package analyzer

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"

	"github.com/opd-ai/go-stats-generator/internal/metrics"
)

func TestNewInterfaceAnalyzer(t *testing.T) {
	fset := token.NewFileSet()
	analyzer := NewInterfaceAnalyzer(fset)

	if analyzer == nil {
		t.Fatal("NewInterfaceAnalyzer returned nil")
	}

	if analyzer.fset != fset {
		t.Error("InterfaceAnalyzer fset not set correctly")
	}
}

func TestAnalyzeInterfaces_EmptyFile(t *testing.T) {
	source := `package test`

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", source, parser.ParseComments)
	if err != nil {
		t.Fatalf("Failed to parse source: %v", err)
	}

	analyzer := NewInterfaceAnalyzer(fset)
	interfaces, err := analyzer.AnalyzeInterfaces(file, "test")

	if err != nil {
		t.Fatalf("AnalyzeInterfaces failed: %v", err)
	}

	if len(interfaces) != 0 {
		t.Errorf("Expected 0 interfaces, got %d", len(interfaces))
	}
}

func TestAnalyzeInterfaces_SimpleInterface(t *testing.T) {
	source := `package test

// Writer represents something that can write data
type Writer interface {
	Write(data []byte) (int, error)
	Close() error
}`

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", source, parser.ParseComments)
	if err != nil {
		t.Fatalf("Failed to parse source: %v", err)
	}

	analyzer := NewInterfaceAnalyzer(fset)
	interfaces, err := analyzer.AnalyzeInterfaces(file, "test")

	if err != nil {
		t.Fatalf("AnalyzeInterfaces failed: %v", err)
	}

	if len(interfaces) != 1 {
		t.Fatalf("Expected 1 interface, got %d", len(interfaces))
	}

	writer := interfaces[0]

	// Test basic fields
	if writer.Name != "Writer" {
		t.Errorf("Expected name 'Writer', got '%s'", writer.Name)
	}

	if writer.Package != "test" {
		t.Errorf("Expected package 'test', got '%s'", writer.Package)
	}

	if !writer.IsExported {
		t.Error("Expected Writer to be exported")
	}

	if writer.MethodCount != 2 {
		t.Errorf("Expected 2 methods, got %d", writer.MethodCount)
	}

	// Test methods
	if len(writer.Methods) != 2 {
		t.Fatalf("Expected 2 methods in slice, got %d", len(writer.Methods))
	}

	// Check Write method
	writeMethod := writer.Methods[0]
	if writeMethod.Name != "Write" {
		t.Errorf("Expected first method 'Write', got '%s'", writeMethod.Name)
	}

	if writeMethod.Signature.ParameterCount != 1 {
		t.Errorf("Expected Write to have 1 parameter, got %d", writeMethod.Signature.ParameterCount)
	}

	if writeMethod.Signature.ReturnCount != 2 {
		t.Errorf("Expected Write to have 2 return values, got %d", writeMethod.Signature.ReturnCount)
	}

	if !writeMethod.Signature.ErrorReturn {
		t.Error("Expected Write to return error")
	}

	// Check Close method
	closeMethod := writer.Methods[1]
	if closeMethod.Name != "Close" {
		t.Errorf("Expected second method 'Close', got '%s'", closeMethod.Name)
	}

	if closeMethod.Signature.ParameterCount != 0 {
		t.Errorf("Expected Close to have 0 parameters, got %d", closeMethod.Signature.ParameterCount)
	}

	if closeMethod.Signature.ReturnCount != 1 {
		t.Errorf("Expected Close to have 1 return value, got %d", closeMethod.Signature.ReturnCount)
	}

	if !closeMethod.Signature.ErrorReturn {
		t.Error("Expected Close to return error")
	}

	// Test documentation
	if !writer.Documentation.HasComment {
		t.Error("Expected interface to have documentation")
	}

	if writer.Documentation.QualityScore <= 0 {
		t.Error("Expected positive documentation quality score")
	}
}

