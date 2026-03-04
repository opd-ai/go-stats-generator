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

// finalizeDuplicationMetrics performs duplication analysis on all collected files
func finalizeDuplicationMetrics(report *metrics.Report, duplicationAnalyzer *analyzer.DuplicationAnalyzer, collectedMetrics *CollectedMetrics, cfg *config.Config) {
	// Get configuration values for duplication detection
	minBlockLines := cfg.Analysis.Duplication.MinBlockLines
	similarityThreshold := cfg.Analysis.Duplication.SimilarityThreshold
	ignoreTestFiles := cfg.Analysis.Duplication.IgnoreTestFiles

	// Filter files if ignoring test files
	filesToAnalyze := collectedMetrics.Files
	if ignoreTestFiles {
		filtered := make(map[string]*ast.File)
		for filename, file := range collectedMetrics.Files {
			if !strings.HasSuffix(filename, "_test.go") {
				filtered[filename] = file
			}
		}
		filesToAnalyze = filtered
	}

	// Skip if no files were collected
	if len(filesToAnalyze) == 0 {
		report.Duplication = metrics.DuplicationMetrics{
			ClonePairs:       0,
			DuplicatedLines:  0,
			DuplicationRatio: 0.0,
			LargestCloneSize: 0,
			Clones:           []metrics.ClonePair{},
		}
		return
	}

	if cfg.Output.Verbose {
		if ignoreTestFiles {
			fmt.Fprintf(os.Stderr, "Running duplication analysis on %d files (excluding test files)...\n", len(filesToAnalyze))
		} else {
			fmt.Fprintf(os.Stderr, "Running duplication analysis on %d files...\n", len(filesToAnalyze))
		}
	}

	// Run duplication analysis
	duplicationMetrics := duplicationAnalyzer.AnalyzeDuplication(
		filesToAnalyze,
		minBlockLines,
		similarityThreshold,
	)

	report.Duplication = duplicationMetrics

	if cfg.Output.Verbose {
		fmt.Fprintf(os.Stderr, "Found %d clone pairs, %d duplicated lines (%.2f%% duplication ratio)\n",
			duplicationMetrics.ClonePairs,
			duplicationMetrics.DuplicatedLines,
			duplicationMetrics.DuplicationRatio*100)
	}
}

// finalizeNamingMetrics performs naming convention analysis on all collected files
func finalizeNamingMetrics(report *metrics.Report, analyzers *AnalyzerSet, collectedMetrics *CollectedMetrics, cfg *config.Config) {
	// Get all file paths from collected metrics
	var filePaths []string
	for filePath := range collectedMetrics.Files {
		filePaths = append(filePaths, filePath)
	}

	// Skip if no files
	if len(filePaths) == 0 {
		report.Naming = metrics.NamingMetrics{
			FileNameViolations:    0,
			IdentifierViolations:  0,
			PackageNameViolations: 0,
			OverallNamingScore:    1.0,
			FileNameIssues:        []metrics.FileNameViolation{},
			IdentifierIssues:      []metrics.IdentifierViolation{},
			PackageNameIssues:     []metrics.PackageNameViolation{},
		}
		return
	}

	if cfg.Output.Verbose {
		fmt.Fprintf(os.Stderr, "Running naming convention analysis on %d files...\n", len(filePaths))
	}

	// Analyze file names
	fileNameViolations := analyzers.Naming.AnalyzeFileNames(filePaths)

	// Analyze identifiers in each file
	var identifierViolations []metrics.IdentifierViolation
	totalIdentifiers := 0
	for filePath, astFile := range collectedMetrics.Files {
		violations := analyzers.Naming.AnalyzeIdentifiers(astFile, filePath, analyzers.fileSet)
		identifierViolations = append(identifierViolations, violations...)
		totalIdentifiers += countIdentifiers(astFile)
	}

	// Analyze package names (track unique packages)
	var packageNameViolations []metrics.PackageNameViolation
	uniquePackages := make(map[string]struct {
		dirName  string
		filePath string
	})
	for filePath, astFile := range collectedMetrics.Files {
		if astFile.Name != nil {
			pkgName := astFile.Name.Name
			dirName := filepath.Base(filepath.Dir(filePath))
			// Only analyze each unique package once
			if _, exists := uniquePackages[pkgName]; !exists {
				uniquePackages[pkgName] = struct {
					dirName  string
					filePath string
				}{dirName, filePath}
			}
		}
	}

	for pkgName, info := range uniquePackages {
		violations := analyzers.Naming.AnalyzePackageName(pkgName, info.dirName, info.filePath)
		packageNameViolations = append(packageNameViolations, violations...)
	}

	// Calculate naming scores
	fileNamingScore := analyzers.Naming.ComputeFileNamingScore(fileNameViolations, len(filePaths))
	identifierScore := analyzers.Naming.ComputeIdentifierQualityScore(identifierViolations, totalIdentifiers)
	packageScore := analyzers.Naming.ComputePackageNamingScore(packageNameViolations, len(uniquePackages))

	// Calculate overall naming score (weighted average)
	overallScore := (fileNamingScore + identifierScore + packageScore) / 3.0

	// Populate naming metrics
	report.Naming = metrics.NamingMetrics{
		FileNameViolations:    len(fileNameViolations),
		IdentifierViolations:  len(identifierViolations),
		PackageNameViolations: len(packageNameViolations),
		OverallNamingScore:    overallScore,
		FileNameIssues:        fileNameViolations,
		IdentifierIssues:      identifierViolations,
		PackageNameIssues:     packageNameViolations,
	}

	if cfg.Output.Verbose {
		fmt.Fprintf(os.Stderr, "Found %d file, %d identifier, %d package naming violations (score: %.2f)\n",
			len(fileNameViolations),
			len(identifierViolations),
			len(packageNameViolations),
			overallScore)
	}
}

