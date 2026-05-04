package cmd

import (
	"fmt"
	"go/ast"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/opd-ai/go-stats-generator/internal/analyzer"
	"github.com/opd-ai/go-stats-generator/internal/config"
	"github.com/opd-ai/go-stats-generator/internal/metrics"
)

// finalizeReport populates the report with collected metrics and generates final package report
func finalizeReport(report *metrics.Report, collectedMetrics *CollectedMetrics, packageAnalyzer *analyzer.PackageAnalyzer, cfg *config.Config) {
	// Generate package report
	packageReport, err := packageAnalyzer.GenerateReport()
	if err != nil {
		if cfg.Output.Verbose {
			fmt.Fprintf(os.Stderr, "Warning: failed to generate package report: %v\n", err)
		}
		packageReport = &metrics.PackageReport{
			Packages:             []metrics.PackageMetrics{},
			CircularDependencies: []metrics.CircularDependency{},
			TotalPackages:        0,
		}
	}

	// Populate main metrics
	report.Functions = collectedMetrics.Functions
	report.Structs = collectedMetrics.Structs
	report.Interfaces = collectedMetrics.Interfaces
	report.Packages = packageReport.Packages
	report.CircularDependencies = packageReport.CircularDependencies

	// Aggregate generics metrics from all files
	aggregateGenericsMetrics(report, collectedMetrics)

	// Calculate overview metrics
	calculateOverviewMetrics(report, collectedMetrics, packageReport)

	// Finalize complexity metrics aggregation
	finalizeComplexityMetrics(report)

	// Finalize concurrency metrics summary statistics
	finalizeConcurrencyMetrics(report)

	// Finalize burden metrics (dead code percentage)
	finalizeBurdenMetrics(report)

	// Calculate MBI scores for all files and packages
	finalizeScoringMetrics(report, cfg)

	// Analyze test coverage correlation if coverage profile provided
	finalizeTestCoverageMetrics(report, cfg)
}

// finalizeScoringMetrics calculates maintenance burden index for files and packages
func finalizeScoringMetrics(report *metrics.Report, cfg *config.Config) {
	scoringAnalyzer := analyzer.NewScoringAnalyzer(cfg.Analysis.Scoring.Weights)
	report.Scores = *scoringAnalyzer.CalculateAllScores(report)
}

// finalizeDuplicationMetrics performs duplication analysis using pre-extracted blocks.
// Blocks are accumulated in CollectedMetrics during the streaming phase (processFileAnalysis)
// using per-file token.FileSets, so this function does not need a file map or shared fset.
func finalizeDuplicationMetrics(report *metrics.Report, duplicationAnalyzer *analyzer.DuplicationAnalyzer, collectedMetrics *CollectedMetrics, cfg *config.Config) {
	blocks := collectedMetrics.DupBlocks
	totalLines := collectedMetrics.DupTotalLines

	if cfg.Analysis.Duplication.IgnoreTestFiles {
		blocks, totalLines = filterTestBlocks(blocks, collectedMetrics)
	}

	if len(blocks) == 0 {
		report.Duplication = createEmptyDuplicationMetrics()
		return
	}

	logDuplicationStart(cfg, len(collectedMetrics.Files))
	duplicationMetrics := duplicationAnalyzer.AnalyzeDuplicationFromBlocks(blocks, totalLines, cfg.Analysis.Duplication.SimilarityThreshold)
	report.Duplication = duplicationMetrics
	logDuplicationResults(cfg, duplicationMetrics)
}

// filterTestBlocks removes blocks belonging to test files and returns the adjusted total line count.
func filterTestBlocks(blocks []analyzer.StatementBlock, collectedMetrics *CollectedMetrics) ([]analyzer.StatementBlock, int) {
	var filtered []analyzer.StatementBlock
	totalLines := 0
	for _, block := range blocks {
		if !strings.HasSuffix(block.File, "_test.go") {
			filtered = append(filtered, block)
		}
	}
	for filePath, lines := range collectedMetrics.FileLinesCount {
		if !strings.HasSuffix(filePath, "_test.go") {
			totalLines += lines
		}
	}
	return filtered, totalLines
}

// createEmptyDuplicationMetrics returns zero-initialized duplication metrics.
func createEmptyDuplicationMetrics() metrics.DuplicationMetrics {
	return metrics.DuplicationMetrics{
		ClonePairs:       0,
		DuplicatedLines:  0,
		DuplicationRatio: 0.0,
		LargestCloneSize: 0,
		Clones:           []metrics.ClonePair{},
	}
}

