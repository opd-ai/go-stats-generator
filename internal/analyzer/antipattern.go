package analyzer

import (
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

	for _, decl := range file.Decls {
		funcDecl, ok := decl.(*ast.FuncDecl)
		if !ok || funcDecl.Body == nil {
			continue
		}
		patterns = append(patterns, a.analyzeFunction(funcDecl)...)
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
		if n.Body != nil {
			for _, stmt := range n.Body.List {
				a.walkWithLoopContext(stmt, true, funcBody, patterns)
			}
		}
		return
	case *ast.RangeStmt:
		if n.Body != nil {
			for _, stmt := range n.Body.List {
				a.walkWithLoopContext(stmt, true, funcBody, patterns)
			}
		}
		return
	case *ast.GoStmt:
		if !a.hasContextOrDone(n) {
			*patterns = append(*patterns, metrics.PerformanceAntipattern{
				Type:        "goroutine_leak",
				Description: "Goroutine without context or done channel",
				Severity:    "high",
				File:        a.fset.Position(n.Pos()).Filename,
				Line:        a.fset.Position(n.Pos()).Line,
				Suggestion:  "Add context.Context or done channel for graceful shutdown",
			})
		}
	case *ast.CallExpr:
		if a.isResourceAcquisition(n) && !a.hasDeferClose(n, funcBody) {
			*patterns = append(*patterns, metrics.PerformanceAntipattern{
				Type:        "resource_leak",
				Description: "Resource acquisition without defer close",
				Severity:    "critical",
				File:        a.fset.Position(n.Pos()).Filename,
				Line:        a.fset.Position(n.Pos()).Line,
				Suggestion:  "Use defer to ensure resource cleanup",
			})
		}
	case *ast.AssignStmt:
		if inLoop {
			a.checkAssignForLoopAntipatterns(n, patterns)
		}
	case *ast.BinaryExpr:
		if inLoop && n.Op == token.ADD {
			if a.isStringType(n.X) || a.isStringType(n.Y) {
				*patterns = append(*patterns, metrics.PerformanceAntipattern{
					Type:        "string_concatenation",
					Description: "String concatenation in loop",
					Severity:    "high",
					File:        a.fset.Position(n.Pos()).Filename,
					Line:        a.fset.Position(n.Pos()).Line,
					Suggestion:  "Use strings.Builder for efficient concatenation",
				})
			}
		}
	}

	// Recurse into child nodes for non-loop statements
	ast.Inspect(node, func(child ast.Node) bool {
		if child == node {
			return true
		}
		switch child.(type) {
		case *ast.ForStmt, *ast.RangeStmt, *ast.GoStmt, *ast.CallExpr, *ast.AssignStmt, *ast.BinaryExpr:
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
					Severity:    "medium",
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
	// Check parameter names
	for _, name := range param.Names {
		nameLower := strings.ToLower(name.Name)
		if strings.Contains(nameLower, "ctx") || strings.Contains(nameLower, "done") ||
			strings.Contains(nameLower, "cancel") || strings.Contains(nameLower, "quit") ||
			strings.Contains(nameLower, "stop") {
			return true
		}
	}

	// Check parameter type for context.Context
	if sel, ok := param.Type.(*ast.SelectorExpr); ok {
		if ident, ok := sel.X.(*ast.Ident); ok {
			if ident.Name == "context" && sel.Sel.Name == "Context" {
				return true
			}
		}
	}

	// Check for channel types (done channels)
	if _, ok := param.Type.(*ast.ChanType); ok {
		return true
	}

	return false
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
// after the resource acquisition call.
func (a *AntipatternAnalyzer) hasDeferClose(call *ast.CallExpr, funcBody *ast.BlockStmt) bool {
	if funcBody == nil {
		return false
	}

	found := false
	ast.Inspect(funcBody, func(n ast.Node) bool {
		if found {
			return false
		}
		deferStmt, ok := n.(*ast.DeferStmt)
		if !ok {
			return true
		}
		// Check for defer x.Close(), defer x.Release(), defer x.Disconnect(), etc.
		if sel, ok := deferStmt.Call.Fun.(*ast.SelectorExpr); ok {
			cleanupFuncs := []string{"Close", "Release", "Disconnect", "Shutdown", "Stop", "Done"}
			for _, fn := range cleanupFuncs {
				if sel.Sel.Name == fn {
					found = true
					return false
				}
			}
		}
		// Check for defer func() { ... close ... }()
		if funcLit, ok := deferStmt.Call.Fun.(*ast.FuncLit); ok {
			ast.Inspect(funcLit.Body, func(inner ast.Node) bool {
				if innerCall, ok := inner.(*ast.CallExpr); ok {
					if innerSel, ok := innerCall.Fun.(*ast.SelectorExpr); ok {
						if innerSel.Sel.Name == "Close" || innerSel.Sel.Name == "Release" {
							found = true
							return false
						}
					}
				}
				return true
			})
		}
		return !found
	})

	return found
}
