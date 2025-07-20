package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/opd-ai/go-stats-generator/internal/config"
	"github.com/opd-ai/go-stats-generator/internal/storage"
	"github.com/spf13/cobra"
)

var (
	trendDays      int
	trendMetric    string
	trendEntity    string
	trendThreshold float64
)

var trendCmd = &cobra.Command{
	Use:   "trend",
	Short: "Analyze trends in code metrics over time",
	Long: `Analyze trends and patterns in code metrics over time.
This command provides trend analysis, forecasting, and regression detection
based on historical metrics snapshots.`,
	RunE: runTrend,
}

var trendAnalyzeCmd = &cobra.Command{
	Use:   "analyze",
	Short: "Analyze trends for specific metrics",
	Long:  "Analyze trends for specific metrics over a time period.",
	RunE:  runTrendAnalyze,
}

var trendForecastCmd = &cobra.Command{
	Use:   "forecast",
	Short: "Forecast future metric values based on trends",
	Long:  "Generate forecasts for future metric values based on historical trends.",
	RunE:  runTrendForecast,
}

var trendRegressionsCmd = &cobra.Command{
	Use:   "regressions",
	Short: "Detect potential regressions in recent snapshots",
	Long:  "Detect potential regressions by analyzing recent changes against historical baselines.",
	RunE:  runTrendRegressions,
}

func init() {
	// Add trend command to root
	rootCmd.AddCommand(trendCmd)

	// Add subcommands to trend
	trendCmd.AddCommand(trendAnalyzeCmd)
	trendCmd.AddCommand(trendForecastCmd)
	trendCmd.AddCommand(trendRegressionsCmd)

	// Flags for trend analysis
	trendCmd.PersistentFlags().IntVarP(&trendDays, "days", "d", 30, "Number of days to analyze")
	trendCmd.PersistentFlags().StringVarP(&trendMetric, "metric", "m", "", "Specific metric to analyze (complexity, lines, functions)")
	trendCmd.PersistentFlags().StringVarP(&trendEntity, "entity", "e", "", "Specific entity to analyze (function, package, file)")
	trendCmd.PersistentFlags().Float64VarP(&trendThreshold, "threshold", "t", 10.0, "Threshold percentage for significance")

	// Global flags inherited from root
	trendCmd.PersistentFlags().StringVarP(&outputFormat, "format", "f", "console", "Output format (json, console)")
	trendCmd.PersistentFlags().StringVarP(&outputFile, "output", "o", "", "Output file (default: stdout)")
	trendCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Verbose output")
}

func runTrend(cmd *cobra.Command, args []string) error {
	// Default behavior is to analyze trends
	return runTrendAnalyze(cmd, args)
}

