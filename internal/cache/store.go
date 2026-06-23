package cache

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"wc-predictor/internal/adapter"
	"wc-predictor/internal/models"
	"wc-predictor/internal/teams"
)

// Config holds DataStore configuration.
type Config struct {
	DataSource string        // "live" or "mock"
	TTL        time.Duration // refresh interval, default 24h
	FMCSVPath  string        // path to optional FM CSV file
}

// DataStore is the central in-memory cache for all adapter data.
// All public methods are safe for concurrent use.
type DataStore struct {
	mu          sync.RWMutex
	teamStats   map[string]*models.TeamRawStats // keyed by canonical name
	lastRefresh time.Time
	cfg         Config

	csvAdapter    *adapter.CSVResultsAdapter
	eloAdapter    *adapter.EloRatingsAdapter
	wbAdapter     *adapter.WorldBankAdapter
	wikiAdapter   *adapter.WikimediaAdapter
	fmAdapter     *adapter.FMCSVAdapter
}

// NewDataStore creates and initializes a DataStore with the given configuration.
func NewDataStore(cfg Config) *DataStore {
	if cfg.TTL == 0 {
		cfg.TTL = 24 * time.Hour
	}
	return &DataStore{
		teamStats:   make(map[string]*models.TeamRawStats),
		cfg:         cfg,
		csvAdapter:  adapter.NewCSVResultsAdapter(cfg.DataSource),
		eloAdapter:  adapter.NewEloRatingsAdapter(cfg.DataSource),
		wbAdapter:   adapter.NewWorldBankAdapter(),
		wikiAdapter: adapter.NewWikimediaAdapter(),
		fmAdapter:   adapter.NewFMCSVAdapter(cfg.FMCSVPath),
	}
}

// Initialize performs the first data fetch synchronously. Call this at startup.
func (s *DataStore) Initialize(ctx context.Context) error {
	return s.refresh(ctx)
}

// StartBackgroundRefresh launches a goroutine that refreshes data on the configured TTL.
// The goroutine stops when ctx is cancelled.
func (s *DataStore) StartBackgroundRefresh(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(s.cfg.TTL)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := s.refresh(ctx); err != nil {
					log.Printf("[cache] background refresh failed: %v", err)
				} else {
					log.Printf("[cache] data refreshed at %s", time.Now().Format(time.RFC3339))
				}
			}
		}
	}()
}

// refresh fetches all adapter data and merges it into the team stats map.
func (s *DataStore) refresh(ctx context.Context) error {
	// Phase 1: Fetch match results (primary data for attack/defense stats).
	if err := s.csvAdapter.Fetch(ctx); err != nil {
		return fmt.Errorf("cache refresh: csv_results: %w", err)
	}

	// Phase 2: Fetch Elo ratings.
	if err := s.eloAdapter.Fetch(ctx); err != nil {
		return fmt.Errorf("cache refresh: elo_ratings: %w", err)
	}

	// Build a fresh stats map from the CSV adapter output.
	freshStats := make(map[string]*models.TeamRawStats)
	for _, id := range adapter.AllTeamIDs() {
		canonical, ok := adapter.CanonicalFromID(id)
		if !ok {
			continue
		}
		csvStats, found := s.csvAdapter.GetTeamStats(canonical)
		if !found {
			// Create a placeholder so Elo can still be merged.
			csvStats = &models.TeamRawStats{
				Source:      "placeholder",
				TeamNameRaw: canonical,
			}
		}
		freshStats[canonical] = csvStats

		// Merge Elo rating.
		if elo, ok := s.eloAdapter.GetElo(canonical); ok {
			csvStats.EloRating = elo
		}
	}

	// Phase 3: World Bank (non-blocking, best-effort).
	if s.cfg.DataSource == "live" {
		wbCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
		_ = s.wbAdapter.Fetch(wbCtx)
		cancel()
		s.wbAdapter.MergeInto(freshStats, adapter.AllCanonicalByID())
	}

	// Phase 4: Wikimedia (non-blocking, best-effort).
	if s.cfg.DataSource == "live" {
		wikiCtx, cancel := context.WithTimeout(ctx, 20*time.Second)
		_ = s.wikiAdapter.Fetch(wikiCtx)
		cancel()
		s.wikiAdapter.MergeInto(freshStats, adapter.AllCanonicalByID())
	}

	// Phase 5: FM CSV (optional, silent if file absent).
	_ = s.fmAdapter.Fetch(ctx)
	s.fmAdapter.MergeInto(freshStats)

	s.mu.Lock()
	s.teamStats = freshStats
	s.lastRefresh = time.Now()
	s.mu.Unlock()

	log.Printf("[cache] refresh complete: %d teams loaded", len(freshStats))
	return nil
}

// RefreshSource triggers a manual refresh of one or all sources.
// source may be "all", "csv", "elo", "worldbank", "wikimedia", or "fm".
func (s *DataStore) RefreshSource(ctx context.Context, source string) error {
	switch source {
	case "all":
		return s.refresh(ctx)
	case "csv":
		return s.csvAdapter.Fetch(ctx)
	case "elo":
		return s.eloAdapter.Fetch(ctx)
	case "worldbank":
		return s.wbAdapter.Fetch(ctx)
	case "wikimedia":
		return s.wikiAdapter.Fetch(ctx)
	case "fm":
		return s.fmAdapter.Fetch(ctx)
	default:
		return fmt.Errorf("unknown source %q; valid values: all, csv, elo, worldbank, wikimedia, fm", source)
	}
}

// GetTeamStats returns the cached stats for the given canonical team name.
func (s *DataStore) GetTeamStats(canonical string) (*models.TeamRawStats, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	st, ok := s.teamStats[canonical]
	return st, ok
}

// AllTeamProfiles returns all teams in TeamProfile format for the /api/teams endpoint.
func (s *DataStore) AllTeamProfiles() []models.TeamProfile {
	s.mu.RLock()
	defer s.mu.RUnlock()

	reg, err := teams.Global()
	if err != nil {
		return nil
	}

	profiles := make([]models.TeamProfile, 0, len(reg.All()))
	for _, entry := range reg.All() {
		profile := models.TeamProfile{
			ID:       entry.ID,
			NameEN:   entry.NameEN,
			NameCN:   entry.NameCN,
			FlagCode: entry.FlagCode,
			Group:    entry.Group,
			IsHost:   entry.IsHost,
		}
		if st, ok := s.teamStats[entry.CanonicalName]; ok {
			profile.Elo = st.EloRating
			profile.AvgGoalsScored = st.AvgGoalsScored
			profile.AvgGoalsConceded = st.AvgGoalsConceded
		}
		profiles = append(profiles, profile)
	}
	return profiles
}

// LastRefresh returns the time of the most recent successful data refresh.
func (s *DataStore) LastRefresh() time.Time {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.lastRefresh
}

