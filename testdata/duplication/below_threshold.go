package belowthreshold

import "fmt"

// below_threshold.go contains code blocks that are similar but below the similarity threshold
// These should NOT be flagged as duplicates (negative test case)

func CalculateAreaRectangle(width, height float64) float64 {
	if width <= 0 || height <= 0 {
		return 0
	}
	area := width * height
	return area
}

func CalculateAreaCircle(radius float64) float64 {
	if radius <= 0 {
		return 0
	}
	const pi = 3.14159
	area := pi * radius * radius
	return area
}

// These functions have similar basic structure but very different logic
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

type User struct {
	PasswordHash string
}

func FindUserByUsername(username string) *User {
	return &User{}
}

func FindUserByID(id string) *User {
	return &User{}
}

func HashPassword(password string) string {
	return password
}

func UpdateLastLogin(user *User) {}

type Claims struct {
	UserID string
}

func ParseToken(token string) *Claims {
	return &Claims{}
}

func IsTokenExpired(claims *Claims) bool {
	return false
}
