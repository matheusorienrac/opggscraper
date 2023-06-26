package model

// Matchup is a struct that contains the name of the champion and its win rate against the champion that is being analyzed
type Matchup struct {
	ChampionName string
	WinRate      string
}

// ChampionCounters is a struct that contains the name of the champion and a list of its matchups
type ChampionCounters struct {
	ChampionName string
	Matchups     []Matchup
}