func TestAnalyzeInterfaces_EmbeddedInterface(t *testing.T) {
	source := `package test

import "io"

// ReadCloser combines reading and closing capabilities
type ReadCloser interface {
	io.Reader       // Embedded external interface
	io.Closer       // Another embedded interface
	Flush() error   // Additional method
}`

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", source, parser.ParseComments)
	if err != nil {
		t.Fatalf("Failed to parse source: %v", err)
	}

	analyzer := NewInterfaceAnalyzer(fset)
	interfaces, err := analyzer.AnalyzeInterfaces(file, "test")

	if err != nil {
		t.Fatalf("AnalyzeInterfaces failed: %v", err)
	}

	if len(interfaces) != 1 {
		t.Fatalf("Expected 1 interface, got %d", len(interfaces))
	}

	readCloser := interfaces[0]

	// Test embedded interfaces
	if len(readCloser.EmbeddedInterfaces) != 2 {
		t.Errorf("Expected 2 embedded interfaces, got %d", len(readCloser.EmbeddedInterfaces))
	}

	embeddedNames := make(map[string]bool)
	for _, embedded := range readCloser.EmbeddedInterfaces {
		embeddedNames[embedded] = true
	}

	if !embeddedNames["io.Reader"] {
		t.Error("Expected io.Reader to be embedded")
	}

	if !embeddedNames["io.Closer"] {
		t.Error("Expected io.Closer to be embedded")
	}

	// Test that we still have the explicit method
	if readCloser.MethodCount != 1 {
		t.Errorf("Expected 1 explicit method, got %d", readCloser.MethodCount)
	}

	if len(readCloser.Methods) != 1 {
		t.Fatalf("Expected 1 method in slice, got %d", len(readCloser.Methods))
	}

	flushMethod := readCloser.Methods[0]
	if flushMethod.Name != "Flush" {
		t.Errorf("Expected method 'Flush', got '%s'", flushMethod.Name)
	}
}

func TestAnalyzeInterfaces_ComplexInterface(t *testing.T) {
	source := `package test

// Handler demonstrates complex method signatures
type Handler interface {
	// Simple method
	Start() error
	
	// Method with multiple parameters
	Process(ctx interface{}, data []byte, options ...string) (interface{}, error)
	
	// Method with function parameter
	WithCallback(callback func(error)) Handler
	
	// Method with interface parameter
	SetLogger(logger interface{}) 
}`

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", source, parser.ParseComments)
	if err != nil {
		t.Fatalf("Failed to parse source: %v", err)
	}

	analyzer := NewInterfaceAnalyzer(fset)
	interfaces, err := analyzer.AnalyzeInterfaces(file, "test")

	if err != nil {
		t.Fatalf("AnalyzeInterfaces failed: %v", err)
	}

	if len(interfaces) != 1 {
		t.Fatalf("Expected 1 interface, got %d", len(interfaces))
	}

	handler := interfaces[0]

	if handler.MethodCount != 4 {
		t.Errorf("Expected 4 methods, got %d", handler.MethodCount)
	}

	// Test Process method with variadic parameters
	var processMethod *metrics.InterfaceMethod
	for _, method := range handler.Methods {
		if method.Name == "Process" {
			processMethod = &method
			break
		}
	}

	if processMethod == nil {
		t.Fatal("Process method not found")
	}

	if processMethod.Signature.ParameterCount != 3 {
		t.Errorf("Expected Process to have 3 parameters, got %d", processMethod.Signature.ParameterCount)
	}

	if !processMethod.Signature.VariadicUsage {
		t.Error("Expected Process to have variadic parameter")
	}

	if processMethod.Signature.InterfaceParams != 1 {
		t.Errorf("Expected Process to have 1 interface parameter, got %d", processMethod.Signature.InterfaceParams)
	}

	if !processMethod.Signature.ErrorReturn {
		t.Error("Expected Process to return error")
	}

	// Test signature complexity calculation
	if processMethod.Signature.ComplexityScore <= 1.0 {
		t.Error("Expected Process to have high signature complexity")
	}
}

