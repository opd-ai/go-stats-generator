package api

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// Run starts the HTTP server on the specified port.
func Run(port int, version string) error {
	server := NewServer(version)

	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/health", server.HandleHealth)
	mux.HandleFunc("/api/v1/analyze", server.HandleAnalyze)
	mux.HandleFunc("/api/v1/report/", server.HandleReport)

	addr := fmt.Sprintf(":%d", port)
	srv := &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	fmt.Printf("Starting API server on %s\n", addr)
	return srv.ListenAndServe()
}

// Shutdown gracefully stops the HTTP server.
func Shutdown(srv *http.Server) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	return srv.Shutdown(ctx)
}
