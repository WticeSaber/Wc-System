package server

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"wc-predictor/internal/adapter"
	"wc-predictor/internal/cache"
	"wc-predictor/internal/engine"
	"wc-predictor/internal/models"
	"wc-predictor/internal/semantic"
)

// Handlers groups all HTTP handler functions with their shared dependencies.
type Handlers struct {
	store        *cache.DataStore
	deepSeekClient *semantic.DeepSeekClient
}

// NewHandlers creates a Handlers instance.
func NewHandlers(store *cache.DataStore, dsClient *semantic.DeepSeekClient) *Handlers {
	return &Handlers{store: store, deepSeekClient: dsClient}
}

// writeJSON serializes v to JSON and writes it to w with the given status code.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("[server] writeJSON error: %v", err)
	}
}

// writeError writes a standard API error response.
func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, models.APIError{Message: message, Code: status})
}

// Health handles GET /api/health
func (h *Handlers) Health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"status":       "ok",
		"last_refresh": h.store.LastRefresh().Format("2006-01-02T15:04:05Z"),
	})
}

// Teams handles GET /api/teams
func (h *Handlers) Teams(w http.ResponseWriter, r *http.Request) {
	profiles := h.store.AllTeamProfiles()
	writeJSON(w, http.StatusOK, profiles)
}

// DataStatus handles GET /api/data/status
func (h *Handlers) DataStatus(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"last_refresh": h.store.LastRefresh(),
		"team_count":   len(h.store.AllTeamProfiles()),
	})
}

// Predict handles POST /api/predict
func (h *Handlers) Predict(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "only POST is allowed")
		return
	}

	var req models.MatchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body: "+err.Error())
		return
	}

	// Default values.
	if req.AvgGoals <= 0 {
		req.AvgGoals = 2.5
	}
	// World Cup: all games on neutral ground by default.
	req.IsNeutral = true

	// Resolve canonical team names (request may use team IDs like "ARG").
	homeCanonical := resolveTeamName(req.HomeTeam)
	awayCanonical := resolveTeamName(req.AwayTeam)
	req.HomeTeam = homeCanonical
	req.AwayTeam = awayCanonical

	// Build optional semantic function.
	var semanticFn engine.SemanticModifierFunc
	if req.UseDeepSeek && h.deepSeekClient != nil {
		semanticFn = h.deepSeekClient.AsSemanticModifierFunc()
	}

	result, err := engine.Calculate(r.Context(), &req, h.store, semanticFn)
	if err != nil {
		switch {
		case errors.Is(err, engine.ErrModifierOutOfRange):
			writeError(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, engine.ErrTeamNotFound):
			writeError(w, http.StatusNotFound, err.Error())
		default:
			log.Printf("[server] calculate error: %v", err)
			writeError(w, http.StatusInternalServerError, "internal calculation error")
		}
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// FetchSource handles POST /api/data/refresh?source=all
func (h *Handlers) FetchSource(w http.ResponseWriter, r *http.Request) {
	source := r.URL.Query().Get("source")
	if source == "" {
		source = "all"
	}
	if err := h.store.RefreshSource(context.Background(), source); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "refreshed", "source": source})
}

// resolveTeamName converts a team ID (e.g. "ARG") to the canonical name
// used internally. If not recognized, the original string is returned.
func resolveTeamName(nameOrID string) string {
	if canonical, ok := adapter.CanonicalFromID(nameOrID); ok {
		return canonical
	}
	return nameOrID
}
