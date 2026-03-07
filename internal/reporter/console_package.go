package reporter

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/opd-ai/go-stats-generator/internal/metrics"
)

// writePackageAnalysis generates comprehensive package analysis output
func (cr *ConsoleReporter) writePackageAnalysis(output io.Writer, report *metrics.Report) {
	fmt.Fprintln(output, "=== PACKAGE ANALYSIS ===")

	packages := report.Packages
	if len(packages) == 0 {
		fmt.Fprintln(output, "No packages found.")
		fmt.Fprintln(output)
		return
	}

	// Sort packages by name for consistent output
	cr.sortPackagesByName(packages)

	// Write summary statistics
	cr.writePackageSummaryStats(output, packages)

	// Write quality issue analysis
	cr.writePackageQualityIssues(output, packages)

	// Write largest packages ranking
	cr.writeLargestPackages(output, packages)

	// Write detailed dependencies (if verbose)
	cr.writePackageDependencies(output, packages)
}

// sortPackagesByName sorts packages alphabetically by name
func (cr *ConsoleReporter) sortPackagesByName(packages []metrics.PackageMetrics) {
	sort.Slice(packages, func(i, j int) bool {
		return packages[i].Name < packages[j].Name
	})
}

// writePackageSummaryStats calculates and writes package summary statistics
func (cr *ConsoleReporter) writePackageSummaryStats(output io.Writer, packages []metrics.PackageMetrics) {
	totalDeps, totalFiles := cr.calculatePackageTotals(packages)
	avgDepsPerPkg := float64(totalDeps) / float64(len(packages))
	avgFilesPerPkg := float64(totalFiles) / float64(len(packages))

	fmt.Fprintf(output, "Total Packages: %d\n", len(packages))
	fmt.Fprintf(output, "Average Dependencies per Package: %.1f\n", avgDepsPerPkg)
	fmt.Fprintf(output, "Average Files per Package: %.1f\n", avgFilesPerPkg)
	fmt.Fprintln(output)
}

// calculatePackageTotals computes total dependencies and files across all packages
func (cr *ConsoleReporter) calculatePackageTotals(packages []metrics.PackageMetrics) (int, int) {
	totalDeps := 0
	totalFiles := 0
	for _, pkg := range packages {
		totalDeps += len(pkg.Dependencies)
		totalFiles += len(pkg.Files)
	}
	return totalDeps, totalFiles
}

// writePackageQualityIssues identifies and reports high coupling and low cohesion packages
func (cr *ConsoleReporter) writePackageQualityIssues(output io.Writer, packages []metrics.PackageMetrics) {
	cr.writeHighCouplingPackages(output, packages)
	cr.writeLowCohesionPackages(output, packages)
}

// writeHighCouplingPackages reports packages with excessive dependencies (>3)
func (cr *ConsoleReporter) writeHighCouplingPackages(output io.Writer, packages []metrics.PackageMetrics) {
	var highCouplingPkgs []metrics.PackageMetrics
	for _, pkg := range packages {
		if len(pkg.Dependencies) > 3 {
			highCouplingPkgs = append(highCouplingPkgs, pkg)
		}
	}

	if len(highCouplingPkgs) > 0 {
		fmt.Fprintln(output, "High Coupling Packages (>3 dependencies):")
		for _, pkg := range highCouplingPkgs {
			fmt.Fprintf(output, "  %s: %d dependencies (coupling: %.1f)\n",
				pkg.Name, len(pkg.Dependencies), pkg.CouplingScore)
		}
		fmt.Fprintln(output)
	}
}

