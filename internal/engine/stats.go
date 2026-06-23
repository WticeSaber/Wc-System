package engine

import "wc-predictor/internal/models"

// AttackDefenseMultipliers holds the dimensionless scaling factors derived
// from each team's recent form relative to the match's macroeconomic baseline.
type AttackDefenseMultipliers struct {
	HomeAttack  float64 // home.AvgGoalsScored / avgGoals
	HomeDef     float64 // home.AvgGoalsConceded / avgGoals  (unused in lambda formula)
	AwayAttack  float64 // away.AvgGoalsScored / avgGoals    (unused in lambda formula)
	AwayDef     float64 // away.AvgGoalsConceded / avgGoals
}

// ComputeMultipliers derives attack and defense multipliers from team stats and the
// match-level baseline (avgGoals). This implements the Dixon-Coles-inspired
// attack × defense decomposition described in the PRD.
//
// lambdaBaseHome = avgGoals * homeAttack * awayDef
// lambdaBaseAway = avgGoals * awayAttack * homeDef
func ComputeMultipliers(home, away *models.TeamRawStats, avgGoals float64) AttackDefenseMultipliers {
	if avgGoals <= 0 {
		avgGoals = 2.5
	}

	homeAtk := home.AvgGoalsScored / avgGoals
	homeDef := home.AvgGoalsConceded / avgGoals
	awayAtk := away.AvgGoalsScored / avgGoals
	awayDef := away.AvgGoalsConceded / avgGoals

	return AttackDefenseMultipliers{
		HomeAttack: homeAtk,
		HomeDef:    homeDef,
		AwayAttack: awayAtk,
		AwayDef:    awayDef,
	}
}

// BaseLambdas returns the baseline expected goals for each team before
// Elo correction and subjective modifiers are applied.
func BaseLambdas(avgGoals float64, m AttackDefenseMultipliers) (lambdaHome, lambdaAway float64) {
	lambdaHome = avgGoals * m.HomeAttack * m.AwayDef
	lambdaAway = avgGoals * m.AwayAttack * m.HomeDef
	return
}

// clampLambda enforces the PRD's soft assertion: λ must never fall below 0.01
// to prevent Poisson distribution edge-case failures.
func clampLambda(lambda float64) float64 {
	const minLambda = 0.01
	if lambda < minLambda {
		return minLambda
	}
	return lambda
}
