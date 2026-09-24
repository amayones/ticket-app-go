package services

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"golang-backend/models"
	"golang-backend/repositories"
	"golang-backend/utils"
)

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

// ErrUsernameRequired kept as alias for backward compatibility.
var ErrUsernameRequired = ErrInputRequired

const (
	MaxRefreshTokensPerUser = 5
	codeGenRetries          = 5
)

// Re-export TTLs so callers have one source of truth.
const (
	AccessTokenTTL  = utils.AccessTokenTTL
	RefreshTokenTTL = utils.RefreshTokenTTL
)

// UserServiceInterface allows handlers to depend on abstraction (mockable).
// Identity is the public user CODE (CPUSER.CODE), never the numeric ID.
type UserServiceInterface interface {
	ListUsers(ctx context.Context, limit, offset int) ([]models.UserResponse, error)
	GetUserByCode(ctx context.Context, code string) (*models.User, error)
	CreateUser(ctx context.Context, username, email, password string) (string, error)
	UpdateUser(ctx context.Context, code string, input models.UpdateUserRequest) error
	DeleteUser(ctx context.Context, code string) error
	Login(ctx context.Context, username, password string) (accessToken, refreshToken string, err error)
	RefreshAccessToken(ctx context.Context, refreshToken string) (string, string, error)
	Logout(ctx context.Context, refreshToken string) error
	LogoutAll(ctx context.Context, userCode string) error
	ListRoles(ctx context.Context) ([]models.Role, error)
	CleanupExpiredTokens(ctx context.Context) (int64, error)
}

type UserService struct {
	users   repositories.UserRepositoryInterface
	refresh repositories.RefreshTokenRepositoryInterface
	roles   repositories.RoleRepositoryInterface
	jwtSecret string
}

func NewUserService(
	users repositories.UserRepositoryInterface,
	refresh repositories.RefreshTokenRepositoryInterface,
	roles repositories.RoleRepositoryInterface,
	jwtSecret string,
) (*UserService, error) {
	if err := utils.ValidateSecret(jwtSecret); err != nil {
		return nil, fmt.Errorf("invalid JWT secret: %w", err)
	}
	return &UserService{users: users, refresh: refresh, roles: roles, jwtSecret: jwtSecret}, nil
}

func (s *UserService) validateUserInput(username, email, password string) error {
	if strings.TrimSpace(username) == "" ||
		strings.TrimSpace(email) == "" ||
		strings.TrimSpace(password) == "" {
		return ErrInputRequired
	}
	if !utils.IsValidUsername(username) {
		return ErrUsernameTooShort
	}
	if !utils.IsValidEmail(email) {
		return ErrInvalidEmail
	}
	if len(password) > utils.MaxPasswordBytes {
		return ErrPasswordTooLong
	}
	if !utils.IsValidPassword(password) {
		return ErrPasswordTooShort
	}
	return nil
}

func (s *UserService) mapDBError(err error) error {
	if err == nil {
		return nil
	}
	msg := err.Error()
	// MSSQL unique violations: 2627 / 2601 (robust against constraint renames),
	// plus legacy constraint-name fallback used by existing tests/mocks.
	isUsernameViolation := strings.Contains(msg, "UQ_CPUSER_USERNAME") ||
		strings.Contains(msg, "UQ_users_username")
	isEmailViolation := strings.Contains(msg, "UQ_CPUSER_EMAIL") ||
		strings.Contains(msg, "UQ_users_email")
	isUniqueViolation := strings.Contains(msg, "2627") || strings.Contains(msg, "2601")
	switch {
	case isEmailViolation:
		return ErrEmailTaken
	case isUsernameViolation:
		return ErrUsernameTaken
	case isUniqueViolation:
		// Generic unique violation without identifiable constraint:
		// attribute to username (most common) — DB pre-checks above
		// already disambiguate the usual cases.
		return ErrUsernameTaken
	default:
		return err
	}
}

// uniqueUserCode generates a USR-XXXXXXXX code not yet taken.
func (s *UserService) uniqueUserCode(ctx context.Context) (string, error) {
	for i := 0; i < codeGenRetries; i++ {
		code, err := utils.GenerateUserCode()
		if err != nil {
			return "", err
		}
		if _, err := s.users.GetByCode(ctx, code); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return code, nil
			}
			return "", fmt.Errorf("check code: %w", err)
		}
	}
	return "", fmt.Errorf("could not generate unique user code after %d tries", codeGenRetries)
}

