package scraper

import (
	"fmt"
	"log"

	"github.com/gocolly/colly"
	"github.com/matheusorienrac/opggscraper/model"
	"github.com/matheusorienrac/opggscraper/utils"
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
