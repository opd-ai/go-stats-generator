package analyzer

import (
	"fmt"
	"go/ast"
	"go/token"

	"github.com/opd-ai/go-stats-generator/internal/metrics"
)

// BurdenAnalyzer detects maintenance burden indicators in Go code
type BurdenAnalyzer struct {
	fset *token.FileSet
}

// NewBurdenAnalyzer creates a new maintenance burden analyzer for detecting code quality issues
// that increase maintenance costs. The analyzer identifies magic numbers (unexplained literals),
// dead code (unreachable statements), complex function signatures, deep nesting patterns, and
// feature envy (excessive coupling to other types). Essential for technical debt assessment.
func NewBurdenAnalyzer(fset *token.FileSet) *BurdenAnalyzer {
	return &BurdenAnalyzer{
		fset: fset,
	}
}

// FileSet returns the token.FileSet used by this analyzer for mapping AST node positions
// to source file locations. This enables accurate line number reporting for detected
// maintenance burden indicators, magic numbers, and code smells. Required for position-aware analysis.
func (ba *BurdenAnalyzer) FileSet() *token.FileSet {
	return ba.fset
}

// DetectMagicNumbers identifies numeric and string literals that lack meaningful names,
// making code harder to understand and maintain. It excludes benign values (0, 1, -1, empty strings)
// and constants declared in const blocks. Magic numbers increase cognitive burden and risk of bugs
// when values need to change. Returns a list of detected magic numbers with file locations.
func (ba *BurdenAnalyzer) DetectMagicNumbers(file *ast.File, pkg string) []metrics.MagicNumber {
	var magicNumbers []metrics.MagicNumber
	var currentFunc string

	ast.Inspect(file, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.FuncDecl:
			if node.Name != nil {
				currentFunc = node.Name.Name
			}
			return true

		case *ast.GenDecl:
			// Skip const declarations - these are intentional constants
			if node.Tok == token.CONST {
				return false
			}
			return true

		case *ast.BasicLit:
			if mn := ba.checkBasicLit(node, file, currentFunc); mn != nil {
				magicNumbers = append(magicNumbers, *mn)
			}
			return true
		}
		return true
	})

	return magicNumbers
}

// checkBasicLit examines a basic literal for magic number patterns
func (ba *BurdenAnalyzer) checkBasicLit(lit *ast.BasicLit, file *ast.File, fn string) *metrics.MagicNumber {
	// Check numeric literals
	if lit.Kind == token.INT || lit.Kind == token.FLOAT {
		if ba.isBenignNumber(lit.Value) {
			return nil
		}
		return ba.createMagicNumber(lit, file, fn, "numeric")
	}

	// Check string literals
	if lit.Kind == token.STRING {
		// Exclude empty strings
		if lit.Value == `""` || lit.Value == "``" {
			return nil
		}
		return ba.createMagicNumber(lit, file, fn, "string")
	}

	return nil
}

// isBenignNumber checks if a numeric value is a common constant that should not be flagged
func (ba *BurdenAnalyzer) isBenignNumber(value string) bool {
	// Common benign values that shouldn't be flagged
	benign := map[string]bool{
		"0":    true,
		"1":    true,
		"2":    true,
		"-1":   true,
		"0.0":  true,
		"1.0":  true,
		"0.5":  true,
		"2.0":  true,
		"10":   true,
		"100":  true,
		"1000": true,
		"8":    true,
		"16":   true,
		"32":   true,
		"64":   true,
		"128":  true,
		"256":  true,
		"512":  true,
		"1024": true,
		"0x00": true,
		"0xff": true,
		"0xFF": true,
	}
	return benign[value]
}

// createMagicNumber constructs a MagicNumber metric from a literal and its context
func (ba *BurdenAnalyzer) createMagicNumber(lit *ast.BasicLit, file *ast.File, fn, typ string) *metrics.MagicNumber {
	pos := ba.fset.Position(lit.Pos())
	severity, suggestion := ba.getMagicNumberSeverityAndSuggestion(typ, lit.Value)

	return &metrics.MagicNumber{
		File:       pos.Filename,
		Line:       pos.Line,
		Column:     pos.Column,
		Value:      lit.Value,
		Type:       typ,
		Context:    ba.extractContext(lit, file),
		Function:   fn,
		Severity:   severity,
		Suggestion: suggestion,
	}
}

// getMagicNumberSeverityAndSuggestion determines severity and suggestion for magic numbers
func (ba *BurdenAnalyzer) getMagicNumberSeverityAndSuggestion(typ, value string) (metrics.SeverityLevel, string) {
	if typ == "string" {
		return metrics.SeverityLevelInfo, "Consider extracting string literal into a const if reused or semantically meaningful"
	}
	return metrics.SeverityLevelWarning, fmt.Sprintf("Extract %s literal '%s' into a named constant for better maintainability", typ, value)
}

// extractContext finds the statement context for a literal by inspecting the AST
func (ba *BurdenAnalyzer) extractContext(lit *ast.BasicLit, file *ast.File) string {
	// Find the statement containing this literal
	var context string
	ast.Inspect(file, func(n ast.Node) bool {
		if ctx := ba.checkNodeContext(n, lit); ctx != "" {
			context = ctx
			return false
		}
		return true
	})
	if context == "" {
		context = "expression"
	}
	return context
}

