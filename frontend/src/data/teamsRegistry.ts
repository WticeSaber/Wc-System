/**
 * @license
 * SPDX-License-Identifier: Apache-2.0
 *
 * 球队名单唯一数据源：data/wc2026_teams.json
 * 维护说明见 data/README.md
 */

import roster from "../../../data/wc2026_teams.json";
import { Team, TeamProfile } from "../types";

export interface TeamRegistryEntry {
  id: string;
  name_en: string;
  name_cn: string;
  flag_code: string;
  iso2: string;
  canonical_name: string;
  csv_aliases: string[];
  elo_code: string;
  group: string;
  is_host: boolean;
  wiki_title?: string;
}

export interface TeamRegistryFile {
  version: string;
  tournament: string;
  updated_at: string;
  teams: TeamRegistryEntry[];
}

const registry = roster as TeamRegistryFile;

/** Default Elo when live stats are unavailable (offline fallback). */
const FALLBACK_ELO = 1500;

/** Load UI teams from the shared JSON registry. */
export function loadTeamsFromRegistry(): Team[] {
  return registry.teams.map((entry) => ({
    id: entry.id,
    name: entry.name_cn,
    nameEn: entry.name_en,
    emoji: "",
    flagCode: entry.flag_code,
    defaultElo: FALLBACK_ELO,
    group: entry.group,
    isHost: entry.is_host,
  }));
}

/** Merge live API stats (Elo, goals) into registry-based team list. */
export function mergeTeamProfiles(base: Team[], profiles: TeamProfile[]): Team[] {
  const byId = new Map(profiles.map((p) => [p.id, p]));
  return base.map((team) => {
    const live = byId.get(team.id);
    if (!live) return team;
    return {
      ...team,
      defaultElo: live.elo > 0 ? live.elo : team.defaultElo,
      avgGoalsScored: live.avg_goals_scored,
      avgGoalsConceded: live.avg_goals_conceded,
    };
  });
}

export function findTeam(teams: Team[], id: string): Team {
  return teams.find((t) => t.id === id) ?? teams[0];
}

export function getRegistryMeta() {
  return {
    version: registry.version,
    tournament: registry.tournament,
    updatedAt: registry.updated_at,
    teamCount: registry.teams.length,
  };
}

export const WORLD_CUP_TEAMS: Team[] = loadTeamsFromRegistry();
