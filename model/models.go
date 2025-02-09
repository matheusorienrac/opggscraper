package model

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
	WinRate     string `json:"WinRate"`
	GamesPlayed string `json:"GamesPlayed"`
}

// Champion is a struct that contains the patch version, the position and the matchups against other champions for that positio
type Champion struct {
	Matchups map[Position]map[string]Matchup `json:"Matchups"`
}
