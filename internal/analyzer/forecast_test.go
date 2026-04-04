package analyzer

import (
	"math"
	"testing"
	"time"

	"github.com/opd-ai/go-stats-generator/internal/metrics"
)

func TestGenerateForecast_LinearTrend(t *testing.T) {
	// Perfect linear trend: y = 2x + 5
	baseTime := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	series := metrics.MetricTimeSeries{
		MetricName: "complexity",
		DataPoints: []metrics.TimeSeriesPoint{
			{Timestamp: baseTime, Value: 5},
			{Timestamp: baseTime.AddDate(0, 0, 1), Value: 7},
			{Timestamp: baseTime.AddDate(0, 0, 2), Value: 9},
			{Timestamp: baseTime.AddDate(0, 0, 3), Value: 11},
		},
	}

	forecast := GenerateForecast(series, 7)

	// Expected value at day 10 (7 days after day 3): 2*10 + 5 = 25
	expectedValue := 25.0
	if math.Abs(forecast.PointEstimate-expectedValue) > 0.1 {
		t.Errorf("Expected forecast=%.1f, got %.1f", expectedValue, forecast.PointEstimate)
	}

	if forecast.Warning != "" {
		t.Errorf("Expected no warning for perfect fit, got: %s", forecast.Warning)
	}

	if forecast.ReliabilityScore < 0.99 {
		t.Errorf("Expected high reliability (R²≈1.0), got %.2f", forecast.ReliabilityScore)
	}

	// Check confidence interval exists
	if forecast.LowerBound >= forecast.PointEstimate {
		t.Errorf("Lower bound should be less than point estimate")
	}
	if forecast.UpperBound <= forecast.PointEstimate {
		t.Errorf("Upper bound should be greater than point estimate")
	}
}

func TestGenerateForecast_InsufficientData(t *testing.T) {
	baseTime := time.Now()
	series := metrics.MetricTimeSeries{
		MetricName: "sparse",
		DataPoints: []metrics.TimeSeriesPoint{
			{Timestamp: baseTime, Value: 10},
		},
	}

	forecast := GenerateForecast(series, 7)

	if forecast.Warning == "" {
		t.Error("Expected warning for insufficient data")
	}
	if forecast.PointEstimate != 0 {
		t.Errorf("Expected zero forecast for insufficient data, got %.2f", forecast.PointEstimate)
	}
}

func TestGenerateForecast_LowReliability(t *testing.T) {
	// Noisy data with weak trend
	baseTime := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	series := metrics.MetricTimeSeries{
		MetricName: "noisy_metric",
		DataPoints: []metrics.TimeSeriesPoint{
			{Timestamp: baseTime, Value: 5},
			{Timestamp: baseTime.AddDate(0, 0, 1), Value: 15},
			{Timestamp: baseTime.AddDate(0, 0, 2), Value: 7},
			{Timestamp: baseTime.AddDate(0, 0, 3), Value: 20},
			{Timestamp: baseTime.AddDate(0, 0, 4), Value: 10},
		},
	}

	forecast := GenerateForecast(series, 7)

	// Should warn about low R²
	if forecast.Warning == "" {
		t.Log("R² value:", forecast.ReliabilityScore)
		if forecast.ReliabilityScore < 0.5 {
			t.Error("Expected warning for low R²")
		}
	}
}

func TestGenerateForecast_DateProgression(t *testing.T) {
	baseTime := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	series := metrics.MetricTimeSeries{
		MetricName: "test",
		DataPoints: []metrics.TimeSeriesPoint{
			{Timestamp: baseTime, Value: 10},
			{Timestamp: baseTime.AddDate(0, 0, 1), Value: 12},
			{Timestamp: baseTime.AddDate(0, 0, 2), Value: 14},
		},
	}

	forecast := GenerateForecast(series, 5)

	expectedDate := baseTime.AddDate(0, 0, 7) // Last point (day 2) + 5 days
	if !forecast.PredictionDate.Equal(expectedDate) {
		t.Errorf("Expected prediction date %v, got %v", expectedDate, forecast.PredictionDate)
	}
}

