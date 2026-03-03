package analyzer

import (
	"go/parser"
	"go/token"
	"testing"
)

// TestNestedInterfaceEmbeddingDepth tests that deeply nested interface embeddings
// calculate the correct depth using graph traversal
func TestNestedInterfaceEmbeddingDepth(t *testing.T) {
	source := `package test

// Base interface with no embedding
type Base interface {
BaseMethod()
}

// Level1 embeds Base (depth should be 1)
type Level1 interface {
Base
Level1Method()
}

// Level2 embeds Level1 (depth should be 2)
type Level2 interface {
Level1
Level2Method()
}

// Level3 embeds Level2 (depth should be 3)
type Level3 interface {
Level2
Level3Method()
}
`

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", source, parser.ParseComments)
	if err != nil {
		t.Fatalf("Failed to parse source: %v", err)
	}

	analyzer := NewInterfaceAnalyzer(fset)

	// First pass to build interface definitions
	interfaces, err := analyzer.AnalyzeInterfaces(file, "test")
	if err != nil {
		t.Fatalf("AnalyzeInterfaces failed: %v", err)
	}

	if len(interfaces) != 4 {
		t.Fatalf("Expected 4 interfaces, got %d", len(interfaces))
	}

	// Create a map for easier lookup
	interfaceMap := make(map[string]int)
	for _, iface := range interfaces {
		interfaceMap[iface.Name] = iface.EmbeddingDepth
	}

	// Verify depths
	tests := []struct {
		name          string
		expectedDepth int
	}{
		{"Base", 0},
		{"Level1", 1},
		{"Level2", 2},
		{"Level3", 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actualDepth, exists := interfaceMap[tt.name]
			if !exists {
				t.Fatalf("Interface %s not found in analysis results", tt.name)
			}
			if actualDepth != tt.expectedDepth {
				t.Errorf("Interface %s: expected embedding depth %d, got %d",
					tt.name, tt.expectedDepth, actualDepth)
			}
		})
	}
}
