package utils

import (
	"regexp"
	"strings"
)

const (
	MaxUsernameLen = 50
	MaxEmailLen    = 254
	MinUsernameLen = 3
	MinPasswordLen = 8
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

// NormalizeEmail trims spaces and lowercases. Call before validate+store
// so "  Budi@X.com " and "budi@x.com" are treated as one identity.
func NormalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

// NormalizeUsername trims surrounding spaces (usernames keep case).
func NormalizeUsername(username string) string {
	return strings.TrimSpace(username)
}

func IsValidEmail(email string) bool {
	email = NormalizeEmail(email)
	if email == "" || len(email) > MaxEmailLen {
		return false
	}
	if strings.Contains(email, "..") {
		return false
	}
	return emailRegex.MatchString(email)
}

func IsValidPassword(password string) bool {
	return len(password) >= MinPasswordLen && len(password) <= MaxPasswordBytes
}

func IsValidUsername(username string) bool {
	username = NormalizeUsername(username)
	if len(username) < MinUsernameLen || len(username) > MaxUsernameLen {
		return false
	}
	// Reject control characters / newlines that break logs and UIs.
	for _, r := range username {
		if r < 0x20 || r == 0x7f {
			return false
		}
	}
	return true
}
