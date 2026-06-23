package teams

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
)

// Entry is a single team record from wc2026_teams.json.
type Entry struct {
	ID            string   `json:"id"`
	NameEN        string   `json:"name_en"`
	NameCN        string   `json:"name_cn"`
	FlagCode      string   `json:"flag_code"`
	ISO2          string   `json:"iso2"`
	CanonicalName string   `json:"canonical_name"`
	CSVAliases    []string `json:"csv_aliases"`
	EloCode       string   `json:"elo_code"` // eloratings.net 2-letter country code
	Group         string   `json:"group"`
	IsHost        bool     `json:"is_host"`
	WikiTitle     string   `json:"wiki_title"`
}

// File is the top-level structure of wc2026_teams.json.
type File struct {
	Version    string  `json:"version"`
	Tournament string  `json:"tournament"`
	UpdatedAt  string  `json:"updated_at"`
	Teams      []Entry `json:"teams"`
}

// Registry holds the loaded team roster and lookup indexes.
type Registry struct {
	File         File
	byID         map[string]*Entry
	byCanonical  map[string]*Entry
	aliasToCanon map[string]string
}

var (
	global   *Registry
	loadOnce sync.Once
	loadErr  error
)

// DefaultPath is the default location of the team roster JSON file.
const DefaultPath = "data/wc2026_teams.json"

// Load reads and indexes the team registry from path.
// If path is empty, uses TEAMS_DATA_PATH env or DefaultPath.
func Load(path string) (*Registry, error) {
	if path == "" {
		path = os.Getenv("TEAMS_DATA_PATH")
		if path == "" {
			path = DefaultPath
		}
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("teams: read %q: %w", path, err)
	}

	var file File
	if err := json.Unmarshal(raw, &file); err != nil {
		return nil, fmt.Errorf("teams: parse JSON: %w", err)
	}
	if len(file.Teams) == 0 {
		return nil, fmt.Errorf("teams: no teams defined in %q", path)
	}

	reg := &Registry{
		File:         file,
		byID:         make(map[string]*Entry, len(file.Teams)),
		byCanonical:  make(map[string]*Entry, len(file.Teams)),
		aliasToCanon: make(map[string]string),
	}

	for i := range file.Teams {
		t := &file.Teams[i]
		reg.byID[t.ID] = t
		reg.byCanonical[t.CanonicalName] = t
		reg.indexAlias(strings.ToLower(t.CanonicalName), t.CanonicalName)
		reg.indexAlias(strings.ToLower(t.ID), t.CanonicalName)
		reg.indexAlias(strings.ToLower(t.NameEN), t.CanonicalName)
		for _, alias := range t.CSVAliases {
			reg.indexAlias(strings.ToLower(strings.TrimSpace(alias)), t.CanonicalName)
		}
	}

	return reg, nil
}

func (r *Registry) indexAlias(lower, canonical string) {
	if lower == "" {
		return
	}
	r.aliasToCanon[lower] = canonical
}

// MustGlobal returns the singleton registry, loading on first call.
func MustGlobal() *Registry {
	loadOnce.Do(func() {
		global, loadErr = Load("")
	})
	if loadErr != nil {
		panic(loadErr)
	}
	return global
}

// Global returns the singleton registry and any load error.
func Global() (*Registry, error) {
	loadOnce.Do(func() {
		global, loadErr = Load("")
	})
	return global, loadErr
}

// All returns all team entries in file order.
func (r *Registry) All() []Entry {
	return r.File.Teams
}

// AllIDs returns all team IDs.
func (r *Registry) AllIDs() []string {
	ids := make([]string, 0, len(r.File.Teams))
	for _, t := range r.File.Teams {
		ids = append(ids, t.ID)
	}
	return ids
}

// ByID looks up a team by its 3-letter ID.
func (r *Registry) ByID(id string) (*Entry, bool) {
	t, ok := r.byID[id]
	return t, ok
}

// CanonicalFromID returns the canonical English name for a team ID.
func (r *Registry) CanonicalFromID(id string) (string, bool) {
	if t, ok := r.byID[id]; ok {
		return t.CanonicalName, true
	}
	return "", false
}

// Normalize maps a raw team name or ID to its canonical name.
func (r *Registry) Normalize(nameOrID string) string {
	trimmed := strings.TrimSpace(nameOrID)
	if trimmed == "" {
		return nameOrID
	}
	if canonical, ok := r.CanonicalFromID(trimmed); ok {
		return canonical
	}
	lower := strings.ToLower(trimmed)
	if canonical, ok := r.aliasToCanon[lower]; ok {
		return canonical
	}
	return trimmed
}

// ISO2FromID returns the ISO2 code for World Bank lookups.
func (r *Registry) ISO2FromID(id string) (string, bool) {
	if t, ok := r.byID[id]; ok {
		return t.ISO2, true
	}
	return "", false
}

// CNFromID returns the Chinese display name.
func (r *Registry) CNFromID(id string) string {
	if t, ok := r.byID[id]; ok {
		return t.NameCN
	}
	return ""
}

// WikiTitleFromID returns the Wikipedia article slug for Wikimedia API.
func (r *Registry) WikiTitleFromID(id string) (string, bool) {
	if t, ok := r.byID[id]; ok && t.WikiTitle != "" {
		return t.WikiTitle, true
	}
	return "", false
}

// IDCanonicalMap returns teamID → canonical name for adapter merge helpers.
func (r *Registry) IDCanonicalMap() map[string]string {
	m := make(map[string]string, len(r.File.Teams))
	for _, t := range r.File.Teams {
		m[t.ID] = t.CanonicalName
	}
	return m
}

// HostIDs returns IDs of host nations (USA, MEX, CAN).
func (r *Registry) HostIDs() []string {
	var hosts []string
	for _, t := range r.File.Teams {
		if t.IsHost {
			hosts = append(hosts, t.ID)
		}
	}
	return hosts
}