func TestAnalyzeInterfaces_EmptyInterface(t *testing.T) {
	source := `package test

// Any represents any value
type Any interface{}`

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", source, parser.ParseComments)
	if err != nil {
		t.Fatalf("Failed to parse source: %v", err)
	}

	analyzer := NewInterfaceAnalyzer(fset)
	interfaces, err := analyzer.AnalyzeInterfaces(file, "test")

	if err != nil {
		t.Fatalf("AnalyzeInterfaces failed: %v", err)
	}

	if len(interfaces) != 1 {
		t.Fatalf("Expected 1 interface, got %d", len(interfaces))
	}

	any := interfaces[0]

	if any.MethodCount != 0 {
		t.Errorf("Expected 0 methods, got %d", any.MethodCount)
	}

	if len(any.Methods) != 0 {
		t.Errorf("Expected 0 methods in slice, got %d", len(any.Methods))
	}

	if len(any.EmbeddedInterfaces) != 0 {
		t.Errorf("Expected 0 embedded interfaces, got %d", len(any.EmbeddedInterfaces))
	}
}

func TestIsInterfaceType(t *testing.T) {
	fset := token.NewFileSet()
	analyzer := NewInterfaceAnalyzer(fset)

	tests := []struct {
		source      string
		expected    bool
		description string
	}{
		{"interface{}", true, "empty interface"},
		{"io.Reader", false, "named interface (we can't know for sure from type alone)"},
		{"Writer", true, "type ending in 'er' (common interface pattern)"},
		{"Handler", true, "type ending in 'er'"},
		{"string", false, "primitive type"},
		{"*int", false, "pointer type"},
	}

	for _, test := range tests {
		source := `package test
type TestInterface interface {
	Method(param ` + test.source + `)
}`

		file, err := parser.ParseFile(fset, "test.go", source, 0)
		if err != nil {
			t.Errorf("Failed to parse %s: %v", test.description, err)
			continue
		}

		// Extract the parameter type
		genDecl := file.Decls[0].(*ast.GenDecl)
		typeSpec := genDecl.Specs[0].(*ast.TypeSpec)
		interfaceType := typeSpec.Type.(*ast.InterfaceType)
		method := interfaceType.Methods.List[0]
		funcType := method.Type.(*ast.FuncType)
		paramType := funcType.Params.List[0].Type

		result := analyzer.isInterfaceType(paramType)
		if result != test.expected {
			t.Errorf("For %s: expected %v, got %v", test.description, test.expected, result)
		}
	}
}

func TestIsErrorType(t *testing.T) {
	fset := token.NewFileSet()
	analyzer := NewInterfaceAnalyzer(fset)

	tests := []struct {
		source      string
		expected    bool
		description string
	}{
		{"error", true, "error type"},
		{"string", false, "string type"},
		{"int", false, "int type"},
		{"MyError", false, "custom error type"},
	}

	for _, test := range tests {
		source := `package test
type TestInterface interface {
	Method() ` + test.source + `
}`

		file, err := parser.ParseFile(fset, "test.go", source, 0)
		if err != nil {
			t.Errorf("Failed to parse %s: %v", test.description, err)
			continue
		}

		// Extract the return type
		genDecl := file.Decls[0].(*ast.GenDecl)
		typeSpec := genDecl.Specs[0].(*ast.TypeSpec)
		interfaceType := typeSpec.Type.(*ast.InterfaceType)
		method := interfaceType.Methods.List[0]
		funcType := method.Type.(*ast.FuncType)
		returnType := funcType.Results.List[0].Type

		result := analyzer.isErrorType(returnType)
		if result != test.expected {
			t.Errorf("For %s: expected %v, got %v", test.description, test.expected, result)
		}
	}
}

func TestCalculateSignatureComplexity(t *testing.T) {
	fset := token.NewFileSet()
	analyzer := NewInterfaceAnalyzer(fset)

	tests := []struct {
		signature   metrics.FunctionSignature
		minExpected float64
		description string
	}{
		{
			signature:   metrics.FunctionSignature{},
			minExpected: 0.3,
			description: "empty signature",
		},
		{
			signature: metrics.FunctionSignature{
				ParameterCount: 2,
				ReturnCount:    1,
				ErrorReturn:    true,
			},
			minExpected: 1.0,
			description: "simple signature with error",
		},
		{
			signature: metrics.FunctionSignature{
				ParameterCount:  3,
				ReturnCount:     2,
				VariadicUsage:   true,
				ErrorReturn:     true,
				InterfaceParams: 1,
			},
			minExpected: 2.0,
			description: "complex signature",
		},
	}

	for _, test := range tests {
		complexity := analyzer.calculateSignatureComplexity(test.signature)
		if complexity < test.minExpected {
			t.Errorf("For %s: expected complexity >= %.1f, got %.1f",
				test.description, test.minExpected, complexity)
		}
	}
}