func (s *UserService) ListUsers(ctx context.Context, limit, offset int) ([]models.UserResponse, error) {
	users, err := s.users.List(ctx, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	out := make([]models.UserResponse, 0, len(users))
	for _, u := range users {
		out = append(out, u.ToResponse())
	}
	return out, nil
}

func (s *UserService) GetUserByCode(ctx context.Context, code string) (*models.User, error) {
	user, err := s.users.GetByCode(ctx, code)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("get user: %w", err)
	}
	return user, nil
}

func (s *UserService) CreateUser(ctx context.Context, username, email, password string) (string, error) {
	username = utils.NormalizeUsername(username)
	email = utils.NormalizeEmail(email)
	if err := s.validateUserInput(username, email, password); err != nil {
		return "", err
	}
	// Pre-check for friendlier errors (DB constraint remains source of truth).
	if _, err := s.users.GetByUsername(ctx, username); err == nil {
		return "", ErrUsernameTaken
	} else if !errors.Is(err, sql.ErrNoRows) {
		return "", fmt.Errorf("check username: %w", err)
	}
	if _, err := s.users.GetByEmail(ctx, email); err == nil {
		return "", ErrEmailTaken
	} else if !errors.Is(err, sql.ErrNoRows) {
		return "", fmt.Errorf("check email: %w", err)
	}
	hashed, err := utils.HashPassword(password)
	if err != nil {
		return "", err
	}
	code, err := s.uniqueUserCode(ctx)
	if err != nil {
		return "", err
	}
	user := &models.User{
		Code:     code,
		Username: username,
		Email:    email,
		Password: hashed,
		RoleCode: models.DefaultRoleCode,
	}
	if err := s.users.Create(ctx, user); err != nil {
		return "", s.mapDBError(err)
	}
	return code, nil
}

func (s *UserService) UpdateUser(ctx context.Context, code string, input models.UpdateUserRequest) error {
	existing, err := s.users.GetByCode(ctx, code)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrUserNotFound
		}
		return fmt.Errorf("get user for update: %w", err)
	}
	username := existing.Username
	email := existing.Email
	passwordHash := existing.Password

	if input.Username != nil {
		username = utils.NormalizeUsername(*input.Username)
	}
	if input.Email != nil {
		email = utils.NormalizeEmail(*input.Email)
	}
	if input.Password != nil && *input.Password != "" {
		if len(*input.Password) > utils.MaxPasswordBytes {
			return ErrPasswordTooLong
		}
		if !utils.IsValidPassword(*input.Password) {
			return ErrPasswordTooShort
		}
		hash, err := utils.HashPassword(*input.Password)
		if err != nil {
			return err
		}
		passwordHash = hash
	}
	if strings.TrimSpace(username) == "" || strings.TrimSpace(email) == "" {
		return ErrInputRequired
	}
	if !utils.IsValidUsername(username) {
		return ErrUsernameTooShort
	}
	if !utils.IsValidEmail(email) {
		return ErrInvalidEmail
	}
	updated := &models.User{Code: code, Username: username, Email: email, Password: passwordHash}
	if err := s.users.Update(ctx, updated); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrUserNotFound
		}
		return s.mapDBError(err)
	}
	return nil
}

func (s *UserService) DeleteUser(ctx context.Context, code string) error {
	if err := s.users.Delete(ctx, code); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrUserNotFound
		}
		return fmt.Errorf("delete user: %w", err)
	}
	// Best-effort session cleanup (CASCADE in DB is the backstop).
	_ = s.refresh.DeleteByUserCode(ctx, code)
	return nil
}

func (s *UserService) evictIfNeeded(ctx context.Context, userCode string) error {
	count, err := s.refresh.CountByUserCode(ctx, userCode)
	if err != nil {
		return fmt.Errorf("count refresh tokens: %w", err)
	}
	for count >= MaxRefreshTokensPerUser {
		if err := s.refresh.DeleteOldestByUserCode(ctx, userCode); err != nil {
			return fmt.Errorf("evict oldest token: %w", err)
		}
		count--
	}
	return nil
}

