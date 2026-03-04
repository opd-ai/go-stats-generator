package cmd

import (
	"fmt"
	"os"

	"github.com/opd-ai/go-stats-generator/internal/api"
	"github.com/spf13/cobra"
)

var serverPort int

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the REST API server",
	Long: `Start the go-stats-generator REST API server.

The API server provides HTTP endpoints for analyzing Go repositories:
  POST /api/v1/analyze  - Submit an analysis request
  GET  /api/v1/report/{id} - Retrieve analysis results
  GET  /api/v1/health   - Check API health status

Example:
  go-stats-generator serve --port 8080`,
	RunE: runServe,
}

func init() {
	rootCmd.AddCommand(serveCmd)
	serveCmd.Flags().IntVarP(&serverPort, "port", "p", 8080, "API server port")
}

func runServe(cmd *cobra.Command, args []string) error {
	version := rootCmd.Version
	if version == "" {
		version = "dev"
	}

	fmt.Fprintf(os.Stderr, "Starting go-stats-generator API server v%s\n", version)
	fmt.Fprintf(os.Stderr, "Listening on port %d\n", serverPort)

	if err := api.Run(serverPort, version); err != nil {
		return fmt.Errorf("server error: %w", err)
	}

	return nil
}