func TestExtractEmbeddedInterfaceName(t *testing.T) {
	fset := token.NewFileSet()
	analyzer := NewInterfaceAnalyzer(fset)

	tests := []struct {
		source      string
		expected    string
		description string
	}{
		{"Reader", "Reader", "simple interface name"},
		{"io.Reader", "io.Reader", "qualified interface name"},
	}

	for _, test := range tests {
		source := `package test
type TestInterface interface {
	` + test.source + `
}`

		file, err := parser.ParseFile(fset, "test.go", source, 0)
		if err != nil {
			t.Errorf("Failed to parse %s: %v", test.description, err)
			continue
		}

		// Extract the embedded interface
		genDecl := file.Decls[0].(*ast.GenDecl)
		typeSpec := genDecl.Specs[0].(*ast.TypeSpec)
		interfaceType := typeSpec.Type.(*ast.InterfaceType)
		embeddedType := interfaceType.Methods.List[0].Type

		result := analyzer.extractEmbeddedInterfaceName(embeddedType)
		if result != test.expected {
			t.Errorf("For %s: expected '%s', got '%s'", test.description, test.expected, result)
		}
	}
}

func TestAnalyzeInterfaces_Integration(t *testing.T) {
	// Test with a more realistic interface
	source := `package test

import (
	"context"
	"io"
)

// Service represents a business service with lifecycle management
type Service interface {
	// Start the service with context
	Start(ctx context.Context) error
	
	// Stop gracefully shuts down the service
	Stop() error
	
	// Status returns current service status
	Status() string
}

// ReadWriteCloser combines multiple standard interfaces
type ReadWriteCloser interface {
	io.Reader
	io.Writer
	io.Closer
	
	// Flush any buffered data
	Flush() error
}`

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", source, parser.ParseComments)
	if err != nil {
		t.Fatalf("Failed to parse source: %v", err)
	}

	analyzer := NewInterfaceAnalyzer(fset)
	interfaces, err := analyzer.AnalyzeInterfaces(file, "test")

	if err != nil {
		t.Fatalf("AnalyzeInterfaces failed: %v", err)
	}

	if len(interfaces) != 2 {
		t.Fatalf("Expected 2 interfaces, got %d", len(interfaces))
	}

	// Find Service interface
	var service metrics.InterfaceMetrics
	for _, intf := range interfaces {
		if intf.Name == "Service" {
			service = intf
			break
		}
	}

	if service.Name == "" {
		t.Fatal("Service interface not found")
	}

	// Test Service interface
	if service.MethodCount != 3 {
		t.Errorf("Expected Service to have 3 methods, got %d", service.MethodCount)
	}

	if !service.IsExported {
		t.Error("Expected Service to be exported")
	}

	if !service.Documentation.HasComment {
		t.Error("Expected Service to have documentation")
	}

	// Find ReadWriteCloser interface
	var rwc metrics.InterfaceMetrics
	for _, intf := range interfaces {
		if intf.Name == "ReadWriteCloser" {
			rwc = intf
			break
		}
	}

	if rwc.Name == "" {
		t.Fatal("ReadWriteCloser interface not found")
	}

	// Test ReadWriteCloser interface
	if len(rwc.EmbeddedInterfaces) != 3 {
		t.Errorf("Expected ReadWriteCloser to have 3 embedded interfaces, got %d", len(rwc.EmbeddedInterfaces))
	}

	if rwc.MethodCount != 1 {
		t.Errorf("Expected ReadWriteCloser to have 1 explicit method, got %d", rwc.MethodCount)
	}

	// Check embedded interfaces
	embeddedNames := make(map[string]bool)
	for _, embedded := range rwc.EmbeddedInterfaces {
		embeddedNames[embedded] = true
	}

	expectedEmbedded := []string{"io.Reader", "io.Writer", "io.Closer"}
	for _, expected := range expectedEmbedded {
		if !embeddedNames[expected] {
			t.Errorf("Expected %s to be embedded in ReadWriteCloser", expected)
		}
	}
}
