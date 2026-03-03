package analyzer

import (
	"go/ast"
	"go/token"
	"strings"

	"github.com/opd-ai/go-stats-generator/internal/metrics"
)

// GenericAnalyzer analyzes Go 1.18+ generic code constructs
type GenericAnalyzer struct {
	fset *token.FileSet
}

// NewGenericAnalyzer creates a new generic code analyzer
func NewGenericAnalyzer(fset *token.FileSet) *GenericAnalyzer {
	return &GenericAnalyzer{fset: fset}
}

// AnalyzeGenerics analyzes generic types and functions in a Go file
func (ga *GenericAnalyzer) AnalyzeGenerics(file *ast.File, pkgName, filePath string) (metrics.GenericMetrics, error) {
	result := metrics.GenericMetrics{
		TypeParameters: metrics.GenericTypeParameters{
			Constraints: make(map[string]int),
		},
		Instantiations:  metrics.GenericInstantiations{},
		ConstraintUsage: make(map[string]int),
	}

	// Walk the AST to collect generic information
	ast.Inspect(file, func(n ast.Node) bool {
		ga.processNode(n, filePath, &result)
		return true
	})

	// Calculate overall complexity score
	result.ComplexityScore = ga.calculateComplexity(&result.TypeParameters)

	return result, nil
}

// processNode processes individual AST nodes
func (ga *GenericAnalyzer) processNode(n ast.Node, filePath string, result *metrics.GenericMetrics) {
	switch node := n.(type) {
	case *ast.FuncDecl:
		ga.processFuncDecl(node, filePath, result)
	case *ast.TypeSpec:
		ga.processTypeSpec(node, filePath, result)
	case *ast.IndexExpr, *ast.IndexListExpr:
		ga.processInstantiation(node, filePath, result)
	}
}

// processFuncDecl analyzes generic function declarations
func (ga *GenericAnalyzer) processFuncDecl(fn *ast.FuncDecl, filePath string, result *metrics.GenericMetrics) {
	if fn.Type.TypeParams == nil {
		return
	}

	params := ga.extractTypeParams(fn.Type.TypeParams)
	result.TypeParameters.Count += len(params.Complexity)
	ga.mergeConstraints(result.TypeParameters.Constraints, params.Constraints)
	ga.mergeConstraints(result.ConstraintUsage, params.Constraints)
	result.TypeParameters.Complexity = append(result.TypeParameters.Complexity, params.Complexity...)
}

// processTypeSpec analyzes generic type declarations
func (ga *GenericAnalyzer) processTypeSpec(ts *ast.TypeSpec, filePath string, result *metrics.GenericMetrics) {
	if ts.TypeParams == nil {
		return
	}

	params := ga.extractTypeParams(ts.TypeParams)
	result.TypeParameters.Count += len(params.Complexity)
	ga.mergeConstraints(result.TypeParameters.Constraints, params.Constraints)
	ga.mergeConstraints(result.ConstraintUsage, params.Constraints)
	result.TypeParameters.Complexity = append(result.TypeParameters.Complexity, params.Complexity...)
}

// extractTypeParams extracts type parameter info
func (ga *GenericAnalyzer) extractTypeParams(fieldList *ast.FieldList) metrics.GenericTypeParameters {
	params := metrics.GenericTypeParameters{
		Constraints: make(map[string]int),
	}

	if fieldList == nil {
		return params
	}

	for _, field := range fieldList.List {
		constraintName := ga.extractConstraint(field.Type)
		params.Constraints[constraintName]++

		for _, name := range field.Names {
			complexity := metrics.GenericComplexity{
				Name:            name.Name,
				ParameterCount:  1,
				ConstraintCount: ga.countConstraints(field.Type),
				ComplexityScore: ga.scoreConstraint(field.Type),
			}
			params.Complexity = append(params.Complexity, complexity)
		}
	}

	return params
}

// extractConstraint extracts constraint name from type
func (ga *GenericAnalyzer) extractConstraint(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.SelectorExpr:
		return ga.selectorName(t)
	case *ast.InterfaceType:
		return "interface{}"
	case *ast.IndexExpr:
		return ga.extractConstraint(t.X)
	default:
		return "any"
	}
}

// selectorName builds name from selector
func (ga *GenericAnalyzer) selectorName(sel *ast.SelectorExpr) string {
	var parts []string
	if id, ok := sel.X.(*ast.Ident); ok {
		parts = append(parts, id.Name)
	}
	parts = append(parts, sel.Sel.Name)
	return strings.Join(parts, ".")
}

// countConstraints counts constraint complexity
func (ga *GenericAnalyzer) countConstraints(expr ast.Expr) int {
	if iface, ok := expr.(*ast.InterfaceType); ok {
		return len(iface.Methods.List)
	}
	return 1
}

// scoreConstraint calculates constraint complexity
func (ga *GenericAnalyzer) scoreConstraint(expr ast.Expr) float64 {
	count := float64(ga.countConstraints(expr))
	if count == 0 {
		return 1.0
	}
	return count
}

// processInstantiation tracks generic instantiations
func (ga *GenericAnalyzer) processInstantiation(n ast.Node, filePath string, result *metrics.GenericMetrics) {
	var inst metrics.GenericInstantiation
	inst.File = filePath

	switch node := n.(type) {
	case *ast.IndexExpr:
		inst.GenericName = ga.exprName(node.X)
		inst.TypeArgs = []string{ga.exprName(node.Index)}
		inst.Line = ga.fset.Position(node.Pos()).Line
	case *ast.IndexListExpr:
		inst.GenericName = ga.exprName(node.X)
		for _, idx := range node.Indices {
			inst.TypeArgs = append(inst.TypeArgs, ga.exprName(idx))
		}
		inst.Line = ga.fset.Position(node.Pos()).Line
	default:
		return
	}

	result.Instantiations.Functions = append(result.Instantiations.Functions, inst)
}

// exprName extracts name from expression
func (ga *GenericAnalyzer) exprName(expr ast.Expr) string {
	switch e := expr.(type) {
	case *ast.Ident:
		return e.Name
	case *ast.SelectorExpr:
		return ga.selectorName(e)
	default:
		return "unknown"
	}
}

// calculateComplexity calculates overall score
func (ga *GenericAnalyzer) calculateComplexity(params *metrics.GenericTypeParameters) float64 {
	if len(params.Complexity) == 0 {
		return 0.0
	}

	total := 0.0
	for _, c := range params.Complexity {
		total += c.ComplexityScore
	}
	return total / float64(len(params.Complexity))
}

// mergeConstraints merges constraint maps
func (ga *GenericAnalyzer) mergeConstraints(dest, src map[string]int) {
	for k, v := range src {
		dest[k] += v
	}
}
