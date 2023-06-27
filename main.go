package main

import (
	"time"

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

	champions := map[string]*model.Champion{}
	// for i, championName := range championList {
	// 	if championName == "Maokai" {
	// 		championList = championList[i:]
	// 		break
	// 	}
	// }

	for _, championName := range championList {
		time.Sleep(30 * time.Second)
		champion := &model.Champion{}
		champion.PatchVersion = patchVersion
		champion.Matchups = scraper.GetChampionMatchups(championName, tier, champion.PatchVersion)

		champions[championName] = champion
	}

	err := utils.SaveJSON(champions, "championStats_"+patchVersion)
	if err != nil {
		panic(err)
	}

}
