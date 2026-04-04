package analyzer

import (
	"fmt"
	"go/ast"
	"go/token"
	"strings"

	"github.com/opd-ai/go-stats-generator/internal/metrics"
)

// AntipatternAnalyzer detects performance anti-patterns in Go code
type AntipatternAnalyzer struct {
	fset *token.FileSet
}

// NewAntipatternAnalyzer creates a new performance anti-pattern analyzer for detecting
// inefficient code patterns in Go source code. The analyzer identifies common performance
// issues including excessive memory allocations, string concatenation in loops, goroutine leaks,
// and improper resource management. Essential for code quality assessment and optimization.
func NewAntipatternAnalyzer(fset *token.FileSet) *AntipatternAnalyzer {
	return &AntipatternAnalyzer{fset: fset}
}

// Analyze detects performance anti-patterns in a Go source file by inspecting the AST for
// common inefficiencies. It checks for memory allocation issues (append without capacity),
// string concatenation in loops, goroutine leaks (missing channel closes), and resource
// management problems (defer in loops, unclosed resources). Returns a list of detected patterns.
func (a *AntipatternAnalyzer) Analyze(file *ast.File) []metrics.PerformanceAntipattern {
	var patterns []metrics.PerformanceAntipattern

	// Check if this is library code (non-main package)
	isLibraryCode := file.Name != nil && file.Name.Name != "main"

	for _, decl := range file.Decls {
		funcDecl, ok := decl.(*ast.FuncDecl)
		if !ok || funcDecl.Body == nil {
			continue
		}
		patterns = append(patterns, a.analyzeFunction(funcDecl)...)
		patterns = append(patterns, a.checkAnyOveruse(funcDecl)...)
		patterns = append(patterns, a.checkInitFunctionComplexity(funcDecl)...)
		patterns = append(patterns, a.checkNakedReturnInLongFunction(funcDecl)...)
		patterns = append(patterns, a.checkPanicInLibraryCode(funcDecl, isLibraryCode)...)
		patterns = append(patterns, a.checkGiantBranchingChains(funcDecl)...)
		patterns = append(patterns, a.checkUnusedReceiverName(funcDecl)...)
	}

	return patterns
}

// analyzeFunction analyzes a single function declaration for anti-patterns,
// using the function body to provide parent-node context for loop detection.
func (a *AntipatternAnalyzer) analyzeFunction(funcDecl *ast.FuncDecl) []metrics.PerformanceAntipattern {
	var patterns []metrics.PerformanceAntipattern
	a.walkWithLoopContext(funcDecl.Body, false, funcDecl.Body, &patterns)
	return patterns
}

// walkWithLoopContext traverses the AST tracking whether we are inside a loop,
// enabling accurate detection of anti-patterns that only apply within loops.
func (a *AntipatternAnalyzer) walkWithLoopContext(node ast.Node, inLoop bool, funcBody *ast.BlockStmt, patterns *[]metrics.PerformanceAntipattern) {
	if node == nil {
		return
	}

	switch n := node.(type) {
	case *ast.ForStmt:
		a.handleForLoop(n, funcBody, patterns)
		return
	case *ast.RangeStmt:
		a.handleRangeLoop(n, funcBody, patterns)
		return
	case *ast.GoStmt:
		a.checkGoroutineLeak(n, patterns)
	case *ast.CallExpr:
		a.checkResourceLeak(n, funcBody, patterns)
	case *ast.AssignStmt:
		if inLoop {
			a.checkAssignForLoopAntipatterns(n, patterns)
		}
	case *ast.BinaryExpr:
		if inLoop {
			a.checkStringConcatInLoop(n, patterns)
		}
	case *ast.IfStmt:
		a.checkBareErrorReturn(n, patterns)
	}

	a.recurseIntoChildren(node, inLoop, funcBody, patterns)
}

func (a *AntipatternAnalyzer) handleForLoop(n *ast.ForStmt, funcBody *ast.BlockStmt, patterns *[]metrics.PerformanceAntipattern) {
	if n.Body != nil {
		for _, stmt := range n.Body.List {
			a.walkWithLoopContext(stmt, true, funcBody, patterns)
		}
	}
}

func (a *AntipatternAnalyzer) handleRangeLoop(n *ast.RangeStmt, funcBody *ast.BlockStmt, patterns *[]metrics.PerformanceAntipattern) {
	if n.Body != nil {
		for _, stmt := range n.Body.List {
			a.walkWithLoopContext(stmt, true, funcBody, patterns)
		}
	}
}

func (a *AntipatternAnalyzer) checkGoroutineLeak(n *ast.GoStmt, patterns *[]metrics.PerformanceAntipattern) {
	if !a.hasContextOrDone(n) {
		*patterns = append(*patterns, metrics.PerformanceAntipattern{
			Type:        "goroutine_leak",
			Description: "Goroutine without context or done channel",
			Severity:    metrics.SeverityLevelViolation,
			File:        a.fset.Position(n.Pos()).Filename,
			Line:        a.fset.Position(n.Pos()).Line,
			Suggestion:  "Add context.Context or done channel for graceful shutdown",
		})
	}
}

func (a *AntipatternAnalyzer) checkResourceLeak(n *ast.CallExpr, funcBody *ast.BlockStmt, patterns *[]metrics.PerformanceAntipattern) {
	if a.isResourceAcquisition(n) && !a.hasDeferClose(n, funcBody) {
		*patterns = append(*patterns, metrics.PerformanceAntipattern{
			Type:        "resource_leak",
			Description: "Resource acquisition without defer close",
			Severity:    metrics.SeverityLevelCritical,
			File:        a.fset.Position(n.Pos()).Filename,
			Line:        a.fset.Position(n.Pos()).Line,
			Suggestion:  "Use defer to ensure resource cleanup",
		})
	}
}

