// Package analyzer provides code analysis capabilities for Go source files.
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

// NewAntipatternAnalyzer creates a new performance anti-pattern analyzer
func NewAntipatternAnalyzer(fset *token.FileSet) *AntipatternAnalyzer {
	return &AntipatternAnalyzer{fset: fset}
}

// Analyze detects performance anti-patterns in a file
func (a *AntipatternAnalyzer) Analyze(file *ast.File) []metrics.PerformanceAntipattern {
	var patterns []metrics.PerformanceAntipattern

	ast.Inspect(file, func(n ast.Node) bool {
		patterns = append(patterns, a.checkMemoryAllocation(n)...)
		patterns = append(patterns, a.checkStringConcatenation(n)...)
		patterns = append(patterns, a.checkGoroutineLeaks(n)...)
		patterns = append(patterns, a.checkResourceManagement(n)...)
		return true
	})

	return patterns
}

// checkMemoryAllocation detects inefficient memory allocation patterns
func (a *AntipatternAnalyzer) checkMemoryAllocation(n ast.Node) []metrics.PerformanceAntipattern {
	var patterns []metrics.PerformanceAntipattern

	switch node := n.(type) {
	case *ast.AssignStmt:
		for _, expr := range node.Rhs {
			if call, ok := expr.(*ast.CallExpr); ok {
				if a.isAppendInLoop(call, node) {
					patterns = append(patterns, metrics.PerformanceAntipattern{
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

	return patterns
}

// checkStringConcatenation detects inefficient string operations
func (a *AntipatternAnalyzer) checkStringConcatenation(n ast.Node) []metrics.PerformanceAntipattern {
	var patterns []metrics.PerformanceAntipattern

	if binExpr, ok := n.(*ast.BinaryExpr); ok {
		if binExpr.Op == token.ADD {
			if a.isStringType(binExpr.X) || a.isStringType(binExpr.Y) {
				if a.isInLoop(binExpr) {
					patterns = append(patterns, metrics.PerformanceAntipattern{
						Type:        "string_concatenation",
						Description: "String concatenation in loop",
						Severity:    "high",
						File:        a.fset.Position(binExpr.Pos()).Filename,
						Line:        a.fset.Position(binExpr.Pos()).Line,
						Suggestion:  "Use strings.Builder for efficient concatenation",
					})
				}
			}
		}
	}

	return patterns
}

// checkGoroutineLeaks detects potential goroutine leaks
func (a *AntipatternAnalyzer) checkGoroutineLeaks(n ast.Node) []metrics.PerformanceAntipattern {
	var patterns []metrics.PerformanceAntipattern

	if goStmt, ok := n.(*ast.GoStmt); ok {
		if !a.hasContextOrDone(goStmt) {
			patterns = append(patterns, metrics.PerformanceAntipattern{
				Type:        "goroutine_leak",
				Description: "Goroutine without context or done channel",
				Severity:    "high",
				File:        a.fset.Position(goStmt.Pos()).Filename,
				Line:        a.fset.Position(goStmt.Pos()).Line,
				Suggestion:  "Add context.Context or done channel for graceful shutdown",
			})
		}
	}

	return patterns
}

// checkResourceManagement detects resource management issues
func (a *AntipatternAnalyzer) checkResourceManagement(n ast.Node) []metrics.PerformanceAntipattern {
	var patterns []metrics.PerformanceAntipattern

	if call, ok := n.(*ast.CallExpr); ok {
		if a.isResourceAcquisition(call) && !a.hasDeferClose(call) {
			patterns = append(patterns, metrics.PerformanceAntipattern{
				Type:        "resource_leak",
				Description: "Resource acquisition without defer close",
				Severity:    "critical",
				File:        a.fset.Position(call.Pos()).Filename,
				Line:        a.fset.Position(call.Pos()).Line,
				Suggestion:  "Use defer to ensure resource cleanup",
			})
		}
	}

	return patterns
}

// isAppendInLoop checks if append is called inside a loop
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

// isInLoop checks if node is inside a loop
func (a *AntipatternAnalyzer) isInLoop(node ast.Node) bool {
	// Simplified check - in production would traverse parent nodes
	return false
}

// hasContextOrDone checks if goroutine has context or done channel
func (a *AntipatternAnalyzer) hasContextOrDone(goStmt *ast.GoStmt) bool {
	if call, ok := goStmt.Call.Fun.(*ast.FuncLit); ok {
		for _, param := range call.Type.Params.List {
			for _, name := range param.Names {
				if strings.Contains(name.Name, "ctx") || strings.Contains(name.Name, "done") {
					return true
				}
			}
		}
	}
	return false
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

// hasDeferClose checks if there's a defer close after resource acquisition
func (a *AntipatternAnalyzer) hasDeferClose(call *ast.CallExpr) bool {
	// Simplified check - in production would check surrounding statements
	return false
}
