package utils

import "regexp"

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

func IsValidEmail(email string) bool {
	return emailRegex.MatchString(email)
}

func IsValidPassword(password string) bool {
	return len(password) >= 8
}

func IsValidUsername(username string) bool {
	return len(username) >= 3
}
