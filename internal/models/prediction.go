package models

import "time"

// ScorePrediction represents a specific scoreline and its joint probability.
type ScorePrediction struct {
	HomeScore   int     `json:"home_score"`
	AwayScore   int     `json:"away_score"`
	Probability float64 `json:"probability"`
}

// PredictionResult is the full output of the Elo-Poisson engine.
type PredictionResult struct {
	HomeExpectedGoals float64           `json:"home_expected_goals"`
	AwayExpectedGoals float64           `json:"away_expected_goals"`
	HomeWinProb       float64           `json:"home_win_prob"`
	DrawProb          float64           `json:"draw_prob"`
	AwayWinProb       float64           `json:"away_win_prob"`
	Matrix            [6][6]float64     `json:"matrix"`
	TopPredictions    []ScorePrediction `json:"top_predictions"`
	AlertTriggered    bool              `json:"alert_triggered"`
	AlertMessage      string            `json:"alert_message"`
	DeepSeekModifier  *float64          `json:"deepseek_modifier,omitempty"`
	EloSource         string            `json:"elo_source"`
	DataFreshnessAt   time.Time         `json:"data_freshness_at"`
}

// APIError is the standard error response body.
type APIError struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
}
