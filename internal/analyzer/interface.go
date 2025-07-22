package analyzer

import (
	"go/ast"
	"go/token"
	"strings"

	"github.com/opd-ai/go-stats-generator/internal/metrics"
)

// InterfaceAnalyzer analyzes interface declarations in Go source code
// It analyzes method signatures, embedded interfaces, and calculates
// complexity metrics to understand interface design patterns and usage.
type InterfaceAnalyzer struct {
	fset *token.FileSet
}

// NewInterfaceAnalyzer creates a new interface analyzer
func NewInterfaceAnalyzer(fset *token.FileSet) *InterfaceAnalyzer {
	return &InterfaceAnalyzer{
		fset: fset,
	}
}

// AnalyzeInterfaces analyzes all interface declarations in an AST file
func (ia *InterfaceAnalyzer) AnalyzeInterfaces(file *ast.File, pkgName string) ([]metrics.InterfaceMetrics, error) {
	var interfaces []metrics.InterfaceMetrics

	// Find all interface declarations
	for _, decl := range file.Decls {
		if genDecl, ok := decl.(*ast.GenDecl); ok && genDecl.Tok == token.TYPE {
			for _, spec := range genDecl.Specs {
				if typeSpec, ok := spec.(*ast.TypeSpec); ok {
					if interfaceType, ok := typeSpec.Type.(*ast.InterfaceType); ok {
						interfaceMetric, err := ia.analyzeInterface(typeSpec, interfaceType, file.Name.Name, pkgName, genDecl.Doc)
						if err != nil {
							continue // Log warning and continue
						}
						interfaces = append(interfaces, interfaceMetric)
					}
				}
			}
		}
	}

	return interfaces, nil
}

// analyzeInterface analyzes a single interface declaration
func (ia *InterfaceAnalyzer) analyzeInterface(typeSpec *ast.TypeSpec, interfaceType *ast.InterfaceType, fileName, pkgName string, doc *ast.CommentGroup) (metrics.InterfaceMetrics, error) {
	pos := ia.fset.Position(typeSpec.Pos())

	interfaceMetric := metrics.InterfaceMetrics{
		Name:       typeSpec.Name.Name,
		Package:    pkgName,
		File:       fileName,
		Line:       pos.Line,
		IsExported: ast.IsExported(typeSpec.Name.Name),
		Methods:    make([]metrics.InterfaceMethod, 0),
	}

	// Analyze interface methods and embedded interfaces
	if interfaceType.Methods != nil {
		for _, field := range interfaceType.Methods.List {
			if field.Names == nil {
				// Embedded interface
				embeddedName := ia.extractEmbeddedInterfaceName(field.Type)
				if embeddedName != "" {
					interfaceMetric.EmbeddedInterfaces = append(interfaceMetric.EmbeddedInterfaces, embeddedName)
				}
			} else {
				// Regular method
				for _, name := range field.Names {
					method := ia.analyzeInterfaceMethod(name, field.Type)
					interfaceMetric.Methods = append(interfaceMetric.Methods, method)
				}
			}
		}
	}

	// Calculate totals
	interfaceMetric.MethodCount = len(interfaceMetric.Methods)

	// Analyze documentation
	interfaceMetric.Documentation = ia.analyzeDocumentation(doc)

	return interfaceMetric, nil
}

// analyzeInterfaceMethod analyzes a method in an interface
func (ia *InterfaceAnalyzer) analyzeInterfaceMethod(name *ast.Ident, methodType ast.Expr) metrics.InterfaceMethod {
	method := metrics.InterfaceMethod{
		Name: name.Name,
	}

	// Analyze method signature
	if funcType, ok := methodType.(*ast.FuncType); ok {
		method.Signature = ia.analyzeFunctionSignature(funcType)
	}

	return method
}