// checkNodeContext determines the usage context of a literal within an AST node,
// categorizing it as assignment, return, function_call, binary_expression, or generic expression.
func (ba *BurdenAnalyzer) checkNodeContext(n ast.Node, lit *ast.BasicLit) string {
	switch node := n.(type) {
	case *ast.AssignStmt:
		return ba.checkAssignStmtContext(node, lit)
	case *ast.ReturnStmt:
		return ba.checkReturnStmtContext(node, lit)
	case *ast.CallExpr:
		return ba.checkCallExprContext(node, lit)
	case *ast.BinaryExpr:
		return ba.checkBinaryExprContext(node, lit)
	}
	return ""
}

// checkAssignStmtContext checks if literal appears in assignment statement
func (ba *BurdenAnalyzer) checkAssignStmtContext(node *ast.AssignStmt, lit *ast.BasicLit) string {
	for _, rhs := range node.Rhs {
		if ba.containsNode(rhs, lit) {
			return "assignment"
		}
	}
	return ""
}

// checkReturnStmtContext checks if literal appears in return statement
func (ba *BurdenAnalyzer) checkReturnStmtContext(node *ast.ReturnStmt, lit *ast.BasicLit) string {
	for _, res := range node.Results {
		if ba.containsNode(res, lit) {
			return "return"
		}
	}
	return ""
}

// checkCallExprContext checks if literal appears in function call
func (ba *BurdenAnalyzer) checkCallExprContext(node *ast.CallExpr, lit *ast.BasicLit) string {
	for _, arg := range node.Args {
		if ba.containsNode(arg, lit) {
			return "function_call"
		}
	}
	return ""
}

// checkBinaryExprContext checks if literal appears in binary expression
func (ba *BurdenAnalyzer) checkBinaryExprContext(node *ast.BinaryExpr, lit *ast.BasicLit) string {
	if ba.containsNode(node, lit) {
		return "binary_expression"
	}
	return ""
}

// containsNode checks if a parent AST node contains a target child node
func (ba *BurdenAnalyzer) containsNode(parent, target ast.Node) bool {
	found := false
	ast.Inspect(parent, func(n ast.Node) bool {
		if n == target {
			found = true
			return false
		}
		return true
	})
	return found
}

// BurdenFileInfo holds a parsed source file together with the token.FileSet it was
// parsed into and the Go package it belongs to. Use this when files from multiple
// worker goroutines — each with its own token.FileSet — are analyzed together at
// package scope so that position lookups remain accurate.
type BurdenFileInfo struct {
	// File is the parsed AST for a single Go source file.
	File *ast.File
	// Fset is the token.FileSet that File was parsed into. Each worker goroutine
	// creates its own FileSet, so Fset must travel with File to ensure that
	// position information (file name, line numbers) can be resolved correctly
	// when files from different goroutines are analyzed together.
	Fset *token.FileSet
	// Pkg is the Go package name this file belongs to (e.g. "mypkg" or "main").
	// It is used to determine which entry points are always-live (init everywhere;
	// main only when Pkg == "main").
	Pkg string
}

// callGraphNode returns the disambiguated call-graph node key for a function declaration.
// Methods are qualified as "ReceiverType.methodName" so they never collide with
// package-level functions of the same name. Standalone functions use their bare name.
func callGraphNode(fn *ast.FuncDecl) string {
	recvType := GetMethodReceiverType(fn)
	if recvType == "" {
		return fn.Name.Name
	}
	return recvType + "." + fn.Name.Name
}

// DetectDeadCode is a convenience wrapper for single-FileSet analysis (e.g. unit tests).
// All files must have been parsed into the BurdenAnalyzer's own FileSet. For multi-file
// package analysis where files come from separate worker goroutines (each with their own
// FileSet), use DetectDeadCodeForPackage with per-file BurdenFileInfo entries instead.
func (ba *BurdenAnalyzer) DetectDeadCode(files []*ast.File, pkg string) *metrics.DeadCodeMetrics {
	fileInfos := make([]BurdenFileInfo, len(files))
	for i, f := range files {
		fileInfos[i] = BurdenFileInfo{File: f, Fset: ba.fset, Pkg: pkg}
	}
	return ba.DetectDeadCodeForPackage(fileInfos)
}

// DetectDeadCodeForPackage runs dead-code analysis at full package scope. It accepts
// one BurdenFileInfo per source file so that each file's position lookups use the
// correct token.FileSet, even when files were parsed by different worker goroutines.
// Public (exported) functions are never considered dead code. Private functions
// reachable — directly or transitively — from any exported function, package-level
// init, or main (in package main) are also excluded.
func (ba *BurdenAnalyzer) DetectDeadCodeForPackage(fileInfos []BurdenFileInfo) *metrics.DeadCodeMetrics {
	if len(fileInfos) == 0 {
		return &metrics.DeadCodeMetrics{}
	}

	pkg := fileInfos[0].Pkg
	files := burdenFileInfoFiles(fileInfos)

	// Build a method-name index so selector calls can be resolved to all matching
	// method declarations without requiring full type information.
	methodIndex := ba.buildMethodIndex(files)

	// Build a per-function call graph across all files in the package.
	callGraph := ba.buildCallGraph(files, methodIndex)

	// Collect always-live entry points: exported functions, package-level init,
	// and main (only in package main).
	roots := ba.collectExportedFunctions(files, pkg)

	// Compute the full reachable set via BFS from every root.
	reachable := ba.findReachableFromExported(callGraph, roots)

	// Find unexported functions not reachable from any live root.
	unreferenced := ba.findUnreferencedSymbolsForPackage(fileInfos, reachable, pkg)

	// Find unreachable code blocks (statements after terminating statements).
	unreachableBlocks := ba.findUnreachableCodeForPackage(fileInfos)

	totalLines := 0
	for _, block := range unreachableBlocks {
		totalLines += block.Lines
	}

	return &metrics.DeadCodeMetrics{
		UnreferencedFunctions: unreferenced,
		UnreachableCode:       unreachableBlocks,
		TotalDeadLines:        totalLines,
		DeadCodePercent:       0.0, // Calculate in analyzer integration
	}
}