// logDuplicationStart prints duplication analysis progress if verbose mode is enabled.
func logDuplicationStart(cfg *config.Config, fileCount int) {
	if !cfg.Output.Verbose {
		return
	}
	msg := fmt.Sprintf("Running duplication analysis on %d files", fileCount)
	if cfg.Analysis.Duplication.IgnoreTestFiles {
		msg += " (excluding test files)"
	}
	fmt.Fprintf(os.Stderr, "%s...\n", msg)
}

// logDuplicationResults prints duplication analysis results if verbose mode is enabled.
func logDuplicationResults(cfg *config.Config, metrics metrics.DuplicationMetrics) {
	if !cfg.Output.Verbose {
		return
	}
	fmt.Fprintf(os.Stderr, "Found %d clone pairs, %d duplicated lines (%.2f%% duplication ratio)\n",
		metrics.ClonePairs,
		metrics.DuplicatedLines,
		metrics.DuplicationRatio*100)
}

// finalizeNamingMetrics performs naming convention analysis on all collected files.
// Identifier violations were accumulated in CollectedMetrics during the streaming phase
// (see processFileAnalysis) using per-file token.FileSets, so this function only runs
// file-name and package-name checks which do not require position lookups.
func finalizeNamingMetrics(report *metrics.Report, analyzers *AnalyzerSet, collectedMetrics *CollectedMetrics, cfg *config.Config) {
	filePaths := extractFilePaths(collectedMetrics)
	if len(filePaths) == 0 {
		report.Naming = createEmptyNamingMetrics()
		return
	}

	logNamingStart(cfg, len(filePaths))

	fileNameViolations := analyzers.Naming.AnalyzeFileNames(filePaths)
	// Use pre-accumulated identifier violations from the streaming phase instead of
	// iterating the Files map with the shared (and now empty) fileSet.
	identifierViolations := collectedMetrics.IdentifierViolations
	totalIdentifiers := collectedMetrics.TotalIdentifiers
	packageNameViolations, uniquePackages := analyzeAllPackageNames(analyzers, collectedMetrics)

	report.Naming = buildNamingMetrics(
		analyzers, fileNameViolations, identifierViolations, packageNameViolations,
		len(filePaths), totalIdentifiers, len(uniquePackages))

	logNamingResults(cfg, report.Naming)
}

// extractFilePaths extracts sorted file paths from collected metrics for analysis.
func extractFilePaths(collectedMetrics *CollectedMetrics) []string {
	var filePaths []string
	for filePath := range collectedMetrics.Files {
		filePaths = append(filePaths, filePath)
	}
	return filePaths
}

// createEmptyNamingMetrics returns zero-initialized naming metrics with empty violation lists.
func createEmptyNamingMetrics() metrics.NamingMetrics {
	return metrics.NamingMetrics{
		FileNameViolations:    0,
		IdentifierViolations:  0,
		PackageNameViolations: 0,
		OverallNamingScore:    1.0,
		FileNameIssues:        []metrics.FileNameViolation{},
		IdentifierIssues:      []metrics.IdentifierViolation{},
		PackageNameIssues:     []metrics.PackageNameViolation{},
	}
}

// logNamingStart prints naming analysis progress if verbose mode is enabled.
func logNamingStart(cfg *config.Config, fileCount int) {
	if cfg.Output.Verbose {
		fmt.Fprintf(os.Stderr, "Running naming convention analysis on %d files...\n", fileCount)
	}
}

// analyzeAllPackageNames extracts unique packages and analyzes package naming conventions.
func analyzeAllPackageNames(analyzers *AnalyzerSet, collectedMetrics *CollectedMetrics) ([]metrics.PackageNameViolation, map[string]struct {
	dirName  string
	filePath string
},
) {
	uniquePackages := collectUniquePackages(collectedMetrics)
	var packageNameViolations []metrics.PackageNameViolation
	for pkgName, info := range uniquePackages {
		violations := analyzers.Naming.AnalyzePackageName(pkgName, info.dirName, info.filePath)
		packageNameViolations = append(packageNameViolations, violations...)
	}
	return packageNameViolations, uniquePackages
}

// collectUniquePackages builds a map of unique package names with their directory and file locations.
func collectUniquePackages(collectedMetrics *CollectedMetrics) map[string]struct {
	dirName  string
	filePath string
} {
	uniquePackages := make(map[string]struct {
		dirName  string
		filePath string
	})
	for filePath, astFile := range collectedMetrics.Files {
		if astFile.Name != nil {
			pkgName := astFile.Name.Name
			dirName := filepath.Base(filepath.Dir(filePath))
			if _, exists := uniquePackages[pkgName]; !exists {
				uniquePackages[pkgName] = struct {
					dirName  string
					filePath string
				}{dirName, filePath}
			}
		}
	}
	return uniquePackages
}

