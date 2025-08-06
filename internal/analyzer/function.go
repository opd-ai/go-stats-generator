package analyzer

import (
	"go/ast"
	"go/token"
	"os"
	"strings"

	"github.com/opd-ai/go-stats-generator/internal/metrics"
)

// FunctionAnalyzer analyzes functions and methods in Go source code
type FunctionAnalyzer struct {
	fset *token.FileSet
}

// NewFunctionAnalyzer creates a new function analyzer
func NewFunctionAnalyzer(fset *token.FileSet) *FunctionAnalyzer {
	return &FunctionAnalyzer{
		fset: fset,
	}
}

// AnalyzeFunctions analyzes all functions in an AST file
func (fa *FunctionAnalyzer) AnalyzeFunctions(file *ast.File, pkgName string) ([]metrics.FunctionMetrics, error) {
	var functions []metrics.FunctionMetrics

	// Analyze top-level functions
	for _, decl := range file.Decls {
		if funcDecl, ok := decl.(*ast.FuncDecl); ok {
			function, err := fa.analyzeFunction(funcDecl, pkgName, file.Name.Name)
			if err != nil {
				continue // Log warning and continue
			}
			functions = append(functions, function)
		}
	}

	return functions, nil
}

// analyzeFunction analyzes a single function declaration
func (fa *FunctionAnalyzer) analyzeFunction(funcDecl *ast.FuncDecl, fileName, pkgName string) (metrics.FunctionMetrics, error) {
	pos := fa.fset.Position(funcDecl.Pos())

	function := metrics.FunctionMetrics{
		Name:       funcDecl.Name.Name,
		Package:    pkgName,
		File:       fileName,
		Line:       pos.Line,
		IsExported: ast.IsExported(funcDecl.Name.Name),
		IsMethod:   funcDecl.Recv != nil,
	}

	// Analyze receiver type for methods
	if funcDecl.Recv != nil {
		function.ReceiverType = fa.extractReceiverType(funcDecl.Recv)
	}

	// Analyze function signature
	function.Signature = fa.analyzeSignature(funcDecl.Type)

	// Count lines
	function.Lines = fa.countLines(funcDecl)

	// Calculate complexity
	function.Complexity = fa.calculateComplexity(funcDecl)

	// Analyze documentation
	function.Documentation = fa.analyzeDocumentation(funcDecl.Doc)

	return function, nil
}

// extractReceiverType extracts the receiver type name from a method
func (fa *FunctionAnalyzer) extractReceiverType(recv *ast.FieldList) string {
	if recv == nil || len(recv.List) == 0 {
		return ""
	}

	field := recv.List[0]
	if field.Type == nil {
		return ""
	}

	switch t := field.Type.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		if ident, ok := t.X.(*ast.Ident); ok {
			return "*" + ident.Name
		}
	}

	return ""
}

// analyzeSignature analyzes function signature complexity
func (fa *FunctionAnalyzer) analyzeSignature(funcType *ast.FuncType) metrics.FunctionSignature {
	signature := metrics.FunctionSignature{}

	fa.analyzeSignatureParameters(funcType, &signature)
	fa.analyzeSignatureReturns(funcType, &signature)
	fa.analyzeGenericParameters(funcType, &signature)

	// Calculate signature complexity score
	signature.ComplexityScore = fa.calculateSignatureComplexity(signature)

	return signature
}

// analyzeSignatureParameters analyzes function parameters
func (fa *FunctionAnalyzer) analyzeSignatureParameters(funcType *ast.FuncType, signature *metrics.FunctionSignature) {
	if funcType.Params == nil {
		return
	}

	signature.ParameterCount = len(funcType.Params.List)

	for _, param := range funcType.Params.List {
		// Check for variadic parameters
		if _, ok := param.Type.(*ast.Ellipsis); ok {
			signature.VariadicUsage = true
		}

		// Check for interface parameters
		if fa.isInterfaceType(param.Type) {
			signature.InterfaceParams++
		}
	}
}

