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

// NewFunctionAnalyzer creates a new function analyzer for computing comprehensive function-level
// metrics including cyclomatic complexity, cognitive complexity, line counts (code vs. comments),
// parameter/return counts, and signature analysis. Essential for identifying complex functions
// that exceed maintainability thresholds and require refactoring.
func NewFunctionAnalyzer(fset *token.FileSet) *FunctionAnalyzer {
	return &FunctionAnalyzer{
		fset: fset,
	}
}

// AnalyzeFunctions analyzes all functions in an AST file and returns metrics.
func (fa *FunctionAnalyzer) AnalyzeFunctions(file *ast.File, pkgName string) ([]metrics.FunctionMetrics, error) {
	return fa.AnalyzeFunctionsWithPath(file, pkgName, file.Name.Name)
}

// AnalyzeFunctionsWithPath analyzes all functions in an AST file with explicit file path for
// accurate source location reporting. It computes comprehensive metrics for each function including
// complexity (cyclomatic, cognitive), line counts (excluding comments/blanks), parameter/return counts,
// and documentation presence. Returns a slice of function metrics used for threshold enforcement.
func (fa *FunctionAnalyzer) AnalyzeFunctionsWithPath(file *ast.File, pkgName, filePath string) ([]metrics.FunctionMetrics, error) {
	var functions []metrics.FunctionMetrics

	// Analyze top-level functions
	for _, decl := range file.Decls {
		if funcDecl, ok := decl.(*ast.FuncDecl); ok {
			function, err := fa.analyzeFunction(funcDecl, filePath, pkgName)
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

	lines, ok := fa.readFileLines(file.Name(), startLine, endLine)
	if !ok {
		return metrics.LineMetrics{}
	}

	codeLines, commentLines, blankLines := fa.classifyLines(lines, startLine, endLine)
	totalLines := codeLines + commentLines + blankLines

	return metrics.LineMetrics{
		Total:    totalLines,
		Code:     codeLines,
		Comments: commentLines,
		Blank:    blankLines,
	}
}

// readFileLines loads and splits a file into lines for the specified line range.
func (fa *FunctionAnalyzer) readFileLines(fileName string, startLine, endLine int) ([]string, bool) {
	src, err := os.ReadFile(fileName)
	if err != nil {
		return nil, false
	}

	lines := strings.Split(string(src), "\n")
	if startLine < 1 || endLine > len(lines) {
		return nil, false
	}

	return lines, true
}

// classifyLines categorizes each line as code, comment, or blank within the given range.
func (fa *FunctionAnalyzer) classifyLines(lines []string, startLine, endLine int) (codeLines, commentLines, blankLines int) {
	inBlockComment := false

	for i := startLine - 1; i < endLine && i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])

		if line == "" {
			blankLines++
			continue
		}

		lineType := fa.classifyLine(line, &inBlockComment)

		switch lineType {
		case "code":
			codeLines++
		case "comment":
			commentLines++
		case "mixed":
			codeLines++
		}
	}

	return codeLines, commentLines, blankLines
}

// findCommentOutsideStrings finds the index of a comment marker (// or /*)
// that is not inside a string literal. Returns -1 if not found outside strings.
// Handles both double-quoted strings with escapes and backtick raw strings.
func (fa *FunctionAnalyzer) findCommentOutsideStrings(line, commentMarker string) int {
	inDoubleQuote := false
	inBacktick := false
	escaped := false

	for i := 0; i < len(line); i++ {
		ch := line[i]

		if inDoubleQuote {
			inDoubleQuote, escaped = fa.processDoubleQuoteChar(ch, escaped)
			continue
		}

		if inBacktick {
			inBacktick = fa.processBacktickChar(ch)
			continue
		}

		if newQuoteState, insideQuote := fa.checkStringStart(ch); insideQuote {
			inDoubleQuote, inBacktick = newQuoteState[0], newQuoteState[1]
			continue
		}

		if fa.matchesCommentMarker(line, i, commentMarker) {
			return i
		}
	}

	return -1
}

// processDoubleQuoteChar handles a character inside a double-quoted string
func (fa *FunctionAnalyzer) processDoubleQuoteChar(ch byte, escaped bool) (stillInQuote, newEscaped bool) {
	if escaped {
		return true, false
	}
	if ch == '\\' {
		return true, true
	}
	if ch == '"' {
		return false, false
	}
	return true, false
}

// processBacktickChar handles a character inside a backtick raw string
func (fa *FunctionAnalyzer) processBacktickChar(ch byte) bool {
	return ch != '`'
}

// checkStringStart checks if the character starts a string literal
func (fa *FunctionAnalyzer) checkStringStart(ch byte) (quoteState [2]bool, isStringStart bool) {
	if ch == '"' {
		return [2]bool{true, false}, true
	}
	if ch == '`' {
		return [2]bool{false, true}, true
	}
	return [2]bool{false, false}, false
}

