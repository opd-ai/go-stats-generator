package placement

// User represents a user
type User struct {
	ID       int
	Username string
	Email    string
}

// NewUser creates a new user
func NewUser(id int, username, email string) *User {
	return &User{
		ID:       id,
		Username: username,
		Email:    email,
	}
}
