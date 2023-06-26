package main

import (
	"github.com/gocolly/colly"
	"github.com/matheusorienrac/opggscraper/model"
	"github.com/matheusorienrac/opggscraper/scraper"
	"github.com/matheusorienrac/opggscraper/utils"
)

func main() {
	c := colly.NewCollector()
	scraper := scraper.NewScraper(c)

	patchVersion := "13.12"
	championList := scraper.GetChampionNames()
	tier := "master"

	for _, championName := range championList {
		champion := &model.Champion{}

		champion.Name = championName
		champion.PatchVersion = patchVersion
		champion.Matchups = scraper.GetChampionMatchups(championName, tier, champion.PatchVersion)

		err := utils.SaveJSON(champion, champion.Name+"_"+champion.PatchVersion)
		if err != nil {
			panic(err)
		}
	}

}
