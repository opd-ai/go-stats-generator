package reporter

import (
	"fmt"
	"html/template"
	"io"
	"time"

	"github.com/opd-ai/go-stats-generator/internal/config"
	"github.com/opd-ai/go-stats-generator/internal/metrics"
)

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
		return fmt.Errorf("failed to parse template: %w", err)
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
	}).Parse(htmlDiffTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse diff template: %w", err)
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

// HTML template for basic reports
const htmlReportTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Go Stats Report - {{.Report.Metadata.Repository}}</title>
    <style>
        {{ template "styles" }}
    </style>
</head>
<body>
    <div class="container">
        <header>
            <h1>Go Source Code Statistics Report</h1>
            <div class="metadata">
                <p><strong>Repository:</strong> {{.Report.Metadata.Repository}}</p>
                <p><strong>Generated:</strong> {{formatTime .Report.Metadata.GeneratedAt}}</p>
                <p><strong>Analysis Time:</strong> {{formatDuration .Report.Metadata.AnalysisTime}}</p>
                <p><strong>Files Processed:</strong> {{.Report.Metadata.FilesProcessed}}</p>
            </div>
        </header>

        {{if .Config.IncludeOverview}}
        <section class="overview">
            <h2>Overview</h2>
            <div class="stats-grid">
                <div class="stat-card">
                    <h3>{{.Report.Overview.TotalFiles}}</h3>
                    <p>Total Files</p>
                </div>
                <div class="stat-card">
                    <h3>{{.Report.Overview.TotalLines}}</h3>
                    <p>Total Lines</p>
                </div>
                <div class="stat-card">
                    <h3>{{.Report.Overview.TotalFunctions}}</h3>
                    <p>Total Functions</p>
                </div>
                <div class="stat-card">
                    <h3>{{formatFloat .Report.Overview.AverageComplexity}}</h3>
                    <p>Average Complexity</p>
                </div>
            </div>
        </section>
        {{end}}

        {{if and .Config.IncludeDetails .Report.Functions}}
        <section class="functions">
            <h2>Function Analysis</h2>
            <table class="data-table">
                <thead>
                    <tr>
                        <th>Function</th>
                        <th>Package</th>
                        <th>Lines</th>
                        <th>Complexity</th>
                        <th>Parameters</th>
                        <th>Returns</th>
                    </tr>
                </thead>
                <tbody>
                    {{range .Report.Functions}}
                    <tr>
                        <td>{{.Name}}</td>
                        <td>{{.Package}}</td>
                        <td>{{.LineCount}}</td>
                        <td>{{.CyclomaticComplexity}}</td>
                        <td>{{.ParameterCount}}</td>
                        <td>{{.ReturnCount}}</td>
                    </tr>
                    {{end}}
                </tbody>
            </table>
        </section>
        {{end}}
    </div>
</body>
</html>`

// HTML template for diff reports
const htmlDiffTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Go Stats Diff Report</title>
    <style>
        {{ template "styles" }}
        {{ template "diffStyles" }}
    </style>
</head>
<body>
    <div class="container">
        <header>
            <h1>Go Source Code Complexity Diff Report</h1>
            <div class="metadata">
                <p><strong>Baseline:</strong> {{.Diff.Baseline.ID}} ({{formatTime .Diff.Baseline.Timestamp}})</p>
                <p><strong>Current:</strong> {{.Diff.Current.ID}} ({{formatTime .Diff.Current.Timestamp}})</p>
                <p><strong>Comparison:</strong> {{formatTime .Diff.Timestamp}}</p>
            </div>
        </header>

        <section class="summary">
            <h2>Summary</h2>
            <div class="summary-grid">
                <div class="summary-card">
                    <h3>{{.Diff.Summary.TotalChanges}}</h3>
                    <p>Total Changes</p>
                </div>
                <div class="summary-card regressions">
                    <h3>{{.Diff.Summary.TotalRegressions}}</h3>
                    <p>Regressions</p>
                </div>
                <div class="summary-card improvements">
                    <h3>{{.Diff.Summary.TotalImprovements}}</h3>
                    <p>Improvements</p>
                </div>
                <div class="summary-card {{severityClass .Diff.Summary.OverallSeverity}}">
                    <h3>{{.Diff.Summary.OverallSeverity}}</h3>
                    <p>Overall Severity</p>
                </div>
            </div>
        </section>

        {{if .Diff.Regressions}}
        <section class="regressions">
            <h2>Regressions ({{len .Diff.Regressions}})</h2>
            <table class="data-table">
                <thead>
                    <tr>
                        <th>Type</th>
                        <th>Entity</th>
                        <th>Metric</th>
                        <th>Before</th>
                        <th>After</th>
                        <th>Change</th>
                        <th>Severity</th>
                    </tr>
                </thead>
                <tbody>
                    {{range .Diff.Regressions}}
                    <tr class="regression-row">
                        <td>{{.EntityType}}</td>
                        <td>{{.EntityName}}</td>
                        <td>{{.MetricName}}</td>
                        <td>{{formatFloat .OldValue}}</td>
                        <td>{{formatFloat .NewValue}}</td>
                        <td class="{{changeClass .PercentChange}}">{{formatChange .PercentChange}}</td>
                        <td class="{{severityClass .Severity}}">{{.Severity}}</td>
                    </tr>
                    {{end}}
                </tbody>
            </table>
        </section>
        {{end}}

        {{if .Diff.Improvements}}
        <section class="improvements">
            <h2>Improvements ({{len .Diff.Improvements}})</h2>
            <table class="data-table">
                <thead>
                    <tr>
                        <th>Type</th>
                        <th>Entity</th>
                        <th>Metric</th>
                        <th>Before</th>
                        <th>After</th>
                        <th>Change</th>
                        <th>Impact</th>
                    </tr>
                </thead>
                <tbody>
                    {{range .Diff.Improvements}}
                    <tr class="improvement-row">
                        <td>{{.EntityType}}</td>
                        <td>{{.EntityName}}</td>
                        <td>{{.MetricName}}</td>
                        <td>{{formatFloat .OldValue}}</td>
                        <td>{{formatFloat .NewValue}}</td>
                        <td class="{{changeClass .PercentChange}}">{{formatChange .PercentChange}}</td>
                        <td>{{.Impact}}</td>
                    </tr>
                    {{end}}
                </tbody>
            </table>
        </section>
        {{end}}

        {{if and .Config.IncludeDetails .Diff.Changes}}
        <section class="all-changes">
            <h2>All Changes ({{len .Diff.Changes}})</h2>
            <table class="data-table">
                <thead>
                    <tr>
                        <th>Type</th>
                        <th>Entity</th>
                        <th>Metric</th>
                        <th>Before</th>
                        <th>After</th>
                        <th>Change</th>
                        <th>Threshold</th>
                    </tr>
                </thead>
                <tbody>
                    {{range .Diff.Changes}}
                    <tr>
                        <td>{{.EntityType}}</td>
                        <td>{{.EntityName}}</td>
                        <td>{{.MetricName}}</td>
                        <td>{{formatFloat .OldValue}}</td>
                        <td>{{formatFloat .NewValue}}</td>
                        <td class="{{changeClass .PercentChange}}">{{formatChange .PercentChange}}</td>
                        <td class="{{thresholdClass .ThresholdExceeded}}">
                            {{if .ThresholdExceeded}}Exceeded{{else}}OK{{end}}
                        </td>
                    </tr>
                    {{end}}
                </tbody>
            </table>
        </section>
        {{end}}
    </div>
</body>
</html>

{{define "styles"}}
body {
    font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
    line-height: 1.6;
    margin: 0;
    padding: 0;
    background-color: #f8f9fa;
    color: #333;
}

.container {
    max-width: 1200px;
    margin: 0 auto;
    padding: 20px;
}

header {
    background: white;
    padding: 30px;
    border-radius: 8px;
    box-shadow: 0 2px 4px rgba(0,0,0,0.1);
    margin-bottom: 30px;
}

header h1 {
    margin: 0 0 20px 0;
    color: #2c3e50;
    font-size: 2.5em;
}

.metadata p {
    margin: 5px 0;
    color: #6c757d;
}

section {
    background: white;
    margin-bottom: 30px;
    padding: 30px;
    border-radius: 8px;
    box-shadow: 0 2px 4px rgba(0,0,0,0.1);
}

section h2 {
    margin: 0 0 20px 0;
    color: #2c3e50;
    border-bottom: 3px solid #3498db;
    padding-bottom: 10px;
}

.stats-grid, .summary-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
    gap: 20px;
    margin-top: 20px;
}

.stat-card, .summary-card {
    text-align: center;
    padding: 20px;
    border-radius: 8px;
    background: #f8f9fa;
    border: 1px solid #e9ecef;
}

.stat-card h3, .summary-card h3 {
    margin: 0;
    font-size: 2em;
    color: #2c3e50;
}

.stat-card p, .summary-card p {
    margin: 10px 0 0 0;
    color: #6c757d;
    font-weight: 500;
}

.data-table {
    width: 100%;
    border-collapse: collapse;
    margin-top: 20px;
}

.data-table th,
.data-table td {
    padding: 12px;
    text-align: left;
    border-bottom: 1px solid #dee2e6;
}

.data-table th {
    background-color: #f8f9fa;
    font-weight: 600;
    color: #495057;
}

.data-table tbody tr:hover {
    background-color: #f8f9fa;
}
{{end}}

{{define "diffStyles"}}
.summary-card.regressions {
    border-color: #dc3545;
    background-color: #f8d7da;
}

.summary-card.improvements {
    border-color: #28a745;
    background-color: #d4edda;
}

.increase {
    color: #dc3545;
    font-weight: bold;
}

.decrease {
    color: #28a745;
    font-weight: bold;
}

.neutral {
    color: #6c757d;
}

.severity-low {
    color: #17a2b8;
}

.severity-medium {
    color: #ffc107;
}

.severity-high {
    color: #fd7e14;
}

.severity-critical {
    color: #dc3545;
    font-weight: bold;
}

.threshold-exceeded {
    color: #dc3545;
    font-weight: bold;
}

.threshold-ok {
    color: #28a745;
}

.regression-row {
    background-color: #fff5f5;
}

.improvement-row {
    background-color: #f0fff4;
}
{{end}}`
