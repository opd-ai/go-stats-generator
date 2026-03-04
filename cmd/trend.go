package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/opd-ai/go-stats-generator/internal/analyzer"
	"github.com/opd-ai/go-stats-generator/internal/config"
	"github.com/opd-ai/go-stats-generator/internal/metrics"
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
	Short: "Forecast future metrics using linear regression",
	Long: `Generate forecasts for future metric values based on historical trends.

Uses linear regression analysis to extrapolate trends and provide:
  - Point estimates for 7, 14, and 30 days ahead
  - 95% confidence intervals
  - Reliability scores (R² values)
  - Warnings for low-quality forecasts

Requires at least 3 historical baseline snapshots.`,
	RunE: runTrendForecast,
}

var trendRegressionsCmd = &cobra.Command{
	Use:   "regressions",
	Short: "Detect metric regressions using statistical analysis",
	Long: `Detect potential regressions by analyzing recent changes against historical trends.

Uses statistical regression detection to identify:
  - Significant deviations from expected values
  - Severity classification (low/medium/high/critical)
  - P-values for statistical significance
  - Multiple metrics: MBI, duplication, documentation, complexity

Requires at least 2 historical and 1 recent snapshot.

For direct baseline comparisons, use the 'diff' command.`,
	RunE: runTrendRegressions,
}

// init registers the trend command and its subcommands with the root command.
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

// runTrend executes the default trend behavior by running trend analysis.
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
// using linear regression analysis with confidence intervals.
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

	storageBackend, err := initializeRegressionStorage()
	if err != nil {
		return err
	}
	defer storageBackend.Close()

	recentSnapshots, historicalSnapshots, err := retrieveRegressionSnapshots(storageBackend)
	if err != nil {
		return err
	}

	regressions := detectRegressions(historicalSnapshots, recentSnapshots, trendThreshold)
	return outputRegressionResults(regressions)
}

// initializeRegressionStorage sets up SQLite storage for regression detection.
func initializeRegressionStorage() (storage.MetricsStorage, error) {
	cfg := config.DefaultConfig()
	sqliteConfig := storage.SQLiteConfig{
		Path:              cfg.Storage.Path,
		EnableWAL:         true,
		MaxConnections:    10,
		EnableCompression: cfg.Storage.Compression,
	}

	storageBackend, err := storage.NewSQLiteStorage(sqliteConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize storage: %w", err)
	}
	return storageBackend, nil
}

// retrieveRegressionSnapshots fetches recent and historical snapshots for comparison.
func retrieveRegressionSnapshots(storageBackend storage.MetricsStorage) ([]storage.SnapshotInfo, []storage.SnapshotInfo, error) {
	ctx := context.Background()

	recentSnapshots, err := getRecentSnapshots(ctx, storageBackend)
	if err != nil {
		return nil, nil, err
	}

	historicalSnapshots, err := getHistoricalSnapshots(ctx, storageBackend)
	if err != nil {
		return nil, nil, err
	}

	return recentSnapshots, historicalSnapshots, nil
}

// getRecentSnapshots retrieves the most recent snapshots for regression analysis.
func getRecentSnapshots(ctx context.Context, storageBackend storage.MetricsStorage) ([]storage.SnapshotInfo, error) {
	recentFilter := storage.SnapshotFilter{
		Limit: 10,
	}

	recentSnapshots, err := storageBackend.List(ctx, recentFilter)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve recent snapshots: %w", err)
	}

	if len(recentSnapshots) < 2 {
		return nil, fmt.Errorf("insufficient snapshots for regression detection (need at least 2, found %d)", len(recentSnapshots))
	}

	return recentSnapshots, nil
}

// getHistoricalSnapshots retrieves snapshots from before the trend cutoff period.
func getHistoricalSnapshots(ctx context.Context, storageBackend storage.MetricsStorage) ([]storage.SnapshotInfo, error) {
	cutoffTime := time.Now().AddDate(0, 0, -trendDays)
	historicalFilter := storage.SnapshotFilter{
		Before: &cutoffTime,
		Limit:  20,
	}

	historicalSnapshots, err := storageBackend.List(ctx, historicalFilter)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve historical snapshots: %w", err)
	}

	return historicalSnapshots, nil
}

