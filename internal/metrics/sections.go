package metrics

import "strings"

// ValidSections lists all valid report section names for --sections/--only filtering.
var ValidSections = map[string]bool{
	"metadata":      true,
	"overview":      true,
	"functions":     true,
	"structs":       true,
	"interfaces":    true,
	"packages":      true,
	"patterns":      true,
	"concurrency":   true, // alias for patterns
	"complexity":    true,
	"documentation": true,
	"generics":      true,
	"duplication":   true,
	"naming":        true,
	"placement":     true,
	"organization":  true,
	"burden":        true,
	"scores":        true,
	"suggestions":   true,
}

// FilterReportSections zeros out report sections that are not in the requested set.
// FilterReportSections returns the report unchanged if sections is empty.
func FilterReportSections(report *Report, sections []string) {
	if len(sections) == 0 {
		return
	}

	// Build lookup set, normalizing to lowercase
	keep := make(map[string]bool, len(sections))
	for _, s := range sections {
		keep[strings.ToLower(strings.TrimSpace(s))] = true
	}

	// "concurrency" is an alias for "patterns"
	if keep["concurrency"] {
		keep["patterns"] = true
	}

	if !keep["metadata"] {
		report.Metadata = ReportMetadata{}
	}
	if !keep["overview"] {
		report.Overview = OverviewMetrics{}
	}
	if !keep["functions"] {
		report.Functions = nil
	}
	if !keep["structs"] {
		report.Structs = nil
	}
	if !keep["interfaces"] {
		report.Interfaces = nil
	}
	if !keep["packages"] {
		report.Packages = nil
	}
	if !keep["patterns"] {
		report.Patterns = PatternMetrics{}
	}
	if !keep["complexity"] {
		report.Complexity = ComplexityMetrics{}
	}
	if !keep["documentation"] {
		report.Documentation = DocumentationMetrics{}
	}
	if !keep["generics"] {
		report.Generics = GenericMetrics{}
	}
	if !keep["duplication"] {
		report.Duplication = DuplicationMetrics{}
	}
	if !keep["naming"] {
		report.Naming = NamingMetrics{}
	}
	if !keep["placement"] {
		report.Placement = PlacementMetrics{}
	}
	if !keep["organization"] {
		report.Organization = OrganizationMetrics{}
	}
	if !keep["burden"] {
		report.Burden = BurdenMetrics{}
	}
	if !keep["scores"] {
		report.Scores = ScoringMetrics{}
	}
	if !keep["suggestions"] {
		report.Suggestions = nil
	}
}
