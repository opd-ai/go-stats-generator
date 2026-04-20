package placement

import "testing"

func TestNewUser(t *testing.T) {
	u := NewUser(1, "alice", "alice@example.com")
	if u == nil {
		t.Fatal("NewUser returned nil")
	}
	if u.ID != 1 || u.Username != "alice" || u.Email != "alice@example.com" {
		t.Errorf("unexpected user fields: %+v", u)
	}
}

func TestValidate_Valid(t *testing.T) {
	u := &User{ID: 1, Username: "bob", Email: "bob@example.com"}
	if !u.Validate() {
		t.Error("expected Validate to return true for complete user")
	}
}

func TestValidate_MissingUsername(t *testing.T) {
	u := &User{ID: 1, Username: "", Email: "x@y.com"}
	if u.Validate() {
		t.Error("expected Validate to return false when username is empty")
	}
}

func TestValidate_MissingEmail(t *testing.T) {
	u := &User{ID: 1, Username: "carol", Email: ""}
	if u.Validate() {
		t.Error("expected Validate to return false when email is empty")
	}
}

func TestValidate_InvalidEmail(t *testing.T) {
	u := &User{ID: 1, Username: "dave", Email: "nodomain"}
	if u.Validate() {
		t.Error("expected Validate to return false when email has no @")
	}
}

func TestIsAdmin_True(t *testing.T) {
	u := &User{ID: 1, Username: "admin_charlie"}
	if !u.IsAdmin() {
		t.Error("expected IsAdmin to return true for admin_ prefix")
	}
}

func TestIsAdmin_False(t *testing.T) {
	u := &User{ID: 1, Username: "regular"}
	if u.IsAdmin() {
		t.Error("expected IsAdmin to return false for non-admin user")
	}
}
