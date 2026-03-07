package reporter

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/opd-ai/go-stats-generator/internal/metrics"
)

// writeDuplicationAnalysis generates duplication analysis output
func (cr *ConsoleReporter) writeDuplicationAnalysis(output io.Writer, report *metrics.Report) {
	fmt.Fprintln(output, "=== DUPLICATION ANALYSIS ===")
	cr.writeDuplicationSummary(output, report.Duplication)
	if len(report.Duplication.Clones) == 0 {
		return
	}
	cr.writeDuplicationTable(output, report.Duplication.Clones)
}

func (cr *ConsoleReporter) writeDuplicationSummary(output io.Writer, dup metrics.DuplicationMetrics) {
	fmt.Fprintf(output, "Clone Pairs Detected: %d\n", dup.ClonePairs)
	fmt.Fprintf(output, "Duplicated Lines: %d\n", dup.DuplicatedLines)
	fmt.Fprintf(output, "Duplication Ratio: %.2f%%\n", dup.DuplicationRatio*100)
	fmt.Fprintf(output, "Largest Clone Size: %d lines\n", dup.LargestCloneSize)
	fmt.Fprintln(output)
}

func (cr *ConsoleReporter) writeDuplicationTable(output io.Writer, clones []metrics.ClonePair) {
	sortedClones := cr.getSortedClones(clones)
	limit := cr.calculateDisplayLimit(len(sortedClones))
	cr.writeDuplicationHeader(output, limit)
	cr.writeDuplicationRows(output, sortedClones, limit)
}

func (cr *ConsoleReporter) getSortedClones(clones []metrics.ClonePair) []metrics.ClonePair {
	sorted := make([]metrics.ClonePair, len(clones))
	copy(sorted, clones)
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].LineCount != sorted[j].LineCount {
			return sorted[i].LineCount < sorted[j].LineCount
		}
		typeI := string(sorted[i].Type)
		typeJ := string(sorted[j].Type)
		if typeI != typeJ {
			return typeI < typeJ
		}
		return cr.formatCloneLocations(sorted[i]) < cr.formatCloneLocations(sorted[j])
	})
	return sorted
}

func (cr *ConsoleReporter) writeDuplicationHeader(output io.Writer, limit int) {
	fmt.Fprintf(output, "Clone Pairs (shortest to longest, %d shown):\n", limit)
	fmt.Fprintf(output, "%-15s %8s %8s %s\n", "Type", "Lines", "Instances", "Locations")
	fmt.Fprintln(output, "--------------------------------------------------------------------------------")
}

func (cr *ConsoleReporter) writeDuplicationRows(output io.Writer, clones []metrics.ClonePair, limit int) {
	for i := 0; i < limit; i++ {
		cr.writeDuplicationRow(output, clones[i])
	}
	fmt.Fprintln(output)
}

func (cr *ConsoleReporter) writeDuplicationRow(output io.Writer, clone metrics.ClonePair) {
	locations := cr.formatCloneLocations(clone)
	fmt.Fprintf(output, "%-15s %8d %8d %s\n",
		string(clone.Type),
		clone.LineCount,
		len(clone.Instances),
		locations,
	)
}

const maxConsoleLocations = 3

func (cr *ConsoleReporter) formatCloneLocations(clone metrics.ClonePair) string {
	if len(clone.Instances) == 0 {
		return ""
	}
	count := len(clone.Instances)
	limit := count
	if limit > maxConsoleLocations {
		limit = maxConsoleLocations
	}
	locations := make([]string, 0, limit)
	for i := 0; i < limit; i++ {
		inst := clone.Instances[i]
		locations = append(locations, fmt.Sprintf("%s:%d-%d", cr.truncate(inst.File, 40), inst.StartLine, inst.EndLine))
	}
	result := strings.Join(locations, ", ")
	if count > maxConsoleLocations {
		result += fmt.Sprintf(" (+%d more)", count-maxConsoleLocations)
	}
	return result
}

