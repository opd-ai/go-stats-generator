package reporter

import (
	"bytes"
	"strings"
	"testing"

	"github.com/opd-ai/go-stats-generator/internal/config"
	"github.com/opd-ai/go-stats-generator/internal/metrics"
)

func TestConsoleReporter_WithOrganization(t *testing.T) {
	report := &metrics.Report{
		Organization: metrics.OrganizationMetrics{
			OversizedFiles: []metrics.OversizedFile{
				{
					File:              "internal/large_file.go",
					Lines:             metrics.LineMetrics{Code: 800},
					FunctionCount:     25,
					TypeCount:         8,
					MaintenanceBurden: 75.5,
					Severity:          "high",
				},
			},
			OversizedPackages: []metrics.OversizedPackage{
				{
					Package:         "github.com/test/bigpackage",
					FileCount:       25,
					ExportedSymbols: 50,
					TotalFunctions:  120,
					CohesionScore:   0.35,
					IsMegaPackage:   true,
					Severity:        "high",
				},
			},
			HighFanInPackages: []metrics.FanInPackage{
				{
					Package:   "github.com/test/core",
					FanIn:     15,
					RiskLevel: "high",
				},
			},
			AvgPackageStability: 0.55,
		},
	}

	cfg := &config.OutputConfig{
		IncludeDetails: true,
		Limit:          10,
	}

	cr := NewConsoleReporter(cfg)
	var buf bytes.Buffer
	err := cr.Generate(report, &buf)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	output := buf.String()

	// Verify organization section is present
	if !strings.Contains(output, "ORGANIZATION HEALTH") {
		t.Error("Output should contain 'ORGANIZATION HEALTH' section")
	}

	// Verify key metrics are displayed
	expectedStrings := []string{
		"Oversized Files: 1",
		"Oversized Packages: 1",
		"High Fan-In Packages: 1",
		"Avg Package Instability: 0.55",
		"Top 1 Oversized Files:",
		"internal/large_file.go",
		"Top 1 Oversized Packages:",
		"github.com/test/bigpackage",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(output, expected) {
			t.Errorf("Output should contain %q", expected)
		}
	}
}
