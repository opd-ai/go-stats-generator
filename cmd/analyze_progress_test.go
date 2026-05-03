package cmd

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"

	"github.com/opd-ai/go-stats-generator/internal/config"
)

// TestProgressOutputDisabledForNonConsoleFormats verifies that progress output
// is automatically disabled for JSON, CSV, HTML, and Markdown formats to prevent
// pollution of structured output.
func TestProgressOutputDisabledForNonConsoleFormats(t *testing.T) {
	tests := []struct {
		name           string
		format         config.OutputFormat
		verbose        bool
		expectProgress bool
	}{
		{
			name:           "JSON format disables progress",
			format:         config.FormatJSON,
			verbose:        false,
			expectProgress: false,
		},
		{
			name:           "JSON format disables progress even with verbose",
			format:         config.FormatJSON,
			verbose:        true,
			expectProgress: false,
		},
		{
			name:           "CSV format disables progress",
			format:         config.FormatCSV,
			verbose:        false,
			expectProgress: false,
		},
		{
			name:           "HTML format disables progress",
			format:         config.FormatHTML,
			verbose:        false,
			expectProgress: false,
		},
		{
			name:           "Markdown format disables progress",
			format:         config.FormatMarkdown,
			verbose:        false,
			expectProgress: false,
		},
		{
			name:           "Console format enables progress",
			format:         config.FormatConsole,
			verbose:        false,
			expectProgress: true,
		},
		{
			name:           "Console format with verbose enables progress",
			format:         config.FormatConsole,
			verbose:        true,
			expectProgress: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset viper state for each test
			viper.Reset()

			// Set up configuration
			cfg := config.DefaultConfig()
			cfg.Output.Format = tt.format
			cfg.Output.Verbose = tt.verbose

			// Apply the configuration logic
			applyVerboseDefaults(cfg)

			// Verify progress setting
			assert.Equal(t, tt.expectProgress, cfg.Output.ShowProgress,
				"ShowProgress should be %v for format=%s verbose=%v",
				tt.expectProgress, tt.format, tt.verbose)
		})
	}
}

// TestProgressOutputFormatPrecedence verifies that format-based progress
// disabling takes precedence over verbose flag.
func TestProgressOutputFormatPrecedence(t *testing.T) {
	// Reset viper
	viper.Reset()

	cfg := config.DefaultConfig()
	cfg.Output.Format = config.FormatJSON
	cfg.Output.Verbose = true

	applyVerboseDefaults(cfg)

	// JSON format should disable progress even when verbose is true
	assert.False(t, cfg.Output.ShowProgress,
		"JSON format should disable progress even with verbose=true")
}

// TestProgressOutputDefaultConsole verifies that console format has
// progress enabled by default.
func TestProgressOutputDefaultConsole(t *testing.T) {
	// Reset viper
	viper.Reset()

	cfg := config.DefaultConfig()
	assert.Equal(t, config.FormatConsole, cfg.Output.Format,
		"Default format should be console")

	applyVerboseDefaults(cfg)

	assert.True(t, cfg.Output.ShowProgress,
		"Console format should have progress enabled by default")
}
