package metrics

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"
)

// DiffOptions configures how the diff is performed
type DiffOptions struct {
	ThresholdPercent float64 // Minimum percentage change to consider significant
	ShowOnlyChanges  bool    // Only show items that changed
	Granularity      ChangeGranularity
}

// DefaultDiffOptions returns sensible default options for diffing
func DefaultDiffOptions() DiffOptions {
	return DiffOptions{
		ThresholdPercent: 5.0, // 5% threshold for significance
		ShowOnlyChanges:  false,
		Granularity:      DefaultChangeGranularity(),
	}
}

// CompareSnapshots compares two metrics snapshots and returns a comprehensive diff
func CompareSnapshots(baseline, current MetricsSnapshot, config ThresholdConfig) (*ComplexityDiff, error) {
	if baseline.ID == "" || current.ID == "" {
		return nil, fmt.Errorf("both baseline and current snapshots must have valid IDs")
	}

	diff := &ComplexityDiff{
		Baseline:  baseline,
		Current:   current,
		Timestamp: time.Now(),
		Config:    config,
	}

	// Generate changes, regressions, and improvements
	changes := generateMetricChanges(baseline.Report, current.Report, config)
	diff.Changes = changes

	// Categorize changes into regressions and improvements
	regressions, improvements := categorizeChanges(changes, config)
	diff.Regressions = regressions
	diff.Improvements = improvements

	// Generate summary
	diff.Summary = generateDiffSummary(changes, regressions, improvements)

	return diff, nil
}

// generateMetricChanges compares reports and generates detailed metric changes
func generateMetricChanges(baseline, current Report, config ThresholdConfig) []MetricChange {
	var changes []MetricChange

	// Compare function metrics
	funcChanges := compareFunctionMetrics(baseline.Functions, current.Functions, config)
	changes = append(changes, funcChanges...)

	// Compare struct metrics
	structChanges := compareStructMetrics(baseline.Structs, current.Structs, config)
	changes = append(changes, structChanges...)

	// Compare package metrics
	packageChanges := comparePackageMetrics(baseline.Packages, current.Packages, config)
	changes = append(changes, packageChanges...)

	// Compare overall complexity
	complexityChanges := compareComplexityMetrics(baseline.Complexity, current.Complexity, config)
	changes = append(changes, complexityChanges...)

	return changes
}

// compareFunctionMetrics compares function metrics between reports
func compareFunctionMetrics(baseline, current []FunctionMetrics, config ThresholdConfig) []MetricChange {
	var changes []MetricChange

	// Create maps for efficient lookup
	baselineMap := make(map[string]FunctionMetrics)
	currentMap := make(map[string]FunctionMetrics)

	for _, f := range baseline {
		key := fmt.Sprintf("%s.%s", f.Package, f.Name)
		baselineMap[key] = f
	}

	for _, f := range current {
		key := fmt.Sprintf("%s.%s", f.Package, f.Name)
		currentMap[key] = f
	}

	// Find all unique functions
	allKeys := make(map[string]bool)
	for key := range baselineMap {
		allKeys[key] = true
	}
	for key := range currentMap {
		allKeys[key] = true
	}

	for key := range allKeys {
		baseFunc, hasBaseline := baselineMap[key]
		currFunc, hasCurrent := currentMap[key]

		if hasBaseline && hasCurrent {
			// Function exists in both - compare complexity
			changes = append(changes, compareFunctionComplexity(baseFunc, currFunc, config)...)
		} else if hasBaseline && !hasCurrent {
			// Function was removed
			changes = append(changes, MetricChange{
				Category:    "function",
				Name:        baseFunc.Name,
				Path:        fmt.Sprintf("%s.%s", baseFunc.Package, baseFunc.Name),
				File:        baseFunc.File,
				Line:        baseFunc.Line,
				OldValue:    baseFunc.Complexity.Overall,
				NewValue:    nil,
				Delta:       Delta{Direction: ChangeDirectionDecrease, Significant: true, Magnitude: ChangeMagnitudeMajor},
				Impact:      ImpactLevelMedium,
				Severity:    SeverityLevelWarning,
				Description: "Function removed",
				Suggestion:  "Verify function removal was intentional",
			})
		} else if !hasBaseline && hasCurrent {
			// Function was added
			changes = append(changes, MetricChange{
				Category:    "function",
				Name:        currFunc.Name,
				Path:        fmt.Sprintf("%s.%s", currFunc.Package, currFunc.Name),
				File:        currFunc.File,
				Line:        currFunc.Line,
				OldValue:    nil,
				NewValue:    currFunc.Complexity.Overall,
				Delta:       Delta{Direction: ChangeDirectionIncrease, Significant: true, Magnitude: ChangeMagnitudeModerate},
				Impact:      ImpactLevelLow,
				Severity:    SeverityLevelInfo,
				Description: "Function added",
				Suggestion:  "Review new function complexity",
			})
		}
	}

	return changes
}

