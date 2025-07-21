package main

import (
	"fmt"
	"path/filepath"
	"strings"
)

func main() {
	path := "src\\vendor\\lib"
	fmt.Printf("Original path: %q\n", path)
	fmt.Printf("ToSlash result: %q\n", filepath.ToSlash(path))

	segments := strings.Split(filepath.ToSlash(path), "/")
	fmt.Printf("Segments: %v\n", segments)

	for _, s := range segments {
		if s == "vendor" {
			fmt.Printf("Found vendor segment: %q\n", s)
		}
	}
}
