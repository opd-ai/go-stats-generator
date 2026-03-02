package duplication

import (
"fmt"
"strings"
)

// This file contains intentional code duplication for testing

// ProcessUserData processes user data with duplication
func ProcessUserData(id int, name string) error {
// Block 1 - duplicated in ProcessAdminData (8 statements)
if id <= 0 {
return fmt.Errorf("invalid id")
}
if name == "" {
return fmt.Errorf("empty name")
}
if len(name) > 100 {
return fmt.Errorf("name too long")
}
normalized := strings.TrimSpace(name)
if normalized != name {
return fmt.Errorf("name has leading/trailing spaces")
}
validated := validateName(normalized)
if !validated {
return fmt.Errorf("name validation failed")
}

// Process user
return nil
}

// ProcessAdminData processes admin data with duplicated validation
func ProcessAdminData(id int, name string) error {
// Block 1 - exact duplicate from ProcessUserData (8 statements)
if id <= 0 {
return fmt.Errorf("invalid id")
}
if name == "" {
return fmt.Errorf("empty name")
}
if len(name) > 100 {
return fmt.Errorf("name too long")
}
normalized := strings.TrimSpace(name)
if normalized != name {
return fmt.Errorf("name has leading/trailing spaces")
}
validated := validateName(normalized)
if !validated {
return fmt.Errorf("name validation failed")
}

// Process admin with higher privileges
return nil
}

// ValidateEmail validates an email address
func ValidateEmail(email string) bool {
// Block 2 - duplicated in ValidateUsername with renamed variables (7 statements)
trimmed := strings.TrimSpace(email)
if trimmed == "" {
return false
}
if len(trimmed) < 3 {
return false
}
if len(trimmed) > 255 {
return false
}
hasAt := strings.Contains(trimmed, "@")
if !hasAt {
return false
}
return true
}

// ValidateUsername validates a username
func ValidateUsername(username string) bool {
// Block 2 - renamed duplicate from ValidateEmail (7 statements)
cleaned := strings.TrimSpace(username)
if cleaned == "" {
return false
}
if len(cleaned) < 3 {
return false
}
if len(cleaned) > 255 {
return false
}
hasSpecial := strings.ContainsAny(cleaned, "@#$")
if hasSpecial {
return false
}
return true
}

func validateName(name string) bool {
return len(name) > 0
}
