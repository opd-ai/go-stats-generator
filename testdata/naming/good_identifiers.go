// Package naming demonstrates proper Go naming conventions
package naming

import "fmt"

// URL uses correct acronym casing
type URL struct {
	Path string
}

// GetUser follows MixedCaps convention
func GetUser(userID int) string {
	return fmt.Sprintf("user-%d", userID)
}

// UserService has no stuttering
type UserService struct {
	db string
}

// NewService creates a new UserService
func NewService() *UserService {
	return &UserService{db: "postgres"}
}

// HTTPClient uses correct acronym casing
type HTTPClient struct {
	URL string
}

// NewHTTPClient creates a new HTTP client
func NewHTTPClient() *HTTPClient {
	return &HTTPClient{URL: "http://example.com"}
}

// Single letter variables are acceptable in short loops
func processItems(items []int) int {
	sum := 0
	for i, v := range items {
		if i%2 == 0 {
			sum += v
		}
	}
	return sum
}

// UserID uses correct acronym casing
type UserID int

// JSONData uses correct acronym casing
type JSONData struct {
	XMLContent string
}
