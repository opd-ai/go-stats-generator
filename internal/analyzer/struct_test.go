package analyzer

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"

	"github.com/opd-ai/go-stats-generator/internal/metrics"
)

func TestNewStructAnalyzer(t *testing.T) {
	fset := token.NewFileSet()
	analyzer := NewStructAnalyzer(fset)

	if analyzer == nil {
		t.Fatal("NewStructAnalyzer returned nil")
	}

	if analyzer.fset != fset {
		t.Error("StructAnalyzer fset not set correctly")
	}
}

func TestAnalyzeStructs_EmptyFile(t *testing.T) {
	source := `package test`

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", source, parser.ParseComments)
	if err != nil {
		t.Fatalf("Failed to parse source: %v", err)
	}

	analyzer := NewStructAnalyzer(fset)
	structs, err := analyzer.AnalyzeStructs(file, "test")

	if err != nil {
		t.Fatalf("AnalyzeStructs failed: %v", err)
	}

	if len(structs) != 0 {
		t.Errorf("Expected 0 structs, got %d", len(structs))
	}
}

func TestAnalyzeStructs_SimpleStruct(t *testing.T) {
	source := `package test

// User represents a user in the system
type User struct {
	ID   int    ` + "`json:\"id\"`" + `
	Name string ` + "`json:\"name\"`" + `
}`

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", source, parser.ParseComments)
	if err != nil {
		t.Fatalf("Failed to parse source: %v", err)
	}

	analyzer := NewStructAnalyzer(fset)
	structs, err := analyzer.AnalyzeStructs(file, "test")

	if err != nil {
		t.Fatalf("AnalyzeStructs failed: %v", err)
	}

	if len(structs) != 1 {
		t.Fatalf("Expected 1 struct, got %d", len(structs))
	}

	user := structs[0]

	// Test basic fields
	if user.Name != "User" {
		t.Errorf("Expected name 'User', got '%s'", user.Name)
	}

	if user.Package != "test" {
		t.Errorf("Expected package 'test', got '%s'", user.Package)
	}

	if !user.IsExported {
		t.Error("Expected User to be exported")
	}

	if user.TotalFields != 2 {
		t.Errorf("Expected 2 fields, got %d", user.TotalFields)
	}

	// Test field categorization
	if user.FieldsByType[metrics.FieldTypePrimitive] != 2 {
		t.Errorf("Expected 2 primitive fields, got %d", user.FieldsByType[metrics.FieldTypePrimitive])
	}

	// Test tag analysis
	if user.Tags["json"] != 2 {
		t.Errorf("Expected 2 json tags, got %d", user.Tags["json"])
	}

	// Test documentation
	if !user.Documentation.HasComment {
		t.Error("Expected struct to have documentation")
	}

	if user.Documentation.QualityScore <= 0 {
		t.Error("Expected positive documentation quality score")
	}
}

