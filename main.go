package main

import (
	"context"
	"flag"
	"log"
	"os/signal"
	"syscall"
	"time"

	"github.com/gocolly/colly"
	"github.com/matheusorienrac/opggscraper/db"
	"github.com/matheusorienrac/opggscraper/model"
	"github.com/matheusorienrac/opggscraper/scraper"
	"github.com/matheusorienrac/opggscraper/utils"
)

const (
	// Define your MongoDB connection string here
	mongoURI = "mongodb://localhost:27017"
	// Riot API for patch versions
	patchApiURL = "https://ddragon.leagueoflegends.com/api/versions.json"
)

func main() {
	// --- Setup ---
	log.Println("Starting OP.GG Scraper...")

	// Create a cancellable context based on OS signals
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel() // Ensure context is cancelled even if not triggered by signal

	// Connect to MongoDB
	dbClient, err := db.ConnectDB(mongoURI)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer func() {
		// Use a separate context for disconnection, as the main one might be cancelled
		disconnectCtx, disconnectCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer disconnectCancel()
		log.Println("Disconnecting from MongoDB...")
		dbClient.Disconnect(disconnectCtx)
	}()

	c := colly.NewCollector()
	scraper := scraper.NewScraper(c)

	// --- Ticker for Daily Execution ---
	ticker := time.NewTicker(24 * time.Hour) // Run once a day
	defer ticker.Stop()

	// --- Initial Run & Loop ---
	// Accept extra patch versions as CLI args (e.g., ./opggscraper 16.2).
	// These are scraped once on startup in addition to the auto-detected latest patch.
	// Use --once with explicit patch args to run a one-shot backfill and exit.
	once := flag.Bool("once", false, "run one scrape cycle and exit")
	flag.Parse()
	extraPatches := flag.Args()

	log.Println("Performing initial scrape...")
	includeLatest := !*once || len(extraPatches) == 0
	scrapeAndSave(ctx, dbClient, scraper, extraPatches, includeLatest)

	if *once {
		log.Println("One-shot scrape complete. Exiting.")
		return
	}
	log.Println("Initial scrape complete. Waiting for next scheduled run...")

	for {
		select {
		case <-ticker.C:
			if ctx.Err() != nil { // Check if context was cancelled while waiting
				log.Println("Context cancelled, skipping scheduled run.")
				return
			}
			log.Println("Scheduled scrape starting...")
			scrapeAndSave(ctx, dbClient, scraper, nil, true)
			log.Println("Scheduled scrape complete. Waiting for next run...")
		case <-ctx.Done(): // Wait for the context to be cancelled by the signal
			log.Println("Shutdown signal received. Exiting...")
			return // Exit the program, defer statements will run
		}
	}
}

