package placement

import "fmt"

// This file has high cohesion - all declarations are related to User

const (
	// StatusActive represents an active user account
	StatusActive = "active"
	// StatusInactive represents an inactive user account
	StatusInactive = "inactive"
	// UserDisplayFormat is the template for user string representation
	UserDisplayFormat = "User %s (%s): %s"
	// MinEmailLength is the minimum valid email length
	MinEmailLength = 3
)

// User represents a user in the system with authentication and profile information.
// This type demonstrates high cohesion where all related methods are co-located
// with the struct definition in the same file.
type User struct {
	ID       int    // Unique identifier for the user
	Username string // Display name for the user
	Email    string // Email address for the user
	Active   bool   // Whether the user account is currently active
}

// NewUser creates a new user with the given ID, username, and email.
// The user is created in an active state by default.
// Returns a pointer to the newly created User.
func NewUser(id int, username, email string) *User {
	return &User{
		ID:       id,
		Username: username,
		Email:    email,
		Active:   true,
	}
}

// Activate sets the user's Active status to true, enabling their account.
// This method is used to re-enable previously deactivated user accounts.
func (u *User) Activate() {
	u.Active = true
}

// Deactivate sets the user's Active status to false, disabling their account.
// This method is used to temporarily suspend user access without deletion.
func (u *User) Deactivate() {
	u.Active = false
}

// Display returns a human-readable string representation of the user.
// The format includes username, email, and current active status.
// Returns a formatted string in the form "User <username> (<email>): <status>".
func (u *User) Display() string {
	status := StatusInactive
	if u.Active {
		status = StatusActive
	}
	return fmt.Sprintf(UserDisplayFormat, u.Username, u.Email, status)
}

// ValidateEmail checks if the user has a valid email address.
// A valid email is defined as non-empty and having at least 3 characters.
// Returns true if the email meets validation criteria, false otherwise.
func (u *User) ValidateEmail() bool {
	return u.Email != "" && len(u.Email) > MinEmailLength
}

// IsActive returns whether the user account is currently active.
// Active users have full access to the system.
// Returns the current value of the Active field.
func (u *User) IsActive() bool {
	return u.Active
}
