package analyzer

import (
	"go/ast"
	"go/token"
)

// CollectFunctions extracts all function declarations from an AST file by traversing
// the entire abstract syntax tree. It returns a slice of FuncDecl pointers representing
// both standalone functions and methods. This utility is commonly used for function-level
// analysis, complexity calculation, and documentation coverage assessment.
func CollectFunctions(file *ast.File) []*ast.FuncDecl {
	var funcs []*ast.FuncDecl
	ast.Inspect(file, func(n ast.Node) bool {
		if fn, ok := n.(*ast.FuncDecl); ok {
			funcs = append(funcs, fn)
		}
		return true
	})
	return funcs
}

// CollectTypes extracts all type declarations from an AST file by inspecting GenDecl nodes
// with TYPE tokens. It returns a slice of TypeSpec pointers representing struct, interface,
// and type alias declarations. This function is essential for structural analysis, dependency
// tracking, and design pattern detection in Go codebases.
func CollectTypes(file *ast.File) []*ast.TypeSpec {
	var types []*ast.TypeSpec
	ast.Inspect(file, func(n ast.Node) bool {
		if genDecl, ok := n.(*ast.GenDecl); ok && genDecl.Tok == token.TYPE {
			for _, spec := range genDecl.Specs {
				if typeSpec, ok := spec.(*ast.TypeSpec); ok {
					types = append(types, typeSpec)
				}
			}
		}
		return true
	})
	return types
}

// ExtractReceiverType extracts the type name from a method receiver expression, handling
// pointer receivers (*Type), generic types (Type[T]), and simple identifiers. This function
// is critical for method-to-type association, calculating method counts per struct, and
// analyzing object-oriented design patterns in Go code. Returns empty string for invalid expressions.
func ExtractReceiverType(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		if ident, ok := t.X.(*ast.Ident); ok {
			return ident.Name
		}
	case *ast.IndexExpr:
		return ExtractReceiverType(t.X)
	case *ast.IndexListExpr:
		return ExtractReceiverType(t.X)
	}
	return ""
}

// IsMethod checks if a FuncDecl represents a method by verifying it has a receiver field.
// Methods in Go are functions with a receiver parameter, distinguishing them from standalone
// functions. This check is essential for accurate method counting, struct analysis, and
// understanding code organization patterns.
func IsMethod(fn *ast.FuncDecl) bool {
	return fn.Recv != nil && len(fn.Recv.List) > 0
}

// GetMethodReceiverType returns the receiver type name for a method declaration, or an empty
// string for standalone functions. This function combines receiver validation with type extraction,
// simplifying method-to-struct association logic in analyzers. It supports pointer receivers,
// generic types, and value receivers.
func GetMethodReceiverType(fn *ast.FuncDecl) string {
	if !IsMethod(fn) {
		return ""
	}
	return ExtractReceiverType(fn.Recv.List[0].Type)
}

// CountNodes counts the total number of AST nodes in a statement list by recursively inspecting
// each statement and its children. This metric is used for structural complexity analysis,
// code duplication detection, and estimating the cognitive burden of code blocks. The count
// includes all non-nil nodes encountered during tree traversal.
func CountNodes(stmts []ast.Stmt) int {
	count := 0
	for _, stmt := range stmts {
		ast.Inspect(stmt, func(n ast.Node) bool {
			if n != nil {
				count++
			}
			return true
		})
	}
	return count
}
