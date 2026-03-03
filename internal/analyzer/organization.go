package analyzer

import (
	"go/ast"
	"go/token"
	"os"
	"strings"

	"github.com/opd-ai/go-stats-generator/internal/metrics"
)

// OrganizationAnalyzer analyzes file and package organization
type OrganizationAnalyzer struct {
	fset *token.FileSet
}

// NewOrganizationAnalyzer creates a new organization analyzer
func NewOrganizationAnalyzer(fset *token.FileSet) *OrganizationAnalyzer {
	return &OrganizationAnalyzer{
		fset: fset,
	}
}

// OrganizationConfig holds threshold configuration
type OrganizationConfig struct {
	MaxFileLines     int
	MaxFileFunctions int
	MaxFileTypes     int
}

// DefaultOrganizationConfig returns default configuration
func DefaultOrganizationConfig() OrganizationConfig {
	return OrganizationConfig{
		MaxFileLines:     500,
		MaxFileFunctions: 20,
		MaxFileTypes:     5,
	}
}

// AnalyzeFileSizes analyzes file sizes and complexity
func (oa *OrganizationAnalyzer) AnalyzeFileSizes(file *ast.File, filePath string, config OrganizationConfig) (*metrics.OversizedFile, error) {
	lines := oa.countFileLines(filePath)
	funcCount := oa.countFunctions(file)
	typeCount := oa.countTypes(file)

	burden := oa.calculateBurden(lines, funcCount, typeCount)

	if !oa.isOversized(lines, funcCount, typeCount, config) {
		return nil, nil
	}

	return &metrics.OversizedFile{
		File:              filePath,
		Lines:             lines,
		FunctionCount:     funcCount,
		TypeCount:         typeCount,
		MaintenanceBurden: burden,
		Severity:          oa.getSeverity(lines, funcCount, typeCount, config),
		Suggestions:       oa.getSuggestions(lines, funcCount, typeCount, config),
	}, nil
}

// countFileLines counts lines in an entire file
func (oa *OrganizationAnalyzer) countFileLines(filePath string) metrics.LineMetrics {
	src, err := os.ReadFile(filePath)
	if err != nil {
		return metrics.LineMetrics{}
	}

	lines := strings.Split(string(src), "\n")
	return oa.analyzeLinesInFile(lines)
}

// analyzeLinesInFile categorizes lines
func (oa *OrganizationAnalyzer) analyzeLinesInFile(lines []string) metrics.LineMetrics {
	var codeLines, commentLines, blankLines int
	inBlockComment := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if trimmed == "" {
			blankLines++
			continue
		}

		lineType := oa.classifyLine(trimmed, &inBlockComment)
		switch lineType {
		case "code":
			codeLines++
		case "comment":
			commentLines++
		case "mixed":
			codeLines++
		}
	}

	total := len(lines)
	return metrics.LineMetrics{
		Total:    total,
		Code:     codeLines,
		Comments: commentLines,
		Blank:    blankLines,
	}
}

// classifyLine determines line type
func (oa *OrganizationAnalyzer) classifyLine(line string, inBlock *bool) string {
	if oa.isBlockCommentStart(line, inBlock) {
		return "comment"
	}

	if *inBlock {
		return oa.handleBlockComment(line, inBlock)
	}

	if strings.HasPrefix(line, "//") {
		return "comment"
	}

	if strings.Contains(line, "//") {
		return "mixed"
	}

	return "code"
}

// isBlockCommentStart checks for block comment start
func (oa *OrganizationAnalyzer) isBlockCommentStart(line string, inBlock *bool) bool {
	if !strings.HasPrefix(line, "/*") {
		return false
	}

	*inBlock = true
	if strings.HasSuffix(line, "*/") {
		*inBlock = false
	}
	return true
}

// handleBlockComment processes lines in block comment
func (oa *OrganizationAnalyzer) handleBlockComment(line string, inBlock *bool) string {
	if strings.Contains(line, "*/") {
		*inBlock = false
	}
	return "comment"
}

// countFunctions counts functions and methods
func (oa *OrganizationAnalyzer) countFunctions(file *ast.File) int {
	count := 0
	for _, decl := range file.Decls {
		if _, ok := decl.(*ast.FuncDecl); ok {
			count++
		}
	}
	return count
}

// countTypes counts type declarations
func (oa *OrganizationAnalyzer) countTypes(file *ast.File) int {
	count := 0
	for _, decl := range file.Decls {
		if genDecl, ok := decl.(*ast.GenDecl); ok {
			if genDecl.Tok == token.TYPE {
				count += len(genDecl.Specs)
			}
		}
	}
	return count
}

// calculateBurden computes composite burden score
func (oa *OrganizationAnalyzer) calculateBurden(lines metrics.LineMetrics, funcCount, typeCount int) float64 {
	lineScore := float64(lines.Total) / 500.0
	funcScore := float64(funcCount) / 20.0
	typeScore := float64(typeCount) / 5.0

	return (lineScore + funcScore + typeScore) / 3.0
}

// isOversized checks if file exceeds thresholds
func (oa *OrganizationAnalyzer) isOversized(lines metrics.LineMetrics, funcCount, typeCount int, config OrganizationConfig) bool {
	return lines.Total > config.MaxFileLines ||
		funcCount > config.MaxFileFunctions ||
		typeCount > config.MaxFileTypes
}

// getSeverity determines severity level
func (oa *OrganizationAnalyzer) getSeverity(lines metrics.LineMetrics, funcCount, typeCount int, config OrganizationConfig) string {
	criticalCount := 0

	if lines.Total > config.MaxFileLines*2 {
		criticalCount++
	}
	if funcCount > config.MaxFileFunctions*2 {
		criticalCount++
	}
	if typeCount > config.MaxFileTypes*2 {
		criticalCount++
	}

	if criticalCount >= 2 {
		return "critical"
	}
	if criticalCount == 1 {
		return "high"
	}
	return "medium"
}

// getSuggestions generates improvement suggestions
func (oa *OrganizationAnalyzer) getSuggestions(lines metrics.LineMetrics, funcCount, typeCount int, config OrganizationConfig) []string {
	var suggestions []string

	if lines.Total > config.MaxFileLines {
		suggestions = append(suggestions, "Consider splitting file - exceeds maximum line count")
	}
	if funcCount > config.MaxFileFunctions {
		suggestions = append(suggestions, "Too many functions - group related functions into separate files")
	}
	if typeCount > config.MaxFileTypes {
		suggestions = append(suggestions, "Too many types - separate type definitions into focused files")
	}

	return suggestions
}
