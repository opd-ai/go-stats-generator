package analyzer

import (
	"go/ast"
	"go/token"

	"github.com/opd-ai/go-stats-generator/internal/metrics"
)

// BurdenAnalyzer detects maintenance burden indicators in Go code
type BurdenAnalyzer struct {
	fset *token.FileSet
}

// NewBurdenAnalyzer creates a new maintenance burden analyzer
func NewBurdenAnalyzer(fset *token.FileSet) *BurdenAnalyzer {
	return &BurdenAnalyzer{
		fset: fset,
	}
}

// FileSet returns the token.FileSet used by this analyzer
func (ba *BurdenAnalyzer) FileSet() *token.FileSet {
	return ba.fset
}

// DetectMagicNumbers identifies numeric and string literals used as magic numbers
// Excludes benign values: 0, 1, -1, and empty strings
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

func (ba *BurdenAnalyzer) isBenignNumber(value string) bool {
	// Common benign values that shouldn't be flagged
	benign := map[string]bool{
		"0":   true,
		"1":   true,
		"-1":  true,
		"0.0": true,
		"1.0": true,
	}
	return benign[value]
}

func (ba *BurdenAnalyzer) createMagicNumber(lit *ast.BasicLit, file *ast.File, fn, typ string) *metrics.MagicNumber {
	pos := ba.fset.Position(lit.Pos())
	return &metrics.MagicNumber{
		File:     pos.Filename,
		Line:     pos.Line,
		Column:   pos.Column,
		Value:    lit.Value,
		Type:     typ,
		Context:  ba.extractContext(lit, file),
		Function: fn,
	}
}

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

func (ba *BurdenAnalyzer) checkNodeContext(n ast.Node, lit *ast.BasicLit) string {
	switch node := n.(type) {
	case *ast.AssignStmt:
		for _, rhs := range node.Rhs {
			if ba.containsNode(rhs, lit) {
				return "assignment"
			}
		}
	case *ast.ReturnStmt:
		for _, res := range node.Results {
			if ba.containsNode(res, lit) {
				return "return"
			}
		}
	case *ast.CallExpr:
		for _, arg := range node.Args {
			if ba.containsNode(arg, lit) {
				return "function_call"
			}
		}
	case *ast.BinaryExpr:
		if ba.containsNode(node, lit) {
			return "binary_expression"
		}
	}
	return ""
}

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

// DetectDeadCode identifies unreferenced unexported symbols and unreachable code
func (ba *BurdenAnalyzer) DetectDeadCode(files []*ast.File, pkg string) *metrics.DeadCodeMetrics {
	// Build symbol references across all files in package
	refs := ba.buildReferenceMap(files)

	// Find unreferenced symbols
	unreferenced := ba.findUnreferencedSymbols(files, refs, pkg)

	// Find unreachable code blocks
	unreachable := ba.findUnreachableCode(files)

	// Calculate total dead lines
	totalLines := 0
	for _, block := range unreachable {
		totalLines += block.Lines
	}

	return &metrics.DeadCodeMetrics{
		UnreferencedFunctions: unreferenced,
		UnreachableCode:       unreachable,
		TotalDeadLines:        totalLines,
		DeadCodePercent:       0.0, // Calculate in analyzer integration
	}
}

func (ba *BurdenAnalyzer) buildReferenceMap(files []*ast.File) map[string]int {
	refs := make(map[string]int)

	// Count function call references
	for _, file := range files {
		ast.Inspect(file, func(n ast.Node) bool {
			switch node := n.(type) {
			case *ast.CallExpr:
				// Direct function call
				if ident, ok := node.Fun.(*ast.Ident); ok {
					refs[ident.Name]++
				}
			}
			return true
		})
	}

	return refs
}

func (ba *BurdenAnalyzer) findUnreferencedSymbols(files []*ast.File, refs map[string]int, pkg string) []metrics.UnreferencedSymbol {
	var unreferenced []metrics.UnreferencedSymbol

	for _, file := range files {
		ast.Inspect(file, func(n ast.Node) bool {
			switch node := n.(type) {
			case *ast.FuncDecl:
				if node.Name != nil && !ast.IsExported(node.Name.Name) {
					// Check if function is referenced
					if refs[node.Name.Name] == 0 {
						pos := ba.fset.Position(node.Pos())
						unreferenced = append(unreferenced, metrics.UnreferencedSymbol{
							Name:    node.Name.Name,
							File:    pos.Filename,
							Line:    pos.Line,
							Type:    "function",
							Package: pkg,
						})
					}
				}
			}
			return true
		})
	}

	return unreferenced
}

func (ba *BurdenAnalyzer) findUnreachableCode(files []*ast.File) []metrics.UnreachableBlock {
	var unreachable []metrics.UnreachableBlock

	for _, file := range files {
		var currentFunc string

		ast.Inspect(file, func(n ast.Node) bool {
			switch node := n.(type) {
			case *ast.FuncDecl:
				if node.Name != nil {
					currentFunc = node.Name.Name
				}
				if node.Body != nil {
					blocks := ba.checkBlockForUnreachable(node.Body, currentFunc)
					unreachable = append(unreachable, blocks...)
				}
				return false
			}
			return true
		})
	}

	return unreachable
}

