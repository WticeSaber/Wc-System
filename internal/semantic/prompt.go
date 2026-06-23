package semantic

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"
	"time"

	"wc-predictor/internal/models"
)

const modifierPromptTemplate = `You are a football analytics assistant for the 2026 FIFA World Cup.
Analyze the following match context and return ONLY a valid JSON object with no markdown formatting.

Match: {{.HomeTeam}} vs {{.AwayTeam}}
{{.HomeTeam}} recent results (last {{.HomeFormLen}} games): {{.HomeFormStr}}
{{.AwayTeam}} recent results (last {{.AwayFormLen}} games): {{.AwayFormStr}}
Tournament: 2026 FIFA World Cup

Return format (strict JSON, no markdown, no explanation outside JSON):
{"modifier": <float between -0.2 and 0.2>, "reasoning": "<max 20 words>"}

Rules:
- "modifier" must be a float64 number between -0.2 and 0.2 (inclusive)
- Positive modifier means {{.HomeTeam}} has a momentum/intelligence advantage
- Negative modifier means {{.AwayTeam}} has a momentum/intelligence advantage
- 0.0 means no meaningful advantage detected`

type promptData struct {
	HomeTeam    string
	AwayTeam    string
	HomeFormStr string
	AwayFormStr string
	HomeFormLen int
	AwayFormLen int
}

var tmpl = template.Must(template.New("modifier").Parse(modifierPromptTemplate))

// BuildPrompt constructs the DeepSeek prompt string for the given match context.
func BuildPrompt(homeTeam, awayTeam string, homeForm, awayForm []models.MatchResult) (string, error) {
	data := promptData{
		HomeTeam:    homeTeam,
		AwayTeam:    awayTeam,
		HomeFormStr: formatForm(homeForm),
		AwayFormStr: formatForm(awayForm),
		HomeFormLen: len(homeForm),
		AwayFormLen: len(awayForm),
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("semantic: build prompt: %w", err)
	}
	return buf.String(), nil
}

// formatForm converts the last N match results into a human-readable string.
// Example: "W(2-1 vs France), D(1-1 vs Brazil), L(0-2 vs Spain)"
func formatForm(form []models.MatchResult) string {
	if len(form) == 0 {
		return "no recent data"
	}

	n := 5
	if len(form) < n {
		n = len(form)
	}

	parts := make([]string, 0, n)
	for _, m := range form[:n] {
		var result string
		switch {
		case m.GoalFor > m.GoalAgainst:
			result = "W"
		case m.GoalFor == m.GoalAgainst:
			result = "D"
		default:
			result = "L"
		}
		parts = append(parts, fmt.Sprintf("%s(%d-%d vs %s on %s)",
			result,
			m.GoalFor,
			m.GoalAgainst,
			m.Opponent,
			m.Date.Format(time.DateOnly),
		))
	}
	return strings.Join(parts, ", ")
}
