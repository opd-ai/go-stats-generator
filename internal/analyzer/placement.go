package analyzer

import (
	"go/ast"
	"go/token"
	"path/filepath"
	"strings"

	"github.com/opd-ai/go-stats-generator/internal/metrics"
)

// PlacementAnalyzer performs placement and cohesion analysis on Go code
type PlacementAnalyzer struct {
	// Symbol definitions: symbolName -> file path
	symbolDefs map[string]string
	// File to symbols: file path -> symbols defined in that file
	fileSymbols map[string][]string
	// Symbol references: symbolName -> list of (file, count) where it's referenced
	symbolRefs map[string]map[string]int
	// File references: file -> symbols it references
	fileRefs map[string]map[string]int
	// Receiver types: receiverType -> file where it's defined
	receiverFiles map[string]string
	// Methods: methodName -> (receiverType, file)
	methods map[string]methodInfo
	// File set for position information
	fset *token.FileSet
	// Configuration
	affinityMargin float64
	minCohesion    float64
}

// methodInfo stores information about a method's receiver type and definition file.
type methodInfo struct {
	receiverType string
	file         string
}

// NewPlacementAnalyzer creates a new placement analyzer with configurable
// NewPlacementAnalyzer uses affinity margin and minimum cohesion thresholds for misplacement detection.
func NewPlacementAnalyzer(affinityMargin, minCohesion float64) *PlacementAnalyzer {
	return &PlacementAnalyzer{
		symbolDefs:     make(map[string]string),
		fileSymbols:    make(map[string][]string),
		symbolRefs:     make(map[string]map[string]int),
		fileRefs:       make(map[string]map[string]int),
		receiverFiles:  make(map[string]string),
		methods:        make(map[string]methodInfo),
		affinityMargin: affinityMargin,
		minCohesion:    minCohesion,
	}
}

// Analyze performs comprehensive placement analysis including function affinity,
// Analyze returns method placement and file cohesion metrics for all provided AST files.
func (pa *PlacementAnalyzer) Analyze(files []*ast.File, fset *token.FileSet) metrics.PlacementMetrics {
	pa.fset = fset
	pa.buildSymbolIndex(files, fset)

	functionIssues := pa.AnalyzeFunctionAffinity()
	methodIssues := pa.AnalyzeMethodPlacement()
	cohesionIssues := pa.AnalyzeFileCohesion()

	avgCohesion := 0.0
	if len(pa.fileRefs) > 0 {
		totalCohesion := 0.0
		for file := range pa.fileRefs {
			totalCohesion += pa.calculateCohesion(file)
		}
		avgCohesion = totalCohesion / float64(len(pa.fileRefs))
	}

	return metrics.PlacementMetrics{
		MisplacedFunctions: len(functionIssues),
		MisplacedMethods:   len(methodIssues),
		LowCohesionFiles:   len(cohesionIssues),
		AvgFileCohesion:    avgCohesion,
		FunctionIssues:     functionIssues,
		MethodIssues:       methodIssues,
		CohesionIssues:     cohesionIssues,
	}
}

// AnalyzeMap performs the same placement analysis as Analyze but accepts a
// map[filename]*ast.File rather than a slice plus a token.FileSet.
// The map keys are used directly as filenames, so no fset position lookup is
// needed. This allows the call site to work with per-file token.FileSets (where
// a shared fset has no position information) without changing the existing Analyze
// API or breaking its tests.
func (pa *PlacementAnalyzer) AnalyzeMap(files map[string]*ast.File) metrics.PlacementMetrics {
	pa.buildSymbolIndexFromMap(files)

	functionIssues := pa.AnalyzeFunctionAffinity()
	methodIssues := pa.AnalyzeMethodPlacement()
	cohesionIssues := pa.AnalyzeFileCohesion()

	avgCohesion := 0.0
	if len(pa.fileRefs) > 0 {
		totalCohesion := 0.0
		for file := range pa.fileRefs {
			totalCohesion += pa.calculateCohesion(file)
		}
		avgCohesion = totalCohesion / float64(len(pa.fileRefs))
	}

	return metrics.PlacementMetrics{
		MisplacedFunctions: len(functionIssues),
		MisplacedMethods:   len(methodIssues),
		LowCohesionFiles:   len(cohesionIssues),
		AvgFileCohesion:    avgCohesion,
		FunctionIssues:     functionIssues,
		MethodIssues:       methodIssues,
		CohesionIssues:     cohesionIssues,
	}
}

// buildSymbolIndexFromMap constructs the symbol table using map keys as filenames,
// avoiding any fset.Position call.
func (pa *PlacementAnalyzer) buildSymbolIndexFromMap(files map[string]*ast.File) {
	for filename, file := range files {
		pa.collectDefinitionsFromFile(file, filepath.ToSlash(filename))
	}
	for filename, file := range files {
		pa.collectReferencesFromFile(file, filepath.ToSlash(filename))
	}
}