// matchesCommentMarker checks if the position starts with the comment marker
func (fa *FunctionAnalyzer) matchesCommentMarker(line string, i int, commentMarker string) bool {
	return i+len(commentMarker) <= len(line) && line[i:i+len(commentMarker)] == commentMarker
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

	// Check for new block comments (outside of string literals)
	blockStartIdx := fa.findCommentOutsideStrings(line, "/*")
	if blockStartIdx >= 0 {
		return fa.classifyLineWithBlockComment(line, blockStartIdx, inBlockComment)
	}

	// Check for line comments (outside of string literals)
	lineCommentIdx := fa.findCommentOutsideStrings(line, "//")
	if lineCommentIdx >= 0 {
		return fa.classifyLineWithLineComment(line, lineCommentIdx)
	}

	// No comments found, must be code
	return "code"
}

// classifyLineInBlockComment handles lines that are within an existing block comment
func (fa *FunctionAnalyzer) classifyLineInBlockComment(line string, inBlockComment *bool) string {
	// Handle nested comments when already in a block comment
	endIdx, endsOnThisLine := fa.findBlockCommentEndFromWithin(line)

	if endsOnThisLine {
		*inBlockComment = false
		return fa.checkCodeAfterBlockCommentEnd(line, endIdx)
	}
	return "comment"
}

// findBlockCommentEndFromWithin finds the end of a block comment when already inside one
// This handles nested comments properly by tracking depth starting from 1
func (fa *FunctionAnalyzer) findBlockCommentEndFromWithin(line string) (int, bool) {
	depth := 1 // We're already inside a comment
	i := 0

	for i < len(line)-1 {
		if i < len(line)-1 && line[i] == '/' && line[i+1] == '*' {
			depth++
			i += 2
		} else if i < len(line)-1 && line[i] == '*' && line[i+1] == '/' {
			depth--
			if depth == 0 {
				return i + 2, true // Found the matching end
			}
			i += 2
		} else {
			i++
		}
	}

	// No matching end found on this line
	return len(line), false
}

// checkCodeAfterBlockCommentEnd checks if there's code after a block comment ends
func (fa *FunctionAnalyzer) checkCodeAfterBlockCommentEnd(line string, endIdx int) string {
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
	// Use proper nested comment parsing to find the real end
	endIdx, endsOnThisLine := fa.findBlockCommentEnd(line, blockStartIdx)

	if endsOnThisLine {
		return fa.classifyLineWithCompleteBlockComment(line, blockStartIdx, endIdx-blockStartIdx-2)
	}

	// Block comment starts but doesn't end on this line
	*inBlockComment = true
	beforeBlock := strings.TrimSpace(line[:blockStartIdx])
	if beforeBlock != "" {
		return "mixed"
	}
	return "comment"
}

// findBlockCommentEnd finds the correct end of a block comment, handling nested comments
// Returns the absolute position of the end and whether it ends on this line
func (fa *FunctionAnalyzer) findBlockCommentEnd(line string, startIdx int) (int, bool) {
	depth := 1
	i := startIdx + 2 // Skip the initial /*

	for i < len(line)-1 {
		if i < len(line)-1 && line[i] == '/' && line[i+1] == '*' {
			depth++
			i += 2
		} else if i < len(line)-1 && line[i] == '*' && line[i+1] == '/' {
			depth--
			if depth == 0 {
				return i + 2, true // Found the matching end
			}
			i += 2
		} else {
			i++
		}
	}

	// No matching end found on this line
	return len(line), false
}

// classifyLineWithCompleteBlockComment handles lines with complete block comments
// blockContentLength is the length of the comment content (excluding /* and */)
func (fa *FunctionAnalyzer) classifyLineWithCompleteBlockComment(line string, blockStartIdx, blockContentLength int) string {
	beforeBlock := strings.TrimSpace(line[:blockStartIdx])
	afterBlock := strings.TrimSpace(line[blockStartIdx+blockContentLength+2:])
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

	// Use a recursive approach to properly track depth
	fa.walkForNestingDepth(block, 0, &maxDepth)

	return maxDepth
}

// walkForNestingDepth recursively walks the AST to calculate nesting depth
func (fa *FunctionAnalyzer) walkForNestingDepth(node ast.Node, currentDepth int, maxDepth *int) {
	if node == nil {
		return
	}

	switch n := node.(type) {
	case *ast.IfStmt:
		fa.walkIfStmtNesting(n, currentDepth, maxDepth)
	case *ast.ForStmt:
		fa.walkForStmtNesting(n, currentDepth, maxDepth)
	case *ast.RangeStmt:
		fa.walkRangeStmtNesting(n, currentDepth, maxDepth)
	case *ast.SwitchStmt:
		fa.walkSwitchStmtNesting(n, currentDepth, maxDepth)
	case *ast.TypeSwitchStmt:
		fa.walkTypeSwitchStmtNesting(n, currentDepth, maxDepth)
	case *ast.SelectStmt:
		fa.walkSelectStmtNesting(n, currentDepth, maxDepth)
	case *ast.BlockStmt:
		fa.walkBlockStmtNesting(n, currentDepth, maxDepth)
	default:
		fa.walkDefaultNodeNesting(node, currentDepth, maxDepth)
	}
}

