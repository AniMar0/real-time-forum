package backend

import (
	"regexp"
	"strings"
	"unicode"
)

// isValidEmail validates email format
func isValidEmail(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}

// isValidPassword checks password strength (min 8 chars, uppercase, lowercase, number)
func isValidPassword(password string) bool {
	if len(password) < 8 {
		return false
	}
	var hasUpper, hasLower, hasNumber bool
	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsNumber(char):
			hasNumber = true
		}
	}
	return hasUpper && hasLower && hasNumber
}

// isValidTextLength validates text length within min and max limits
func isValidTextLength(text string, min, max int) bool {
	length := len(strings.TrimSpace(text))
	return length >= min && length <= max
}

// isValidAge validates age range
func isValidAge(age int) bool {
	return age >= 13 && age <= 120
}

// isValidNickname validates nickname (alphanumeric, underscore, 3-20 chars)
func isValidNickname(nickname string) bool {
	if len(nickname) < 3 || len(nickname) > 20 {
		return false
	}
	nicknameRegex := regexp.MustCompile(`^[a-zA-Z0-9_]+$`)
	return nicknameRegex.MatchString(nickname)
}