// buildSymbolIndex constructs the complete symbol table for all files
func (pa *PlacementAnalyzer) buildSymbolIndex(files []*ast.File, fset *token.FileSet) {
	pa.collectDefinitions(files, fset)
	pa.collectReferences(files, fset)
}

// collectDefinitions performs first pass to collect all symbol definitions
func (pa *PlacementAnalyzer) collectDefinitions(files []*ast.File, fset *token.FileSet) {
	for _, file := range files {
		filename := fset.Position(file.Pos()).Filename
		filename = filepath.ToSlash(filename)
		pa.collectDefinitionsFromFile(file, filename)
	}
}

// collectDefinitionsFromFile extracts symbol definitions from a single file
func (pa *PlacementAnalyzer) collectDefinitionsFromFile(file *ast.File, filename string) {
	ast.Inspect(file, func(n ast.Node) bool {
		switch decl := n.(type) {
		case *ast.FuncDecl:
			pa.processFuncDecl(decl, filename)
		case *ast.GenDecl:
			pa.processGenDecl(decl, filename)
		}
		return true
	})
}

// processFuncDecl handles function and method declarations
func (pa *PlacementAnalyzer) processFuncDecl(decl *ast.FuncDecl, filename string) {
	funcName := decl.Name.Name
	if decl.Recv != nil && len(decl.Recv.List) > 0 {
		recvType := ExtractReceiverType(decl.Recv.List[0].Type)
		methodName := recvType + "." + funcName
		pa.methods[methodName] = methodInfo{
			receiverType: recvType,
			file:         filename,
		}
		pa.recordSymbolDef(methodName, filename)
	} else {
		pa.recordSymbolDef(funcName, filename)
	}
}

// processGenDecl handles type, var, and const declarations
func (pa *PlacementAnalyzer) processGenDecl(decl *ast.GenDecl, filename string) {
	for _, spec := range decl.Specs {
		switch s := spec.(type) {
		case *ast.TypeSpec:
			typeName := s.Name.Name
			pa.recordSymbolDef(typeName, filename)
			pa.receiverFiles[typeName] = filename
		case *ast.ValueSpec:
			for _, name := range s.Names {
				pa.recordSymbolDef(name.Name, filename)
			}
		}
	}
}

// collectReferences performs second pass to collect all symbol references
func (pa *PlacementAnalyzer) collectReferences(files []*ast.File, fset *token.FileSet) {
	for _, file := range files {
		filename := fset.Position(file.Pos()).Filename
		filename = filepath.ToSlash(filename)
		pa.collectReferencesFromFile(file, filename)
	}
}

// collectReferencesFromFile extracts symbol references from a single file
func (pa *PlacementAnalyzer) collectReferencesFromFile(file *ast.File, filename string) {
	var currentFunc string

	ast.Inspect(file, func(n ast.Node) bool {
		if funcDecl, ok := n.(*ast.FuncDecl); ok {
			currentFunc = pa.getFuncDeclName(funcDecl)
		}

		if ident, ok := n.(*ast.Ident); ok {
			pa.processIdentRef(ident, filename, currentFunc)
		}

		return true
	})
}

// getFuncDeclName gets the full name of a function or method declaration
func (pa *PlacementAnalyzer) getFuncDeclName(decl *ast.FuncDecl) string {
	if decl.Recv != nil && len(decl.Recv.List) > 0 {
		recvType := ExtractReceiverType(decl.Recv.List[0].Type)
		return recvType + "." + decl.Name.Name
	}
	return decl.Name.Name
}

// processIdentRef processes an identifier reference
func (pa *PlacementAnalyzer) processIdentRef(ident *ast.Ident, filename, currentFunc string) {
	if ident.Obj != nil && ident.Obj.Pos() == ident.Pos() {
		return
	}

	if _, exists := pa.symbolDefs[ident.Name]; exists {
		if ident.Name != currentFunc {
			pa.recordSymbolRef(ident.Name, filename)
		}
	}
}

// recordSymbolDef records that a symbol is defined in a file
func (pa *PlacementAnalyzer) recordSymbolDef(symbol, file string) {
	pa.symbolDefs[symbol] = file
	pa.fileSymbols[file] = append(pa.fileSymbols[file], symbol)
}

// recordSymbolRef records that a symbol is referenced in a file
func (pa *PlacementAnalyzer) recordSymbolRef(symbol, file string) {
	if pa.symbolRefs[symbol] == nil {
		pa.symbolRefs[symbol] = make(map[string]int)
	}
	pa.symbolRefs[symbol][file]++

	if pa.fileRefs[file] == nil {
		pa.fileRefs[file] = make(map[string]int)
	}
	pa.fileRefs[file][symbol]++
}

