package engine

import (
	"context"
	"errors"
	"fmt"
	"math"

	"wc-predictor/internal/cache"
	"wc-predictor/internal/models"
)

// ErrModifierOutOfRange is returned when HomeMod or AwayMod exceeds ±0.5.
var ErrModifierOutOfRange = errors.New("modifier must be between -0.5 and +0.5")

// ErrTeamNotFound is returned when a requested team has no data in the store.
var ErrTeamNotFound = errors.New("team not found in data store")

// SemanticModifierFunc is a function type for the optional DeepSeek semantic layer.
// It returns a float modifier and a non-nil error if the call failed.
type SemanticModifierFunc func(ctx context.Context, homeCanonical, awayCanonical string, homeForm, awayForm []models.MatchResult) (float64, error)

// Calculate runs the complete three-stage Elo-Poisson prediction pipeline.
//
// Stage 1: Elo delta → net goal correction (linear regression).
// Stage 2: Attack/defense multipliers from recent form.
// Stage 3: Modifier application + optional semantic layer + Poisson matrix.
//
// semanticFn may be nil; if non-nil and req.UseDeepSeek is true, it is called
// after stage 2 and its result is applied to the lambda values.
func Calculate(
	ctx context.Context,
	req *models.MatchRequest,
	store *cache.DataStore,
	semanticFn SemanticModifierFunc,
) (*models.PredictionResult, error) {
	// --- Input validation ---
	if math.Abs(req.HomeMod) > 0.5 || math.Abs(req.AwayMod) > 0.5 {
		return nil, fmt.Errorf("%w: home=%.2f away=%.2f", ErrModifierOutOfRange, req.HomeMod, req.AwayMod)
	}
	if req.AvgGoals <= 0 {
		req.AvgGoals = 2.5
	}

	// --- Stage 1: Fetch team data from cache ---
	homeStats, ok := store.GetTeamStats(req.HomeTeam)
	if !ok {
		return nil, fmt.Errorf("%w: %q", ErrTeamNotFound, req.HomeTeam)
	}
	awayStats, ok := store.GetTeamStats(req.AwayTeam)
	if !ok {
		return nil, fmt.Errorf("%w: %q", ErrTeamNotFound, req.AwayTeam)
	}

	// Resolve Elo: use manually provided value if given, otherwise use store value.
	homeElo := homeStats.EloRating
	if req.HomeElo != nil {
		homeElo = *req.HomeElo
	}
	awayElo := awayStats.EloRating
	if req.AwayElo != nil {
		awayElo = *req.AwayElo
	}

	// World Cup is neutral ground for all non-host teams.
	isNeutral := req.IsNeutral

	// --- Stage 2: Compute expected goals (λ) ---
	// 2a. Elo correction term: distributes net goal advantage across both lambdas.
	deltaGoals := EloToDeltaGoals(homeElo, awayElo, isNeutral)

	// 2b. Attack/defense multipliers from recent form.
	mults := ComputeMultipliers(homeStats, awayStats, req.AvgGoals)
	lambdaHome, lambdaAway := BaseLambdas(req.AvgGoals, mults)

	// 2c. Apply Elo correction symmetrically.
	lambdaHome += deltaGoals / 2.0
	lambdaAway -= deltaGoals / 2.0

	// --- Stage 3: Modifiers and finalization ---
	// 3a. Subjective white-box micro-operation modifier (user-controlled).
	lambdaHome = lambdaHome * (1.0 + req.HomeMod)
	lambdaAway = lambdaAway * (1.0 + req.AwayMod)

	// 3b. Optional semantic momentum modifier from DeepSeek.
	var deepSeekModifier *float64
	if req.UseDeepSeek && semanticFn != nil {
		mod, err := semanticFn(ctx, req.HomeTeam, req.AwayTeam, homeStats.RecentForm, awayStats.RecentForm)
		if err == nil {
			lambdaHome = lambdaHome * (1.0 + mod)
			lambdaAway = lambdaAway * (1.0 - mod)
			deepSeekModifier = &mod
		}
		// On DeepSeek failure, log is handled in the semantic layer; we continue unmodified.
	}

	// 3c. Soft assertion: clamp to minimum to prevent Poisson edge cases.
	lambdaHome = clampLambda(lambdaHome)
	lambdaAway = clampLambda(lambdaAway)

	// --- Poisson matrix generation ---
	matrix := BuildMatrix(lambdaHome, lambdaAway)
	homeWin, draw, awayWin := ExtractOutcomeProbs(matrix)
	top3 := TopNPredictions(matrix, 3)
	alertTriggered, alertMessage := CheckAlertThreshold(matrix)

	eloSource := "eloratings.net"
	if req.HomeElo != nil || req.AwayElo != nil {
		eloSource = "manual_override"
	}

	return &models.PredictionResult{
		HomeExpectedGoals: lambdaHome,
		AwayExpectedGoals: lambdaAway,
		HomeWinProb:       homeWin,
		DrawProb:          draw,
		AwayWinProb:       awayWin,
		Matrix:            matrix,
		TopPredictions:    top3,
		AlertTriggered:    alertTriggered,
		AlertMessage:      alertMessage,
		DeepSeekModifier:  deepSeekModifier,
		EloSource:         eloSource,
		DataFreshnessAt:   store.LastRefresh(),
	}, nil
}