func (a *AntipatternAnalyzer) checkStringConcatInLoop(n *ast.BinaryExpr, patterns *[]metrics.PerformanceAntipattern) {
	if n.Op == token.ADD {
		if a.isStringType(n.X) || a.isStringType(n.Y) {
			*patterns = append(*patterns, metrics.PerformanceAntipattern{
				Type:        "string_concatenation",
				Description: "String concatenation in loop",
				Severity:    metrics.SeverityLevelViolation,
				File:        a.fset.Position(n.Pos()).Filename,
				Line:        a.fset.Position(n.Pos()).Line,
				Suggestion:  "Use strings.Builder for efficient concatenation",
			})
		}
	}
}

func (a *AntipatternAnalyzer) recurseIntoChildren(node ast.Node, inLoop bool, funcBody *ast.BlockStmt, patterns *[]metrics.PerformanceAntipattern) {
	ast.Inspect(node, func(child ast.Node) bool {
		if child == node {
			return true
		}
		switch child.(type) {
		case *ast.ForStmt, *ast.RangeStmt, *ast.GoStmt, *ast.CallExpr, *ast.AssignStmt, *ast.BinaryExpr, *ast.IfStmt:
			a.walkWithLoopContext(child, inLoop, funcBody, patterns)
			return false
		}
		return true
	})
}

// checkAssignForLoopAntipatterns checks assignment statements inside loops for append without pre-allocation
func (a *AntipatternAnalyzer) checkAssignForLoopAntipatterns(node *ast.AssignStmt, patterns *[]metrics.PerformanceAntipattern) {
	for _, expr := range node.Rhs {
		if call, ok := expr.(*ast.CallExpr); ok {
			if a.isAppendCall(call) {
				*patterns = append(*patterns, metrics.PerformanceAntipattern{
					Type:        "memory_allocation",
					Description: "append() in loop without pre-allocation",
					Severity:    metrics.SeverityLevelWarning,
					File:        a.fset.Position(node.Pos()).Filename,
					Line:        a.fset.Position(node.Pos()).Line,
					Suggestion:  "Pre-allocate slice with make() for known capacity",
				})
			}
		}
	}
}

// isAppendCall checks if a call expression is an append() call
func (a *AntipatternAnalyzer) isAppendCall(call *ast.CallExpr) bool {
	if ident, ok := call.Fun.(*ast.Ident); ok {
		return ident.Name == "append"
	}
	return false
}

// isAppendInLoop checks if append is called inside a loop (kept for backward compatibility)
func (a *AntipatternAnalyzer) isAppendInLoop(call *ast.CallExpr, node ast.Node) bool {
	if ident, ok := call.Fun.(*ast.Ident); ok {
		return ident.Name == "append"
	}
	return false
}

// isStringType checks if expression is string type
func (a *AntipatternAnalyzer) isStringType(expr ast.Expr) bool {
	if lit, ok := expr.(*ast.BasicLit); ok {
		return lit.Kind == token.STRING
	}
	return false
}

// isInLoop checks if node is inside a loop (kept for backward compatibility)
func (a *AntipatternAnalyzer) isInLoop(node ast.Node) bool {
	return false
}

// hasContextOrDone checks if goroutine has context or done channel by examining
// parameter names and types of the launched function.
func (a *AntipatternAnalyzer) hasContextOrDone(goStmt *ast.GoStmt) bool {
	funcLit, ok := goStmt.Call.Fun.(*ast.FuncLit)
	if !ok {
		return a.hasContextOrDoneInArgs(goStmt.Call)
	}

	// Check parameter names
	if funcLit.Type.Params != nil {
		for _, param := range funcLit.Type.Params.List {
			if a.isContextOrDoneParam(param) {
				return true
			}
		}
	}

	// Check if the function body references context or done channels via closure
	if a.bodyReferencesContextOrDone(funcLit.Body) {
		return true
	}

	return false
}

// isContextOrDoneParam checks if a parameter represents a context or done channel
func (a *AntipatternAnalyzer) isContextOrDoneParam(param *ast.Field) bool {
	if a.hasContextOrDoneParamName(param) {
		return true
	}
	if a.isContextType(param) {
		return true
	}
	if a.isChannelType(param) {
		return true
	}
	return false
}

// hasContextOrDoneParamName checks if parameter name suggests context/cancellation usage.
func (a *AntipatternAnalyzer) hasContextOrDoneParamName(param *ast.Field) bool {
	contextKeywords := []string{"ctx", "done", "cancel", "quit", "stop"}
	for _, name := range param.Names {
		nameLower := strings.ToLower(name.Name)
		for _, keyword := range contextKeywords {
			if strings.Contains(nameLower, keyword) {
				return true
			}
		}
	}
	return false
}

// isContextType checks if parameter type is context.Context.
func (a *AntipatternAnalyzer) isContextType(param *ast.Field) bool {
	sel, ok := param.Type.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	ident, ok := sel.X.(*ast.Ident)
	if !ok {
		return false
	}
	return ident.Name == "context" && sel.Sel.Name == "Context"
}

// isChannelType checks if parameter is a channel (often used for done signals).
func (a *AntipatternAnalyzer) isChannelType(param *ast.Field) bool {
	_, ok := param.Type.(*ast.ChanType)
	return ok
}

