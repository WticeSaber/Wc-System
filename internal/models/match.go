package models

// MatchRequest is the input to the prediction engine, received from CLI or HTTP.
type MatchRequest struct {
	HomeTeam    string   `json:"home_team"`
	AwayTeam    string   `json:"away_team"`
	AvgGoals    float64  `json:"avg_goals"`    // default 2.5
	HomeElo     *float64 `json:"home_elo"`     // nil = auto-fetch from store
	AwayElo     *float64 `json:"away_elo"`     // nil = auto-fetch from store
	HomeMod     float64  `json:"home_mod"`     // subjective modifier, -0.5 to +0.5
	AwayMod     float64  `json:"away_mod"`     // subjective modifier, -0.5 to +0.5
	UseDeepSeek bool     `json:"use_deepseek"` // enable semantic momentum modifier
	IsNeutral   bool     `json:"is_neutral"`   // World Cup = true for non-hosts
}
