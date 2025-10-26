package utils

import (
	"regexp"
	"strings"
	"unicode"
)

func Validate(username string, min, max int) bool {
	if len(username) < min {
		return false
	}

	if len(username) > max {
		return false
	}

	return true
}

func ValidateUsername(username string, min, max int) bool {
	if len(username) < min {
		return false
	}

	if len(username) > max {
		return false
	}

	return true
}

func ValidateEmail(email string, max_length int) bool {
	if len(email) >= max_length {
		return false
	}
	
	match, err := regexp.Match(`^([\w\-_.]*[^.])(@\w+)(\.\w+(\.\w+)?[^.\W])$`, []byte(email))
	
	if err != nil {
		return false
	}
	
	return match
}

func ValidatePassword(password string) bool {
	if len(password) < 8 {
		return false
	}
    
	var (
		hasUpper   = false
		hasLower   = false
		hasDigit   = false
		hasSpecial = false
	)

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsDigit(char):
			hasDigit = true
		case strings.ContainsRune("#?!@$%^&*-", char):
			hasSpecial = true
		}
	}

    return hasUpper && hasLower && hasDigit && hasSpecial
}