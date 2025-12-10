package util

import (
	"errors"
	"regexp"
	"strings"
	"unicode"
)

// Username validation constants
const (
	MinUsernameLength = 3
	MaxUsernameLength = 255
)

// Password validation constants
const (
	MinPasswordLength = 8
	MaxPasswordLength = 256
)

var usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

type PasswordStrength struct {
	HasMinLength  bool
	HasUppercase  bool
	HasLowercase  bool
	HasNumber     bool
	HasSpecial    bool
	ComplexityMet bool
}

// ValidateUsername validates username format and constraints
func ValidateUsername(username string) error {
	if len(username) < MinUsernameLength {
		return errors.New("username must be at least 3 characters")
	}
	if len(username) > MaxUsernameLength {
		return errors.New("username must not exceed 255 characters")
	}

	if !usernameRegex.MatchString(username) {
		return errors.New("username must contain only letters, numbers, underscores, and hyphens")
	}

	// Check against reserved words
	reservedWords := []string{"admin", "root", "system", "api", "null", "undefined"}
	usernameLower := strings.ToLower(username)
	for _, reserved := range reservedWords {
		if usernameLower == reserved {
			return errors.New("username is reserved")
		}
	}

	return nil
}

// ValidatePasswordStrength validates password strength and complexity
func ValidatePasswordStrength(password string) (*PasswordStrength, error) {
	strength := &PasswordStrength{}

	if len(password) < MinPasswordLength {
		return strength, errors.New("password must be at least 8 characters")
	}
	if len(password) > MaxPasswordLength {
		return strength, errors.New("password must not exceed 256 characters")
	}
	strength.HasMinLength = true

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			strength.HasUppercase = true
		case unicode.IsLower(char):
			strength.HasLowercase = true
		case unicode.IsNumber(char):
			strength.HasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			strength.HasSpecial = true
		}
	}

	// Require at least 2 of: uppercase, number, special character
	complexityCount := 0
	if strength.HasUppercase {
		complexityCount++
	}
	if strength.HasNumber {
		complexityCount++
	}
	if strength.HasSpecial {
		complexityCount++
	}

	if complexityCount < 2 {
		return strength, errors.New("password must contain at least 2 of: uppercase letter, number, special character")
	}

	strength.ComplexityMet = true
	return strength, nil
}