// writeNamingAnalysis generates naming convention analysis output
func (cr *ConsoleReporter) writeNamingAnalysis(output io.Writer, report *metrics.Report) {
	naming := report.Naming
	content := sectionContent{
		header: "=== NAMING CONVENTION ANALYSIS ===",
		summaryLines: []string{
			fmt.Sprintf("File Name Violations: %d", naming.FileNameViolations),
			fmt.Sprintf("Identifier Violations: %d", naming.IdentifierViolations),
			fmt.Sprintf("Package Name Violations: %d", naming.PackageNameViolations),
			fmt.Sprintf("Overall Naming Score: %.2f", naming.OverallNamingScore),
		},
	}
	if len(naming.IdentifierIssues) > 0 {
		content.detailWriters = append(content.detailWriters, func() {
			cr.writeIdentifierViolations(output, naming.IdentifierIssues)
		})
	}
	if len(naming.PackageNameIssues) > 0 {
		content.detailWriters = append(content.detailWriters, func() {
			cr.writePackageNameViolations(output, naming.PackageNameIssues)
		})
	}
	if len(naming.FileNameIssues) > 0 {
		content.detailWriters = append(content.detailWriters, func() {
			cr.writeFileNameViolations(output, naming.FileNameIssues)
		})
	}
	cr.writeSectionWithDetails(output, content)
}

// writeIdentifierViolations displays identifier naming violations
func (cr *ConsoleReporter) writeIdentifierViolations(output io.Writer, violations []metrics.IdentifierViolation) {
	// Sort by severity (high > medium > low) then by file
	sorted := make([]metrics.IdentifierViolation, len(violations))
	copy(sorted, violations)
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].Severity != sorted[j].Severity {
			return severityWeight(sorted[i].Severity) > severityWeight(sorted[j].Severity)
		}
		return sorted[i].File < sorted[j].File
	})

	limit := cr.calculateDisplayLimit(len(sorted))

	fmt.Fprintf(output, "Top %d Identifier Violations:\n", limit)
	fmt.Fprintf(output, "%-25s %-10s %-12s %-40s\n", "Name", "Type", "Violation", "File:Line")
	fmt.Fprintln(output, "--------------------------------------------------------------------------------")

	for i := 0; i < limit; i++ {
		v := sorted[i]
		location := fmt.Sprintf("%s:%d", cr.truncate(v.File, 30), v.Line)
		fmt.Fprintf(output, "%-25s %-10s %-12s %-40s\n",
			cr.truncate(v.Name, 25),
			v.Type,
			cr.truncate(v.ViolationType, 12),
			location,
		)
	}
	fmt.Fprintln(output)
}

// writePackageNameViolations displays package naming violations
func (cr *ConsoleReporter) writePackageNameViolations(output io.Writer, violations []metrics.PackageNameViolation) {
	// Sort by severity
	sorted := make([]metrics.PackageNameViolation, len(violations))
	copy(sorted, violations)
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].Severity != sorted[j].Severity {
			return severityWeight(sorted[i].Severity) > severityWeight(sorted[j].Severity)
		}
		return sorted[i].Package < sorted[j].Package
	})

	fmt.Fprintln(output, "Package Name Violations:")
	fmt.Fprintf(output, "%-20s %-20s %-40s\n", "Package", "Violation", "Description")
	fmt.Fprintln(output, "--------------------------------------------------------------------------------")

	for _, v := range sorted {
		fmt.Fprintf(output, "%-20s %-20s %-40s\n",
			cr.truncate(v.Package, 20),
			cr.truncate(v.ViolationType, 20),
			cr.truncate(v.Description, 40),
		)
	}
	fmt.Fprintln(output)
}

