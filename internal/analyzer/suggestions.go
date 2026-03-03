package analyzer

import (
	"fmt"
	"sort"

	"github.com/opd-ai/go-stats-generator/internal/metrics"
)

// ActionType represents the type of refactoring action suggested.
// Each action type corresponds to a specific code improvement strategy
// that can be applied to reduce maintenance burden.
type ActionType string

const (
	// ActionExtractFunction suggests extracting code into a separate function
	ActionExtractFunction ActionType = "extract_function"
	// ActionRename suggests renaming an identifier to follow conventions
	ActionRename ActionType = "rename"
	// ActionMoveToFile suggests moving code to a different file for better cohesion
	ActionMoveToFile ActionType = "move_to_file"
	// ActionAddDocumentation suggests adding missing GoDoc comments
	ActionAddDocumentation ActionType = "add_documentation"
	// ActionReduceComplexity suggests simplifying complex logic
	ActionReduceComplexity ActionType = "reduce_complexity"
	// ActionDeduplicate suggests extracting duplicated code
	ActionDeduplicate ActionType = "deduplicate"
)

// EffortLevel represents estimated implementation effort.
// Effort classification helps prioritize suggestions based on
// available time and resources.
type EffortLevel string

const (
	// EffortLow represents tasks taking <1 hour with <30 LoC changed
	EffortLow EffortLevel = "low" // <1 hour, <30 LoC changed
	// EffortMedium represents tasks taking 1-4 hours with 30-100 LoC changed
	EffortMedium EffortLevel = "medium" // 1-4 hours, 30-100 LoC
	// EffortHigh represents tasks taking >4 hours with >100 LoC changed
	EffortHigh EffortLevel = "high" // >4 hours, >100 LoC
)

// RefactoringSuggestion represents an actionable code improvement.
// Each suggestion includes impact estimates, effort classification,
// and prioritization metrics to help developers make informed decisions
// about which technical debt to address first.
type RefactoringSuggestion struct {
	Action        ActionType  `json:"action"`         // Type of refactoring action
	Target        string      `json:"target"`         // File, function, or symbol affected
	Location      string      `json:"location"`       // File:Line or Package/File
	Description   string      `json:"description"`    // Human-readable explanation
	Effort        EffortLevel `json:"effort"`         // Estimated effort to implement
	MBIImpact     float64     `json:"mbi_impact"`     // Estimated MBI reduction
	ImpactEffort  float64     `json:"impact_effort"`  // ROI ratio (higher = better)
	Category      string      `json:"category"`       // duplication, naming, etc.
	AffectedLines int         `json:"affected_lines"` // Estimated lines changed
}

// SuggestionGenerator creates prioritized refactoring suggestions.
// It analyzes metrics reports and generates actionable recommendations
// sorted by impact-to-effort ratio to maximize return on investment.
type SuggestionGenerator struct {
	scorer *ScoringAnalyzer
}

// NewSuggestionGenerator creates a new suggestion generator.
// The generator uses the provided scorer to estimate MBI impact
// for each suggestion.
func NewSuggestionGenerator(scorer *ScoringAnalyzer) *SuggestionGenerator {
	return &SuggestionGenerator{
		scorer: scorer,
	}
}

// GenerateSuggestions creates prioritized refactoring suggestions from a report.
// It analyzes all maintenance burden categories (duplication, complexity,
// documentation, naming, placement, organization) and generates actionable
// recommendations sorted by impact-to-effort ratio.
//
// The returned slice is ordered from highest to lowest ROI, making it easy
// to identify the most valuable improvements to implement first.
func (sg *SuggestionGenerator) GenerateSuggestions(report *metrics.Report) []RefactoringSuggestion {
	suggestions := make([]RefactoringSuggestion, 0)

	// Generate category-specific suggestions
	suggestions = append(suggestions, sg.generateDuplicationSuggestions(report)...)
	suggestions = append(suggestions, sg.generateComplexitySuggestions(report)...)
	suggestions = append(suggestions, sg.generateDocumentationSuggestions(report)...)
	suggestions = append(suggestions, sg.generateNamingSuggestions(report)...)
	suggestions = append(suggestions, sg.generatePlacementSuggestions(report)...)
	suggestions = append(suggestions, sg.generateOrganizationSuggestions(report)...)

	// Calculate impact-effort ratio for all suggestions
	for i := range suggestions {
		suggestions[i].ImpactEffort = sg.calculateImpactEffortRatio(&suggestions[i])
	}

	// Sort by impact/effort ratio (highest ROI first)
	sort.Slice(suggestions, func(i, j int) bool {
		return suggestions[i].ImpactEffort > suggestions[j].ImpactEffort
	})

	return suggestions
}

