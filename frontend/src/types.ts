/**
 * @license
 * SPDX-License-Identifier: Apache-2.0
 */

export interface Team {
  id: string;
  name: string;
  nameEn: string;
  emoji: string;
  defaultElo: number;
  flagCode: string; // 2-letter lower case ISO code for FlagCDN
}

export interface MatchParams {
  homeTeamId: string;
  awayTeamId: string;
  avgGoals: number;
  homeElo: number | "";
  awayElo: number | "";
  homeMod: number; // -0.5 to +0.5
  awayMod: number; // -0.5 to +0.5
}

export interface ScorePrediction {
  homeScore: number;
  awayScore: number;
  probability: number;
}

export interface PredictionResult {
  homeExpectedGoals: number;
  awayExpectedGoals: number;
  homeWinProb: number;
  drawProb: number;
  awayWinProb: number;
  matrix: number[][]; // 6x6 matrix of probabilities [homeGoals][awayGoals]
  topPredictions: ScorePrediction[]; // List of most likely scorelines
  alertTriggered: boolean;
  alertMessage: string;
}
