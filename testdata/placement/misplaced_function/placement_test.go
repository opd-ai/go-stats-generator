package placement

import "testing"

func TestNewDatabase(t *testing.T) {
	db := NewDatabase()
	if db == nil {
		t.Fatal("NewDatabase returned nil")
	}
}

func TestAddAndGetUser(t *testing.T) {
	db := NewDatabase()
	db.AddUser(User{ID: 1, Name: "Alice"})

	user := db.GetUser(1)
	if user == nil {
		t.Fatal("GetUser returned nil for existing user")
	}
	if user.Name != "Alice" {
		t.Errorf("expected Name=Alice, got %s", user.Name)
	}
}

func TestGetUser_NotFound(t *testing.T) {
	db := NewDatabase()
	if db.GetUser(99) != nil {
		t.Error("expected nil for non-existent user")
	}
}

func TestValidateUser(t *testing.T) {
	db := NewDatabase()
	db.AddUser(User{ID: 1, Name: "Bob"})

	if !ValidateUser(db, 1) {
		t.Error("expected ValidateUser to return true for valid user")
	}
	if ValidateUser(db, 99) {
		t.Error("expected ValidateUser to return false for missing user")
	}
}

func TestValidateUser_EmptyName(t *testing.T) {
	db := NewDatabase()
	db.AddUser(User{ID: 1, Name: ""})
	if ValidateUser(db, 1) {
		t.Error("expected ValidateUser to return false for user with empty name")
	}
}

func TestProcessUser(t *testing.T) {
	db := NewDatabase()
	db.AddUser(User{ID: 1, Name: "Carol"})
	// ProcessUser prints to stdout; just verify it doesn't panic
	ProcessUser(db, 1)
	ProcessUser(db, 99) // non-existent user — should be a no-op
}

func TestHandleRequest(t *testing.T) {
	db := NewDatabase()
	db.AddUser(User{ID: 1, Name: "Dave"})

	got := HandleRequest(db, 1)
	want := "Hello, Dave!"
	if got != want {
		t.Errorf("HandleRequest returned %q, want %q", got, want)
	}
}

func TestHandleRequest_InvalidUser(t *testing.T) {
	db := NewDatabase()
	got := HandleRequest(db, 42)
	if got != "invalid user" {
		t.Errorf("expected 'invalid user', got %q", got)
	}
}

func TestBatchProcess(t *testing.T) {
	db := NewDatabase()
	for i := 1; i <= 3; i++ {
		db.AddUser(User{ID: i, Name: "User"})
	}
	// Just verify no panic
	BatchProcess(db)
}

func TestCheckUser(t *testing.T) {
	db := NewDatabase()
	db.AddUser(User{ID: 1, Name: "Eve"})

	if !CheckUser(db, 1) {
		t.Error("expected CheckUser to return true for existing user")
	}
	if CheckUser(db, 99) {
		t.Error("expected CheckUser to return false for missing user")
	}
}

func TestVerifyUser(t *testing.T) {
	db := NewDatabase()
	db.AddUser(User{ID: 1, Name: "Frank"})

	if !VerifyUser(db, 1) {
		t.Error("expected VerifyUser to return true for existing user")
	}
	if VerifyUser(db, 99) {
		t.Error("expected VerifyUser to return false for missing user")
	}
}
