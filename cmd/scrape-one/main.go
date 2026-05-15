// One-shot end-to-end scrape harness: fetches matchups + synergies for a single champion
// and saves to MongoDB through the same code path the daemon uses.
//
// Usage:
//
//	go run ./cmd/scrape-one                          # defaults: jinx, master_plus, latest patch
//	go run ./cmd/scrape-one <champion> <tier>        # e.g. go run ./cmd/scrape-one leona master_plus
//	go run ./cmd/scrape-one <champion> <tier> <patch> # e.g. go run ./cmd/scrape-one ashe master_plus 16.9
//
// When run inside a Claude Code sandbox, the SOCKS5 proxy at 127.0.0.1:1080 is auto-detected
// and used to reach the host's MongoDB — without it, the sandbox's empty loopback yields
// connection-refused on port 27017.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gocolly/colly"
	"github.com/matheusorienrac/opggscraper/db"
	"github.com/matheusorienrac/opggscraper/model"
	"github.com/matheusorienrac/opggscraper/scraper"
	"github.com/matheusorienrac/opggscraper/utils"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/net/proxy"
)

const patchAPIURL = "https://ddragon.leagueoflegends.com/api/versions.json"

// buildMongoOptions returns ClientOptions targeting localhost:27017. Inside a Claude Code
// sandbox (CLAUDECODE=1) it attaches a SOCKS5 dialer at 127.0.0.1:1080 so connections route
// out of the sandbox to the host network where MongoDB actually listens.
func buildMongoOptions() (*options.ClientOptions, string, error) {
	uri := "mongodb://localhost:27017"
	opts := options.Client().ApplyURI(uri)
	if os.Getenv("CLAUDECODE") != "1" {
		return opts, uri, nil
	}

	socks, err := proxy.SOCKS5("tcp", "127.0.0.1:1080", nil, proxy.Direct)
	if err != nil {
		return nil, uri, fmt.Errorf("socks5 dialer: %w", err)
	}
	cd, ok := socks.(proxy.ContextDialer)
	if !ok {
		return nil, uri, fmt.Errorf("socks5 dialer does not implement ContextDialer")
	}
	opts = opts.SetDialer(cd)
	return opts, uri + " (via socks5 127.0.0.1:1080)", nil
}

func main() {
	championName := "jinx"
	tier := "master_plus"
	if len(os.Args) > 1 {
		championName = os.Args[1]
	}
	if len(os.Args) > 2 {
		tier = os.Args[2]
	}
	patch := ""
	if len(os.Args) > 3 {
		patch = os.Args[3]
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// --- Patch ---
	if patch == "" {
		latestPatch, err := utils.GetLatestPatchVersion(patchAPIURL)
		if err != nil {
			log.Fatalf("could not fetch latest patch: %v", err)
		}
		patch = latestPatch
	}
	dbPatch := utils.FormatPatchVersion(patch)
	opggPatch := utils.FormatPatchVersionForOpGG(patch)
	fmt.Printf("Champion=%s  Tier=%s  Patch(db)=%s  Patch(opgg)=%s\n", championName, tier, dbPatch, opggPatch)

	// --- DB connect ---
	mongoOpts, mongoLabel, err := buildMongoOptions()
	if err != nil {
		log.Fatalf("mongo options: %v", err)
	}
	fmt.Printf("Mongo: %s\n\n", mongoLabel)
	dbClient, err := db.ConnectDBWithOptions(mongoOpts)
	if err != nil {
		log.Fatalf("connect: %v", err)
	}
	defer dbClient.Disconnect(ctx)

	// --- Scrape ---
	s := scraper.NewScraper(colly.NewCollector())

	fmt.Println(">>> Scraping matchups …")
	matchups := s.GetChampionMatchups(championName, tier, opggPatch)
	if len(matchups) == 0 || !utils.ValidateChampionData(matchups) {
		log.Fatalf("matchups missing or invalid")
	}
	matchupCount := 0
	for _, byChamp := range matchups {
		matchupCount += len(byChamp)
	}
	fmt.Printf("    matchups: %d entries across %d source roles\n\n", matchupCount, len(matchups))

	fmt.Println(">>> Scraping synergies …")
	synergies := s.GetChampionSynergies(championName, tier, opggPatch, utils.PositionsWithMatchups(matchups))
	pairs := 0
	for _, byPartner := range synergies {
		for _, partners := range byPartner {
			pairs += len(partners)
		}
	}
	fmt.Printf("    synergies: %d pairs across %d source roles\n\n", pairs, len(synergies))

	// --- Save ---
	stats := model.RankedChampionStats{
		ChampionName: championName,
		Patch:        dbPatch,
		Tier:         tier,
		ScrapedAt:    time.Now(),
		Matchups:     matchups,
		Synergies:    synergies,
	}
	fmt.Println(">>> Upserting to MongoDB …")
	if err := dbClient.SaveChampionStats(ctx, stats); err != nil {
		log.Fatalf("save: %v", err)
	}
	fmt.Printf("    saved.\n\n")

	// --- Summary of synergies (just the ADC source role to keep output short) ---
	if adcSynergies, ok := synergies[model.Adc]; ok {
		out, _ := json.MarshalIndent(adcSynergies, "", "  ")
		fmt.Printf(">>> Synergies for %s (source role: ADC):\n%s\n", championName, string(out))
	} else {
		fmt.Println(">>> No ADC source-role synergies (champion may not be played ADC).")
	}
}