// buildNamingMetrics combines all naming analysis results into a comprehensive metrics structure.
func buildNamingMetrics(analyzers *AnalyzerSet, fileNameViolations []metrics.FileNameViolation,
	identifierViolations []metrics.IdentifierViolation, packageNameViolations []metrics.PackageNameViolation,
	fileCount, totalIdentifiers, packageCount int,
) metrics.NamingMetrics {
	fileNamingScore := analyzers.Naming.ComputeFileNamingScore(fileNameViolations, fileCount)
	identifierScore := analyzers.Naming.ComputeIdentifierQualityScore(identifierViolations, totalIdentifiers)
	packageScore := analyzers.Naming.ComputePackageNamingScore(packageNameViolations, packageCount)
	overallScore := (fileNamingScore + identifierScore + packageScore) / 3.0

	return metrics.NamingMetrics{
		FileNameViolations:    len(fileNameViolations),
		IdentifierViolations:  len(identifierViolations),
		PackageNameViolations: len(packageNameViolations),
		OverallNamingScore:    overallScore,
		FileNameIssues:        fileNameViolations,
		IdentifierIssues:      identifierViolations,
		PackageNameIssues:     packageNameViolations,
	}
}

// logNamingResults prints naming analysis summary if verbose mode is enabled.
func logNamingResults(cfg *config.Config, naming metrics.NamingMetrics) {
	if cfg.Output.Verbose {
		fmt.Fprintf(os.Stderr, "Found %d file, %d identifier, %d package naming violations (score: %.2f)\n",
			naming.FileNameViolations,
			naming.IdentifierViolations,
			naming.PackageNameViolations,
			naming.OverallNamingScore)
	}
}

// finalizePlacementMetrics performs placement and cohesion analysis on all collected files.
// It uses AnalyzeMap (which derives filenames from the map keys) rather than Analyze
// (which calls fset.Position on each file), so that a shared token.FileSet is not needed.
func finalizePlacementMetrics(report *metrics.Report, analyzers *AnalyzerSet, collectedMetrics *CollectedMetrics, cfg *config.Config) {
	if len(collectedMetrics.Files) == 0 {
		report.Placement = metrics.PlacementMetrics{
			MisplacedFunctions: 0,
			MisplacedMethods:   0,
			LowCohesionFiles:   0,
			AvgFileCohesion:    0.0,
			FunctionIssues:     []metrics.MisplacedFunctionIssue{},
			MethodIssues:       []metrics.MisplacedMethodIssue{},
			CohesionIssues:     []metrics.FileCohesionIssue{},
		}
		return
	}

	if cfg.Output.Verbose {
		fmt.Fprintf(os.Stderr, "Running placement analysis on %d files...\n", len(collectedMetrics.Files))
	}

	placementMetrics := analyzers.Placement.AnalyzeMap(collectedMetrics.Files)
	report.Placement = placementMetrics

	if cfg.Output.Verbose {
		fmt.Fprintf(os.Stderr, "Found %d misplaced functions, %d misplaced methods, %d low cohesion files (avg cohesion: %.2f)\n",
			placementMetrics.MisplacedFunctions,
			placementMetrics.MisplacedMethods,
			placementMetrics.LowCohesionFiles,
			placementMetrics.AvgFileCohesion)
	}
}

// countIdentifiers counts total identifiers in an AST for scoring
func countIdentifiers(file *ast.File) int {
	count := 0
	ast.Inspect(file, func(n ast.Node) bool {
		switch n.(type) {
		case *ast.FuncDecl, *ast.TypeSpec, *ast.ValueSpec:
			count++
		}
		return true
	})
	return count
}

