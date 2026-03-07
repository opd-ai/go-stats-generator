package reporter

import (
	"fmt"
	"io"

	"github.com/opd-ai/go-stats-generator/internal/metrics"
)

// Reporter interface defines the contract for generating reports
type Reporter interface {
	Generate(report *metrics.Report, output io.Writer) error
	WriteDiff(output io.Writer, diff *metrics.ComplexityDiff) error
}

// Type represents the type of reporter
type Type string

const (
	TypeConsole  Type = "console"
	TypeJSON     Type = "json"
	TypeCSV      Type = "csv"
	TypeHTML     Type = "html"
	TypeMarkdown Type = "markdown"
)

// NewReporter creates a new reporter of the specified type (console, JSON, CSV, HTML, or Markdown).
// Returns an error if the reporterType is unsupported or invalid. Console reporter uses default configuration
// (colors enabled, overview included). For custom configuration, create reporters directly with their New*WithConfig constructors.
func NewReporter(reporterType string) (Reporter, error) {
	switch Type(reporterType) {
	case TypeJSON:
		return NewJSONReporter(), nil
	case TypeCSV:
		return NewCSVReporter(), nil
	case TypeHTML:
		return NewHTMLReporter(), nil
	case TypeMarkdown:
		return NewMarkdownReporter(), nil
	case TypeConsole:
		return NewConsoleReporter(nil), nil
	default:
		return nil, fmt.Errorf("unsupported reporter type: %s", reporterType)
	}
}

// CreateReporter creates a new reporter of the specified type using typed Type enum (legacy function).
// The options parameter is ignored in the current implementation for backward compatibility.
// Prefer using NewReporter with string type or individual New*Reporter constructors for new code.
func CreateReporter(reporterType Type, options interface{}) Reporter {
	switch reporterType {
	case TypeJSON:
		return NewJSONReporter()
	case TypeCSV:
		return NewCSVReporter()
	case TypeHTML:
		return NewHTMLReporter()
	case TypeMarkdown:
		return NewMarkdownReporter()
	case TypeConsole:
		fallthrough
	default:
		return NewConsoleReporter(nil)
	}
}
