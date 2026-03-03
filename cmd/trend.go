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

This command provides time-series analysis of burden metrics including:
  - MBI (Maintenance Burden Index) score trends
  - Duplication ratio changes
  - Documentation coverage evolution
  - Complexity violation trends
  - Naming convention compliance

Trends show improvement/degradation over time with visual indicators.`,
	RunE: runTrend,
}

var trendAnalyzeCmd = &cobra.Command{
	Use:   "analyze",
	Short: "Analyze burden metric trends over time",
	Long: `Analyze burden metric trends over a time period.

Displays trend analysis for:
  - MBI (Maintenance Burden Index) score
  - Duplication ratio
  - Documentation coverage
  - Complexity violations
  - Naming violations

Shows start/end values, delta, and direction (improving/degrading/stable).`,
	RunE: runTrendAnalyze,
}

var trendForecastCmd = &cobra.Command{
	Use:   "forecast",
	Short: "Forecast future metrics (PLACEHOLDER - implementation planned)",
	Long: `Generate forecasts for future metric values based on historical trends.

⚠️  PLACEHOLDER: This command currently returns structural output only.
Full implementation with regression analysis and time series forecasting
(ARIMA, exponential smoothing) is planned for a future release.`,
	RunE: runTrendForecast,
}

var trendRegressionsCmd = &cobra.Command{
	Use:   "regressions",
	Short: "Detect metric regressions (PLACEHOLDER - implementation planned)",
	Long: `Detect potential regressions by analyzing recent changes.

⚠️  PLACEHOLDER: This command currently returns structural output only.
Full implementation with statistical hypothesis testing and significance
analysis is planned for a future release.

