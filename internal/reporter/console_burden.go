package reporter

import (
	"fmt"
	"io"
	"sort"

	"github.com/opd-ai/go-stats-generator/internal/metrics"
)

// writeBurdenAnalysis generates maintenance burden analysis output
func (cr *ConsoleReporter) writeBurdenAnalysis(output io.Writer, report *metrics.Report) {
	fmt.Fprintln(output, "=== MAINTENANCE BURDEN ===")

	burden := report.Burden

	// Summary metrics
	fmt.Fprintf(output, "Magic Numbers: %d\n", len(burden.MagicNumbers))
	fmt.Fprintf(output, "Dead Code (Unreferenced): %d functions\n", len(burden.DeadCode.UnreferencedFunctions))
	fmt.Fprintf(output, "Dead Code (Unreachable): %d blocks\n", len(burden.DeadCode.UnreachableCode))
	fmt.Fprintf(output, "Dead Code Percentage: %.2f%%\n", burden.DeadCode.DeadCodePercent)
	fmt.Fprintf(output, "Complex Signatures: %d\n", len(burden.ComplexSignatures))
	fmt.Fprintf(output, "Deeply Nested Functions: %d\n", len(burden.DeeplyNestedFunctions))
	fmt.Fprintf(output, "Feature Envy Methods: %d\n", len(burden.FeatureEnvyMethods))
	fmt.Fprintln(output)

	cr.writeTopBurdenIssues(output, burden)

	fmt.Fprintln(output)
}

// writeTopBurdenIssues displays top burden violations
func (cr *ConsoleReporter) writeTopBurdenIssues(output io.Writer, burden metrics.BurdenMetrics) {
	cr.writeTopComplexSignatures(output, burden.ComplexSignatures)
	cr.writeTopDeeplyNestedFunctions(output, burden.DeeplyNestedFunctions)
	cr.writeTopMagicNumbers(output, burden.MagicNumbers)
}

// writeTopComplexSignatures displays functions with complex signatures
func (cr *ConsoleReporter) writeTopComplexSignatures(output io.Writer, signatures []metrics.SignatureIssue) {
	if len(signatures) == 0 {
		return
	}

	limit := cr.calculateDisplayLimit(len(signatures))
	fmt.Fprintf(output, "Top %d Complex Signatures:\n", limit)
	fmt.Fprintf(output, "%-40s %-20s %8s %8s\n", "Function", "File", "Params", "Returns")
	fmt.Fprintln(output, "--------------------------------------------------------------------------------")

	for i := 0; i < limit; i++ {
		s := signatures[i]
		fmt.Fprintf(output, "%-40s %-20s %8d %8d\n",
			cr.truncate(s.Function, 40),
			cr.truncate(s.File, 20),
			s.ParameterCount,
			s.ReturnCount,
		)
	}
	fmt.Fprintln(output)
}

// writeTopDeeplyNestedFunctions displays functions with deep nesting
func (cr *ConsoleReporter) writeTopDeeplyNestedFunctions(output io.Writer, nesting []metrics.NestingIssue) {
	if len(nesting) == 0 {
		return
	}

	limit := cr.calculateDisplayLimit(len(nesting))
	fmt.Fprintf(output, "Top %d Deeply Nested Functions:\n", limit)
	fmt.Fprintf(output, "%-40s %-20s %8s\n", "Function", "File", "Max Depth")
	fmt.Fprintln(output, "--------------------------------------------------------------------------------")

	for i := 0; i < limit; i++ {
		n := nesting[i]
		fmt.Fprintf(output, "%-40s %-20s %8d\n",
			cr.truncate(n.Function, 40),
			cr.truncate(n.File, 20),
			n.MaxDepth,
		)
	}
	fmt.Fprintln(output)
}

// writeTopMagicNumbers displays top magic number occurrences
func (cr *ConsoleReporter) writeTopMagicNumbers(output io.Writer, numbers []metrics.MagicNumber) {
	if len(numbers) == 0 {
		return
	}

	limit := cr.calculateDisplayLimit(len(numbers))
	fmt.Fprintf(output, "Top %d Magic Numbers:\n", limit)
	fmt.Fprintf(output, "%-20s %-30s %8s %s\n", "Value", "File", "Line", "Context")
	fmt.Fprintln(output, "--------------------------------------------------------------------------------")

	for i := 0; i < limit; i++ {
		m := numbers[i]
		fmt.Fprintf(output, "%-20s %-30s %8d %s\n",
			cr.truncate(m.Value, 20),
			cr.truncate(m.File, 30),
			m.Line,
			cr.truncate(m.Context, 30),
		)
	}
	fmt.Fprintln(output)
}

// writeOrganizationAnalysis generates organization health analysis output
func (cr *ConsoleReporter) writeOrganizationAnalysis(output io.Writer, report *metrics.Report) {
	org := report.Organization
	content := sectionContent{
		header: "=== ORGANIZATION HEALTH ===",
		summaryLines: []string{
			fmt.Sprintf("Oversized Files: %d", len(org.OversizedFiles)),
			fmt.Sprintf("Oversized Packages: %d", len(org.OversizedPackages)),
			fmt.Sprintf("Deep Directories: %d", len(org.DeepDirectories)),
			fmt.Sprintf("High Fan-In Packages: %d", len(org.HighFanInPackages)),
			fmt.Sprintf("High Fan-Out Packages: %d", len(org.HighFanOutPackages)),
			fmt.Sprintf("Avg Package Instability: %.2f", org.AvgPackageStability),
		},
		detailWriters: []func(){
			func() { cr.writeOversizedFiles(output, org.OversizedFiles) },
			func() { cr.writeOversizedPackages(output, org.OversizedPackages) },
			func() { cr.writeDeepDirectories(output, org.DeepDirectories) },
			func() { cr.writeHighFanInPackages(output, org.HighFanInPackages) },
			func() { cr.writeHighFanOutPackages(output, org.HighFanOutPackages) },
		},
	}
	cr.writeSectionWithDetails(output, content)
}