// finalizeDocumentationMetrics performs documentation analysis on all collected files
func finalizeDocumentationMetrics(report *metrics.Report, analyzers *AnalyzerSet, collectedMetrics *CollectedMetrics, cfg *config.Config) {
	// Skip if documentation analysis is disabled or no files
	if !cfg.Analysis.IncludeDocumentation || len(collectedMetrics.DocFiles) == 0 {
		report.Documentation = metrics.DocumentationMetrics{}
		return
	}

	// Build packages map from DocFiles (mirrors what prepareDocumentationInput did from Files map).
	pkgs := buildPkgsFromDocFiles(collectedMetrics.DocFiles)

	if cfg.Output.Verbose {
		fmt.Fprintf(os.Stderr, "Running documentation analysis on %d files in %d packages...\n", len(collectedMetrics.DocFiles), len(pkgs))
	}

	// Use AnalyzeWithFileSets so that annotation line numbers are resolved against each
	// file's own FileSet rather than the shared discoverer FileSet.
	docMetrics := analyzers.Documentation.AnalyzeWithFileSets(collectedMetrics.DocFiles, pkgs)
	report.Documentation = *docMetrics

	if cfg.Output.Verbose {
		fmt.Fprintf(os.Stderr, "Documentation coverage: %.1f%% (%.1f%% packages, %.1f%% functions, %.1f%% types)\n",
			docMetrics.Coverage.Overall,
			docMetrics.Coverage.Packages,
			docMetrics.Coverage.Functions,
			docMetrics.Coverage.Types)
	}
}

// buildPkgsFromDocFiles builds an ast.Package map keyed by package name from DocFileInfo entries.
func buildPkgsFromDocFiles(docFiles []analyzer.DocFileInfo) map[string]*ast.Package {
	pkgs := make(map[string]*ast.Package)
	for _, fi := range docFiles {
		if fi.File == nil || fi.File.Name == nil {
			continue
		}
		pkgName := fi.File.Name.Name
		if _, exists := pkgs[pkgName]; !exists {
			pkgs[pkgName] = &ast.Package{
				Name:  pkgName,
				Files: make(map[string]*ast.File),
			}
		}
		pkgs[pkgName].Files[fi.Path] = fi.File
	}
	return pkgs
}

// finalizeOrganizationMetrics performs organization analysis on all collected files and packages
func finalizeOrganizationMetrics(report *metrics.Report, analyzers *AnalyzerSet, collectedMetrics *CollectedMetrics, cfg *config.Config, targetPath string) {
	if len(collectedMetrics.Files) == 0 {
		report.Organization = metrics.OrganizationMetrics{}
		return
	}

	orgConfig := getOrganizationConfig(cfg)
	logOrganizationStart(cfg, len(collectedMetrics.Files))

	oversizedFiles := analyzeOversizedFiles(analyzers, collectedMetrics, orgConfig)
	oversizedPackages := analyzeOversizedPackages(analyzers, collectedMetrics, report, orgConfig)
	deepDirs := analyzeDeepDirectories(analyzers, collectedMetrics, targetPath, orgConfig)
	highFanIn, highFanOut, avgStability := analyzeImportGraph(analyzers, collectedMetrics, orgConfig)

	report.Organization = metrics.OrganizationMetrics{
		OversizedFiles:      oversizedFiles,
		OversizedPackages:   oversizedPackages,
		DeepDirectories:     deepDirs,
		HighFanInPackages:   highFanIn,
		HighFanOutPackages:  highFanOut,
		AvgPackageStability: avgStability,
	}

	logOrganizationResults(cfg, len(oversizedFiles), len(oversizedPackages), len(deepDirs))
}

// logOrganizationStart prints verbose logging for organization analysis start
func logOrganizationStart(cfg *config.Config, fileCount int) {
	if cfg.Output.Verbose {
		fmt.Fprintf(os.Stderr, "Running organization analysis on %d files...\n", fileCount)
	}
}

// analyzeOversizedFiles analyzes all files for size violations using pre-computed line counts.
func analyzeOversizedFiles(analyzers *AnalyzerSet, collectedMetrics *CollectedMetrics, orgConfig analyzer.OrganizationConfig) []metrics.OversizedFile {
	var oversizedFiles []metrics.OversizedFile
	for filePath, astFile := range collectedMetrics.Files {
		lineCount := collectedMetrics.FileLinesCount[filePath]
		result, err := analyzers.Organization.AnalyzeFileSizesWithLines(astFile, filePath, lineCount, orgConfig)
		if err == nil && result != nil {
			oversizedFiles = append(oversizedFiles, *result)
		}
	}
	return oversizedFiles
}

// analyzeOversizedPackages analyzes all packages for size violations
func analyzeOversizedPackages(analyzers *AnalyzerSet, collectedMetrics *CollectedMetrics, report *metrics.Report, orgConfig analyzer.OrganizationConfig) []metrics.OversizedPackage {
	pkgInfo := buildPackageInfo(collectedMetrics, report)
	return analyzers.Organization.AnalyzePackageSizes(pkgInfo, orgConfig)
}

