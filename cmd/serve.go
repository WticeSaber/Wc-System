package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"wc-predictor/internal/cache"
	"wc-predictor/internal/semantic"
	"wc-predictor/internal/server"
	"wc-predictor/internal/teams"
)

var (
	servePort int
	serveHost string
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the HTTP API server",
	Long:  `Starts the REST API server used by the frontend. Initializes the data store on startup.`,
	RunE:  runServe,
}

func init() {
	serveCmd.Flags().IntVar(&servePort, "port", getPortFromEnv(), "TCP port to listen on")
	serveCmd.Flags().StringVar(&serveHost, "host", "0.0.0.0", "host/IP to bind to")
	rootCmd.AddCommand(serveCmd)
}

func getPortFromEnv() int {
	if p := os.Getenv("PORT"); p != "" {
		var port int
		if _, err := fmt.Sscanf(p, "%d", &port); err == nil && port > 0 {
			return port
		}
	}
	return 8080
}

func runServe(cmd *cobra.Command, args []string) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if _, err := teams.Global(); err != nil {
		return fmt.Errorf("serve: load team registry: %w", err)
	}

	cfg := cache.Config{
		DataSource: dataSource,
		TTL:        24 * time.Hour,
		FMCSVPath:  os.Getenv("FM_CSV_PATH"),
	}

	store := cache.NewDataStore(cfg)
	log.Printf("[serve] initializing data store (source=%s)...", dataSource)
	if err := store.Initialize(ctx); err != nil {
		return fmt.Errorf("serve: data store init: %w", err)
	}
	store.StartBackgroundRefresh(ctx)

	dsClient := semantic.NewDeepSeekClient()
	if dsClient == nil {
		log.Println("[serve] DEEPSEEK_API_KEY not set; semantic layer disabled")
	}

	srv := server.New(serveHost, servePort, store, dsClient)
	return srv.Run(ctx)
}