func (s *UserService) Login(ctx context.Context, username, password string) (string, string, error) {
	username = utils.NormalizeUsername(username)
	if username == "" || password == "" {
		return "", "", ErrInvalidLogin
	}
	user, err := s.users.GetByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", "", ErrInvalidLogin
		}
		return "", "", fmt.Errorf("login lookup: %w", err)
	}
	if !utils.CheckPasswordHash(password, user.Password) {
		return "", "", ErrInvalidLogin
	}
	access, err := utils.GenerateAccessToken(s.jwtSecret, user.Code, user.Username, user.RoleCode)
	if err != nil {
		return "", "", err
	}
	refresh, err := utils.GenerateRefreshToken()
	if err != nil {
		return "", "", err
	}
	if err := s.evictIfNeeded(ctx, user.Code); err != nil {
		return "", "", err
	}
	rt := models.RefreshToken{
		UserCode:  user.Code,
		Token:     utils.HashRefreshToken(refresh),
		ExpiresAt: time.Now().Add(RefreshTokenTTL),
	}
	if err := s.refresh.Create(ctx, &rt); err != nil {
		return "", "", fmt.Errorf("store refresh token: %w", err)
	}
	return access, refresh, nil
}

func (s *UserService) RefreshAccessToken(ctx context.Context, refreshToken string) (string, string, error) {
	if strings.TrimSpace(refreshToken) == "" {
		return "", "", ErrInvalidRefresh
	}
	hash := utils.HashRefreshToken(strings.TrimSpace(refreshToken))
	rt, err := s.refresh.GetByTokenHash(ctx, hash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", "", ErrInvalidRefresh
		}
		return "", "", fmt.Errorf("lookup refresh token: %w", err)
	}
	if time.Now().After(rt.ExpiresAt) {
		_, _ = s.refresh.DeleteByTokenHash(ctx, hash)
		return "", "", ErrInvalidRefresh
	}
	user, err := s.users.GetByCode(ctx, rt.UserCode)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			_, _ = s.refresh.DeleteByTokenHash(ctx, hash)
			return "", "", ErrInvalidRefresh
		}
		return "", "", fmt.Errorf("refresh user lookup: %w", err)
	}
	newAccess, err := utils.GenerateAccessToken(s.jwtSecret, user.Code, user.Username, user.RoleCode)
	if err != nil {
		return "", "", err
	}
	newRefresh, err := utils.GenerateRefreshToken()
	if err != nil {
		return "", "", err
	}
	// Rotate: invalidate old first, then store new. Crash between the two
	// forces re-login (safe) instead of session duplication.
	if _, err := s.refresh.DeleteByTokenHash(ctx, hash); err != nil {
		return "", "", fmt.Errorf("invalidate old token: %w", err)
	}
	newRT := models.RefreshToken{
		UserCode:  user.Code,
		Token:     utils.HashRefreshToken(newRefresh),
		ExpiresAt: time.Now().Add(RefreshTokenTTL),
	}
	if err := s.refresh.Create(ctx, &newRT); err != nil {
		return "", "", fmt.Errorf("store rotated token: %w", err)
	}
	return newAccess, newRefresh, nil
}

func (s *UserService) Logout(ctx context.Context, refreshToken string) error {
	if strings.TrimSpace(refreshToken) == "" {
		return ErrInvalidRefresh
	}
	deleted, err := s.refresh.DeleteByTokenHash(ctx, utils.HashRefreshToken(strings.TrimSpace(refreshToken)))
	if err != nil {
		return fmt.Errorf("logout: %w", err)
	}
	if !deleted {
		return ErrInvalidRefresh
	}
	return nil
}

func (s *UserService) LogoutAll(ctx context.Context, userCode string) error {
	return s.refresh.DeleteByUserCode(ctx, userCode)
}

func (s *UserService) ListRoles(ctx context.Context) ([]models.Role, error) {
	roles, err := s.roles.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("list roles: %w", err)
	}
	return roles, nil
}

func (s *UserService) CleanupExpiredTokens(ctx context.Context) (int64, error) {
	return s.refresh.DeleteExpired(ctx)
}
