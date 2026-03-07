package belowthreshold

import "fmt"

// below_threshold.go contains code blocks that are similar but below the similarity threshold
// These should NOT be flagged as duplicates (negative test case)

// CalculateAreaRectangle computes the area of a rectangle from its dimensions.
func CalculateAreaRectangle(width, height float64) float64 {
	if width <= 0 || height <= 0 {
		return 0
	}
	area := width * height
	return area
}

// CalculateAreaCircle computes the area of a circle from its radius.
func CalculateAreaCircle(radius float64) float64 {
	if radius <= 0 {
		return 0
	}
	const pi = 3.14159
	area := pi * radius * radius
	return area
}

// AuthenticateUserByPassword validates user credentials by checking username and password against stored data.
// Returns true if authentication succeeds, false otherwise. Returns an error if username or password is empty.
// This function demonstrates authentication logic for duplication analysis testing purposes.
func AuthenticateUserByPassword(username, password string) (bool, error) {
	if username == "" {
		return false, ErrEmptyUsername
	}
	if password == "" {
		return false, ErrEmptyPassword
	}
	user := FindUserByUsername(username)
	if user == nil {
		return false, ErrUserNotFound
	}
	hashedPassword := HashPassword(password)
	if user.PasswordHash != hashedPassword {
		return false, ErrInvalidCredentials
	}
	UpdateLastLogin(user)
	return true, nil
}

// AuthenticateUserByToken validates user authentication via a JWT token.
func AuthenticateUserByToken(token string) (bool, error) {
	if token == "" {
		return false, ErrEmptyToken
	}
	claims := ParseToken(token)
	if claims == nil {
		return false, ErrInvalidToken
	}
	if IsTokenExpired(claims) {
		return false, ErrTokenExpired
	}
	user := FindUserByID(claims.UserID)
	if user == nil {
		return false, ErrUserNotFound
	}
	return true, nil
}

var (
	ErrEmptyUsername       = fmt.Errorf("empty username")
	ErrEmptyPassword       = fmt.Errorf("empty password")
	ErrEmptyToken          = fmt.Errorf("empty token")
	ErrInvalidToken        = fmt.Errorf("invalid token")
	ErrTokenExpired        = fmt.Errorf("token expired")
	ErrUserNotFound        = fmt.Errorf("user not found")
	ErrInvalidCredentials  = fmt.Errorf("invalid credentials")
)

// User represents an authenticated user with stored credentials.
type User struct {
	PasswordHash string
}

// FindUserByUsername retrieves a user record by their username.
func FindUserByUsername(username string) *User {
	return &User{}
}

// FindUserByID retrieves a user record by their unique identifier.
func FindUserByID(id string) *User {
	return &User{}
}

// HashPassword generates a hash of the provided password for secure storage.
func HashPassword(password string) string {
	return password
}

// UpdateLastLogin records the timestamp of the user's most recent login.
func UpdateLastLogin(user *User) {}

// Claims represents JWT token claims including the user identifier.
type Claims struct {
	UserID string
}

// ParseToken decodes a JWT token string and extracts its claims.
func ParseToken(token string) *Claims {
	return &Claims{}
}

// IsTokenExpired checks whether the token claims indicate expiration.
func IsTokenExpired(claims *Claims) bool {
	return false
}