// writeFileNameViolations displays file naming violations
func (cr *ConsoleReporter) writeFileNameViolations(output io.Writer, violations []metrics.FileNameViolation) {
	// Sort by severity
	sorted := make([]metrics.FileNameViolation, len(violations))
	copy(sorted, violations)
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].Severity != sorted[j].Severity {
			return severityWeight(sorted[i].Severity) > severityWeight(sorted[j].Severity)
		}
		return sorted[i].File < sorted[j].File
	})

	limit := cr.calculateDisplayLimit(len(sorted))

	fmt.Fprintf(output, "Top %d File Name Violations:\n", limit)
	fmt.Fprintf(output, "%-40s %-20s %-30s\n", "File", "Violation", "Suggested Name")
	fmt.Fprintln(output, "--------------------------------------------------------------------------------")

	for i := 0; i < limit; i++ {
		v := sorted[i]
		fmt.Fprintf(output, "%-40s %-20s %-30s\n",
			cr.truncate(v.File, 40),
			cr.truncate(v.ViolationType, 20),
			cr.truncate(v.SuggestedName, 30),
		)
	}
	fmt.Fprintln(output)
}

// severityWeight returns numeric weight for severity sorting
func severityWeight(severity string) int {
	switch severity {
	case "high":
		return 3
	case "medium":
		return 2
	case "low":
		return 1
	default:
		return 0
	}
}

// writePlacementAnalysis generates placement analysis output
func (cr *ConsoleReporter) writePlacementAnalysis(output io.Writer, report *metrics.Report) {
	placement := report.Placement
	content := sectionContent{
		header: "=== PLACEMENT ANALYSIS ===",
		summaryLines: []string{
			fmt.Sprintf("Misplaced Functions: %d", placement.MisplacedFunctions),
			fmt.Sprintf("Misplaced Methods: %d", placement.MisplacedMethods),
			fmt.Sprintf("Low Cohesion Files: %d", placement.LowCohesionFiles),
			fmt.Sprintf("Average File Cohesion: %.2f", placement.AvgFileCohesion),
		},
	}
	if len(placement.FunctionIssues) > 0 {
		content.detailWriters = append(content.detailWriters, func() {
			cr.writeMisplacedFunctions(output, placement.FunctionIssues)
		})
	}
	if len(placement.MethodIssues) > 0 {
		content.detailWriters = append(content.detailWriters, func() {
			cr.writeMisplacedMethods(output, placement.MethodIssues)
		})
	}
	if len(placement.CohesionIssues) > 0 {
		content.detailWriters = append(content.detailWriters, func() {
			cr.writeFileCohesionIssues(output, placement.CohesionIssues)
		})
	}
	cr.writeSectionWithDetails(output, content)
}

// writeMisplacedFunctions displays misplaced function issues
func (cr *ConsoleReporter) writeMisplacedFunctions(output io.Writer, issues []metrics.MisplacedFunctionIssue) {
	// Sort by severity (high > medium > low) then by suggested affinity (descending)
	sorted := make([]metrics.MisplacedFunctionIssue, len(issues))
	copy(sorted, issues)
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].Severity != sorted[j].Severity {
			return severityWeight(sorted[i].Severity) > severityWeight(sorted[j].Severity)
		}
		return sorted[i].SuggestedAffinity > sorted[j].SuggestedAffinity
	})

	limit := cr.calculateDisplayLimit(len(sorted))

	fmt.Fprintf(output, "Top %d Misplaced Functions:\n", limit)
	fmt.Fprintf(output, "%-30s %-25s %-25s %s\n", "Function", "Current File", "Suggested File", "Affinity Gain")
	fmt.Fprintln(output, "--------------------------------------------------------------------------------")

	for i := 0; i < limit; i++ {
		issue := sorted[i]
		affinityGain := issue.SuggestedAffinity - issue.CurrentAffinity
		fmt.Fprintf(output, "%-30s %-25s %-25s +%.2f\n",
			cr.truncate(issue.Name, 30),
			cr.truncate(issue.CurrentFile, 25),
			cr.truncate(issue.SuggestedFile, 25),
			affinityGain,
		)
	}
	fmt.Fprintln(output)
}

