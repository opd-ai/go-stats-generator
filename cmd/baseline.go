package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/opd-ai/go-stats-generator/internal/config"
	"github.com/opd-ai/go-stats-generator/internal/metrics"
	"github.com/opd-ai/go-stats-generator/internal/storage"
	"github.com/opd-ai/go-stats-generator/pkg/gostats"
	"github.com/spf13/cobra"
)

var (
	baselineID      string
	baselineMessage string
	tags            []string
	overwriteFlag   bool
)

var baselineCmd = &cobra.Command{
	Use:   "baseline [path]",
	Short: "Create or manage baseline snapshots for comparison",
	Long: `Create a new baseline snapshot from the current codebase state,
or manage existing baselines. Baselines are used as reference points
for complexity comparison and regression detection.`,
	RunE: runBaseline,
}

var createBaselineCmd = &cobra.Command{
	Use:   "create [path]",
	Short: "Create a new baseline snapshot",
	Long: `Create a new baseline snapshot from the current codebase state.
This snapshot can be used later for comparisons and trend analysis.`,
	RunE: runCreateBaseline,
}

var listBaselinesCmd = &cobra.Command{
	Use:   "list",
	Short: "List all available baseline snapshots",
	Long:  "List all available baseline snapshots with their metadata.",
	RunE:  runListBaselines,
}

var deleteBaselineCmd = &cobra.Command{
	Use:   "delete <baseline-id>",
	Short: "Delete a baseline snapshot",
	Long:  "Delete a specific baseline snapshot by its ID.",
	Args:  cobra.ExactArgs(1),
	RunE:  runDeleteBaseline,
}

func init() {
	// Add baseline command to root
	rootCmd.AddCommand(baselineCmd)

	// Add subcommands to baseline
	baselineCmd.AddCommand(createBaselineCmd)
	baselineCmd.AddCommand(listBaselinesCmd)
	baselineCmd.AddCommand(deleteBaselineCmd)

	// Flags for create baseline
	createBaselineCmd.Flags().StringVarP(&baselineID, "id", "i", "", "Custom ID for the baseline (auto-generated if not specified)")
	createBaselineCmd.Flags().StringVarP(&baselineMessage, "message", "m", "", "Description message for the baseline")
	createBaselineCmd.Flags().StringSliceVarP(&tags, "tags", "t", []string{}, "Tags to associate with the baseline")
	createBaselineCmd.Flags().BoolVar(&overwriteFlag, "overwrite", false, "Overwrite existing baseline with same ID")

	// Global flags inherited from root
	createBaselineCmd.Flags().StringVarP(&outputFormat, "format", "f", "json", "Output format (json, console)")
	createBaselineCmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output file (default: stdout)")
	createBaselineCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Verbose output")
}

func runBaseline(cmd *cobra.Command, args []string) error {
	// Default behavior is to create a baseline
	return runCreateBaseline(cmd, args)
}

func runCreateBaseline(cmd *cobra.Command, args []string) error {
	targetPath := "."
	if len(args) > 0 {
		targetPath = args[0]
	}

	if verbose {
		fmt.Printf("Creating baseline snapshot for: %s\n", targetPath)
	}

	// Initialize storage
	cfg := config.DefaultConfig()
	sqliteConfig := storage.SQLiteConfig{
		Path:              cfg.Storage.Path,
		EnableWAL:         true,
		MaxConnections:    10,
		EnableCompression: cfg.Storage.Compression,
	}

	storageBackend, err := storage.NewSQLiteStorage(sqliteConfig)
	if err != nil {
		return fmt.Errorf("failed to initialize storage: %w", err)
	}
	defer storageBackend.Close()

	// Analyze the codebase
	report, err := analyzeCodebase(targetPath)
	if err != nil {
		return fmt.Errorf("failed to analyze codebase: %w", err)
	}

	// Generate baseline ID if not provided
	if baselineID == "" {
		baselineID = generateBaselineID()
	}

	// Create snapshot metadata
	metadata := metrics.SnapshotMetadata{
		Timestamp:   time.Now(),
		GitBranch:   getCurrentBranch(),
		GitCommit:   getCurrentCommit(),
		Description: baselineMessage,
		Tags:        convertToTagMap(tags),
	}

	// Create snapshot
	snapshot := metrics.MetricsSnapshot{
		ID:       baselineID,
		Report:   *report,
		Metadata: metadata,
	}

	// Store the snapshot
	ctx := context.Background()
	err = storageBackend.Store(ctx, snapshot, metadata)
	if err != nil {
		if !overwriteFlag {
			return fmt.Errorf("failed to store baseline snapshot: %w", err)
		}
		// Try to overwrite
		if err := storageBackend.Delete(ctx, baselineID); err != nil {
			if verbose {
				fmt.Printf("Warning: could not delete existing baseline: %v\n", err)
			}
		}
		if err := storageBackend.Store(ctx, snapshot, metadata); err != nil {
			return fmt.Errorf("failed to overwrite baseline snapshot: %w", err)
		}
	}

	// Output the result
	if outputFormat == "console" {
		fmt.Printf("✓ Baseline snapshot created successfully\n")
		fmt.Printf("  ID: %s\n", snapshot.ID)
		fmt.Printf("  Timestamp: %s\n", snapshot.Metadata.Timestamp.Format("2006-01-02 15:04:05"))
		if snapshot.Metadata.Description != "" {
			fmt.Printf("  Message: %s\n", snapshot.Metadata.Description)
		}
		if len(snapshot.Metadata.Tags) > 0 {
			fmt.Printf("  Tags: %v\n", snapshot.Metadata.Tags)
		}
	} else {
		// JSON output
		output := map[string]interface{}{
			"status":   "success",
			"baseline": snapshot,
		}

		var outputWriter = os.Stdout
		if outputFile != "" {
			file, err := os.Create(outputFile)
			if err != nil {
				return fmt.Errorf("failed to create output file: %w", err)
			}
			defer file.Close()
			outputWriter = file
		}

		encoder := json.NewEncoder(outputWriter)
		encoder.SetIndent("", "  ")
		return encoder.Encode(output)
	}

	return nil
}

