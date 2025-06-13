package model

import "time"

type Position string

const (
	// Top is the top lane
	Top Position = "top"
	// Jungle is the jungle
	Jungle Position = "jungle"
	// Mid is the mid lane
	Mid Position = "mid"
	// Adc is the adc role. This would be normally called bot, but it is called adc here because opgg calls it adc
	Adc Position = "adc"
	// Support is the support role
	Support Position = "support"
)

// Positions is a slice of all the positions
var Positions = []Position{Top, Jungle, Mid, Adc, Support}

// Matchup is a struct that contains the champion name and the win rate against that champion
type Matchup struct {
	WinRate     string `json:"WinRate" bson:"winRate"`
	GamesPlayed string `json:"GamesPlayed" bson:"gamesPlayed"`
}

// RankedChampionStats holds the scraped matchup data for a specific champion, patch, tier, and time.
type RankedChampionStats struct {
	ChampionName string                          `bson:"championName"`
	Patch        string                          `bson:"patch"`
	Tier         string                          `bson:"tier"`
	ScrapedAt    time.Time                       `bson:"scrapedAt"`
	Matchups     map[Position]map[string]Matchup `bson:"matchups"`
}

// Champion is a struct that contains the patch version, the position and the matchups against other champions for that positio
// Deprecated: Use RankedChampionStats instead for MongoDB storage.
type Champion struct {
	Matchups map[Position]map[string]Matchup `json:"Matchups"`
}
