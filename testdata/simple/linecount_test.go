package test

import "fmt"

// ComplexLineCountingTest demonstrates various line counting scenarios
func ComplexLineCountingTest() error {
	// This is a single line comment
	var x int = 42 // Inline comment

	/*
	 * Multi-line block comment
	 * with multiple lines
	 */

	if x > 0 { // Another inline comment
		fmt.Printf("x is positive: %d\n", x) /* inline block */

		/* single line block comment */
		return nil
	}

	// Final comment
	return fmt.Errorf("x is not positive")
}
