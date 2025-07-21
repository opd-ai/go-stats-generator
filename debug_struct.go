package main

import (
	"fmt"
	"go/parser"
	"go/token"

	"github.com/opd-ai/go-stats-generator/internal/analyzer"
)

func main() {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "testdata/simple/user.go", nil, parser.ParseComments)
	if err != nil {
		fmt.Printf("Parse error: %v\n", err)
		return
	}

	structAnalyzer := analyzer.NewStructAnalyzer(fset)
	structs, err := structAnalyzer.AnalyzeStructs(file, "simple")
	if err != nil {
		fmt.Printf("Analysis error: %v\n", err)
		return
	}

	fmt.Printf("Found %d structs:\n", len(structs))
	for _, s := range structs {
		fmt.Printf("- %s: %d fields\n", s.Name, s.TotalFields)
	}
}