func TestAnalyzeStructs_ComplexStruct(t *testing.T) {
	source := `package test

import (
	"context"
	"time"
)

// ComplexStruct demonstrates various field types
type ComplexStruct struct {
	// Primitives
	ID       int
	Name     string
	Active   bool
	Price    float64
	
	// Slices and arrays
	Tags     []string
	Numbers  [10]int
	
	// Maps
	Metadata map[string]interface{}
	Counts   map[int]string
	
	// Channels
	Events   chan string
	Results  <-chan int
	Commands chan<- bool
	
	// Interfaces
	Handler  interface{}
	Writer   interface{ Write([]byte) (int, error) }
	
	// Pointers
	Parent   *ComplexStruct
	Config   *Config
	
	// Functions
	Callback func() error
	Transform func(string) string
	
	// External types
	CreatedAt time.Time
	Context   context.Context
}

type Config struct {
	Debug bool
}`

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", source, parser.ParseComments)
	if err != nil {
		t.Fatalf("Failed to parse source: %v", err)
	}

	analyzer := NewStructAnalyzer(fset)
	structs, err := analyzer.AnalyzeStructs(file, "test")

	if err != nil {
		t.Fatalf("AnalyzeStructs failed: %v", err)
	}

	if len(structs) != 2 {
		t.Fatalf("Expected 2 structs, got %d", len(structs))
	}

	var complexStruct metrics.StructMetrics
	for _, s := range structs {
		if s.Name == "ComplexStruct" {
			complexStruct = s
			break
		}
	}

	if complexStruct.Name == "" {
		t.Fatal("ComplexStruct not found")
	}

	// Test field counts by type
	expected := map[metrics.FieldType]int{
		metrics.FieldTypePrimitive: 4, // ID, Name, Active, Price
		metrics.FieldTypeSlice:     2, // Tags, Numbers
		metrics.FieldTypeMap:       2, // Metadata, Counts
		metrics.FieldTypeChannel:   3, // Events, Results, Commands
		metrics.FieldTypeInterface: 2, // Handler, Writer
		metrics.FieldTypePointer:   2, // Parent, Config
		metrics.FieldTypeFunction:  2, // Callback, Transform
		metrics.FieldTypeStruct:    2, // CreatedAt, Context (external types)
	}

	for fieldType, expectedCount := range expected {
		if complexStruct.FieldsByType[fieldType] != expectedCount {
			t.Errorf("Expected %d %s fields, got %d",
				expectedCount, fieldType, complexStruct.FieldsByType[fieldType])
		}
	}

	// Test total field count
	expectedTotal := 19
	if complexStruct.TotalFields != expectedTotal {
		t.Errorf("Expected %d total fields, got %d", expectedTotal, complexStruct.TotalFields)
	}

	// Test complexity
	if complexStruct.Complexity.Overall <= 0 {
		t.Error("Expected positive overall complexity")
	}
}

func TestAnalyzeStructs_EmbeddedTypes(t *testing.T) {
	source := `package test

import "fmt"

type Base struct {
	ID int
}

type External struct {
	Name string
}

type Embedded struct {
	Base                    // Embedded same-package type
	*External              // Embedded pointer to same-package type
	fmt.Stringer           // Embedded external interface
	Value        string    // Regular field
}`

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", source, parser.ParseComments)
	if err != nil {
		t.Fatalf("Failed to parse source: %v", err)
	}

	analyzer := NewStructAnalyzer(fset)
	structs, err := analyzer.AnalyzeStructs(file, "test")

	if err != nil {
		t.Fatalf("AnalyzeStructs failed: %v", err)
	}

	var embedded metrics.StructMetrics
	for _, s := range structs {
		if s.Name == "Embedded" {
			embedded = s
			break
		}
	}

	if embedded.Name == "" {
		t.Fatal("Embedded struct not found")
	}

	// Test embedded types
	if len(embedded.EmbeddedTypes) != 3 {
		t.Errorf("Expected 3 embedded types, got %d", len(embedded.EmbeddedTypes))
	}

	// Check specific embedded types
	embeddedNames := make(map[string]bool)
	for _, emb := range embedded.EmbeddedTypes {
		embeddedNames[emb.Name] = true

		if emb.Name == "External" && !emb.IsPointer {
			t.Error("Expected External to be marked as pointer")
		}

		if emb.Name == "Stringer" && emb.Package != "fmt" {
			t.Errorf("Expected Stringer package to be 'fmt', got '%s'", emb.Package)
		}
	}

	expectedNames := []string{"Base", "External", "Stringer"}
	for _, name := range expectedNames {
		if !embeddedNames[name] {
			t.Errorf("Expected embedded type '%s' not found", name)
		}
	}

	// Test field type categorization with embedded
	if embedded.FieldsByType[metrics.FieldTypeEmbedded] != 3 {
		t.Errorf("Expected 3 embedded fields, got %d", embedded.FieldsByType[metrics.FieldTypeEmbedded])
	}

	if embedded.FieldsByType[metrics.FieldTypePrimitive] != 1 {
		t.Errorf("Expected 1 primitive field, got %d", embedded.FieldsByType[metrics.FieldTypePrimitive])
	}
}