// hasContextOrDoneInArgs checks if a named function call includes context arguments
func (a *AntipatternAnalyzer) hasContextOrDoneInArgs(call *ast.CallExpr) bool {
	for _, arg := range call.Args {
		if ident, ok := arg.(*ast.Ident); ok {
			nameLower := strings.ToLower(ident.Name)
			if strings.Contains(nameLower, "ctx") || strings.Contains(nameLower, "done") ||
				strings.Contains(nameLower, "cancel") || strings.Contains(nameLower, "quit") ||
				strings.Contains(nameLower, "stop") {
				return true
			}
		}
	}
	return false
}

// bodyReferencesContextOrDone checks if a function body uses context or done channel variables
func (a *AntipatternAnalyzer) bodyReferencesContextOrDone(body *ast.BlockStmt) bool {
	if body == nil {
		return false
	}
	found := false
	ast.Inspect(body, func(n ast.Node) bool {
		if found {
			return false
		}
		switch node := n.(type) {
		case *ast.Ident:
			nameLower := strings.ToLower(node.Name)
			if strings.Contains(nameLower, "ctx") || strings.Contains(nameLower, "done") ||
				strings.Contains(nameLower, "cancel") || strings.Contains(nameLower, "quit") {
				found = true
				return false
			}
		case *ast.SelectStmt:
			found = true
			return false
		}
		return true
	})
	return found
}

// isResourceAcquisition checks if call acquires a resource
func (a *AntipatternAnalyzer) isResourceAcquisition(call *ast.CallExpr) bool {
	if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
		resourceFuncs := []string{"Open", "Create", "Dial", "Connect", "Acquire"}
		for _, fn := range resourceFuncs {
			if sel.Sel.Name == fn {
				return true
			}
		}
	}
	return false
}

// hasDeferClose checks if the enclosing function body contains a defer close/Close statement
// for the specific resource acquired by the given call expression.
func (a *AntipatternAnalyzer) hasDeferClose(call *ast.CallExpr, funcBody *ast.BlockStmt) bool {
	if funcBody == nil {
		return false
	}

	resourceVar := a.findAssignedVar(call, funcBody)
	found := false

	ast.Inspect(funcBody, func(n ast.Node) bool {
		if found {
			return false
		}
		deferStmt, ok := n.(*ast.DeferStmt)
		if !ok {
			return true
		}

		if a.checkDirectDeferCleanup(deferStmt, resourceVar) {
			found = true
			return false
		}

		if a.checkDeferFuncLitCleanup(deferStmt, resourceVar) {
			found = true
			return false
		}

		return true
	})

	return found
}

func (a *AntipatternAnalyzer) checkDirectDeferCleanup(deferStmt *ast.DeferStmt, resourceVar string) bool {
	sel, ok := deferStmt.Call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}

	if !a.isCleanupMethod(sel.Sel.Name) {
		return false
	}

	return a.matchesResourceVar(sel.X, resourceVar)
}

func (a *AntipatternAnalyzer) checkDeferFuncLitCleanup(deferStmt *ast.DeferStmt, resourceVar string) bool {
	funcLit, ok := deferStmt.Call.Fun.(*ast.FuncLit)
	if !ok {
		return false
	}

	found := false
	ast.Inspect(funcLit.Body, func(inner ast.Node) bool {
		innerCall, ok := inner.(*ast.CallExpr)
		if !ok {
			return true
		}

		innerSel, ok := innerCall.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}

		if innerSel.Sel.Name == "Close" || innerSel.Sel.Name == "Release" {
			if a.matchesResourceVar(innerSel.X, resourceVar) {
				found = true
				return false
			}
		}
		return true
	})

	return found
}

func (a *AntipatternAnalyzer) isCleanupMethod(methodName string) bool {
	cleanupFuncs := []string{"Close", "Release", "Disconnect", "Shutdown", "Stop", "Done"}
	for _, fn := range cleanupFuncs {
		if methodName == fn {
			return true
		}
	}
	return false
}

func (a *AntipatternAnalyzer) matchesResourceVar(expr ast.Expr, resourceVar string) bool {
	if resourceVar == "" {
		return true
	}
	ident, ok := expr.(*ast.Ident)
	return ok && ident.Name == resourceVar
}

// findAssignedVar finds the variable name that a call expression result is assigned to
// within the function body. Returns empty string if unable to determine.
func (a *AntipatternAnalyzer) findAssignedVar(call *ast.CallExpr, funcBody *ast.BlockStmt) string {
	var varName string
	ast.Inspect(funcBody, func(n ast.Node) bool {
		if varName != "" {
			return false
		}
		assign, ok := n.(*ast.AssignStmt)
		if !ok {
			return true
		}
		for _, rhs := range assign.Rhs {
			if rhs == call && len(assign.Lhs) > 0 {
				if ident, ok := assign.Lhs[0].(*ast.Ident); ok {
					varName = ident.Name
					return false
				}
			}
		}
		return true
	})
	return varName
}

