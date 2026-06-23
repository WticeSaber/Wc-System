package engine

import "os"
import "strconv"

const (
	defaultEloSlope      = 0.0011 // linear regression coefficient: ΔElo → goal advantage
	homeAdvantagePoints  = 75.0   // added to home Elo for non-neutral venues
)

// eloSlope reads the ELO_SLOPE environment variable, falling back to the default.
func eloSlope() float64 {
	if raw := os.Getenv("ELO_SLOPE"); raw != "" {
		if k, err := strconv.ParseFloat(raw, 64); err == nil && k > 0 {
			return k
		}
	}
	return defaultEloSlope
}

// EloToDeltaGoals converts the Elo rating difference between two teams
// into an expected net goal advantage for the home (or designated) team.
//
// For World Cup group-stage matches, isNeutral should be true for all fixtures
// except those involving the host nations (USA, Canada, Mexico).
//
// Returns a positive value when eloA > eloB (team A advantage),
// negative when eloB > eloA (team B advantage).
func EloToDeltaGoals(eloA, eloB float64, isNeutral bool) float64 {
	effectiveEloA := eloA
	if !isNeutral {
		// Non-neutral venue: inflate the home team's effective rating.
		effectiveEloA += homeAdvantagePoints
	}
	return eloSlope() * (effectiveEloA - eloB)
}
