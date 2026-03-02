package analyzer

import (
	"go/parser"
	"go/token"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPackageNameAnalysis_Integration(t *testing.T) {
	na := NewNamingAnalyzer()

	testCases := []struct {
		name               string
		sourceCode         string
		pkgName            string
		dirName            string
		expectedViolations int
		expectedTypes      []string
	}{
		{
			name: "directory mismatch - package service in user dir",
			sourceCode: `package service

type User struct {
	ID string
}`,
			pkgName:            "service",
			dirName:            "user",
			expectedViolations: 1,
			expectedTypes:      []string{"directory_mismatch"},
		},
		{
			name: "generic package name - util",
			sourceCode: `package util

func Helper() {}`,
			pkgName:            "util",
			dirName:            "util",
			expectedViolations: 1,
			expectedTypes:      []string{"generic_package_name"},
		},
		{
			name: "stdlib collision - http",
			sourceCode: `package http

type Client struct {}`,
			pkgName:            "http",
			dirName:            "http",
			expectedViolations: 1,
			expectedTypes:      []string{"stdlib_collision"},
		},
		{
			name: "package with underscore",
			sourceCode: `package user_service

type User struct {}`,
			pkgName:            "user_service",
			dirName:            "user_service",
			expectedViolations: 1,
			expectedTypes:      []string{"non_conventional_name"},
		},
		{
			name: "multiple violations - generic and mismatch",
			sourceCode: `package util

func Helper() {}`,
			pkgName:            "util",
			dirName:            "utilities",
			expectedViolations: 2,
			expectedTypes:      []string{"generic_package_name", "directory_mismatch"},
		},
		{
			name: "valid package - no violations",
			sourceCode: `package authentication

type User struct {}`,
			pkgName:            "authentication",
			dirName:            "authentication",
			expectedViolations: 0,
			expectedTypes:      []string{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, "test.go", tc.sourceCode, parser.AllErrors)
			require.NoError(t, err)
			require.NotNil(t, file)

			violations := na.AnalyzePackageName(tc.pkgName, tc.dirName, "/path/to/"+tc.dirName+"/test.go")

			assert.Len(t, violations, tc.expectedViolations, "expected %d violations", tc.expectedViolations)

			if len(tc.expectedTypes) > 0 {
				violationTypes := make(map[string]bool)
				for _, v := range violations {
					violationTypes[v.ViolationType] = true
				}

				for _, expectedType := range tc.expectedTypes {
					assert.True(t, violationTypes[expectedType], "expected violation type %s", expectedType)
				}
			}
		})
	}
}