// outputRegressionResults writes regression data to console or JSON.
func outputRegressionResults(regressions map[string]interface{}) error {
	if outputFormat == "console" {
		outputRegressionsConsole(regressions)
		return nil
	}

	outputWriter, closer, err := createRegressionOutputWriter()
	if err != nil {
		return err
	}
	if closer != nil {
		defer closer()
	}

	encoder := json.NewEncoder(outputWriter)
	encoder.SetIndent("", "  ")
	return encoder.Encode(regressions)
}

// createRegressionOutputWriter sets up the output destination for regressions.
func createRegressionOutputWriter() (io.Writer, func(), error) {
	if outputFile == "" {
		return os.Stdout, nil, nil
	}

	file, err := os.Create(outputFile)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create output file: %w", err)
	}

	return file, func() { file.Close() }, nil
}

// Helper functions for trend analysis

// analyzeTrends computes trend metrics from historical snapshots over the configured period.
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

// generateForecasts produces metric forecasts based on historical trend data using linear regression.
func generateForecasts(snapshots []storage.SnapshotInfo, metric, entity string) map[string]interface{} {
	if len(snapshots) < 2 {
		return map[string]interface{}{
			"error": "Insufficient data for forecasting",
		}
	}

	// Build time series from snapshots based on requested metric
	series := buildTimeSeriesFromSnapshots(snapshots, metric)
	if len(series.DataPoints) < 2 {
		return map[string]interface{}{
			"error": fmt.Sprintf("No data available for metric: %s", metric),
		}
	}

	// Generate forecasts for 7, 14, and 30 days ahead
	forecast7 := analyzer.GenerateForecast(series, 7)
	forecast14 := analyzer.GenerateForecast(series, 14)
	forecast30 := analyzer.GenerateForecast(series, 30)

	result := map[string]interface{}{
		"metric":      metric,
		"entity":      entity,
		"method":      "linear_regression",
		"data_points": len(series.DataPoints),
		"forecasts": []map[string]interface{}{
			{
				"horizon_days": 7,
				"date":         forecast7.PredictionDate.Format("2006-01-02"),
				"value":        forecast7.PointEstimate,
				"lower_bound":  forecast7.LowerBound,
				"upper_bound":  forecast7.UpperBound,
				"reliability":  forecast7.ReliabilityScore,
				"warning":      forecast7.Warning,
			},
			{
				"horizon_days": 14,
				"date":         forecast14.PredictionDate.Format("2006-01-02"),
				"value":        forecast14.PointEstimate,
				"lower_bound":  forecast14.LowerBound,
				"upper_bound":  forecast14.UpperBound,
				"reliability":  forecast14.ReliabilityScore,
				"warning":      forecast14.Warning,
			},
			{
				"horizon_days": 30,
				"date":         forecast30.PredictionDate.Format("2006-01-02"),
				"value":        forecast30.PointEstimate,
				"lower_bound":  forecast30.LowerBound,
				"upper_bound":  forecast30.UpperBound,
				"reliability":  forecast30.ReliabilityScore,
				"warning":      forecast30.Warning,
			},
		},
		"trend_statistics": map[string]interface{}{
			"slope":       forecast7.TrendStatistics.Slope,
			"intercept":   forecast7.TrendStatistics.Intercept,
			"r_squared":   forecast7.TrendStatistics.RSquared,
			"correlation": forecast7.TrendStatistics.Correlation,
			"start_date":  forecast7.TrendStatistics.StartDate,
			"end_date":    forecast7.TrendStatistics.EndDate,
		},
	}

	return result
}

