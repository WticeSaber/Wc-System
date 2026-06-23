/**
 * @license
 * SPDX-License-Identifier: Apache-2.0
 */

import React, { useState, useEffect } from "react";
import { Sidebar } from "./components/Sidebar";
import { Dashboard } from "./components/Dashboard";
import { MatchParams, PredictionResult } from "./types";
import { runEloPoissonPrediction } from "./utils/predictor";
import { Globe, RefreshCw, BarChart2, ShieldAlert } from "lucide-react";

export default function App() {
  // Setup standard initial parameters (Argentina vs France)
  const [params, setParams] = useState<MatchParams>({
    homeTeamId: "ARG", 
    awayTeamId: "FRA", 
    avgGoals: 2.5,
    homeElo: "",
    awayElo: "",
    homeMod: 0.0,
    awayMod: 0.0,
  });

  // Calculate initial metrics on load
  const [result, setResult] = useState<PredictionResult>(() =>
    runEloPoissonPrediction({
      homeTeamId: "ARG",
      awayTeamId: "FRA",
      avgGoals: 2.5,
      homeElo: "",
      awayElo: "",
      homeMod: 0.0,
      awayMod: 0.0,
    })
  );

  const [isPredicting, setIsPredicting] = useState(false);

  // Trigger automatic subtle computation updates on team selection parameter shifts 
  // so the board behaves dynamically, but preserve the heavy "Run Prediction" action for the primary button!
  const handleParamsChange = (newParams: MatchParams) => {
    setParams(newParams);
    // Silent background pre-calculation ensures direct interactivity
    const instantResult = runEloPoissonPrediction(newParams);
    setResult(instantResult);
  };

  // Heavy simulation prediction calculation mimicking actual model loaders
  const handleRunPrediction = () => {
    setIsPredicting(true);
    setTimeout(() => {
      const freshResult = runEloPoissonPrediction(params);
      setResult(freshResult);
      setIsPredicting(false);
    }, 450); // 450ms perfect cognitive wait-time for matrix calculations
  };

  return (
    <div className="min-h-screen bg-[#0b0f19] bg-[radial-gradient(ellipse_80%_80%_at_50%_-20%,rgba(16,29,54,0.6),rgba(11,15,25,1))] text-white font-sans selection:bg-[#00FF87] selection:text-black">
      
      {/* Exquisite Top Header Section */}
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
                  V2.4 PRO
                </span>
              </div>
              <p className="text-xs text-gray-400 mt-0.5 font-sans">
                2026世界杯比分预测引擎 · 基于Elo等级分、攻击干预系数与双向泊松联合决策概率矩阵
              </p>
            </div>
          </div>

          {/* Top Status Badges */}
          <div className="flex items-center space-x-3 text-xs">
            <div className="bg-[#161a23] border border-[#232a3b] rounded-xl px-3.5 py-1.5 flex items-center space-x-2 font-mono">
              <span className="w-2 h-2 rounded-full bg-[#00FF87] animate-pulse" />
              <span className="text-gray-300">数据源: 实时Elo库</span>
            </div>
            <div className="bg-[#161a23] border border-[#232a3b] rounded-xl px-3.5 py-1.5 flex items-center space-x-2 font-mono">
              <Globe className="w-3.5 h-3.5 text-[#00E5FF]" />
              <span className="text-gray-300">FIFA 2026 赛事库独立解耦</span>
            </div>
          </div>
        </div>
      </header>

      {/* Main Content Area */}
      <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-6">
        <div className="grid grid-cols-1 lg:grid-cols-12 gap-6 items-stretch">
          
          {/* LHS Sidebar Layout - 3.5 Columns on LG screen */}
          <div className="lg:col-span-4 xl:col-span-3.5">
            <Sidebar
              params={params}
              onChange={handleParamsChange}
              onRunPredict={handleRunPrediction}
              isPredicting={isPredicting}
            />
          </div>

          {/* RHS Predictions Dashboard Layout - 8.5 Columns on LG screen */}
          <div className="lg:col-span-8 xl:col-span-8.5">
            <Dashboard
              homeTeamId={params.homeTeamId}
              awayTeamId={params.awayTeamId}
              result={result}
              isPredicting={isPredicting}
            />
          </div>

        </div>
      </main>

      {/* Exquisite Footer */}
      <footer className="border-t border-[#1c2438] bg-[#070a12] py-6 mt-12 text-center text-xs text-gray-500">
        <div className="max-w-7xl mx-auto px-4 flex flex-col sm:flex-row items-center justify-between gap-4 font-mono">
          <p>© 2026 World Cup Poisson Analytics Inc. 所有运算策略独立发布并受Elo算法框架授权保护。</p>
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