// checkBareErrorReturn detects the pattern `if err != nil { return err }` without
// error wrapping using fmt.Errorf, errors.New, or custom error types. This is the
// most common LLM slop pattern in Go where generated code loses error context.
func (a *AntipatternAnalyzer) checkBareErrorReturn(ifStmt *ast.IfStmt, patterns *[]metrics.PerformanceAntipattern) {
	// Pattern: if err != nil { return err } or if err != nil { return nil, err }
	// First check the condition is err != nil or nil != err
	if !a.isErrorNilCheck(ifStmt.Cond) {
		return
	}

	// Check the body contains a return statement returning err
	if ifStmt.Body == nil || len(ifStmt.Body.List) == 0 {
		return
	}

	// Look for return statements in the if body
	for _, stmt := range ifStmt.Body.List {
		retStmt, ok := stmt.(*ast.ReturnStmt)
		if !ok {
			continue
		}

		// Check if any return value is a bare err identifier
		if a.hasBareErrorReturn(retStmt) {
			*patterns = append(*patterns, metrics.PerformanceAntipattern{
				Type:        "bare_error_return",
				Description: "Error returned without context wrapping",
				Severity:    metrics.SeverityLevelViolation,
				File:        a.fset.Position(ifStmt.Pos()).Filename,
				Line:        a.fset.Position(ifStmt.Pos()).Line,
				Suggestion:  "Wrap error with fmt.Errorf(\"context: %w\", err) to preserve error chain",
			})
			return
		}
	}
}

// isErrorNilCheck checks if condition is `err != nil` or `nil != err`
func (a *AntipatternAnalyzer) isErrorNilCheck(cond ast.Expr) bool {
	binExpr, ok := cond.(*ast.BinaryExpr)
	if !ok || binExpr.Op != token.NEQ {
		return false
	}

	// Check left side is err and right is nil, or vice versa
	leftErr := a.isErrIdentifier(binExpr.X)
	leftNil := a.isNilIdentifier(binExpr.X)
	rightErr := a.isErrIdentifier(binExpr.Y)
	rightNil := a.isNilIdentifier(binExpr.Y)

	return (leftErr && rightNil) || (leftNil && rightErr)
}

// isErrIdentifier checks if expression is an identifier named "err"
func (a *AntipatternAnalyzer) isErrIdentifier(expr ast.Expr) bool {
	ident, ok := expr.(*ast.Ident)
	return ok && ident.Name == "err"
}

// isNilIdentifier checks if expression is the nil identifier
func (a *AntipatternAnalyzer) isNilIdentifier(expr ast.Expr) bool {
	ident, ok := expr.(*ast.Ident)
	return ok && ident.Name == "nil"
}

// hasBareErrorReturn checks if return statement returns err without wrapping
func (a *AntipatternAnalyzer) hasBareErrorReturn(retStmt *ast.ReturnStmt) bool {
	for _, result := range retStmt.Results {
		// Skip if result is wrapped (CallExpr like fmt.Errorf)
		if _, ok := result.(*ast.CallExpr); ok {
			continue
		}

		// Check if it's a bare err identifier
		if a.isErrIdentifier(result) {
			return true
		}
	}
	return false
}

// checkAnyOveruse detects excessive use of interface{} or any in function signatures.
// Flags functions with multiple any/interface{} parameters or returns, suggesting more
// specific types. Excludes genuinely generic utility functions (single type parameter).
func (a *AntipatternAnalyzer) checkAnyOveruse(funcDecl *ast.FuncDecl) []metrics.PerformanceAntipattern {
	var patterns []metrics.PerformanceAntipattern

	if funcDecl.Type == nil {
		return patterns
	}

	anyCount, totalParams := a.countAnyUsage(funcDecl.Type)

	if a.exceedsAnyThreshold(anyCount, totalParams) {
		patterns = append(patterns, metrics.PerformanceAntipattern{
			Type:        "any_overuse",
			Description: "Excessive use of interface{}/any in function signature",
			Severity:    metrics.SeverityLevelWarning,
			File:        a.fset.Position(funcDecl.Pos()).Filename,
			Line:        a.fset.Position(funcDecl.Pos()).Line,
			Suggestion:  "Use concrete types or constrained generics instead of interface{}/any for type safety",
		})
	}

	return patterns
}

// countAnyUsage counts any/interface{} parameters and returns in function type
func (a *AntipatternAnalyzer) countAnyUsage(funcType *ast.FuncType) (anyCount, totalParams int) {
	anyCount += a.countAnyInParams(funcType.Params)
	totalParams = a.countTotalParams(funcType.Params)
	anyCount += a.countAnyInResults(funcType.Results)
	return anyCount, totalParams
}

// countAnyInParams counts any/interface{} in parameter list
func (a *AntipatternAnalyzer) countAnyInParams(params *ast.FieldList) int {
	if params == nil {
		return 0
	}

	count := 0
	for _, param := range params.List {
		if a.isAnyType(param.Type) {
			paramCount := len(param.Names)
			if paramCount == 0 {
				paramCount = 1
			}
			count += paramCount
		}
	}
	return count
}

// countTotalParams counts total parameters in parameter list
func (a *AntipatternAnalyzer) countTotalParams(params *ast.FieldList) int {
	if params == nil {
		return 0
	}

	total := 0
	for _, param := range params.List {
		paramCount := len(param.Names)
		if paramCount == 0 {
			paramCount = 1
		}
		total += paramCount
	}
	return total
}

// countAnyInResults counts any/interface{} in result list
func (a *AntipatternAnalyzer) countAnyInResults(results *ast.FieldList) int {
	if results == nil {
		return 0
	}

	count := 0
	for _, result := range results.List {
		if a.isAnyType(result.Type) {
			count++
		}
	}
	return count
}

// exceedsAnyThreshold checks if any usage exceeds acceptable threshold
func (a *AntipatternAnalyzer) exceedsAnyThreshold(anyCount, totalParams int) bool {
	// Flag if more than 2 any parameters
	if anyCount > 2 {
		return true
	}

	// Flag if >30% any ratio (excluding single-param generic utilities)
	if totalParams > 1 && anyCount > 0 && float64(anyCount)/float64(totalParams) > 0.3 {
		return true
	}

	return false
}

