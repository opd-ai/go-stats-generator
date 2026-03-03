package analyzer

import (
	"math"

	"github.com/opd-ai/go-stats-generator/internal/metrics"
)

// ComputeLinearRegression performs least-squares linear regression on time series data.
// Returns regression coefficients (slope, intercept) and goodness of fit (R²).
func ComputeLinearRegression(series metrics.MetricTimeSeries) metrics.TrendStatistics {
	n := len(series.DataPoints)
	if n < 2 {
		return metrics.TrendStatistics{DataPoints: n}
	}

	// Convert timestamps to days since first observation
	var x, y []float64
	firstTime := series.DataPoints[0].Timestamp
	for _, pt := range series.DataPoints {
		days := pt.Timestamp.Sub(firstTime).Hours() / 24.0
		x = append(x, days)
		y = append(y, pt.Value)
	}

	// Compute means
	meanX, meanY := mean(x), mean(y)

	// Compute slope and intercept using least squares
	var numerator, denominator float64
	for i := range x {
		numerator += (x[i] - meanX) * (y[i] - meanY)
		denominator += (x[i] - meanX) * (x[i] - meanX)
	}

	slope := 0.0
	if denominator != 0 {
		slope = numerator / denominator
	}
	intercept := meanY - slope*meanX

	// Compute R² (coefficient of determination)
	rSquared := computeRSquared(x, y, slope, intercept, meanY)

	// Compute Pearson correlation coefficient
	correlation := 0.0
	if rSquared >= 0 {
		correlation = math.Sqrt(rSquared)
		if slope < 0 {
			correlation = -correlation
		}
	}

	return metrics.TrendStatistics{
		Slope:       slope,
		Intercept:   intercept,
		RSquared:    rSquared,
		DataPoints:  n,
		StartDate:   series.DataPoints[0].Timestamp.Format("2006-01-02"),
		EndDate:     series.DataPoints[n-1].Timestamp.Format("2006-01-02"),
		Correlation: correlation,
	}
}

// computeRSquared calculates coefficient of determination (R²).
func computeRSquared(x, y []float64, slope, intercept, meanY float64) float64 {
	var ssRes, ssTot float64
	for i := range x {
		predicted := slope*x[i] + intercept
		ssRes += (y[i] - predicted) * (y[i] - predicted)
		ssTot += (y[i] - meanY) * (y[i] - meanY)
	}

	if ssTot == 0 {
		return 0
	}
	return 1 - (ssRes / ssTot)
}

// mean calculates arithmetic mean of a slice.
func mean(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}