// analyzeDeepDirectories analyzes directory structure for excessive nesting
func analyzeDeepDirectories(analyzers *AnalyzerSet, collectedMetrics *CollectedMetrics, targetPath string, orgConfig analyzer.OrganizationConfig) []metrics.DeepDirectory {
	filePaths := extractFilePaths(collectedMetrics)
	return analyzers.Organization.AnalyzeDirectoryDepth(filePaths, targetPath, orgConfig)
}

// analyzeImportGraph analyzes import relationships and returns fan-in, fan-out, and stability metrics
func analyzeImportGraph(analyzers *AnalyzerSet, collectedMetrics *CollectedMetrics, orgConfig analyzer.OrganizationConfig) ([]metrics.FanInPackage, []metrics.FanOutPackage, float64) {
	graphData := buildImportGraphData(collectedMetrics)
	importMetrics, _ := analyzers.Organization.AnalyzeImportGraph(graphData, orgConfig)
	return extractOrgImportMetrics(importMetrics)
}

// extractOrgImportMetrics extracts fan-in, fan-out, and stability from import metrics
func extractOrgImportMetrics(importMetrics *analyzer.ImportGraphMetrics) ([]metrics.FanInPackage, []metrics.FanOutPackage, float64) {
	var highFanIn []metrics.FanInPackage
	var highFanOut []metrics.FanOutPackage
	avgStability := 0.0
	if importMetrics != nil {
		highFanIn = importMetrics.HighFanInPackages
		highFanOut = importMetrics.HighFanOutPackages
		avgStability = importMetrics.AvgInstability
	}
	return highFanIn, highFanOut, avgStability
}

// logOrganizationResults prints verbose logging for organization analysis results
func logOrganizationResults(cfg *config.Config, filesCount, packagesCount, dirsCount int) {
	if cfg.Output.Verbose {
		fmt.Fprintf(os.Stderr, "Found %d oversized files, %d oversized packages, %d deep directories\n",
			filesCount, packagesCount, dirsCount)
	}
}

// prepareDocumentationInput converts files map to slice and groups by package
func prepareDocumentationInput(filesMap map[string]*ast.File) ([]*ast.File, map[string]*ast.Package) {
	var files []*ast.File
	pkgs := make(map[string]*ast.Package)

	for filePath, astFile := range filesMap {
		files = append(files, astFile)
		if astFile.Name != nil {
			pkgName := astFile.Name.Name
			if _, exists := pkgs[pkgName]; !exists {
				pkgs[pkgName] = &ast.Package{
					Name:  pkgName,
					Files: make(map[string]*ast.File),
				}
			}
			pkgs[pkgName].Files[filePath] = astFile
		}
	}

	return files, pkgs
}

// buildPackageInfo constructs package metadata from collected metrics for placement analysis.
func buildPackageInfo(collectedMetrics *CollectedMetrics, report *metrics.Report) map[string]*analyzer.PackageInfo {
	pkgInfo := make(map[string]*analyzer.PackageInfo)

	for filePath, astFile := range collectedMetrics.Files {
		if astFile.Name == nil {
			continue
		}
		pkgName := astFile.Name.Name
		if _, exists := pkgInfo[pkgName]; !exists {
			pkgInfo[pkgName] = &analyzer.PackageInfo{
				Name:  pkgName,
				Files: []string{},
			}
		}
		pkgInfo[pkgName].Files = append(pkgInfo[pkgName].Files, filePath)
	}

	for _, pkg := range report.Packages {
		if info, exists := pkgInfo[pkg.Name]; exists {
			info.ExportedSymbols = countExportedSymbols(pkg)
			info.TotalFunctions = pkg.Functions
			info.CohesionScore = pkg.CohesionScore
		}
	}

	return pkgInfo
}

// countExportedSymbols returns the total count of exported symbols in a package.
func countExportedSymbols(pkg metrics.PackageMetrics) int {
	return pkg.Functions + pkg.Structs + pkg.Interfaces
}

// buildImportGraphData constructs import relationship data for placement analysis.
func buildImportGraphData(collectedMetrics *CollectedMetrics) *analyzer.ImportGraphData {
	graphData := &analyzer.ImportGraphData{
		FileImports:    make(map[string]int),
		PackageFanIn:   make(map[string][]string),
		PackageFanOut:  make(map[string][]string),
		FilePackageMap: make(map[string]string),
	}

	for filePath, astFile := range collectedMetrics.Files {
		if astFile.Name != nil {
			graphData.FilePackageMap[filePath] = astFile.Name.Name
		}

		importCount := 0
		for _, imp := range astFile.Imports {
			if imp.Path != nil {
				importCount++
			}
		}
		graphData.FileImports[filePath] = importCount
	}

	return graphData
}

