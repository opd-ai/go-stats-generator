package metrics

import (
	"testing"
)

func TestFilterReportSections_EmptySectionsNoOp(t *testing.T) {
	report := &Report{
		Functions:  []FunctionMetrics{{Name: "foo"}},
		Structs:    []StructMetrics{{Name: "Bar"}},
		Interfaces: []InterfaceMetrics{{Name: "Baz"}},
	}

	FilterReportSections(report, nil)

	if len(report.Functions) != 1 {
		t.Errorf("expected functions preserved, got %d", len(report.Functions))
	}
	if len(report.Structs) != 1 {
		t.Errorf("expected structs preserved, got %d", len(report.Structs))
	}
	if len(report.Interfaces) != 1 {
		t.Errorf("expected interfaces preserved, got %d", len(report.Interfaces))
	}
}

func TestFilterReportSections_KeepOnlyFunctions(t *testing.T) {
	report := &Report{
		Functions:  []FunctionMetrics{{Name: "foo"}},
		Structs:    []StructMetrics{{Name: "Bar"}},
		Interfaces: []InterfaceMetrics{{Name: "Baz"}},
		Packages:   []PackageMetrics{{Name: "pkg"}},
	}

	FilterReportSections(report, []string{"functions"})

	if len(report.Functions) != 1 {
		t.Errorf("expected functions preserved, got %d", len(report.Functions))
	}
	if report.Structs != nil {
		t.Errorf("expected structs zeroed, got %v", report.Structs)
	}
	if report.Interfaces != nil {
		t.Errorf("expected interfaces zeroed, got %v", report.Interfaces)
	}
	if report.Packages != nil {
		t.Errorf("expected packages zeroed, got %v", report.Packages)
	}
}

func TestFilterReportSections_MultipleSections(t *testing.T) {
	report := &Report{
		Functions: []FunctionMetrics{{Name: "foo"}},
		Structs:   []StructMetrics{{Name: "Bar"}},
		Duplication: DuplicationMetrics{
			ClonePairs: 5,
		},
	}

	FilterReportSections(report, []string{"functions", "duplication"})

	if len(report.Functions) != 1 {
		t.Errorf("expected functions preserved, got %d", len(report.Functions))
	}
	if report.Duplication.ClonePairs != 5 {
		t.Errorf("expected duplication preserved, got %d", report.Duplication.ClonePairs)
	}
	if report.Structs != nil {
		t.Errorf("expected structs zeroed, got %v", report.Structs)
	}
}

func TestFilterReportSections_CaseInsensitive(t *testing.T) {
	report := &Report{
		Functions: []FunctionMetrics{{Name: "foo"}},
		Structs:   []StructMetrics{{Name: "Bar"}},
	}

	FilterReportSections(report, []string{"Functions"})

	if len(report.Functions) != 1 {
		t.Errorf("expected functions preserved with uppercase input, got %d", len(report.Functions))
	}
	if report.Structs != nil {
		t.Errorf("expected structs zeroed, got %v", report.Structs)
	}
}

func TestFilterReportSections_ConcurrencyAlias(t *testing.T) {
	report := &Report{
		Patterns: PatternMetrics{
			ConcurrencyPatterns: ConcurrencyPatternMetrics{
				Goroutines: GoroutineMetrics{TotalCount: 3},
			},
		},
		Functions: []FunctionMetrics{{Name: "foo"}},
	}

	FilterReportSections(report, []string{"concurrency"})

	if report.Patterns.ConcurrencyPatterns.Goroutines.TotalCount != 3 {
		t.Errorf("expected patterns preserved via concurrency alias, got %d",
			report.Patterns.ConcurrencyPatterns.Goroutines.TotalCount)
	}
	if report.Functions != nil {
		t.Errorf("expected functions zeroed, got %v", report.Functions)
	}
}

func TestFilterReportSections_WhitespaceHandling(t *testing.T) {
	report := &Report{
		Functions: []FunctionMetrics{{Name: "foo"}},
		Structs:   []StructMetrics{{Name: "Bar"}},
	}

	FilterReportSections(report, []string{" functions ", " structs"})

	if len(report.Functions) != 1 {
		t.Errorf("expected functions preserved with whitespace, got %d", len(report.Functions))
	}
	if len(report.Structs) != 1 {
		t.Errorf("expected structs preserved with whitespace, got %d", len(report.Structs))
	}
}

func TestFilterReportSections_MetadataAlwaysPreserved(t *testing.T) {
	report := &Report{
		Metadata: ReportMetadata{
			Repository:     "test/repo",
			FilesProcessed: 42,
			ToolVersion:    "1.0.0",
		},
		Overview: OverviewMetrics{
			TotalFunctions: 100,
			TotalFiles:     42,
		},
		Functions: []FunctionMetrics{{Name: "foo"}},
		Structs:   []StructMetrics{{Name: "Bar"}},
	}

	FilterReportSections(report, []string{"functions"})

	if report.Metadata.Repository != "test/repo" {
		t.Errorf("expected metadata preserved, got empty repository")
	}
	if report.Metadata.FilesProcessed != 42 {
		t.Errorf("expected metadata.FilesProcessed=42, got %d", report.Metadata.FilesProcessed)
	}
	if report.Overview.TotalFunctions != 100 {
		t.Errorf("expected overview.TotalFunctions=100, got %d", report.Overview.TotalFunctions)
	}
	if len(report.Functions) != 1 {
		t.Errorf("expected functions preserved, got %d", len(report.Functions))
	}
	if report.Structs != nil {
		t.Errorf("expected structs zeroed, got %v", report.Structs)
	}
}

func TestValidSections_AllPresent(t *testing.T) {
	expected := []string{
		"metadata", "overview", "functions", "structs", "interfaces",
		"packages", "patterns", "concurrency", "complexity", "documentation",
		"generics", "duplication", "naming", "placement", "organization",
		"burden", "scores", "suggestions",
	}

	for _, s := range expected {
		if !ValidSections[s] {
			t.Errorf("expected %q in ValidSections", s)
		}
	}
}
