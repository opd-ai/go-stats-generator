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
	"test_coverage": true,
	"test_quality":  true,
	"suggestions":   true,
}

// sectionHandler defines how to clear a specific report section.
type sectionHandler func(*Report)

// sectionHandlers maps section names to their clearing functions.
var sectionHandlers = map[string]sectionHandler{
	"metadata":      func(r *Report) { r.Metadata = ReportMetadata{} },
	"overview":      func(r *Report) { r.Overview = OverviewMetrics{} },
	"functions":     func(r *Report) { r.Functions = nil },
	"structs":       func(r *Report) { r.Structs = nil },
	"interfaces":    func(r *Report) { r.Interfaces = nil },
	"packages":      clearPackageSection,
	"patterns":      func(r *Report) { r.Patterns = PatternMetrics{} },
	"complexity":    func(r *Report) { r.Complexity = ComplexityMetrics{} },
	"documentation": func(r *Report) { r.Documentation = DocumentationMetrics{} },
	"generics":      func(r *Report) { r.Generics = GenericMetrics{} },
	"duplication":   func(r *Report) { r.Duplication = DuplicationMetrics{} },
	"naming":        func(r *Report) { r.Naming = NamingMetrics{} },
	"placement":     func(r *Report) { r.Placement = PlacementMetrics{} },
	"organization":  func(r *Report) { r.Organization = OrganizationMetrics{} },
	"burden":        func(r *Report) { r.Burden = BurdenMetrics{} },
	"scores":        func(r *Report) { r.Scores = ScoringMetrics{} },
	"test_coverage": func(r *Report) { r.TestCoverage = TestCoverageMetrics{} },
	"test_quality":  func(r *Report) { r.TestQuality = TestQualityMetrics{} },
	"suggestions":   func(r *Report) { r.Suggestions = nil },
}

// clearPackageSection clears both packages and circular dependencies.
func clearPackageSection(r *Report) {
	r.Packages = nil
	r.CircularDependencies = nil
}

// FilterReportSections zeros out report sections that are not in the requested set.
// Returns the report unchanged if sections is empty.
func FilterReportSections(report *Report, sections []string) {
	if len(sections) == 0 {
		return
	}

	keep := buildSectionKeepSet(sections)
	clearUnrequestedSections(report, keep)
}

// buildSectionKeepSet creates a set of section names to keep, normalized to lowercase.
func buildSectionKeepSet(sections []string) map[string]bool {
	keep := make(map[string]bool, len(sections))
	for _, s := range sections {
		keep[strings.ToLower(strings.TrimSpace(s))] = true
	}
	if keep["concurrency"] {
		keep["patterns"] = true
	}
	return keep
}

// clearUnrequestedSections removes all sections not in the keep set.
func clearUnrequestedSections(report *Report, keep map[string]bool) {
	for section, handler := range sectionHandlers {
		if !keep[section] {
			handler(report)
		}
	}
}
