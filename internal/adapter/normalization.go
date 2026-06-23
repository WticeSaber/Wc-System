package adapter

import "wc-predictor/internal/teams"

// Normalize converts any known variant of a team name to its canonical form.
func Normalize(name string) string {
	reg, err := teams.Global()
	if err != nil {
		return name
	}
	return reg.Normalize(name)
}

// CanonicalFromID returns the canonical English name for a given team ID.
func CanonicalFromID(id string) (string, bool) {
	reg, err := teams.Global()
	if err != nil {
		return "", false
	}
	return reg.CanonicalFromID(id)
}

// ISO2FromID returns the ISO 3166-1 alpha-2 code for a given team ID.
func ISO2FromID(id string) (string, bool) {
	reg, err := teams.Global()
	if err != nil {
		return "", false
	}
	return reg.ISO2FromID(id)
}

// CNFromID returns the Chinese name for a given team ID.
func CNFromID(id string) string {
	reg, err := teams.Global()
	if err != nil {
		return ""
	}
	return reg.CNFromID(id)
}

// AllTeamIDs returns the list of all supported team IDs from wc2026_teams.json.
func AllTeamIDs() []string {
	reg, err := teams.Global()
	if err != nil {
		return nil
	}
	return reg.AllIDs()
}

// AllCanonicalByID returns the full map of teamID → canonical name.
func AllCanonicalByID() map[string]string {
	reg, err := teams.Global()
	if err != nil {
		return map[string]string{}
	}
	return reg.IDCanonicalMap()
}
