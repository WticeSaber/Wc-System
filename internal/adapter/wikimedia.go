package adapter

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"wc-predictor/internal/models"
	"wc-predictor/internal/teams"
)

const wikimediaBaseURL = "https://en.wikipedia.org/api/rest_v1/page/summary/%s"

// WikimediaAdapter fetches team article summaries from the Wikipedia REST API.
type WikimediaAdapter struct {
	summaries map[string]string
	mu        sync.RWMutex
}

// NewWikimediaAdapter creates a new Wikimedia adapter.
func NewWikimediaAdapter() *WikimediaAdapter {
	return &WikimediaAdapter{
		summaries: make(map[string]string),
	}
}

// Fetch concurrently retrieves Wikipedia summaries for all roster teams.
func (a *WikimediaAdapter) Fetch(ctx context.Context) error {
	reg, err := teams.Global()
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	client := &http.Client{Timeout: 8 * time.Second}

	for _, entry := range reg.All() {
		if entry.WikiTitle == "" {
			continue
		}
		wg.Add(1)
		go func(id, title string) {
			defer wg.Done()
			summary := a.fetchSummary(ctx, client, title)
			a.mu.Lock()
			a.summaries[id] = summary
			a.mu.Unlock()
		}(entry.ID, entry.WikiTitle)
	}

	wg.Wait()
	return nil
}

func (a *WikimediaAdapter) fetchSummary(ctx context.Context, client *http.Client, wikiTitle string) string {
	reqURL := fmt.Sprintf(wikimediaBaseURL, wikiTitle)
	if !strings.Contains(wikiTitle, "%") {
		reqURL = fmt.Sprintf(wikimediaBaseURL, url.PathEscape(wikiTitle))
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return ""
	}
	req.Header.Set("User-Agent", "wc-predictor/1.0 (contact: predictor@example.com)")

	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		return ""
	}
	defer resp.Body.Close()

	var result struct {
		Extract string `json:"extract"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return ""
	}
	if len(result.Extract) > 500 {
		return result.Extract[:500]
	}
	return result.Extract
}

// MergeInto copies wiki summaries into TeamRawStats keyed by canonical name.
func (a *WikimediaAdapter) MergeInto(statsMap map[string]*models.TeamRawStats, idMap map[string]string) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	for teamID, summary := range a.summaries {
		canonical, ok := idMap[teamID]
		if !ok {
			continue
		}
		if s, ok := statsMap[canonical]; ok {
			s.MacroContext.WikiSummary = summary
		}
	}
}
