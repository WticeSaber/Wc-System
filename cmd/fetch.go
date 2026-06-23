package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"wc-predictor/internal/cache"
)

var fetchSource string

var fetchCmd = &cobra.Command{
	Use:   "fetch",
	Short: "Manually trigger a data refresh from external sources",
	Long: `Fetches and caches fresh data from external sources. Useful after a tournament
day to ensure the latest match results and Elo ratings are loaded before running predictions.`,
	Example: `  predictor fetch --source all
  predictor fetch --source elo
  predictor fetch --source csv`,
	RunE: runFetch,
}

func init() {
	fetchCmd.Flags().StringVar(&fetchSource, "source", "all",
		`source to refresh: all, csv, elo, worldbank, wikimedia, fm`)
	rootCmd.AddCommand(fetchCmd)
}

func runFetch(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	cfg := cache.Config{
		DataSource: dataSource,
		TTL:        24 * time.Hour,
		FMCSVPath:  os.Getenv("FM_CSV_PATH"),
	}

	store := cache.NewDataStore(cfg)
	fmt.Printf("[fetch] refreshing source=%q ...\n", fetchSource)

	if err := store.RefreshSource(ctx, fetchSource); err != nil {
		return fmt.Errorf("fetch: %w", err)
	}

	fmt.Printf("[fetch] done at %s\n", time.Now().Format(time.RFC3339))
	return nil
}
