package utils

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/matheusorienrac/opggscraper/model"
)

// Unmarshalls champion data to JSON and saves it to a file
func SaveJSON(champions map[string]model.Champion, filename string) error {

	// Marshal the map into JSON
	jsonData, err := json.MarshalIndent(champions, "", "    ")
	if err != nil {
		return err
	}

	// Save the JSON data to a file
	err = ioutil.WriteFile(filename+".json", jsonData, 0644)
	if err != nil {
		return err
	}

	fmt.Println("JSON data saved to " + filename + ".json")

	return nil
}

// CleanChampionName takes a champion name and returns a cleaned up version of it so it can be used in a url
func CleanChampionName(championName string) string {
	// special cases

	switch championName {
	case "Nunu & Willump":
		return "nunu"
	case "Wukong":
		return "monkeyking"
	case "Renata Glasc":
		return "renata"
	}

	championName = strings.Replace(championName, "'", "", -1)
	championName = strings.Replace(championName, ".", "", -1)
	championName = strings.Replace(championName, " ", "", -1)

	// make everything lower case because riot is not very consistent about which letters are capitalized
	championName = strings.ToLower(championName)

	return championName
}

// ValidateChampionData checks if the champion data contains at least one valid winrate
func ValidateChampionData(matchups map[model.Position]map[string]model.Matchup) bool {
	// Check if we have any matchup data
	if len(matchups) == 0 {
		return false
	}

	// Look for at least one valid winrate (contains %)
	for _, positionMatchups := range matchups {
		for _, matchup := range positionMatchups {
			if strings.Contains(matchup.WinRate, "%") {
				return true
			}
		}
	}

	return false
}
