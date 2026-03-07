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

// BatchProcess demonstrates a misplaced function that heavily calls ValidateUser and ProcessUser from handler.go.
// This function illustrates poor code placement where a function resides in the wrong package relative to its dependencies.
func BatchProcess(db *Database) {
	for i := 0; i < 10; i++ {
		ValidateUser(db, i)
		ValidateUser(db, i)
		ValidateUser(db, i)
		ProcessUser(db, i)
		ProcessUser(db, i)
	}
}

// CheckUser validates a user by ID using the ValidateUser function from handler.go.
// This function demonstrates code placement issues where dependencies are in different packages.
func CheckUser(db *Database, id int) bool {
	return ValidateUser(db, id)
}

// VerifyUser confirms user existence by delegating to ValidateUser from handler.go.
// This function illustrates misplaced code where heavy cross-package dependencies indicate poor module organization.
func VerifyUser(db *Database, id int) bool {
	return ValidateUser(db, id)
}
