/**
 * @license
 * SPDX-License-Identifier: Apache-2.0
 */

import React, { useState } from "react";
import { Team, PredictionResult, ScorePrediction } from "../types";
import { WORLD_CUP_TEAMS } from "../data/teams";
import { ShieldAlert, TrendingUp, Info, HelpCircle, Activity } from "lucide-react";

interface DashboardProps {
  homeTeamId: string;
  awayTeamId: string;
  result: PredictionResult;
  isPredicting: boolean;
}

export function Dashboard({ homeTeamId, awayTeamId, result, isPredicting }: DashboardProps) {
  const homeTeam = WORLD_CUP_TEAMS.find((t) => t.id === homeTeamId) || WORLD_CUP_TEAMS[0];
  const awayTeam = WORLD_CUP_TEAMS.find((t) => t.id === awayTeamId) || WORLD_CUP_TEAMS[1];

  // Hover state for the 6x6 heatmap cells
  const [hoveredCell, setHoveredCell] = useState<{ h: number; a: number; prob: number } | null>(null);

  // Parse percentages
  const homeWinPct = (result.homeWinProb * 100).toFixed(1);
  const drawPct = (result.drawProb * 100).toFixed(1);
  const awayWinPct = (result.awayWinProb * 100).toFixed(1);

  // Find max probability in the matrix to scale the opacity gradients correctly
  let maxProb = 0.001;
  for (let h = 0; h < 6; h++) {
    for (let a = 0; a < 6; a++) {
      if (result.matrix[h][a] > maxProb) {
        maxProb = result.matrix[h][a];
      }
    }
  }

  // Define podium scores in order (1st in center, 2nd on left, 3rd on right)
  const p1 = result.topPredictions[0];
  const p2 = result.topPredictions[1];
  const p3 = result.topPredictions[2];

  // Helper to color each cell in heatmap
  const getCellBgColor = (prob: number) => {
    const ratio = prob / maxProb;
    // We'll use our signature fluorescent green: rgb(0, 255, 135) mapped by ratio with standard dark background overlay
    return {
      backgroundColor: `rgba(0, 255, 135, ${Math.max(0.04, ratio * 0.85)})`,
      borderColor: ratio > 0.6 ? "rgba(0, 255, 135, 0.4)" : "rgba(35, 42, 59, 0.4)",
    };
  };

  return (
    <div className="space-y-6 text-gray-100 flex flex-col h-full">
      {/* ⚠️ Extreme Alert Banner */}
      {result.alertTriggered && (
        <div
          id="extreme-defense-alert"
          className="bg-gradient-to-r from-red-950 via-red-900 to-red-950 border border-red-500/30 text-red-200 p-4 rounded-xl shadow-lg flex items-center space-x-3.5 animate-pulse relative overflow-hidden transition-all"
        >
          {/* Subtle animated red background glow */}
          <div className="absolute inset-0 bg-red-600/5 select-none pointer-events-none" />
          <ShieldAlert className="w-5 h-5 text-red-400 shrink-0" />
          <div className="text-xs font-sans leading-relaxed tracking-wide">
            {result.alertMessage}
          </div>
        </div>
      )}

      {/* Main Grid Wrapper */}
      <div className="grid grid-cols-1 lg:grid-cols-10 gap-6">
        
        {/* Left Component: Macro Results Distribution (Gauge & Legend Stats) - 5 Cols */}
        <div className="lg:col-span-4 bg-[#161a23] border border-[#232a3b] rounded-2xl p-5 shadow-xl flex flex-col justify-between">
          <div>
            <h3 className="text-gray-400 text-xs font-bold uppercase tracking-widest flex items-center mb-4">
              <Activity className="w-3.5 h-3.5 text-[#00E5FF] mr-1.5" />
              宏观赛果概率分布 <span className="text-[9px] font-mono font-normal ml-2 text-gray-500">MACRO DISTRIBUTION</span>
            </h3>

            {/* Custom SVG Half Donut / Arch Gauge */}
            <div className="relative flex flex-col items-center justify-center my-4">
              <svg width="220" height="130" className="rotate-0">
                {/* Background Arch track */}
                <path
                  d="M 20 120 A 90 90 0 0 1 200 120"
                  fill="none"
                  stroke="#1c2333"
                  strokeWidth="16"
                  strokeLinecap="round"
                />
                
                {/* Dynamically drawing three arcs representing Home Win, Draw, Away Win */}
                {/* Total circumference = PI * radius = 3.14159 * 90 = 282.7px */}
                {/* Draw Home Win */}
                <path
                  d="M 20 120 A 90 90 0 0 1 200 120"
                  fill="none"
                  stroke="#00FF87"
                  strokeWidth="16"
                  strokeLinecap="round"
                  strokeDasharray={`${result.homeWinProb * 282.7} 282.7`}
                  strokeDashoffset="0"
                  className="transition-all duration-1000 ease-out"
                />
                
                {/* Draw Draw (offset by home win) */}
                <path
                  d="M 20 120 A 90 90 0 0 1 200 120"
                  fill="none"
                  stroke="#00E5FF"
                  strokeWidth="16"
                  strokeLinecap="round"
                  strokeDasharray={`${result.drawProb * 282.7} 282.7`}
                  strokeDashoffset={`-${result.homeWinProb * 282.7}`}
                  className="transition-all duration-1000 ease-out"
                />
                
                {/* Draw Away Win (offset by home win + draw) */}
                <path
                  d="M 20 120 A 90 90 0 0 1 200 120"
                  fill="none"
                  stroke="#FF4A6B"
                  strokeWidth="16"
                  strokeLinecap="round"
                  strokeDasharray={`${result.awayWinProb * 282.7} 282.7`}
                  strokeDashoffset={`-${(result.homeWinProb + result.drawProb) * 282.7}`}
                  className="transition-all duration-1000 ease-out"
                />
              </svg>

              {/* Central Details Overlay */}
              <div className="absolute bottom-1 text-center">
                <span className="text-[10px] text-gray-500 uppercase tracking-widest block font-mono">
                  期待进球数 (xG)
                </span>
                <span className="text-lg font-mono font-extrabold text-white">
                  {result.homeExpectedGoals.toFixed(2)} : {result.awayExpectedGoals.toFixed(2)}
                </span>
                <span className="text-[10px] text-gray-400 flex items-center justify-center gap-1.5 mt-0.5 font-semibold">
                  <img src={`https://flagcdn.com/w40/${homeTeam.flagCode}.png`} alt="" className="w-4.5 h-3 object-cover rounded-sm border border-gray-800/40" referrerPolicy="no-referrer" />
                  <span>{homeTeam.name}</span>
                  <span className="text-gray-600 font-mono">VS</span>
                  <span>{awayTeam.name}</span>
                  <img src={`https://flagcdn.com/w40/${awayTeam.flagCode}.png`} alt="" className="w-4.5 h-3 object-cover rounded-sm border border-gray-800/40" referrerPolicy="no-referrer" />
                </span>
              </div>
            </div>

            {/* Structured Stats Percent Legend */}
            <div className="space-y-3 mt-4">
              {/* Home Win info */}
              <div className="flex items-center justify-between text-xs font-sans bg-[#111622]/60 p-2.5 rounded-xl border border-[#1e2536]">
                <div className="flex items-center space-x-2">
                  <div className="w-2.5 h-2.5 rounded-full bg-[#00FF87]" />
                  <span className="text-gray-300 font-medium">主胜 (Home Win)</span>
                </div>
                <div className="text-right">
                  <span className="font-mono text-sm font-extrabold text-[#00FF87]">{homeWinPct}%</span>
                </div>
              </div>

              {/* Draw info */}
              <div className="flex items-center justify-between text-xs font-sans bg-[#111622]/60 p-2.5 rounded-xl border border-[#1e2536]">
                <div className="flex items-center space-x-2">
                  <div className="w-2.5 h-2.5 rounded-full bg-[#00E5FF]" />
                  <span className="text-gray-300 font-medium">平局 (Draw)</span>
                </div>
                <div className="text-right">
                  <span className="font-mono text-sm font-extrabold text-[#00E5FF]">{drawPct}%</span>
                </div>
              </div>

              {/* Away Win info */}
              <div className="flex items-center justify-between text-xs font-sans bg-[#111622]/60 p-2.5 rounded-xl border border-[#1e2536]">
                <div className="flex items-center space-x-2">
                  <div className="w-2.5 h-2.5 rounded-full bg-[#FF4A6B]" />
                  <span className="text-gray-300 font-medium">客胜 (Away Win)</span>
                </div>
                <div className="text-right">
                  <span className="font-mono text-sm font-extrabold text-[#FF4A6B]">{awayWinPct}%</span>
                </div>
              </div>
            </div>
          </div>

          {/* Three-segment horizontal solid progress bar */}
          <div className="mt-5 pt-4 border-t border-[#232a3b]/60">
            <div className="flex justify-between text-[10px] text-gray-400 font-mono mb-1.5 uppercase tracking-wider">
              <span>比例趋势</span>
              <span>SUM: 100%</span>
            </div>
            <div className="h-3.5 w-full rounded-lg overflow-hidden flex bg-[#111622] border border-[#232a3b]">
              <div
                style={{ width: `${homeWinPct}%` }}
                className="bg-[#00FF87] h-full transition-all duration-1000 ease-out"
                title={`主胜: ${homeWinPct}%`}
              />
              <div
                style={{ width: `${drawPct}%` }}
                className="bg-[#00E5FF] h-full transition-all duration-1000 ease-out"
                title={`平局: ${drawPct}%`}
              />
              <div
                style={{ width: `${awayWinPct}%` }}
                className="bg-[#FF4A6B] h-full transition-all duration-1000 ease-out"
                title={`客胜: ${awayWinPct}%`}
              />
            </div>
            <div className="flex justify-between items-center text-[9px] font-mono text-gray-500 mt-1">
              <span>{homeTeam.id} {homeWinPct}%</span>
              <span>DRAW {drawPct}%</span>
              <span>{awayTeam.id} {awayWinPct}%</span>
            </div>
          </div>
        </div>

        {/* Right Component: Podium Scoreline Predictions (领奖台样式) - 6 Cols */}
        <div className="lg:col-span-6 bg-[#161a23] border border-[#232a3b] rounded-2xl p-5 shadow-xl flex flex-col justify-between">
          <div>
            <h3 className="text-gray-400 text-xs font-bold uppercase tracking-widest flex items-center mb-4">
              <TrendingUp className="w-3.5 h-3.5 text-[#00FF87] mr-1.5" />
              核心推荐及候补比分预测 <span className="text-[9px] font-mono font-normal ml-2 text-gray-500">SCOREBOARD EXPECTANCY</span>
            </h3>

            {/* Podium Visual Frame */}
            {p1 && p2 && p3 ? (
              <div className="grid grid-cols-3 gap-3.5 items-end pt-8 pb-4 relative min-h-[220px]">
                
                {/* 2nd Place: Backup 1 (Left Podium) */}
                <div
                  id="podium-rank-2"
                  className="bg-[#1c2231]/80 hover:bg-[#1c2231] border border-[#2d3a56]/80 rounded-2xl p-4 flex flex-col items-center justify-between transition-all duration-300 transform hover:-translate-y-1 h-[155px] text-center shadow-lg relative group"
                >
                  <div className="absolute top-0 -translate-y-1/2 bg-[#00E5FF] text-[#0b111e] rounded-full w-5 h-5 flex items-center justify-center font-mono text-xs font-extrabold border border-[#161a23]">
                    2
                  </div>
                  <div className="text-[10px] text-gray-400 font-sans tracking-wide truncate w-full uppercase">
                    候补预测 1
                  </div>
                  <div className="my-2">
                    <span className="text-2xl font-mono font-extrabold text-[#00E5FF] tracking-tight">
                      {p2.homeScore}-{p2.awayScore}
                    </span>
                  </div>
                  <div className="bg-[#111622] rounded-lg px-2.5 py-1 font-mono text-xs text-gray-300 font-bold border border-[#26324d]">
                    {(p2.probability * 100).toFixed(2)}%
                  </div>
                </div>

                {/* 1st Place: Top Prediction (Center Podium - Tallest, Glow Bordered) */}
                <div
                  id="podium-rank-1"
                  className="bg-gradient-to-b from-[#1d273a] to-[#121824] border-2 border-[#00FF87] rounded-2xl p-4 flex flex-col items-center justify-between transition-all duration-300 transform hover:-translate-y-1.5 h-[190px] text-center shadow-[0_0_25px_rgba(0,255,135,0.15)] relative group z-10"
                >
                  {/* Glowing halo indicator */}
                  <div className="absolute -inset-0.5 bg-gradient-to-r from-[#00FF87] to-[#00E5FF] rounded-2xl opacity-10 blur group-hover:opacity-20 transition duration-300 -z-10" />
                  
                  <div className="absolute top-0 -translate-y-1/2 bg-[#00FF87] text-[#0b111e] rounded-full w-6.5 h-6.5 flex items-center justify-center font-mono text-sm font-extrabold border-2 border-[#161a23] shadow-md shadow-[#00FF87]/20">
                    1
                  </div>
                  <div className="text-[11px] font-bold text-[#00FF87] font-sans tracking-widest uppercase">
                    第一顺位预测
                  </div>
                  <div className="my-3">
                    <span className="text-3xl font-mono font-black text-white tracking-tight drop-shadow-[0_0_10px_rgba(255,255,255,0.1)]">
                      {p1.homeScore}-{p1.awayScore}
                    </span>
                  </div>
                  <div className="bg-[#111622] rounded-xl px-4 py-1.5 font-mono text-xs text-[#00FF87] font-extrabold border border-[#00FF87]/30 shadow-inner">
                    {(p1.probability * 100).toFixed(2)}%
                  </div>
                </div>

                {/* 3rd Place: Backup 2 (Right Podium) */}
                <div
                  id="podium-rank-3"
                  className="bg-[#1c2231]/80 hover:bg-[#1c2231] border border-[#2d3a56]/80 rounded-2xl p-4 flex flex-col items-center justify-between transition-all duration-300 transform hover:-translate-y-1 h-[135px] text-center shadow-lg relative group"
                >
                  <div className="absolute top-0 -translate-y-1/2 bg-[#FF4A6B] text-[#0b111e] rounded-full w-5 h-5 flex items-center justify-center font-mono text-xs font-extrabold border border-[#161a23]">
                    3
                  </div>
                  <div className="text-[10px] text-gray-400 font-sans tracking-wide truncate w-full uppercase">
                    候补预测 2
                  </div>
                  <div className="my-2">
                    <span className="text-2xl font-mono font-extrabold text-[#FF4A6B] tracking-tight">
                      {p3.homeScore}-{p3.awayScore}
                    </span>
                  </div>
                  <div className="bg-[#111622] rounded-lg px-2.5 py-1 font-mono text-xs text-gray-300 font-bold border border-[#26324d]">
                    {(p3.probability * 100).toFixed(2)}%
                  </div>
                </div>

              </div>
            ) : (
              <div className="h-[200px] flex items-center justify-center text-gray-400">数据运算未激活</div>
            )}
          </div>

          {/* Quick analysis summary paragraph */}
          <div className="bg-[#111622]/80 border border-[#212b3f] rounded-xl p-3 mt-4 text-xs text-gray-300 leading-relaxed font-sans flex items-start space-x-2">
            <Info className="w-4 h-4 text-[#00E5FF] shrink-0 mt-0.5" />
            <p>
              领奖台算法基于独立泊松分布模型联合运算所得。根据最新胜率判定：
              <span className="text-[#00FF87] font-bold"> {homeTeam.name}</span> 有 
              <span className="font-mono text-[#00FF87] font-bold"> {homeWinPct}%</span> 概率胜出，
              最看好的赛果总进球期望为 
              <span className="font-mono text-[#00E5FF] font-bold"> {(result.homeExpectedGoals + result.awayExpectedGoals).toFixed(2)}</span> 个。
            </p>
          </div>
        </div>

      </div>

      {/* 📊 Section 2: Deep 6x6 Joint Probability Heatmap */}
      <div className="bg-[#161a23] border border-[#232a3b] rounded-2xl p-6 shadow-xl flex-1 flex flex-col justify-between" id="heatmap-section-card">
        <div>
          <div className="flex flex-col sm:flex-row sm:items-center justify-between pb-4 border-b border-[#232a3b] mb-4 space-y-2 sm:space-y-0">
            <div>
              <h3 className="text-gray-400 text-xs font-bold uppercase tracking-widest flex items-center">
                <span className="text-lg mr-2">📊</span>
                深度概率分布矩阵 (Heatmap Matrix)
              </h3>
              <p className="text-[11px] text-gray-400 mt-1">
                行 (Row) 代表主队 {homeTeam.name} 进球数，列 (Col) 代表客队 {awayTeam.name} 进球数。单元格代表特定比分的联合概率。
              </p>
            </div>
            
            {/* Color mapping key */}
            <div className="flex items-center space-x-2.5">
              <span className="text-[10px] font-mono text-gray-500 uppercase">概率分布深浅:</span>
              <div className="flex items-center space-x-1">
                <span className="text-[9px] text-gray-500 font-mono">0%</span>
                <div className="h-2 w-16 rounded-full bg-gradient-to-r from-rgba(0,255,135,0.05) to-[#00FF87] border border-[#2d3a56]" style={{ background: "linear-gradient(to right, rgba(0,255,135,0.05), rgba(0,255,135,0.85))" }} />
                <span className="text-[9px] text-gray-500 font-mono">Max ({(maxProb * 100).toFixed(1)}%)</span>
              </div>
            </div>
          </div>

          {/* Heatmap Matrix Grid */}
          <div className="relative overflow-auto py-2">
            <div className="min-w-[480px] max-w-full mx-auto">
              
              {/* Top X Axis Header: Away goals */}
              <div className="flex">
                {/* Empty cell for row labels offsets */}
                <div className="w-[110px] shrink-0" />
                <div className="flex-1 grid grid-cols-6 text-center text-xs font-mono font-bold py-1.5 uppercase tracking-wide text-gray-400 border-b border-[#232a3b]/40">
                  <div className="col-span-6 text-[10px] font-sans text-center pb-1 text-[#FF4A6B]">客队 {awayTeam.name} 进球数 (0-5)</div>
                  {Array.from({ length: 6 }).map((_, i) => (
                    <div key={i} className="py-1">{i} 球</div>
                  ))}
                </div>
              </div>

              {/* Grid content */}
              <div className="space-y-1 mt-2">
                {Array.from({ length: 6 }).map((_, hScore) => (
                  <div key={hScore} className="flex items-stretch h-14">
                    {/* Left Y Axis Label: Home Goals */}
                    <div className="w-[110px] shrink-0 flex flex-col justify-center pr-3 border-r border-[#232a3b]/40 text-right">
                      {hScore === 0 && (
                        <div className="text-[9px] font-sans uppercase text-[#00FF87] tracking-wider mb-0.5 leading-none font-bold">
                          主队进球 (0-5)
                        </div>
                      )}
                      <span className="text-xs font-mono font-bold text-gray-300">
                        {hScore} 球
                      </span>
                    </div>

                    {/* Row loop */}
                    <div className="flex-1 grid grid-cols-6 gap-1 pl-1">
                      {Array.from({ length: 6 }).map((_, aScore) => {
                        const cellProb = result.matrix[hScore][aScore];
                        const cellPct = (cellProb * 100).toFixed(2);
                        const isRank1 = p1 && p1.homeScore === hScore && p1.awayScore === aScore;
                        const isHovered = hoveredCell && hoveredCell.h === hScore && hoveredCell.a === aScore;

                        return (
                          <div
                            key={aScore}
                            id={`matrix-cell-${hScore}-${aScore}`}
                            style={getCellBgColor(cellProb)}
                            onMouseEnter={() => setHoveredCell({ h: hScore, a: aScore, prob: cellProb })}
                            onMouseLeave={() => setHoveredCell(null)}
                            className={`rounded-lg border flex flex-col items-center justify-center p-1 cursor-crosshair transition-all duration-150 relative group select-none ${
                              isRank1
                                ? "ring-2 ring-[#00FF87] ring-offset-2 ring-offset-[#111622] z-10"
                                : "hover:scale-[1.03] hover:shadow-lg hover:shadow-[#00FF87]/10"
                            }`}
                          >
                            <span className="text-[12px] font-mono font-black text-white hover:scale-105 transition-transform">
                              {cellPct}%
                            </span>
                            <span className="text-[9px] font-mono text-gray-400 mt-0.5 opacity-50 group-hover:opacity-100 transition-opacity">
                              {hScore}-{aScore}
                            </span>

                            {/* Small Star Indicator for Rank 1 */}
                            {isRank1 && (
                              <div className="absolute top-0.5 right-1 rounded-full bg-[#00FF87] w-1.5 h-1.5 shadow-[0_0_4px_#00FF87]" title="第一顺位比分预测" />
                            )}
                          </div>
                        );
                      })}
                    </div>
                  </div>
                ))}
              </div>

            </div>
          </div>
        </div>

        {/* Hover Inspector Box */}
        <div className="mt-4 pt-4 border-t border-[#232a3b]/60 flex items-center justify-between min-h-[44px]">
          {hoveredCell ? (
            <div className="flex flex-wrap items-center gap-x-4 gap-y-1 text-xs text-gray-300 animate-fadeIn font-sans">
              <span className="font-semibold text-white bg-[#222c42] px-2.5 py-1 rounded border border-[#2d3a56]">
                🔬 浮动数据探针 (Inspector)
              </span>
              <span>
                预测比分: <strong className="font-mono text-[#00E5FF]">{hoveredCell.h} - {hoveredCell.a}</strong>
              </span>
              <span>
                联合概率: <strong className="font-mono text-[#00FF87]">{(hoveredCell.prob * 100).toFixed(4)}%</strong>
              </span>
              <span>
                主队进球分布: <strong className="font-mono text-gray-400">P({hoveredCell.h}) = {(result.matrix[hoveredCell.h].reduce((sum, item) => sum + item, 0) * 100).toFixed(1)}%</strong>
              </span>
              <span>
                客队进球分布: <strong className="font-mono text-gray-400">P({hoveredCell.a}) = {(Array.from({ length: 6 }).map((_, r) => result.matrix[r][hoveredCell.a]).reduce((sum, item) => sum + item, 0) * 100).toFixed(1)}%</strong>
              </span>
            </div>
          ) : (
            <div className="text-xs text-gray-500 font-sans italic flex items-center space-x-1.5">
              <Info className="w-3.5 h-3.5" />
              <span>将鼠标悬停在矩阵网格上，可以查看特定比分的深度概率探针指标。超过 5 球的超大比分概率极低，因对决策影响甚微未作显示。</span>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
