package analyzer

import (
	"go/ast"
	"go/token"
	"strings"

	"github.com/opd-ai/go-stats-generator/internal/metrics"
)

// InterfaceAnalyzer analyzes interface declarations in Go source code
// It analyzes method signatures, embedded interfaces, implementation tracking,
// and calculates complexity metrics to understand interface design patterns and usage.
type InterfaceAnalyzer struct {
	fset *token.FileSet
	// Implementation tracking across files
	typeImplementations  map[string][]string           // interface -> []implementer
	interfaceDefinitions map[string]*ast.InterfaceType // interface name -> definition
	structDefinitions    map[string]*ast.StructType    // struct name -> definition
}

// NewInterfaceAnalyzer creates a new interface analyzer
func NewInterfaceAnalyzer(fset *token.FileSet) *InterfaceAnalyzer {
	return &InterfaceAnalyzer{
		fset:                 fset,
		typeImplementations:  make(map[string][]string),
		interfaceDefinitions: make(map[string]*ast.InterfaceType),
		structDefinitions:    make(map[string]*ast.StructType),
	}
}

// AnalyzeInterfaces analyzes all interface declarations in an AST file
func (ia *InterfaceAnalyzer) AnalyzeInterfaces(file *ast.File, pkgName string) ([]metrics.InterfaceMetrics, error) {
	var interfaces []metrics.InterfaceMetrics

	// First pass: collect all type definitions for implementation analysis
	ia.collectTypeDefinitions(file, pkgName)

	// Second pass: analyze interfaces
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

	// Third pass: analyze method implementations and calculate ratios
	ia.analyzeImplementations(file, pkgName)

	// Update interface metrics with implementation data
	for i := range interfaces {
		ia.updateImplementationMetrics(&interfaces[i])
	}

	return interfaces, nil
}

// collectTypeDefinitions collects all type definitions for implementation analysis
func (ia *InterfaceAnalyzer) collectTypeDefinitions(file *ast.File, pkgName string) {
	for _, decl := range file.Decls {
		if genDecl, ok := decl.(*ast.GenDecl); ok && genDecl.Tok == token.TYPE {
			for _, spec := range genDecl.Specs {
				if typeSpec, ok := spec.(*ast.TypeSpec); ok {
					typeName := pkgName + "." + typeSpec.Name.Name

					switch t := typeSpec.Type.(type) {
					case *ast.InterfaceType:
						ia.interfaceDefinitions[typeName] = t
					case *ast.StructType:
						ia.structDefinitions[typeName] = t
					}
				}
			}
		}
	}
}

// analyzeImplementations finds which types implement which interfaces
func (ia *InterfaceAnalyzer) analyzeImplementations(file *ast.File, pkgName string) {
	// Find all method declarations
	methodsByType := make(map[string][]string)

	for _, decl := range file.Decls {
		if funcDecl, ok := decl.(*ast.FuncDecl); ok && funcDecl.Recv != nil {
			// This is a method
			receiverType := ia.extractReceiverType(funcDecl.Recv)
			if receiverType != "" {
				fullTypeName := pkgName + "." + receiverType
				methodsByType[fullTypeName] = append(methodsByType[fullTypeName], funcDecl.Name.Name)
			}
		}
	}

	// Check which types implement which interfaces
	for interfaceName, interfaceType := range ia.interfaceDefinitions {
		requiredMethods := ia.extractInterfaceMethods(interfaceType)

		for typeName, typeMethods := range methodsByType {
			if ia.implementsInterface(typeMethods, requiredMethods) {
				ia.typeImplementations[interfaceName] = append(ia.typeImplementations[interfaceName], typeName)
			}
		}
	}
}

// extractReceiverType extracts the type name from a method receiver
func (ia *InterfaceAnalyzer) extractReceiverType(recv *ast.FieldList) string {
	if recv == nil || len(recv.List) == 0 {
		return ""
	}

	field := recv.List[0]
	switch t := field.Type.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		if ident, ok := t.X.(*ast.Ident); ok {
			return ident.Name
		}
	}
	return ""
}

