package placement

import "testing"

func TestNewUser(t *testing.T) {
	id, username, email := 1, "johndoe", "john@example.com"
	user := NewUser(id, username, email)

	if user == nil {
		t.Fatal("NewUser returned nil")
	}
	if user.ID != id {
		t.Errorf("Expected ID %d, got %d", id, user.ID)
	}
	if user.Username != username {
		t.Errorf("Expected username %s, got %s", username, user.Username)
	}
	if user.Email != email {
		t.Errorf("Expected email %s, got %s", email, user.Email)
	}
	if !user.Active {
		t.Error("Expected new user to be active by default")
	}
}

func TestActivate(t *testing.T) {
	user := &User{ID: 1, Username: "test", Active: false}
	user.Activate()

	if !user.Active {
		t.Error("Expected user to be active after Activate()")
	}
}

func TestDeactivate(t *testing.T) {
	user := &User{ID: 1, Username: "test", Active: true}
	user.Deactivate()

	if user.Active {
		t.Error("Expected user to be inactive after Deactivate()")
	}
}

func TestDisplay(t *testing.T) {
	tests := []struct {
		name     string
		user     *User
		expected string
	}{
		{
			name:     "active user",
			user:     &User{ID: 1, Username: "john", Email: "john@example.com", Active: true},
			expected: "User john (john@example.com): active",
		},
		{
			name:     "inactive user",
			user:     &User{ID: 2, Username: "jane", Email: "jane@example.com", Active: false},
			expected: "User jane (jane@example.com): inactive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.user.Display()
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestValidateEmail(t *testing.T) {
	tests := []struct {
		name     string
		email    string
		expected bool
	}{
		{"valid email", "user@example.com", true},
		{"exactly 4 chars", "a@bc", true},    // len = 4, which is > 3
		{"exactly 3 chars", "a@b", false},    // len = 3, needs > 3
		{"minimum valid", "a@b.c", true},     // len = 5
		{"empty email", "", false},
		{"single char", "a", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &User{Email: tt.email}
			result := user.ValidateEmail()
			if result != tt.expected {
				t.Errorf("ValidateEmail(%q) = %v, expected %v", tt.email, result, tt.expected)
			}
		})
	}
}

func TestIsActive(t *testing.T) {
	tests := []struct {
		name   string
		active bool
	}{
		{"active user", true},
		{"inactive user", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &User{Active: tt.active}
			result := user.IsActive()
			if result != tt.active {
				t.Errorf("Expected IsActive() = %v, got %v", tt.active, result)
			}
		})
	}
}
