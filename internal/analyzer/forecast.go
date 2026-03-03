package analyzer

import (
	"fmt"
	"math"

	"github.com/opd-ai/go-stats-generator/internal/metrics"
)

// GenerateForecast predicts future metric values using linear regression.
// daysAhead specifies how many days into the future to forecast.
func GenerateForecast(series metrics.MetricTimeSeries, daysAhead int) metrics.ForecastResult {
	stats := ComputeLinearRegression(series)

	if stats.DataPoints < 2 {
		return metrics.ForecastResult{
			MetricName: series.MetricName,
			Warning:    "Insufficient data for forecasting (need at least 2 points)",
		}
	}

	// Calculate forecast date and point estimate
	lastPoint := series.DataPoints[len(series.DataPoints)-1]
	forecastDate := lastPoint.Timestamp.AddDate(0, 0, daysAhead)
	daysSinceStart := forecastDate.Sub(series.DataPoints[0].Timestamp).Hours() / 24.0
	pointEstimate := stats.Slope*daysSinceStart + stats.Intercept

	// Calculate standard error and confidence interval
	stdError := calculateStandardError(series, stats)
	confidenceLevel := 0.95
	marginOfError := 1.96 * stdError // 95% CI uses z=1.96

	// For perfect fit (R²=1.0), use small nominal margin
	if stdError < 0.01 {
		marginOfError = math.Abs(pointEstimate) * 0.05 // 5% of point estimate
	}

	result := metrics.ForecastResult{
		MetricName:       series.MetricName,
		PredictionDate:   forecastDate,
		PointEstimate:    pointEstimate,
		LowerBound:       pointEstimate - marginOfError,
		UpperBound:       pointEstimate + marginOfError,
		ConfidenceLevel:  confidenceLevel,
		TrendStatistics:  stats,
		ReliabilityScore: stats.RSquared,
	}

	// Warn if low R²
	if stats.RSquared < 0.5 {
		result.Warning = fmt.Sprintf("Low reliability (R²=%.2f): forecast may be inaccurate", stats.RSquared)
	}

	return result
}

// calculateStandardError computes standard error of the estimate for regression.
func calculateStandardError(series metrics.MetricTimeSeries, stats metrics.TrendStatistics) float64 {
	n := len(series.DataPoints)
	if n <= 2 {
		return 0
	}

	firstTime := series.DataPoints[0].Timestamp
	var sumSquaredResiduals float64

	for _, pt := range series.DataPoints {
		days := pt.Timestamp.Sub(firstTime).Hours() / 24.0
		predicted := stats.Slope*days + stats.Intercept
		residual := pt.Value - predicted
		sumSquaredResiduals += residual * residual
	}

	variance := sumSquaredResiduals / float64(n-2)
	return math.Sqrt(variance)
}