// AnalyzeFunctionAffinity identifies functions that may be misplaced based on
// AnalyzeFunctionAffinity suggests files where a function would have higher affinity based on reference patterns.
func (pa *PlacementAnalyzer) AnalyzeFunctionAffinity() []metrics.MisplacedFunctionIssue {
	var issues []metrics.MisplacedFunctionIssue

	for symbol, defFile := range pa.symbolDefs {
		if issue := pa.checkFunctionPlacement(symbol, defFile); issue != nil {
			issues = append(issues, *issue)
		}
	}

	return issues
}

// checkFunctionPlacement analyzes a single function for potential misplacement
func (pa *PlacementAnalyzer) checkFunctionPlacement(symbol, defFile string) *metrics.MisplacedFunctionIssue {
	if pa.shouldSkipSymbol(symbol) {
		return nil
	}

	refs := pa.symbolRefs[symbol]
	if refs == nil || len(refs) == 0 {
		return nil
	}

	totalRefs := pa.calculateTotalRefs(refs)
	currentAffinity := pa.calculateAffinity(refs[defFile], totalRefs)
	bestFile, bestAffinity := pa.findBestAffinityFile(refs, totalRefs, defFile, currentAffinity)

	if pa.isMisplaced(bestFile, defFile, bestAffinity, currentAffinity) {
		return pa.createMisplacedIssue(symbol, defFile, bestFile, currentAffinity, bestAffinity)
	}

	return nil
}

// shouldSkipSymbol returns true if the symbol should be skipped (e.g., methods)
func (pa *PlacementAnalyzer) shouldSkipSymbol(symbol string) bool {
	return strings.Contains(symbol, ".")
}

// calculateTotalRefs sums all references across files
func (pa *PlacementAnalyzer) calculateTotalRefs(refs map[string]int) int {
	total := 0
	for _, count := range refs {
		total += count
	}
	return total
}

// calculateAffinity computes affinity score as ratio of refs to total
func (pa *PlacementAnalyzer) calculateAffinity(fileRefs, totalRefs int) float64 {
	if totalRefs == 0 {
		return 0.0
	}
	return float64(fileRefs) / float64(totalRefs)
}

// findBestAffinityFile identifies the file with highest reference affinity
func (pa *PlacementAnalyzer) findBestAffinityFile(refs map[string]int, totalRefs int, defFile string, currentAffinity float64) (string, float64) {
	bestFile := defFile
	bestAffinity := currentAffinity

	for file, count := range refs {
		affinity := pa.calculateAffinity(count, totalRefs)
		if affinity > bestAffinity+pa.affinityMargin {
			bestFile = file
			bestAffinity = affinity
		}
	}

	return bestFile, bestAffinity
}

// isMisplaced checks if function should be flagged as misplaced
func (pa *PlacementAnalyzer) isMisplaced(bestFile, defFile string, bestAffinity, currentAffinity float64) bool {
	return bestFile != defFile && bestAffinity > currentAffinity+pa.affinityMargin
}

// createMisplacedIssue builds a misplaced function issue with appropriate severity
func (pa *PlacementAnalyzer) createMisplacedIssue(symbol, defFile, bestFile string, currentAffinity, bestAffinity float64) *metrics.MisplacedFunctionIssue {
	severity := pa.calculateSeverity(bestAffinity, currentAffinity)

	return &metrics.MisplacedFunctionIssue{
		Name:              symbol,
		CurrentFile:       defFile,
		SuggestedFile:     bestFile,
		CurrentAffinity:   currentAffinity,
		SuggestedAffinity: bestAffinity,
		ReferencedSymbols: []string{},
		Severity:          severity,
	}
}

// calculateSeverity determines issue severity based on affinity difference
func (pa *PlacementAnalyzer) calculateSeverity(bestAffinity, currentAffinity float64) metrics.SeverityLevel {
	if bestAffinity-currentAffinity > 2*pa.affinityMargin {
		return metrics.SeverityLevelCritical
	}
	return metrics.SeverityLevelWarning
}

// AnalyzeMethodPlacement checks if methods are defined in the same file as their receiver
func (pa *PlacementAnalyzer) AnalyzeMethodPlacement() []metrics.MisplacedMethodIssue {
	var issues []metrics.MisplacedMethodIssue
	for methodName, info := range pa.methods {
		if issue := pa.checkMethodPlacement(methodName, info); issue != nil {
			issues = append(issues, *issue)
		}
	}
	return issues
}

