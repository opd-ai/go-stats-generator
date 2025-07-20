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

// ReporterType represents the type of reporter
type ReporterType string

const (
	TypeConsole  ReporterType = "console"
	TypeJSON     ReporterType = "json"
	TypeCSV      ReporterType = "csv"
	TypeHTML     ReporterType = "html"
	TypeMarkdown ReporterType = "markdown"
)

// NewReporter creates a new reporter of the specified type
func NewReporter(reporterType string) (Reporter, error) {
	switch ReporterType(reporterType) {
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

// CreateReporter creates a new reporter of the specified type (legacy)
func CreateReporter(reporterType ReporterType, options interface{}) Reporter {
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
