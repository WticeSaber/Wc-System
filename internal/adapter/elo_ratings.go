package adapter

import (
	"bufio"
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"wc-predictor/internal/models"
	"wc-predictor/internal/teams"
)

const (
	// eloratingsWorldTSV is the live World Football Elo table (headerless TSV).
	// Columns: rank, rank, country-code, elo, ...
	eloratingsWorldTSV = "https://www.eloratings.net/World.tsv"
	mockEloPath        = "mock/elo_ratings.csv"
)

// EloRatingsAdapter loads national team Elo ratings for roster teams.
type EloRatingsAdapter struct {
	dataSource string
	ratings    map[string]float64 // canonical name → Elo score
}

// NewEloRatingsAdapter creates a new Elo ratings adapter.
func NewEloRatingsAdapter(dataSource string) *EloRatingsAdapter {
	return &EloRatingsAdapter{
		dataSource: dataSource,
		ratings:    make(map[string]float64),
	}
}

// Fetch downloads live Elo data or reads the local mock CSV.
func (a *EloRatingsAdapter) Fetch(ctx context.Context) error {
	if a.dataSource == "mock" {
		return a.loadMockCSV()
	}

	if err := a.fetchLiveTSV(ctx); err != nil {
		log.Printf("[elo_ratings] live fetch failed: %v; falling back to %s", err, mockEloPath)
		if mockErr := a.loadMockCSV(); mockErr != nil {
			return fmt.Errorf("elo_ratings: live failed (%v) and mock fallback failed (%v)", err, mockErr)
		}
		log.Printf("[elo_ratings] using mock fallback (%d ratings)", len(a.ratings))
	}
	return nil
}

func (a *EloRatingsAdapter) fetchLiveTSV(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, eloratingsWorldTSV, nil)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("User-Agent", "wc-predictor/1.0")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("fetch URL: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status %d", resp.StatusCode)
	}

	codeToElo, err := parseEloRatingsTSV(resp.Body)
	if err != nil {
		return err
	}

	reg, err := teams.Global()
	if err != nil {
		return err
	}

	a.ratings = make(map[string]float64)
	matched := 0
	for _, entry := range reg.All() {
		code := entry.EloCode
		if code == "" {
			continue
		}
		elo, ok := codeToElo[strings.ToUpper(code)]
		if !ok {
			continue
		}
		a.ratings[entry.CanonicalName] = elo
		matched++
	}

	if matched == 0 {
		return fmt.Errorf("no roster teams matched eloratings.net codes")
	}
	log.Printf("[elo_ratings] loaded %d/%d team Elo ratings from eloratings.net", matched, len(reg.All()))
	return nil
}

// parseEloRatingsTSV parses the headerless World.tsv from eloratings.net.
// Returns a map of country-code → Elo (e.g. "AR" → 2144).
func parseEloRatingsTSV(r io.Reader) (map[string]float64, error) {
	out := make(map[string]float64)
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		fields := strings.Split(line, "\t")
		if len(fields) < 4 {
			continue
		}
		code := strings.ToUpper(strings.TrimSpace(fields[2]))
		elo, err := strconv.ParseFloat(strings.TrimSpace(fields[3]), 64)
		if err != nil || code == "" {
			continue
		}
		out[code] = elo
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read TSV: %w", err)
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("empty TSV payload")
	}
	return out, nil
}

func (a *EloRatingsAdapter) loadMockCSV() error {
	f, err := os.Open(mockEloPath)
	if err != nil {
		return fmt.Errorf("open mock file: %w", err)
	}
	defer f.Close()

	a.ratings = make(map[string]float64)
	return a.parseCSV(f)
}

// parseCSV reads mock/local CSV with Team and Elo columns.
func (a *EloRatingsAdapter) parseCSV(reader io.Reader) error {
	csvReader := csv.NewReader(reader)
	csvReader.LazyQuotes = true

	header, err := csvReader.Read()
	if err != nil {
		return fmt.Errorf("read header: %w", err)
	}

	colIdx := make(map[string]int)
	for i, h := range header {
		colIdx[strings.TrimSpace(strings.ToLower(h))] = i
	}

	teamCol, hasTeam := colIdx["team"]
	eloCol, hasElo := colIdx["elo"]
	if !hasTeam || !hasElo {
		teamCol, hasTeam = colIdx["country"]
		if !hasTeam {
			return fmt.Errorf("cannot find team/country and elo columns in header %v", header)
		}
	}

	for {
		row, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			continue
		}
		if len(row) <= teamCol || len(row) <= eloCol {
			continue
		}

		teamName := Normalize(strings.TrimSpace(row[teamCol]))
		eloStr := strings.TrimSpace(row[eloCol])

		elo, err := strconv.ParseFloat(eloStr, 64)
		if err != nil {
			continue
		}

		a.ratings[teamName] = elo
	}

	return nil
}

// GetElo returns the Elo rating for a given canonical team name.
func (a *EloRatingsAdapter) GetElo(teamName string) (float64, bool) {
	elo, ok := a.ratings[teamName]
	return elo, ok
}

// MergeInto copies Elo ratings into TeamRawStats keyed by canonical name.
func (a *EloRatingsAdapter) MergeInto(statsMap map[string]*models.TeamRawStats) {
	for teamName, elo := range a.ratings {
		if s, ok := statsMap[teamName]; ok {
			s.EloRating = elo
		}
	}
}
