package models

import "time"

// MatchResult represents a single historical match outcome.
type MatchResult struct {
	Date       time.Time
	Opponent   string
	GoalFor    int
	GoalAgainst int
	IsNeutral  bool
}

// MacroContext holds World Bank and Wikimedia contextual data.
type MacroContext struct {
	ISO2        string
	GDP         float64
	Population  int64
	WikiSummary string
}

// FMTeamAttributes holds aggregated Football Manager player attributes per national team.
type FMTeamAttributes struct {
	AvgAttack   float64
	AvgDefense  float64
	PlayerCount int
}

// TeamRawStats is the unified output struct produced by all adapters.
type TeamRawStats struct {
	Source           string
	TeamNameRaw      string
	AvgGoalsScored   float64
	AvgGoalsConceded float64
	EloRating        float64
	RecentForm       []MatchResult
	MacroContext     MacroContext
	FMAttributes     FMTeamAttributes
}

// TeamProfile is the public-facing team representation served by the API.
type TeamProfile struct {
	ID               string  `json:"id"`
	NameEN           string  `json:"name_en"`
	NameCN           string  `json:"name_cn"`
	FlagCode         string  `json:"flag_code"`
	Group            string  `json:"group"`
	IsHost           bool    `json:"is_host"`
	Elo              float64 `json:"elo"`
	AvgGoalsScored   float64 `json:"avg_goals_scored"`
	AvgGoalsConceded float64 `json:"avg_goals_conceded"`
}
