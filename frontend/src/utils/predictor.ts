/**
 * @license
 * SPDX-License-Identifier: Apache-2.0
 */

import { Team, MatchParams, PredictionResult, ScorePrediction } from "../types";
import { WORLD_CUP_TEAMS } from "../data/teams";

// Standard factorial calculations
function factorial(n: number): number {
  if (n <= 1) return 1;
  let result = 1;
  for (let i = 2; i <= n; i++) {
    result *= i;
  }
  return result;
}

// Poisson probability distribution: P(k; lambda) = (lambda^k * e^-lambda) / k!
export function poissonProbability(k: number, lambda: number): number {
  if (lambda <= 0) return k === 0 ? 1 : 0;
  return (Math.pow(lambda, k) * Math.exp(-lambda)) / factorial(k);
}

export function runEloPoissonPrediction(params: MatchParams): PredictionResult {
  const homeTeam = WORLD_CUP_TEAMS.find((t) => t.id === params.homeTeamId) || WORLD_CUP_TEAMS[0];
  const awayTeam = WORLD_CUP_TEAMS.find((t) => t.id === params.awayTeamId) || WORLD_CUP_TEAMS[1];

  // Determine actual Elo ratings used
  const homeElo = typeof params.homeElo === "number" ? params.homeElo : homeTeam.defaultElo;
  const awayElo = typeof params.awayElo === "number" ? params.awayElo : awayTeam.defaultElo;

  // Expected goals (lambda) calculations based on Elo delta & Avg goals
  // Let lambda baseline be half of avg goals
  const lambdaBaseline = params.avgGoals / 2;

  // Elo rating difference impact (for every 400 Elo points delta, expected goals scales by 10^(ΔElo/800))
  const eloDelta = homeElo - awayElo;
  const homeScaling = Math.pow(10, eloDelta / 800);
  const awayScaling = Math.pow(10, -eloDelta / 800);

  // Apply subjective micro-modifications
  const homeExpectedGoals = Math.max(0.05, lambdaBaseline * homeScaling + params.homeMod);
  const awayExpectedGoals = Math.max(0.05, lambdaBaseline * awayScaling + params.awayMod);

  // Compute full joint probabilities (from 0 to 12 goals) for high-accuracy outcome states
  const maxSize = 15;
  const homeProbList = Array.from({ length: maxSize }, (_, i) => poissonProbability(i, homeExpectedGoals));
  const awayProbList = Array.from({ length: maxSize }, (_, i) => poissonProbability(i, awayExpectedGoals));

  let homeWinProb = 0;
  let drawProb = 0;
  let awayWinProb = 0;

  for (let h = 0; h < maxSize; h++) {
    for (let a = 0; a < maxSize; a++) {
      const p = homeProbList[h] * awayProbList[a];
      if (h > a) {
        homeWinProb += p;
      } else if (h === a) {
        drawProb += p;
      } else {
        awayWinProb += p;
      }
    }
  }

  // Normalize outcomes to sum to exactly 1.0
  const totalOutcomeProb = homeWinProb + drawProb + awayWinProb;
  if (totalOutcomeProb > 0) {
    homeWinProb /= totalOutcomeProb;
    drawProb /= totalOutcomeProb;
    awayWinProb /= totalOutcomeProb;
  }

  // Generate the 6x6 representation matrix (goals from 0 to 5 for both Home and Away)
  const matrix: number[][] = [];
  const allScorePredictions: ScorePrediction[] = [];

  for (let h = 0; h < 6; h++) {
    matrix[h] = [];
    for (let a = 0; a < 6; a++) {
      // Joint probability
      const p = poissonProbability(h, homeExpectedGoals) * poissonProbability(a, awayExpectedGoals);
      matrix[h][a] = p;
      
      allScorePredictions.push({
        homeScore: h,
        awayScore: a,
        probability: p
      });
    }
  }

  // Grab the top score candidates, sorted descending by probability
  allScorePredictions.sort((a, b) => b.probability - a.probability);
  const topPredictions = allScorePredictions.slice(0, 3);

  // Calculate the raw 0-0 draw probability within the entire system
  const p00 = poissonProbability(0, homeExpectedGoals) * poissonProbability(0, awayExpectedGoals);
  const alertTriggered = p00 >= 0.12; // Trigger warning above 12% 0-0 probability
  const alertPercentageStr = (p00 * 100).toFixed(1);
  const alertMessage = `⚠️ 极值预警：当前计算的 0-0 平局概率已突破 ${alertPercentageStr}%（阈值 12%）！双方进攻期望极度看衰，预计将陷入窒息式僵局态势。`;

  return {
    homeExpectedGoals,
    awayExpectedGoals,
    homeWinProb,
    drawProb,
    awayWinProb,
    matrix,
    topPredictions,
    alertTriggered,
    alertMessage,
  };
}