// writeMisplacedMethods displays misplaced method issues
func (cr *ConsoleReporter) writeMisplacedMethods(output io.Writer, issues []metrics.MisplacedMethodIssue) {
	// Sort by severity (high > medium > low) then by distance
	sorted := make([]metrics.MisplacedMethodIssue, len(issues))
	copy(sorted, issues)
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].Severity != sorted[j].Severity {
			return severityWeight(sorted[i].Severity) > severityWeight(sorted[j].Severity)
		}
		// "different_package" before "same_package"
		if sorted[i].Distance != sorted[j].Distance {
			return sorted[i].Distance > sorted[j].Distance
		}
		return sorted[i].MethodName < sorted[j].MethodName
	})

	limit := cr.calculateDisplayLimit(len(sorted))

	fmt.Fprintf(output, "Top %d Misplaced Methods:\n", limit)
	fmt.Fprintf(output, "%-30s %-20s %-25s %-25s\n", "Method", "Receiver Type", "Current File", "Receiver File")
	fmt.Fprintln(output, "--------------------------------------------------------------------------------")

	for i := 0; i < limit; i++ {
		issue := sorted[i]
		fmt.Fprintf(output, "%-30s %-20s %-25s %-25s\n",
			cr.truncate(issue.MethodName, 30),
			cr.truncate(issue.ReceiverType, 20),
			cr.truncate(issue.CurrentFile, 25),
			cr.truncate(issue.ReceiverFile, 25),
		)
	}
	fmt.Fprintln(output)
}

// writeFileCohesionIssues displays file cohesion issues
func (cr *ConsoleReporter) writeFileCohesionIssues(output io.Writer, issues []metrics.FileCohesionIssue) {
	sorted := cr.sortCohesionIssues(issues)
	limit := cr.calculateDisplayLimit(len(sorted))
	cr.writeCohesionHeader(output, limit)
	cr.writeCohesionRows(output, sorted, limit)
}

func (cr *ConsoleReporter) sortCohesionIssues(issues []metrics.FileCohesionIssue) []metrics.FileCohesionIssue {
	sorted := make([]metrics.FileCohesionIssue, len(issues))
	copy(sorted, issues)
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].Severity != sorted[j].Severity {
			return severityWeight(sorted[i].Severity) > severityWeight(sorted[j].Severity)
		}
		return sorted[i].CohesionScore < sorted[j].CohesionScore
	})
	return sorted
}

func (cr *ConsoleReporter) writeCohesionHeader(output io.Writer, limit int) {
	fmt.Fprintf(output, "Top %d Low Cohesion Files:\n", limit)
	fmt.Fprintf(output, "%-40s %-12s %s\n", "File", "Cohesion", "Suggested Splits")
	fmt.Fprintln(output, "--------------------------------------------------------------------------------")
}

func (cr *ConsoleReporter) writeCohesionRows(output io.Writer, sorted []metrics.FileCohesionIssue, limit int) {
	for i := 0; i < limit; i++ {
		cr.writeCohesionRow(output, sorted[i])
	}
	fmt.Fprintln(output)
}

func (cr *ConsoleReporter) writeCohesionRow(output io.Writer, issue metrics.FileCohesionIssue) {
	splits := cr.formatSuggestedSplits(issue.SuggestedSplits)
	fmt.Fprintf(output, "%-40s %-12.2f %s\n",
		cr.truncate(issue.File, 40),
		issue.CohesionScore,
		splits,
	)
}

func (cr *ConsoleReporter) formatSuggestedSplits(splits []string) string {
	if len(splits) == 0 {
		return ""
	}
	result := splits[0]
	if len(splits) > 1 {
		result += fmt.Sprintf(" (+%d more)", len(splits)-1)
	}
	return result
}