// analyzeFunctionSignature analyzes the signature complexity of a function type
func (ia *InterfaceAnalyzer) analyzeFunctionSignature(funcType *ast.FuncType) metrics.FunctionSignature {
	signature := metrics.FunctionSignature{}

	// Count parameters
	if funcType.Params != nil {
		for _, field := range funcType.Params.List {
			if len(field.Names) == 0 {
				// Unnamed parameter (like in interface methods)
				signature.ParameterCount++
			} else {
				signature.ParameterCount += len(field.Names)
			}

			// Check for variadic parameters
			if _, isEllipsis := field.Type.(*ast.Ellipsis); isEllipsis {
				signature.VariadicUsage = true
			}

			// Check for interface parameters
			if ia.isInterfaceType(field.Type) {
				signature.InterfaceParams++
			}
		}
	}

	// Count return values
	if funcType.Results != nil {
		for _, field := range funcType.Results.List {
			if len(field.Names) == 0 {
				signature.ReturnCount++
			} else {
				signature.ReturnCount += len(field.Names)
			}

			// Check if returns error
			if ia.isErrorType(field.Type) {
				signature.ErrorReturn = true
			}
		}
	}

	// Calculate signature complexity
	signature.ComplexityScore = ia.calculateSignatureComplexity(signature)

	return signature
}

// isInterfaceType checks if a type expression represents an interface
func (ia *InterfaceAnalyzer) isInterfaceType(expr ast.Expr) bool {
	switch t := expr.(type) {
	case *ast.InterfaceType:
		return true
	case *ast.Ident:
		// Common interface types
		return t.Name == "interface{}" || strings.HasSuffix(t.Name, "er")
	default:
		return false
	}
}

// isErrorType checks if a type expression represents the error interface
func (ia *InterfaceAnalyzer) isErrorType(expr ast.Expr) bool {
	if ident, ok := expr.(*ast.Ident); ok {
		return ident.Name == "error"
	}
	return false
}

// calculateSignatureComplexity calculates complexity score for a function signature
func (ia *InterfaceAnalyzer) calculateSignatureComplexity(signature metrics.FunctionSignature) float64 {
	complexity := 0.3 // Base complexity

	// Add complexity based on parameters
	complexity += float64(signature.ParameterCount) * 0.2

	// Add complexity based on return values
	complexity += float64(signature.ReturnCount) * 0.3

	// Add complexity for variadic parameters
	if signature.VariadicUsage {
		complexity += 0.2
	}

	// Add complexity for interface parameters
	complexity += float64(signature.InterfaceParams) * 0.3

	// Add complexity for error returns
	if signature.ErrorReturn {
		complexity += 0.3
	}

	return complexity
}

// extractEmbeddedInterfaceName extracts the name of an embedded interface
func (ia *InterfaceAnalyzer) extractEmbeddedInterfaceName(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.SelectorExpr:
		// pkg.Interface
		if pkgIdent, ok := t.X.(*ast.Ident); ok {
			return pkgIdent.Name + "." + t.Sel.Name
		}
	}
	return ""
}

// analyzeDocumentation analyzes interface documentation quality
func (ia *InterfaceAnalyzer) analyzeDocumentation(doc *ast.CommentGroup) metrics.DocumentationInfo {
	docInfo := metrics.DocumentationInfo{}

	if doc == nil {
		return docInfo
	}

	docInfo.HasComment = true

	// Combine all comment lines
	var docText strings.Builder
	for _, comment := range doc.List {
		docText.WriteString(comment.Text)
		docText.WriteString(" ")
	}

	text := docText.String()
	docInfo.CommentLength = len(text)

	// Check for code examples (simple heuristic)
	docInfo.HasExample = strings.Contains(text, "Example") ||
		strings.Contains(text, "example") ||
		strings.Contains(text, "//")

	// Calculate quality score
	docInfo.QualityScore = ia.calculateDocQualityScore(text)

	return docInfo
}

// calculateDocQualityScore calculates documentation quality score
func (ia *InterfaceAnalyzer) calculateDocQualityScore(docText string) float64 {
	if len(docText) == 0 {
		return 0.0
	}

	score := 0.0

	// Base score for having documentation
	score += 0.3

	// Length-based scoring
	if len(docText) > 50 {
		score += 0.2
	}
	if len(docText) > 100 {
		score += 0.2
	}

	// Content quality indicators
	if strings.Contains(strings.ToLower(docText), "interface") ||
		strings.Contains(strings.ToLower(docText), "contract") ||
		strings.Contains(strings.ToLower(docText), "behavior") {
		score += 0.2
	}

	// Example presence
	if strings.Contains(strings.ToLower(docText), "example") {
		score += 0.1
	}

	// Cap at 1.0
	if score > 1.0 {
		score = 1.0
	}

	return score
}