// isAnyType checks if the type is interface{} or any
func (a *AntipatternAnalyzer) isAnyType(expr ast.Expr) bool {
	// Check for "any" identifier (Go 1.18+)
	if ident, ok := expr.(*ast.Ident); ok && ident.Name == "any" {
		return true
	}

	// Check for empty interface{}
	ifaceType, ok := expr.(*ast.InterfaceType)
	if !ok {
		return false
	}

	// Empty interface has no methods
	return ifaceType.Methods == nil || len(ifaceType.Methods.List) == 0
}

// checkInitFunctionComplexity detects complex init() functions, which violate
// Go best practices. Init functions should be simple and focused on setup.
// Complex init() functions (high cyclomatic complexity) make code harder to test
// and understand, and can hide initialization bugs. This is a common LLM slop pattern.
func (a *AntipatternAnalyzer) checkInitFunctionComplexity(funcDecl *ast.FuncDecl) []metrics.PerformanceAntipattern {
	var patterns []metrics.PerformanceAntipattern

	// Check if this is an init() function
	if !a.isInitFunction(funcDecl) {
		return patterns
	}

	// Calculate cyclomatic complexity for the init function
	complexity := a.calculateCyclomaticComplexity(funcDecl.Body)

	// Flag if complexity exceeds threshold (default: 5)
	const maxInitComplexity = 5
	if complexity > maxInitComplexity {
		patterns = append(patterns, metrics.PerformanceAntipattern{
			Type:        "init_complexity",
			Description: "init() function has high cyclomatic complexity",
			Severity:    metrics.SeverityLevelWarning,
			File:        a.fset.Position(funcDecl.Pos()).Filename,
			Line:        a.fset.Position(funcDecl.Pos()).Line,
			Suggestion:  "Simplify init() function or move complex initialization to explicit functions",
		})
	}

	return patterns
}

// isInitFunction checks if function declaration is an init() function
func (a *AntipatternAnalyzer) isInitFunction(funcDecl *ast.FuncDecl) bool {
	return funcDecl.Name != nil && funcDecl.Name.Name == "init" &&
		funcDecl.Recv == nil && // Not a method
		(funcDecl.Type.Params == nil || len(funcDecl.Type.Params.List) == 0) && // No parameters
		(funcDecl.Type.Results == nil || len(funcDecl.Type.Results.List) == 0) // No returns
}

// calculateCyclomaticComplexity computes cyclomatic complexity for a statement block
func (a *AntipatternAnalyzer) calculateCyclomaticComplexity(body *ast.BlockStmt) int {
	if body == nil {
		return 1
	}

	complexity := 1 // Base complexity

	ast.Inspect(body, func(n ast.Node) bool {
		switch n.(type) {
		case *ast.IfStmt, *ast.ForStmt, *ast.RangeStmt, *ast.CaseClause, *ast.CommClause:
			complexity++
		case *ast.BinaryExpr:
			// Count logical operators (&&, ||)
			binExpr := n.(*ast.BinaryExpr)
			if binExpr.Op == token.LAND || binExpr.Op == token.LOR {
				complexity++
			}
		}
		return true
	})

	return complexity
}

// checkNakedReturnInLongFunction detects naked returns (return without explicit values)
// in functions exceeding a line threshold. Short functions with named returns are idiomatic
// Go; long functions with naked returns harm readability. This is a common LLM slop pattern
// where generated code uses named returns inappropriately in complex functions.
func (a *AntipatternAnalyzer) checkNakedReturnInLongFunction(funcDecl *ast.FuncDecl) []metrics.PerformanceAntipattern {
	var patterns []metrics.PerformanceAntipattern

	// Only check functions with named return parameters
	if !a.hasNamedReturns(funcDecl) {
		return patterns
	}

	// Check if function body exists
	if funcDecl.Body == nil {
		return patterns
	}

	// Count lines in function body
	lineCount := a.countFunctionLines(funcDecl)

	// Threshold: only flag if function > 10 lines
	const maxLinesForNakedReturn = 10
	if lineCount <= maxLinesForNakedReturn {
		return patterns
	}

	// Check for naked return statements
	if a.hasNakedReturn(funcDecl.Body) {
		patterns = append(patterns, metrics.PerformanceAntipattern{
			Type:        "naked_return_long_function",
			Description: "Naked return in long function with named returns",
			Severity:    metrics.SeverityLevelWarning,
			File:        a.fset.Position(funcDecl.Pos()).Filename,
			Line:        a.fset.Position(funcDecl.Pos()).Line,
			Suggestion:  "Use explicit return values in long functions to improve readability",
		})
	}

	return patterns
}

// hasNamedReturns checks if function has named return parameters
func (a *AntipatternAnalyzer) hasNamedReturns(funcDecl *ast.FuncDecl) bool {
	if funcDecl.Type == nil || funcDecl.Type.Results == nil {
		return false
	}

	for _, result := range funcDecl.Type.Results.List {
		// If any result has a name, it's a named return
		if len(result.Names) > 0 {
			return true
		}
	}

	return false
}

// countFunctionLines counts the number of lines in a function body
func (a *AntipatternAnalyzer) countFunctionLines(funcDecl *ast.FuncDecl) int {
	if funcDecl.Body == nil {
		return 0
	}

	start := a.fset.Position(funcDecl.Body.Lbrace)
	end := a.fset.Position(funcDecl.Body.Rbrace)

	// Return number of lines between braces (exclusive of braces themselves)
	return end.Line - start.Line - 1
}

