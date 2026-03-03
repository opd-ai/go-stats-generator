package config

// CustomMetricsConfig controls user-defined custom metric calculation
type CustomMetricsConfig struct {
	Enabled bool                     `mapstructure:"enabled" json:"enabled"`
	Metrics []CustomMetricDefinition `mapstructure:"metrics" json:"metrics"`
}

// CustomMetricDefinition defines a single custom metric
type CustomMetricDefinition struct {
	Name        string `mapstructure:"name" json:"name"`
	Type        string `mapstructure:"type" json:"type"` // counter, ratio, measurement
	Description string `mapstructure:"description" json:"description"`
	Pattern     string `mapstructure:"pattern" json:"pattern"`

	// For ratio type metrics
	NumeratorPattern   string `mapstructure:"numerator_pattern" json:"numerator_pattern,omitempty"`
	DenominatorPattern string `mapstructure:"denominator_pattern" json:"denominator_pattern,omitempty"`

	// For measurement type metrics
	Aggregation string `mapstructure:"aggregation" json:"aggregation,omitempty"` // sum, avg, max, min, count
	Property    string `mapstructure:"property" json:"property,omitempty"`
}

// DefaultCustomMetricsConfig returns default configuration for custom metrics
func DefaultCustomMetricsConfig() CustomMetricsConfig {
	return CustomMetricsConfig{
		Enabled: false,
		Metrics: []CustomMetricDefinition{},
	}
}