// calculateOverviewMetrics calculates and sets the overview metrics in the report
func calculateOverviewMetrics(report *metrics.Report, collectedMetrics *CollectedMetrics, packageReport *metrics.PackageReport) {
	// Calculate total lines of code from all functions
	totalLOC := 0
	for _, fn := range collectedMetrics.Functions {
		totalLOC += fn.Lines.Code
	}

	report.Overview = metrics.OverviewMetrics{
		TotalLinesOfCode: totalLOC,
		TotalFunctions:   len(collectedMetrics.Functions),
		TotalStructs:     len(collectedMetrics.Structs),
		TotalInterfaces:  len(collectedMetrics.Interfaces),
		TotalPackages:    packageReport.TotalPackages,
		TotalFiles:       report.Metadata.FilesProcessed,
	}

	// Count methods vs functions
	for _, fn := range collectedMetrics.Functions {
		if fn.IsMethod {
			report.Overview.TotalMethods++
		}
	}
	report.Overview.TotalFunctions -= report.Overview.TotalMethods
}

// finalizeConcurrencyMetrics calculates final concurrency metric summaries
func finalizeConcurrencyMetrics(report *metrics.Report) {
	report.Patterns.ConcurrencyPatterns.Goroutines.TotalCount = len(report.Patterns.ConcurrencyPatterns.Goroutines.Instances)
	for _, instance := range report.Patterns.ConcurrencyPatterns.Goroutines.Instances {
		if instance.IsAnonymous {
			report.Patterns.ConcurrencyPatterns.Goroutines.AnonymousCount++
		} else {
			report.Patterns.ConcurrencyPatterns.Goroutines.NamedCount++
		}
	}

	report.Patterns.ConcurrencyPatterns.Channels.TotalCount = len(report.Patterns.ConcurrencyPatterns.Channels.Instances)
	for _, instance := range report.Patterns.ConcurrencyPatterns.Channels.Instances {
		if instance.IsBuffered {
			report.Patterns.ConcurrencyPatterns.Channels.BufferedCount++
		} else {
			report.Patterns.ConcurrencyPatterns.Channels.UnbufferedCount++
		}
		if instance.IsDirectional {
			report.Patterns.ConcurrencyPatterns.Channels.DirectionalCount++
		}
	}
}

// finalizeBurdenMetrics calculates derived burden statistics
func finalizeBurdenMetrics(report *metrics.Report) {
	if report.Overview.TotalLinesOfCode > 0 {
		totalDeadLines := float64(report.Burden.DeadCode.TotalDeadLines)
		totalLines := float64(report.Overview.TotalLinesOfCode)
		report.Burden.DeadCode.DeadCodePercent = (totalDeadLines / totalLines) * 100
	}
}

// finalizeDeadCodeMetrics groups the accumulated BurdenFiles by package name and runs
// package-scope dead-code detection for each package. Results are merged into the report.
// This must be called after the streaming phase so all files of every package are present.
func finalizeDeadCodeMetrics(report *metrics.Report, collectedMetrics *CollectedMetrics, burdenAnalyzer *analyzer.BurdenAnalyzer) {
	// Group files by package name.
	pkgFiles := make(map[string][]analyzer.BurdenFileInfo)
	for _, fi := range collectedMetrics.BurdenFiles {
		pkgName := fi.Pkg
		if pkgName == "" {
			// Pkg should always be populated by processFileAnalysis; if it's empty
			// that indicates a data quality issue upstream.  Fall back to the AST
			// package name and emit a warning so the problem is visible.
			if fi.File != nil && fi.File.Name != nil {
				pkgName = fi.File.Name.Name
				fmt.Fprintf(os.Stderr, "Warning: BurdenFileInfo has empty Pkg field; falling back to AST package name %q\n", pkgName)
			}
			if pkgName == "" {
				continue // cannot determine package; skip this file
			}
		}
		pkgFiles[pkgName] = append(pkgFiles[pkgName], fi)
	}

	// Run dead-code detection at package scope and merge results.
	for _, fileInfos := range pkgFiles {
		deadCode := burdenAnalyzer.DetectDeadCodeForPackage(fileInfos)
		if deadCode == nil {
			continue
		}
		report.Burden.DeadCode.UnreferencedFunctions = append(
			report.Burden.DeadCode.UnreferencedFunctions, deadCode.UnreferencedFunctions...)
		report.Burden.DeadCode.UnreachableCode = append(
			report.Burden.DeadCode.UnreachableCode, deadCode.UnreachableCode...)
		report.Burden.DeadCode.TotalDeadLines += deadCode.TotalDeadLines
	}
}

