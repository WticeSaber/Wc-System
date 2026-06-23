# Changelog

All notable changes to the WC Predictor project are documented here.
Breaking changes are marked with **[BREAKING]**.

---

## Phase 4b — 2026-06-23

### Frontend: API Integration & UI Enhancements

**New files**
- `frontend/src/api.ts` — HTTP client for the Go backend. Maps backend `snake_case` JSON to TypeScript `camelCase`.

**Modified files**
- `frontend/src/types.ts` — Added `useDeepSeek: boolean` to `MatchParams`; added `deepseekModifier`, `eloSource`, `dataFreshnessAt` to `PredictionResult`; added `TeamProfile` interface.
- `frontend/src/App.tsx` — `handleRunPrediction` is now async; calls `fetchPrediction`, falls back to `runEloPoissonPrediction` on network failure; shows yellow warning banner on fallback; checks `/api/health` at startup.
- `frontend/src/utils/predictor.ts` — `runEloPoissonPrediction` now returns `deepseekModifier: null`, `eloSource: null`, `dataFreshnessAt: null` for backward compatibility with the updated `PredictionResult` type.
- `frontend/src/components/Sidebar.tsx` — Added DeepSeek AI toggle switch in the advanced panel.
- `frontend/src/components/Dashboard.tsx` — Added data source + freshness badge; added DeepSeek modifier badge when `deepseekModifier != null`.

**API changes**: None (frontend-only).

**Known limitations**
- `GET /api/teams` response is fetched at startup but the team dropdown still uses the static `WORLD_CUP_TEAMS` data; live Elo values are only reflected in the prediction calculation, not the team selector UI. This is a future enhancement.

---

## Phase 4a — 2026-06-23

### Backend: HTTP Server

**New files**
- `internal/server/server.go` — `net/http` server, route registration, graceful shutdown.
- `internal/server/handlers.go` — Request handlers for `/api/health`, `/api/teams`, `/api/predict`, `/api/data/status`, `/api/data/refresh`.
- `internal/server/middleware.go` — CORS, JSON Content-Type, request logging middleware chain.
- `cmd/serve.go` — `predictor serve` Cobra command.

**API changes**: All endpoints introduced in this phase (no prior version).

---

## Phase 3 — 2026-06-23

### Backend: DeepSeek Semantic Layer

**New files**
- `internal/semantic/prompt.go` — Go template for the DeepSeek momentum modifier prompt. Formats recent form as `W(2-1 vs France on 2024-07-09)`.
- `internal/semantic/deepseek.go` — OpenAI-compatible DeepSeek API client. Returns float in `[-0.2, +0.2]`.

**Integration point**: `internal/engine/calculator.go` Stage 3 — modifier applied after `(1+β)` micro-operation, before `λ` clamping.

**Known limitations**
- DeepSeek requires valid `DEEPSEEK_API_KEY` env var. Without it, `use_deepseek` is silently ignored.
- 10-second API timeout. On timeout, the prediction proceeds without the modifier (non-fatal).

---

## Phase 2 — 2026-06-23

### Backend: Elo-Poisson Engine + Cobra CLI

**New files**
- `internal/engine/elo.go` — `EloToDeltaGoals()`: Elo delta → net goal correction. Slope `k=0.0011` (configurable via `ELO_SLOPE` env).
- `internal/engine/stats.go` — `ComputeMultipliers()`, `BaseLambdas()`, `clampLambda()`.
- `internal/engine/poisson.go` — `BuildMatrix()`, `ExtractOutcomeProbs()`, `TopNPredictions()`, `CheckAlertThreshold()` using `gonum/stat/distuv`.
- `internal/engine/calculator.go` — `Calculate()`: orchestrates the full 3-stage pipeline. Exports `ErrModifierOutOfRange`, `ErrTeamNotFound`.
- `cmd/root.go`, `cmd/calc.go`, `cmd/fetch.go` — Cobra CLI commands.

**Verification**: `DATA_SOURCE=mock go run . calc --home ARG --away FRA` runs successfully with colorized terminal output.

---

## Phase 1 — 2026-06-23

### Backend: Adapter Layer + Cache

**New files**
- `internal/adapter/normalization.go` — `canonicalNames` map, `Normalize()`, `CanonicalFromID()`, `AllTeamIDs()`, `AllCanonicalByID()`.
- `internal/adapter/csv_results.go` — Fetches `martj42/international_results` CSV; computes last-10-game attack/defense stats.
- `internal/adapter/elo_ratings.go` — Fetches `JGravier/soccer-elo` CSV; parses Elo ratings by team name.
- `internal/adapter/worldbank.go` — Fetches GDP data from World Bank Open Data API (concurrent goroutines).
- `internal/adapter/wikimedia.go` — Fetches Wikipedia article summaries for each national team.
- `internal/adapter/fm_csv.go` — Parses optional Football Manager squad export CSV.
- `internal/cache/store.go` — `DataStore` with `sync.RWMutex`, TTL-based background refresh, `Initialize()`, `StartBackgroundRefresh()`.
- `internal/models/team.go`, `match.go`, `prediction.go` — Shared data models.

**Data source priority**: JGravier Elo CSV (primary) → `mock/elo_ratings.csv` (fallback with `DATA_SOURCE=mock`).

---

## Phase 0 — 2026-06-23

### Infrastructure Setup

**New files**
- `go.mod` — `module wc-predictor`, `go 1.22`, dependencies: cobra, gonum, go-openai, goquery, fatih/color.
- `go.sum` — Locked dependency checksums.
- `main.go` — Entry point delegating to `cmd.Execute()`.
- `Dockerfile` — Multi-stage build: `golang:1.22-alpine` builder → `alpine:latest` runtime.
- `docker-compose.yml` — `backend` service on port 8080 with FM CSV volume mount.
- `.env.example` — Template with all configurable environment variables.
- `mock/results.csv` — ~160 rows of 2024 international match results for offline testing.
- `mock/elo_ratings.csv` — Elo ratings for 24 World Cup 2026 teams.
- `data/.gitignore` — Excludes FM CSV and user data files from git.
- `frontend/vite.config.ts` — Added `/api` proxy to `http://localhost:8080` for dev mode.
