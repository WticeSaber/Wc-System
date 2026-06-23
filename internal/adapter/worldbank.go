package adapter

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"wc-predictor/internal/models"
	"wc-predictor/internal/teams"
)

const worldBankBaseURL = "https://api.worldbank.org/v2/country/%s/indicator/NY.GDP.MKTP.CD?format=json&mrv=1"

// WorldBankAdapter fetches GDP context data from the World Bank Open Data API.
type WorldBankAdapter struct {
	contexts map[string]models.MacroContext
	mu       sync.RWMutex
}

// NewWorldBankAdapter creates a new World Bank adapter.
func NewWorldBankAdapter() *WorldBankAdapter {
	return &WorldBankAdapter{
		contexts: make(map[string]models.MacroContext),
	}
}

// Fetch concurrently retrieves GDP data for all roster team ISO2 codes.
func (a *WorldBankAdapter) Fetch(ctx context.Context) error {
	reg, err := teams.Global()
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	client := &http.Client{Timeout: 8 * time.Second}

	for _, entry := range reg.All() {
		wg.Add(1)
		go func(id, iso string) {
			defer wg.Done()
			gdp := a.fetchGDP(ctx, client, iso)
			a.mu.Lock()
			a.contexts[id] = models.MacroContext{ISO2: iso, GDP: gdp}
			a.mu.Unlock()
		}(entry.ID, entry.ISO2)
	}

	wg.Wait()
	return nil
}

func (a *WorldBankAdapter) fetchGDP(ctx context.Context, client *http.Client, iso2 string) float64 {
	url := fmt.Sprintf(worldBankBaseURL, iso2)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return 0
	}

	resp, err := client.Do(req)
	if err != nil {
		return 0
	}
	defer resp.Body.Close()

	var raw []json.RawMessage
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil || len(raw) < 2 {
		return 0
	}

	var records []struct {
		Value *float64 `json:"value"`
	}
	if err := json.Unmarshal(raw[1], &records); err != nil || len(records) == 0 {
		return 0
	}
	if records[0].Value == nil {
		return 0
	}
	return *records[0].Value
}

// MergeInto copies macro context into TeamRawStats keyed by canonical name.
func (a *WorldBankAdapter) MergeInto(statsMap map[string]*models.TeamRawStats, idMap map[string]string) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	for teamID, ctx := range a.contexts {
		canonical, ok := idMap[teamID]
		if !ok {
			continue
		}
		if s, ok := statsMap[canonical]; ok {
			s.MacroContext.ISO2 = ctx.ISO2
			s.MacroContext.GDP = ctx.GDP
		}
	}
}
