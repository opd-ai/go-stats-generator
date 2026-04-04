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
	// ForecastARIMA uses ARIMA(1,1,1) for forecasting.
	ForecastARIMA ForecastMethod = "arima"
)

// GenerateForecastWithMethod predicts future metric values using the specified method.
// Supported methods: "linear" (default), "exponential" (exponential smoothing), "arima".
func GenerateForecastWithMethod(series metrics.MetricTimeSeries, daysAhead int, method ForecastMethod) metrics.ForecastResult {
	switch method {
	case ForecastExponential:
		return generateExponentialForecast(series, daysAhead)
	case ForecastARIMA:
		return generateARIMAForecast(series, daysAhead)
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

// generateARIMAForecast predicts future values using a simplified ARIMA(1,1,1) model.
// ARIMA combines autoregressive (AR), integrated (I), and moving average (MA) components.
// This implementation uses first-order differencing with AR(1) and MA(1) terms.
func generateARIMAForecast(series metrics.MetricTimeSeries, daysAhead int) metrics.ForecastResult {
	n := len(series.DataPoints)
	if n < 4 {
		return metrics.ForecastResult{
			MetricName: series.MetricName,
			Warning:    "Insufficient data for ARIMA forecasting (need at least 4 points)",
		}
	}

	// First-order differencing (I=1)
	diff := computeDifferences(series)
	if len(diff) < 3 {
		return metrics.ForecastResult{
			MetricName: series.MetricName,
			Warning:    "Insufficient differenced data for ARIMA",
		}
	}

	// Estimate AR(1) and MA(1) coefficients
	phi, theta := estimateARIMACoefficients(diff)

	// Compute residuals and forecast
	residuals := computeARIMAResiduals(diff, phi, theta)
	lastValue := series.DataPoints[n-1].Value
	lastDiff := diff[len(diff)-1]
	lastResidual := 0.0
	if len(residuals) > 0 {
		lastResidual = residuals[len(residuals)-1]
	}

	// Forecast differenced value
	forecastDiff := phi*lastDiff + theta*lastResidual
	pointEstimate := lastValue + forecastDiff*float64(daysAhead)

	// Calculate forecast date
	forecastDate := series.DataPoints[n-1].Timestamp.AddDate(0, 0, daysAhead)

	// Calculate standard error
	stdError := computeARIMAStdError(residuals)
	marginOfError := 1.96 * stdError * math.Sqrt(float64(daysAhead))

	// Minimum margin for very stable series
	if marginOfError < math.Abs(pointEstimate)*0.02 {
		marginOfError = math.Abs(pointEstimate) * 0.05
	}

	// Create trend statistics
	stats := buildARIMAStats(series, phi, theta, residuals)

	return metrics.ForecastResult{
		MetricName:       series.MetricName,
		PredictionDate:   forecastDate,
		PointEstimate:    pointEstimate,
		LowerBound:       pointEstimate - marginOfError,
		UpperBound:       pointEstimate + marginOfError,
		ConfidenceLevel:  0.95,
		TrendStatistics:  stats,
		ReliabilityScore: computeARIMAReliability(residuals, diff),
	}
}

// computeDifferences calculates first-order differences for ARIMA.
func computeDifferences(series metrics.MetricTimeSeries) []float64 {
	n := len(series.DataPoints)
	if n < 2 {
		return nil
	}

	diff := make([]float64, n-1)
	for i := 1; i < n; i++ {
		diff[i-1] = series.DataPoints[i].Value - series.DataPoints[i-1].Value
	}
	return diff
}

// estimateARIMACoefficients estimates AR(1) and MA(1) coefficients using Yule-Walker equations.
func estimateARIMACoefficients(diff []float64) (phi, theta float64) {
	n := len(diff)
	if n < 2 {
		return 0, 0
	}

	// Estimate AR(1) coefficient phi using autocorrelation
	meanDiff := mean(diff)
	var gamma0, gamma1 float64

	for i := 0; i < n; i++ {
		dev := diff[i] - meanDiff
		gamma0 += dev * dev
	}
	for i := 1; i < n; i++ {
		gamma1 += (diff[i] - meanDiff) * (diff[i-1] - meanDiff)
	}

	if gamma0 > 0 {
		phi = gamma1 / gamma0
	}

	// Clamp phi to ensure stationarity
	if phi > 0.99 {
		phi = 0.99
	} else if phi < -0.99 {
		phi = -0.99
	}

	// Estimate MA(1) coefficient theta using innovation method
	// Simplified: use a fraction of the AR coefficient
	theta = -0.5 * phi
	if theta > 0.99 {
		theta = 0.99
	} else if theta < -0.99 {
		theta = -0.99
	}

	return phi, theta
}

// computeARIMAResiduals calculates residuals from the ARIMA model.
func computeARIMAResiduals(diff []float64, phi, theta float64) []float64 {
	n := len(diff)
	if n < 2 {
		return nil
	}

	residuals := make([]float64, n)
	residuals[0] = diff[0]

	for i := 1; i < n; i++ {
		predicted := phi * diff[i-1]
		if i > 1 {
			predicted += theta * residuals[i-1]
		}
		residuals[i] = diff[i] - predicted
	}

	return residuals
}

// computeARIMAStdError calculates standard error from ARIMA residuals.
func computeARIMAStdError(residuals []float64) float64 {
	n := len(residuals)
	if n <= 2 {
		return 0
	}

	var sumSq float64
	for _, r := range residuals {
		sumSq += r * r
	}

	return math.Sqrt(sumSq / float64(n-2))
}

// computeARIMAReliability estimates model reliability based on residual variance.
func computeARIMAReliability(residuals, diff []float64) float64 {
	if len(diff) == 0 || len(residuals) == 0 {
		return 0
	}

	// Calculate variance ratio (1 - residual variance / original variance)
	diffVar := variance(diff)
	resVar := variance(residuals)

	if diffVar == 0 {
		return 0
	}

	reliability := 1 - (resVar / diffVar)
	if reliability < 0 {
		reliability = 0
	}
	if reliability > 1 {
		reliability = 1
	}

	return reliability
}

// variance calculates the variance of a slice.
func variance(values []float64) float64 {
	n := len(values)
	if n < 2 {
		return 0
	}

	m := mean(values)
	var sumSq float64
	for _, v := range values {
		dev := v - m
		sumSq += dev * dev
	}

	return sumSq / float64(n-1)
}

// buildARIMAStats creates TrendStatistics for ARIMA model.
func buildARIMAStats(series metrics.MetricTimeSeries, phi, theta float64, residuals []float64) metrics.TrendStatistics {
	n := len(series.DataPoints)

	// Approximate slope from differenced values mean
	diff := computeDifferences(series)
	slope := 0.0
	if len(diff) > 0 {
		slope = mean(diff)
	}

	// R-squared approximation from residuals
	rSquared := computeARIMAReliability(residuals, diff)

	return metrics.TrendStatistics{
		Slope:       slope,
		Intercept:   series.DataPoints[0].Value,
		RSquared:    rSquared,
		DataPoints:  n,
		StartDate:   series.DataPoints[0].Timestamp.Format("2006-01-02"),
		EndDate:     series.DataPoints[n-1].Timestamp.Format("2006-01-02"),
		Correlation: phi, // Use AR coefficient as correlation indicator
	}
}
