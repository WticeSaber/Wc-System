package adapter

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"wc-predictor/internal/models"
)

const (
	resultsCSVURL  = "https://raw.githubusercontent.com/martj42/international_results/master/results.csv"
	mockResultsPath = "mock/results.csv"
	recentGamesN   = 10
)

// CSVResultsAdapter fetches and parses historical match results from martj42/international_results.
type CSVResultsAdapter struct {
	dataSource string // "live" or "mock"
	stats      map[string]*models.TeamRawStats
}

// NewCSVResultsAdapter creates a new adapter. dataSource should be "live" or "mock".
func NewCSVResultsAdapter(dataSource string) *CSVResultsAdapter {
	return &CSVResultsAdapter{
		dataSource: dataSource,
		stats:      make(map[string]*models.TeamRawStats),
	}
}

// Fetch downloads (or reads locally) the results CSV and populates the internal stats map.
func (a *CSVResultsAdapter) Fetch(ctx context.Context) error {
	var reader io.Reader

	if a.dataSource == "mock" {
		f, err := os.Open(mockResultsPath)
		if err != nil {
			return fmt.Errorf("csv_results: open mock file: %w", err)
		}
		defer f.Close()
		reader = f
	} else {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, resultsCSVURL, nil)
		if err != nil {
			return fmt.Errorf("csv_results: build request: %w", err)
		}
		req.Header.Set("User-Agent", "wc-predictor/1.0")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return fmt.Errorf("csv_results: fetch URL: %w", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("csv_results: unexpected status %d", resp.StatusCode)
		}
		reader = resp.Body
	}

	return a.parseCSV(reader)
}

// parseCSV reads all rows and computes per-team stats from the last N competitive matches.
func (a *CSVResultsAdapter) parseCSV(reader io.Reader) error {
	// Collect all match records per team first, then slice to last N.
	type rawMatch struct {
		date        time.Time
		goalsFor    int
		goalsAgainst int
		isNeutral   bool
		opponent    string
	}

	teamMatches := make(map[string][]rawMatch)

	csvReader := csv.NewReader(reader)
	header, err := csvReader.Read()
	if err != nil {
		return fmt.Errorf("csv_results: read header: %w", err)
	}

	// Locate column indices by header name (order-independent).
	colIdx := make(map[string]int)
	for i, h := range header {
		colIdx[strings.TrimSpace(h)] = i
	}

	required := []string{"date", "home_team", "away_team", "home_score", "away_score", "neutral"}
	for _, col := range required {
		if _, ok := colIdx[col]; !ok {
			return fmt.Errorf("csv_results: missing required column %q", col)
		}
	}

	for {
		row, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			continue // skip malformed rows
		}

		dateStr := strings.TrimSpace(row[colIdx["date"]])
		homeTeam := Normalize(strings.TrimSpace(row[colIdx["home_team"]]))
		awayTeam := Normalize(strings.TrimSpace(row[colIdx["away_team"]]))
		homeScoreStr := strings.TrimSpace(row[colIdx["home_score"]])
		awayScoreStr := strings.TrimSpace(row[colIdx["away_score"]])
		neutralStr := strings.ToLower(strings.TrimSpace(row[colIdx["neutral"]]))

		// Skip rows with missing scores (future fixtures).
		if homeScoreStr == "" || awayScoreStr == "" {
			continue
		}

		date, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			continue
		}

		homeScore, err := strconv.Atoi(homeScoreStr)
		if err != nil {
			continue
		}
		awayScore, err := strconv.Atoi(awayScoreStr)
		if err != nil {
			continue
		}

		isNeutral := neutralStr == "true"

		teamMatches[homeTeam] = append(teamMatches[homeTeam], rawMatch{
			date:         date,
			goalsFor:     homeScore,
			goalsAgainst: awayScore,
			isNeutral:    isNeutral,
			opponent:     awayTeam,
		})

		teamMatches[awayTeam] = append(teamMatches[awayTeam], rawMatch{
			date:         date,
			goalsFor:     awayScore,
			goalsAgainst: homeScore,
			isNeutral:    isNeutral,
			opponent:     homeTeam,
		})
	}

	// For each team, sort by date descending and take the last N matches.
	for teamName, matches := range teamMatches {
		sort.Slice(matches, func(i, j int) bool {
			return matches[i].date.After(matches[j].date)
		})

		n := recentGamesN
		if len(matches) < n {
			n = len(matches)
		}
		recent := matches[:n]

		var totalFor, totalAgainst float64
		recentForm := make([]models.MatchResult, 0, len(recent))
		for _, m := range recent {
			totalFor += float64(m.goalsFor)
			totalAgainst += float64(m.goalsAgainst)
			recentForm = append(recentForm, models.MatchResult{
				Date:        m.date,
				Opponent:    m.opponent,
				GoalFor:     m.goalsFor,
				GoalAgainst: m.goalsAgainst,
				IsNeutral:   m.isNeutral,
			})
		}

		count := float64(len(recent))
		if count == 0 {
			count = 1
		}

		stats, exists := a.stats[teamName]
		if !exists {
			stats = &models.TeamRawStats{Source: "csv_results", TeamNameRaw: teamName}
			a.stats[teamName] = stats
		}
		stats.AvgGoalsScored = totalFor / count
		stats.AvgGoalsConceded = totalAgainst / count
		stats.RecentForm = recentForm
	}

	return nil
}

// GetTeamStats returns the parsed stats for a given canonical team name.
func (a *CSVResultsAdapter) GetTeamStats(teamName string) (*models.TeamRawStats, bool) {
	s, ok := a.stats[teamName]
	return s, ok
}