func TestCategorizeFieldType(t *testing.T) {
	fset := token.NewFileSet()
	analyzer := NewStructAnalyzer(fset)

	tests := []struct {
		source       string
		expectedType metrics.FieldType
		description  string
	}{
		{"int", metrics.FieldTypePrimitive, "primitive int"},
		{"string", metrics.FieldTypePrimitive, "primitive string"},
		{"bool", metrics.FieldTypePrimitive, "primitive bool"},
		{"[]string", metrics.FieldTypeSlice, "slice type"},
		{"[10]int", metrics.FieldTypeSlice, "array type"},
		{"map[string]int", metrics.FieldTypeMap, "map type"},
		{"chan string", metrics.FieldTypeChannel, "channel type"},
		{"<-chan int", metrics.FieldTypeChannel, "receive channel"},
		{"chan<- bool", metrics.FieldTypeChannel, "send channel"},
		{"interface{}", metrics.FieldTypeInterface, "empty interface"},
		{"*int", metrics.FieldTypePointer, "pointer type"},
		{"func() error", metrics.FieldTypeFunction, "function type"},
		{"CustomType", metrics.FieldTypeStruct, "custom type"},
		{"pkg.Type", metrics.FieldTypeStruct, "external type"},
	}

	for _, test := range tests {
		source := `package test
type TestStruct struct {
	Field ` + test.source + `
}`

		file, err := parser.ParseFile(fset, "test.go", source, 0)
		if err != nil {
			t.Errorf("Failed to parse %s: %v", test.description, err)
			continue
		}

		// Extract the field type
		genDecl := file.Decls[0].(*ast.GenDecl)
		typeSpec := genDecl.Specs[0].(*ast.TypeSpec)
		structType := typeSpec.Type.(*ast.StructType)
		field := structType.Fields.List[0]

		fieldType := analyzer.categorizeFieldType(field.Type)
		if fieldType != test.expectedType {
			t.Errorf("For %s: expected %s, got %s",
				test.description, test.expectedType, fieldType)
		}
	}
}

func TestIsPrimitiveType(t *testing.T) {
	fset := token.NewFileSet()
	analyzer := NewStructAnalyzer(fset)

	primitives := []string{
		"bool", "string", "int", "int8", "int16", "int32", "int64",
		"uint", "uint8", "uint16", "uint32", "uint64", "uintptr",
		"byte", "rune", "float32", "float64", "complex64", "complex128",
	}

	for _, primitive := range primitives {
		if !analyzer.isPrimitiveType(primitive) {
			t.Errorf("Expected %s to be primitive", primitive)
		}
	}

	nonPrimitives := []string{"CustomType", "time.Time", "error", "interface{}"}
	for _, nonPrimitive := range nonPrimitives {
		if analyzer.isPrimitiveType(nonPrimitive) {
			t.Errorf("Expected %s to not be primitive", nonPrimitive)
		}
	}
}

func TestAnalyzeTags(t *testing.T) {
	fset := token.NewFileSet()
	analyzer := NewStructAnalyzer(fset)

	tests := []struct {
		tagValue    string
		expectedTag string
		description string
	}{
		{"`json:\"name\"`", "json", "json tag"},
		{"`xml:\"data\"`", "xml", "xml tag"},
		{"`json:\"id\" xml:\"id\"`", "json", "multiple tags"},
		{"`validate:\"required\"`", "validate", "validate tag"},
		{"`db:\"user_id\"`", "db", "db tag"},
		{"`form:\"username\"`", "form", "form tag"},
		{"`yaml:\"config\"`", "yaml", "yaml tag"},
		{"`binding:\"required\"`", "binding", "binding tag"},
	}

	for _, test := range tests {
		structMetric := &metrics.StructMetrics{
			Tags: make(map[string]int),
		}

		analyzer.analyzeTags(test.tagValue, structMetric)

		if structMetric.Tags[test.expectedTag] != 1 {
			t.Errorf("For %s: expected %s tag count 1, got %d",
				test.description, test.expectedTag, structMetric.Tags[test.expectedTag])
		}
	}
}