// analyzeSignatureReturns analyzes function return values
func (fa *FunctionAnalyzer) analyzeSignatureReturns(funcType *ast.FuncType, signature *metrics.FunctionSignature) {
	if funcType.Results == nil {
		return
	}

	signature.ReturnCount = len(funcType.Results.List)

	// Check if function returns error
	for _, result := range funcType.Results.List {
		if fa.isErrorType(result.Type) {
			signature.ErrorReturn = true
			break
		}
	}
}

// analyzeGenericParameters analyzes generic type parameters (Go 1.18+)
func (fa *FunctionAnalyzer) analyzeGenericParameters(funcType *ast.FuncType, signature *metrics.FunctionSignature) {
	if funcType.TypeParams == nil {
		return
	}

	for _, param := range funcType.TypeParams.List {
		for _, name := range param.Names {
			genericParam := metrics.GenericParam{
				Name:        name.Name,
				Constraints: fa.extractConstraints(param.Type),
			}
			signature.GenericParams = append(signature.GenericParams, genericParam)
		}
	}
}

// countLines counts various types of lines in a function with precise categorization
func (fa *FunctionAnalyzer) countLines(funcDecl *ast.FuncDecl) metrics.LineMetrics {
	if funcDecl.Body == nil {
		return metrics.LineMetrics{}
	}

	start := fa.fset.Position(funcDecl.Body.Lbrace)
	end := fa.fset.Position(funcDecl.Body.Rbrace)

	// Get the source file to analyze line by line
	file := fa.fset.File(funcDecl.Pos())
	if file == nil {
		return metrics.LineMetrics{}
	}

	return fa.countLinesInRange(file, start.Line+1, end.Line-1)
}

// countLinesInRange performs precise line counting between start and end lines (inclusive)
func (fa *FunctionAnalyzer) countLinesInRange(file *token.File, startLine, endLine int) metrics.LineMetrics {
	if startLine > endLine {
		return metrics.LineMetrics{}
	}

	// Read the source file
	src, err := os.ReadFile(file.Name())
	if err != nil {
		return metrics.LineMetrics{}
	}

	lines := strings.Split(string(src), "\n")
	if startLine < 1 || endLine > len(lines) {
		return metrics.LineMetrics{}
	}

	var codeLines, commentLines, blankLines int

	// Track if we're inside a multi-line comment
	inBlockComment := false

	for i := startLine - 1; i < endLine && i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])

		// Check for blank lines first
		if line == "" {
			blankLines++
			continue
		}

		// Parse the line character by character to handle complex cases
		lineType := fa.classifyLine(line, &inBlockComment)

		switch lineType {
		case "code":
			codeLines++
		case "comment":
			commentLines++
		case "mixed":
			// Line has both code and comment - count as code
			codeLines++
		}
	}

	totalLines := codeLines + commentLines + blankLines

	return metrics.LineMetrics{
		Total:    totalLines,
		Code:     codeLines,
		Comments: commentLines,
		Blank:    blankLines,
	}
}

// classifyLine determines the type of a line (code, comment, or mixed)
func (fa *FunctionAnalyzer) classifyLine(line string, inBlockComment *bool) string {
	if line == "" {
		return "blank"
	}

	// Handle lines within existing block comments
	if *inBlockComment {
		return fa.classifyLineInBlockComment(line, inBlockComment)
	}

	// Check for new block comments
	if blockStartIdx := strings.Index(line, "/*"); blockStartIdx >= 0 {
		return fa.classifyLineWithBlockComment(line, blockStartIdx, inBlockComment)
	}

	// Check for line comments
	if lineCommentIdx := strings.Index(line, "//"); lineCommentIdx >= 0 {
		return fa.classifyLineWithLineComment(line, lineCommentIdx)
	}

	// No comments found, must be code
	return "code"
}