// complexityEntry holds temporary complexity data for sorting and analysis
type complexityEntry struct {
	name       string
	complexity float64
	itemType   string
	file       string
	line       int
}

// finalizeComplexityMetrics calculates aggregated complexity statistics
func finalizeComplexityMetrics(report *metrics.Report) {
	calculateAverageComplexities(report)
	buildHighestComplexityList(report)
	buildComplexityDistribution(report)
}

// calculateAverageComplexities computes average complexity for functions and structs
func calculateAverageComplexities(report *metrics.Report) {
	var totalFunctionComplexity float64
	for _, fn := range report.Functions {
		totalFunctionComplexity += fn.Complexity.Overall
	}
	if len(report.Functions) > 0 {
		report.Complexity.AverageFunction = totalFunctionComplexity / float64(len(report.Functions))
	}

	var totalStructComplexity float64
	for _, s := range report.Structs {
		totalStructComplexity += s.Complexity.Overall
	}
	if len(report.Structs) > 0 {
		report.Complexity.AverageStruct = totalStructComplexity / float64(len(report.Structs))
	}
}

// buildHighestComplexityList creates a sorted list of top 20 most complex items
func buildHighestComplexityList(report *metrics.Report) {
	allItems := collectComplexityEntries(report)

	sort.Slice(allItems, func(i, j int) bool {
		return allItems[i].complexity > allItems[j].complexity
	})

	topCount := 20
	if len(allItems) < topCount {
		topCount = len(allItems)
	}

	report.Complexity.HighestComplexity = make([]metrics.ComplexityItem, topCount)
	for i := 0; i < topCount; i++ {
		report.Complexity.HighestComplexity[i] = metrics.ComplexityItem{
			Name:       allItems[i].name,
			Type:       allItems[i].itemType,
			File:       allItems[i].file,
			Line:       allItems[i].line,
			Complexity: allItems[i].complexity,
		}
	}
}

// collectComplexityEntries gathers complexity data from functions and structs
func collectComplexityEntries(report *metrics.Report) []complexityEntry {
	var entries []complexityEntry

	for _, fn := range report.Functions {
		entries = append(entries, complexityEntry{
			name:       fn.Name,
			complexity: fn.Complexity.Overall,
			itemType:   "function",
			file:       fn.File,
			line:       fn.Line,
		})
	}

	for _, s := range report.Structs {
		entries = append(entries, complexityEntry{
			name:       s.Name,
			complexity: s.Complexity.Overall,
			itemType:   "struct",
			file:       s.File,
			line:       s.Line,
		})
	}

	return entries
}

// buildComplexityDistribution creates histogram of complexity ranges
func buildComplexityDistribution(report *metrics.Report) {
	allItems := collectComplexityEntries(report)

	report.Complexity.Distribution = make(map[string]int)
	for _, item := range allItems {
		switch {
		case item.complexity <= 5:
			report.Complexity.Distribution["0-5"]++
		case item.complexity <= 10:
			report.Complexity.Distribution["6-10"]++
		case item.complexity <= 15:
			report.Complexity.Distribution["11-15"]++
		case item.complexity <= 20:
			report.Complexity.Distribution["16-20"]++
		default:
			report.Complexity.Distribution["20+"]++
		}
	}
}

// finalizeRefactoringSuggestions generates prioritized refactoring recommendations
// after all metrics have been finalized (duplication, naming, placement, etc.)
func finalizeRefactoringSuggestions(report *metrics.Report, cfg *config.Config) {
	scoringAnalyzer := analyzer.NewScoringAnalyzer(cfg.Analysis.Scoring.Weights)
	suggestionGen := analyzer.NewSuggestionGenerator(scoringAnalyzer)
	rawSuggestions := suggestionGen.GenerateSuggestions(report)

	// Convert to metrics.SuggestionInfo for report serialization
	report.Suggestions = make([]metrics.SuggestionInfo, 0, len(rawSuggestions))
	for _, s := range rawSuggestions {
		report.Suggestions = append(report.Suggestions, metrics.SuggestionInfo{
			Action:        string(s.Action),
			Target:        s.Target,
			Location:      s.Location,
			Description:   s.Description,
			Effort:        string(s.Effort),
			MBIImpact:     s.MBIImpact,
			ImpactEffort:  s.ImpactEffort,
			Category:      s.Category,
			AffectedLines: s.AffectedLines,
		})
	}
}

