package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	dataSource string
	logLevel   string
)

var rootCmd = &cobra.Command{
	Use:   "predictor",
	Short: "2026 World Cup Elo-Poisson match prediction engine",
	Long: `wc-predictor is a local-first football match prediction engine for the 2026
FIFA World Cup. It combines Elo ratings with Poisson distribution modelling
to generate scoreline probabilities and match outcome forecasts.`,
}

// Execute is the main entry point called from main.go.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&dataSource, "data-source", getEnvOrDefault("DATA_SOURCE", "live"),
		`data source mode: "live" fetches from GitHub, "mock" uses local files in mock/`)
	rootCmd.PersistentFlags().StringVar(&logLevel, "log-level", "info",
		`log verbosity: "info" or "debug"`)
}

func getEnvOrDefault(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}