// classifyLineInBlockComment handles lines that are within an existing block comment
func (fa *FunctionAnalyzer) classifyLineInBlockComment(line string, inBlockComment *bool) string {
	if strings.Contains(line, "*/") {
		*inBlockComment = false
		return fa.checkCodeAfterBlockCommentEnd(line)
	}
	return "comment"
}

// checkCodeAfterBlockCommentEnd checks if there's code after a block comment ends
func (fa *FunctionAnalyzer) checkCodeAfterBlockCommentEnd(line string) string {
	endIdx := strings.Index(line, "*/") + 2
	if endIdx < len(line) {
		remaining := strings.TrimSpace(line[endIdx:])
		if remaining != "" && !strings.HasPrefix(remaining, "//") {
			return "mixed"
		}
	}
	return "comment"
}

// classifyLineWithBlockComment handles lines that contain block comment starts
func (fa *FunctionAnalyzer) classifyLineWithBlockComment(line string, blockStartIdx int, inBlockComment *bool) string {
	// Check if block comment ends on same line
	if blockEndIdx := strings.Index(line[blockStartIdx:], "*/"); blockEndIdx >= 0 {
		return fa.classifyLineWithCompleteBlockComment(line, blockStartIdx, blockEndIdx)
	}

	// Block comment starts but doesn't end
	*inBlockComment = true
	beforeBlock := strings.TrimSpace(line[:blockStartIdx])
	if beforeBlock != "" {
		return "mixed"
	}
	return "comment"
}

// classifyLineWithCompleteBlockComment handles lines with complete block comments
func (fa *FunctionAnalyzer) classifyLineWithCompleteBlockComment(line string, blockStartIdx, blockEndIdx int) string {
	beforeBlock := strings.TrimSpace(line[:blockStartIdx])
	afterBlock := strings.TrimSpace(line[blockStartIdx+blockEndIdx+2:])
	hasCodeBefore := beforeBlock != ""
	hasCodeAfter := afterBlock != "" && !strings.HasPrefix(afterBlock, "//")

	if hasCodeBefore || hasCodeAfter {
		return "mixed"
	}
	return "comment"
}

// classifyLineWithLineComment handles lines that contain line comments
func (fa *FunctionAnalyzer) classifyLineWithLineComment(line string, lineCommentIdx int) string {
	beforeComment := strings.TrimSpace(line[:lineCommentIdx])
	if beforeComment != "" {
		return "mixed"
	}
	return "comment"
}

// calculateComplexity calculates various complexity metrics
func (fa *FunctionAnalyzer) calculateComplexity(funcDecl *ast.FuncDecl) metrics.ComplexityScore {
	if funcDecl.Body == nil {
		return metrics.ComplexityScore{}
	}

	complexity := metrics.ComplexityScore{
		Cyclomatic:   fa.calculateCyclomaticComplexity(funcDecl.Body),
		NestingDepth: fa.calculateNestingDepth(funcDecl.Body),
	}

	// Calculate cognitive complexity (simplified for now)
	complexity.Cognitive = complexity.Cyclomatic

	// Calculate overall complexity score
	complexity.Overall = float64(complexity.Cyclomatic) +
		float64(complexity.NestingDepth)*0.5 +
		float64(complexity.Cognitive)*0.3

	return complexity
}

// calculateCyclomaticComplexity calculates cyclomatic complexity
func (fa *FunctionAnalyzer) calculateCyclomaticComplexity(block *ast.BlockStmt) int {
	complexity := 1 // Base complexity

	ast.Inspect(block, func(n ast.Node) bool {
		switch n.(type) {
		case *ast.IfStmt, *ast.ForStmt, *ast.RangeStmt, *ast.SwitchStmt,
			*ast.TypeSwitchStmt, *ast.SelectStmt:
			complexity++
		case *ast.CaseClause:
			// Each case adds complexity, but we'll count the switch itself
		case *ast.CommClause:
			// Each select case adds complexity
			complexity++
		}
		return true
	})

	return complexity
}