// burdenFileInfoFiles extracts the *ast.File slice from a BurdenFileInfo slice.
func burdenFileInfoFiles(fileInfos []BurdenFileInfo) []*ast.File {
	files := make([]*ast.File, len(fileInfos))
	for i, fi := range fileInfos {
		files[i] = fi.File
	}
	return files
}

// buildMethodIndex builds a mapping from bare method name to all qualified method-node
// keys ("RecvType.methodName") across all provided files. This lets the call graph
// resolve selector calls (h.method()) to their declaration nodes without type information:
// since we don't know the concrete type of h at parse time, we conservatively add edges
// to every method with the matching name, preventing false positives.
func (ba *BurdenAnalyzer) buildMethodIndex(files []*ast.File) map[string][]string {
	index := make(map[string][]string)
	for _, file := range files {
		ast.Inspect(file, func(n ast.Node) bool {
			fn, ok := n.(*ast.FuncDecl)
			if !ok || fn.Name == nil {
				return true
			}
			if recvType := GetMethodReceiverType(fn); recvType != "" {
				key := recvType + "." + fn.Name.Name
				index[fn.Name.Name] = append(index[fn.Name.Name], key)
			}
			return true
		})
	}
	return index
}

// buildCallGraph constructs a call graph mapping each function/method node key to the
// set of node keys it calls directly. Plain calls (helper()) target the bare function
// name. Selector calls (h.method()) target all method node keys with that method name,
// using methodIndex to resolve without full type information.
func (ba *BurdenAnalyzer) buildCallGraph(files []*ast.File, methodIndex map[string][]string) map[string]map[string]bool {
	graph := make(map[string]map[string]bool)
	for _, file := range files {
		ba.buildFileCallGraph(file, graph, methodIndex)
	}
	return graph
}

// buildFileCallGraph populates the call graph with all calls made inside each
// function declaration in file.
func (ba *BurdenAnalyzer) buildFileCallGraph(file *ast.File, graph map[string]map[string]bool, methodIndex map[string][]string) {
	ast.Inspect(file, func(n ast.Node) bool {
		funcDecl, ok := n.(*ast.FuncDecl)
		if !ok {
			return true
		}
		if funcDecl.Name == nil || funcDecl.Body == nil {
			return false
		}

		// Use the disambiguated node key so methods and package-level functions
		// with the same bare name get separate call-graph nodes.
		callerKey := callGraphNode(funcDecl)
		if graph[callerKey] == nil {
			graph[callerKey] = make(map[string]bool)
		}

		// Walk the function body to collect all direct call targets.
		// Note: calls inside nested function literals (closures) are attributed
		// to the enclosing named function. This is conservative — it prevents
		// false positives where a function used only inside a closure would be
		// incorrectly flagged as dead code. The trade-off is that a closure that
		// is defined but never invoked will not be caught as dead code by this
		// pass; that case is rare and is handled separately by the compiler
		// (unused variables cause compile errors) or other analysis passes.
		ast.Inspect(funcDecl.Body, func(inner ast.Node) bool {
			call, ok := inner.(*ast.CallExpr)
			if !ok {
				return true
			}
			if ident, ok := call.Fun.(*ast.Ident); ok {
				// Plain call foo() targets the package-level function "foo".
				// Built-in names (make, len, …) are harmless extras since they
				// will never match a declared unexported function.
				graph[callerKey][ident.Name] = true
			} else if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
				// Selector call h.method() on a local variable: add edges to
				// ALL methods named "method" across all receiver types. Without
				// full type information we cannot be more precise; being
				// conservative here avoids false positives.
				// ident.Obj is non-nil for local identifiers and nil for
				// imported package names (os, log, …), filtering cross-package
				// calls that are outside the analyzed package.
				if ident, ok := sel.X.(*ast.Ident); ok && ident.Obj != nil {
					for _, qualifiedName := range methodIndex[sel.Sel.Name] {
						graph[callerKey][qualifiedName] = true
					}
				}
			}
			return true
		})

		// We handled this FuncDecl's children ourselves; don't re-visit them.
		return false
	})
}

// collectExportedFunctions returns a set of call-graph node keys for all always-live
// entry points: exported functions/methods, package-level init (any package), and
// main (only when pkg == "main"). Using qualified keys ensures that e.g. an exported
// method "T.Method" and a package-level function "Method" are tracked independently.
func (ba *BurdenAnalyzer) collectExportedFunctions(files []*ast.File, pkg string) map[string]bool {
	roots := make(map[string]bool)
	for _, file := range files {
		ast.Inspect(file, func(n ast.Node) bool {
			funcDecl, ok := n.(*ast.FuncDecl)
			if !ok || funcDecl.Name == nil {
				return true
			}
			name := funcDecl.Name.Name
			nodeKey := callGraphNode(funcDecl)
			isPackageLevel := !IsMethod(funcDecl)
			switch {
			case ast.IsExported(name):
				roots[nodeKey] = true
			case name == "init" && isPackageLevel:
				// Package-level init is always invoked by the runtime.
				roots[nodeKey] = true
			case name == "main" && isPackageLevel && pkg == "main":
				// main() is an entry point only in package main.
				roots[nodeKey] = true
			}
			return true
		})
	}
	return roots
}

