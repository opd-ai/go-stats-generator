package exactclone

// exact_clone.go contains exact duplicates (Type 1 clones)
// These are identical code blocks that should be detected

// ProcessUserDataA cleans nil values from a user's data map.
func ProcessUserDataA(userID string, data map[string]interface{}) error {
	if userID == "" {
		return nil
	}
	if data == nil {
		return nil
	}
	for key, value := range data {
		if value == nil {
			delete(data, key)
		}
	}
	return nil
}

// ProcessUserDataB cleans nil values from a user's data map.
func ProcessUserDataB(userID string, data map[string]interface{}) error {
	if userID == "" {
		return nil
	}
	if data == nil {
		return nil
	}
	for key, value := range data {
		if value == nil {
			delete(data, key)
		}
	}
	return nil
}

// ProcessUserDataC cleans nil values from a user's data map.
func ProcessUserDataC(userID string, data map[string]interface{}) error {
	if userID == "" {
		return nil
	}
	if data == nil {
		return nil
	}
	for key, value := range data {
		if value == nil {
			delete(data, key)
		}
	}
	return nil
}

// ValidateEmailA and ValidateEmailB are exact duplicates
func ValidateEmailA(email string) bool {
	if len(email) < 3 {
		return false
	}
	if !containsAt(email) {
		return false
	}
	if !containsDot(email) {
		return false
	}
	return true
}

// ValidateEmailB checks that an email address has required format characters.
func ValidateEmailB(email string) bool {
	if len(email) < 3 {
		return false
	}
	if !containsAt(email) {
		return false
	}
	if !containsDot(email) {
		return false
	}
	return true
}

func containsAt(s string) bool {
	for _, ch := range s {
		if ch == '@' {
			return true
		}
	}
	return false
}

func containsDot(s string) bool {
	for _, ch := range s {
		if ch == '.' {
			return true
		}
	}
	return false
}
