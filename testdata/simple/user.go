package simple

import (
	"fmt"
	"time"
)

// User represents a user in the system
type User struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	IsActive  bool      `json:"is_active"`
}

// UserService provides user-related operations
type UserService struct {
	users map[int]*User
}

// NewUserService creates a new user service
func NewUserService() *UserService {
	return &UserService{
		users: make(map[int]*User),
	}
}

// CreateUser creates a new user
func (us *UserService) CreateUser(name, email string) (*User, error) {
	if name == "" {
		return nil, fmt.Errorf("name cannot be empty")
	}

	if email == "" {
		return nil, fmt.Errorf("email cannot be empty")
	}

	user := &User{
		ID:        len(us.users) + 1,
		Name:      name,
		Email:     email,
		CreatedAt: time.Now(),
		IsActive:  true,
	}

	us.users[user.ID] = user
	return user, nil
}

// GetUser retrieves a user by ID
func (us *UserService) GetUser(id int) (*User, error) {
	user, exists := us.users[id]
	if !exists {
		return nil, fmt.Errorf("user with ID %d not found", id)
	}
	return user, nil
}

// GetAllUsers returns all users
func (us *UserService) GetAllUsers() []*User {
	users := make([]*User, 0, len(us.users))
	for _, user := range us.users {
		users = append(users, user)
	}
	return users
}

// UpdateUser updates an existing user
func (us *UserService) UpdateUser(id int, name, email string) error {
	user, exists := us.users[id]
	if !exists {
		return fmt.Errorf("user with ID %d not found", id)
	}

	if name != "" {
		user.Name = name
	}

	if email != "" {
		user.Email = email
	}

	return nil
}

// DeleteUser deletes a user by ID
func (us *UserService) DeleteUser(id int) error {
	_, exists := us.users[id]
	if !exists {
		return fmt.Errorf("user with ID %d not found", id)
	}

	delete(us.users, id)
	return nil
}

// DeactivateUser deactivates a user
func (us *UserService) DeactivateUser(id int) error {
	user, exists := us.users[id]
	if !exists {
		return fmt.Errorf("user with ID %d not found", id)
	}

	user.IsActive = false
	return nil
}

// GetActiveUsers returns only active users
func (us *UserService) GetActiveUsers() []*User {
	var activeUsers []*User
	for _, user := range us.users {
		if user.IsActive {
			activeUsers = append(activeUsers, user)
		}
	}
	return activeUsers
}

// ComplexFunction demonstrates a function with higher complexity
func (us *UserService) ComplexFunction(criteria map[string]interface{}) ([]*User, error) {
	var results []*User

	for _, user := range us.users {
		match := true

		if name, ok := criteria["name"]; ok {
			if nameStr, ok := name.(string); ok {
				if user.Name != nameStr {
					match = false
				}
			}
		}

		if email, ok := criteria["email"]; ok {
			if emailStr, ok := email.(string); ok {
				if user.Email != emailStr {
					match = false
				}
			}
		}

		if active, ok := criteria["active"]; ok {
			if activeBool, ok := active.(bool); ok {
				if user.IsActive != activeBool {
					match = false
				}
			}
		}

		if createdAfter, ok := criteria["created_after"]; ok {
			if timeVal, ok := createdAfter.(time.Time); ok {
				if user.CreatedAt.Before(timeVal) {
					match = false
				}
			}
		}

		if match {
			results = append(results, user)
		}
	}

	return results, nil
}
