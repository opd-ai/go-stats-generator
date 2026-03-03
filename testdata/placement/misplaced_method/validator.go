package placement

import "strings"

// Validate is a method on User but defined in a different file
// This should be flagged as misplaced - methods should be with their receiver type
func (u *User) Validate() bool {
	if u.Username == "" || u.Email == "" {
		return false
	}
	return strings.Contains(u.Email, "@")
}

// IsAdmin is also misplaced - should be in user.go with User type
func (u *User) IsAdmin() bool {
	return strings.HasPrefix(u.Username, "admin_")
}