// calculateImpactEffortRatio computes ROI for a suggestion.
// It converts effort level to hours and divides MBI impact by effort,
// producing a ratio where higher values indicate better return on investment.
func (sg *SuggestionGenerator) calculateImpactEffortRatio(s *RefactoringSuggestion) float64 {
	// Convert effort to hours
	effortHours := 1.0
	switch s.Effort {
	case EffortLow:
		effortHours = 0.5
	case EffortMedium:
		effortHours = 2.0
	case EffortHigh:
		effortHours = 6.0
	}

	// Avoid division by zero
	if effortHours == 0 {
		effortHours = 0.1
	}

	return s.MBIImpact / effortHours
}

// generateDuplicationSuggestions creates suggestions for code clones.
// It identifies clone pairs with 10+ lines and at least 2 instances,
// estimating impact based on clone size and affected lines.
func (sg *SuggestionGenerator) generateDuplicationSuggestions(report *metrics.Report) []RefactoringSuggestion {
	suggestions := make([]RefactoringSuggestion, 0)

	// Only suggest for the largest clone pairs (>= 10 lines)
	for _, clone := range report.Duplication.Clones {
		if clone.LineCount < 10 || len(clone.Instances) < 2 {
			continue
		}

		// Estimate impact: larger clones = higher impact
		impact := float64(clone.LineCount) * 0.5
		if impact > 20 {
			impact = 20
		}

		// Estimate effort based on clone size
		effort := classifyEffort(clone.LineCount)

		loc := clone.Instances[0]
		suggestions = append(suggestions, RefactoringSuggestion{
			Action:        ActionDeduplicate,
			Target:        loc.File,
			Location:      formatLocation(loc.File, loc.StartLine),
			Description:   formatDuplicationDesc(clone),
			Effort:        effort,
			MBIImpact:     impact,
			Category:      "duplication",
			AffectedLines: clone.LineCount * len(clone.Instances),
		})
	}

	return suggestions
}

// generateComplexitySuggestions creates suggestions for high-complexity functions.
// It targets functions with cyclomatic complexity > 15, estimating impact
// based on the potential complexity reduction.
func (sg *SuggestionGenerator) generateComplexitySuggestions(report *metrics.Report) []RefactoringSuggestion {
	suggestions := make([]RefactoringSuggestion, 0)

	for _, fn := range report.Functions {
		// Only suggest for functions with cyclomatic complexity > 15
		if fn.Complexity.Cyclomatic <= 15 {
			continue
		}

		// Estimate impact based on complexity reduction
		impact := float64(fn.Complexity.Cyclomatic-10) * 0.8
		if impact > 25 {
			impact = 25
		}

		// Effort increases with complexity
		effort := EffortMedium
		if fn.Complexity.Cyclomatic > 20 {
			effort = EffortHigh
		}

		suggestions = append(suggestions, RefactoringSuggestion{
			Action:        ActionReduceComplexity,
			Target:        fn.Name,
			Location:      formatLocation(fn.File, fn.Line),
			Description:   formatComplexityDesc(fn),
			Effort:        effort,
			MBIImpact:     impact,
			Category:      "burden",
			AffectedLines: estimateRefactoringLines(fn.Lines.Code),
		})
	}

	return suggestions
}

// generateDocumentationSuggestions creates suggestions for missing documentation.
// It identifies exported functions without GoDoc comments and estimates
// the MBI impact of adding documentation.
func (sg *SuggestionGenerator) generateDocumentationSuggestions(report *metrics.Report) []RefactoringSuggestion {
	suggestions := make([]RefactoringSuggestion, 0)

	for _, fn := range report.Functions {
		// Only suggest for exported functions without documentation
		if fn.Documentation.HasComment || !isExported(fn.Name) {
			continue
		}

		// Documentation has moderate impact
		impact := 5.0

		suggestions = append(suggestions, RefactoringSuggestion{
			Action:        ActionAddDocumentation,
			Target:        fn.Name,
			Location:      formatLocation(fn.File, fn.Line),
			Description:   formatDocDesc(fn.Name),
			Effort:        EffortLow,
			MBIImpact:     impact,
			Category:      "documentation",
			AffectedLines: 3, // Typical doc comment size
		})
	}

	return suggestions
}

// generateNamingSuggestions creates suggestions for naming violations.
// It processes identifier issues from the naming analyzer and generates
// rename suggestions to improve code clarity.
func (sg *SuggestionGenerator) generateNamingSuggestions(report *metrics.Report) []RefactoringSuggestion {
	suggestions := make([]RefactoringSuggestion, 0)

	// Identifier violations
	for _, issue := range report.Naming.IdentifierIssues {
		suggestions = append(suggestions, RefactoringSuggestion{
			Action:        ActionRename,
			Target:        issue.Name,
			Location:      formatLocation(issue.File, issue.Line),
			Description:   formatNamingDesc(issue),
			Effort:        EffortLow,
			MBIImpact:     3.0,
			Category:      "naming",
			AffectedLines: 5,
		})
	}

	return suggestions
}