// writeLowCohesionPackages reports packages with poor internal cohesion (<2.0)
func (cr *ConsoleReporter) writeLowCohesionPackages(output io.Writer, packages []metrics.PackageMetrics) {
	var lowCohesionPkgs []metrics.PackageMetrics
	for _, pkg := range packages {
		if pkg.CohesionScore < 2.0 {
			lowCohesionPkgs = append(lowCohesionPkgs, pkg)
		}
	}

	if len(lowCohesionPkgs) > 0 {
		fmt.Fprintln(output, "Low Cohesion Packages (<2.0 cohesion score):")
		for _, pkg := range lowCohesionPkgs {
			fmt.Fprintf(output, "  %s: %.1f cohesion, %d files, %d functions\n",
				pkg.Name, pkg.CohesionScore, len(pkg.Files), pkg.Functions)
		}
		fmt.Fprintln(output)
	}
}

// writeLargestPackages reports the largest packages ranked by function count
func (cr *ConsoleReporter) writeLargestPackages(output io.Writer, packages []metrics.PackageMetrics) {
	// Sort by function count descending
	sort.Slice(packages, func(i, j int) bool {
		return packages[i].Functions > packages[j].Functions
	})

	limit := len(packages)
	if limit > 10 {
		limit = 10
	}

	fmt.Fprintln(output, "Largest Packages (by function count):")
	for i := 0; i < limit; i++ {
		pkg := packages[i]
		fmt.Fprintf(output, "  %s: %d functions, %d structs, %d interfaces, %d files\n",
			pkg.Name, pkg.Functions, pkg.Structs, pkg.Interfaces, len(pkg.Files))
	}
	fmt.Fprintln(output)
}

// writePackageDependencies writes detailed dependency information in verbose mode
func (cr *ConsoleReporter) writePackageDependencies(output io.Writer, packages []metrics.PackageMetrics) {
	if !cr.config.Verbose || len(packages) > 5 {
		return
	}

	fmt.Fprintln(output, "Package Dependencies:")
	for _, pkg := range packages {
		fmt.Fprintf(output, "  %s:\n", pkg.Name)
		if len(pkg.Dependencies) == 0 {
			fmt.Fprintln(output, "    (no internal dependencies)")
		} else {
			for _, dep := range pkg.Dependencies {
				fmt.Fprintf(output, "    → %s\n", dep)
			}
		}
	}
	fmt.Fprintln(output)
}

// writeCircularDependencies displays circular dependency detection results
func (cr *ConsoleReporter) writeCircularDependencies(output io.Writer, report *metrics.Report) {
	fmt.Fprintln(output, "=== CIRCULAR DEPENDENCIES ===")
	if cr.writeCircularDepsEmpty(output, report) {
		return
	}
	fmt.Fprintf(output, "Found %d circular dependency chain(s):\n\n", len(report.CircularDependencies))
	cr.writeCircularDepsList(output, report.CircularDependencies)
}

func (cr *ConsoleReporter) writeCircularDepsEmpty(output io.Writer, report *metrics.Report) bool {
	if len(report.CircularDependencies) == 0 {
		fmt.Fprintln(output, "No circular dependencies detected.")
		fmt.Fprintln(output)
		return true
	}
	return false
}

func (cr *ConsoleReporter) writeCircularDepsList(output io.Writer, cycles []metrics.CircularDependency) {
	for i, cycle := range cycles {
		cr.writeCircularDepsEntry(output, i+1, cycle)
	}
}

func (cr *ConsoleReporter) writeCircularDepsEntry(output io.Writer, index int, cycle metrics.CircularDependency) {
	severity := cycle.Severity
	if severity == "" {
		severity = metrics.SeverityLevelInfo
	}
	fmt.Fprintf(output, "%d. [%s SEVERITY] ", index, toUpperCase(string(severity)))
	cr.writeCircularDepsChain(output, cycle.Packages)
	fmt.Fprintln(output)
}

func (cr *ConsoleReporter) writeCircularDepsChain(output io.Writer, packages []string) {
	for j, pkg := range packages {
		if j > 0 {
			fmt.Fprint(output, " → ")
		}
		fmt.Fprint(output, pkg)
	}
	if len(packages) > 0 {
		fmt.Fprintf(output, " → %s\n", packages[0])
	}
}

// toUpperCase converts a string to uppercase
func toUpperCase(s string) string {
	return strings.ToUpper(s)
}
