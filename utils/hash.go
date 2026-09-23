package utils

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
)

// MaxPasswordBytes is bcrypt's hard limit (72 bytes). Longer inputs are
// silently truncated by bcrypt, so reject them explicitly.
const MaxPasswordBytes = 72

var ErrPasswordTooLong = errors.New("password must not exceed 72 bytes")

func HashPassword(password string) (string, error) {
	if len(password) > MaxPasswordBytes {
		return "", ErrPasswordTooLong
	}
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func CheckPasswordHash(password, hash string) bool {
	if password == "" || hash == "" {
		return false
	}
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