// finalizePlacementMetrics performs placement and cohesion analysis on all collected files
func finalizePlacementMetrics(report *metrics.Report, analyzers *AnalyzerSet, collectedMetrics *CollectedMetrics, cfg *config.Config) {
	// Convert collected files map to slice
	var astFiles []*ast.File
	for _, file := range collectedMetrics.Files {
		astFiles = append(astFiles, file)
	}

	// Skip if no files
	if len(astFiles) == 0 {
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
		fmt.Fprintf(os.Stderr, "Running placement analysis on %d files...\n", len(astFiles))
	}

	// Perform placement analysis
	placementMetrics := analyzers.Placement.Analyze(astFiles, analyzers.fileSet)
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
	if !cfg.Analysis.IncludeDocumentation || len(collectedMetrics.Files) == 0 {
		report.Documentation = metrics.DocumentationMetrics{}
		return
	}

	// Convert files map to slice and group by package
	files, pkgs := prepareDocumentationInput(collectedMetrics.Files)

	if cfg.Output.Verbose {
		fmt.Fprintf(os.Stderr, "Running documentation analysis on %d files in %d packages...\n", len(files), len(pkgs))
	}

	// Run documentation analysis
	docMetrics := analyzers.Documentation.Analyze(files, pkgs)
	report.Documentation = *docMetrics

	if cfg.Output.Verbose {
		fmt.Fprintf(os.Stderr, "Documentation coverage: %.1f%% (%.1f%% packages, %.1f%% functions, %.1f%% types)\n",
			docMetrics.Coverage.Overall,
			docMetrics.Coverage.Packages,
			docMetrics.Coverage.Functions,
			docMetrics.Coverage.Types)
	}
}

// finalizeOrganizationMetrics performs organization analysis on all collected files and packages
func finalizeOrganizationMetrics(report *metrics.Report, analyzers *AnalyzerSet, collectedMetrics *CollectedMetrics, cfg *config.Config, targetPath string) {
	if len(collectedMetrics.Files) == 0 {
		report.Organization = metrics.OrganizationMetrics{}
		return
	}

	orgConfig := getOrganizationConfig(cfg)

	if cfg.Output.Verbose {
		fmt.Fprintf(os.Stderr, "Running organization analysis on %d files...\n", len(collectedMetrics.Files))
	}

	var oversizedFiles []metrics.OversizedFile
	for filePath, astFile := range collectedMetrics.Files {
		result, err := analyzers.Organization.AnalyzeFileSizes(astFile, filePath, orgConfig)
		if err == nil && result != nil {
			oversizedFiles = append(oversizedFiles, *result)
		}
	}

	pkgInfo := buildPackageInfo(collectedMetrics, report)
	oversizedPackages := analyzers.Organization.AnalyzePackageSizes(pkgInfo, orgConfig)

	var filePaths []string
	for filePath := range collectedMetrics.Files {
		filePaths = append(filePaths, filePath)
	}
	deepDirs := analyzers.Organization.AnalyzeDirectoryDepth(filePaths, targetPath, orgConfig)

	graphData := buildImportGraphData(collectedMetrics)
	importMetrics, _ := analyzers.Organization.AnalyzeImportGraph(graphData, orgConfig)

	var highFanIn []metrics.FanInPackage
	var highFanOut []metrics.FanOutPackage
	avgStability := 0.0
	if importMetrics != nil {
		highFanIn = importMetrics.HighFanInPackages
		highFanOut = importMetrics.HighFanOutPackages
		avgStability = importMetrics.AvgInstability
	}

	report.Organization = metrics.OrganizationMetrics{
		OversizedFiles:      oversizedFiles,
		OversizedPackages:   oversizedPackages,
		DeepDirectories:     deepDirs,
		HighFanInPackages:   highFanIn,
		HighFanOutPackages:  highFanOut,
		AvgPackageStability: avgStability,
	}

	if cfg.Output.Verbose {
		fmt.Fprintf(os.Stderr, "Found %d oversized files, %d oversized packages, %d deep directories\n",
			len(oversizedFiles), len(oversizedPackages), len(deepDirs))
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
		merged.TypeParameters.Count += gen.TypeParameters.Count
		merged.TypeParameters.Complexity = append(
			merged.TypeParameters.Complexity,
			gen.TypeParameters.Complexity...)

		for k, v := range gen.TypeParameters.Constraints {
			merged.TypeParameters.Constraints[k] += v
		}
		for k, v := range gen.ConstraintUsage {
			merged.ConstraintUsage[k] += v
		}

		merged.Instantiations.Functions = append(
			merged.Instantiations.Functions,
			gen.Instantiations.Functions...)
		merged.Instantiations.Types = append(
			merged.Instantiations.Types,
			gen.Instantiations.Types...)
		merged.Instantiations.Methods = append(
			merged.Instantiations.Methods,
			gen.Instantiations.Methods...)
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
	wd, _ := os.Getwd()
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
	repoPath, _ := os.Getwd()
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