// hasNakedReturn checks if function body contains any naked return statements
func (a *AntipatternAnalyzer) hasNakedReturn(body *ast.BlockStmt) bool {
	found := false

	ast.Inspect(body, func(n ast.Node) bool {
		if found {
			return false
		}

		retStmt, ok := n.(*ast.ReturnStmt)
		if !ok {
			return true
		}

		// Naked return: return statement with no explicit values
		if len(retStmt.Results) == 0 {
			found = true
			return false
		}

		return true
	})

	return found
}

// checkPanicInLibraryCode detects panic() and log.Fatal() calls in library code (non-main packages).
// Library code should return errors instead of terminating the process. The check excludes init()
// functions where panic() is acceptable for configuration validation during initialization.
func (a *AntipatternAnalyzer) checkPanicInLibraryCode(funcDecl *ast.FuncDecl, isLibraryCode bool) []metrics.PerformanceAntipattern {
	var patterns []metrics.PerformanceAntipattern

	// Skip if this is not library code (main package is OK to use panic/log.Fatal)
	if !isLibraryCode {
		return patterns
	}

	// Skip init() functions - panic is acceptable during initialization
	if a.isInitFunction(funcDecl) {
		return patterns
	}

	// Walk the function body looking for panic() or log.Fatal() calls
	ast.Inspect(funcDecl.Body, func(n ast.Node) bool {
		callExpr, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}

		if a.isPanicCall(callExpr) {
			patterns = append(patterns, metrics.PerformanceAntipattern{
				Type:        "panic_in_library",
				Description: "panic() call in library code (non-main package)",
				Severity:    metrics.SeverityLevelViolation,
				File:        a.fset.Position(callExpr.Pos()).Filename,
				Line:        a.fset.Position(callExpr.Pos()).Line,
				Suggestion:  "Return error instead of panic() - library code should not terminate the process",
			})
		} else if a.isLogFatalCall(callExpr) {
			patterns = append(patterns, metrics.PerformanceAntipattern{
				Type:        "log_fatal_in_library",
				Description: "log.Fatal() call in library code (non-main package)",
				Severity:    metrics.SeverityLevelCritical,
				File:        a.fset.Position(callExpr.Pos()).Filename,
				Line:        a.fset.Position(callExpr.Pos()).Line,
				Suggestion:  "Return error instead of log.Fatal() - library code should not terminate the process",
			})
		}

		return true
	})

	return patterns
}

// isPanicCall checks if a call expression is a panic() call
func (a *AntipatternAnalyzer) isPanicCall(callExpr *ast.CallExpr) bool {
	ident, ok := callExpr.Fun.(*ast.Ident)
	return ok && ident.Name == "panic"
}

// isLogFatalCall checks if a call expression is a log.Fatal() or log.Fatalf() call
func (a *AntipatternAnalyzer) isLogFatalCall(callExpr *ast.CallExpr) bool {
	sel, ok := callExpr.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}

	// Check if selector is log.Fatal, log.Fatalf, or log.Fatalln
	pkgIdent, ok := sel.X.(*ast.Ident)
	if !ok || pkgIdent.Name != "log" {
		return false
	}

	return sel.Sel.Name == "Fatal" || sel.Sel.Name == "Fatalf" || sel.Sel.Name == "Fatalln"
}

// checkGiantBranchingChains detects switch statements or if-else chains with excessive branches.
// Giant branching structures indicate poor control flow design and suggest refactoring to
// dispatch maps, strategy patterns, or polymorphic design. Configurable threshold defaults to 10 branches.
func (a *AntipatternAnalyzer) checkGiantBranchingChains(funcDecl *ast.FuncDecl) []metrics.PerformanceAntipattern {
	var patterns []metrics.PerformanceAntipattern
	const maxBranches = 10
	processedIfs := make(map[*ast.IfStmt]bool)

	ast.Inspect(funcDecl.Body, func(n ast.Node) bool {
		if p := a.checkBranchingNode(n, maxBranches, processedIfs); p != nil {
			patterns = append(patterns, *p)
		}
		return true
	})
	return patterns
}

// checkBranchingNode checks a single AST node for excessive branching
func (a *AntipatternAnalyzer) checkBranchingNode(n ast.Node, maxBranches int, processedIfs map[*ast.IfStmt]bool) *metrics.PerformanceAntipattern {
	switch stmt := n.(type) {
	case *ast.SwitchStmt:
		return a.checkSwitchBranches(stmt, maxBranches)
	case *ast.TypeSwitchStmt:
		return a.checkTypeSwitchBranches(stmt, maxBranches)
	case *ast.IfStmt:
		return a.checkIfElseChain(stmt, maxBranches, processedIfs)
	}
	return nil
}

// checkSwitchBranches checks if a switch statement has excessive branches
func (a *AntipatternAnalyzer) checkSwitchBranches(stmt *ast.SwitchStmt, maxBranches int) *metrics.PerformanceAntipattern {
	if a.countSwitchBranches(stmt) <= maxBranches {
		return nil
	}
	return a.createBranchPattern(stmt.Pos(), "giant_switch", "Switch statement with excessive branches",
		"Consider using a dispatch map or strategy pattern instead of giant switch statement")
}

// checkTypeSwitchBranches checks if a type switch statement has excessive branches
func (a *AntipatternAnalyzer) checkTypeSwitchBranches(stmt *ast.TypeSwitchStmt, maxBranches int) *metrics.PerformanceAntipattern {
	if a.countTypeSwitchBranches(stmt) <= maxBranches {
		return nil
	}
	return a.createBranchPattern(stmt.Pos(), "giant_type_switch", "Type switch statement with excessive branches",
		"Consider using interface methods or type-specific handlers instead of giant type switch")
}