func runListBaselines(cmd *cobra.Command, args []string) error {
	// Initialize storage
	cfg := config.DefaultConfig()
	sqliteConfig := storage.SQLiteConfig{
		Path:              cfg.Storage.Path,
		EnableWAL:         true,
		MaxConnections:    10,
		EnableCompression: cfg.Storage.Compression,
	}

	storageBackend, err := storage.NewSQLiteStorage(sqliteConfig)
	if err != nil {
		return fmt.Errorf("failed to initialize storage: %w", err)
	}
	defer storageBackend.Close()

	// Get all snapshots
	ctx := context.Background()
	filter := storage.SnapshotFilter{
		Limit: 100,
	}
	snapshots, err := storageBackend.List(ctx, filter)
	if err != nil {
		return fmt.Errorf("failed to list snapshots: %w", err)
	}

	if outputFormat == "console" {
		if len(snapshots) == 0 {
			fmt.Println("No baseline snapshots found.")
			return nil
		}

		fmt.Printf("Found %d baseline snapshot(s):\n\n", len(snapshots))
		for _, info := range snapshots {
			fmt.Printf("ID: %s\n", info.ID)
			fmt.Printf("  Timestamp: %s\n", info.Timestamp.Format("2006-01-02 15:04:05"))
			if info.Description != "" {
				fmt.Printf("  Message: %s\n", info.Description)
			}
			if len(info.Tags) > 0 {
				fmt.Printf("  Tags: %v\n", info.Tags)
			}
			if info.GitBranch != "" {
				fmt.Printf("  Branch: %s\n", info.GitBranch)
			}
			if info.GitCommit != "" {
				fmt.Printf("  Commit: %s\n", info.GitCommit)
			}
			fmt.Printf("  Size: %d bytes\n", info.Size)
			fmt.Println()
		}
	} else {
		// JSON output
		var outputWriter = os.Stdout
		if outputFile != "" {
			file, err := os.Create(outputFile)
			if err != nil {
				return fmt.Errorf("failed to create output file: %w", err)
			}
			defer file.Close()
			outputWriter = file
		}

		encoder := json.NewEncoder(outputWriter)
		encoder.SetIndent("", "  ")
		return encoder.Encode(map[string]interface{}{
			"snapshots": snapshots,
			"count":     len(snapshots),
		})
	}

	return nil
}

func runDeleteBaseline(cmd *cobra.Command, args []string) error {
	baselineID := args[0]

	// Initialize storage
	cfg := config.DefaultConfig()
	sqliteConfig := storage.SQLiteConfig{
		Path:              cfg.Storage.Path,
		EnableWAL:         true,
		MaxConnections:    10,
		EnableCompression: cfg.Storage.Compression,
	}

	storageBackend, err := storage.NewSQLiteStorage(sqliteConfig)
	if err != nil {
		return fmt.Errorf("failed to initialize storage: %w", err)
	}
	defer storageBackend.Close()

	// Delete the snapshot
	ctx := context.Background()
	err = storageBackend.Delete(ctx, baselineID)
	if err != nil {
		return fmt.Errorf("failed to delete baseline snapshot: %w", err)
	}

	if outputFormat == "console" {
		fmt.Printf("✓ Baseline snapshot '%s' deleted successfully\n", baselineID)
	} else {
		// JSON output
		output := map[string]interface{}{
			"status":     "success",
			"message":    "Baseline snapshot deleted",
			"baselineId": baselineID,
		}

		var outputWriter = os.Stdout
		if outputFile != "" {
			file, err := os.Create(outputFile)
			if err != nil {
				return fmt.Errorf("failed to create output file: %w", err)
			}
			defer file.Close()
			outputWriter = file
		}

		encoder := json.NewEncoder(outputWriter)
		encoder.SetIndent("", "  ")
		return encoder.Encode(output)
	}

	return nil
}

// Helper functions

func analyzeCodebase(targetPath string) (*metrics.Report, error) {
	// Use the public API to analyze the project
	api := gostats.NewAnalyzer()
	report, err := api.AnalyzeDirectory(context.Background(), targetPath)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze project: %w", err)
	}

	return report, nil
}

func generateBaselineID() string {
	timestamp := time.Now().Format("20060102-150405")
	return fmt.Sprintf("baseline-%s", timestamp)
}

func getCurrentBranch() string {
	// Try to get current git branch
	// This is a placeholder - in real implementation you'd use git commands
	return ""
}

func getCurrentCommit() string {
	// Try to get current git commit hash
	// This is a placeholder - in real implementation you'd use git commands
	return ""
}

func convertToTagMap(tagSlice []string) map[string]string {
	tagMap := make(map[string]string)
	for _, tag := range tagSlice {
		parts := strings.SplitN(tag, "=", 2)
		if len(parts) == 2 {
			tagMap[parts[0]] = parts[1]
		} else {
			tagMap[tag] = "true"
		}
	}
	return tagMap
}