func runTrendAnalyze(cmd *cobra.Command, args []string) error {
	if verbose {
		fmt.Printf("Analyzing trends for the last %d days\n", trendDays)
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

	// Get historical snapshots
	ctx := context.Background()
	cutoffTime := time.Now().AddDate(0, 0, -trendDays)
	filter := storage.SnapshotFilter{
		After: &cutoffTime,
		Limit: 100,
	}

	snapshots, err := storageBackend.List(ctx, filter)
	if err != nil {
		return fmt.Errorf("failed to retrieve snapshots: %w", err)
	}

	if len(snapshots) < 2 {
		return fmt.Errorf("insufficient snapshots for trend analysis (need at least 2, found %d)", len(snapshots))
	}

	// Analyze trends
	trendAnalysis := analyzeTrends(snapshots, trendMetric, trendEntity, trendThreshold)

	// Output results
	if outputFormat == "console" {
		outputTrendAnalysisConsole(trendAnalysis)
	} else {
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
		return encoder.Encode(trendAnalysis)
	}

	return nil
}

func runTrendForecast(cmd *cobra.Command, args []string) error {
	if verbose {
		fmt.Printf("Generating forecasts based on %d days of data\n", trendDays)
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

	// Get historical snapshots
	ctx := context.Background()
	cutoffTime := time.Now().AddDate(0, 0, -trendDays)
	filter := storage.SnapshotFilter{
		After: &cutoffTime,
		Limit: 100,
	}

	snapshots, err := storageBackend.List(ctx, filter)
	if err != nil {
		return fmt.Errorf("failed to retrieve snapshots: %w", err)
	}

	if len(snapshots) < 3 {
		return fmt.Errorf("insufficient snapshots for forecasting (need at least 3, found %d)", len(snapshots))
	}

	// Generate forecasts
	forecasts := generateForecasts(snapshots, trendMetric, trendEntity)

	// Output results
	if outputFormat == "console" {
		outputForecastsConsole(forecasts)
	} else {
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
		return encoder.Encode(forecasts)
	}

	return nil
}

func runTrendRegressions(cmd *cobra.Command, args []string) error {
	if verbose {
		fmt.Printf("Detecting regressions in the last %d days\n", trendDays)
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

	// Get recent snapshots
	ctx := context.Background()
	recentFilter := storage.SnapshotFilter{
		Limit: 10, // Get last 10 snapshots
	}

	recentSnapshots, err := storageBackend.List(ctx, recentFilter)
	if err != nil {
		return fmt.Errorf("failed to retrieve recent snapshots: %w", err)
	}

	if len(recentSnapshots) < 2 {
		return fmt.Errorf("insufficient snapshots for regression detection (need at least 2, found %d)", len(recentSnapshots))
	}

	// Get historical baseline
	cutoffTime := time.Now().AddDate(0, 0, -trendDays)
	historicalFilter := storage.SnapshotFilter{
		Before: &cutoffTime,
		Limit:  20,
	}

	historicalSnapshots, err := storageBackend.List(ctx, historicalFilter)
	if err != nil {
		return fmt.Errorf("failed to retrieve historical snapshots: %w", err)
	}

	// Detect regressions
	regressions := detectRegressions(historicalSnapshots, recentSnapshots, trendThreshold)

	// Output results
	if outputFormat == "console" {
		outputRegressionsConsole(regressions)
	} else {
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
		return encoder.Encode(regressions)
	}

	return nil
}

// Helper functions for trend analysis

func analyzeTrends(snapshots []storage.SnapshotInfo, metric, entity string, threshold float64) map[string]interface{} {
	// This is a simplified trend analysis implementation
	// In a real implementation, you would perform statistical analysis

	result := map[string]interface{}{
		"period":    fmt.Sprintf("%d days", trendDays),
		"snapshots": len(snapshots),
		"metric":    metric,
		"entity":    entity,
		"threshold": threshold,
		"trends":    []map[string]interface{}{},
		"summary":   map[string]interface{}{},
	}

	if len(snapshots) >= 2 {
		first := snapshots[len(snapshots)-1] // Oldest
		last := snapshots[0]                 // Newest

		result["summary"] = map[string]interface{}{
			"start_date": first.Timestamp.Format("2006-01-02"),
			"end_date":   last.Timestamp.Format("2006-01-02"),
			"timespan":   last.Timestamp.Sub(first.Timestamp).Hours() / 24,
		}
	}

	return result
}

func generateForecasts(snapshots []storage.SnapshotInfo, metric, entity string) map[string]interface{} {
	// This is a simplified forecasting implementation
	// In a real implementation, you would use regression analysis or time series forecasting

	result := map[string]interface{}{
		"metric":     metric,
		"entity":     entity,
		"method":     "linear_regression",
		"forecasts":  []map[string]interface{}{},
		"confidence": 0.0,
	}

	return result
}

func detectRegressions(historical, recent []storage.SnapshotInfo, threshold float64) map[string]interface{} {
	// This is a simplified regression detection implementation
	// In a real implementation, you would compare metrics statistically

	result := map[string]interface{}{
		"threshold":            threshold,
		"historical_count":     len(historical),
		"recent_count":         len(recent),
		"detected_regressions": []map[string]interface{}{},
		"severity":             "low",
	}

	return result
}

// Output functions

func outputTrendAnalysisConsole(analysis map[string]interface{}) {
	fmt.Println("=== TREND ANALYSIS ===")
	fmt.Printf("Period: %v\n", analysis["period"])
	fmt.Printf("Snapshots analyzed: %v\n", analysis["snapshots"])

	if metric := analysis["metric"]; metric != nil && metric != "" {
		fmt.Printf("Metric: %v\n", metric)
	}

	if entity := analysis["entity"]; entity != nil && entity != "" {
		fmt.Printf("Entity: %v\n", entity)
	}

	fmt.Printf("Threshold: %.1f%%\n", analysis["threshold"])
	fmt.Println()

	if summary, ok := analysis["summary"].(map[string]interface{}); ok {
		fmt.Println("Summary:")
		if startDate := summary["start_date"]; startDate != nil {
			fmt.Printf("  Start Date: %v\n", startDate)
		}
		if endDate := summary["end_date"]; endDate != nil {
			fmt.Printf("  End Date: %v\n", endDate)
		}
		if timespan := summary["timespan"]; timespan != nil {
			fmt.Printf("  Timespan: %.1f days\n", timespan)
		}
	}
}

func outputForecastsConsole(forecasts map[string]interface{}) {
	fmt.Println("=== METRIC FORECASTS ===")
	fmt.Printf("Method: %v\n", forecasts["method"])

	if metric := forecasts["metric"]; metric != nil && metric != "" {
		fmt.Printf("Metric: %v\n", metric)
	}

	if entity := forecasts["entity"]; entity != nil && entity != "" {
		fmt.Printf("Entity: %v\n", entity)
	}

	fmt.Printf("Confidence: %.1f%%\n", forecasts["confidence"])
	fmt.Println()

	fmt.Println("Forecasting functionality is under development.")
	fmt.Println("This will provide predictions for future metric values based on historical trends.")
}

func outputRegressionsConsole(regressions map[string]interface{}) {
	fmt.Println("=== REGRESSION DETECTION ===")
	fmt.Printf("Threshold: %.1f%%\n", regressions["threshold"])
	fmt.Printf("Historical snapshots: %v\n", regressions["historical_count"])
	fmt.Printf("Recent snapshots: %v\n", regressions["recent_count"])
	fmt.Printf("Severity: %v\n", regressions["severity"])
	fmt.Println()

	fmt.Println("Regression detection functionality is under development.")
	fmt.Println("This will identify significant negative changes in code quality metrics.")
}
