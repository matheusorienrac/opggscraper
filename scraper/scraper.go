package scraper

import (
	"fmt"
	"log"
	"regexp"
	"strconv"

	"github.com/gocolly/colly"
	"github.com/matheusorienrac/opggscraper/model"
	"github.com/matheusorienrac/opggscraper/utils"
)

// Synergy entries arrive inside Next.js RSC stream chunks. OP.GG has served the
// object payload both as escaped JSON and as plain JSON, so the quote matcher
// accepts either form. Field order on op.gg: play, synergy_position, win_rate,
// pick_rate, synergy_champion_name, synergy_champion_image_url,
// synergy_champion_key, tier_rank.
var synergyEntryPattern = regexp.MustCompile(
	`\\?"play\\?":(\d+),` +
		`\\?"synergy_position\\?":\\?"([A-Z]+)\\?",` +
		`\\?"win_rate\\?":([0-9.]+),` +
		`\\?"pick_rate\\?":[0-9.]+,` +
		`\\?"synergy_champion_name\\?":\\?"[^"\\]+\\?",` +
		`\\?"synergy_champion_image_url\\?":\\?"[^"\\]+\\?",` +
		`\\?"synergy_champion_key\\?":\\?"([a-z0-9]+)\\?",` +
		`\\?"tier_rank\\?":(\d+)`,
)

type Scraper struct {
	// Base configuration for creating collectors
	userAgent string
}

// creates a new scraper
func NewScraper(c *colly.Collector) *Scraper {
	return &Scraper{
		userAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
	}
}

// createCollector creates a new collector with standard callbacks
func (s *Scraper) createCollector() *colly.Collector {
	c := colly.NewCollector()

	c.OnRequest(func(r *colly.Request) {
		r.Headers.Set("User-Agent", s.userAgent)
		fmt.Println("Visiting: ", r.URL)
	})

	c.OnError(func(_ *colly.Response, err error) {
		log.Println("Something went wrong: ", err)
	})

	c.OnResponse(func(r *colly.Response) {
		fmt.Println("Page visited: ", r.Request.URL)
	})

	c.OnScraped(func(r *colly.Response) {
		fmt.Println(r.Request.URL, " scraped!")
	})

	return c
}

// Gets the champion matchups from the website by Position
func (s *Scraper) GetChampionMatchupsByPosition(championName string, pos model.Position, tier string, patchVersion string) map[string]model.Matchup {
	matchups := map[string]model.Matchup{}

	championNames := []string{}
	championWinrates := []string{}
	championGamesPlayed := []string{}

	// Create a new collector for this specific scrape
	c := s.createCollector()

	c.OnHTML("ul > li", func(e *colly.HTMLElement) {
		championNames = append(championNames, utils.CleanChampionName(e.ChildText("div:nth-child(2) > span")))
		championWinrates = append(championWinrates, e.ChildText("div:nth-child(3) > strong"))
		championGamesPlayed = append(championGamesPlayed, e.ChildText("div:nth-child(4) > span"))
	})

	c.Visit("https://www.op.gg/champions/" + championName + "/counters/" + string(pos) + "?region=global&tier=" + tier + "&patch=" + patchVersion)
	fmt.Println(championNames)

	for i := 0; i < len(championNames); i++ {
		matchup := model.Matchup{}
		matchup.WinRate = championWinrates[i]
		matchup.GamesPlayed = championGamesPlayed[i]
		matchups[championNames[i]] = matchup
	}
	fmt.Println(matchups)
	return matchups
}

// Gets champion names from the website
func (s *Scraper) GetChampionNames() []string {
	championNames := []string{}

	// Create a new collector for this specific scrape
	c := s.createCollector()

	c.OnHTML("span.truncate", func(e *colly.HTMLElement) {
		championName := e.Text
		if championName != "" {
			championNames = append(championNames, championName)
			//fmt.Println("Added champion:", championName) // Debug print
		}
	})

	err := c.Visit("https://www.op.gg/champions")
	if err != nil {
		fmt.Println("Error visiting page:", err)
		return championNames
	}
	fmt.Println("Total champions found:", len(championNames)) // Debug print

	return championNames
}