// findReachableFromExported performs a breadth-first traversal of the call graph
// starting from all root (always-live) functions and returns the set of all
// node keys that are reachable from at least one root.
func (ba *BurdenAnalyzer) findReachableFromExported(graph map[string]map[string]bool, roots map[string]bool) map[string]bool {
	reachable := make(map[string]bool)
	queue := make([]string, 0, len(roots))
	for name := range roots {
		if !reachable[name] {
			reachable[name] = true
			queue = append(queue, name)
		}
	}
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		for callee := range graph[current] {
			if !reachable[callee] {
				reachable[callee] = true
				queue = append(queue, callee)
			}
		}
	}
	return reachable
}

// findUnreferencedSymbolsForPackage identifies unexported functions that are not
// reachable from any always-live root. It uses the per-file FileSet from each
// BurdenFileInfo entry for accurate position reporting.
func (ba *BurdenAnalyzer) findUnreferencedSymbolsForPackage(fileInfos []BurdenFileInfo, reachable map[string]bool, pkg string) []metrics.UnreferencedSymbol {
	var unreferenced []metrics.UnreferencedSymbol

	for _, fi := range fileInfos {
		ast.Inspect(fi.File, func(n ast.Node) bool {
			node, ok := n.(*ast.FuncDecl)
			if !ok || node.Name == nil {
				return true
			}
			// Exported functions are never dead code.
			if ast.IsExported(node.Name.Name) {
				return true
			}
			// Flag only if the qualified node key is not reachable.
			nodeKey := callGraphNode(node)
			if !reachable[nodeKey] {
				pos := fi.Fset.Position(node.Pos())
				unreferenced = append(unreferenced, metrics.UnreferencedSymbol{
					Name:    node.Name.Name,
					File:    pos.Filename,
					Line:    pos.Line,
					Type:    "function",
					Package: pkg,
				})
			}
			return true
		})
	}

	return unreferenced
}

// findUnreachableCodeForPackage identifies unreachable code blocks across all files,
// using each file's own FileSet for accurate position reporting.
func (ba *BurdenAnalyzer) findUnreachableCodeForPackage(fileInfos []BurdenFileInfo) []metrics.UnreachableBlock {
	var unreachable []metrics.UnreachableBlock
	for _, fi := range fileInfos {
		unreachable = append(unreachable, ba.findUnreachableCodeInFile(fi.File, fi.Fset)...)
	}
	return unreachable
}

// findUnreachableCodeInFile identifies unreachable code blocks within a single file,
// using the provided FileSet for position lookups.
func (ba *BurdenAnalyzer) findUnreachableCodeInFile(file *ast.File, fset *token.FileSet) []metrics.UnreachableBlock {
	var unreachable []metrics.UnreachableBlock
	var currentFunc string

	ast.Inspect(file, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.FuncDecl:
			if node.Name != nil {
				currentFunc = node.Name.Name
			}
			if node.Body != nil {
				blocks := ba.checkBlockForUnreachable(node.Body, currentFunc, fset)
				unreachable = append(unreachable, blocks...)
			}
			return false
		}
		return true
	})

	return unreachable
}

// checkBlockForUnreachable scans a code block for statements after terminating statements.
func (ba *BurdenAnalyzer) checkBlockForUnreachable(block *ast.BlockStmt, fn string, fset *token.FileSet) []metrics.UnreachableBlock {
	var unreachable []metrics.UnreachableBlock

	for i, stmt := range block.List {
		// Check if this statement is a terminating statement
		if ba.isTerminating(stmt) {
			// Check if there are statements after this
			if i+1 < len(block.List) {
				startPos := fset.Position(block.List[i+1].Pos())
				endPos := fset.Position(block.List[len(block.List)-1].End())

				unreachable = append(unreachable, metrics.UnreachableBlock{
					File:      startPos.Filename,
					StartLine: startPos.Line,
					EndLine:   endPos.Line,
					Function:  fn,
					Reason:    ba.getTerminationReason(stmt),
					Lines:     endPos.Line - startPos.Line + 1,
				})
				break
			}
		}

		// Recursively check nested blocks
		unreachable = append(unreachable, ba.checkStmtForUnreachable(stmt, fn, fset)...)
	}

	return unreachable
}

// checkStmtForUnreachable analyzes a statement for unreachable code blocks,
// recursively checking nested control structures (if/else, for, switch, select).
// Returns all unreachable blocks found within the statement and its children.
func (ba *BurdenAnalyzer) checkStmtForUnreachable(stmt ast.Stmt, fn string, fset *token.FileSet) []metrics.UnreachableBlock {
	var unreachable []metrics.UnreachableBlock

	switch s := stmt.(type) {
	case *ast.IfStmt:
		unreachable = append(unreachable, ba.checkIfStmtUnreachable(s, fn, fset)...)
	case *ast.ForStmt:
		unreachable = append(unreachable, ba.checkLoopBodyUnreachable(s.Body, fn, fset)...)
	case *ast.RangeStmt:
		unreachable = append(unreachable, ba.checkLoopBodyUnreachable(s.Body, fn, fset)...)
	case *ast.SwitchStmt:
		unreachable = append(unreachable, ba.checkSwitchCasesUnreachable(s.Body, fn, fset)...)
	case *ast.TypeSwitchStmt:
		unreachable = append(unreachable, ba.checkSwitchCasesUnreachable(s.Body, fn, fset)...)
	}

	return unreachable
}