// aggregateGenericsMetrics merges generic metrics from all analyzed files
func aggregateGenericsMetrics(report *metrics.Report, collected *CollectedMetrics) {
	if len(collected.Generics) == 0 {
		return
	}

	merged := metrics.GenericMetrics{
		TypeParameters: metrics.GenericTypeParameters{
			Constraints: make(map[string]int),
		},
		ConstraintUsage: make(map[string]int),
	}

	for _, gen := range collected.Generics {
		metrics.MergeGenericsData(&merged, gen)
	}

	if len(merged.TypeParameters.Complexity) > 0 {
		total := 0.0
		for _, c := range merged.TypeParameters.Complexity {
			total += c.ComplexityScore
		}
		merged.ComplexityScore = total / float64(len(merged.TypeParameters.Complexity))
	}

	report.Generics = merged
}

// finalizeTestCoverageMetrics analyzes test coverage if provided
func finalizeTestCoverageMetrics(report *metrics.Report, cfg *config.Config) {
	if cfg.Analysis.CoverageProfile == "" {
		return
	}
	loadAndAnalyzeCoverage(report, cfg)
	analyzeTestQualityMetrics(report, cfg)
}

// loadAndAnalyzeCoverage loads coverage profile and analyzes correlation
func loadAndAnalyzeCoverage(report *metrics.Report, cfg *config.Config) {
	coveragePath := resolveCoveragePath(cfg.Analysis.CoverageProfile)
	logVerbose(cfg, "Loading coverage profile from: %s\n", coveragePath)

	covAnalyzer := analyzer.NewTestCoverageAnalyzer()
	if err := covAnalyzer.LoadCoverageProfile(coveragePath); err != nil {
		logVerbose(cfg, "Warning: failed to load coverage profile: %v\n", err)
		return
	}

	logVerbose(cfg, "Analyzing test coverage correlation for %d functions...\n", len(report.Functions))
	report.TestCoverage = covAnalyzer.AnalyzeCorrelation(report.Functions)
	logCoverageResults(report, cfg)
}

// resolveCoveragePath returns absolute path for coverage profile
func resolveCoveragePath(path string) string {
	if filepath.IsAbs(path) {
		return path
	}
	wd, err := os.Getwd()
	if err != nil {
		// If we can't get working directory, return the relative path as-is
		// Caller will handle any path resolution errors
		return path
	}
	return filepath.Join(wd, path)
}

// logCoverageResults outputs coverage analysis summary
func logCoverageResults(report *metrics.Report, cfg *config.Config) {
	logVerbose(cfg, "Coverage analysis: %.2f%% function coverage, %d high-risk functions, %d coverage gaps\n",
		report.TestCoverage.FunctionCoverageRate*100,
		len(report.TestCoverage.HighRiskFunctions),
		len(report.TestCoverage.CoverageGaps))
}

// analyzeTestQualityMetrics performs test quality analysis
func analyzeTestQualityMetrics(report *metrics.Report, cfg *config.Config) {
	repoPath, err := os.Getwd()
	if err != nil {
		logVerbose(cfg, "Warning: failed to get working directory for test quality analysis: %v\n", err)
		return
	}
	testQuality, err := analyzer.AnalyzeTestQuality(repoPath)
	if err != nil {
		logVerbose(cfg, "Warning: failed to analyze test quality: %v\n", err)
		return
	}
	report.TestQuality = testQuality
	logVerbose(cfg, "Test quality analysis: %d tests in %d files\n",
		testQuality.TotalTests, len(testQuality.TestFiles))
}

// logVerbose prints message if verbose mode is enabled
func logVerbose(cfg *config.Config, format string, args ...interface{}) {
	if cfg.Output.Verbose {
		fmt.Fprintf(os.Stderr, format, args...)
	}
}

// finalizeTeamMetrics analyzes Git history for team productivity
func finalizeTeamMetrics(report *metrics.Report, targetPath string, cfg *config.Config) {
	// Skip if feature disabled
	if !cfg.Analysis.EnableTeamMetrics {
		return
	}

	teamAnalyzer := analyzer.NewTeamAnalyzer(targetPath)
	teamMetrics, err := teamAnalyzer.AnalyzeTeamMetrics()
	if err != nil {
		if cfg.Output.Verbose {
			fmt.Fprintf(os.Stderr, "Warning: team metrics unavailable (not a Git repo?): %v\n", err)
		}
		return
	}

	report.Team = teamMetrics
	if cfg.Output.Verbose {
		fmt.Fprintf(os.Stderr, "Team metrics: %d developers analyzed\n",
			teamMetrics.TotalDevelopers)
	}
}
