/**
 * @license
 * SPDX-License-Identifier: Apache-2.0
 */

import React, { useState, useRef, useEffect } from "react";
import { Team, MatchParams } from "../types";
import { findTeam } from "../data/teamsRegistry";
import { Search, ChevronDown, Sliders, Play, RotateCcw, HelpCircle, BrainCircuit } from "lucide-react";

interface SidebarProps {
  teams: Team[];
  params: MatchParams;
  onChange: (newParams: MatchParams) => void;
  onRunPredict: () => void;
  isPredicting: boolean;
}

export function Sidebar({ teams, params, onChange, onRunPredict, isPredicting }: SidebarProps) {
  // Dropdown states
  const [homeSearch, setHomeSearch] = useState("");
  const [awaySearch, setAwaySearch] = useState("");
  const [homeOpen, setHomeOpen] = useState(false);
  const [awayOpen, setAwayOpen] = useState(false);
  const [accordionOpen, setAccordionOpen] = useState(false);

  const homeRef = useRef<HTMLDivElement>(null);
  const awayRef = useRef<HTMLDivElement>(null);

  // Close dropdowns on outside click
  useEffect(() => {
    function handleClickOutside(event: MouseEvent) {
      if (homeRef.current && !homeRef.current.contains(event.target as Node)) {
        setHomeOpen(false);
      }
      if (awayRef.current && !awayRef.current.contains(event.target as Node)) {
        setAwayOpen(false);
      }
    }
    document.addEventListener("mousedown", handleClickOutside);
    return () => document.removeEventListener("mousedown", handleClickOutside);
  }, []);

  const homeTeam = findTeam(teams, params.homeTeamId);
  const awayTeam = findTeam(teams, params.awayTeamId);

  const filteredHomeTeams = teams.filter(
    (t) =>
      t.id !== params.awayTeamId &&
      (t.name.toLowerCase().includes(homeSearch.toLowerCase()) ||
        t.nameEn.toLowerCase().includes(homeSearch.toLowerCase()) ||
        t.id.toLowerCase().includes(homeSearch.toLowerCase()))
  );

  const filteredAwayTeams = teams.filter(
    (t) =>
      t.id !== params.homeTeamId &&
      (t.name.toLowerCase().includes(awaySearch.toLowerCase()) ||
        t.nameEn.toLowerCase().includes(awaySearch.toLowerCase()) ||
        t.id.toLowerCase().includes(awaySearch.toLowerCase()))
  );

  const handleSelectHome = (teamId: string) => {
    onChange({
      ...params,
      homeTeamId: teamId,
      // Reset customElo value so it fetches team default default Elo or lets the user know
      homeElo: "",
    });
    setHomeOpen(false);
    setHomeSearch("");
  };

  const handleSelectAway = (teamId: string) => {
    onChange({
      ...params,
      awayTeamId: teamId,
      awayElo: "",
    });
    setAwayOpen(false);
    setAwaySearch("");
  };

  return (
    <div className="bg-[#161a23] border border-[#232a3b] rounded-2xl p-6 shadow-2xl flex flex-col justify-between space-y-6 h-full text-gray-100">
      <div className="space-y-6">
        {/* Title */}
        <div className="border-b border-[#232a3b] pb-4">
          <div className="flex items-center space-x-2">
            <span className="text-2xl">⚽</span>
            <h2 className="text-lg font-bold font-sans tracking-tight text-white uppercase sm:text-base md:text-lg">
              微操控制台 <span className="text-xs font-mono text-[#00E5FF] block mt-0.5 font-normal tracking-widest leading-none">MICRO CONTROL CONSOLE</span>
            </h2>
          </div>
          <p className="text-xs text-gray-400 mt-1.5 leading-relaxed font-sans">
            微调两队核心评级及攻击/防守主观微操干预值，启动实时泊松矩阵运算。
          </p>
        </div>

        {/* Home & Away Team Selection */}
        <div className="space-y-4">
          {/* Home Team Dropdown */}
          <div className="relative" ref={homeRef} id="home-team-picker">
            <label className="block text-xs font-semibold text-gray-300 font-sans tracking-wider mb-1.5 uppercase">
              主队 (Home Team)
            </label>
            <button
              type="button"
              id="home-team-dropdown-btn"
              onClick={() => {
                setHomeOpen(!homeOpen);
                setAwayOpen(false);
              }}
              className="w-full flex items-center justify-between bg-[#1f2638] hover:bg-[#28324a] border border-[#2d3a56] focus:border-[#00FF87] hover:border-gray-500 rounded-xl px-4 py-3 text-left transition-all duration-200 shadow-md group cursor-pointer"
            >
              <div className="flex items-center space-x-3.5">
                <div className="w-10 h-10 rounded-full bg-[#121622] border-2 border-[#2d3a56] group-hover:border-[#00FF87] overflow-hidden flex items-center justify-center shadow-inner shrink-0 select-none transition-colors">
                  <img
                    src={`https://flagcdn.com/w80/${homeTeam.flagCode}.png`}
                    alt={`${homeTeam.name} flag`}
                    className="w-full h-full object-cover scale-110"
                    referrerPolicy="no-referrer"
                  />
                </div>
                <div>
                  <div className="text-sm font-bold text-white leading-tight flex items-center">
                    {homeTeam.name}
                    <span className="ml-1.5 text-[10px] font-mono text-gray-400 font-normal uppercase bg-[#161a23] px-1.5 py-0.5 rounded border border-[#252f44]">
                      {homeTeam.id}{homeTeam.group ? ` · ${homeTeam.group}` : ""}
                    </span>
                  </div>
                  <div className="text-[11px] text-gray-400 font-mono mt-0.5">
                    基础 Elo: {params.homeElo || homeTeam.defaultElo}
                  </div>
                </div>
              </div>
              <ChevronDown className={`w-4 h-4 text-gray-400 group-hover:text-white transition-transform duration-200 ${homeOpen ? "transform rotate-180" : ""}`} />
            </button>

            {/* Dropdown popup */}
            {homeOpen && (
              <div className="absolute left-0 mt-2 w-full bg-[#1e2536] border border-[#2d3a56] rounded-xl shadow-2xl z-50 overflow-hidden transform duration-200 animate-fadeIn">
                <div className="p-2 border-b border-[#2d3a56] bg-[#161d2b] flex items-center">
                  <Search className="w-4 h-4 text-gray-400 mr-2 shrink-0" />
                  <input
                    type="text"
                    id="home-search-input"
                    value={homeSearch}
                    onChange={(e) => setHomeSearch(e.target.value)}
                    placeholder="输入国名搜索..."
                    className="w-full bg-transparent text-sm text-white focus:outline-none placeholder-gray-500 py-1"
                    autoFocus
                  />
                </div>
                <ul className="max-h-60 overflow-y-auto divide-y divide-[#252f44] scrollbar-thin scrollbar-thumb-gray-800">
                  {filteredHomeTeams.length > 0 ? (
                    filteredHomeTeams.map((t) => (
                      <li key={t.id}>
                        <button
                          type="button"
                          id={`home-select-${t.id}`}
                          onClick={() => handleSelectHome(t.id)}
                          className="w-full flex items-center justify-between px-4 py-2.5 hover:bg-[#28324a] text-left transition-all duration-150 cursor-pointer"
                        >
                          <div className="flex items-center space-x-3.5">
                            <div className="w-8 h-8 rounded-full bg-[#111522] border border-[#2d3a56] overflow-hidden flex items-center justify-center shadow-inner shrink-0 select-none">
                              <img
                                src={`https://flagcdn.com/w40/${t.flagCode}.png`}
                                alt={`${t.name} flag`}
                                className="w-full h-full object-cover scale-110"
                                referrerPolicy="no-referrer"
                              />
                            </div>
                            <div>
                              <span className="text-sm font-semibold text-white block">{t.name}</span>
                              <span className="text-[10px] text-gray-400 uppercase font-mono block">{t.nameEn}</span>
                            </div>
                          </div>
                          <span className="text-xs font-mono text-gray-400 bg-[#161a23] px-2 py-0.5 rounded border border-[#2d3a56]">
                            Elo {t.defaultElo}
                          </span>
                        </button>
                      </li>
                    ))
                  ) : (
                    <li className="px-4 py-3 text-center text-xs text-gray-500">未找到相关队伍</li>
                  )}
                </ul>
              </div>
            )}
          </div>

          {/* Away Team Dropdown */}
          <div className="relative" ref={awayRef} id="away-team-picker">
            <label className="block text-xs font-semibold text-gray-300 font-sans tracking-wider mb-1.5 uppercase">
              客队 (Away Team)
            </label>
            <button
              type="button"
              id="away-team-dropdown-btn"
              onClick={() => {
                setAwayOpen(!awayOpen);
                setHomeOpen(false);
              }}
              className="w-full flex items-center justify-between bg-[#1f2638] hover:bg-[#28324a] border border-[#2d3a56] focus:border-[#00FF87] hover:border-gray-500 rounded-xl px-4 py-3 text-left transition-all duration-200 shadow-md group cursor-pointer"
            >
              <div className="flex items-center space-x-3.5">
                <div className="w-10 h-10 rounded-full bg-[#121622] border-2 border-[#2d3a56] group-hover:border-[#00E5FF] overflow-hidden flex items-center justify-center shadow-inner shrink-0 select-none transition-colors">
                  <img
                    src={`https://flagcdn.com/w80/${awayTeam.flagCode}.png`}
                    alt={`${awayTeam.name} flag`}
                    className="w-full h-full object-cover scale-110"
                    referrerPolicy="no-referrer"
                  />
                </div>
                <div>
                  <div className="text-sm font-bold text-white leading-tight flex items-center">
                    {awayTeam.name}
                    <span className="ml-1.5 text-[10px] font-mono text-gray-400 font-normal uppercase bg-[#161a23] px-1.5 py-0.5 rounded border border-[#252f44]">
                      {awayTeam.id}{awayTeam.group ? ` · ${awayTeam.group}` : ""}
                    </span>
                  </div>
                  <div className="text-[11px] text-gray-400 font-mono mt-0.5">
                    基础 Elo: {params.awayElo || awayTeam.defaultElo}
                  </div>
                </div>
              </div>
              <ChevronDown className={`w-4 h-4 text-gray-400 group-hover:text-white transition-transform duration-200 ${awayOpen ? "transform rotate-180" : ""}`} />
            </button>

            {/* Dropdown popup */}
            {awayOpen && (
              <div className="absolute left-0 mt-2 w-full bg-[#1e2536] border border-[#2d3a56] rounded-xl shadow-2xl z-50 overflow-hidden transform duration-200 animate-fadeIn">
                <div className="p-2 border-b border-[#2d3a56] bg-[#161d2b] flex items-center">
                  <Search className="w-4 h-4 text-gray-400 mr-2 shrink-0" />
                  <input
                    type="text"
                    id="away-search-input"
                    value={awaySearch}
                    onChange={(e) => setAwaySearch(e.target.value)}
                    placeholder="输入国名搜索..."
                    className="w-full bg-transparent text-sm text-white focus:outline-none placeholder-gray-500 py-1"
                    autoFocus
                  />
                </div>
                <ul className="max-h-60 overflow-y-auto divide-y divide-[#252f44] scrollbar-thin scrollbar-thumb-gray-800">
                  {filteredAwayTeams.length > 0 ? (
                    filteredAwayTeams.map((t) => (
                      <li key={t.id}>
                        <button
                          type="button"
                          id={`away-select-${t.id}`}
                          onClick={() => handleSelectAway(t.id)}
                          className="w-full flex items-center justify-between px-4 py-2.5 hover:bg-[#28324a] text-left transition-all duration-150 cursor-pointer"
                        >
                          <div className="flex items-center space-x-3.5">
                            <div className="w-8 h-8 rounded-full bg-[#111522] border border-[#2d3a56] overflow-hidden flex items-center justify-center shadow-inner shrink-0 select-none">
                              <img
                                src={`https://flagcdn.com/w40/${t.flagCode}.png`}
                                alt={`${t.name} flag`}
                                className="w-full h-full object-cover scale-110"
                                referrerPolicy="no-referrer"
                              />
                            </div>
                            <div>
                              <span className="text-sm font-semibold text-white block">{t.name}</span>
                              <span className="text-[10px] text-gray-400 uppercase font-mono block">{t.nameEn}</span>
                            </div>
                          </div>
                          <span className="text-xs font-mono text-gray-400 bg-[#161a23] px-2 py-0.5 rounded border border-[#2d3a56]">
                            Elo {t.defaultElo}
                          </span>
                        </button>
                      </li>
                    ))
                  ) : (
                    <li className="px-4 py-3 text-center text-xs text-gray-500">未找到相关队伍</li>
                  )}
                </ul>
              </div>
            )}
          </div>
        </div>

        {/* Macro Benchmark Expected Goals */}
        <div className="bg-[#1a202f] p-4 rounded-xl border border-[#26324d] space-y-2">
          <div className="flex items-center justify-between">
            <label className="text-xs font-semibold text-gray-300 font-sans tracking-wide uppercase flex items-center">
              赛事场均进球期望 (Avg Goals)
              <span className="group relative ml-1.5 cursor-pointer">
                <HelpCircle className="w-3.5 h-3.5 text-gray-400 hover:text-white" />
                <span className="absolute left-1/2 -translate-x-1/2 bottom-full mb-2 w-48 bg-[#0b111e] text-[10px] text-gray-300 p-2 rounded-lg border border-[#2c374e] opacity-0 group-hover:opacity-100 transition-opacity duration-200 pointer-events-none z-50 text-center leading-normal">
                  国际足联世界杯场均进球的标准平均参数（推荐值：2.3 至 2.8）
                </span>
              </span>
            </label>
            <span className="text-xs font-mono text-[#00E5FF] font-bold bg-[#111622] px-2 py-0.5 rounded border border-[#242e44]">
              {params.avgGoals}
            </span>
          </div>
          <input
            type="number"
            id="avg-goals-input"
            value={params.avgGoals}
            min="0.5"
            max="6.0"
            step="0.1"
            onChange={(e) => {
              const val = parseFloat(e.target.value);
              onChange({ ...params, avgGoals: isNaN(val) ? 2.5 : val });
            }}
            className="w-full bg-[#111622] text-sm text-white font-mono border border-[#2e3b56] focus:border-[#00FF87] focus:outline-none rounded-lg px-3 py-2 transition-all"
          />
        </div>

        {/* Collapsible Advanced Configuration Accordion */}
        <div className="border border-[#232a3b] rounded-xl overflow-hidden bg-[#161a23]">
          <button
            type="button"
            id="advanced-accordion-trigger"
            onClick={() => setAccordionOpen(!accordionOpen)}
            className="w-full flex items-center justify-between p-4 bg-[#1f2638] hover:bg-[#28324a] text-left transition-colors cursor-pointer"
          >
            <span className="text-xs font-bold text-gray-200 flex items-center font-sans tracking-wider uppercase">
              <Sliders className="w-3.5 h-3.5 text-[#00E5FF] mr-2" />
              高级干预面板 (Advanced Config)
            </span>
            <ChevronDown className={`w-4 h-4 text-gray-400 transition-transform duration-200 ${accordionOpen ? "transform rotate-180" : ""}`} />
          </button>

          {/* Accordion Content Panel */}
          <div className={`transition-all duration-300 ease-in-out ${accordionOpen ? "max-h-[500px] border-t border-[#232a3b] p-4 opacity-100" : "max-h-0 opacity-0 overflow-hidden"}`}>
            <div className="space-y-4">
              {/* Manual Elo Overrides */}
              <div className="grid grid-cols-2 gap-3">
                <div className="space-y-1">
                  <label className="text-[10px] uppercase tracking-wider font-mono text-gray-400 block">
                    主队 Elo 分数
                  </label>
                  <input
                    type="number"
                    id="home-elo-override"
                    value={params.homeElo}
                    placeholder="留空则自动抓取最新"
                    onChange={(e) => {
                      const val = e.target.value === "" ? "" : parseInt(e.target.value);
                      onChange({ ...params, homeElo: val });
                    }}
                    className="w-full bg-[#111622] text-xs font-mono text-white placeholder-gray-600 border border-[#2d3a56] focus:border-[#00FF87] focus:outline-none rounded-lg px-2.5 py-1.5 transition-all text-center"
                  />
                </div>
                <div className="space-y-1">
                  <label className="text-[10px] uppercase tracking-wider font-mono text-gray-400 block">
                    客队 Elo 分数
                  </label>
                  <input
                    type="number"
                    id="away-elo-override"
                    value={params.awayElo}
                    placeholder="留空则自动抓取最新"
                    onChange={(e) => {
                      const val = e.target.value === "" ? "" : parseInt(e.target.value);
                      onChange({ ...params, awayElo: val });
                    }}
                    className="w-full bg-[#111622] text-xs font-mono text-white placeholder-gray-600 border border-[#2d3a56] focus:border-[#00FF87] focus:outline-none rounded-lg px-2.5 py-1.5 transition-all text-center"
                  />
                </div>
              </div>

              {/* Subjective Intervention Modifiers */}
              <div className="space-y-4 pt-1">
                {/* Home Mod Slider */}
                <div className="space-y-1.5">
                  <div className="flex items-center justify-between text-[11px] font-sans">
                    <span className="text-gray-300 font-medium">主队主观微操干预 (Home Mod)</span>
                    <span className={`font-mono font-bold px-1.5 py-0.5 rounded text-[10px] ${params.homeMod > 0 ? "text-[#00FF87] bg-[#112a20]" : params.homeMod < 0 ? "text-[#FF4A6B] bg-[#2a131b]" : "text-gray-400 bg-[#161a23]"}`}>
                      {params.homeMod > 0 ? `+${params.homeMod.toFixed(2)}` : params.homeMod.toFixed(2)}
                    </span>
                  </div>
                  <input
                    type="range"
                    id="home-mod-slider"
                    min="-0.5"
                    max="0.5"
                    step="0.05"
                    value={params.homeMod}
                    onChange={(e) => onChange({ ...params, homeMod: parseFloat(e.target.value) })}
                    className="w-full accent-[#00FF87] bg-[#1a202f] h-1.5 rounded-lg appearance-none cursor-pointer"
                  />
                  <div className="flex justify-between text-[8px] font-mono text-gray-500 px-1">
                    <span>-0.5 (极度降解)</span>
                    <span>0.0 (中性白板)</span>
                    <span>+0.5 (完全增压)</span>
                  </div>
                </div>

                {/* Away Mod Slider */}
                <div className="space-y-1.5">
                  <div className="flex items-center justify-between text-[11px] font-sans">
                    <span className="text-gray-300 font-medium">客队主观干预 (Away Mod)</span>
                    <span className={`font-mono font-bold px-1.5 py-0.5 rounded text-[10px] ${params.awayMod > 0 ? "text-[#00FF87] bg-[#112a20]" : params.awayMod < 0 ? "text-[#FF4A6B] bg-[#2a131b]" : "text-gray-400 bg-[#161a23]"}`}>
                      {params.awayMod > 0 ? `+${params.awayMod.toFixed(2)}` : params.awayMod.toFixed(2)}
                    </span>
                  </div>
                  <input
                    type="range"
                    id="away-mod-slider"
                    min="-0.5"
                    max="0.5"
                    step="0.05"
                    value={params.awayMod}
                    onChange={(e) => onChange({ ...params, awayMod: parseFloat(e.target.value) })}
                    className="w-full accent-[#00E5FF] bg-[#1a202f] h-1.5 rounded-lg appearance-none cursor-pointer"
                  />
                  <div className="flex justify-between text-[8px] font-mono text-gray-500 px-1">
                    <span>-0.5 (极度降解)</span>
                    <span>0.0 (中性白板)</span>
                    <span>+0.5 (完全增压)</span>
                  </div>
                </div>

                {/* Quick resets */}
                <button
                  type="button"
                  id="reset-modifiers-btn"
                  onClick={() => onChange({ ...params, homeMod: 0.0, awayMod: 0.0, homeElo: "", awayElo: "" })}
                  className="w-full flex items-center justify-center space-x-1 py-1.5 text-[10px] font-semibold text-gray-400 hover:text-white bg-[#111622] hover:bg-[#1c2438] hover:border-gray-500 border border-[#2d3a56] rounded-lg transition-all cursor-pointer"
                >
                  <RotateCcw className="w-3 h-3" />
                  <span>复位所有变数 (Reset Modifiers)</span>
                </button>

                {/* DeepSeek AI Semantic Modifier Toggle */}
                <div className="pt-2 border-t border-[#232a3b]/60">
                  <div className="flex items-center justify-between">
                    <div className="flex items-center space-x-2">
                      <BrainCircuit className="w-3.5 h-3.5 text-[#00E5FF]" />
                      <span className="text-[11px] font-semibold text-gray-300 font-sans">
                        DeepSeek AI 情报修正
                      </span>
                      <span className="group relative cursor-pointer">
                        <HelpCircle className="w-3 h-3 text-gray-500 hover:text-gray-300" />
                        <span className="absolute left-1/2 -translate-x-1/2 bottom-full mb-2 w-52 bg-[#0b111e] text-[10px] text-gray-300 p-2 rounded-lg border border-[#2c374e] opacity-0 group-hover:opacity-100 transition-opacity duration-200 pointer-events-none z-50 text-center leading-normal">
                          调用 DeepSeek 分析近期比赛状态，生成 [-0.2, +0.2] 范围内的情报修正系数叠加至期望进球计算。需配置 DEEPSEEK_API_KEY。
                        </span>
                      </span>
                    </div>
                    <button
                      type="button"
                      id="deepseek-toggle"
                      onClick={() => onChange({ ...params, useDeepSeek: !params.useDeepSeek })}
                      className={`relative inline-flex h-5 w-9 items-center rounded-full transition-colors duration-200 cursor-pointer focus:outline-none ${
                        params.useDeepSeek ? "bg-[#00E5FF]" : "bg-[#2d3a56]"
                      }`}
                      role="switch"
                      aria-checked={params.useDeepSeek}
                    >
                      <span
                        className={`inline-block h-3.5 w-3.5 transform rounded-full bg-white shadow transition-transform duration-200 ${
                          params.useDeepSeek ? "translate-x-4" : "translate-x-0.5"
                        }`}
                      />
                    </button>
                  </div>
                  {params.useDeepSeek && (
                    <p className="text-[9px] text-[#00E5FF] font-mono mt-1.5 leading-tight">
                      AI 修正已启用 — 点击「运行矩阵运算」时将调用 DeepSeek API
                    </p>
                  )}
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      {/* Main operational action trigger */}
      <div className="pt-4 border-t border-[#232a3b]">
        <button
          type="button"
          id="run-prediction-action-btn"
          disabled={isPredicting}
          onClick={onRunPredict}
          className={`w-full py-3.5 px-4 rounded-xl font-bold flex items-center justify-center space-x-2 transition-all text-sm uppercase tracking-wider relative overflow-hidden group cursor-pointer ${
            isPredicting
              ? "bg-[#252f44] text-gray-500 cursor-not-allowed border border-transparent shadow shadow-none"
              : "bg-[#00FF87] hover:bg-[#17ff93] active:scale-95 text-black font-extrabold shadow-[0_0_20px_rgba(0,255,135,0.35)] hover:shadow-[0_0_25px_rgba(0,255,135,0.5)]"
          }`}
        >
          {isPredicting ? (
            <div className="flex items-center space-x-2">
              <svg className="animate-spin h-4 w-4 text-gray-400" fill="none" viewBox="0 0 24 24">
                <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" />
                <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
              </svg>
              <span className="font-mono text-xs tracking-widest text-gray-300">MATRIX RE-COMPUTING...</span>
            </div>
          ) : (
            <>
              <Play className="w-4 h-4 fill-black text-black group-hover:scale-110 transition-transform" />
              <span className="font-sans">运行矩阵运算 (Run Prediction)</span>
            </>
          )}
        </button>
      </div>
    </div>
  );
}
