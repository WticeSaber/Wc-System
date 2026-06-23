package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"wc-predictor/internal/adapter"
	"wc-predictor/internal/cache"
	"wc-predictor/internal/engine"
	"wc-predictor/internal/models"
	"wc-predictor/internal/semantic"
	"wc-predictor/internal/teams"
)

var (
	calcHome      string
	calcAway      string
	calcAvgGoals  float64
	calcHomeElo   float64
	calcAwayElo   float64
	calcHomeMod   float64
	calcAwayMod   float64
	calcDeepSeek  bool
	calcHomeEloSet bool
	calcAwayEloSet bool
)

var calcCmd = &cobra.Command{
	Use:   "calc",
	Short: "Run a single match prediction and print results to the terminal",
	Example: `  predictor calc --home Argentina --away France
  predictor calc --home ARG --away FRA --home-mod -0.15 --avg-goals 2.3
  predictor calc --home Brazil --away England --home-elo 2050 --deepseek`,
	RunE: runCalc,
}

func init() {
	calcCmd.Flags().StringVar(&calcHome, "home", "", "home team name or ID (required)")
	calcCmd.Flags().StringVar(&calcAway, "away", "", "away team name or ID (required)")
	calcCmd.Flags().Float64Var(&calcAvgGoals, "avg-goals", 2.5, "match baseline expected goals (default: 2.5)")
	calcCmd.Flags().Float64Var(&calcHomeElo, "home-elo", 0, "override home team Elo (optional)")
	calcCmd.Flags().Float64Var(&calcAwayElo, "away-elo", 0, "override away team Elo (optional)")
	calcCmd.Flags().Float64Var(&calcHomeMod, "home-mod", 0, "subjective home modifier [-0.5, +0.5]")
	calcCmd.Flags().Float64Var(&calcAwayMod, "away-mod", 0, "subjective away modifier [-0.5, +0.5]")
	calcCmd.Flags().BoolVar(&calcDeepSeek, "deepseek", false, "enable DeepSeek semantic momentum modifier")

	_ = calcCmd.MarkFlagRequired("home")
	_ = calcCmd.MarkFlagRequired("away")

	rootCmd.AddCommand(calcCmd)
}

func runCalc(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	if _, err := teams.Global(); err != nil {
		return fmt.Errorf("calc: load team registry: %w", err)
	}

	cfg := cache.Config{
		DataSource: dataSource,
		TTL:        24 * time.Hour,
		FMCSVPath:  os.Getenv("FM_CSV_PATH"),
	}

	store := cache.NewDataStore(cfg)
	log.Printf("[calc] loading data (source=%s)...", dataSource)
	if err := store.Initialize(ctx); err != nil {
		return fmt.Errorf("calc: data store init: %w", err)
	}

	// Resolve team names / IDs.
	homeCanonical := resolveCalcTeamName(calcHome)
	awayCanonical := resolveCalcTeamName(calcAway)

	req := &models.MatchRequest{
		HomeTeam:    homeCanonical,
		AwayTeam:    awayCanonical,
		AvgGoals:    calcAvgGoals,
		HomeMod:     calcHomeMod,
		AwayMod:     calcAwayMod,
		IsNeutral:   true,
		UseDeepSeek: calcDeepSeek,
	}

	// Apply manual Elo overrides if the flags were explicitly set.
	if cmd.Flags().Changed("home-elo") {
		req.HomeElo = &calcHomeElo
	}
	if cmd.Flags().Changed("away-elo") {
		req.AwayElo = &calcAwayElo
	}

	var semanticFn engine.SemanticModifierFunc
	if calcDeepSeek {
		dsClient := semantic.NewDeepSeekClient()
		if dsClient == nil {
			fmt.Fprintln(os.Stderr, "WARNING: DEEPSEEK_API_KEY not set; --deepseek flag ignored")
		} else {
			semanticFn = dsClient.AsSemanticModifierFunc()
		}
	}

	result, err := engine.Calculate(ctx, req, store, semanticFn)
	if err != nil {
		return err
	}

	printCalcResult(homeCanonical, awayCanonical, result)
	return nil
}

func resolveCalcTeamName(nameOrID string) string {
	if canonical, ok := adapter.CanonicalFromID(nameOrID); ok {
		return canonical
	}
	return nameOrID
}

// printCalcResult renders the prediction result in a colorful terminal format.
func printCalcResult(home, away string, r *models.PredictionResult) {
	bold := color.New(color.Bold)
	green := color.New(color.FgGreen, color.Bold)
	cyan := color.New(color.FgCyan, color.Bold)
	red := color.New(color.FgRed, color.Bold)
	yellow := color.New(color.FgYellow)

	fmt.Println()
	bold.Printf("  ╔═══════════════════════════════════════════╗\n")
	bold.Printf("  ║   2026 WC Elo-Poisson Prediction Engine   ║\n")
	bold.Printf("  ╚═══════════════════════════════════════════╝\n")
	fmt.Println()

	bold.Printf("  %s  vs  %s\n", home, away)
	fmt.Printf("  Expected Goals: ")
	green.Printf("%.2f", r.HomeExpectedGoals)
	fmt.Printf(" — ")
	red.Printf("%.2f\n", r.AwayExpectedGoals)

	if r.DeepSeekModifier != nil {
		cyan.Printf("  AI Modifier: %+.4f\n", *r.DeepSeekModifier)
	}

	fmt.Println()
	fmt.Println("  ─── OUTCOME PROBABILITIES ─────────────────")
	printBar(green, "Home Win", r.HomeWinProb, color.FgGreen)
	printBar(cyan, "Draw    ", r.DrawProb, color.FgCyan)
	printBar(red, "Away Win", r.AwayWinProb, color.FgRed)

	fmt.Println()
	fmt.Println("  ─── TOP 3 SCORELINE PREDICTIONS ───────────")
	medals := []string{"🥇", "🥈", "🥉"}
	for i, p := range r.TopPredictions {
		if i >= 3 {
			break
		}
		fmt.Printf("  %s  %d-%d  →  %.2f%%\n", medals[i], p.HomeScore, p.AwayScore, p.Probability*100)
	}

	if r.AlertTriggered {
		fmt.Println()
		red.Printf("  %s\n", r.AlertMessage)
	}

	fmt.Println()
	yellow.Printf("  Data: %s  |  Freshness: %s\n",
		r.EloSource,
		r.DataFreshnessAt.Format("2006-01-02 15:04 UTC"),
	)
	fmt.Println()
}

// printBar renders a simple ASCII percentage bar for terminal output.
func printBar(c *color.Color, label string, prob float64, fgColor color.Attribute) {
	barWidth := 30
	filled := int(prob * float64(barWidth))
	bar := strings.Repeat("█", filled) + strings.Repeat("░", barWidth-filled)
	c.Printf("  %s  [%s] %.1f%%\n", label, bar, prob*100)
}