// writeDocumentationAnalysis generates documentation analysis output
func (cr *ConsoleReporter) writeDocumentationAnalysis(output io.Writer, report *metrics.Report) {
	fmt.Fprintln(output, "=== DOCUMENTATION ANALYSIS ===")

	doc := report.Documentation

	// Coverage summary
	fmt.Fprintf(output, "Overall Coverage: %.1f%%\n", doc.Coverage.Overall)
	fmt.Fprintf(output, "Package Coverage: %.1f%%\n", doc.Coverage.Packages)
	fmt.Fprintf(output, "Function Coverage: %.1f%%\n", doc.Coverage.Functions)
	fmt.Fprintf(output, "Type Coverage: %.1f%%\n", doc.Coverage.Types)
	fmt.Fprintf(output, "Method Coverage: %.1f%%\n", doc.Coverage.Methods)
	fmt.Fprintln(output)

	// Annotation summary
	totalAnnotations := len(doc.TODOComments) + len(doc.FIXMEComments) + len(doc.HACKComments) + len(doc.BUGComments) + len(doc.XXXComments) + len(doc.DEPRECATEDComments) + len(doc.NOTEComments)
	if totalAnnotations > 0 {
		fmt.Fprintln(output, "Annotation Summary:")
		fmt.Fprintf(output, "  TODO: %d\n", len(doc.TODOComments))
		fmt.Fprintf(output, "  FIXME: %d (critical)\n", len(doc.FIXMEComments))
		fmt.Fprintf(output, "  HACK: %d\n", len(doc.HACKComments))
		fmt.Fprintf(output, "  BUG: %d (critical)\n", len(doc.BUGComments))
		fmt.Fprintf(output, "  XXX: %d\n", len(doc.XXXComments))
		fmt.Fprintf(output, "  DEPRECATED: %d\n", len(doc.DEPRECATEDComments))
		fmt.Fprintf(output, "  NOTE: %d\n", len(doc.NOTEComments))
		fmt.Fprintf(output, "  Total: %d\n", totalAnnotations)
		fmt.Fprintln(output)

		// Show top annotations by severity
		cr.writeTopAnnotations(output, doc)
	}

	fmt.Fprintln(output)
}

// annotationItem represents a code annotation with its metadata for console display.
type annotationItem struct {
	category string
	file     string
	line     int
	desc     string
	severity string
}

// collectAnnotations gathers all annotations from documentation metrics
func collectAnnotations(doc metrics.DocumentationMetrics) []annotationItem {
	var annotations []annotationItem

	for _, c := range doc.FIXMEComments {
		annotations = append(annotations, annotationItem{"FIXME", c.File, c.Line, c.Description, "critical"})
	}
	for _, c := range doc.BUGComments {
		annotations = append(annotations, annotationItem{"BUG", c.File, c.Line, c.Description, "critical"})
	}
	for _, c := range doc.HACKComments {
		annotations = append(annotations, annotationItem{"HACK", c.File, c.Line, c.Reason, "high"})
	}
	for _, c := range doc.TODOComments {
		annotations = append(annotations, annotationItem{"TODO", c.File, c.Line, c.Description, "medium"})
	}
	for _, c := range doc.XXXComments {
		annotations = append(annotations, annotationItem{"XXX", c.File, c.Line, c.Description, "medium"})
	}

	return annotations
}

// writeTopAnnotations displays top annotations by severity
func (cr *ConsoleReporter) writeTopAnnotations(output io.Writer, doc metrics.DocumentationMetrics) {
	annotations := collectAnnotations(doc)
	if len(annotations) == 0 {
		return
	}

	limit := 10
	if limit > len(annotations) {
		limit = len(annotations)
	}

	fmt.Fprintf(output, "Top %d Annotations by Severity:\n", limit)
	fmt.Fprintf(output, "%-10s %-50s %-6s %s\n", "Category", "File", "Line", "Description")
	fmt.Fprintln(output, "--------------------------------------------------------------------------------")

	for i := 0; i < limit; i++ {
		a := annotations[i]
		fmt.Fprintf(output, "%-10s %-50s %-6d %s\n",
			a.category,
			cr.truncate(a.file, 50),
			a.line,
			cr.truncate(a.desc, 40),
		)
	}
	fmt.Fprintln(output)
}