// Gets the champion matchups for all positions from the website
func (s *Scraper) GetChampionMatchups(championName string, tier string, patchVersion string) map[model.Position]map[string]model.Matchup {
	matchupsAllPositions := map[model.Position]map[string]model.Matchup{}

	for _, position := range model.Positions {
		matchupsByPosition := s.GetChampionMatchupsByPosition(championName, position, tier, patchVersion)

		// Create a new map for the current position
		matchupsForPosition := make(map[string]model.Matchup)
		for key, value := range matchupsByPosition {
			matchupsForPosition[key] = value
		}

		// Store the matchups for the current position
		matchupsAllPositions[position] = matchupsForPosition
	}

	return matchupsAllPositions
}

// Gets champion synergies (best duo partners) from OP.GG for a single source position.
// Returns a map keyed by partner role → partner champion key → Synergy.
// The op.gg synergies page lists 10 partners per partner-role for the four roles other than the source.
func (s *Scraper) GetChampionSynergiesByPosition(championName string, pos model.Position, tier string, patchVersion string) map[model.Position]map[string]model.Synergy {
	out := map[model.Position]map[string]model.Synergy{}

	c := s.createCollector()

	c.OnResponse(func(r *colly.Response) {
		body := string(r.Body)
		matches := synergyEntryPattern.FindAllStringSubmatch(body, -1)
		for _, m := range matches {
			play, _ := strconv.Atoi(m[1])
			partnerRole, ok := utils.OPGGPositionToModel(m[2])
			if !ok {
				continue
			}
			wr, err := strconv.ParseFloat(m[3], 64)
			if err != nil {
				continue
			}
			partnerKey := m[4]
			rank, _ := strconv.Atoi(m[5])

			if _, exists := out[partnerRole]; !exists {
				out[partnerRole] = map[string]model.Synergy{}
			}
			// Defensive: skip if we already have an entry with at least as many games.
			if existing, dup := out[partnerRole][partnerKey]; dup && existing.GamesPlayed >= play {
				continue
			}
			out[partnerRole][partnerKey] = model.Synergy{
				WinRate:     wr,
				GamesPlayed: play,
				TierRank:    rank,
			}
		}
	})

	url := "https://www.op.gg/lol/champions/" + championName + "/synergies/" + string(pos) +
		"?region=global&tier=" + tier + "&patch=" + patchVersion
	if err := c.Visit(url); err != nil {
		log.Printf("    ERROR visiting synergies for %s/%s: %v", championName, pos, err)
	}
	return out
}

// Gets champion synergies for the given source positions.
// Returns map keyed as result[sourceRole][partnerRole][partnerChampion].
// When positions is nil, all five are scraped; pass a filtered list to avoid wasted
// HTTP requests for off-role pages (~60-75% of champions are only played in 1-2 roles).
func (s *Scraper) GetChampionSynergies(championName string, tier string, patchVersion string, positions []model.Position) map[model.Position]map[model.Position]map[string]model.Synergy {
	if positions == nil {
		positions = model.Positions
	}
	all := map[model.Position]map[model.Position]map[string]model.Synergy{}
	for _, pos := range positions {
		byPartnerRole := s.GetChampionSynergiesByPosition(championName, pos, tier, patchVersion)
		if len(byPartnerRole) == 0 {
			continue
		}
		all[pos] = byPartnerRole
	}
	return all
}

// // Gets matchups for all championNames in the list. Requires colly async to be true
// func GetChampionMatchupsFromList(championNames []string, tier string, patchVersion string) map[string]map[model.Position][]model.Matchup {
// 	matchups := map[string]map[model.Position][]model.Matchup{}

// 	// list of urls to visit
// 	urls := []string{}

// 	for _, championName := range championNames {
// 		for _, position := range model.Positions {
// 			urls = append(urls, "https://www.op.gg/champions/"+championName+"/"+string(position)+"/counters?region=global&tier="+tier+"&patch="+patchVersion)
// 		}

// 		matchups[championName] = scraper.GetChampionMatchups(championName, tier, patchVersion)

// 	scraper.Collector.Wait()

// 	return matchups
// }
// j
