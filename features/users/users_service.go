package users

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"golang-backend/models"
	"golang-backend/services"
	"golang-backend/utils"
)

const (
	MaxRefreshTokensPerUser = 5
	codeGenRetries          = 5
)

// ServiceInterface allows handlers to depend on abstraction (mockable).
// Identity is the public user CODE (CPUSER.CODE), never the numeric ID.
type ServiceInterface interface {
	ListUsers(ctx context.Context, limit, offset int) ([]models.UserResponse, error)
	GetUserByCode(ctx context.Context, code string) (*models.User, error)
	CreateUser(ctx context.Context, username, email, password, roleCode string) (string, error)
	UpdateUser(ctx context.Context, code string, input models.UpdateUserRequest) error
	DeleteUser(ctx context.Context, code string) error
	Login(ctx context.Context, username, password string) (accessToken, refreshToken string, err error)
	RefreshAccessToken(ctx context.Context, refreshToken string) (string, string, error)
	Logout(ctx context.Context, refreshToken string) error
	LogoutAll(ctx context.Context, userCode string) error
	CountUsers(ctx context.Context) (int, error)
}

// roleChecker dipenuhi roles.Repository (tanpa import antar-fitur).
type roleChecker interface {
	RoleExists(ctx context.Context, code string) (bool, error)
}

// refreshStore dipenuhi sessions.Repository (tanpa import antar-fitur).
type refreshStore interface {
	Create(ctx context.Context, token *models.RefreshToken) error
	GetByTokenHash(ctx context.Context, tokenHash string) (*models.RefreshToken, error)
	DeleteByTokenHash(ctx context.Context, tokenHash string) (bool, error)
	DeleteByUserCode(ctx context.Context, userCode string) error
	DeleteExpired(ctx context.Context) (int64, error)
	CountByUserCode(ctx context.Context, userCode string) (int, error)
	DeleteOldestByUserCode(ctx context.Context, userCode string) error
}

type Service struct {
	users       RepositoryInterface
	refresh     refreshStore
	roles       roleChecker
	jwtSecret   string
	accessTTL   time.Duration
	refreshTTL  time.Duration
}

func NewService(
	users RepositoryInterface,
	refresh refreshStore,
	roles roleChecker,
	jwtSecret string,
	accessTTL, refreshTTL time.Duration,
) (*Service, error) {
	if err := utils.ValidateSecret(jwtSecret); err != nil {
		return nil, fmt.Errorf("invalid JWT secret: %w", err)
	}
	if accessTTL <= 0 {
		accessTTL = utils.AccessTokenTTL
	}
	if refreshTTL <= 0 {
		refreshTTL = utils.RefreshTokenTTL
	}
	return &Service{users: users, refresh: refresh, roles: roles, jwtSecret: jwtSecret, accessTTL: accessTTL, refreshTTL: refreshTTL}, nil
}

func (s *Service) validateUserInput(username, email, password string) error {
	if strings.TrimSpace(username) == "" ||
		strings.TrimSpace(email) == "" ||
		strings.TrimSpace(password) == "" {
		return services.ErrInputRequired
	}
	if !utils.IsValidUsername(username) {
		return services.ErrUsernameTooShort
	}
	if !utils.IsValidEmail(email) {
		return services.ErrInvalidEmail
	}
	if len(password) > utils.MaxPasswordBytes {
		return services.ErrPasswordTooLong
	}
	if !utils.IsValidPassword(password) {
		return services.ErrPasswordTooShort
	}
	return nil
}

