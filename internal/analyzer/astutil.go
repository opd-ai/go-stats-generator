package analyzer

import (
	"go/ast"
	"go/token"
)

// CollectFunctions extracts all function declarations from an AST file.
// CollectFunctions traverses the entire AST tree and returns a slice of FuncDecl pointers.
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

// CollectTypes extracts all type declarations from an AST file.
// CollectTypes inspects GenDecl nodes with TYPE token and returns all TypeSpec found.
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

// ExtractReceiverType extracts the type name from a receiver expression.
// ExtractReceiverType handles pointer receivers, generic types, and simple identifiers.
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

// IsMethod checks if a FuncDecl is a method by verifying it has a receiver.
func IsMethod(fn *ast.FuncDecl) bool {
	return fn.Recv != nil && len(fn.Recv.List) > 0
}

// GetMethodReceiverType returns the receiver type for a method, empty string for functions
func GetMethodReceiverType(fn *ast.FuncDecl) string {
	if !IsMethod(fn) {
		return ""
	}
	return ExtractReceiverType(fn.Recv.List[0].Type)
}

// CountNodes counts the number of AST nodes in a statement list.
// CountNodes recursively inspects each statement and counts all non-nil nodes.
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
