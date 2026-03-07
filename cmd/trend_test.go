package cmd

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTrendCommand(t *testing.T) {
	t.Run("flag parsing", func(t *testing.T) {
		cmd := trendCmd
		assert.NotNil(t, cmd)
		assert.Equal(t, "trend", cmd.Use)
		assert.Equal(t, "Analyze trends in code metrics over time", cmd.Short)

		daysFlag := cmd.PersistentFlags().Lookup("days")
		assert.NotNil(t, daysFlag)
		assert.Equal(t, "30", daysFlag.DefValue)

		metricFlag := cmd.PersistentFlags().Lookup("metric")
		assert.NotNil(t, metricFlag)

		entityFlag := cmd.PersistentFlags().Lookup("entity")
		assert.NotNil(t, entityFlag)

		thresholdFlag := cmd.PersistentFlags().Lookup("threshold")
		assert.NotNil(t, thresholdFlag)
		assert.Equal(t, "10", thresholdFlag.DefValue)
	})

	t.Run("subcommands exist", func(t *testing.T) {
		cmd := trendCmd

		analyzeCmd := cmd.Commands()[0]
		assert.NotNil(t, analyzeCmd)

		forecastCmd := cmd.Commands()[1]
		assert.NotNil(t, forecastCmd)

		regressionsCmd := cmd.Commands()[2]
		assert.NotNil(t, regressionsCmd)
	})
}

func TestTrendAnalyzeCommand(t *testing.T) {
	t.Run("command structure", func(t *testing.T) {
		cmd := trendAnalyzeCmd
		assert.NotNil(t, cmd)
		assert.Equal(t, "analyze", cmd.Use)
		assert.Equal(t, "Analyze burden metric trends over time", cmd.Short)
	})
}

func TestTrendForecastCommand(t *testing.T) {
	t.Run("command structure", func(t *testing.T) {
		cmd := trendForecastCmd
		assert.NotNil(t, cmd)
		assert.Equal(t, "forecast", cmd.Use)
		assert.Equal(t, "Forecast future metrics using linear regression", cmd.Short)
	})
}

func TestTrendRegressionsCommand(t *testing.T) {
	t.Run("command structure", func(t *testing.T) {
		cmd := trendRegressionsCmd
		assert.NotNil(t, cmd)
		assert.Equal(t, "regressions", cmd.Use)
		assert.Equal(t, "Detect metric regressions using statistical analysis", cmd.Short)
	})
}

func TestTrendCommandHelp(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		expectedOutput []string
	}{
		{
			name: "trend help",
			args: []string{"trend", "--help"},
			expectedOutput: []string{
				"Analyze trends and patterns in code metrics over time",
				"--days",
				"--metric",
				"--entity",
			},
		},
		{
			name: "trend analyze help",
			args: []string{"trend", "analyze", "--help"},
			expectedOutput: []string{
				"Analyze burden metric trends",
				"MBI",
				"Duplication ratio",
			},
		},
		{
			name: "trend forecast help",
			args: []string{"trend", "forecast", "--help"},
			expectedOutput: []string{
				"Generate forecasts",
				"linear regression",
				"confidence intervals",
			},
		},
		{
			name: "trend regressions help",
			args: []string{"trend", "regressions", "--help"},
			expectedOutput: []string{
				"Detect potential regressions",
				"statistical",
				"P-values",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rootCmd.SetArgs(tt.args)
			buf := new(bytes.Buffer)
			rootCmd.SetOut(buf)
			rootCmd.SetErr(buf)

			err := rootCmd.Execute()
			assert.NoError(t, err)

			output := buf.String()
			for _, expected := range tt.expectedOutput {
				assert.Contains(t, output, expected)
			}
		})
	}
}