func TestCalculateComplexity(t *testing.T) {
	fset := token.NewFileSet()
	analyzer := NewStructAnalyzer(fset)

	structMetric := metrics.StructMetrics{
		TotalFields: 5,
		FieldsByType: map[metrics.FieldType]int{
			metrics.FieldTypePrimitive: 2,
			metrics.FieldTypeMap:       1,
			metrics.FieldTypeFunction:  1,
			metrics.FieldTypeEmbedded:  1,
		},
		EmbeddedTypes: []metrics.EmbeddedType{
			{Name: "Base"},
		},
	}

	complexity := analyzer.calculateComplexity(structMetric)

	// Test basic complexity calculation
	if complexity.Cyclomatic <= 0 {
		t.Error("Expected positive cyclomatic complexity")
	}

	if complexity.NestingDepth != 1 {
		t.Errorf("Expected nesting depth 1, got %d", complexity.NestingDepth)
	}

	if complexity.Overall <= 0 {
		t.Error("Expected positive overall complexity")
	}

	// Test that complex field types increase complexity more
	expectedBase := 5 + 2 + 1*2 + 1*3 + 1*3 // totalFields + primitives + map*2 + function*3 + embedded*3
	if complexity.Cyclomatic != expectedBase {
		t.Errorf("Expected cyclomatic complexity %d, got %d", expectedBase, complexity.Cyclomatic)
	}
}

func TestAnalyzeStructs_Integration(t *testing.T) {
	// Test with real testdata
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "../../testdata/simple/user.go", nil, parser.ParseComments)
	if err != nil {
		t.Fatalf("Failed to parse testdata: %v", err)
	}

	analyzer := NewStructAnalyzer(fset)
	structs, err := analyzer.AnalyzeStructs(file, "simple")

	if err != nil {
		t.Fatalf("AnalyzeStructs failed: %v", err)
	}

	// Should find User and UserService structs
	if len(structs) < 2 {
		t.Errorf("Expected at least 2 structs, got %d", len(structs))
	}

	// Find User struct
	var userStruct metrics.StructMetrics
	for _, s := range structs {
		if s.Name == "User" {
			userStruct = s
			break
		}
	}

	if userStruct.Name == "" {
		t.Fatal("User struct not found in testdata")
	}

	// Verify User struct analysis
	if userStruct.TotalFields != 5 {
		t.Errorf("Expected User to have 5 fields, got %d", userStruct.TotalFields)
	}

	if !userStruct.IsExported {
		t.Error("Expected User to be exported")
	}

	if userStruct.Documentation.HasComment {
		if userStruct.Documentation.QualityScore <= 0 {
			t.Error("Expected positive documentation quality score for User")
		}
	}
}