// checkIfStmtUnreachable analyzes if/else statements for unreachable blocks,
// checking both the main body and any else clause for dead code patterns.
func (ba *BurdenAnalyzer) checkIfStmtUnreachable(s *ast.IfStmt, fn string, fset *token.FileSet) []metrics.UnreachableBlock {
	var unreachable []metrics.UnreachableBlock

	if s.Body != nil {
		unreachable = append(unreachable, ba.checkBlockForUnreachable(s.Body, fn, fset)...)
	}

	if s.Else != nil {
		unreachable = append(unreachable, ba.checkElseClauseUnreachable(s.Else, fn, fset)...)
	}

	return unreachable
}

// checkElseClauseUnreachable analyzes else clauses for unreachable blocks,
// handling both block statements and chained if-else statements recursively.
func (ba *BurdenAnalyzer) checkElseClauseUnreachable(elseStmt ast.Stmt, fn string, fset *token.FileSet) []metrics.UnreachableBlock {
	switch e := elseStmt.(type) {
	case *ast.BlockStmt:
		return ba.checkBlockForUnreachable(e, fn, fset)
	case *ast.IfStmt:
		return ba.checkStmtForUnreachable(e, fn, fset)
	}
	return nil
}

// checkLoopBodyUnreachable analyzes loop body for unreachable blocks,
// supporting both for loops and range loops with nil-safe checking.
func (ba *BurdenAnalyzer) checkLoopBodyUnreachable(body *ast.BlockStmt, fn string, fset *token.FileSet) []metrics.UnreachableBlock {
	if body == nil {
		return nil
	}
	return ba.checkBlockForUnreachable(body, fn, fset)
}

// checkSwitchCasesUnreachable analyzes switch/type switch cases for unreachable blocks,
// examining each case clause body for dead code after terminating statements.
func (ba *BurdenAnalyzer) checkSwitchCasesUnreachable(body *ast.BlockStmt, fn string, fset *token.FileSet) []metrics.UnreachableBlock {
	if body == nil {
		return nil
	}

	var unreachable []metrics.UnreachableBlock
	for _, clause := range body.List {
		if cc, ok := clause.(*ast.CaseClause); ok {
			unreachable = append(unreachable, ba.checkBlockForUnreachable(&ast.BlockStmt{List: cc.Body}, fn, fset)...)
		}
	}
	return unreachable
}

// isTerminating checks if a statement unconditionally terminates execution,
// detecting return statements, os.Exit calls, panic calls, and log.Fatal* calls.
func (ba *BurdenAnalyzer) isTerminating(stmt ast.Stmt) bool {
	if ba.isReturnStmt(stmt) {
		return true
	}
	if ba.isOsExitCall(stmt) {
		return true
	}
	if ba.isPanicCall(stmt) {
		return true
	}
	if ba.isLogFatalCall(stmt) {
		return true
	}
	return false
}

// isReturnStmt checks if a statement is a return statement
func (ba *BurdenAnalyzer) isReturnStmt(stmt ast.Stmt) bool {
	_, ok := stmt.(*ast.ReturnStmt)
	return ok
}

// extractCallExpr extracts a CallExpr from a statement if it's an expression statement.
func extractCallExpr(stmt ast.Stmt) (*ast.CallExpr, bool) {
	exprStmt, ok := stmt.(*ast.ExprStmt)
	if !ok {
		return nil, false
	}
	call, ok := exprStmt.X.(*ast.CallExpr)
	return call, ok
}

// isOsExitCall checks if a statement is an os.Exit call
func (ba *BurdenAnalyzer) isOsExitCall(stmt ast.Stmt) bool {
	return ba.isSelectorCall(stmt, "os", "Exit")
}

// isPanicCall checks if a statement is a panic call
func (ba *BurdenAnalyzer) isPanicCall(stmt ast.Stmt) bool {
	call, ok := extractCallExpr(stmt)
	if !ok {
		return false
	}
	ident, ok := call.Fun.(*ast.Ident)
	if !ok {
		return false
	}
	return ident.Name == "panic"
}

// isLogFatalCall checks if a statement is a log.Fatal, log.Fatalf, or log.Fatalln call
func (ba *BurdenAnalyzer) isLogFatalCall(stmt ast.Stmt) bool {
	call, ok := extractCallExpr(stmt)
	if !ok {
		return false
	}
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	ident, ok := sel.X.(*ast.Ident)
	if !ok {
		return false
	}
	if ident.Name == "log" {
		return sel.Sel.Name == "Fatal" || sel.Sel.Name == "Fatalf" || sel.Sel.Name == "Fatalln"
	}
	return false
}

// isSelectorCall checks if a statement is a call to pkg.method
func (ba *BurdenAnalyzer) isSelectorCall(stmt ast.Stmt, pkg, method string) bool {
	call, ok := extractCallExpr(stmt)
	if !ok {
		return false
	}
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	ident, ok := sel.X.(*ast.Ident)
	if !ok {
		return false
	}
	return ident.Name == pkg && sel.Sel.Name == method
}

// getTerminationReason returns a human-readable description of why a statement
// terminates execution (e.g., "return statement", "os.Exit call", "panic call", "log.Fatal call").
func (ba *BurdenAnalyzer) getTerminationReason(stmt ast.Stmt) string {
	if ba.isReturnStmt(stmt) {
		return "return statement"
	}
	if ba.isOsExitCall(stmt) {
		return "os.Exit call"
	}
	if ba.isPanicCall(stmt) {
		return "panic call"
	}
	if ba.isLogFatalCall(stmt) {
		return "log.Fatal call"
	}
	return "terminating statement"
}

