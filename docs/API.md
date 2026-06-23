# WC Predictor REST API Documentation

## Overview

The Go backend exposes a JSON REST API on port `8080` (configurable via `PORT` env var).
All endpoints are prefixed with `/api`. The frontend Vite dev server proxies `/api/*` requests
to `http://localhost:8080` automatically during development.

---

## Endpoints

### `GET /api/health`

Health check. Returns the server status and time of last data refresh.

**Response 200**
```json
{
  "status": "ok",
  "last_refresh": "2026-06-23T20:00:00Z"
}
```

---

### `GET /api/teams`

Returns all supported national teams with their current stats from the data store.

**Response 200** — array of `TeamProfile`
```json
[
  {
    "id": "ARG",
    "name_en": "Argentina",
    "name_cn": "阿根廷",
    "flag_code": "ar",
    "group": "J",
    "is_host": false,
    "elo": 2141.0,
    "avg_goals_scored": 2.3,
    "avg_goals_conceded": 0.7
  }
]
```

> When the backend hasn't yet fetched live data, `elo`, `avg_goals_scored`, and
> `avg_goals_conceded` may be `0`. The frontend falls back to `frontend/src/data/teams.ts`
> static defaults in this case.

---

### `POST /api/predict`

Runs the complete Elo-Poisson prediction pipeline for a single match.

**Request body**
```json
{
  "home_team": "ARG",
  "away_team": "FRA",
  "avg_goals": 2.5,
  "home_elo": null,
  "away_elo": null,
  "home_mod": 0.0,
  "away_mod": 0.0,
  "use_deepseek": false,
  "is_neutral": true
}
```

| Field | Type | Required | Description |
|---|---|---|---|
| `home_team` | string | ✓ | Team ID (e.g. `"ARG"`) or canonical name (e.g. `"Argentina"`) |
| `away_team` | string | ✓ | Same as above |
| `avg_goals` | float | — | Match baseline expected goals. Default: `2.5` |
| `home_elo` | float\|null | — | Manual Elo override for home team. `null` = auto-fetch |
| `away_elo` | float\|null | — | Manual Elo override for away team. `null` = auto-fetch |
| `home_mod` | float | — | Subjective home modifier. Range: `[-0.5, +0.5]`. Default: `0.0` |
| `away_mod` | float | — | Subjective away modifier. Range: `[-0.5, +0.5]`. Default: `0.0` |
| `use_deepseek` | bool | — | Enable DeepSeek semantic momentum modifier. Requires `DEEPSEEK_API_KEY`. Default: `false` |
| `is_neutral` | bool | — | Whether the match is at a neutral venue. Default: `true` (World Cup) |

**Response 200** — `PredictionResult`
```json
{
  "home_expected_goals": 1.42,
  "away_expected_goals": 1.18,
  "home_win_prob": 0.451,
  "draw_prob": 0.267,
  "away_win_prob": 0.282,
  "matrix": [
    [0.0763, 0.0901, ...],
    ...
  ],
  "top_predictions": [
    { "home_score": 1, "away_score": 1, "probability": 0.112 },
    { "home_score": 2, "away_score": 1, "probability": 0.098 },
    { "home_score": 1, "away_score": 0, "probability": 0.096 }
  ],
  "alert_triggered": false,
  "alert_message": "",
  "deepseek_modifier": 0.05,
  "elo_source": "github_mirror",
  "data_freshness_at": "2026-06-23T20:00:00Z"
}
```

| Field | Type | Description |
|---|---|---|
| `home_expected_goals` | float | Final λ for home team after all corrections |
| `away_expected_goals` | float | Final λ for away team after all corrections |
| `home_win_prob` | float | Normalized probability of home team winning |
| `draw_prob` | float | Normalized probability of a draw |
| `away_win_prob` | float | Normalized probability of away team winning |
| `matrix` | `float[6][6]` | Joint probability matrix: `matrix[homeGoals][awayGoals]` |
| `top_predictions` | array | Top 3 most likely scorelines, sorted by probability descending |
| `alert_triggered` | bool | `true` if P(0-0) ≥ 15% |
| `alert_message` | string | Human-readable warning (Chinese) when alert fires |
| `deepseek_modifier` | float\|null | The AI momentum modifier applied. `null` if DeepSeek was not used |
| `elo_source` | string | `"github_mirror"` or `"manual_override"` |
| `data_freshness_at` | datetime | ISO 8601 timestamp of last data store refresh |

**Error Responses**

| Status | Condition |
|---|---|
| `400` | `home_mod` or `away_mod` out of range `[-0.5, +0.5]`, or invalid JSON |
| `404` | Team name not found in data store |
| `500` | Internal calculation error |

**Error body**
```json
{ "message": "modifier must be between -0.5 and +0.5: home=0.80 away=0.00", "code": 400 }
```

---

### `GET /api/data/status`

Returns metadata about the current state of the data cache.

**Response 200**
```json
{
  "last_refresh": "2026-06-23T20:00:00Z",
  "team_count": 24
}
```

---

### `POST /api/data/refresh?source=all`

Manually triggers a data refresh. Useful after a match day.

**Query parameters**

| Param | Values | Default |
|---|---|---|
| `source` | `all`, `csv`, `elo`, `worldbank`, `wikimedia`, `fm` | `all` |

**Response 200**
```json
{ "status": "refreshed", "source": "all" }
```

---

## Environment Configuration

| Variable | Default | Description |
|---|---|---|
| `DATA_SOURCE` | `live` | `live` = fetch from GitHub; `mock` = use `mock/` CSV files |
| `PORT` | `8080` | HTTP server port |
| `DEEPSEEK_API_KEY` | — | Optional. Required for `use_deepseek=true` requests |
| `ELO_SLOPE` | `0.0011` | Elo-to-goal linear regression slope (range: 0.0009–0.0014) |
| `FM_CSV_PATH` | — | Optional. Path to Football Manager squad export CSV |
| `DATA_REFRESH_INTERVAL` | `24h` | Background cache refresh interval |

---

## Calculation Pipeline (Summary)

The `/api/predict` endpoint executes the following three-stage pipeline internally:

```
Stage 1: Data fetch
  home_stats ← DataStore.GetTeamStats(home_team)
  away_stats ← DataStore.GetTeamStats(away_team)
  home_elo, away_elo ← manual override || store value

Stage 2: λ computation
  delta_goals = ELO_SLOPE × (home_elo − away_elo)          [neutral venue]
  home_atk = home.avg_goals_scored / avg_goals
  away_def = away.avg_goals_conceded / avg_goals
  λ_base_home = avg_goals × home_atk × away_def
  λ_base_home += delta_goals / 2
  λ_base_away -= delta_goals / 2

Stage 3: Modifiers + Poisson matrix
  λ_final_home = λ_base_home × (1 + home_mod)
  λ_final_away = λ_base_away × (1 + away_mod)
  [if use_deepseek] λ_final_home *= (1 + ai_modifier)
                    λ_final_away *= (1 − ai_modifier)
  λ_final_home = max(0.01, λ_final_home)
  λ_final_away = max(0.01, λ_final_away)

  matrix[i][j] = Poisson(λ_home).Prob(i) × Poisson(λ_away).Prob(j)
  for i,j in [0,5]
```
