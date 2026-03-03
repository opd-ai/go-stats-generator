package placement

import "fmt"

// This file has high cohesion - all declarations are related to User

// User represents a user in the system
type User struct {
	ID       int
	Username string
	Email    string
	Active   bool
}

// NewUser creates a new user
func NewUser(id int, username, email string) *User {
	return &User{
		ID:       id,
		Username: username,
		Email:    email,
		Active:   true,
	}
}

// Activate activates the user
func (u *User) Activate() {
	u.Active = true
}

// Deactivate deactivates the user
func (u *User) Deactivate() {
	u.Active = false
}

// Display returns a string representation
func (u *User) Display() string {
	status := "inactive"
	if u.Active {
		status = "active"
	}
	return fmt.Sprintf("User %s (%s): %s", u.Username, u.Email, status)
}

// ValidateEmail checks if the user has a valid email
func (u *User) ValidateEmail() bool {
	return u.Email != "" && len(u.Email) > 3
}

// IsActive returns whether the user is active
func (u *User) IsActive() bool {
	return u.Active
}