// generatePlacementSuggestions creates suggestions for misplaced functions.
// It identifies functions with high affinity for other files (>0.5 gain)
// and suggests relocating them for better cohesion.
func (sg *SuggestionGenerator) generatePlacementSuggestions(report *metrics.Report) []RefactoringSuggestion {
	suggestions := make([]RefactoringSuggestion, 0)

	// Only suggest for high-affinity misplacements
	for _, issue := range report.Placement.FunctionIssues {
		affinityGain := issue.SuggestedAffinity - issue.CurrentAffinity
		if affinityGain < 0.5 {
			continue
		}

		suggestions = append(suggestions, RefactoringSuggestion{
			Action:        ActionMoveToFile,
			Target:        issue.Name,
			Location:      issue.CurrentFile,
			Description:   formatPlacementDesc(issue),
			Effort:        EffortLow,
			MBIImpact:     affinityGain * 10,
			Category:      "placement",
			AffectedLines: 10,
		})
	}

	return suggestions
}

// generateOrganizationSuggestions creates suggestions for organizational issues.
// It identifies oversized files (>500 lines) and suggests splitting them
// into smaller, more maintainable modules.
func (sg *SuggestionGenerator) generateOrganizationSuggestions(report *metrics.Report) []RefactoringSuggestion {
	suggestions := make([]RefactoringSuggestion, 0)

	// Oversized files
	for _, file := range report.Organization.OversizedFiles {
		if file.Lines.Total < 500 {
			continue
		}

		impact := (float64(file.Lines.Total) - 500) / 50
		if impact > 30 {
			impact = 30
		}

		suggestions = append(suggestions, RefactoringSuggestion{
			Action:        ActionExtractFunction,
			Target:        file.File,
			Location:      file.File,
			Description:   formatOversizedDesc(file),
			Effort:        EffortHigh,
			MBIImpact:     impact,
			Category:      "organization",
			AffectedLines: file.Lines.Total / 3,
		})
	}

	return suggestions
}

// Helper functions for formatting and classification

// classifyEffort determines effort level based on lines of code.
// Returns low (<30 lines), medium (30-100 lines), or high (>100 lines).
func classifyEffort(lines int) EffortLevel {
	if lines < 30 {
		return EffortLow
	} else if lines < 100 {
		return EffortMedium
	}
	return EffortHigh
}

// estimateRefactoringLines estimates how many lines will be modified.
// Assumes refactoring touches approximately 60% of original lines.
func estimateRefactoringLines(originalLines int) int {
	// Assume refactoring touches ~60% of lines
	return int(float64(originalLines) * 0.6)
}

// isExported checks if a name is exported (starts with uppercase letter).
func isExported(name string) bool {
	if len(name) == 0 {
		return false
	}
	firstChar := name[0]
	return firstChar >= 'A' && firstChar <= 'Z'
}

// formatLocation formats a file and line number as "file:line".
func formatLocation(file string, line int) string {
	return fmt.Sprintf("%s:%d", file, line)
}

// formatDuplicationDesc generates a description for duplication suggestions.
func formatDuplicationDesc(clone metrics.ClonePair) string {
	return fmt.Sprintf("Extract duplicated code block (%d lines) to shared function", clone.LineCount)
}

// formatComplexityDesc generates a description for complexity suggestions.
func formatComplexityDesc(fn metrics.FunctionMetrics) string {
	return fmt.Sprintf("Split function into smaller helpers (current complexity: %d)", fn.Complexity.Cyclomatic)
}

// formatDocDesc generates a description for documentation suggestions.
func formatDocDesc(name string) string {
	return fmt.Sprintf("Add GoDoc comment for exported function %s", name)
}

// formatNamingDesc generates a description for naming suggestions.
func formatNamingDesc(issue metrics.IdentifierViolation) string {
	return fmt.Sprintf("Rename to follow Go conventions: %s", issue.Description)
}

// formatPlacementDesc generates a description for placement suggestions.
func formatPlacementDesc(issue metrics.MisplacedFunctionIssue) string {
	return fmt.Sprintf("Move to %s for better cohesion", issue.SuggestedFile)
}

// formatOversizedDesc generates a description for oversized file suggestions.
func formatOversizedDesc(file metrics.OversizedFile) string {
	return fmt.Sprintf("Split file into smaller modules (current: %d lines)", file.Lines.Total)
}
