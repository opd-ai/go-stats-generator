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