// checkMethodPlacement checks if a single method is misplaced
func (pa *PlacementAnalyzer) checkMethodPlacement(methodName string, info methodInfo) *metrics.MisplacedMethodIssue {
	receiverFile, exists := pa.receiverFiles[info.receiverType]
	if !exists || info.file == receiverFile {
		return nil
	}
	distance := pa.calculateMethodDistance(info.file, receiverFile)
	return &metrics.MisplacedMethodIssue{
		MethodName:   methodName,
		ReceiverType: info.receiverType,
		CurrentFile:  info.file,
		ReceiverFile: receiverFile,
		Distance:     distance,
		Severity:     pa.determinePlacementSeverity(distance),
	}
}

// calculateMethodDistance determines the distance between method and receiver
func (pa *PlacementAnalyzer) calculateMethodDistance(methodFile, receiverFile string) string {
	if filepath.Dir(methodFile) != filepath.Dir(receiverFile) {
		return "different_package"
	}
	return "same_package"
}

// determinePlacementSeverity determines severity based on distance
func (pa *PlacementAnalyzer) determinePlacementSeverity(distance string) metrics.SeverityLevel {
	if distance == "different_package" {
		return metrics.SeverityLevelCritical
	}
	return metrics.SeverityLevelWarning
}

// AnalyzeFileCohesion identifies files with low internal cohesion based on
// AnalyzeFileCohesion measures the ratio of internal symbol references versus external references.
func (pa *PlacementAnalyzer) AnalyzeFileCohesion() []metrics.FileCohesionIssue {
	var issues []metrics.FileCohesionIssue

	for file := range pa.fileRefs {
		cohesion := pa.calculateCohesion(file)

		if cohesion < pa.minCohesion {
			issue := pa.buildCohesionIssue(file, cohesion)
			issues = append(issues, issue)
		}
	}

	return issues
}

// buildCohesionIssue creates a FileCohesionIssue for a file with low cohesion
func (pa *PlacementAnalyzer) buildCohesionIssue(file string, cohesion float64) metrics.FileCohesionIssue {
	intraRefs, totalRefs := pa.countFileReferences(file)
	suggestedSplits := pa.suggestSplits(file)
	severity := pa.determineCohesionSeverity(cohesion)

	return metrics.FileCohesionIssue{
		File:            file,
		CohesionScore:   cohesion,
		IntraFileRefs:   intraRefs,
		TotalRefs:       totalRefs,
		SuggestedSplits: suggestedSplits,
		Severity:        severity,
	}
}

// countFileReferences counts intra-file and total references for a file
func (pa *PlacementAnalyzer) countFileReferences(file string) (intraRefs, totalRefs int) {
	for symbol, count := range pa.fileRefs[file] {
		totalRefs += count
		if pa.symbolDefs[symbol] == file {
			intraRefs += count
		}
	}
	return intraRefs, totalRefs
}

// determineCohesionSeverity calculates severity based on cohesion score
func (pa *PlacementAnalyzer) determineCohesionSeverity(cohesion float64) metrics.SeverityLevel {
	if cohesion < pa.minCohesion/3 {
		return metrics.SeverityLevelViolation
	}
	if cohesion < pa.minCohesion/2 {
		return metrics.SeverityLevelWarning
	}
	return metrics.SeverityLevelInfo
}

// calculateCohesion computes the cohesion score for a file
func (pa *PlacementAnalyzer) calculateCohesion(file string) float64 {
	refs := pa.fileRefs[file]
	if len(refs) == 0 {
		return 1.0 // Perfect cohesion for files with no external refs
	}

	intraRefs := 0
	totalRefs := 0

	for symbol, count := range refs {
		totalRefs += count
		if pa.symbolDefs[symbol] == file {
			intraRefs += count
		}
	}

	if totalRefs == 0 {
		return 1.0
	}

	return float64(intraRefs) / float64(totalRefs)
}

// suggestSplits suggests logical file splits based on symbol clustering
func (pa *PlacementAnalyzer) suggestSplits(file string) []string {
	// Simplified implementation: suggest splitting by dominant external reference
	externalRefs := make(map[string]int)

	for symbol, count := range pa.fileRefs[file] {
		defFile := pa.symbolDefs[symbol]
		if defFile != file && defFile != "" {
			externalRefs[defFile] += count
		}
	}

	var suggestions []string
	for refFile, count := range externalRefs {
		if count >= 3 { // Threshold for suggesting a split
			baseName := filepath.Base(refFile)
			baseName = strings.TrimSuffix(baseName, ".go")
			suggestions = append(suggestions, baseName+"_related.go")
		}
	}

	if len(suggestions) == 0 {
		suggestions = []string{"part1.go", "part2.go"}
	}

	return suggestions
}
