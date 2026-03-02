package smallblocks

// small_blocks.go contains small code blocks below min_block_lines threshold
// These should NOT be flagged as duplicates (negative test case)

// Simple getters - too small to be considered duplicates (< 6 lines)
func GetUserID(u *User) string {
	if u == nil {
		return ""
	}
	return u.ID
}

func GetUserName(u *User) string {
	if u == nil {
		return ""
	}
	return u.Name
}

func GetUserEmail(u *User) string {
	if u == nil {
		return ""
	}
	return u.Email
}

// Simple validation - identical structure but too small
func IsValidID(id string) bool {
	if id == "" {
		return false
	}
	return true
}

func IsValidName(name string) bool {
	if name == "" {
		return false
	}
	return true
}

func IsValidEmail(email string) bool {
	if email == "" {
		return false
	}
	return true
}

// Trivial error checks - common pattern but too small to flag
func CheckErrorA(err error) {
	if err != nil {
		panic(err)
	}
}

func CheckErrorB(err error) {
	if err != nil {
		panic(err)
	}
}

func CheckErrorC(err error) {
	if err != nil {
		panic(err)
	}
}

type User struct {
	ID    string
	Name  string
	Email string
}