// compareFunctionComplexity compares complexity between two function versions
func compareFunctionComplexity(baseline, current FunctionMetrics, config ThresholdConfig) []MetricChange {
	var changes []MetricChange

	// Compare cyclomatic complexity
	if baseline.Complexity.Cyclomatic != current.Complexity.Cyclomatic {
		delta := calculateDelta(float64(baseline.Complexity.Cyclomatic), float64(current.Complexity.Cyclomatic), config.Global.SignificanceLevel)

		change := MetricChange{
			Category:    "function_complexity",
			Name:        current.Name,
			Path:        fmt.Sprintf("%s.%s", current.Package, current.Name),
			File:        current.File,
			Line:        current.Line,
			OldValue:    baseline.Complexity.Cyclomatic,
			NewValue:    current.Complexity.Cyclomatic,
			Delta:       delta,
			Description: "Cyclomatic complexity changed",
		}

		// Determine impact and severity
		if current.Complexity.Cyclomatic > config.FunctionComplexity.Error {
			change.Impact = ImpactLevelCritical
			change.Severity = SeverityLevelCritical
			change.Suggestion = "Critical: Function complexity exceeds error threshold, refactoring required"
		} else if current.Complexity.Cyclomatic > config.FunctionComplexity.Warning {
			change.Impact = ImpactLevelHigh
			change.Severity = SeverityLevelWarning
			change.Suggestion = "Warning: Function complexity exceeds warning threshold"
		} else if delta.Direction == ChangeDirectionIncrease && delta.Percentage > config.FunctionComplexity.MaxIncrease {
			change.Impact = ImpactLevelMedium
			change.Severity = SeverityLevelWarning
			change.Suggestion = "Complexity increase exceeds threshold"
		} else {
			change.Impact = ImpactLevelLow
			change.Severity = SeverityLevelInfo
		}

		changes = append(changes, change)
	}

	// Compare overall complexity
	if baseline.Complexity.Overall != current.Complexity.Overall {
		delta := calculateDelta(baseline.Complexity.Overall, current.Complexity.Overall, config.Global.SignificanceLevel)

		changes = append(changes, MetricChange{
			Category:    "function_overall_complexity",
			Name:        current.Name,
			Path:        fmt.Sprintf("%s.%s", current.Package, current.Name),
			File:        current.File,
			Line:        current.Line,
			OldValue:    baseline.Complexity.Overall,
			NewValue:    current.Complexity.Overall,
			Delta:       delta,
			Impact:      determineImpactLevel(delta),
			Severity:    determineSeverityLevel(delta),
			Description: "Overall function complexity changed",
		})
	}

	return changes
}