// checkIfElseChain checks if an if-else chain has excessive branches
func (a *AntipatternAnalyzer) checkIfElseChain(stmt *ast.IfStmt, maxBranches int, processedIfs map[*ast.IfStmt]bool) *metrics.PerformanceAntipattern {
	if processedIfs[stmt] {
		return nil
	}
	if a.countIfElseChainLength(stmt) <= maxBranches {
		return nil
	}
	a.markIfElseChain(stmt, processedIfs)
	return a.createBranchPattern(stmt.Pos(), "giant_if_else_chain", "If-else chain with excessive branches",
		"Refactor to dispatch map, early returns, or extract condition checks into named functions")
}

// createBranchPattern creates a PerformanceAntipattern for branching issues
func (a *AntipatternAnalyzer) createBranchPattern(pos token.Pos, patType, desc, suggestion string) *metrics.PerformanceAntipattern {
	return &metrics.PerformanceAntipattern{
		Type:        patType,
		Description: desc,
		Severity:    metrics.SeverityLevelWarning,
		File:        a.fset.Position(pos).Filename,
		Line:        a.fset.Position(pos).Line,
		Suggestion:  suggestion,
	}
}

// markIfElseChain marks all if statements in an if-else chain as processed
func (a *AntipatternAnalyzer) markIfElseChain(ifStmt *ast.IfStmt, processedIfs map[*ast.IfStmt]bool) {
	current := ifStmt
	for current != nil {
		processedIfs[current] = true

		// Check if the else is another if statement
		nextIf, ok := current.Else.(*ast.IfStmt)
		if !ok {
			break
		}
		current = nextIf
	}
}

// countSwitchBranches counts the number of case clauses in a switch statement
func (a *AntipatternAnalyzer) countSwitchBranches(switchStmt *ast.SwitchStmt) int {
	if switchStmt.Body == nil {
		return 0
	}
	return len(switchStmt.Body.List)
}

// countTypeSwitchBranches counts the number of case clauses in a type switch statement
func (a *AntipatternAnalyzer) countTypeSwitchBranches(typeSwitchStmt *ast.TypeSwitchStmt) int {
	if typeSwitchStmt.Body == nil {
		return 0
	}
	return len(typeSwitchStmt.Body.List)
}

// countIfElseChainLength counts the total length of an if-else-if chain
func (a *AntipatternAnalyzer) countIfElseChainLength(ifStmt *ast.IfStmt) int {
	count := 1 // Count the initial if statement

	// Follow the else chain
	current := ifStmt
	for current.Else != nil {
		count++
		// Check if the else is another if statement (else if)
		nextIf, ok := current.Else.(*ast.IfStmt)
		if !ok {
			// It's a final else block, stop counting
			break
		}
		current = nextIf
	}

	return count
}

// checkUnusedReceiverName detects method receivers that are never referenced in the method body.
// A receiver name that is never used suggests the method could be a plain function or the receiver
// should be named _ (underscore) following Go idioms. This improves code clarity and signals intent.
func (a *AntipatternAnalyzer) checkUnusedReceiverName(funcDecl *ast.FuncDecl) []metrics.PerformanceAntipattern {
	var patterns []metrics.PerformanceAntipattern

	// Only check methods (functions with receivers)
	if funcDecl.Recv == nil || len(funcDecl.Recv.List) == 0 {
		return patterns
	}

	// Get the receiver field
	receiver := funcDecl.Recv.List[0]

	// Skip if receiver is already named _ (idiomatic Go for unused receivers)
	if len(receiver.Names) == 0 || receiver.Names[0].Name == "_" {
		return patterns
	}

	receiverName := receiver.Names[0].Name

	// Check if receiver is referenced in the function body
	if !a.isReceiverUsed(funcDecl.Body, receiverName) {
		patterns = append(patterns, metrics.PerformanceAntipattern{
			Type:        "unused_receiver",
			Description: "Method receiver is never referenced in method body",
			Severity:    metrics.SeverityLevelInfo,
			File:        a.fset.Position(funcDecl.Pos()).Filename,
			Line:        a.fset.Position(funcDecl.Pos()).Line,
			Suggestion:  "Use _ as receiver name or convert to plain function if receiver is not needed",
		})
	}

	return patterns
}

// isReceiverUsed checks if a receiver name is referenced anywhere in the function body
func (a *AntipatternAnalyzer) isReceiverUsed(body *ast.BlockStmt, receiverName string) bool {
	if body == nil {
		return false
	}

	used := false
	ast.Inspect(body, func(n ast.Node) bool {
		if used {
			return false
		}

		// Check for identifier references
		if ident, ok := n.(*ast.Ident); ok && ident.Name == receiverName {
			used = true
			return false
		}

		return true
	})

	return used
}

// CheckTestOnlyExports detects exported symbols with zero cross-package references outside
// test files. This identifies symbols that are exported solely for test access, which could
// instead use export_test.go patterns or be restructured to test via the public API. This
// prevents API surface bloat and clarifies what is truly public vs. test infrastructure.
func CheckTestOnlyExports(files map[string]*ast.File, fset *token.FileSet, packagePath string) []metrics.PerformanceAntipattern {
	exportedSymbols, symbolReferences := collectExportData(files, fset, packagePath)
	return detectTestOnlyExports(exportedSymbols, symbolReferences)
}

