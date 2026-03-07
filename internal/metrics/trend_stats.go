package metrics

import "time"

// TrendStatistics represents linear regression analysis results for metric trends.
type TrendStatistics struct {
	Slope       float64 `json:"slope"`       // Rate of change per day
	Intercept   float64 `json:"intercept"`   // Y-intercept value
	RSquared    float64 `json:"r_squared"`   // Goodness of fit (0-1)
	DataPoints  int     `json:"data_points"` // Number of observations
	StartDate   string  `json:"start_date"`  // ISO 8601 format
	EndDate     string  `json:"end_date"`    // ISO 8601 format
	Correlation float64 `json:"correlation"` // Pearson correlation coefficient
}

// ForecastResult represents a predicted future metric value with confidence bounds.
type ForecastResult struct {
	MetricName       string          `json:"metric_name"`
	PredictionDate   time.Time       `json:"prediction_date"`
	PointEstimate    float64         `json:"point_estimate"`
	LowerBound       float64         `json:"lower_bound"` // 95% CI lower
	UpperBound       float64         `json:"upper_bound"` // 95% CI upper
	ConfidenceLevel  float64         `json:"confidence_level"`
	TrendStatistics  TrendStatistics `json:"trend_statistics"`
	ReliabilityScore float64         `json:"reliability_score"` // 0-1, based on r²
	Warning          string          `json:"warning,omitempty"` // Set if r² < 0.5
}

// RegressionResult represents detection of metric regression or improvement.
type RegressionResult struct {
	MetricName        string          `json:"metric_name"`
	CurrentValue      float64         `json:"current_value"`
	ExpectedValue     float64         `json:"expected_value"`     // Based on trend
	PercentDeviation  float64         `json:"percent_deviation"`  // Positive = worse than expected
	Classification    string          `json:"classification"`     // "regression" / "improvement" / "stable"
	Severity          SeverityLevel   `json:"severity"`           // "low" / "medium" / "high" / "critical"
	PValue            float64         `json:"p_value"`            // Statistical significance
	SignificanceLevel float64         `json:"significance_level"` // Threshold used (e.g., 0.05)
	DetectedAt        time.Time       `json:"detected_at"`
	TrendStatistics   TrendStatistics `json:"trend_statistics"`
}

// MetricTimeSeries represents historical data points for a single metric.
type MetricTimeSeries struct {
	MetricName string            `json:"metric_name"`
	DataPoints []TimeSeriesPoint `json:"data_points"`
}

// TimeSeriesPoint represents a single observation at a point in time.
type TimeSeriesPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Value     float64   `json:"value"`
}