// extractInterfaceMethods extracts method names from an interface
func (ia *InterfaceAnalyzer) extractInterfaceMethods(interfaceType *ast.InterfaceType) []string {
	var methods []string

	if interfaceType.Methods != nil {
		for _, field := range interfaceType.Methods.List {
			if field.Names != nil {
				// Regular method
				for _, name := range field.Names {
					methods = append(methods, name.Name)
				}
			}
		}
	}

	return methods
}

// implementsInterface checks if a type implements all methods of an interface
func (ia *InterfaceAnalyzer) implementsInterface(typeMethods, requiredMethods []string) bool {
	if len(requiredMethods) == 0 {
		return false
	}

	methodSet := make(map[string]bool)
	for _, method := range typeMethods {
		methodSet[method] = true
	}

	for _, required := range requiredMethods {
		if !methodSet[required] {
			return false
		}
	}

	return true
}

// updateImplementationMetrics updates interface metrics with implementation data
func (ia *InterfaceAnalyzer) updateImplementationMetrics(interfaceMetric *metrics.InterfaceMetrics) {
	interfaceName := interfaceMetric.Package + "." + interfaceMetric.Name

	if implementations, exists := ia.typeImplementations[interfaceName]; exists {
		interfaceMetric.Implementations = implementations
		interfaceMetric.ImplementationCount = len(implementations)

		// Calculate implementation ratio (implementations per method)
		if interfaceMetric.MethodCount > 0 {
			interfaceMetric.ImplementationRatio = float64(len(implementations)) / float64(interfaceMetric.MethodCount)
		}
	}
}

// analyzeInterface analyzes a single interface declaration
func (ia *InterfaceAnalyzer) analyzeInterface(typeSpec *ast.TypeSpec, interfaceType *ast.InterfaceType, fileName, pkgName string, doc *ast.CommentGroup) (metrics.InterfaceMetrics, error) {
	pos := ia.fset.Position(typeSpec.Pos())

	interfaceMetric := metrics.InterfaceMetrics{
		Name:               typeSpec.Name.Name,
		Package:            pkgName,
		File:               fileName,
		Line:               pos.Line,
		IsExported:         ast.IsExported(typeSpec.Name.Name),
		Methods:            make([]metrics.InterfaceMethod, 0),
		EmbeddedInterfaces: make([]string, 0),
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

	// Calculate totals and complexity
	interfaceMetric.MethodCount = len(interfaceMetric.Methods)
	interfaceMetric.EmbeddingDepth = ia.calculateEmbeddingDepth(interfaceMetric.EmbeddedInterfaces)
	interfaceMetric.ComplexityScore = ia.calculateInterfaceComplexity(interfaceMetric)

	// Analyze documentation
	interfaceMetric.Documentation = ia.analyzeDocumentation(doc)

	return interfaceMetric, nil
}

// calculateEmbeddingDepth calculates the depth of interface embedding
func (ia *InterfaceAnalyzer) calculateEmbeddingDepth(embeddedInterfaces []string) int {
	if len(embeddedInterfaces) == 0 {
		return 0
	}

	maxDepth := 1
	for _, embedded := range embeddedInterfaces {
		// This is a simplified calculation - in a full implementation,
		// we would recursively check embedded interfaces for their own embeddings
		if strings.Contains(embedded, ".") {
			maxDepth = 2 // External interface embedding
		}
	}

	return maxDepth
}

// calculateInterfaceComplexity calculates overall interface complexity
func (ia *InterfaceAnalyzer) calculateInterfaceComplexity(interfaceMetric metrics.InterfaceMetrics) float64 {
	complexity := 0.5 // Base complexity

	// Add complexity for methods
	complexity += float64(interfaceMetric.MethodCount) * 0.3

	// Add complexity for embedded interfaces
	complexity += float64(len(interfaceMetric.EmbeddedInterfaces)) * 0.4

	// Add complexity for embedding depth
	complexity += float64(interfaceMetric.EmbeddingDepth) * 0.2

	// Add complexity based on method signatures
	for _, method := range interfaceMetric.Methods {
		complexity += method.Signature.ComplexityScore * 0.1
	}

	return complexity
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
