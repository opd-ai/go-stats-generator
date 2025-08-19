package reporter

import (
	_ "embed"
	"fmt"
	"html/template"
	"io"
	"time"

	"github.com/opd-ai/go-stats-generator/internal/config"
	"github.com/opd-ai/go-stats-generator/internal/metrics"
)

//go:embed templates/html/report.html
var htmlReportTemplate string

//go:embed templates/html/diff.html
var htmlDiffTemplate string

// HTMLReporterImpl generates HTML reports with interactive charts
type HTMLReporterImpl struct {
	config *config.OutputConfig
}

// NewHTMLReporterWithConfig creates a new HTML reporter with config
func NewHTMLReporterWithConfig(cfg *config.OutputConfig) *HTMLReporterImpl {
	if cfg == nil {
		cfg = &config.OutputConfig{
			IncludeOverview: true,
			IncludeDetails:  true,
			Limit:           50,
		}
	}

	return &HTMLReporterImpl{
		config: cfg,
	}
}

// Generate generates an HTML report
func (hr *HTMLReporterImpl) Generate(report *metrics.Report, output io.Writer) error {
	tmpl, err := template.New("report").Funcs(template.FuncMap{
		"formatTime":     formatTime,
		"formatDuration": formatDuration,
		"formatFloat":    formatFloat,
		"formatPercent":  formatPercent,
	}).Parse(htmlReportTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse embedded report template: %w", err)
	}

	data := struct {
		Report *metrics.Report
		Config *config.OutputConfig
	}{
		Report: report,
		Config: hr.config,
	}

	return tmpl.Execute(output, data)
}

// WriteDiff generates an HTML diff report
func (hr *HTMLReporterImpl) WriteDiff(output io.Writer, diff *metrics.ComplexityDiff) error {
	tmpl, err := template.New("diff").Funcs(template.FuncMap{
		"formatTime":     formatTime,
		"formatDuration": formatDuration,
		"formatFloat":    formatFloat,
		"formatPercent":  formatPercent,
		"formatChange":   formatChange,
		"changeClass":    changeClass,
		"severityClass":  severityClass,
		"thresholdClass": thresholdClass,
		"trendClass":     trendClass,
	}).Parse(htmlDiffTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse embedded diff template: %w", err)
	}

	data := struct {
		Diff   *metrics.ComplexityDiff
		Config *config.OutputConfig
	}{
		Diff:   diff,
		Config: hr.config,
	}

	return tmpl.Execute(output, data)
}

// Template helper functions
func formatTime(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}

func formatDuration(d time.Duration) string {
	return d.Round(time.Millisecond).String()
}

func formatFloat(f float64) string {
	return fmt.Sprintf("%.2f", f)
}

func formatPercent(f float64) string {
	return fmt.Sprintf("%.1f%%", f)
}

func formatChange(change float64) string {
	if change > 0 {
		return fmt.Sprintf("+%.1f%%", change)
	}
	return fmt.Sprintf("%.1f%%", change)
}

func changeClass(change float64) string {
	if change > 0 {
		return "increase"
	} else if change < 0 {
		return "decrease"
	}
	return "neutral"
}

func severityClass(severity metrics.SeverityLevel) string {
	switch severity {
	case metrics.SeverityLevelInfo:
		return "severity-info"
	case metrics.SeverityLevelWarning:
		return "severity-warning"
	case metrics.SeverityLevelError:
		return "severity-error"
	case metrics.SeverityLevelCritical:
		return "severity-critical"
	default:
		return "severity-info"
	}
}

func thresholdClass(exceeded bool) string {
	if exceeded {
		return "threshold-exceeded"
	}
	return "threshold-ok"
}

func trendClass(trend metrics.TrendDirection) string {
	switch trend {
	case metrics.TrendImproving:
		return "trend-improving"
	case metrics.TrendDegrading:
		return "trend-degrading"
	case metrics.TrendVolatile:
		return "trend-volatile"
	default:
		return "trend-stable"
	}
}
