package analyzer

import (
	"math"

	"github.com/opd-ai/go-stats-generator/internal/metrics"
)

// DetectRegressions analyzes whether recent metric values deviate from historical trends.
// threshold specifies the percent deviation that qualifies as a regression (e.g., 10.0 for 10%).
func DetectRegressions(series metrics.MetricTimeSeries, threshold float64) []metrics.RegressionResult {
	if len(series.DataPoints) < 3 {
		return nil
	}

	stats := ComputeLinearRegression(series)
	lastPoint := series.DataPoints[len(series.DataPoints)-1]

	// Calculate expected value based on trend
	daysSinceStart := lastPoint.Timestamp.Sub(series.DataPoints[0].Timestamp).Hours() / 24.0
	expectedValue := stats.Slope*daysSinceStart + stats.Intercept

	// Calculate percent deviation
	deviation := 0.0
	if expectedValue != 0 {
		deviation = ((lastPoint.Value - expectedValue) / math.Abs(expectedValue)) * 100
	}

	// Classify deviation
	classification := classifyDeviation(deviation, threshold)
	severity := calculateSeverity(math.Abs(deviation), threshold)

	// Simple significance test based on standard error
	pValue := calculatePValue(series, stats, lastPoint.Value)

	result := metrics.RegressionResult{
		MetricName:        series.MetricName,
		CurrentValue:      lastPoint.Value,
		ExpectedValue:     expectedValue,
		PercentDeviation:  deviation,
		Classification:    classification,
		Severity:          severity,
		PValue:            pValue,
		SignificanceLevel: 0.05,
		DetectedAt:        lastPoint.Timestamp,
		TrendStatistics:   stats,
	}

	return []metrics.RegressionResult{result}
}

// classifyDeviation determines if a deviation is a regression, improvement, or stable.
func classifyDeviation(deviation, threshold float64) string {
	absDeviation := math.Abs(deviation)
	if absDeviation < threshold {
		return "stable"
	}
	if deviation > 0 {
		return "regression" // Higher than expected (worse for metrics like complexity)
	}
	return "improvement" // Lower than expected (better)
}

// calculateSeverity assigns severity level based on deviation magnitude.
func calculateSeverity(absDeviation, threshold float64) string {
	if absDeviation < threshold {
		return "low"
	}
	if absDeviation < threshold*2 {
		return "medium"
	}
	if absDeviation < threshold*3 {
		return "high"
	}
	return "critical"
}

// calculatePValue estimates statistical significance of the deviation.
// Simplified approach: compares deviation to standard error.
func calculatePValue(series metrics.MetricTimeSeries, stats metrics.TrendStatistics, observedValue float64) float64 {
	stdError := calculateStandardError(series, stats)
	if stdError == 0 {
		return 0.0 // Perfect fit, any deviation is significant
	}

	lastPoint := series.DataPoints[len(series.DataPoints)-1]
	daysSinceStart := lastPoint.Timestamp.Sub(series.DataPoints[0].Timestamp).Hours() / 24.0
	expectedValue := stats.Slope*daysSinceStart + stats.Intercept

	// Z-score approximation
	zScore := math.Abs(observedValue-expectedValue) / stdError

	// Approximate p-value using complementary error function
	// For simplicity, use rough approximation
	if zScore < 1.0 {
		return 0.32 // Not significant
	}
	if zScore < 2.0 {
		return 0.05 // Marginally significant
	}
	return 0.01 // Significant
}