// compareStructMetrics compares struct metrics between reports
func compareStructMetrics(baseline, current []StructMetrics, config ThresholdConfig) []MetricChange {
	var changes []MetricChange

	baselineMap := make(map[string]StructMetrics)
	currentMap := make(map[string]StructMetrics)

	for _, s := range baseline {
		key := fmt.Sprintf("%s.%s", s.Package, s.Name)
		baselineMap[key] = s
	}

	for _, s := range current {
		key := fmt.Sprintf("%s.%s", s.Package, s.Name)
		currentMap[key] = s
	}

	// Find all unique structs
	allKeys := make(map[string]bool)
	for key := range baselineMap {
		allKeys[key] = true
	}
	for key := range currentMap {
		allKeys[key] = true
	}

	for key := range allKeys {
		baseStruct, hasBaseline := baselineMap[key]
		currStruct, hasCurrent := currentMap[key]

		if hasBaseline && hasCurrent {
			// Compare field count
			if baseStruct.TotalFields != currStruct.TotalFields {
				delta := calculateDelta(float64(baseStruct.TotalFields), float64(currStruct.TotalFields), config.Global.SignificanceLevel)

				change := MetricChange{
					Category:    "struct_fields",
					Name:        currStruct.Name,
					Path:        fmt.Sprintf("%s.%s", currStruct.Package, currStruct.Name),
					File:        currStruct.File,
					Line:        currStruct.Line,
					OldValue:    baseStruct.TotalFields,
					NewValue:    currStruct.TotalFields,
					Delta:       delta,
					Description: "Struct field count changed",
				}

				if currStruct.TotalFields > config.StructComplexity.MaxFields {
					change.Impact = ImpactLevelHigh
					change.Severity = SeverityLevelWarning
					change.Suggestion = "Struct has too many fields, consider composition"
				} else if delta.Direction == ChangeDirectionIncrease && delta.Percentage > config.StructComplexity.FieldIncrease {
					change.Impact = ImpactLevelMedium
					change.Severity = SeverityLevelWarning
					change.Suggestion = "Field count increase exceeds threshold"
				} else {
					change.Impact = determineImpactLevel(delta)
					change.Severity = determineSeverityLevel(delta)
				}

				changes = append(changes, change)
			}
		} else if hasBaseline && !hasCurrent {
			// Struct removed
			changes = append(changes, MetricChange{
				Category:    "struct",
				Name:        baseStruct.Name,
				Path:        fmt.Sprintf("%s.%s", baseStruct.Package, baseStruct.Name),
				File:        baseStruct.File,
				Line:        baseStruct.Line,
				OldValue:    baseStruct.TotalFields,
				NewValue:    nil,
				Delta:       Delta{Direction: ChangeDirectionDecrease, Significant: true, Magnitude: ChangeMagnitudeMajor},
				Impact:      ImpactLevelMedium,
				Severity:    SeverityLevelWarning,
				Description: "Struct removed",
			})
		} else if !hasBaseline && hasCurrent {
			// Struct added
			changes = append(changes, MetricChange{
				Category:    "struct",
				Name:        currStruct.Name,
				Path:        fmt.Sprintf("%s.%s", currStruct.Package, currStruct.Name),
				File:        currStruct.File,
				Line:        currStruct.Line,
				OldValue:    nil,
				NewValue:    currStruct.TotalFields,
				Delta:       Delta{Direction: ChangeDirectionIncrease, Significant: true, Magnitude: ChangeMagnitudeModerate},
				Impact:      ImpactLevelLow,
				Severity:    SeverityLevelInfo,
				Description: "Struct added",
			})
		}
	}

	return changes
}

// comparePackageMetrics compares package metrics between reports
func comparePackageMetrics(baseline, current []PackageMetrics, config ThresholdConfig) []MetricChange {
	var changes []MetricChange

	baselineMap := make(map[string]PackageMetrics)
	currentMap := make(map[string]PackageMetrics)

	for _, p := range baseline {
		baselineMap[p.Path] = p
	}

	for _, p := range current {
		currentMap[p.Path] = p
	}

	// Find all unique packages
	allPaths := make(map[string]bool)
	for path := range baselineMap {
		allPaths[path] = true
	}
	for path := range currentMap {
		allPaths[path] = true
	}

	for path := range allPaths {
		basePkg, hasBaseline := baselineMap[path]
		currPkg, hasCurrent := currentMap[path]

		if hasBaseline && hasCurrent {
			// Compare coupling score
			if basePkg.CouplingScore != currPkg.CouplingScore {
				delta := calculateDelta(basePkg.CouplingScore, currPkg.CouplingScore, config.Global.SignificanceLevel)

				change := MetricChange{
					Category:    "package_coupling",
					Name:        currPkg.Name,
					Path:        currPkg.Path,
					OldValue:    basePkg.CouplingScore,
					NewValue:    currPkg.CouplingScore,
					Delta:       delta,
					Description: "Package coupling changed",
				}

				if currPkg.CouplingScore > config.PackageMetrics.MaxCoupling {
					change.Impact = ImpactLevelCritical
					change.Severity = SeverityLevelError
					change.Suggestion = "Package coupling exceeds threshold, reduce dependencies"
				} else {
					change.Impact = determineImpactLevel(delta)
					change.Severity = determineSeverityLevel(delta)
				}

				changes = append(changes, change)
			}

			// Compare cohesion score
			if basePkg.CohesionScore != currPkg.CohesionScore {
				delta := calculateDelta(basePkg.CohesionScore, currPkg.CohesionScore, config.Global.SignificanceLevel)

				change := MetricChange{
					Category:    "package_cohesion",
					Name:        currPkg.Name,
					Path:        currPkg.Path,
					OldValue:    basePkg.CohesionScore,
					NewValue:    currPkg.CohesionScore,
					Delta:       delta,
					Description: "Package cohesion changed",
				}

				if currPkg.CohesionScore < config.PackageMetrics.MinCohesion {
					change.Impact = ImpactLevelHigh
					change.Severity = SeverityLevelWarning
					change.Suggestion = "Package cohesion below threshold, group related functionality"
				} else {
					change.Impact = determineImpactLevel(delta)
					change.Severity = determineSeverityLevel(delta)
				}

				changes = append(changes, change)
			}
		}
	}

	return changes
}

