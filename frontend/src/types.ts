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
  flagCode: string;
  group?: string;
  isHost?: boolean;
  avgGoalsScored?: number;
  avgGoalsConceded?: number;
}

export interface MatchParams {
  homeTeamId: string;
  awayTeamId: string;
  avgGoals: number;
  homeElo: number | "";
  awayElo: number | "";
  homeMod: number; // -0.5 to +0.5
  awayMod: number; // -0.5 to +0.5
  useDeepSeek: boolean; // enable DeepSeek semantic momentum modifier
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
  // Backend-enriched fields (null when running local fallback)
  deepseekModifier: number | null;
  eloSource: string | null;
  dataFreshnessAt: string | null; // ISO 8601 datetime string
}

// TeamProfile as returned by GET /api/teams
export interface TeamProfile {
  id: string;
  name_en: string;
  name_cn: string;
  flag_code: string;
  group: string;
  is_host: boolean;
  elo: number;
  avg_goals_scored: number;
  avg_goals_conceded: number;
}