// walkIfStmtNesting processes if statements for nesting depth tracking
func (fa *FunctionAnalyzer) walkIfStmtNesting(n *ast.IfStmt, currentDepth int, maxDepth *int) {
	newDepth := currentDepth + 1
	fa.updateMaxNestingDepth(newDepth, maxDepth)
	fa.walkForNestingDepth(n.Body, newDepth, maxDepth)
	if n.Else != nil {
		fa.walkForNestingDepth(n.Else, newDepth, maxDepth)
	}
}

// walkForStmtNesting processes for statements for nesting depth tracking
func (fa *FunctionAnalyzer) walkForStmtNesting(n *ast.ForStmt, currentDepth int, maxDepth *int) {
	newDepth := currentDepth + 1
	fa.updateMaxNestingDepth(newDepth, maxDepth)
	fa.walkForNestingDepth(n.Body, newDepth, maxDepth)
}

// walkRangeStmtNesting processes range statements for nesting depth tracking
func (fa *FunctionAnalyzer) walkRangeStmtNesting(n *ast.RangeStmt, currentDepth int, maxDepth *int) {
	newDepth := currentDepth + 1
	fa.updateMaxNestingDepth(newDepth, maxDepth)
	fa.walkForNestingDepth(n.Body, newDepth, maxDepth)
}

// walkSwitchStmtNesting processes switch statements for nesting depth tracking
func (fa *FunctionAnalyzer) walkSwitchStmtNesting(n *ast.SwitchStmt, currentDepth int, maxDepth *int) {
	newDepth := currentDepth + 1
	fa.updateMaxNestingDepth(newDepth, maxDepth)
	fa.walkForNestingDepth(n.Body, newDepth, maxDepth)
}

// walkTypeSwitchStmtNesting processes type switch statements for nesting depth tracking
func (fa *FunctionAnalyzer) walkTypeSwitchStmtNesting(n *ast.TypeSwitchStmt, currentDepth int, maxDepth *int) {
	newDepth := currentDepth + 1
	fa.updateMaxNestingDepth(newDepth, maxDepth)
	fa.walkForNestingDepth(n.Body, newDepth, maxDepth)
}

// walkSelectStmtNesting processes select statements for nesting depth tracking
func (fa *FunctionAnalyzer) walkSelectStmtNesting(n *ast.SelectStmt, currentDepth int, maxDepth *int) {
	newDepth := currentDepth + 1
	fa.updateMaxNestingDepth(newDepth, maxDepth)
	fa.walkForNestingDepth(n.Body, newDepth, maxDepth)
}

// walkBlockStmtNesting processes block statements for nesting depth tracking
func (fa *FunctionAnalyzer) walkBlockStmtNesting(n *ast.BlockStmt, currentDepth int, maxDepth *int) {
	for _, stmt := range n.List {
		fa.walkForNestingDepth(stmt, currentDepth, maxDepth)
	}
}

// walkDefaultNodeNesting handles other node types without changing depth
func (fa *FunctionAnalyzer) walkDefaultNodeNesting(node ast.Node, currentDepth int, maxDepth *int) {
	ast.Inspect(node, func(child ast.Node) bool {
		if child != node {
			fa.walkForNestingDepth(child, currentDepth, maxDepth)
			return false
		}
		return true
	})
}

// updateMaxNestingDepth updates the maximum depth if a new maximum is found
func (fa *FunctionAnalyzer) updateMaxNestingDepth(newDepth int, maxDepth *int) {
	if newDepth > *maxDepth {
		*maxDepth = newDepth
	}
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

	score := calculateLengthScore(docText)
	score += calculateContentQualityScore(docText)

	return normalizeScore(score)
}

// calculateLengthScore calculates score based on documentation length
func calculateLengthScore(docText string) float64 {
	lengthScore := float64(len(docText)) / 100.0
	if lengthScore > 1.0 {
		lengthScore = 1.0
	}
	return lengthScore * 0.4
}

// calculateContentQualityScore calculates score based on content quality indicators
func calculateContentQualityScore(docText string) float64 {
	lowerDoc := strings.ToLower(docText)
	score := 0.0

	qualityIndicators := []struct {
		keywords []string
		points   float64
	}{
		{[]string{"example"}, 0.2},
		{[]string{"param", "argument"}, 0.1},
		{[]string{"return"}, 0.1},
		{[]string{"error"}, 0.1},
		{[]string{"note", "warning"}, 0.1},
	}

	for _, indicator := range qualityIndicators {
		if containsAnyKeyword(lowerDoc, indicator.keywords) {
			score += indicator.points
		}
	}

	return score
}

// containsAnyKeyword checks if text contains any of the specified keywords
func containsAnyKeyword(text string, keywords []string) bool {
	for _, keyword := range keywords {
		if strings.Contains(text, keyword) {
			return true
		}
	}
	return false
}

// normalizeScore ensures score is within valid range [0,1]
func normalizeScore(score float64) float64 {
	if score > 1.0 {
		return 1.0
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