func (s *Service) mapDBError(err error) error {
	if err == nil {
		return nil
	}
	msg := err.Error()
	lower := strings.ToLower(msg)
	// MSSQL unique violations: 2627 / 2601 (robust against constraint renames),
	// plus legacy constraint-name fallback used by existing tests/mocks.
	// Postgres: "duplicate key value violates unique constraint ...",
	// SQLite (modernc): "UNIQUE constraint failed: CPUSER.EMAIL".
	isUsernameViolation := strings.Contains(msg, "UQ_CPUSER_USERNAME") ||
		strings.Contains(msg, "UQ_users_username")
	isEmailViolation := strings.Contains(msg, "UQ_CPUSER_EMAIL") ||
		strings.Contains(msg, "UQ_users_email")
	isUniqueViolation := strings.Contains(msg, "2627") || strings.Contains(msg, "2601") ||
		strings.Contains(lower, "unique constraint failed") ||
		strings.Contains(lower, "duplicate key")
	switch {
	case isEmailViolation:
		return services.ErrEmailTaken
	case isUsernameViolation:
		return services.ErrUsernameTaken
	case isUniqueViolation:
		// Postgres/SQLite sering tidak menyebut nama constraint secara jelas
		// di pesan; bedakan via nama kolom bila ada, jika ambigu kembalikan
		// error mentah agar tidak salah menuduh username.
		if strings.Contains(lower, "email") {
			return services.ErrEmailTaken
		}
		if strings.Contains(lower, "username") {
			return services.ErrUsernameTaken
		}
		return err
	default:
		return err
	}
}

// uniqueUserCode generates a USR-XXXXXXXX code not yet taken.
func (s *Service) uniqueUserCode(ctx context.Context) (string, error) {
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

func (s *Service) ListUsers(ctx context.Context, limit, offset int) ([]models.UserResponse, error) {
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

func (s *Service) GetUserByCode(ctx context.Context, code string) (*models.User, error) {
	user, err := s.users.GetByCode(ctx, code)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, services.ErrUserNotFound
		}
		return nil, fmt.Errorf("get user: %w", err)
	}
	return user, nil
}

func (s *Service) CreateUser(ctx context.Context, username, email, password, roleCode string) (string, error) {
	username = utils.NormalizeUsername(username)
	email = utils.NormalizeEmail(email)
	if err := s.validateUserInput(username, email, password); err != nil {
		return "", err
	}
	roleCode = strings.ToUpper(strings.TrimSpace(roleCode))
	if roleCode == "" {
		roleCode = models.DefaultRoleCode
	} else if ok, err := s.roles.RoleExists(ctx, roleCode); err != nil {
		return "", fmt.Errorf("check role: %w", err)
	} else if !ok {
		return "", services.ErrRoleNotFound
	}
	// Pre-check for friendlier errors (DB constraint remains source of truth).
	if _, err := s.users.GetByUsername(ctx, username); err == nil {
		return "", services.ErrUsernameTaken
	} else if !errors.Is(err, sql.ErrNoRows) {
		return "", fmt.Errorf("check username: %w", err)
	}
	if _, err := s.users.GetByEmail(ctx, email); err == nil {
		return "", services.ErrEmailTaken
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
		RoleCode: roleCode,
	}
	if err := s.users.Create(ctx, user); err != nil {
		return "", s.mapDBError(err)
	}
	return code, nil
}

func (s *Service) UpdateUser(ctx context.Context, code string, input models.UpdateUserRequest) error {
	existing, err := s.users.GetByCode(ctx, code)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return services.ErrUserNotFound
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
			return services.ErrPasswordTooLong
		}
		if !utils.IsValidPassword(*input.Password) {
			return services.ErrPasswordTooShort
		}
		hash, err := utils.HashPassword(*input.Password)
		if err != nil {
			return err
		}
		passwordHash = hash
	}
	if strings.TrimSpace(username) == "" || strings.TrimSpace(email) == "" {
		return services.ErrInputRequired
	}
	if !utils.IsValidUsername(username) {
		return services.ErrUsernameTooShort
	}
	if !utils.IsValidEmail(email) {
		return services.ErrInvalidEmail
	}
	// Pre-check duplikat (abaikan milik sendiri). Constraint DB tetap
	// menjadi sumber kebenaran untuk race condition.
	// NOTE: RoleCode sengaja tidak diubah di sini; gunakan UpdateUserRole.
	if !strings.EqualFold(username, existing.Username) {
		if _, err := s.users.GetByUsername(ctx, username); err == nil {
			return services.ErrUsernameTaken
		} else if !errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("check username: %w", err)
		}
	}
	if !strings.EqualFold(email, existing.Email) {
		if _, err := s.users.GetByEmail(ctx, email); err == nil {
			return services.ErrEmailTaken
		} else if !errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("check email: %w", err)
		}
	}
	updated := &models.User{Code: code, Username: username, Email: email, Password: passwordHash, RoleCode: existing.RoleCode}
	if err := s.users.Update(ctx, updated); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return services.ErrUserNotFound
		}
		return s.mapDBError(err)
	}
	return nil
}