// calculateNestingDepth calculates maximum nesting depth
func (fa *FunctionAnalyzer) calculateNestingDepth(block *ast.BlockStmt) int {
	maxDepth := 0
	currentDepth := 0

	ast.Inspect(block, func(n ast.Node) bool {
		switch n.(type) {
		case *ast.BlockStmt:
			currentDepth++
			if currentDepth > maxDepth {
				maxDepth = currentDepth
			}
		}
		return true
	})

	return maxDepth
}

// analyzeDocumentation analyzes function documentation
func (fa *FunctionAnalyzer) analyzeDocumentation(doc *ast.CommentGroup) metrics.DocumentationInfo {
	if doc == nil {
		return metrics.DocumentationInfo{
			HasComment: false,
		}
	}

	docText := doc.Text()

	info := metrics.DocumentationInfo{
		HasComment:    true,
		CommentLength: len(docText),
		HasExample:    strings.Contains(strings.ToLower(docText), "example"),
	}

	// Calculate quality score based on length and content
	info.QualityScore = fa.calculateDocQualityScore(docText)

	return info
}

// calculateDocQualityScore calculates documentation quality score
func (fa *FunctionAnalyzer) calculateDocQualityScore(docText string) float64 {
	if len(docText) == 0 {
		return 0.0
	}

	score := 0.0

	// Length score (diminishing returns)
	lengthScore := float64(len(docText)) / 100.0
	if lengthScore > 1.0 {
		lengthScore = 1.0
	}
	score += lengthScore * 0.4

	// Content quality indicators
	lowerDoc := strings.ToLower(docText)

	if strings.Contains(lowerDoc, "example") {
		score += 0.2
	}
	if strings.Contains(lowerDoc, "param") || strings.Contains(lowerDoc, "argument") {
		score += 0.1
	}
	if strings.Contains(lowerDoc, "return") {
		score += 0.1
	}
	if strings.Contains(lowerDoc, "error") {
		score += 0.1
	}
	if strings.Contains(lowerDoc, "note") || strings.Contains(lowerDoc, "warning") {
		score += 0.1
	}

	if score > 1.0 {
		score = 1.0
	}

	return score
}

// calculateSignatureComplexity calculates function signature complexity
func (fa *FunctionAnalyzer) calculateSignatureComplexity(sig metrics.FunctionSignature) float64 {
	complexity := 0.0

	// Parameter count contributes to complexity
	complexity += float64(sig.ParameterCount) * 0.5

	// Return count contributes to complexity
	complexity += float64(sig.ReturnCount) * 0.3

	// Interface parameters increase complexity
	complexity += float64(sig.InterfaceParams) * 0.8

	// Variadic parameters increase complexity
	if sig.VariadicUsage {
		complexity += 1.0
	}

	// Generic parameters increase complexity
	complexity += float64(len(sig.GenericParams)) * 1.5

	return complexity
}

// isInterfaceType checks if a type is an interface
func (fa *FunctionAnalyzer) isInterfaceType(expr ast.Expr) bool {
	switch t := expr.(type) {
	case *ast.InterfaceType:
		return true
	case *ast.Ident:
		// Common interface types
		return t.Name == "interface{}" || t.Name == "any"
	}
	return false
}

// isErrorType checks if a type is the error interface
func (fa *FunctionAnalyzer) isErrorType(expr ast.Expr) bool {
	if ident, ok := expr.(*ast.Ident); ok {
		return ident.Name == "error"
	}
	return false
}

// extractConstraints extracts type constraints from generic parameters
func (fa *FunctionAnalyzer) extractConstraints(expr ast.Expr) []string {
	var constraints []string

	// This is a simplified implementation
	// In practice, you'd need to handle more complex constraint expressions
	if ident, ok := expr.(*ast.Ident); ok {
		constraints = append(constraints, ident.Name)
	}

	return constraints
}
