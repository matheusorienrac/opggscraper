package main

import (
	"fmt"

	"github.com/gocolly/colly"
	"github.com/matheusorienrac/opggscraper/scraper"
)

func main() {
	c := colly.NewCollector()
	scraper := scraper.NewScraper(c)

	scraper.GetChampionCounters("Aatrox", "top", "master")

	fmt.Println(scraper.GetChampionNames())
}

// Gets counters for a champion
// func getChampionCounters(championName string) *model.ChampionCounters {