func (ba *BurdenAnalyzer) checkBlockForUnreachable(block *ast.BlockStmt, fn string) []metrics.UnreachableBlock {
	var unreachable []metrics.UnreachableBlock

	for i, stmt := range block.List {
		// Check if this statement is a terminating statement
		if ba.isTerminating(stmt) {
			// Check if there are statements after this
			if i+1 < len(block.List) {
				startPos := ba.fset.Position(block.List[i+1].Pos())
				endPos := ba.fset.Position(block.List[len(block.List)-1].End())

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
		unreachable = append(unreachable, ba.checkStmtForUnreachable(stmt, fn)...)
	}

	return unreachable
}

func (ba *BurdenAnalyzer) checkStmtForUnreachable(stmt ast.Stmt, fn string) []metrics.UnreachableBlock {
	var unreachable []metrics.UnreachableBlock

	switch s := stmt.(type) {
	case *ast.IfStmt:
		if s.Body != nil {
			unreachable = append(unreachable, ba.checkBlockForUnreachable(s.Body, fn)...)
		}
		if s.Else != nil {
			switch elseStmt := s.Else.(type) {
			case *ast.BlockStmt:
				unreachable = append(unreachable, ba.checkBlockForUnreachable(elseStmt, fn)...)
			case *ast.IfStmt:
				unreachable = append(unreachable, ba.checkStmtForUnreachable(elseStmt, fn)...)
			}
		}
	case *ast.ForStmt:
		if s.Body != nil {
			unreachable = append(unreachable, ba.checkBlockForUnreachable(s.Body, fn)...)
		}
	case *ast.RangeStmt:
		if s.Body != nil {
			unreachable = append(unreachable, ba.checkBlockForUnreachable(s.Body, fn)...)
		}
	case *ast.SwitchStmt:
		if s.Body != nil {
			for _, clause := range s.Body.List {
				if cc, ok := clause.(*ast.CaseClause); ok {
					unreachable = append(unreachable, ba.checkBlockForUnreachable(&ast.BlockStmt{List: cc.Body}, fn)...)
				}
			}
		}
	case *ast.TypeSwitchStmt:
		if s.Body != nil {
			for _, clause := range s.Body.List {
				if cc, ok := clause.(*ast.CaseClause); ok {
					unreachable = append(unreachable, ba.checkBlockForUnreachable(&ast.BlockStmt{List: cc.Body}, fn)...)
				}
			}
		}
	}

	return unreachable
}

func (ba *BurdenAnalyzer) isTerminating(stmt ast.Stmt) bool {
	switch s := stmt.(type) {
	case *ast.ReturnStmt:
		return true
	case *ast.ExprStmt:
		if call, ok := s.X.(*ast.CallExpr); ok {
			if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
				// Check for os.Exit
				if ident, ok := sel.X.(*ast.Ident); ok {
					if ident.Name == "os" && sel.Sel.Name == "Exit" {
						return true
					}
				}
			}
			// Check for panic
			if ident, ok := call.Fun.(*ast.Ident); ok {
				if ident.Name == "panic" {
					return true
				}
			}
		}
	}
	return false
}

func (ba *BurdenAnalyzer) getTerminationReason(stmt ast.Stmt) string {
	switch s := stmt.(type) {
	case *ast.ReturnStmt:
		return "return statement"
	case *ast.ExprStmt:
		if call, ok := s.X.(*ast.CallExpr); ok {
			if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
				if ident, ok := sel.X.(*ast.Ident); ok {
					if ident.Name == "os" && sel.Sel.Name == "Exit" {
						return "os.Exit call"
					}
				}
			}
			if ident, ok := call.Fun.(*ast.Ident); ok {
				if ident.Name == "panic" {
					return "panic call"
				}
			}
		}
	}
	return "terminating statement"
}

// AnalyzeSignatureComplexity flags functions with excessive parameters or returns
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

	return &metrics.SignatureIssue{
		Function:       fn.Name.Name,
		File:           pos.Filename,
		Line:           pos.Line,
		ParameterCount: paramCount,
		ReturnCount:    returnCount,
		BoolParams:     boolParams,
		Severity:       severity,
	}
}

func (ba *BurdenAnalyzer) countParameters(fnType *ast.FuncType) (int, []string) {
	paramCount := 0
	var boolParams []string

	if fnType.Params == nil {
		return 0, nil
	}

	for _, field := range fnType.Params.List {
		numNames := len(field.Names)
		if numNames == 0 {
			numNames = 1
		}
		paramCount += numNames

		// Check for bool parameters (flag arguments)
		if ident, ok := field.Type.(*ast.Ident); ok && ident.Name == "bool" {
			for _, name := range field.Names {
				boolParams = append(boolParams, name.Name)
			}
		}
	}

	return paramCount, boolParams
}

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

func (ba *BurdenAnalyzer) calculateSeverity(paramCount, returnCount, maxParams, maxReturns int) string {
	if paramCount > maxParams*2 || returnCount > maxReturns*2 {
		return "high"
	}
	if paramCount > maxParams || returnCount > maxReturns {
		return "medium"
	}
	return "low"
}

// DetectDeepNesting identifies functions exceeding nesting depth threshold
func (ba *BurdenAnalyzer) DetectDeepNesting(fn *ast.FuncDecl, maxNesting int) *metrics.NestingIssue {
	// TODO: Implement deep nesting detection
	return nil
}

// DetectFeatureEnvy identifies methods with excessive external references
func (ba *BurdenAnalyzer) DetectFeatureEnvy(fn *ast.FuncDecl, ratio float64) *metrics.FeatureEnvyIssue {
	// TODO: Implement feature envy detection
	return nil
}