func (s *Service) DeleteUser(ctx context.Context, code string) error {
	if err := s.users.Delete(ctx, code); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return services.ErrUserNotFound
		}
		return fmt.Errorf("delete user: %w", err)
	}
	// Best-effort session cleanup (CASCADE in DB is the backstop).
	_ = s.refresh.DeleteByUserCode(ctx, code)
	return nil
}

func (s *Service) evictIfNeeded(ctx context.Context, userCode string) error {
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

func (s *Service) Login(ctx context.Context, username, password string) (string, string, error) {
	username = utils.NormalizeUsername(username)
	if username == "" || password == "" {
		return "", "", services.ErrInvalidLogin
	}
	user, err := s.users.GetByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", "", services.ErrInvalidLogin
		}
		return "", "", fmt.Errorf("login lookup: %w", err)
	}
	if !utils.CheckPasswordHash(password, user.Password) {
		return "", "", services.ErrInvalidLogin
	}
	access, err := s.accessTokenFor(user)
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
		ExpiresAt: time.Now().Add(s.refreshTTL),
	}
	if err := s.refresh.Create(ctx, &rt); err != nil {
		return "", "", fmt.Errorf("store refresh token: %w", err)
	}
	return access, refresh, nil
}

func (s *Service) RefreshAccessToken(ctx context.Context, refreshToken string) (string, string, error) {
	if strings.TrimSpace(refreshToken) == "" {
		return "", "", services.ErrInvalidRefresh
	}
	hash := utils.HashRefreshToken(strings.TrimSpace(refreshToken))
	rt, err := s.refresh.GetByTokenHash(ctx, hash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", "", services.ErrInvalidRefresh
		}
		return "", "", fmt.Errorf("lookup refresh token: %w", err)
	}
	if time.Now().After(rt.ExpiresAt) {
		_, _ = s.refresh.DeleteByTokenHash(ctx, hash)
		return "", "", services.ErrInvalidRefresh
	}
	user, err := s.users.GetByCode(ctx, rt.UserCode)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			_, _ = s.refresh.DeleteByTokenHash(ctx, hash)
			return "", "", services.ErrInvalidRefresh
		}
		return "", "", fmt.Errorf("refresh user lookup: %w", err)
	}
	newAccess, err := s.accessTokenFor(user)
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
		ExpiresAt: time.Now().Add(s.refreshTTL),
	}
	if err := s.refresh.Create(ctx, &newRT); err != nil {
		return "", "", fmt.Errorf("store rotated token: %w", err)
	}
	return newAccess, newRefresh, nil
}

func (s *Service) Logout(ctx context.Context, refreshToken string) error {
	if strings.TrimSpace(refreshToken) == "" {
		return services.ErrInvalidRefresh
	}
	deleted, err := s.refresh.DeleteByTokenHash(ctx, utils.HashRefreshToken(strings.TrimSpace(refreshToken)))
	if err != nil {
		return fmt.Errorf("logout: %w", err)
	}
	if !deleted {
		return services.ErrInvalidRefresh
	}
	return nil
}

func (s *Service) LogoutAll(ctx context.Context, userCode string) error {
	return s.refresh.DeleteByUserCode(ctx, userCode)
}

func (s *Service) CountUsers(ctx context.Context) (int, error) {
	return s.users.Count(ctx)
}

// accessTokenFor mengemas pembuatan token akses (dipakai login & refresh flow).
func (s *Service) accessTokenFor(user *models.User) (string, error) {
	return utils.GenerateAccessTokenWithTTL(s.jwtSecret, user.Code, user.Username, user.RoleCode, s.accessTTL)
}