func TestGenerateForecastWithMethod_Linear(t *testing.T) {
	baseTime := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	series := metrics.MetricTimeSeries{
		MetricName: "complexity",
		DataPoints: []metrics.TimeSeriesPoint{
			{Timestamp: baseTime, Value: 5},
			{Timestamp: baseTime.AddDate(0, 0, 1), Value: 7},
			{Timestamp: baseTime.AddDate(0, 0, 2), Value: 9},
		},
	}

	forecast := GenerateForecastWithMethod(series, 7, ForecastLinear)

	// Should produce same result as GenerateForecast
	directForecast := GenerateForecast(series, 7)

	if math.Abs(forecast.PointEstimate-directForecast.PointEstimate) > 0.01 {
		t.Errorf("Linear method mismatch: got %.2f, expected %.2f",
			forecast.PointEstimate, directForecast.PointEstimate)
	}
}

func TestGenerateForecastWithMethod_Exponential(t *testing.T) {
	baseTime := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	series := metrics.MetricTimeSeries{
		MetricName: "complexity",
		DataPoints: []metrics.TimeSeriesPoint{
			{Timestamp: baseTime, Value: 10},
			{Timestamp: baseTime.AddDate(0, 0, 1), Value: 12},
			{Timestamp: baseTime.AddDate(0, 0, 2), Value: 11},
			{Timestamp: baseTime.AddDate(0, 0, 3), Value: 15},
			{Timestamp: baseTime.AddDate(0, 0, 4), Value: 14},
		},
	}

	forecast := GenerateForecastWithMethod(series, 7, ForecastExponential)

	// Verify basic structure
	if forecast.MetricName != "complexity" {
		t.Errorf("Expected metric name 'complexity', got %s", forecast.MetricName)
	}

	// Point estimate should be reasonable (between min and max with some margin)
	if forecast.PointEstimate < 5 || forecast.PointEstimate > 25 {
		t.Errorf("Exponential forecast out of expected range: %.2f", forecast.PointEstimate)
	}

	// Should have confidence interval
	if forecast.LowerBound >= forecast.UpperBound {
		t.Error("Expected lower bound < upper bound")
	}

	// Reliability score should be in valid range
	if forecast.ReliabilityScore < 0 || forecast.ReliabilityScore > 1 {
		t.Errorf("Reliability score out of range: %.2f", forecast.ReliabilityScore)
	}
}

func TestGenerateForecastWithMethod_ExponentialInsufficientData(t *testing.T) {
	baseTime := time.Now()
	series := metrics.MetricTimeSeries{
		MetricName: "sparse",
		DataPoints: []metrics.TimeSeriesPoint{
			{Timestamp: baseTime, Value: 10},
		},
	}

	forecast := GenerateForecastWithMethod(series, 7, ForecastExponential)

	if forecast.Warning == "" {
		t.Error("Expected warning for insufficient data")
	}
}

func TestComputeOptimalAlpha(t *testing.T) {
	baseTime := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	series := metrics.MetricTimeSeries{
		MetricName: "test",
		DataPoints: []metrics.TimeSeriesPoint{
			{Timestamp: baseTime, Value: 10},
			{Timestamp: baseTime.AddDate(0, 0, 1), Value: 11},
			{Timestamp: baseTime.AddDate(0, 0, 2), Value: 12},
			{Timestamp: baseTime.AddDate(0, 0, 3), Value: 13},
		},
	}

	alpha := computeOptimalAlpha(series)

	// Alpha should be in valid range [0.1, 0.9]
	if alpha < 0.1 || alpha > 0.9 {
		t.Errorf("Alpha out of expected range: %.2f", alpha)
	}
}

func TestComputeSmoothedValues(t *testing.T) {
	baseTime := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	series := metrics.MetricTimeSeries{
		MetricName: "test",
		DataPoints: []metrics.TimeSeriesPoint{
			{Timestamp: baseTime, Value: 10},
			{Timestamp: baseTime.AddDate(0, 0, 1), Value: 20},
			{Timestamp: baseTime.AddDate(0, 0, 2), Value: 15},
		},
	}

	alpha := 0.5
	smoothed := computeSmoothedValues(series, alpha)

	// First value should equal first data point
	if smoothed[0] != 10 {
		t.Errorf("First smoothed value should be 10, got %.2f", smoothed[0])
	}

	// Second value: 0.5*20 + 0.5*10 = 15
	if math.Abs(smoothed[1]-15) > 0.01 {
		t.Errorf("Second smoothed value should be 15, got %.2f", smoothed[1])
	}

	// Third value: 0.5*15 + 0.5*15 = 15
	if math.Abs(smoothed[2]-15) > 0.01 {
		t.Errorf("Third smoothed value should be 15, got %.2f", smoothed[2])
	}
}
