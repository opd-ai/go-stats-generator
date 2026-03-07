package smallblocks

// small_blocks.go contains small code blocks below min_block_lines threshold
// These should NOT be flagged as duplicates (negative test case)

// GetUserID extracts the ID field from a User struct, returning an empty string if the user pointer is nil.
// This function demonstrates small code blocks that fall below minimum duplication detection thresholds.
func GetUserID(u *User) string {
	if u == nil {
		return ""
	}
	return u.ID
}

// GetUserName returns the user's name or empty string if user is nil.
func GetUserName(u *User) string {
	if u == nil {
		return ""
	}
	return u.Name
}

// GetUserEmail returns the user's email or empty string if user is nil.
func GetUserEmail(u *User) string {
	if u == nil {
		return ""
	}
	return u.Email
}

// IsValidID checks whether the provided ID string is non-empty, returning false for empty strings.
// This function demonstrates small validation patterns that fall below duplication detection thresholds.
func IsValidID(id string) bool {
	if id == "" {
		return false
	}
	return true
}

// IsValidName checks whether the provided name is non-empty.
func IsValidName(name string) bool {
	if name == "" {
		return false
	}
	return true
}

// IsValidEmail checks whether the provided email is non-empty.
func IsValidEmail(email string) bool {
	if email == "" {
		return false
	}
	return true
}

// CheckErrorA panics if the provided error is non-nil, implementing a simple error-handling pattern.
// This function demonstrates trivial error checks that are too small for duplication flagging.
func CheckErrorA(err error) {
	if err != nil {
		panic(err)
	}
}

// CheckErrorB panics if the provided error is non-nil.
func CheckErrorB(err error) {
	if err != nil {
		panic(err)
	}
}

// CheckErrorC panics if the provided error is non-nil.
func CheckErrorC(err error) {
	if err != nil {
		panic(err)
	}
}

// User represents a user with identifying information.
type User struct {
	ID    string
	Name  string
	Email string
}
