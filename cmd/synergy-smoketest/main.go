// Smoke test: fetch synergies for a single champion+tier+patch and print the result.
// Does NOT touch MongoDB. Intended to be run by a developer to sanity-check scraping
// against the live op.gg site after model/scraper changes.
//
// Usage:
//
//	go run ./cmd/synergy-smoketest                  # defaults: jinx, emerald_plus, latest patch
//	go run ./cmd/synergy-smoketest leona support_tier_optional
//	go run ./cmd/synergy-smoketest ashe master_plus 16.9
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/gocolly/colly"
	"github.com/matheusorienrac/opggscraper/scraper"
	"github.com/matheusorienrac/opggscraper/utils"
)

const patchAPIURL = "https://ddragon.leagueoflegends.com/api/versions.json"

func main() {
	championName := "jinx"
	tier := "emerald_plus"
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
	if patch == "" {
		latestPatch, err := utils.GetLatestPatchVersion(patchAPIURL)
		if err != nil {
			log.Fatalf("could not fetch latest patch: %v", err)
		}
		patch = latestPatch
	}

	opggPatch := utils.FormatPatchVersionForOpGG(patch)
	fmt.Printf("Champion=%s Tier=%s Patch=%s (op.gg=%s)\n\n", championName, tier, patch, opggPatch)

	s := scraper.NewScraper(colly.NewCollector())
	synergies := s.GetChampionSynergies(championName, tier, opggPatch, nil) // nil = all 5 positions

	if len(synergies) == 0 {
		log.Fatalf("FAIL: no synergies returned (regex broke or page blocked)")
	}

	totalPairs := 0
	for sourceRole, byPartner := range synergies {
		for partnerRole, partners := range byPartner {
			totalPairs += len(partners)
			fmt.Printf("[%s] partner role %s: %d entries\n", sourceRole, partnerRole, len(partners))
		}
	}
	fmt.Printf("\nTotal synergy pairs across all source roles: %d\n\n", totalPairs)

	// Pretty-print the structure so we can eyeball field correctness.
	out, _ := json.MarshalIndent(synergies, "", "  ")
	fmt.Println(string(out))
}
