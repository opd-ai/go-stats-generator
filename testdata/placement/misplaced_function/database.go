package placement

// User represents a user in the database
type User struct {
	ID   int
	Name string
}

// Database provides data access
type Database struct {
	users []User
}

// NewDatabase creates a new database
func NewDatabase() *Database {
	return &Database{users: make([]User, 0)}
}

// GetUser retrieves a user by ID
func (db *Database) GetUser(id int) *User {
	for i := range db.users {
		if db.users[i].ID == id {
			return &db.users[i]
		}
	}
	return nil
}

// AddUser adds a user to the database
func (db *Database) AddUser(u User) {
	db.users = append(db.users, u)
}

// Helper function that calls ValidateUser and ProcessUser
// These functions are actually defined in handler.go but heavily used here
func BatchProcess(db *Database) {
	for i := 0; i < 10; i++ {
		ValidateUser(db, i)
		ValidateUser(db, i)
		ValidateUser(db, i)
		ProcessUser(db, i)
		ProcessUser(db, i)
	}
}

// Another helper that calls ValidateUser
func CheckUser(db *Database, id int) bool {
	return ValidateUser(db, id)
}

// Yet another function calling ValidateUser
func VerifyUser(db *Database, id int) bool {
	return ValidateUser(db, id)
}
