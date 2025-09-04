package go_stats_generator_test

import (
	"context"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"testing"

	go_stats_generator "github.com/opd-ai/go-stats-generator/pkg/go-stats-generator"
	"github.com/stretchr/testify/assert"
)

// TestPublicAPILimitedFunctionality validates that the public API is missing features
func TestPublicAPILimitedFunctionality(t *testing.T) {
	// Create a sample Go file with multiple types of constructs
	sampleCode := `package testpkg

import "sync"

// User represents a user in the system
type User struct {
	ID   int    ` + "`json:\"id\"`" + `
	Name string ` + "`json:\"name\"`" + `
}

// UserService provides user operations
type UserService interface {
	GetUser(id int) (*User, error)
	CreateUser(name string) (*User, error)
}

// ProcessUsers processes users concurrently
func ProcessUsers(users []User) {
	var wg sync.WaitGroup
	ch := make(chan User, 10)
	
	// Worker goroutine
	go func() {
		for user := range ch {
			// Process user
			_ = user
		}
	}()
	
	// Send users to channel
	for _, user := range users {
		wg.Add(1)
		ch <- user
		wg.Done()
	}
	
	close(ch)
	wg.Wait()
}

func SimpleFunction() {
	// Simple function
}`

	// Parse the code to validate it has the expected constructs
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", sampleCode, parser.ParseComments)
	assert.NoError(t, err)

	// Verify the sample contains all types of constructs
	var hasStruct, hasInterface, hasFunction, hasConcurrency bool
	ast.Inspect(file, func(n ast.Node) bool {
		switch n.(type) {
		case *ast.StructType:
			hasStruct = true
		case *ast.InterfaceType:
			hasInterface = true
		case *ast.FuncDecl:
			hasFunction = true
		case *ast.GoStmt:
			hasConcurrency = true
		}
		return true
	})

	assert.True(t, hasStruct, "Sample code should contain structs")
	assert.True(t, hasInterface, "Sample code should contain interfaces")
	assert.True(t, hasFunction, "Sample code should contain functions")
	assert.True(t, hasConcurrency, "Sample code should contain concurrency patterns")

	// Now test the current API limitation
	analyzer := go_stats_generator.NewAnalyzer()

	// Create a temporary file
	tempFile := "/tmp/test_api_limitation.go"
	err = writeStringToFile(tempFile, sampleCode)
	assert.NoError(t, err)
	defer removeFile(tempFile)

	// Analyze the file using the public API
	report, err := analyzer.AnalyzeFile(context.Background(), tempFile)
	assert.NoError(t, err)
	assert.NotNil(t, report)

	// Current limitation: Only functions are analyzed
	assert.Greater(t, len(report.Functions), 0, "Should detect functions")

	// BUG: These should all be populated but are currently missing from the API
	assert.Empty(t, report.Structs, "BUG: Structs analysis is missing from public API")
	assert.Empty(t, report.Interfaces, "BUG: Interface analysis is missing from public API")
	assert.Empty(t, report.Packages, "BUG: Package analysis is missing from public API")
	assert.Empty(t, report.Patterns.ConcurrencyPatterns.Goroutines.Instances, "BUG: Concurrency analysis is missing from public API")
}

// Helper function to write string to file
func writeStringToFile(filePath, content string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.WriteString(content)
	return err
}

// Helper function to remove file
func removeFile(filePath string) {
	os.Remove(filePath)
}