For production regression detection, use the 'diff' command to compare
specific baseline snapshots.`,
	RunE: runTrendRegressions,
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

// initStorageWithSnapshots initializes storage backend and retrieves historical snapshots
func initStorageWithSnapshots(days int) (storage.MetricsStorage, []storage.SnapshotInfo, error) {
	cfg := config.DefaultConfig()
	sqliteConfig := storage.SQLiteConfig{
		Path:              cfg.Storage.Path,
		EnableWAL:         true,
		MaxConnections:    10,
		EnableCompression: cfg.Storage.Compression,
	}

	storageBackend, err := storage.NewSQLiteStorage(sqliteConfig)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to initialize storage: %w", err)
	}

	ctx := context.Background()
	cutoffTime := time.Now().AddDate(0, 0, -days)
	filter := storage.SnapshotFilter{
		After: &cutoffTime,
		Limit: 100,
	}

	snapshots, err := storageBackend.List(ctx, filter)
	if err != nil {
		storageBackend.Close()
		return nil, nil, fmt.Errorf("failed to retrieve snapshots: %w", err)
	}

	return storageBackend, snapshots, nil
}

// runTrendAnalyze retrieves historical snapshots from storage within the specified
// time window and outputs trend statistics (avg complexity, duplication, doc coverage)
// in console or JSON format. Currently provides basic aggregation; advanced statistical
// analysis (regression, forecasting) is planned for future releases.
func runTrendAnalyze(cmd *cobra.Command, args []string) error {
	if verbose {
		fmt.Printf("Analyzing trends for the last %d days\n", trendDays)
	}

	storageBackend, snapshots, err := initStorageWithSnapshots(trendDays)
	if err != nil {
		return err
	}
	defer storageBackend.Close()

	if len(snapshots) < 2 {
		return fmt.Errorf("insufficient snapshots for trend analysis (need at least 2, found %d)", len(snapshots))
	}

	// Analyze trends
	trendAnalysis := analyzeTrends(snapshots, trendMetric, trendEntity, trendThreshold)

	// Output results
	if outputFormat == "console" {
		outputTrendAnalysisConsole(trendAnalysis)
	} else {
		outputWriter := os.Stdout
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

// runTrendForecast generates future trend predictions based on historical data
// within the specified time window. Currently provides placeholder forecasts; advanced
// statistical forecasting (ARIMA, linear regression) is planned for future releases.
func runTrendForecast(cmd *cobra.Command, args []string) error {
	if verbose {
		fmt.Printf("Generating forecasts based on %d days of data\n", trendDays)
	}

	storageBackend, snapshots, err := initStorageWithSnapshots(trendDays)
	if err != nil {
		return err
	}
	defer storageBackend.Close()

	if len(snapshots) < 3 {
		return fmt.Errorf("insufficient snapshots for forecasting (need at least 3, found %d)", len(snapshots))
	}

	// Generate forecasts
	forecasts := generateForecasts(snapshots, trendMetric, trendEntity)

	// Output results
	if outputFormat == "console" {
		outputForecastsConsole(forecasts)
	} else {
		outputWriter := os.Stdout
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

// runTrendRegressions detects quality regressions by comparing recent snapshots
// against older baselines within the specified time window. Identifies increases
// in complexity, duplication, or decreases in documentation coverage.
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
		outputWriter := os.Stdout
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
		first := snapshots[len(snapshots)-1]
		last := snapshots[0]

		summary := map[string]interface{}{
			"start_date": first.Timestamp.Format("2006-01-02"),
			"end_date":   last.Timestamp.Format("2006-01-02"),
			"timespan":   last.Timestamp.Sub(first.Timestamp).Hours() / 24,
		}

		trends := calculateBurdenTrends(first, last)
		result["burden_trends"] = trends
		result["summary"] = summary
	}

	return result
}

func generateForecasts(snapshots []storage.SnapshotInfo, metric, entity string) map[string]interface{} {
	// PLACEHOLDER IMPLEMENTATION
	// Full forecasting with regression analysis, ARIMA, or exponential smoothing
	// will be implemented in a future release

	result := map[string]interface{}{
		"placeholder_notice": "⚠️  PLACEHOLDER: Full forecasting implementation (regression, ARIMA, exponential smoothing) planned for future release.",
		"metric":             metric,
		"entity":             entity,
		"method":             "linear_regression (planned)",
		"forecasts":          []map[string]interface{}{},
		"confidence":         0.0,
	}

	return result
}

func detectRegressions(historical, recent []storage.SnapshotInfo, threshold float64) map[string]interface{} {
	// PLACEHOLDER IMPLEMENTATION
	// Full regression detection with statistical hypothesis testing and
	// significance analysis will be implemented in a future release

	result := map[string]interface{}{
		"placeholder_notice":   "⚠️  PLACEHOLDER: Full regression detection with statistical testing planned for future release. Use 'diff' command for production comparisons.",
		"threshold":            threshold,
		"historical_count":     len(historical),
		"recent_count":         len(recent),
		"detected_regressions": []map[string]interface{}{},
		"severity":             "low",
	}

	return result
}

// Output functions

// calculateBurdenTrends computes trend direction and delta for burden metrics
func calculateBurdenTrends(first, last storage.SnapshotInfo) map[string]interface{} {
	trends := make(map[string]interface{})

	if first.MBIScoreAvg != nil && last.MBIScoreAvg != nil {
		delta := *last.MBIScoreAvg - *first.MBIScoreAvg
		trends["mbi_score"] = map[string]interface{}{
			"start":     *first.MBIScoreAvg,
			"end":       *last.MBIScoreAvg,
			"delta":     delta,
			"direction": getTrendDirection(delta, 5.0),
		}
	}

	if first.DuplicationRatio != nil && last.DuplicationRatio != nil {
		delta := *last.DuplicationRatio - *first.DuplicationRatio
		trends["duplication_ratio"] = map[string]interface{}{
			"start":     *first.DuplicationRatio,
			"end":       *last.DuplicationRatio,
			"delta":     delta,
			"direction": getTrendDirection(delta, 0.01),
		}
	}

	if first.DocCoverage != nil && last.DocCoverage != nil {
		delta := *last.DocCoverage - *first.DocCoverage
		trends["doc_coverage"] = map[string]interface{}{
			"start":     *first.DocCoverage,
			"end":       *last.DocCoverage,
			"delta":     delta,
			"direction": getTrendDirection(-delta, 0.01),
		}
	}

	if first.ComplexityViolations != nil && last.ComplexityViolations != nil {
		delta := *last.ComplexityViolations - *first.ComplexityViolations
		trends["complexity_violations"] = map[string]interface{}{
			"start":     *first.ComplexityViolations,
			"end":       *last.ComplexityViolations,
			"delta":     delta,
			"direction": getTrendDirection(float64(delta), 1.0),
		}
	}

	if first.NamingViolations != nil && last.NamingViolations != nil {
		delta := *last.NamingViolations - *first.NamingViolations
		trends["naming_violations"] = map[string]interface{}{
			"start":     *first.NamingViolations,
			"end":       *last.NamingViolations,
			"delta":     delta,
			"direction": getTrendDirection(float64(delta), 1.0),
		}
	}

	return trends
}

// getTrendDirection returns trend direction indicator based on delta
func getTrendDirection(delta, threshold float64) string {
	if delta > threshold {
		return "degrading ↑"
	} else if delta < -threshold {
		return "improving ↓"
	}
	return "stable →"
}

// Output functions

// outputTrendAnalysisConsole displays trend analysis results to the console,
// including time period, snapshot count, and burden metrics trends
// (MBI score, duplication, documentation, complexity, naming) with visual indicators.
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

	displayTrendSummary(analysis)
	displayBurdenTrends(analysis)
}

func displayTrendSummary(analysis map[string]interface{}) {
	summary, ok := analysis["summary"].(map[string]interface{})
	if !ok {
		return
	}

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

func displayBurdenTrends(analysis map[string]interface{}) {
	burdenTrends, ok := analysis["burden_trends"].(map[string]interface{})
	if !ok || len(burdenTrends) == 0 {
		return
	}

	fmt.Println()
	fmt.Println("=== BURDEN METRICS TRENDS ===")

	displayMBITrend(burdenTrends)
	displayDuplicationTrend(burdenTrends)
	displayDocCoverageTrend(burdenTrends)
	displayComplexityViolationsTrend(burdenTrends)
	displayNamingViolationsTrend(burdenTrends)
}

func displayMBITrend(burdenTrends map[string]interface{}) {
	mbi, ok := burdenTrends["mbi_score"].(map[string]interface{})
	if !ok {
		return
	}
	fmt.Printf("\nMBI Score (Maintenance Burden Index):\n")
	fmt.Printf("  Start:     %.2f\n", mbi["start"])
	fmt.Printf("  End:       %.2f\n", mbi["end"])
	fmt.Printf("  Delta:     %+.2f\n", mbi["delta"])
	fmt.Printf("  Trend:     %s\n", mbi["direction"])
}

func displayDuplicationTrend(burdenTrends map[string]interface{}) {
	dup, ok := burdenTrends["duplication_ratio"].(map[string]interface{})
	if !ok {
		return
	}
	fmt.Printf("\nDuplication Ratio:\n")
	fmt.Printf("  Start:     %.2f%%\n", dup["start"].(float64)*100)
	fmt.Printf("  End:       %.2f%%\n", dup["end"].(float64)*100)
	fmt.Printf("  Delta:     %+.2f%%\n", dup["delta"].(float64)*100)
	fmt.Printf("  Trend:     %s\n", dup["direction"])
}

func displayDocCoverageTrend(burdenTrends map[string]interface{}) {
	doc, ok := burdenTrends["doc_coverage"].(map[string]interface{})
	if !ok {
		return
	}
	fmt.Printf("\nDocumentation Coverage:\n")
	fmt.Printf("  Start:     %.2f%%\n", doc["start"].(float64)*100)
	fmt.Printf("  End:       %.2f%%\n", doc["end"].(float64)*100)
	fmt.Printf("  Delta:     %+.2f%%\n", doc["delta"].(float64)*100)
	fmt.Printf("  Trend:     %s\n", doc["direction"])
}

func displayComplexityViolationsTrend(burdenTrends map[string]interface{}) {
	comp, ok := burdenTrends["complexity_violations"].(map[string]interface{})
	if !ok {
		return
	}
	fmt.Printf("\nComplexity Violations:\n")
	fmt.Printf("  Start:     %d\n", comp["start"])
	fmt.Printf("  End:       %d\n", comp["end"])
	fmt.Printf("  Delta:     %+d\n", comp["delta"])
	fmt.Printf("  Trend:     %s\n", comp["direction"])
}

func displayNamingViolationsTrend(burdenTrends map[string]interface{}) {
	naming, ok := burdenTrends["naming_violations"].(map[string]interface{})
	if !ok {
		return
	}
	fmt.Printf("\nNaming Violations:\n")
	fmt.Printf("  Start:     %d\n", naming["start"])
	fmt.Printf("  End:       %d\n", naming["end"])
	fmt.Printf("  Delta:     %+d\n", naming["delta"])
	fmt.Printf("  Trend:     %s\n", naming["direction"])
}

func outputForecastsConsole(forecasts map[string]interface{}) {
	fmt.Println("=== METRIC FORECASTS ===")

	// Display placeholder notice prominently
	if notice, ok := forecasts["placeholder_notice"].(string); ok {
		fmt.Println()
		fmt.Println(notice)
		fmt.Println()
	}

	fmt.Printf("Method: %v\n", forecasts["method"])

	if metric := forecasts["metric"]; metric != nil && metric != "" {
		fmt.Printf("Metric: %v\n", metric)
	}

	if entity := forecasts["entity"]; entity != nil && entity != "" {
		fmt.Printf("Entity: %v\n", entity)
	}

	fmt.Printf("Confidence: %.1f%%\n", forecasts["confidence"])
	fmt.Println()
}

func outputRegressionsConsole(regressions map[string]interface{}) {
	fmt.Println("=== REGRESSION DETECTION ===")

	// Display placeholder notice prominently
	if notice, ok := regressions["placeholder_notice"].(string); ok {
		fmt.Println()
		fmt.Println(notice)
		fmt.Println()
	}

	fmt.Printf("Threshold: %.1f%%\n", regressions["threshold"])
	fmt.Printf("Historical snapshots: %v\n", regressions["historical_count"])
	fmt.Printf("Recent snapshots: %v\n", regressions["recent_count"])
	fmt.Printf("Severity: %v\n", regressions["severity"])
	fmt.Println()
}
