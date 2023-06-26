package scraper

import (
	"fmt"
	"log"

	"github.com/gocolly/colly"
	"github.com/matheusorienrac/opggscraper/model"
)

type Scraper struct {
	collector *colly.Collector
}

// creates a new scraper
func NewScraper(c *colly.Collector) *Scraper {
	c.OnRequest(func(r *colly.Request) {
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

	return &Scraper{collector: c}
}

// Gets the champion counters from the website
func (s *Scraper) GetChampionCounters(championName string, role string, tier string) *model.ChampionCounters {
	championCounters := &model.ChampionCounters{}

	s.collector.OnHTML("tr.eocu2m74", func(e *colly.HTMLElement) {
		matchup := model.Matchup{}

		matchup.ChampionName = e.ChildText("td:nth-child(2) > div > div.eocu2m71")
		matchup.WinRate = e.ChildText("td:nth-child(3) > span")
		championCounters.Matchups = append(championCounters.Matchups, matchup)
		fmt.Println(matchup)
	})

	s.collector.Visit("https://www.op.gg/champions/" + championName + "/" + role + "/counters?region=global&tier=" + tier)

	return championCounters
}

// Gets champion names from the website
func (s *Scraper) GetChampionNames() []string {
	championNames := []string{}

	s.collector.OnHTML("nav.e1y3xkpj1 > ul", func(e *colly.HTMLElement) {
		e.ForEach("li", func(_ int, el *colly.HTMLElement) {
			championNames = append(championNames, el.ChildText("a > span"))
		})

		championNames = append(championNames, e.ChildText("div.champion-index__champion-item__name"))
	})

	s.collector.Visit("https://www.op.gg/champions")

	return championNames
}