// scrapeAndSave performs the full scraping and saving process, respecting context cancellation.
// extraPatches are additional patch versions to scrape (e.g., passed via CLI args on startup).
func scrapeAndSave(ctx context.Context, dbClient *db.Client, scraper *scraper.Scraper, extraPatches []string, includeLatest bool) {
	patchVersions := make([]string, 0, 1+len(extraPatches))
	seen := make(map[string]bool, 1+len(extraPatches))

	if includeLatest {
		latestPatch, err := utils.GetLatestPatchVersion(patchApiURL)
		if err != nil {
			if len(extraPatches) == 0 {
				log.Printf("ERROR: Could not fetch latest patch version: %v. Skipping scrape cycle.", err)
				return
			}
			log.Printf("WARN: Could not fetch latest patch version: %v. Scraping explicit patches only: %v", err, extraPatches)
		} else {
			log.Printf("Latest patch version identified: %s", latestPatch)
			patchVersions = append(patchVersions, latestPatch)
			seen[latestPatch] = true
		}
	} else {
		log.Printf("Skipping latest patch lookup; scraping explicit patches only: %v", extraPatches)
	}

	for _, p := range extraPatches {
		if !seen[p] {
			patchVersions = append(patchVersions, p)
			seen[p] = true
		}
	}
	if len(patchVersions) > 1 {
		log.Printf("Will scrape %d patches: %v", len(patchVersions), patchVersions)
	}
	tiers := []string{"emerald_plus", "diamond_plus", "master_plus", "grandmaster", "challenger"}

	// Check for cancellation before starting heavy work
	if ctx.Err() != nil {
		log.Println("Context cancelled before starting scrape cycle.")
		return
	}

	championList := scraper.GetChampionNames()
	// Champion names need to be cleaned up before they can be used in a URL
	cleanedChampionList := make([]string, len(championList))
	for i := 0; i < len(championList); i++ {
		cleanedChampionList[i] = utils.CleanChampionName(championList[i])
	}

	// Process each patch and tier (now typically only the latest patch)
	for _, patchVersion := range patchVersions {
		// Format patch for DB (e.g., 15.7 -> 15.7, 15.10 -> 15.10)
		dbFormattedPatch := utils.FormatPatchVersion(patchVersion)
		// Format patch for OP.GG URL (e.g., 15.7 -> 15.07, 15.10 -> 15.10)
		opggFormattedPatch := utils.FormatPatchVersionForOpGG(patchVersion)

		for _, tier := range tiers {
			// --- Check for cancellation before starting tier ---
			select {
			case <-ctx.Done():
				log.Printf("Context cancelled before starting tier %s for patch %s.", tier, patchVersion)
				return // Exit scrapeAndSave
			default:
				// Continue if not cancelled
			}

			// Context-aware sleep between tiers to avoid rate limiting
			log.Printf("Waiting 15 minutes before scraping tier: %s...", tier)
			timer := time.NewTimer(15 * time.Minute)
			select {
			case <-timer.C:
				// Timer finished
			case <-ctx.Done():
				timer.Stop() // Stop the timer if cancelled
				log.Printf("Context cancelled during wait for tier %s.", tier)
				return // Exit scrapeAndSave
			}

			log.Printf("Scraping data for Patch: %s (OP.GG: %s, DB: %s), Tier: %s", patchVersion, opggFormattedPatch, dbFormattedPatch, tier)
			now := time.Now()

			// Scrape and save data for each champion individually
			for _, championName := range cleanedChampionList {
				// --- Check for cancellation before scraping champion ---
				select {
				case <-ctx.Done():
					log.Printf("Context cancelled before scraping champion %s for tier %s.", championName, tier)
					return // Exit scrapeAndSave
				default:
					// Continue
				}

				log.Printf("  Scraping matchups for: %s", championName)

				// Context-aware sleep between champions
				champTimer := time.NewTimer(2 * time.Second)
				select {
				case <-champTimer.C:
					// Timer finished
				case <-ctx.Done():
					champTimer.Stop()
					log.Printf("Context cancelled during wait for champion %s.", championName)
					return // Exit scrapeAndSave
				}

				// Use the OP.GG formatted patch for scraping
				matchups := scraper.GetChampionMatchups(championName, tier, opggFormattedPatch)

				// Validate the scraped data
				if len(matchups) == 0 {
					log.Printf("    WARN: No matchups found for %s (Patch: %s, Tier: %s). Skipping save.", championName, opggFormattedPatch, tier)
					continue
				}

				// Use the new validation function
				if !utils.ValidateChampionData(matchups) {
					log.Printf("    WARN: No valid winrates found for %s (Patch: %s, Tier: %s). Skipping save.", championName, opggFormattedPatch, tier)
					continue
				}

				// Scrape synergies — only for positions where the matchup scrape returned data,
				// since off-role synergy pages are almost always empty and waste HTTP requests.
				activePositions := utils.PositionsWithMatchups(matchups)
				synergies := scraper.GetChampionSynergies(championName, tier, opggFormattedPatch, activePositions)
				synergyPairs := 0
				for _, byPartner := range synergies {
					for _, partners := range byPartner {
						synergyPairs += len(partners)
					}
				}
				if synergyPairs == 0 {
					log.Printf("    INFO: No synergies found for %s (Patch: %s, Tier: %s). Saving matchups only.", championName, opggFormattedPatch, tier)
				} else {
					log.Printf("    Found %d synergy pairs for %s", synergyPairs, championName)
				}

				// Save immediately after successful scraping and validation
				stats := model.RankedChampionStats{
					ChampionName: championName,
					Patch:        dbFormattedPatch, // Use the DB formatted patch
					Tier:         tier,
					ScrapedAt:    now,
					Matchups:     matchups,
					Synergies:    synergies,
				}

				err := dbClient.SaveChampionStats(ctx, stats)
				if err != nil {
					// Log error but continue with next champion unless context is cancelled
					if ctx.Err() != nil {
						log.Printf("Context cancelled during save operation for %s: %v", championName, ctx.Err())
						return
					}
					log.Printf("    ERROR saving stats for %s: %v", championName, err)
				} else {
					log.Printf("    Successfully saved data for %s", championName)
				}
			}

			// Check for cancellation after all champions
			if ctx.Err() == nil {
				log.Printf("Finished processing all champions for Patch: %s, Tier: %s.", patchVersion, tier)
			}
		}
	}
	// Check for cancellation at the very end
	if ctx.Err() == nil {
		log.Println("Finished scraping cycle.")
	}
}