// collectExportData performs two-pass collection of exported symbols and their references
func collectExportData(files map[string]*ast.File, fset *token.FileSet, packagePath string) (map[string]*exportedSymbolInfo, map[string][]string) {
	exportedSymbols := make(map[string]*exportedSymbolInfo)
	symbolReferences := make(map[string][]string)

	// First pass: Collect all exported symbols from non-test files
	for filePath, file := range files {
		if !isTestFile(filePath) {
			collectExportedSymbols(file, filePath, fset, packagePath, exportedSymbols)
		}
	}

	// Second pass: Track references to exported symbols
	for filePath, file := range files {
		trackSymbolReferences(file, filePath, exportedSymbols, symbolReferences)
	}

	return exportedSymbols, symbolReferences
}

// detectTestOnlyExports identifies exported symbols only used in tests
func detectTestOnlyExports(exportedSymbols map[string]*exportedSymbolInfo, symbolReferences map[string][]string) []metrics.PerformanceAntipattern {
	var patterns []metrics.PerformanceAntipattern

	for symbolName, info := range exportedSymbols {
		if isTestOnlySymbol(info, symbolReferences[symbolName]) {
			patterns = append(patterns, createTestOnlyPattern(symbolName, info))
		}
	}

	return patterns
}

// isTestOnlySymbol checks if a symbol is only referenced in test files
func isTestOnlySymbol(info *exportedSymbolInfo, refs []string) bool {
	for _, refFile := range refs {
		if refFile != info.File && !isTestFile(refFile) {
			return false
		}
	}
	return true
}

// createTestOnlyPattern creates an antipattern report for test-only exports
func createTestOnlyPattern(symbolName string, info *exportedSymbolInfo) metrics.PerformanceAntipattern {
	return metrics.PerformanceAntipattern{
		Type:        "test_only_export",
		Description: fmt.Sprintf("Exported %s '%s' has zero cross-package references outside test files", info.SymbolType, symbolName),
		Severity:    metrics.SeverityLevelInfo,
		File:        info.File,
		Line:        info.Line,
		Suggestion:  "Consider using export_test.go patterns, making symbol unexported, or restructuring tests to use the public API",
	}
}

// exportedSymbolInfo tracks information about an exported symbol
type exportedSymbolInfo struct {
	Name       string
	Package    string
	File       string
	Line       int
	SymbolType string // "function", "type", "variable", "constant"
}

// collectExportedSymbols collects all exported symbols from a file
func collectExportedSymbols(file *ast.File, filePath string, fset *token.FileSet, packagePath string, exports map[string]*exportedSymbolInfo) {
	for _, decl := range file.Decls {
		switch d := decl.(type) {
		case *ast.FuncDecl:
			collectExportedFunc(d, filePath, fset, packagePath, exports)
		case *ast.GenDecl:
			collectExportedGenDecl(d, filePath, fset, packagePath, exports)
		}
	}
}

// collectExportedFunc handles exported function/method declarations
func collectExportedFunc(d *ast.FuncDecl, filePath string, fset *token.FileSet, packagePath string, exports map[string]*exportedSymbolInfo) {
	if d.Name != nil && d.Name.IsExported() {
		exports[d.Name.Name] = &exportedSymbolInfo{
			Name:       d.Name.Name,
			Package:    packagePath,
			File:       filePath,
			Line:       fset.Position(d.Pos()).Line,
			SymbolType: "function",
		}
	}
}

// collectExportedGenDecl handles exported type/value declarations
func collectExportedGenDecl(d *ast.GenDecl, filePath string, fset *token.FileSet, packagePath string, exports map[string]*exportedSymbolInfo) {
	for _, spec := range d.Specs {
		switch s := spec.(type) {
		case *ast.TypeSpec:
			collectExportedType(s, filePath, fset, packagePath, exports)
		case *ast.ValueSpec:
			collectExportedValue(s, d.Tok, filePath, fset, packagePath, exports)
		}
	}
}

// collectExportedType handles exported type declarations
func collectExportedType(s *ast.TypeSpec, filePath string, fset *token.FileSet, packagePath string, exports map[string]*exportedSymbolInfo) {
	if s.Name != nil && s.Name.IsExported() {
		exports[s.Name.Name] = &exportedSymbolInfo{
			Name:       s.Name.Name,
			Package:    packagePath,
			File:       filePath,
			Line:       fset.Position(s.Pos()).Line,
			SymbolType: "type",
		}
	}
}

// collectExportedValue handles exported variable/constant declarations
func collectExportedValue(s *ast.ValueSpec, tok token.Token, filePath string, fset *token.FileSet, packagePath string, exports map[string]*exportedSymbolInfo) {
	for _, name := range s.Names {
		if name != nil && name.IsExported() {
			symbolType := "variable"
			if tok == token.CONST {
				symbolType = "constant"
			}
			exports[name.Name] = &exportedSymbolInfo{
				Name:       name.Name,
				Package:    packagePath,
				File:       filePath,
				Line:       fset.Position(name.Pos()).Line,
				SymbolType: symbolType,
			}
		}
	}
}

// trackSymbolReferences tracks references to exported symbols
func trackSymbolReferences(file *ast.File, filePath string, exports map[string]*exportedSymbolInfo, references map[string][]string) {
	ast.Inspect(file, func(n ast.Node) bool {
		// Check for identifier usage
		if ident, ok := n.(*ast.Ident); ok {
			if info, exists := exports[ident.Name]; exists {
				// Don't count references in the same file as the definition
				if filePath != info.File {
					references[ident.Name] = append(references[ident.Name], filePath)
				}
			}
		}
		return true
	})
}

// isTestFile checks if a file is a test file
func isTestFile(filePath string) bool {
	return strings.HasSuffix(filePath, "_test.go")
}
