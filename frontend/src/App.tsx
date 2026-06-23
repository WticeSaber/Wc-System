/**
 * @license
 * SPDX-License-Identifier: Apache-2.0
 */

import React, { useState, useEffect } from "react";
import { Sidebar } from "./components/Sidebar";
import { Dashboard } from "./components/Dashboard";
import { MatchParams, PredictionResult, Team } from "./types";
import { runEloPoissonPrediction } from "./utils/predictor";
import { fetchPrediction, fetchTeams } from "./api";
import { loadTeamsFromRegistry, mergeTeamProfiles } from "./data/teamsRegistry";
import { Globe } from "lucide-react";

const DEFAULT_PARAMS: MatchParams = {
  homeTeamId: "ARG",
  awayTeamId: "FRA",
  avgGoals: 2.5,
  homeElo: "",
  awayElo: "",
  homeMod: 0.0,
  awayMod: 0.0,
  useDeepSeek: false,
};

export default function App() {
  const [teams, setTeams] = useState<Team[]>(() => loadTeamsFromRegistry());
  const [params, setParams] = useState<MatchParams>(DEFAULT_PARAMS);

  const [result, setResult] = useState<PredictionResult>(() =>
    runEloPoissonPrediction(DEFAULT_PARAMS, loadTeamsFromRegistry())
  );

  const [isPredicting, setIsPredicting] = useState(false);
  const [backendOnline, setBackendOnline] = useState<boolean | null>(null);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);

  useEffect(() => {
    fetch("/api/health")
      .then((r) => {
        if (r.ok) {
          setBackendOnline(true);
          fetchTeams()
            .then((profiles) => {
              setTeams((prev) => mergeTeamProfiles(prev, profiles));
            })
            .catch(() => {
              // Keep registry-based list when API teams fetch fails.
            });
        } else {
          setBackendOnline(false);
        }
      })
      .catch(() => setBackendOnline(false));
  }, []);

  const handleParamsChange = (newParams: MatchParams) => {
    setParams(newParams);
    const instantResult = runEloPoissonPrediction(newParams, teams);
    setResult(instantResult);
    setErrorMessage(null);
  };

  const handleRunPrediction = async () => {
    setIsPredicting(true);
    setErrorMessage(null);
    try {
      const freshResult = await fetchPrediction(params);
      setResult(freshResult);
    } catch (err) {
      const fallback = runEloPoissonPrediction(params, teams);
      setResult(fallback);
      const msg = err instanceof Error ? err.message : String(err);
      setErrorMessage(`后端不可达，已切换至本地计算模式。(${msg})`);
    } finally {
      setIsPredicting(false);
    }
  };

  return (
    <div className="min-h-screen bg-[#0b0f19] bg-[radial-gradient(ellipse_80%_80%_at_50%_-20%,rgba(16,29,54,0.6),rgba(11,15,25,1))] text-white font-sans selection:bg-[#00FF87] selection:text-black">

      <header className="border-b border-[#1c2438] bg-[#0b0f19]/80 backdrop-blur-md sticky top-0 z-40">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-4 flex flex-col md:flex-row items-center justify-between gap-4">
          <div className="flex items-center space-x-3 text-center md:text-left">
            <div className="w-10 h-10 rounded-xl bg-gradient-to-tr from-[#00FF87] to-[#00E5FF] p-0.5 shadow-lg shadow-[#00FF87]/15">
              <div className="w-full h-full bg-[#0b0f19] rounded-[10px] flex items-center justify-center">
                <span className="text-xl select-none">🏆</span>
              </div>
            </div>
            <div>
              <div className="flex flex-wrap items-center gap-2 justify-center md:justify-start">
                <h1 className="text-xl font-black tracking-tight text-white uppercase font-sans">
                  2026 World Cup Elo-Poisson Predictor
                </h1>
                <span className="bg-gradient-to-r from-[#00FF87] to-[#00E5FF] text-black font-extrabold text-[9px] px-2 py-0.5 rounded-full font-mono uppercase tracking-widest leading-none">
                  V3.0 · {teams.length} TEAMS
                </span>
              </div>
              <p className="text-xs text-gray-400 mt-0.5 font-sans">
                2026世界杯比分预测引擎 · 球队名单维护于 data/wc2026_teams.json
              </p>
            </div>
          </div>

          <div className="flex items-center space-x-3 text-xs">
            <div className="bg-[#161a23] border border-[#232a3b] rounded-xl px-3.5 py-1.5 flex items-center space-x-2 font-mono">
              <span
                className={`w-2 h-2 rounded-full ${
                  backendOnline === null
                    ? "bg-gray-500"
                    : backendOnline
                    ? "bg-[#00FF87] animate-pulse"
                    : "bg-[#FF4A6B]"
                }`}
              />
              <span className="text-gray-300">
                {backendOnline === null
                  ? "连接中..."
                  : backendOnline
                  ? "Go 引擎: 在线"
                  : "Go 引擎: 离线 (本地模式)"}
              </span>
            </div>
            <div className="bg-[#161a23] border border-[#232a3b] rounded-xl px-3.5 py-1.5 flex items-center space-x-2 font-mono">
              <Globe className="w-3.5 h-3.5 text-[#00E5FF]" />
              <span className="text-gray-300">FIFA 2026 赛事库独立解耦</span>
            </div>
          </div>
        </div>
      </header>

      {errorMessage && (
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 pt-4">
          <div className="bg-yellow-950/60 border border-yellow-600/30 text-yellow-300 text-xs font-mono px-4 py-2.5 rounded-xl flex items-center gap-2">
            <span className="text-yellow-400">⚠</span>
            {errorMessage}
          </div>
        </div>
      )}

      <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-6">
        <div className="grid grid-cols-1 lg:grid-cols-12 gap-6 items-stretch">
          <div className="lg:col-span-4 xl:col-span-3">
            <Sidebar
              teams={teams}
              params={params}
              onChange={handleParamsChange}
              onRunPredict={handleRunPrediction}
              isPredicting={isPredicting}
            />
          </div>
          <div className="lg:col-span-8 xl:col-span-9">
            <Dashboard
              teams={teams}
              homeTeamId={params.homeTeamId}
              awayTeamId={params.awayTeamId}
              result={result}
              isPredicting={isPredicting}
            />
          </div>
        </div>
      </main>

      <footer className="border-t border-[#1c2438] bg-[#070a12] py-6 mt-12 text-center text-xs text-gray-500">
        <div className="max-w-7xl mx-auto px-4 flex flex-col sm:flex-row items-center justify-between gap-4 font-mono">
          <p>© 2026 World Cup Poisson Analytics — 球队数据: data/wc2026_teams.json</p>
          <div className="flex space-x-4">
            <span className="hover:text-gray-300 cursor-help" title="基于独立事件联合泊松假设">泊松分布说明</span>
            <span className="text-gray-700">|</span>
            <span className="hover:text-gray-300 cursor-help" title="Elo评级是由世界足坛代表权重的评级模型推导">Elo等级标准</span>
          </div>
        </div>
      </footer>
    </div>
  );
}
