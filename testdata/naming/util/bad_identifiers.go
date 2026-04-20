// Package util has a generic name that violates Go conventions.
// This package intentionally contains naming violations for go-stats-generator testing:
// generic package name, underscore identifiers, and incorrect acronym casing.
package util

import "fmt"

// Url should be URL (acronym casing violation)
type Url struct {
	Path string
}

// get_user violates MixedCaps convention (uses underscores)
func get_user(userId int) string {
	return fmt.Sprintf("user-%d", userId)
}

// User_Service also violates MixedCaps
type User_Service struct {
	db string
}

// NewUserService stutters with package name (util.NewUserService -> util.UserService)
func (s *User_Service) NewUserService() *User_Service {
	return &User_Service{db: "postgres"}
}

// HttpClient should be HTTPClient
type HttpClient struct {
	Url string // Url should be URL
}

// GetHttpClient should be GetHTTPClient
func GetHttpClient() *HttpClient {
	return &HttpClient{Url: "http://example.com"}
}

// Single letter variable in wrong context
func process() {
	x := 42
	y := 100
	z := x + y
	fmt.Println(z)
}

// UserId represents a user identifier as an integer. The type name violates Go naming conventions
// (should be UserID with capitalized initialism) and is used for naming analysis testing.
type UserId int

// JsonData should be JSONData
type JsonData struct {
	XmlContent string // XmlContent should be XMLContent
}
