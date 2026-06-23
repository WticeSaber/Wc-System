/**
 * @license
 * SPDX-License-Identifier: Apache-2.0
 */

import { MatchParams, PredictionResult, TeamProfile } from "./types";

/**
 * fetchPrediction calls the Go backend POST /api/predict endpoint.
 * Maps camelCase MatchParams to the backend's snake_case JSON contract.
 */
export async function fetchPrediction(params: MatchParams): Promise<PredictionResult> {
  const res = await fetch("/api/predict", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({
      home_team: params.homeTeamId,
      away_team: params.awayTeamId,
      avg_goals: params.avgGoals,
      home_elo: params.homeElo !== "" ? params.homeElo : null,
      away_elo: params.awayElo !== "" ? params.awayElo : null,
      home_mod: params.homeMod,
      away_mod: params.awayMod,
      use_deepseek: params.useDeepSeek,
      is_neutral: true,
    }),
  });

  if (!res.ok) {
    const err = await res.json().catch(() => ({ message: `HTTP ${res.status}` }));
    throw new Error(err.message || `API error ${res.status}`);
  }

  // Map snake_case backend response to camelCase TypeScript interface.
  const raw = await res.json();
  return mapBackendResponse(raw);
}

/**
 * fetchTeams calls GET /api/teams and returns the team profile list.
 * Used to refresh the team roster from live backend data at startup.
 */
export async function fetchTeams(): Promise<TeamProfile[]> {
  const res = await fetch("/api/teams");
  if (!res.ok) {
    throw new Error(`Teams API error ${res.status}`);
  }
  return res.json();
}

/**
 * mapBackendResponse converts the backend's snake_case JSON response
 * into the camelCase PredictionResult expected by React components.
 *
 * The matrix from the backend is [6][6]float64 — it arrives as a
 * 2D JSON array and is stored directly (no transformation needed).
 */
function mapBackendResponse(raw: Record<string, unknown>): PredictionResult {
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const rawPredictions = (raw.top_predictions as any[]) ?? [];
  const topPredictions = rawPredictions.map((p) => ({
    homeScore: p.home_score,
    awayScore: p.away_score,
    probability: p.probability,
  }));

  return {
    homeExpectedGoals: raw.home_expected_goals as number,
    awayExpectedGoals: raw.away_expected_goals as number,
    homeWinProb: raw.home_win_prob as number,
    drawProb: raw.draw_prob as number,
    awayWinProb: raw.away_win_prob as number,
    matrix: raw.matrix as number[][],
    topPredictions,
    alertTriggered: raw.alert_triggered as boolean,
    alertMessage: raw.alert_message as string,
    deepseekModifier: raw.deepseek_modifier != null ? (raw.deepseek_modifier as number) : null,
    eloSource: raw.elo_source as string ?? null,
    dataFreshnessAt: raw.data_freshness_at as string ?? null,
  };
}