// AnalyzeSignatureComplexity flags functions with excessive parameters or return values that
// increase cognitive load and error risk. It also detects boolean parameters which often indicate
// control-flow coupling. Functions exceeding the parameter/return thresholds or using boolean flags
// are harder to test, understand, and maintain. Returns signature complexity metrics with severity.
func (ba *BurdenAnalyzer) AnalyzeSignatureComplexity(fn *ast.FuncDecl, maxParams, maxReturns int) *metrics.SignatureIssue {
	if fn == nil || fn.Type == nil {
		return nil
	}

	paramCount, boolParams := ba.countParameters(fn.Type)
	returnCount := ba.countReturns(fn.Type)

	if paramCount <= maxParams && returnCount <= maxReturns && len(boolParams) == 0 {
		return nil
	}

	severity := ba.calculateSeverity(paramCount, returnCount, maxParams, maxReturns)
	pos := ba.fset.Position(fn.Pos())
	suggestion := ba.getSignatureSuggestion(paramCount, returnCount, maxParams, maxReturns, boolParams)

	return &metrics.SignatureIssue{
		Function:       fn.Name.Name,
		File:           pos.Filename,
		Line:           pos.Line,
		ParameterCount: paramCount,
		ReturnCount:    returnCount,
		BoolParams:     boolParams,
		Severity:       severity,
		Suggestion:     suggestion,
	}
}

// getSignatureSuggestion generates an actionable suggestion for complex function signatures
func (ba *BurdenAnalyzer) getSignatureSuggestion(paramCount, returnCount, maxParams, maxReturns int, boolParams []string) string {
	if paramCount > maxParams && returnCount > maxReturns {
		return fmt.Sprintf("Function has %d parameters and %d returns. Consider using a config struct for parameters and a result struct for returns", paramCount, returnCount)
	}
	if paramCount > maxParams {
		return fmt.Sprintf("Function has %d parameters. Consider grouping related parameters into a config struct or using functional options pattern", paramCount)
	}
	if returnCount > maxReturns {
		return fmt.Sprintf("Function has %d return values. Consider using a result struct to group related returns", returnCount)
	}
	if len(boolParams) > 0 {
		return fmt.Sprintf("Function has boolean parameters %v. Consider using functional options or command objects instead", boolParams)
	}
	return ""
}

// countParameters counts the total parameters in a function signature and
// identifies boolean parameters by name, returning the count and a list of
// boolean parameter names for complexity analysis.
func (ba *BurdenAnalyzer) countParameters(fnType *ast.FuncType) (int, []string) {
	if fnType.Params == nil {
		return 0, nil
	}
	paramCount := 0
	var boolParams []string
	for _, field := range fnType.Params.List {
		paramCount += ba.countFieldParameters(field)
		boolParams = append(boolParams, ba.extractBoolParameters(field)...)
	}
	return paramCount, boolParams
}

// countFieldParameters counts parameters in a single field
func (ba *BurdenAnalyzer) countFieldParameters(field *ast.Field) int {
	if len(field.Names) == 0 {
		return 1
	}
	return len(field.Names)
}

// extractBoolParameters extracts boolean parameter names from a field
func (ba *BurdenAnalyzer) extractBoolParameters(field *ast.Field) []string {
	if ident, ok := field.Type.(*ast.Ident); ok && ident.Name == "bool" {
		var boolParams []string
		for _, name := range field.Names {
			boolParams = append(boolParams, name.Name)
		}
		return boolParams
	}
	return nil
}

// countReturns calculates the number of return values in a function signature
func (ba *BurdenAnalyzer) countReturns(fnType *ast.FuncType) int {
	if fnType.Results == nil {
		return 0
	}

	returnCount := 0
	for _, field := range fnType.Results.List {
		numNames := len(field.Names)
		if numNames == 0 {
			numNames = 1
		}
		returnCount += numNames
	}

	return returnCount
}

// calculateSeverity determines severity level based on parameter and return count thresholds
func (ba *BurdenAnalyzer) calculateSeverity(paramCount, returnCount, maxParams, maxReturns int) metrics.SeverityLevel {
	if paramCount > maxParams*2 || returnCount > maxReturns*2 {
		return metrics.SeverityLevelViolation
	}
	if paramCount > maxParams || returnCount > maxReturns {
		return metrics.SeverityLevelWarning
	}
	return metrics.SeverityLevelInfo
}

// DetectDeepNesting identifies functions with excessive nesting levels that harm readability
// and increase cognitive complexity. Deep nesting (>3-4 levels) makes code harder to understand,
// test, and debug. It often indicates opportunities for refactoring into smaller functions or
// using early returns. Returns nesting issue details if threshold is exceeded, nil otherwise.
func (ba *BurdenAnalyzer) DetectDeepNesting(fn *ast.FuncDecl, maxNesting int) *metrics.NestingIssue {
	if fn == nil || fn.Body == nil {
		return nil
	}

	maxDepth := 0
	var deepestLoc token.Pos
	ba.walkForNestingDepth(fn.Body, 0, &maxDepth, &deepestLoc)

	if maxDepth <= maxNesting {
		return nil
	}

	pos := ba.fset.Position(fn.Pos())
	locPos := ba.fset.Position(deepestLoc)

	// Determine severity based on depth
	severity := metrics.SeverityLevelWarning
	if maxDepth > maxNesting+3 {
		severity = metrics.SeverityLevelViolation
	}

	return &metrics.NestingIssue{
		Function:   fn.Name.Name,
		File:       pos.Filename,
		Line:       pos.Line,
		MaxDepth:   maxDepth,
		Location:   locPos.String(),
		Severity:   severity,
		Suggestion: "Consider extracting nested logic into separate functions or using early returns/guard clauses",
	}
}

