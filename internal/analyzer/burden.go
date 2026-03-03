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
	// TODO: Implement dead code detection
	return nil
}

// AnalyzeSignatureComplexity flags functions with excessive parameters or returns
func (ba *BurdenAnalyzer) AnalyzeSignatureComplexity(fn *ast.FuncDecl, maxParams, maxReturns int) *metrics.SignatureIssue {
	// TODO: Implement signature complexity analysis
	return nil
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
