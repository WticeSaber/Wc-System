package engine

import (
	"fmt"
	"sort"

	"gonum.org/v1/gonum/stat/distuv"
	"wc-predictor/internal/models"
)

const (
	matrixSize     = 6   // probabilities computed for scores 0-5 goals
	alertThreshold = 0.15 // P(0,0) > 15% triggers the extreme-defense warning
)

// BuildMatrix computes the 6×6 joint probability matrix using independent
// Poisson distributions for home and away goals. The cell [i][j] holds
// P(homeGoals=i) × P(awayGoals=j).
func BuildMatrix(lambdaHome, lambdaAway float64) [6][6]float64 {
	poisHome := distuv.Poisson{Lambda: lambdaHome}
	poisAway := distuv.Poisson{Lambda: lambdaAway}

	var matrix [6][6]float64
	for i := 0; i < matrixSize; i++ {
		ph := poisHome.Prob(float64(i))
		for j := 0; j < matrixSize; j++ {
			pa := poisAway.Prob(float64(j))
			matrix[i][j] = ph * pa
		}
	}
	return matrix
}

// ExtractOutcomeProbs applies the diagonal-separation method to the 6×6 matrix:
//   - i > j → home win
//   - i == j → draw
//   - i < j → away win
func ExtractOutcomeProbs(matrix [6][6]float64) (homeWin, draw, awayWin float64) {
	for i := 0; i < matrixSize; i++ {
		for j := 0; j < matrixSize; j++ {
			p := matrix[i][j]
			switch {
			case i > j:
				homeWin += p
			case i == j:
				draw += p
			default:
				awayWin += p
			}
		}
	}
	// Normalize to handle the truncated tail (scores > 5 goals are omitted).
	total := homeWin + draw + awayWin
	if total > 0 {
		homeWin /= total
		draw /= total
		awayWin /= total
	}
	return
}

// TopNPredictions flattens the 6×6 matrix into a sorted slice and returns
// the N highest-probability scorelines, as required by the PRD.
func TopNPredictions(matrix [6][6]float64, n int) []models.ScorePrediction {
	predictions := make([]models.ScorePrediction, 0, matrixSize*matrixSize)
	for i := 0; i < matrixSize; i++ {
		for j := 0; j < matrixSize; j++ {
			predictions = append(predictions, models.ScorePrediction{
				HomeScore:   i,
				AwayScore:   j,
				Probability: matrix[i][j],
			})
		}
	}
	sort.Slice(predictions, func(a, b int) bool {
		return predictions[a].Probability > predictions[b].Probability
	})
	if n > len(predictions) {
		n = len(predictions)
	}
	return predictions[:n]
}

// CheckAlertThreshold evaluates whether P(0,0) breaches the 15% warning threshold.
// When triggered, this indicates both teams' attacking expectations are critically low.
func CheckAlertThreshold(matrix [6][6]float64) (triggered bool, message string) {
	p00 := matrix[0][0]
	if p00 >= alertThreshold {
		triggered = true
		message = fmt.Sprintf(
			"⚠️ 极值预警：当前计算的 0-0 平局概率已突破 %.1f%%（阈值 15%%）！双方进攻期望极度看衰，预计将陷入窒息式僵局态势。",
			p00*100,
		)
	}
	return
}