// writeOversizedFiles displays files exceeding size thresholds
func (cr *ConsoleReporter) writeOversizedFiles(output io.Writer, files []metrics.OversizedFile) {
	if len(files) == 0 {
		return
	}

	sorted := make([]metrics.OversizedFile, len(files))
	copy(sorted, files)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].MaintenanceBurden > sorted[j].MaintenanceBurden
	})

	limit := cr.calculateDisplayLimit(len(sorted))

	fmt.Fprintf(output, "Top %d Oversized Files:\n", limit)
	fmt.Fprintf(output, "%-50s %8s %8s %8s %s\n", "File", "Lines", "Funcs", "Types", "Burden")
	fmt.Fprintln(output, "--------------------------------------------------------------------------------")

	for i := 0; i < limit; i++ {
		f := sorted[i]
		fmt.Fprintf(output, "%-50s %8d %8d %8d %.2f\n",
			cr.truncate(f.File, 50),
			f.Lines.Code,
			f.FunctionCount,
			f.TypeCount,
			f.MaintenanceBurden,
		)
	}
	fmt.Fprintln(output)
}

// writeOversizedPackages displays packages exceeding size thresholds
// writeOversizedPackages displays packages exceeding size thresholds
func (cr *ConsoleReporter) writeOversizedPackages(output io.Writer, pkgs []metrics.OversizedPackage) {
	if len(pkgs) == 0 {
		return
	}

	sorted := make([]metrics.OversizedPackage, len(pkgs))
	copy(sorted, pkgs)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].TotalFunctions > sorted[j].TotalFunctions
	})

	limit := cr.calculateDisplayLimit(len(sorted))
	fmt.Fprintf(output, "Top %d Oversized Packages:\n", limit)
	fmt.Fprintf(output, "%-30s %8s %8s %8s %s\n", "Package", "Files", "Exports", "Funcs", "Mega?")
	fmt.Fprintln(output, "--------------------------------------------------------------------------------")

	for i := 0; i < limit; i++ {
		p := sorted[i]
		mega := "No"
		if p.IsMegaPackage {
			mega = "Yes"
		}
		fmt.Fprintf(output, "%-30s %8d %8d %8d %s\n",
			cr.truncate(p.Package, 30),
			p.FileCount,
			p.ExportedSymbols,
			p.TotalFunctions,
			mega,
		)
	}
	fmt.Fprintln(output)
}

// writeDeepDirectories displays directory structures exceeding depth thresholds
func (cr *ConsoleReporter) writeDeepDirectories(output io.Writer, dirs []metrics.DeepDirectory) {
	if len(dirs) == 0 {
		return
	}

	sorted := make([]metrics.DeepDirectory, len(dirs))
	copy(sorted, dirs)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Depth > sorted[j].Depth
	})

	limit := cr.calculateDisplayLimit(len(sorted))

	fmt.Fprintf(output, "Top %d Deep Directories:\n", limit)
	fmt.Fprintf(output, "%-60s %8s %8s\n", "Path", "Depth", "Files")
	fmt.Fprintln(output, "--------------------------------------------------------------------------------")

	for i := 0; i < limit; i++ {
		d := sorted[i]
		fmt.Fprintf(output, "%-60s %8d %8d\n",
			cr.truncate(d.Path, 60),
			d.Depth,
			d.FileCount,
		)
	}
	fmt.Fprintln(output)
}

// writeHighFanInPackages displays packages with high incoming dependencies
func (cr *ConsoleReporter) writeHighFanInPackages(output io.Writer, pkgs []metrics.FanInPackage) {
	if len(pkgs) == 0 {
		return
	}

	sorted := make([]metrics.FanInPackage, len(pkgs))
	copy(sorted, pkgs)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].FanIn > sorted[j].FanIn
	})

	limit := cr.calculateDisplayLimit(len(sorted))

	fmt.Fprintf(output, "Top %d High Fan-In Packages (Bottlenecks):\n", limit)
	fmt.Fprintf(output, "%-40s %8s %s\n", "Package", "Fan-In", "Risk Level")
	fmt.Fprintln(output, "--------------------------------------------------------------------------------")

	for i := 0; i < limit; i++ {
		p := sorted[i]
		fmt.Fprintf(output, "%-40s %8d %s\n",
			cr.truncate(p.Package, 40),
			p.FanIn,
			p.RiskLevel,
		)
	}
	fmt.Fprintln(output)
}

// writeHighFanOutPackages displays packages with high outgoing dependencies
func (cr *ConsoleReporter) writeHighFanOutPackages(output io.Writer, pkgs []metrics.FanOutPackage) {
	if len(pkgs) == 0 {
		return
	}

	sorted := make([]metrics.FanOutPackage, len(pkgs))
	copy(sorted, pkgs)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].FanOut > sorted[j].FanOut
	})

	limit := cr.calculateDisplayLimit(len(sorted))

	fmt.Fprintf(output, "Top %d High Fan-Out Packages (Authority):\n", limit)
	fmt.Fprintf(output, "%-40s %8s %12s %s\n", "Package", "Fan-Out", "Instability", "Risk")
	fmt.Fprintln(output, "--------------------------------------------------------------------------------")

	for i := 0; i < limit; i++ {
		p := sorted[i]
		fmt.Fprintf(output, "%-40s %8d %12.2f %s\n",
			cr.truncate(p.Package, 40),
			p.FanOut,
			p.Instability,
			p.CouplingRisk,
		)
	}
	fmt.Fprintln(output)
}
