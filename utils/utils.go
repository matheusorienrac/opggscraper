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

// PositionsWithMatchups returns the source positions where the scraped matchups contain
// at least one valid winrate. Used to skip synergy fetches for roles a champion isn't played in.
func PositionsWithMatchups(matchups map[model.Position]map[string]model.Matchup) []model.Position {
	var out []model.Position
	for _, pos := range model.Positions {
		for _, m := range matchups[pos] {
			if strings.Contains(m.WinRate, "%") {
				out = append(out, pos)
				break
			}
		}
	}
	return out
}

// OPGGPositionToModel maps OP.GG's two-letter position codes (TO, JU, MI, AD, SU) to model.Position.
// Returns ok=false for unknown codes so callers can skip them defensively.
func OPGGPositionToModel(code string) (model.Position, bool) {
	switch strings.ToUpper(code) {
	case "TO", "TOP":
		return model.Top, true
	case "JU", "JUNGLE":
		return model.Jungle, true
	case "MI", "MID", "MIDDLE":
		return model.Mid, true
	case "AD", "ADC", "BOT", "BOTTOM":
		return model.Adc, true
	case "SU", "SUP", "SUPPORT":
		return model.Support, true
	}
	return "", false
}