// detectRegressions identifies metric regressions by comparing historical and recent snapshots using statistical analysis.
func detectRegressions(historical, recent []storage.SnapshotInfo, threshold float64) map[string]interface{} {
	if len(historical) < 2 || len(recent) < 1 {
		return map[string]interface{}{
			"error": "Insufficient data for regression detection",
		}
	}

	allSnapshots := append(historical, recent...)
	allRegressions := collectMetricRegressions(allSnapshots, threshold)
	overallSeverity := determineOverallSeverity(allRegressions)

	return map[string]interface{}{
		"threshold":            threshold,
		"historical_count":     len(historical),
		"recent_count":         len(recent),
		"detected_regressions": allRegressions,
		"severity":             overallSeverity,
		"total_regressions":    len(allRegressions),
	}
}

// collectMetricRegressions analyzes multiple metrics across snapshots and collects detected regressions.
func collectMetricRegressions(snapshots []storage.SnapshotInfo, threshold float64) []map[string]interface{} {
	metrics := []string{"mbi_score", "duplication_ratio", "doc_coverage", "complexity_violations"}
	allRegressions := []map[string]interface{}{}

	for _, metricName := range metrics {
		series := buildTimeSeriesFromSnapshots(snapshots, metricName)
		if len(series.DataPoints) < 3 {
			continue
		}

		results := analyzer.DetectRegressions(series, threshold*100)
		for _, reg := range results {
			allRegressions = append(allRegressions, map[string]interface{}{
				"metric":            reg.MetricName,
				"current_value":     reg.CurrentValue,
				"expected_value":    reg.ExpectedValue,
				"percent_deviation": reg.PercentDeviation,
				"classification":    reg.Classification,
				"severity":          reg.Severity,
				"p_value":           reg.PValue,
				"detected_at":       reg.DetectedAt.Format("2006-01-02"),
			})
		}
	}
	return allRegressions
}

