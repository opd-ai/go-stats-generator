package placement

import "fmt"

// ValidateUser should be in database.go since it heavily references Database
// This function has high affinity to database.go, not handler.go
func ValidateUser(db *Database, id int) bool {
	user := db.GetUser(id)
	if user == nil {
		return false
	}
	if user.Name == "" {
		return false
	}
	return true
}

// ProcessUser also heavily uses Database methods
func ProcessUser(db *Database, userID int) {
	user := db.GetUser(userID)
	if user != nil {
		fmt.Printf("Processing user: %s\n", user.Name)
	}
}

// HandleRequest is correctly placed - it uses ValidateUser
func HandleRequest(db *Database, userID int) string {
	if !ValidateUser(db, userID) {
		return "invalid user"
	}
	user := db.GetUser(userID)
	return fmt.Sprintf("Hello, %s!", user.Name)
}
