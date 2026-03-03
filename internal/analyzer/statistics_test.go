package analyzer

import (
	"math"
	"testing"
	"time"

	"github.com/opd-ai/go-stats-generator/internal/metrics"
)

func TestComputeLinearRegression_PerfectLinear(t *testing.T) {
	// Perfect linear relationship: y = 2x + 5
	baseTime := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	series := metrics.MetricTimeSeries{
		MetricName: "test_metric",
		DataPoints: []metrics.TimeSeriesPoint{
			{Timestamp: baseTime, Value: 5},
			{Timestamp: baseTime.AddDate(0, 0, 1), Value: 7},
			{Timestamp: baseTime.AddDate(0, 0, 2), Value: 9},
			{Timestamp: baseTime.AddDate(0, 0, 3), Value: 11},
			{Timestamp: baseTime.AddDate(0, 0, 4), Value: 13},
		},
	}

	result := ComputeLinearRegression(series)

	if math.Abs(result.Slope-2.0) > 0.001 {
		t.Errorf("Expected slope=2.0, got %.3f", result.Slope)
	}
	if math.Abs(result.Intercept-5.0) > 0.001 {
		t.Errorf("Expected intercept=5.0, got %.3f", result.Intercept)
	}
	if math.Abs(result.RSquared-1.0) > 0.001 {
		t.Errorf("Expected R²=1.0 for perfect fit, got %.3f", result.RSquared)
	}
	if result.DataPoints != 5 {
		t.Errorf("Expected 5 data points, got %d", result.DataPoints)
	}
	if math.Abs(result.Correlation-1.0) > 0.001 {
		t.Errorf("Expected correlation=1.0, got %.3f", result.Correlation)
	}
}

func TestComputeLinearRegression_NegativeSlope(t *testing.T) {
	// Decreasing trend: y = -1.5x + 10
	baseTime := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	series := metrics.MetricTimeSeries{
		MetricName: "declining_metric",
		DataPoints: []metrics.TimeSeriesPoint{
			{Timestamp: baseTime, Value: 10.0},
			{Timestamp: baseTime.AddDate(0, 0, 1), Value: 8.5},
			{Timestamp: baseTime.AddDate(0, 0, 2), Value: 7.0},
			{Timestamp: baseTime.AddDate(0, 0, 3), Value: 5.5},
		},
	}

	result := ComputeLinearRegression(series)

	if math.Abs(result.Slope+1.5) > 0.001 {
		t.Errorf("Expected slope=-1.5, got %.3f", result.Slope)
	}
	if result.Correlation >= 0 {
		t.Errorf("Expected negative correlation for negative slope, got %.3f", result.Correlation)
	}
}

func TestComputeLinearRegression_InsufficientData(t *testing.T) {
	baseTime := time.Now()
	series := metrics.MetricTimeSeries{
		MetricName: "sparse_metric",
		DataPoints: []metrics.TimeSeriesPoint{
			{Timestamp: baseTime, Value: 10},
		},
	}

	result := ComputeLinearRegression(series)

	if result.DataPoints != 1 {
		t.Errorf("Expected DataPoints=1, got %d", result.DataPoints)
	}
	if result.Slope != 0 || result.Intercept != 0 {
		t.Errorf("Expected zero coefficients for insufficient data")
	}
}

func TestComputeLinearRegression_ConstantValue(t *testing.T) {
	// No variation: y = 5 (constant)
	baseTime := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	series := metrics.MetricTimeSeries{
		MetricName: "constant_metric",
		DataPoints: []metrics.TimeSeriesPoint{
			{Timestamp: baseTime, Value: 5},
			{Timestamp: baseTime.AddDate(0, 0, 1), Value: 5},
			{Timestamp: baseTime.AddDate(0, 0, 2), Value: 5},
		},
	}

	result := ComputeLinearRegression(series)

	if math.Abs(result.Slope) > 0.001 {
		t.Errorf("Expected slope=0 for constant values, got %.3f", result.Slope)
	}
	if math.Abs(result.Intercept-5.0) > 0.001 {
		t.Errorf("Expected intercept=5.0, got %.3f", result.Intercept)
	}
}

func TestMean(t *testing.T) {
	tests := []struct {
		name     string
		values   []float64
		expected float64
	}{
		{"empty", []float64{}, 0},
		{"single", []float64{5}, 5},
		{"positive", []float64{1, 2, 3, 4, 5}, 3},
		{"mixed", []float64{-2, 0, 2}, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mean(tt.values)
			if math.Abs(result-tt.expected) > 0.001 {
				t.Errorf("mean(%v) = %.3f, expected %.3f", tt.values, result, tt.expected)
			}
		})
	}
}
