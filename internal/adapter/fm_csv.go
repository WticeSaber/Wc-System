package adapter

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"wc-predictor/internal/models"
	"wc-predictor/internal/teams"
)

// FMCSVAdapter parses a Football Manager squad export CSV to derive
// aggregate attack and defense attributes per national team.
//
// Expected FM export columns (minimum required):
//
//	Name, Club, Nat, CA, PA, Pac, Sho, Pas, Dri, Def, Phy
type FMCSVAdapter struct {
	fmCSVPath  string
	attributes map[string]*models.FMTeamAttributes // canonical name → aggregated stats
}

// NewFMCSVAdapter creates an adapter that reads from the given file path.
// If path is empty, the adapter is disabled and returns empty attributes.
func NewFMCSVAdapter(fmCSVPath string) *FMCSVAdapter {
	return &FMCSVAdapter{
		fmCSVPath:  fmCSVPath,
		attributes: make(map[string]*models.FMTeamAttributes),
	}
}

// Fetch reads the FM CSV file if it exists and populates aggregated team attributes.
// If the file does not exist, it silently returns nil (non-fatal).
func (a *FMCSVAdapter) Fetch(_ context.Context) error {
	if a.fmCSVPath == "" {
		return nil
	}

	f, err := os.Open(a.fmCSVPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // file optional, not an error
		}
		return fmt.Errorf("fm_csv: open file: %w", err)
	}
	defer f.Close()

	return a.parseCSV(f)
}

type fmPlayerRow struct {
	nat string
	sho float64
	dri float64
	def float64
	phy float64
}

func (a *FMCSVAdapter) parseCSV(reader io.Reader) error {
	csvReader := csv.NewReader(reader)
	csvReader.LazyQuotes = true
	csvReader.TrimLeadingSpace = true

	header, err := csvReader.Read()
	if err != nil {
		return fmt.Errorf("fm_csv: read header: %w", err)
	}

	colIdx := make(map[string]int)
	for i, h := range header {
		colIdx[strings.ToLower(strings.TrimSpace(h))] = i
	}

	reg, err := teams.Global()
	if err != nil {
		return fmt.Errorf("fm_csv: team registry: %w", err)
	}
	natToCanonical := make(map[string]string)
	for id, name := range reg.IDCanonicalMap() {
		natToCanonical[strings.ToLower(id)] = name
	}

	// Accumulate per-player stats, then aggregate per team.
	type teamAcc struct {
		shoTotal, driTotal, defTotal, phyTotal float64
		count                                  int
	}
	accMap := make(map[string]*teamAcc)

	for {
		row, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			continue
		}

		natRaw := strings.ToLower(strings.TrimSpace(safeGet(row, colIdx, "nat")))
		canonical, ok := natToCanonical[natRaw]
		if !ok {
			continue
		}

		sho := parseFloatSafe(safeGet(row, colIdx, "sho"))
		dri := parseFloatSafe(safeGet(row, colIdx, "dri"))
		def := parseFloatSafe(safeGet(row, colIdx, "def"))
		phy := parseFloatSafe(safeGet(row, colIdx, "phy"))

		if _, exists := accMap[canonical]; !exists {
			accMap[canonical] = &teamAcc{}
		}
		acc := accMap[canonical]
		acc.shoTotal += sho
		acc.driTotal += dri
		acc.defTotal += def
		acc.phyTotal += phy
		acc.count++
	}

	for canonical, acc := range accMap {
		if acc.count == 0 {
			continue
		}
		cnt := float64(acc.count)
		a.attributes[canonical] = &models.FMTeamAttributes{
			AvgAttack:   (acc.shoTotal/cnt + acc.driTotal/cnt) / 2.0,
			AvgDefense:  (acc.defTotal/cnt + acc.phyTotal/cnt) / 2.0,
			PlayerCount: acc.count,
		}
	}

	return nil
}

// safeGet retrieves a column value by name from a row, returning "" if not found.
func safeGet(row []string, colIdx map[string]int, colName string) string {
	idx, ok := colIdx[colName]
	if !ok || idx >= len(row) {
		return ""
	}
	return strings.TrimSpace(row[idx])
}

func parseFloatSafe(s string) float64 {
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return f
}

// GetAttributes returns FM aggregated attributes for a given canonical team name.
// Returns an empty struct if no data is available (file absent or team not found).
func (a *FMCSVAdapter) GetAttributes(teamName string) models.FMTeamAttributes {
	if attr, ok := a.attributes[teamName]; ok {
		return *attr
	}
	return models.FMTeamAttributes{}
}

// MergeInto copies FM attributes into a map of TeamRawStats keyed by canonical name.
func (a *FMCSVAdapter) MergeInto(statsMap map[string]*models.TeamRawStats) {
	for canonical, attr := range a.attributes {
		if s, ok := statsMap[canonical]; ok {
			s.FMAttributes = *attr
		}
	}
}
