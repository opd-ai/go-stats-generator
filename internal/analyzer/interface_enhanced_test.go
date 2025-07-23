package analyzer

import (
	"go/parser"
	"go/token"
	"testing"

	"github.com/opd-ai/go-stats-generator/internal/metrics"
)

func TestInterfaceImplementationTracking(t *testing.T) {
	source := `package test

// Writer interface for writing data
type Writer interface {
	Write(data []byte) (int, error)
	Close() error
}

// File implements Writer interface
type File struct {
	name string
}

func (f *File) Write(data []byte) (int, error) {
	return len(data), nil
}

func (f *File) Close() error {
	return nil
}

// PartialImpl only implements one method
type PartialImpl struct{}

func (p *PartialImpl) Write(data []byte) (int, error) {
	return 0, nil
}
`

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

	// Test implementation tracking
	if writer.ImplementationCount != 1 {
		t.Errorf("Expected 1 implementation, got %d", writer.ImplementationCount)
	}

	if len(writer.Implementations) != 1 {
		t.Errorf("Expected 1 implementation in slice, got %d", len(writer.Implementations))
	}

	if writer.Implementations[0] != "test.File" {
		t.Errorf("Expected implementation 'test.File', got '%s'", writer.Implementations[0])
	}

	// Test implementation ratio
	expectedRatio := float64(1) / float64(2) // 1 implementation / 2 methods
	if writer.ImplementationRatio != expectedRatio {
		t.Errorf("Expected implementation ratio %.2f, got %.2f", expectedRatio, writer.ImplementationRatio)
	}

	// Test complexity score is calculated
	if writer.ComplexityScore == 0 {
		t.Error("Expected complexity score to be calculated")
	}

	// Test embedding depth
	if writer.EmbeddingDepth != 0 {
		t.Errorf("Expected embedding depth 0, got %d", writer.EmbeddingDepth)
	}
}

func TestEmbeddedInterfaceAnalysis(t *testing.T) {
	source := `package test

import "io"

// ReadWriter embeds io.Reader and io.Writer
type ReadWriter interface {
	io.Reader
	io.Writer
	Flush() error
}
`

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

	readWriter := interfaces[0]

	// Test embedded interfaces detection
	if len(readWriter.EmbeddedInterfaces) != 2 {
		t.Errorf("Expected 2 embedded interfaces, got %d", len(readWriter.EmbeddedInterfaces))
	}

	// Test method count (only direct methods, not embedded)
	if readWriter.MethodCount != 1 {
		t.Errorf("Expected 1 direct method, got %d", readWriter.MethodCount)
	}

	// Test embedding depth (external package interfaces)
	if readWriter.EmbeddingDepth != 2 {
		t.Errorf("Expected embedding depth 2 (external), got %d", readWriter.EmbeddingDepth)
	}

	// Test complexity includes embedding
	if readWriter.ComplexityScore <= 1.0 {
		t.Errorf("Expected complexity score > 1.0 due to embedding, got %.2f", readWriter.ComplexityScore)
	}
}

func TestComplexMethodSignatures(t *testing.T) {
	source := `package test

// ComplexInterface has complex method signatures
type ComplexInterface interface {
	SimpleMethod()
	WithParams(a int, b string) error
	VariadicMethod(base string, others ...interface{}) ([]string, error)
	WithInterfaces(reader io.Reader, writer io.Writer) error
}
`

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

	complex := interfaces[0]

	// Test method count
	if complex.MethodCount != 4 {
		t.Errorf("Expected 4 methods, got %d", complex.MethodCount)
	}

	// Test method signatures are analyzed
	for _, method := range complex.Methods {
		if method.Signature.ComplexityScore == 0 {
			t.Errorf("Expected method '%s' to have signature complexity > 0", method.Name)
		}
	}

	// Find variadic method
	var variadicMethod *metrics.InterfaceMethod
	for _, method := range complex.Methods {
		if method.Name == "VariadicMethod" {
			variadicMethod = &method
			break
		}
	}

	if variadicMethod == nil {
		t.Fatal("VariadicMethod not found")
	}

	if !variadicMethod.Signature.VariadicUsage {
		t.Error("Expected VariadicMethod to have variadic usage")
	}

	if !variadicMethod.Signature.ErrorReturn {
		t.Error("Expected VariadicMethod to return error")
	}
}
