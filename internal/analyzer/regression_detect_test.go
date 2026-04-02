package analyzer

import (
	"math"
	"testing"
	"time"

	"github.com/opd-ai/go-stats-generator/internal/metrics"
)

func TestDetectRegressions_StableTrend(t *testing.T) {
	baseTime := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	series := metrics.MetricTimeSeries{
		MetricName: "complexity",
		DataPoints: []metrics.TimeSeriesPoint{
			{Timestamp: baseTime, Value: 10},
			{Timestamp: baseTime.AddDate(0, 0, 1), Value: 12},
			{Timestamp: baseTime.AddDate(0, 0, 2), Value: 14},
			{Timestamp: baseTime.AddDate(0, 0, 3), Value: 16},
		},
	}

	results := DetectRegressions(series, 10.0)

	if len(results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(results))
	}

	result := results[0]
	if result.Classification != "stable" {
		t.Errorf("Expected stable classification, got %s", result.Classification)
	}
	if result.Severity != metrics.SeverityLevelInfo {
		t.Errorf("Expected low severity for stable trend, got %s", result.Severity)
	}
}

func TestDetectRegressions_SignificantIncrease(t *testing.T) {
	baseTime := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	series := metrics.MetricTimeSeries{
		MetricName: "complexity",
		DataPoints: []metrics.TimeSeriesPoint{
			{Timestamp: baseTime, Value: 10},
			{Timestamp: baseTime.AddDate(0, 0, 1), Value: 11},
			{Timestamp: baseTime.AddDate(0, 0, 2), Value: 12},
			{Timestamp: baseTime.AddDate(0, 0, 3), Value: 20}, // Spike
		},
	}

	results := DetectRegressions(series, 10.0)

	if len(results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(results))
	}

	result := results[0]
	if result.Classification != "regression" {
		t.Errorf("Expected regression classification, got %s", result.Classification)
	}
	if math.Abs(result.PercentDeviation) < 10 {
		t.Errorf("Expected significant deviation, got %.2f%%", result.PercentDeviation)
	}
	if result.Severity == "low" {
		t.Errorf("Expected higher severity for large deviation, got %s", result.Severity)
	}
}

func TestDetectRegressions_Improvement(t *testing.T) {
	baseTime := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	series := metrics.MetricTimeSeries{
		MetricName: "complexity",
		DataPoints: []metrics.TimeSeriesPoint{
			{Timestamp: baseTime, Value: 20},
			{Timestamp: baseTime.AddDate(0, 0, 1), Value: 19},
			{Timestamp: baseTime.AddDate(0, 0, 2), Value: 18},
			{Timestamp: baseTime.AddDate(0, 0, 3), Value: 10}, // Improvement
		},
	}

	results := DetectRegressions(series, 10.0)

	if len(results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(results))
	}

	result := results[0]
	if result.Classification != "improvement" {
		t.Errorf("Expected improvement classification, got %s", result.Classification)
	}
	if result.PercentDeviation >= 0 {
		t.Errorf("Expected negative deviation for improvement, got %.2f%%", result.PercentDeviation)
	}
}

func TestDetectRegressions_InsufficientData(t *testing.T) {
	baseTime := time.Now()
	series := metrics.MetricTimeSeries{
		MetricName: "test",
		DataPoints: []metrics.TimeSeriesPoint{
			{Timestamp: baseTime, Value: 10},
			{Timestamp: baseTime.AddDate(0, 0, 1), Value: 12},
		},
	}

	results := DetectRegressions(series, 10.0)

	if results != nil {
		t.Errorf("Expected nil for insufficient data, got %d results", len(results))
	}
}

func TestCalculateSeverity(t *testing.T) {
	tests := []struct {
		deviation float64
		threshold float64
		expected  metrics.SeverityLevel
	}{
		{5.0, 10.0, metrics.SeverityLevelInfo},
		{15.0, 10.0, metrics.SeverityLevelWarning},
		{25.0, 10.0, metrics.SeverityLevelViolation},
		{35.0, 10.0, metrics.SeverityLevelCritical},
	}

	for _, tt := range tests {
		result := calculateSeverity(tt.deviation, tt.threshold)
		if result != tt.expected {
			t.Errorf("calculateSeverity(%.1f, %.1f) = %s, expected %s",
				tt.deviation, tt.threshold, result, tt.expected)
		}
	}
}
