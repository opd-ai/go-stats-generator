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

type methodInfo struct {
	receiverType string
	file         string
}

// NewPlacementAnalyzer creates a new placement analyzer with default thresholds
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

// Analyze performs comprehensive placement analysis
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

// buildSymbolIndex constructs the complete symbol table for all files
func (pa *PlacementAnalyzer) buildSymbolIndex(files []*ast.File, fset *token.FileSet) {
	// First pass: collect all definitions
	for _, file := range files {
		filename := fset.Position(file.Pos()).Filename
		filename = filepath.ToSlash(filename)

		ast.Inspect(file, func(n ast.Node) bool {
			switch decl := n.(type) {
			case *ast.FuncDecl:
				// Handle methods and functions
				funcName := decl.Name.Name
				if decl.Recv != nil && len(decl.Recv.List) > 0 {
					// It's a method
					recvType := ExtractReceiverType(decl.Recv.List[0].Type)
					methodName := recvType + "." + funcName
					pa.methods[methodName] = methodInfo{
						receiverType: recvType,
						file:         filename,
					}
					pa.recordSymbolDef(methodName, filename)
				} else {
					// Regular function
					pa.recordSymbolDef(funcName, filename)
				}

			case *ast.GenDecl:
				// Handle type, var, const declarations
				for _, spec := range decl.Specs {
					switch s := spec.(type) {
					case *ast.TypeSpec:
						typeName := s.Name.Name
						pa.recordSymbolDef(typeName, filename)
						// Track receiver types for later method matching
						pa.receiverFiles[typeName] = filename

					case *ast.ValueSpec:
						for _, name := range s.Names {
							pa.recordSymbolDef(name.Name, filename)
						}
					}
				}
			}
			return true
		})
	}

	// Second pass: collect all references
	for _, file := range files {
		filename := fset.Position(file.Pos()).Filename
		filename = filepath.ToSlash(filename)

		// Track if we're inside a function/method declaration
		var currentFunc string

		ast.Inspect(file, func(n ast.Node) bool {
			// Track which function we're in
			if funcDecl, ok := n.(*ast.FuncDecl); ok {
				currentFunc = funcDecl.Name.Name
				if funcDecl.Recv != nil && len(funcDecl.Recv.List) > 0 {
					recvType := ExtractReceiverType(funcDecl.Recv.List[0].Type)
					currentFunc = recvType + "." + funcDecl.Name.Name
				}
			}

			// Look for identifier references (function calls, type usage, etc.)
			if ident, ok := n.(*ast.Ident); ok {
				// Skip if this is the identifier being defined
				if ident.Obj != nil && ident.Obj.Pos() == ident.Pos() {
					return true
				}

				// Only count references to symbols we know about
				if _, exists := pa.symbolDefs[ident.Name]; exists {
					// Don't count self-references
					if ident.Name != currentFunc {
						pa.recordSymbolRef(ident.Name, filename)
					}
				}
			}

			return true
		})
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

// AnalyzeFunctionAffinity identifies functions that may be misplaced
func (pa *PlacementAnalyzer) AnalyzeFunctionAffinity() []metrics.MisplacedFunctionIssue {
	var issues []metrics.MisplacedFunctionIssue

	for symbol, defFile := range pa.symbolDefs {
		// Skip methods (handled separately)
		if strings.Contains(symbol, ".") {
			continue
		}

		// Get references to this symbol from different files
		refs := pa.symbolRefs[symbol]
		if refs == nil || len(refs) == 0 {
			continue
		}

		// Calculate affinity scores
		totalRefs := 0
		for _, count := range refs {
			totalRefs += count
		}

		sameFileRefs := refs[defFile]
		currentAffinity := float64(sameFileRefs) / float64(totalRefs)

		// Find the file with highest affinity
		bestFile := defFile
		bestAffinity := currentAffinity
		var referencedSymbols []string

		for file, count := range refs {
			affinity := float64(count) / float64(totalRefs)
			if affinity > bestAffinity+pa.affinityMargin {
				bestFile = file
				bestAffinity = affinity
			}
		}

		// Collect symbols referenced by this function
		// (This would require AST analysis of function body - simplified for now)
		referencedSymbols = []string{}

		// If best file is different from current file, flag it
		if bestFile != defFile && bestAffinity > currentAffinity+pa.affinityMargin {
			severity := "medium"
			if bestAffinity-currentAffinity > 2*pa.affinityMargin {
				severity = "high"
			}

			issues = append(issues, metrics.MisplacedFunctionIssue{
				Name:              symbol,
				CurrentFile:       defFile,
				SuggestedFile:     bestFile,
				CurrentAffinity:   currentAffinity,
				SuggestedAffinity: bestAffinity,
				ReferencedSymbols: referencedSymbols,
				Severity:          severity,
			})
		}
	}

	return issues
}

// AnalyzeMethodPlacement checks if methods are defined in the same file as their receiver
func (pa *PlacementAnalyzer) AnalyzeMethodPlacement() []metrics.MisplacedMethodIssue {
	var issues []metrics.MisplacedMethodIssue

	for methodName, info := range pa.methods {
		receiverFile, exists := pa.receiverFiles[info.receiverType]
		if !exists {
			// Receiver type not found (possibly from another package)
			continue
		}

		if info.file != receiverFile {
			// Method is in a different file from its receiver
			distance := "same_package"
			// Simple heuristic: if files don't share directory, assume different package
			methodDir := filepath.Dir(info.file)
			receiverDir := filepath.Dir(receiverFile)
			if methodDir != receiverDir {
				distance = "different_package"
			}

			severity := "medium"
			if distance == "different_package" {
				severity = "high"
			}

			issues = append(issues, metrics.MisplacedMethodIssue{
				MethodName:   methodName,
				ReceiverType: info.receiverType,
				CurrentFile:  info.file,
				ReceiverFile: receiverFile,
				Distance:     distance,
				Severity:     severity,
			})
		}
	}

	return issues
}

// AnalyzeFileCohesion identifies files with low internal cohesion
func (pa *PlacementAnalyzer) AnalyzeFileCohesion() []metrics.FileCohesionIssue {
	var issues []metrics.FileCohesionIssue

	for file := range pa.fileRefs {
		cohesion := pa.calculateCohesion(file)

		if cohesion < pa.minCohesion {
			intraRefs := 0
			totalRefs := 0

			for symbol, count := range pa.fileRefs[file] {
				totalRefs += count
				if pa.symbolDefs[symbol] == file {
					intraRefs += count
				}
			}

			// Suggest splits based on symbol clustering (simplified)
			suggestedSplits := pa.suggestSplits(file)

			severity := "low"
			if cohesion < pa.minCohesion/2 {
				severity = "medium"
			}
			if cohesion < pa.minCohesion/3 {
				severity = "high"
			}

			issues = append(issues, metrics.FileCohesionIssue{
				File:            file,
				CohesionScore:   cohesion,
				IntraFileRefs:   intraRefs,
				TotalRefs:       totalRefs,
				SuggestedSplits: suggestedSplits,
				Severity:        severity,
			})
		}
	}

	return issues
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
