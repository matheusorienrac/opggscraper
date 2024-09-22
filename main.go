package main

import (
	"strings"
	"time"

	"github.com/gocolly/colly"
	"github.com/matheusorienrac/opggscraper/model"
	"github.com/matheusorienrac/opggscraper/scraper"
	"github.com/matheusorienrac/opggscraper/utils"
)

func main() {
	c := colly.NewCollector()

	scraper := scraper.NewScraper(c)

	patchVersions := []string{"14.18"}
	tier := "emerald_plus"

	championList := scraper.GetChampionNames()
	// Champion names need to be cleaned up before they can be used in a ur
	for i := 0; i < len(championList); i++ {
		championList[i] = utils.CleanChampionName(championList[i])
	}

	champions := map[string]model.Champion{}

	for _, patchVersion := range patchVersions {
		for _, championName := range championList {
			// sleeps for 1 second
			championName = strings.ToLower(championName)
			time.Sleep(5 * time.Second)

			champion := model.Champion{}
			champion.Matchups = scraper.GetChampionMatchups(championName, tier, patchVersion)

			champions[championName] = champion
		}
		err := utils.SaveJSON(champions, "championStats_"+patchVersion)
		if err != nil {
			panic(err)
		}

	}

}
