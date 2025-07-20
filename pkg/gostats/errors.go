package gostats

import "errors"

var (
	// ErrNoGoFiles is returned when no Go files are found in the target directory
	ErrNoGoFiles = errors.New("no Go files found")

	// ErrInvalidDirectory is returned when the target directory doesn't exist
	ErrInvalidDirectory = errors.New("invalid directory")

	// ErrParsingFailed is returned when file parsing fails
	ErrParsingFailed = errors.New("failed to parse Go file")

	// ErrAnalysisFailed is returned when analysis fails
	ErrAnalysisFailed = errors.New("analysis failed")
)
