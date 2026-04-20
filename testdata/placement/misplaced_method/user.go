package placement

// User represents a user in the system identified by a unique ID, username, and email.
// This is an intentionally minimal struct definition used to test placement analysis
// when methods for the receiver type are defined in separate files (validator.go).
type User struct {
	ID       int
	Username string
	Email    string
}

// NewUser creates a new user with the specified ID, username, and email address.
// Returns a pointer to the newly created User with all fields initialized.
func NewUser(id int, username, email string) *User {
	return &User{
		ID:       id,
		Username: username,
		Email:    email,
	}
}
