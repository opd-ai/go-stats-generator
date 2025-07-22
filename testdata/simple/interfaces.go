package simple

import (
	"context"
	"io"
)

// UserRepository defines the contract for user data access
type UserRepository interface {
	// GetUser retrieves a user by ID
	GetUser(ctx context.Context, id int) (*User, error)

	// SaveUser persists a user
	SaveUser(ctx context.Context, user *User) error

	// DeleteUser removes a user
	DeleteUser(ctx context.Context, id int) error

	// ListUsers returns all users with pagination
	ListUsers(ctx context.Context, limit, offset int) ([]*User, error)
}

// Logger represents a logging interface
type Logger interface {
	Info(msg string, args ...interface{})
	Error(msg string, args ...interface{})
	Debug(msg string, args ...interface{})
}

// ReadWriteCloser combines standard I/O interfaces
type ReadWriteCloser interface {
	io.Reader
	io.Writer
	io.Closer
	Flush() error
}

// Service represents a generic service interface
type Service interface {
	Start() error
	Stop() error
	IsRunning() bool
}

// Empty interface for any type
type Any interface{}
