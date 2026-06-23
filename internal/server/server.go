package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"wc-predictor/internal/cache"
	"wc-predictor/internal/semantic"
)

// Server encapsulates the HTTP server and its dependencies.
type Server struct {
	httpServer *http.Server
	store      *cache.DataStore
}

// New creates a configured HTTP server. The server is not started until Run() is called.
func New(host string, port int, store *cache.DataStore, dsClient *semantic.DeepSeekClient) *Server {
	h := NewHandlers(store, dsClient)
	mux := http.NewServeMux()

	// Route registration — all routes are prefixed with /api.
	mux.HandleFunc("/api/health", h.Health)
	mux.HandleFunc("/api/teams", h.Teams)
	mux.HandleFunc("/api/predict", h.Predict)
	mux.HandleFunc("/api/data/status", h.DataStatus)
	mux.HandleFunc("/api/data/refresh", h.FetchSource)

	// Apply middleware stack: logging → CORS → JSON content-type.
	handler := chain(mux, loggingMiddleware, corsMiddleware, jsonMiddleware)

	return &Server{
		store: store,
		httpServer: &http.Server{
			Addr:         fmt.Sprintf("%s:%d", host, port),
			Handler:      handler,
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 30 * time.Second,
			IdleTimeout:  60 * time.Second,
		},
	}
}

// Run starts the HTTP server and blocks until ctx is cancelled.
func (s *Server) Run(ctx context.Context) error {
	errCh := make(chan error, 1)
	go func() {
		log.Printf("[server] listening on %s", s.httpServer.Addr)
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	select {
	case <-ctx.Done():
		shutCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		log.Println("[server] shutting down gracefully...")
		return s.httpServer.Shutdown(shutCtx)
	case err := <-errCh:
		return fmt.Errorf("server: %w", err)
	}
}
