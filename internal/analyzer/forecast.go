package analyzer

import (
	"fmt"
	"math"

	"github.com/opd-ai/go-stats-generator/internal/metrics"
)

// ForecastMethod specifies the forecasting algorithm to use.
type ForecastMethod string

const (
	// ForecastLinear uses linear regression for forecasting.
	ForecastLinear ForecastMethod = "linear"
	// ForecastExponential uses exponential smoothing for forecasting.
	ForecastExponential ForecastMethod = "exponential"
)

// GenerateForecastWithMethod predicts future metric values using the specified method.
// Supported methods: "linear" (default), "exponential" (exponential smoothing).
func GenerateForecastWithMethod(series metrics.MetricTimeSeries, daysAhead int, method ForecastMethod) metrics.ForecastResult {
	switch method {
	case ForecastExponential:
		return generateExponentialForecast(series, daysAhead)
	default:
		return GenerateForecast(series, daysAhead)
	}
}

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

// generateExponentialForecast predicts future values using simple exponential smoothing.
// This method weights recent observations more heavily than older ones.
func generateExponentialForecast(series metrics.MetricTimeSeries, daysAhead int) metrics.ForecastResult {
	n := len(series.DataPoints)
	if n < 2 {
		return metrics.ForecastResult{
			MetricName: series.MetricName,
			Warning:    "Insufficient data for forecasting (need at least 2 points)",
		}
	}

	// Optimal alpha for simple exponential smoothing
	alpha := computeOptimalAlpha(series)

	// Compute smoothed values
	smoothed := computeSmoothedValues(series, alpha)
	pointEstimate := smoothed[n-1]

	// Calculate forecast date
	lastPoint := series.DataPoints[n-1]
	forecastDate := lastPoint.Timestamp.AddDate(0, 0, daysAhead)

	// Calculate standard error from smoothing residuals
	stdError := computeSmoothingError(series, smoothed)
	confidenceLevel := 0.95
	marginOfError := 1.96 * stdError

	// Minimum margin for very stable series
	if marginOfError < math.Abs(pointEstimate)*0.02 {
		marginOfError = math.Abs(pointEstimate) * 0.05
	}

	// Create trend statistics from smoothed data
	stats := buildSmoothingStats(series, alpha, smoothed)

	return metrics.ForecastResult{
		MetricName:       series.MetricName,
		PredictionDate:   forecastDate,
		PointEstimate:    pointEstimate,
		LowerBound:       pointEstimate - marginOfError,
		UpperBound:       pointEstimate + marginOfError,
		ConfidenceLevel:  confidenceLevel,
		TrendStatistics:  stats,
		ReliabilityScore: 1 - (stdError / (stdError + math.Abs(pointEstimate) + 0.001)),
	}
}

// computeOptimalAlpha finds the best smoothing parameter using grid search.
func computeOptimalAlpha(series metrics.MetricTimeSeries) float64 {
	bestAlpha := 0.3
	bestMSE := math.MaxFloat64

	// Grid search over alpha values
	for alpha := 0.1; alpha <= 0.9; alpha += 0.1 {
		mse := computeSmothingMSE(series, alpha)
		if mse < bestMSE {
			bestMSE = mse
			bestAlpha = alpha
		}
	}

	return bestAlpha
}

// computeSmothing computes mean squared error for a given alpha.
func computeSmothingMSE(series metrics.MetricTimeSeries, alpha float64) float64 {
	smoothed := computeSmoothedValues(series, alpha)
	var sumSq float64
	for i := 1; i < len(series.DataPoints); i++ {
		residual := series.DataPoints[i].Value - smoothed[i-1]
		sumSq += residual * residual
	}
	return sumSq / float64(len(series.DataPoints)-1)
}

// computeSmoothedValues applies simple exponential smoothing to the series.
func computeSmoothedValues(series metrics.MetricTimeSeries, alpha float64) []float64 {
	n := len(series.DataPoints)
	smoothed := make([]float64, n)
	smoothed[0] = series.DataPoints[0].Value

	for i := 1; i < n; i++ {
		smoothed[i] = alpha*series.DataPoints[i].Value + (1-alpha)*smoothed[i-1]
	}

	return smoothed
}

// computeSmoothingError computes standard error from smoothing residuals.
func computeSmoothingError(series metrics.MetricTimeSeries, smoothed []float64) float64 {
	n := len(series.DataPoints)
	if n <= 2 {
		return 0
	}

	var sumSq float64
	for i := 1; i < n; i++ {
		residual := series.DataPoints[i].Value - smoothed[i-1]
		sumSq += residual * residual
	}

	return math.Sqrt(sumSq / float64(n-1))
}

// buildSmoothingStats creates TrendStatistics for exponential smoothing.
func buildSmoothingStats(series metrics.MetricTimeSeries, alpha float64, smoothed []float64) metrics.TrendStatistics {
	n := len(series.DataPoints)

	// Approximate slope from smoothed values
	var slope float64
	if n >= 2 {
		firstTime := series.DataPoints[0].Timestamp
		lastTime := series.DataPoints[n-1].Timestamp
		daysBetween := lastTime.Sub(firstTime).Hours() / 24.0
		if daysBetween > 0 {
			slope = (smoothed[n-1] - smoothed[0]) / daysBetween
		}
	}

	return metrics.TrendStatistics{
		Slope:       slope,
		Intercept:   smoothed[0],
		RSquared:    alpha, // Use alpha as indicator of adaptiveness
		DataPoints:  n,
		StartDate:   series.DataPoints[0].Timestamp.Format("2006-01-02"),
		EndDate:     series.DataPoints[n-1].Timestamp.Format("2006-01-02"),
		Correlation: 0, // Not applicable for exponential smoothing
	}
}