// compareComplexityMetrics compares overall complexity metrics
func compareComplexityMetrics(baseline, current ComplexityMetrics, config ThresholdConfig) []MetricChange {
	var changes []MetricChange

	// Compare average function complexity
	if baseline.AverageFunction != current.AverageFunction {
		delta := calculateDelta(baseline.AverageFunction, current.AverageFunction, config.Global.SignificanceLevel)

		changes = append(changes, MetricChange{
			Category:    "overall_complexity",
			Name:        "average_function_complexity",
			Path:        "global",
			OldValue:    baseline.AverageFunction,
			NewValue:    current.AverageFunction,
			Delta:       delta,
			Impact:      determineImpactLevel(delta),
			Severity:    determineSeverityLevel(delta),
			Description: "Average function complexity changed",
		})
	}

	// Compare average struct complexity
	if baseline.AverageStruct != current.AverageStruct {
		delta := calculateDelta(baseline.AverageStruct, current.AverageStruct, config.Global.SignificanceLevel)

		changes = append(changes, MetricChange{
			Category:    "overall_complexity",
			Name:        "average_struct_complexity",
			Path:        "global",
			OldValue:    baseline.AverageStruct,
			NewValue:    current.AverageStruct,
			Delta:       delta,
			Impact:      determineImpactLevel(delta),
			Severity:    determineSeverityLevel(delta),
			Description: "Average struct complexity changed",
		})
	}

	return changes
}

// categorizeChanges separates changes into regressions and improvements
func categorizeChanges(changes []MetricChange, config ThresholdConfig) ([]Regression, []Improvement) {
	var regressions []Regression
	var improvements []Improvement

	for _, change := range changes {
		if isRegression(change, config) {
			regressions = append(regressions, Regression{
				Type:        categorizeRegressionType(change),
				Location:    change.Path,
				File:        change.File,
				Line:        change.Line,
				Function:    extractFunctionName(change.Path),
				Description: change.Description,
				OldValue:    change.OldValue,
				NewValue:    change.NewValue,
				Delta:       change.Delta,
				Impact:      change.Impact,
				Severity:    change.Severity,
				Suggestion:  change.Suggestion,
				Priority:    calculateRegressionPriority(change),
			})
		} else if isImprovement(change) {
			improvements = append(improvements, Improvement{
				Type:        categorizeImprovementType(change),
				Location:    change.Path,
				File:        change.File,
				Line:        change.Line,
				Function:    extractFunctionName(change.Path),
				Description: change.Description,
				OldValue:    change.OldValue,
				NewValue:    change.NewValue,
				Delta:       change.Delta,
				Impact:      change.Impact,
				Benefit:     generateBenefitDescription(change),
			})
		}
	}

	// Sort regressions by priority (highest first)
	sort.Slice(regressions, func(i, j int) bool {
		return regressions[i].Priority > regressions[j].Priority
	})

	return regressions, improvements
}

// generateDiffSummary creates a summary of all changes
func generateDiffSummary(changes []MetricChange, regressions []Regression, improvements []Improvement) DiffSummary {
	summary := DiffSummary{
		TotalChanges:       len(changes),
		RegressionCount:    len(regressions),
		ImprovementCount:   len(improvements),
		NeutralChangeCount: len(changes) - len(regressions) - len(improvements),
	}

	// Count significant changes
	for _, change := range changes {
		if change.Delta.Significant {
			summary.SignificantChanges++
		}
	}

	// Count critical issues
	for _, regression := range regressions {
		if regression.Severity == SeverityLevelCritical {
			summary.CriticalIssues++
		}
	}

	// Determine overall trend
	if len(regressions) > len(improvements) {
		summary.OverallTrend = TrendDegrading
	} else if len(improvements) > len(regressions) {
		summary.OverallTrend = TrendImproving
	} else {
		summary.OverallTrend = TrendStable
	}

	// Calculate quality score (0-100, higher is better)
	totalSignificantChanges := summary.SignificantChanges
	if totalSignificantChanges == 0 {
		summary.QualityScore = 100.0
	} else {
		improvementRatio := float64(len(improvements)) / float64(totalSignificantChanges)
		summary.QualityScore = improvementRatio * 100.0
	}

	return summary
}

// Helper functions

