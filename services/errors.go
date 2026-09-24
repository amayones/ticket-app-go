package services

// Shared error values used by every feature (features/<menu>).
// Handlers map these to HTTP statuses via errors.Is; keep them here
// so cross-feature checks stay comparable (no duplicates).
import "errors"

var (
	ErrInputRequired    = errors.New("username, email, and password are required")
	ErrUsernameTooShort = errors.New("username must be at least 3 characters")
	ErrInvalidEmail     = errors.New("invalid email format")
	ErrPasswordTooShort = errors.New("password must be at least 8 characters")
	ErrPasswordTooLong  = errors.New("password must not exceed 72 bytes")
	ErrUsernameTaken    = errors.New("username already taken")
	ErrEmailTaken       = errors.New("email already registered")
	ErrInvalidLogin     = errors.New("invalid username or password")
	ErrUserNotFound     = errors.New("user not found")
	ErrInvalidRefresh   = errors.New("invalid or expired refresh token")
	ErrForbidden        = errors.New("forbidden")
)
