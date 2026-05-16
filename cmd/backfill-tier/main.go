// One-off patch+tier backfill.
//
// Usage:
//
//	go run ./cmd/backfill-tier master_plus 16.9
//	go build -o backfill-tier ./cmd/backfill-tier
//	./backfill-tier master_plus 16.9
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gocolly/colly"
	"github.com/matheusorienrac/opggscraper/db"
	"github.com/matheusorienrac/opggscraper/model"
	"github.com/matheusorienrac/opggscraper/scraper"
	"github.com/matheusorienrac/opggscraper/utils"
)

const mongoURI = "mongodb://localhost:27017"

func main() {
	tier := "master_plus"
	patch := "16.9"
	if len(os.Args) > 1 {
		tier = os.Args[1]
	}
	if len(os.Args) > 2 {
		patch = os.Args[2]
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	dbPatch := utils.FormatPatchVersion(patch)
	opggPatch := utils.FormatPatchVersionForOpGG(patch)
	log.Printf("Starting one-off backfill tier=%s patch(db)=%s patch(opgg)=%s", tier, dbPatch, opggPatch)

	dbClient, err := db.ConnectDB(mongoURI)
	if err != nil {
		log.Fatalf("connect MongoDB: %v", err)
	}
	defer dbClient.Disconnect(context.Background())

	s := scraper.NewScraper(colly.NewCollector())
	championNames := s.GetChampionNames()
	if len(championNames) == 0 {
		log.Fatalf("no champions found")
	}
	log.Printf("Found %d champions", len(championNames))

	started := time.Now()
	successes := 0
	failures := 0
	for i, rawName := range championNames {
		if ctx.Err() != nil {
			log.Printf("Backfill cancelled after %d/%d champions", i, len(championNames))
			return
		}

		championName := utils.CleanChampionName(rawName)
		log.Printf("[%d/%d] scraping %s", i+1, len(championNames), championName)

		matchups := s.GetChampionMatchups(championName, tier, opggPatch)
		if len(matchups) == 0 || !utils.ValidateChampionData(matchups) {
			failures++
			log.Printf("[%d/%d] WARN no valid matchups for %s; skipping save", i+1, len(championNames), championName)
			sleep(ctx, 2*time.Second)
			continue
		}

		synergies := s.GetChampionSynergies(championName, tier, opggPatch, utils.PositionsWithMatchups(matchups))
		stats := model.RankedChampionStats{
			ChampionName: championName,
			Patch:        dbPatch,
			Tier:         tier,
			ScrapedAt:    time.Now(),
			Matchups:     matchups,
			Synergies:    synergies,
		}

		if err := saveWithReconnect(ctx, &dbClient, stats); err != nil {
			failures++
			log.Printf("[%d/%d] ERROR saving %s: %v", i+1, len(championNames), championName, err)
			sleep(ctx, 2*time.Second)
			continue
		}

		successes++
		log.Printf("[%d/%d] saved %s with %d synergy pairs", i+1, len(championNames), championName, countSynergies(synergies))
		sleep(ctx, 2*time.Second)
	}

	log.Printf("Backfill complete tier=%s patch=%s successes=%d failures=%d elapsed=%s", tier, dbPatch, successes, failures, time.Since(started).Round(time.Second))
}

func saveWithReconnect(ctx context.Context, clientRef **db.Client, stats model.RankedChampionStats) error {
	if err := (*clientRef).SaveChampionStats(ctx, stats); err == nil {
		return nil
	} else {
		log.Printf("save failed for %s; reconnecting MongoDB once: %v", stats.ChampionName, err)
	}

	(*clientRef).Disconnect(context.Background())
	next, err := db.ConnectDB(mongoURI)
	if err != nil {
		return fmt.Errorf("reconnect MongoDB: %w", err)
	}
	*clientRef = next
	return (*clientRef).SaveChampionStats(ctx, stats)
}

func countSynergies(synergies map[model.Position]map[model.Position]map[string]model.Synergy) int {
	total := 0
	for _, byRole := range synergies {
		for _, partners := range byRole {
			total += len(partners)
		}
	}
	return total
}

func sleep(ctx context.Context, d time.Duration) {
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-timer.C:
	case <-ctx.Done():
	}
}