func TestAnalyzeStructs_WithMethods(t *testing.T) {
	source := `package test

// User represents a user in the system with associated methods
type User struct {
	ID   int    ` + "`json:\"id\"`" + `
	Name string ` + "`json:\"name\"`" + `
}

// GetID returns the user ID (value receiver)
func (u User) GetID() int {
	return u.ID
}

// SetName sets the user name (pointer receiver)
func (u *User) SetName(name string) {
	u.Name = name
}

// ValidateComplex performs complex validation
func (u *User) ValidateComplex(options map[string]interface{}) (bool, error) {
	if u.ID <= 0 {
		return false, errors.New("invalid ID")
	}
	
	if len(u.Name) == 0 {
		return false, errors.New("name required")
	}
	
	for key, value := range options {
		switch key {
		case "strict":
			if strict, ok := value.(bool); ok && strict {
				if len(u.Name) < 3 {
					return false, errors.New("name too short")
				}
			}
		case "format":
			// Additional format validation
			if format, ok := value.(string); ok {
				if format == "email" && len(u.Name) > 0 {
					// Simple email check without strings package
					hasAt := false
					for _, char := range u.Name {
						if char == '@' {
							hasAt = true
							break
						}
					}
					if !hasAt {
						return false, errors.New("invalid email format")
					}
				}
			}
		}
	}
	
	return true, nil
}

// unexportedMethod is not exported
func (u *User) unexportedMethod() {
	// Do nothing
}`

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", source, parser.ParseComments)
	if err != nil {
		t.Fatalf("Failed to parse source: %v", err)
	}

	analyzer := NewStructAnalyzer(fset)
	structs, err := analyzer.AnalyzeStructs(file, "test")

	if err != nil {
		t.Fatalf("AnalyzeStructs failed: %v", err)
	}

	if len(structs) != 1 {
		t.Fatalf("Expected 1 struct, got %d", len(structs))
	}

	user := structs[0]

	// Test that methods were found
	if len(user.Methods) != 4 {
		t.Errorf("Expected 4 methods, got %d", len(user.Methods))
		for i, method := range user.Methods {
			t.Logf("Method %d: %s", i, method.Name)
		}
	}

	// Check specific methods
	methodNames := make(map[string]metrics.MethodInfo)
	for _, method := range user.Methods {
		methodNames[method.Name] = method
	}

	// Test GetID method (value receiver)
	if getID, exists := methodNames["GetID"]; exists {
		if !getID.IsExported {
			t.Error("GetID should be exported")
		}
		if getID.IsPointer {
			t.Error("GetID should not have pointer receiver")
		}
		if getID.Signature.ReturnCount != 1 {
			t.Errorf("GetID should return 1 value, got %d", getID.Signature.ReturnCount)
		}
	} else {
		t.Error("GetID method not found")
	}

	// Test SetName method (pointer receiver)
	if setName, exists := methodNames["SetName"]; exists {
		if !setName.IsExported {
			t.Error("SetName should be exported")
		}
		if !setName.IsPointer {
			t.Error("SetName should have pointer receiver")
		}
		if setName.Signature.ParameterCount != 1 {
			t.Errorf("SetName should have 1 parameter, got %d", setName.Signature.ParameterCount)
		}
	} else {
		t.Error("SetName method not found")
	}

	// Test ValidateComplex method (complex signature)
	if validate, exists := methodNames["ValidateComplex"]; exists {
		if !validate.IsExported {
			t.Error("ValidateComplex should be exported")
		}
		if !validate.IsPointer {
			t.Error("ValidateComplex should have pointer receiver")
		}
		if validate.Signature.ParameterCount != 1 {
			t.Errorf("ValidateComplex should have 1 parameter, got %d", validate.Signature.ParameterCount)
		}
		if validate.Signature.ReturnCount != 2 {
			t.Errorf("ValidateComplex should return 2 values, got %d", validate.Signature.ReturnCount)
		}
		if !validate.Signature.ErrorReturn {
			t.Error("ValidateComplex should return error")
		}
		if validate.Complexity.Cyclomatic < 5 {
			t.Errorf("ValidateComplex should have high complexity, got %d", validate.Complexity.Cyclomatic)
		}
	} else {
		t.Error("ValidateComplex method not found")
	}

	// Test unexported method
	if unexported, exists := methodNames["unexportedMethod"]; exists {
		if unexported.IsExported {
			t.Error("unexportedMethod should not be exported")
		}
		if !unexported.IsPointer {
			t.Error("unexportedMethod should have pointer receiver")
		}
	} else {
		t.Error("unexportedMethod not found")
	}

	// Test overall struct complexity includes methods
	if user.Complexity.Overall <= 2.0 {
		t.Errorf("Expected higher complexity with methods, got %f", user.Complexity.Overall)
	}
}