func calculateDelta(oldValue, newValue, threshold float64) Delta {
	absolute := newValue - oldValue
	var percentage float64

	if oldValue != 0 {
		percentage = math.Abs(absolute) / math.Abs(oldValue) * 100
	} else if newValue != 0 {
		percentage = 100.0 // New value from zero
	}

	direction := ChangeDirectionNeutral
	if absolute > 0 {
		direction = ChangeDirectionIncrease
	} else if absolute < 0 {
		direction = ChangeDirectionDecrease
	}

	magnitude := ChangeMagnitudeMinor
	if percentage >= 50 {
		magnitude = ChangeMagnitudeCritical
	} else if percentage >= 25 {
		magnitude = ChangeMagnitudeMajor
	} else if percentage >= 10 {
		magnitude = ChangeMagnitudeSignificant
	} else if percentage >= 5 {
		magnitude = ChangeMagnitudeModerate
	}

	return Delta{
		Absolute:    absolute,
		Percentage:  percentage,
		Direction:   direction,
		Significant: percentage >= threshold,
		Magnitude:   magnitude,
	}
}

func determineImpactLevel(delta Delta) ImpactLevel {
	if delta.Magnitude == ChangeMagnitudeCritical {
		return ImpactLevelCritical
	} else if delta.Magnitude == ChangeMagnitudeMajor {
		return ImpactLevelHigh
	} else if delta.Magnitude == ChangeMagnitudeSignificant {
		return ImpactLevelMedium
	}
	return ImpactLevelLow
}

func determineSeverityLevel(delta Delta) SeverityLevel {
	if delta.Magnitude == ChangeMagnitudeCritical {
		return SeverityLevelCritical
	} else if delta.Magnitude == ChangeMagnitudeMajor {
		return SeverityLevelError
	} else if delta.Magnitude == ChangeMagnitudeSignificant {
		return SeverityLevelWarning
	}
	return SeverityLevelInfo
}

func isRegression(change MetricChange, config ThresholdConfig) bool {
	// Consider it a regression if it's a negative change that exceeds thresholds
	return change.Delta.Direction == ChangeDirectionIncrease &&
		change.Delta.Significant &&
		(change.Severity == SeverityLevelWarning ||
			change.Severity == SeverityLevelError ||
			change.Severity == SeverityLevelCritical)
}

func isImprovement(change MetricChange) bool {
	// Consider it an improvement if it's a positive change
	return change.Delta.Direction == ChangeDirectionDecrease &&
		change.Delta.Significant
}

func categorizeRegressionType(change MetricChange) RegressionType {
	switch {
	case strings.Contains(change.Category, "complexity"):
		return ComplexityRegression
	case strings.Contains(change.Category, "coupling"):
		return CouplingRegression
	case strings.Contains(change.Category, "cohesion"):
		return CohesionRegression
	case strings.Contains(change.Category, "size") || strings.Contains(change.Category, "field"):
		return SizeRegression
	case strings.Contains(change.Category, "documentation"):
		return DocumentationRegression
	default:
		return ComplexityRegression
	}
}

func categorizeImprovementType(change MetricChange) ImprovementType {
	switch {
	case strings.Contains(change.Category, "complexity"):
		return ComplexityImprovement
	case strings.Contains(change.Category, "coupling"):
		return CouplingImprovement
	case strings.Contains(change.Category, "cohesion"):
		return CohesionImprovement
	case strings.Contains(change.Category, "size") || strings.Contains(change.Category, "field"):
		return SizeImprovement
	case strings.Contains(change.Category, "documentation"):
		return DocumentationImprovement
	default:
		return ComplexityImprovement
	}
}

func calculateRegressionPriority(change MetricChange) int {
	priority := 1

	// Base priority on severity
	switch change.Severity {
	case SeverityLevelCritical:
		priority += 8
	case SeverityLevelError:
		priority += 6
	case SeverityLevelWarning:
		priority += 4
	case SeverityLevelInfo:
		priority += 2
	}

	// Adjust based on magnitude
	switch change.Delta.Magnitude {
	case ChangeMagnitudeCritical:
		priority += 4
	case ChangeMagnitudeMajor:
		priority += 3
	case ChangeMagnitudeSignificant:
		priority += 2
	case ChangeMagnitudeModerate:
		priority += 1
	}

	// Cap at 10
	if priority > 10 {
		priority = 10
	}

	return priority
}

func generateBenefitDescription(change MetricChange) string {
	switch change.Delta.Magnitude {
	case ChangeMagnitudeCritical:
		return "Major improvement in code quality"
	case ChangeMagnitudeMajor:
		return "Significant improvement in maintainability"
	case ChangeMagnitudeSignificant:
		return "Notable improvement in code structure"
	case ChangeMagnitudeModerate:
		return "Moderate improvement in code quality"
	default:
		return "Minor improvement detected"
	}
}

func extractFunctionName(path string) string {
	parts := strings.Split(path, ".")
	if len(parts) > 1 {
		return parts[len(parts)-1]
	}
	return ""
}