// walkForNestingDepth recursively traverses the AST to find the maximum nesting depth
// of control flow structures (if, for, switch, select). Updates maxDepth and deepestLoc
// when a deeper nesting level is found.
func (ba *BurdenAnalyzer) walkForNestingDepth(node ast.Node, currentDepth int, maxDepth *int, deepestLoc *token.Pos) {
	if node == nil {
		return
	}

	switch n := node.(type) {
	case *ast.IfStmt:
		ba.walkIfStmtNesting(n, currentDepth, maxDepth, deepestLoc)
	case *ast.ForStmt:
		ba.walkForStmtNesting(n, currentDepth, maxDepth, deepestLoc)
	case *ast.RangeStmt:
		ba.walkRangeStmtNesting(n, currentDepth, maxDepth, deepestLoc)
	case *ast.SwitchStmt:
		ba.walkSwitchStmtNesting(n, currentDepth, maxDepth, deepestLoc)
	case *ast.TypeSwitchStmt:
		ba.walkTypeSwitchStmtNesting(n, currentDepth, maxDepth, deepestLoc)
	case *ast.SelectStmt:
		ba.walkSelectStmtNesting(n, currentDepth, maxDepth, deepestLoc)
	case *ast.CaseClause:
		ba.walkCaseClauseNesting(n, currentDepth, maxDepth, deepestLoc)
	case *ast.CommClause:
		ba.walkCommClauseNesting(n, currentDepth, maxDepth, deepestLoc)
	case *ast.BlockStmt:
		ba.walkBlockStmtNesting(n, currentDepth, maxDepth, deepestLoc)
	}
}

// walkIfStmtNesting processes if statements for nesting depth tracking
func (ba *BurdenAnalyzer) walkIfStmtNesting(n *ast.IfStmt, currentDepth int, maxDepth *int, deepestLoc *token.Pos) {
	newDepth := currentDepth + 1
	ba.updateMaxDepth(newDepth, n.Pos(), maxDepth, deepestLoc)
	ba.walkForNestingDepth(n.Body, newDepth, maxDepth, deepestLoc)
	if n.Else != nil {
		ba.walkForNestingDepth(n.Else, newDepth, maxDepth, deepestLoc)
	}
}

// walkForStmtNesting processes for statements for nesting depth tracking
func (ba *BurdenAnalyzer) walkForStmtNesting(n *ast.ForStmt, currentDepth int, maxDepth *int, deepestLoc *token.Pos) {
	newDepth := currentDepth + 1
	ba.updateMaxDepth(newDepth, n.Pos(), maxDepth, deepestLoc)
	ba.walkForNestingDepth(n.Body, newDepth, maxDepth, deepestLoc)
}

// walkRangeStmtNesting processes range statements for nesting depth tracking
func (ba *BurdenAnalyzer) walkRangeStmtNesting(n *ast.RangeStmt, currentDepth int, maxDepth *int, deepestLoc *token.Pos) {
	newDepth := currentDepth + 1
	ba.updateMaxDepth(newDepth, n.Pos(), maxDepth, deepestLoc)
	ba.walkForNestingDepth(n.Body, newDepth, maxDepth, deepestLoc)
}

// walkSwitchStmtNesting processes switch statements for nesting depth tracking
func (ba *BurdenAnalyzer) walkSwitchStmtNesting(n *ast.SwitchStmt, currentDepth int, maxDepth *int, deepestLoc *token.Pos) {
	newDepth := currentDepth + 1
	ba.updateMaxDepth(newDepth, n.Pos(), maxDepth, deepestLoc)
	for _, stmt := range n.Body.List {
		ba.walkForNestingDepth(stmt, newDepth, maxDepth, deepestLoc)
	}
}

// walkTypeSwitchStmtNesting processes type switch statements for nesting depth tracking
func (ba *BurdenAnalyzer) walkTypeSwitchStmtNesting(n *ast.TypeSwitchStmt, currentDepth int, maxDepth *int, deepestLoc *token.Pos) {
	newDepth := currentDepth + 1
	ba.updateMaxDepth(newDepth, n.Pos(), maxDepth, deepestLoc)
	ba.walkForNestingDepth(n.Body, newDepth, maxDepth, deepestLoc)
}

// walkSelectStmtNesting processes select statements for nesting depth tracking
func (ba *BurdenAnalyzer) walkSelectStmtNesting(n *ast.SelectStmt, currentDepth int, maxDepth *int, deepestLoc *token.Pos) {
	newDepth := currentDepth + 1
	ba.updateMaxDepth(newDepth, n.Pos(), maxDepth, deepestLoc)
	ba.walkForNestingDepth(n.Body, newDepth, maxDepth, deepestLoc)
}

// walkCaseClauseNesting processes case clauses for nesting depth tracking
func (ba *BurdenAnalyzer) walkCaseClauseNesting(n *ast.CaseClause, currentDepth int, maxDepth *int, deepestLoc *token.Pos) {
	for _, stmt := range n.Body {
		ba.walkForNestingDepth(stmt, currentDepth, maxDepth, deepestLoc)
	}
}

// walkCommClauseNesting processes comm clauses for nesting depth tracking
func (ba *BurdenAnalyzer) walkCommClauseNesting(n *ast.CommClause, currentDepth int, maxDepth *int, deepestLoc *token.Pos) {
	for _, stmt := range n.Body {
		ba.walkForNestingDepth(stmt, currentDepth, maxDepth, deepestLoc)
	}
}

