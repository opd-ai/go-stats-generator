package analyzer

import (
	"go/parser"
	"go/token"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheckUnusedReceiverName(t *testing.T) {
	tests := []struct {
		name          string
		code          string
		expectPattern bool
		description   string
	}{
		{
			name: "receiver is used",
			code: `package main
type Client struct { name string }
func (c *Client) GetName() string {
	return c.name
}`,
			expectPattern: false,
			description:   "Receiver c is referenced in method body",
		},
		{
			name: "receiver is not used",
			code: `package main
type Client struct { name string }
func (c *Client) DoNothing() {
	println("nothing")
}`,
			expectPattern: true,
			description:   "Receiver c is never referenced in method body",
		},
		{
			name: "receiver is underscore (idiomatic)",
			code: `package main
type Client struct { name string }
func (_ *Client) StaticMethod() {
	println("static")
}`,
			expectPattern: false,
			description:   "Underscore receiver is idiomatic for unused receivers",
		},
		{
			name: "not a method (no receiver)",
			code: `package main
func DoSomething() {
	println("something")
}`,
			expectPattern: false,
			description:   "Plain function without receiver should not trigger",
		},
		{
			name: "receiver used in nested function",
			code: `package main
type Client struct { id int }
func (c *Client) Process() {
	helper := func() int {
		return c.id
	}
	_ = helper()
}`,
			expectPattern: false,
			description:   "Receiver used in closure should count as used",
		},
		{
			name: "receiver used in field assignment",
			code: `package main
type Client struct { value int }
func (c *Client) SetValue(v int) {
	c.value = v
}`,
			expectPattern: false,
			description:   "Receiver used in field assignment should count as used",
		},
		{
			name: "receiver used in method call",
			code: `package main
type Client struct { value int }
func (c *Client) helper() int { return c.value }
func (c *Client) DoWork() {
	_ = c.helper()
}`,
			expectPattern: false,
			description:   "Receiver used in method call should count as used",
		},
		{
			name: "value receiver not used",
			code: `package main
type Config struct { debug bool }
func (cfg Config) IsProduction() bool {
	return false
}`,
			expectPattern: true,
			description:   "Unused value receiver should be flagged",
		},
		{
			name: "receiver shadowed but original used",
			code: `package main
type Client struct { name string }
func (c *Client) GetName() string {
	name := c.name
	return name
}`,
			expectPattern: false,
			description:   "Receiver used even if value is shadowed",
		},
		{
			name: "receiver used in return statement",
			code: `package main
type Client struct {}
func (c *Client) Self() *Client {
	return c
}`,
			expectPattern: false,
			description:   "Receiver returned directly should count as used",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, "test.go", tt.code, 0)
			assert.NoError(t, err, "Failed to parse test code")

			analyzer := NewAntipatternAnalyzer(fset)
			patterns := analyzer.Analyze(file)

			// Filter for unused_receiver patterns
			var unusedReceiverPatterns int
			for _, p := range patterns {
				if p.Type == "unused_receiver" {
					unusedReceiverPatterns++
				}
			}

			if tt.expectPattern {
				assert.Equal(t, 1, unusedReceiverPatterns, "Expected one unused_receiver pattern: %s", tt.description)
				if unusedReceiverPatterns > 0 {
					// Verify the pattern has the expected fields
					for _, p := range patterns {
						if p.Type == "unused_receiver" {
							assert.Equal(t, "low", p.Severity)
							assert.Contains(t, p.Description, "receiver")
							assert.Contains(t, p.Suggestion, "_")
						}
					}
				}
			} else {
				assert.Equal(t, 0, unusedReceiverPatterns, "Expected no unused_receiver pattern: %s", tt.description)
			}
		})
	}
}

func TestCheckUnusedReceiverName_MultipleReceivers(t *testing.T) {
	code := `package main
type Client struct { name string }
func (c *Client) UsedReceiver() string {
	return c.name
}
func (c *Client) UnusedReceiver() {
	println("hello")
}
func (_ *Client) UnderscoreReceiver() {
	println("static")
}`

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", code, 0)
	assert.NoError(t, err)

	analyzer := NewAntipatternAnalyzer(fset)
	patterns := analyzer.Analyze(file)

	// Count unused_receiver patterns
	var unusedReceiverPatterns int
	for _, p := range patterns {
		if p.Type == "unused_receiver" {
			unusedReceiverPatterns++
		}
	}

	// Should only flag UnusedReceiver method (not UsedReceiver or UnderscoreReceiver)
	assert.Equal(t, 1, unusedReceiverPatterns, "Should flag exactly one unused receiver")
}

func TestCheckUnusedReceiverName_InterfaceImplementation(t *testing.T) {
	// Common pattern: implementing an interface method that doesn't need the receiver
	code := `package main
type Stringer interface {
	String() string
}
type Empty struct{}
func (e *Empty) String() string {
	return "empty"
}`

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", code, 0)
	assert.NoError(t, err)

	analyzer := NewAntipatternAnalyzer(fset)
	patterns := analyzer.Analyze(file)

	// Count unused_receiver patterns
	var unusedReceiverPatterns int
	for _, p := range patterns {
		if p.Type == "unused_receiver" {
			unusedReceiverPatterns++
		}
	}

	// Should flag the receiver even though it implements an interface
	// (developer can choose to use _ if they want to suppress the warning)
	assert.Equal(t, 1, unusedReceiverPatterns, "Should flag interface method with unused receiver")
}
