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

func TestComputePearsonCorrelation_PerfectPositive(t *testing.T) {
	baseTime := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	series1 := metrics.MetricTimeSeries{
		MetricName: "metric_a",
		DataPoints: []metrics.TimeSeriesPoint{
			{Timestamp: baseTime, Value: 1},
			{Timestamp: baseTime.AddDate(0, 0, 1), Value: 2},
			{Timestamp: baseTime.AddDate(0, 0, 2), Value: 3},
			{Timestamp: baseTime.AddDate(0, 0, 3), Value: 4},
			{Timestamp: baseTime.AddDate(0, 0, 4), Value: 5},
		},
	}
	series2 := metrics.MetricTimeSeries{
		MetricName: "metric_b",
		DataPoints: []metrics.TimeSeriesPoint{
			{Timestamp: baseTime, Value: 2},
			{Timestamp: baseTime.AddDate(0, 0, 1), Value: 4},
			{Timestamp: baseTime.AddDate(0, 0, 2), Value: 6},
			{Timestamp: baseTime.AddDate(0, 0, 3), Value: 8},
			{Timestamp: baseTime.AddDate(0, 0, 4), Value: 10},
		},
	}

	result := ComputePearsonCorrelation(series1, series2)

	if math.Abs(result.Correlation-1.0) > 0.001 {
		t.Errorf("Expected correlation=1.0, got %.3f", result.Correlation)
	}
	if result.Strength != "strong" {
		t.Errorf("Expected strength='strong', got '%s'", result.Strength)
	}
	if result.Direction != "positive" {
		t.Errorf("Expected direction='positive', got '%s'", result.Direction)
	}
	if result.DataPoints != 5 {
		t.Errorf("Expected DataPoints=5, got %d", result.DataPoints)
	}
}

func TestComputePearsonCorrelation_PerfectNegative(t *testing.T) {
	baseTime := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	series1 := metrics.MetricTimeSeries{
		MetricName: "metric_a",
		DataPoints: []metrics.TimeSeriesPoint{
			{Timestamp: baseTime, Value: 1},
			{Timestamp: baseTime.AddDate(0, 0, 1), Value: 2},
			{Timestamp: baseTime.AddDate(0, 0, 2), Value: 3},
			{Timestamp: baseTime.AddDate(0, 0, 3), Value: 4},
		},
	}
	series2 := metrics.MetricTimeSeries{
		MetricName: "metric_b",
		DataPoints: []metrics.TimeSeriesPoint{
			{Timestamp: baseTime, Value: 10},
			{Timestamp: baseTime.AddDate(0, 0, 1), Value: 8},
			{Timestamp: baseTime.AddDate(0, 0, 2), Value: 6},
			{Timestamp: baseTime.AddDate(0, 0, 3), Value: 4},
		},
	}

	result := ComputePearsonCorrelation(series1, series2)

	if math.Abs(result.Correlation+1.0) > 0.001 {
		t.Errorf("Expected correlation=-1.0, got %.3f", result.Correlation)
	}
	if result.Strength != "strong" {
		t.Errorf("Expected strength='strong', got '%s'", result.Strength)
	}
	if result.Direction != "negative" {
		t.Errorf("Expected direction='negative', got '%s'", result.Direction)
	}
}

func TestComputePearsonCorrelation_InsufficientData(t *testing.T) {
	baseTime := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	series1 := metrics.MetricTimeSeries{
		MetricName: "metric_a",
		DataPoints: []metrics.TimeSeriesPoint{
			{Timestamp: baseTime, Value: 1},
			{Timestamp: baseTime.AddDate(0, 0, 1), Value: 2},
		},
	}
	series2 := metrics.MetricTimeSeries{
		MetricName: "metric_b",
		DataPoints: []metrics.TimeSeriesPoint{
			{Timestamp: baseTime, Value: 5},
			{Timestamp: baseTime.AddDate(0, 0, 1), Value: 6},
		},
	}

	result := ComputePearsonCorrelation(series1, series2)

	if result.Strength != "insufficient_data" {
		t.Errorf("Expected strength='insufficient_data', got '%s'", result.Strength)
	}
	if result.DataPoints != 2 {
		t.Errorf("Expected DataPoints=2, got %d", result.DataPoints)
	}
}

func TestComputeCorrelationMatrix(t *testing.T) {
	baseTime := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	seriesList := []metrics.MetricTimeSeries{
		{
			MetricName: "metric_a",
			DataPoints: []metrics.TimeSeriesPoint{
				{Timestamp: baseTime, Value: 1},
				{Timestamp: baseTime.AddDate(0, 0, 1), Value: 2},
				{Timestamp: baseTime.AddDate(0, 0, 2), Value: 3},
			},
		},
		{
			MetricName: "metric_b",
			DataPoints: []metrics.TimeSeriesPoint{
				{Timestamp: baseTime, Value: 2},
				{Timestamp: baseTime.AddDate(0, 0, 1), Value: 4},
				{Timestamp: baseTime.AddDate(0, 0, 2), Value: 6},
			},
		},
		{
			MetricName: "metric_c",
			DataPoints: []metrics.TimeSeriesPoint{
				{Timestamp: baseTime, Value: 10},
				{Timestamp: baseTime.AddDate(0, 0, 1), Value: 8},
				{Timestamp: baseTime.AddDate(0, 0, 2), Value: 6},
			},
		},
	}

	result := ComputeCorrelationMatrix(seriesList)

	if len(result.Metrics) != 3 {
		t.Errorf("Expected 3 metrics, got %d", len(result.Metrics))
	}
	// 3 metrics should produce 3 pairwise correlations: (a,b), (a,c), (b,c)
	if len(result.Correlations) != 3 {
		t.Errorf("Expected 3 correlations, got %d", len(result.Correlations))
	}
	if result.DataPoints != 3 {
		t.Errorf("Expected DataPoints=3, got %d", result.DataPoints)
	}
}

func TestClassifyCorrelationStrength(t *testing.T) {
	tests := []struct {
		r        float64
		expected string
	}{
		{0.9, "strong"},
		{-0.85, "strong"},
		{0.5, "moderate"},
		{-0.45, "moderate"},
		{0.3, "weak"},
		{-0.25, "weak"},
		{0.1, "none"},
		{0.0, "none"},
	}

	for _, tt := range tests {
		result := classifyCorrelationStrength(tt.r)
		if result != tt.expected {
			t.Errorf("classifyCorrelationStrength(%.2f) = '%s', expected '%s'", tt.r, result, tt.expected)
		}
	}
}

func TestClassifyCorrelationDirection(t *testing.T) {
	tests := []struct {
		r        float64
		expected string
	}{
		{0.5, "positive"},
		{0.15, "positive"},
		{-0.5, "negative"},
		{-0.2, "negative"},
		{0.05, "none"},
		{0.0, "none"},
		{-0.08, "none"},
	}

	for _, tt := range tests {
		result := classifyCorrelationDirection(tt.r)
		if result != tt.expected {
			t.Errorf("classifyCorrelationDirection(%.2f) = '%s', expected '%s'", tt.r, result, tt.expected)
		}
	}
}