// walkBlockStmtNesting processes block statements for nesting depth tracking
func (ba *BurdenAnalyzer) walkBlockStmtNesting(n *ast.BlockStmt, currentDepth int, maxDepth *int, deepestLoc *token.Pos) {
	for _, stmt := range n.List {
		ba.walkForNestingDepth(stmt, currentDepth, maxDepth, deepestLoc)
	}
}

// updateMaxDepth updates the maximum depth and location if a new maximum is found
func (ba *BurdenAnalyzer) updateMaxDepth(newDepth int, pos token.Pos, maxDepth *int, deepestLoc *token.Pos) {
	if newDepth > *maxDepth {
		*maxDepth = newDepth
		*deepestLoc = pos
	}
}

// DetectFeatureEnvy identifies methods that reference external types more than their own receiver,
// suggesting the method may belong to a different type. Feature envy is a code smell indicating
// poor cohesion and potential design issues. When a method uses another object's data/methods
// extensively, it often should be moved to that object's type. Returns issue details if ratio exceeded.
func (ba *BurdenAnalyzer) DetectFeatureEnvy(fn *ast.FuncDecl, file *ast.File, ratio float64) *metrics.FeatureEnvyIssue {
	if fn == nil || fn.Body == nil || fn.Recv == nil || len(fn.Recv.List) == 0 {
		return nil
	}

	receiverType := ba.extractReceiverType(fn)
	if receiverType == "" {
		return nil
	}

	receiverVar := ba.getReceiverVarName(fn)
	selfRefs, externalRefs := ba.countReferences(fn.Body, receiverVar)

	maxExtType, maxExtCount := ba.findMostReferencedType(externalRefs)
	if !ba.hasFeatureEnvy(maxExtCount, selfRefs, ratio) {
		return nil
	}

	return ba.buildFeatureEnvyIssue(fn, receiverType, selfRefs, maxExtType, maxExtCount)
}

// hasFeatureEnvy determines if the reference counts indicate feature envy based on the ratio threshold.
func (ba *BurdenAnalyzer) hasFeatureEnvy(extCount, selfRefs int, threshold float64) bool {
	if extCount == 0 {
		return false
	}
	return float64(extCount)/float64(max(selfRefs, 1)) >= threshold
}

// buildFeatureEnvyIssue creates a FeatureEnvyIssue from the analyzed method data.
func (ba *BurdenAnalyzer) buildFeatureEnvyIssue(fn *ast.FuncDecl, receiverType string, selfRefs int, extType string, extRefs int) *metrics.FeatureEnvyIssue {
	pos := ba.fset.Position(fn.Pos())
	envyRatio := float64(extRefs) / float64(max(selfRefs, 1))

	severity := metrics.SeverityLevelWarning
	if envyRatio > 3.0 {
		severity = metrics.SeverityLevelViolation
	}

	return &metrics.FeatureEnvyIssue{
		Method:         fn.Name.Name,
		File:           pos.Filename,
		Line:           pos.Line,
		ReceiverType:   receiverType,
		SelfReferences: selfRefs,
		ExternalType:   extType,
		ExternalRefs:   extRefs,
		Ratio:          envyRatio,
		Severity:       severity,
		SuggestedMove:  "Consider moving this method to " + extType + " or extracting shared logic",
	}
}

// getReceiverVarName extracts the receiver variable name from a method declaration,
// returning an empty string if the function is not a method or has no receiver name.
func (ba *BurdenAnalyzer) getReceiverVarName(fn *ast.FuncDecl) string {
	if fn.Recv == nil || len(fn.Recv.List) == 0 || len(fn.Recv.List[0].Names) == 0 {
		return ""
	}
	return fn.Recv.List[0].Names[0].Name
}

// extractReceiverType extracts the receiver type name from a method declaration,
// handling both value and pointer receivers, returning empty string for non-methods.
func (ba *BurdenAnalyzer) extractReceiverType(fn *ast.FuncDecl) string {
	if fn.Recv == nil || len(fn.Recv.List) == 0 {
		return ""
	}

	recvType := fn.Recv.List[0].Type

	switch t := recvType.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		if ident, ok := t.X.(*ast.Ident); ok {
			return ident.Name
		}
	}

	return ""
}

// countReferences analyzes a function body to count self-references (to the receiver) and
// external references (to other variables), returning counts for feature envy detection.
func (ba *BurdenAnalyzer) countReferences(body *ast.BlockStmt, receiverVar string) (int, map[string]int) {
	selfRefs := 0
	externalRefs := make(map[string]int)

	ast.Inspect(body, func(n ast.Node) bool {
		if sel, ok := n.(*ast.SelectorExpr); ok {
			varName := ba.getVarName(sel.X)

			if varName == receiverVar {
				selfRefs++
			} else if varName != "" {
				externalRefs[varName]++
			}
		}
		return true
	})

	return selfRefs, externalRefs
}

// getVarName extracts the variable name from an identifier expression,
// used to identify which object a selector expression is accessing.
func (ba *BurdenAnalyzer) getVarName(expr ast.Expr) string {
	if ident, ok := expr.(*ast.Ident); ok {
		return ident.Name
	}
	return ""
}

// findMostReferencedType identifies the most frequently referenced external type
// from a map of type names to reference counts, used in feature envy detection.
func (ba *BurdenAnalyzer) findMostReferencedType(refs map[string]int) (string, int) {
	maxType := ""
	maxCount := 0

	for typeName, count := range refs {
		if count > maxCount {
			maxType = typeName
			maxCount = count
		}
	}

	return maxType, maxCount
}

// max returns the larger of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