// determineOverallSeverity returns the highest severity level found across all regressions.
func determineOverallSeverity(regressions []map[string]interface{}) string {
	overallSeverity := "low"
	for _, reg := range regressions {
		sev, ok := reg["severity"].(string)
		if !ok {
			continue
		}
		if sev == "critical" {
			return "critical"
		}
		if sev == "high" && overallSeverity != "critical" {
			overallSeverity = "high"
		}
		if sev == "medium" && overallSeverity == "low" {
			overallSeverity = "medium"
		}
	}
	return overallSeverity
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

// buildTimeSeriesFromSnapshots extracts a time series for a specific metric from snapshot data.
func buildTimeSeriesFromSnapshots(snapshots []storage.SnapshotInfo, metricName string) metrics.MetricTimeSeries {
	series := metrics.MetricTimeSeries{
		MetricName: metricName,
		DataPoints: []metrics.TimeSeriesPoint{},
	}

	for _, snap := range snapshots {
		var value *float64

		switch metricName {
		case "mbi_score":
			value = snap.MBIScoreAvg
		case "duplication_ratio":
			value = snap.DuplicationRatio
		case "doc_coverage":
			value = snap.DocCoverage
		case "complexity_violations":
			if snap.ComplexityViolations != nil {
				floatVal := float64(*snap.ComplexityViolations)
				value = &floatVal
			}
		case "naming_violations":
			if snap.NamingViolations != nil {
				floatVal := float64(*snap.NamingViolations)
				value = &floatVal
			}
		}

		if value != nil {
			series.DataPoints = append(series.DataPoints, metrics.TimeSeriesPoint{
				Timestamp: snap.Timestamp,
				Value:     *value,
			})
		}
	}

	return series
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

// displayTrendSummary outputs the summary section of trend analysis to console.
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

// displayBurdenTrends outputs the burden metrics trends section to console.
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

// displayMBITrend outputs the MBI (Maintenance Burden Index) trend to console.
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

// displayDuplicationTrend outputs the duplication ratio trend to console.
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

// displayDocCoverageTrend outputs the documentation coverage trend to console.
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

// displayComplexityViolationsTrend outputs the complexity violations trend to console.
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

// displayNamingViolationsTrend outputs the naming violations trend to console.
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

// outputForecastsConsole displays metric forecasts in human-readable console format.
func outputForecastsConsole(forecasts map[string]interface{}) {
	fmt.Println("=== METRIC FORECASTS ===")
	fmt.Println()
	printForecastHeader(forecasts)
	printTrendStatistics(forecasts)
	printForecastValues(forecasts)
	fmt.Println()
}

func printForecastHeader(forecasts map[string]interface{}) {
	if metric := forecasts["metric"]; metric != nil && metric != "" {
		fmt.Printf("Metric: %v\n", metric)
	}
	if entity := forecasts["entity"]; entity != nil && entity != "" {
		fmt.Printf("Entity: %v\n", entity)
	}
	fmt.Printf("Method: %v\n", forecasts["method"])
	fmt.Printf("Data Points: %v\n", forecasts["data_points"])
	fmt.Println()
}

func printTrendStatistics(forecasts map[string]interface{}) {
	trendStats, ok := forecasts["trend_statistics"].(map[string]interface{})
	if !ok {
		return
	}

	fmt.Println("Trend Line:")
	fmt.Printf("  y = %.4f·x + %.4f\n", trendStats["slope"], trendStats["intercept"])
	fmt.Printf("  R² = %.4f", trendStats["r_squared"])

	if r2, ok := trendStats["r_squared"].(float64); ok {
		printReliabilityIndicator(r2)
	}
	fmt.Println()
	fmt.Println()
}

func printReliabilityIndicator(r2 float64) {
	if r2 >= 0.8 {
		fmt.Print(" (excellent fit) ✓")
	} else if r2 >= 0.5 {
		fmt.Print(" (moderate fit)")
	} else {
		fmt.Print(" (poor fit) ⚠")
	}
}

func printForecastValues(forecasts map[string]interface{}) {
	forecastsList, ok := forecasts["forecasts"].([]map[string]interface{})
	if !ok {
		return
	}

	fmt.Println("Forecasts:")
	for _, fc := range forecastsList {
		fmt.Printf("  %2d days (%v): %.2f  [%.2f - %.2f]\n",
			fc["horizon_days"], fc["date"], fc["value"],
			fc["lower_bound"], fc["upper_bound"])

		if warning, ok := fc["warning"].(string); ok && warning != "" {
			fmt.Printf("           ⚠  %s\n", warning)
		}
	}
}

// outputRegressionsConsole displays detected regressions in human-readable console format.
func outputRegressionsConsole(regressions map[string]interface{}) {
	fmt.Println("=== REGRESSION DETECTION ===")
	fmt.Println()
	printRegressionHeader(regressions)
	printRegressionList(regressions)
	fmt.Println()
}

func printRegressionHeader(regressions map[string]interface{}) {
	fmt.Printf("Threshold: %.1f%%\n", regressions["threshold"])
	fmt.Printf("Historical snapshots: %v\n", regressions["historical_count"])
	fmt.Printf("Recent snapshots: %v\n", regressions["recent_count"])
	fmt.Printf("Overall Severity: %v\n", regressions["severity"])
	fmt.Println()
}

func printRegressionList(regressions map[string]interface{}) {
	regList, ok := regressions["regressions"].([]map[string]interface{})
	if !ok {
		return
	}

	if len(regList) == 0 {
		fmt.Println("No significant regressions detected ✓")
		return
	}

	fmt.Printf("Detected %d regression(s):\n\n", len(regList))
	for i, reg := range regList {
		printRegressionItem(i+1, reg)
	}
}

func printRegressionItem(index int, reg map[string]interface{}) {
	indicator := getRegressionIndicator(reg["classification"])
	fmt.Printf("%d. [%s] %s %s\n", index, reg["severity"], indicator, reg["metric"])
	fmt.Printf("   Current: %.2f  |  Expected: %.2f  |  Deviation: %.1f%%\n",
		reg["current_value"], reg["expected_value"], reg["percent_deviation"])
	fmt.Printf("   P-value: %.4f\n", reg["p_value"])
	fmt.Println()
}

func getRegressionIndicator(classification interface{}) string {
	switch classification {
	case "regression":
		return "▼"
	case "improvement":
		return "▲"
	default:
		return "→"
	}
}
