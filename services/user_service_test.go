package services

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"
	"time"

	"golang-backend/models"
)

var sqlErrNoRows = sql.ErrNoRows
var errUsernameConstraint = errors.New(`mssql: Violation of UNIQUE KEY constraint 'UQ_CPUSER_USERNAME'`)
var errEmailConstraint = errors.New(`mssql: Violation of UNIQUE KEY constraint 'UQ_CPUSER_EMAIL'`)
var errCodeConstraint = errors.New(`mssql: Violation of UNIQUE KEY constraint 'UQ_CPUSER_CODE'`)

const testJWTSecret = "test-secret-key-for-unit-tests-32-chars"

func newTestService(t *testing.T) (*UserService, *MockUserRepository, *MockRefreshTokenRepository) {
	t.Helper()
	t.Setenv("JWT_SECRET", testJWTSecret)
	repo := NewMockUserRepository()
	refreshRepo := NewMockRefreshTokenRepository()
	roleRepo := NewMockRoleRepository()
	svc, err := NewUserService(repo, refreshRepo, roleRepo, testJWTSecret)
	if err != nil {
		t.Fatalf("NewUserService: %v", err)
	}
	return svc, repo, refreshRepo
}

func strptr(s string) *string { return &s }

func TestCreateUser_Success(t *testing.T) {
	svc, repo, _ := newTestService(t)
	code, err := svc.CreateUser(context.Background(), "budi", "budi@example.com", "password123", "")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if code == "" {
		t.Fatal("expected non-empty code")
	}
	if !strings.HasPrefix(code, "USR-") {
		t.Fatalf("expected USR- prefixed code, got %q", code)
	}
	if len(repo.Users) != 1 {
		t.Fatalf("expected 1 user in repo, got %d", len(repo.Users))
	}
	if repo.Users[code].RoleCode != models.DefaultRoleCode {
		t.Fatalf("expected default role %q, got %q", models.DefaultRoleCode, repo.Users[code].RoleCode)
	}
}

func TestCreateUser_Table(t *testing.T) {
	cases := []struct {
		name     string
		username string
		email    string
		password string
		wantErr  error
	}{
		{"invalid email", "budi", "bukan-email", "password123", ErrInvalidEmail},
		{"short password", "budi", "budi@example.com", "123", ErrPasswordTooShort},
		{"short username", "ab", "budi@example.com", "password123", ErrUsernameTooShort},
		{"blank spaces", "   ", "budi@example.com", "password123", ErrInputRequired},
		{"too long password", "budi", "budi@example.com", strings.Repeat("x", 73), ErrPasswordTooLong},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			svc, _, _ := newTestService(t)
			_, err := svc.CreateUser(context.Background(), tc.username, tc.email, tc.password, "")
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("expected %v, got %v", tc.wantErr, err)
			}
		})
	}
}

func TestCreateUser_NormalizesEmail(t *testing.T) {
	svc, repo, _ := newTestService(t)
	code, err := svc.CreateUser(context.Background(), "budi", "  Budi@Example.COM ", "password123", "")
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	u := repo.Users[code]
	if u.Email != "budi@example.com" {
		t.Fatalf("expected normalized email, got %q", u.Email)
	}
}

func TestCreateUser_Duplicates(t *testing.T) {
	svc, _, _ := newTestService(t)
	if _, err := svc.CreateUser(context.Background(), "budi", "budi@example.com", "password123", ""); err != nil {
		t.Fatalf("seed: %v", err)
	}
	if _, err := svc.CreateUser(context.Background(), "budi", "lain@example.com", "password123", ""); !errors.Is(err, ErrUsernameTaken) {
		t.Fatalf("expected ErrUsernameTaken, got %v", err)
	}
	if _, err := svc.CreateUser(context.Background(), "lain", "budi@example.com", "password123", ""); !errors.Is(err, ErrEmailTaken) {
		t.Fatalf("expected ErrEmailTaken, got %v", err)
	}
}

func TestCreateUser_WithRole(t *testing.T) {
	svc, repo, _ := newTestService(t)
	ctx := context.Background()
	code, err := svc.CreateUser(ctx, "budi", "budi@example.com", "password123", "ADMIN")
	if err != nil {
		t.Fatalf("create with role: %v", err)
	}
	if repo.Users[code].RoleCode != models.RoleAdmin {
		t.Fatalf("expected role ADMIN, got %q", repo.Users[code].RoleCode)
	}
	if _, err := svc.CreateUser(ctx, "siti", "siti@example.com", "password123", "ROLE_TAK_ADA"); err == nil {
		t.Fatal("expected error for unknown role")
	}
}

func TestLogin(t *testing.T) {
	svc, _, _ := newTestService(t)
	ctx := context.Background()
	if _, err := svc.CreateUser(ctx, "budi", "budi@example.com", "password123", ""); err != nil {
		t.Fatalf("seed: %v", err)
	}
	access, refresh, err := svc.Login(ctx, "budi", "password123")
	if err != nil || access == "" || refresh == "" {
		t.Fatalf("expected tokens, got %v", err)
	}
	if _, _, err := svc.Login(ctx, "budi", "passwordsalah"); !errors.Is(err, ErrInvalidLogin) {
		t.Fatalf("expected ErrInvalidLogin, got %v", err)
	}
	if _, _, err := svc.Login(ctx, "tidakada", "password123"); !errors.Is(err, ErrInvalidLogin) {
		t.Fatalf("expected ErrInvalidLogin, got %v", err)
	}
}

func TestLogin_EvictsBeyondCap(t *testing.T) {
	svc, _, refreshRepo := newTestService(t)
	ctx := context.Background()
	code, err := svc.CreateUser(ctx, "budi", "budi@example.com", "password123", "")
	if err != nil {
		t.Fatalf("seed: %v", err)
	}
	for i := 0; i < MaxRefreshTokensPerUser+3; i++ {
		if _, _, err := svc.Login(ctx, "budi", "password123"); err != nil {
			t.Fatalf("login %d: %v", i, err)
		}
	}
	n, _ := refreshRepo.CountByUserCode(ctx, code)
	if n > MaxRefreshTokensPerUser {
		t.Fatalf("expected at most %d tokens, got %d", MaxRefreshTokensPerUser, n)
	}
}

func TestGetUserByCode_NotFound(t *testing.T) {
	svc, _, _ := newTestService(t)
	if _, err := svc.GetUserByCode(context.Background(), "USR-TIDAKADA"); !errors.Is(err, ErrUserNotFound) {
		t.Fatalf("expected ErrUserNotFound, got %v", err)
	}
}

func TestListRoles(t *testing.T) {
	svc, _, _ := newTestService(t)
	roles, err := svc.ListRoles(context.Background())
	if err != nil {
		t.Fatalf("list roles: %v", err)
	}
	if len(roles) != 2 {
		t.Fatalf("expected 2 seeded roles, got %d", len(roles))
	}
}

func TestUpdateUser_Partial(t *testing.T) {
	svc, repo, _ := newTestService(t)
	ctx := context.Background()
	code, _ := svc.CreateUser(ctx, "budi", "budi@example.com", "password123", "")
	oldHash := repo.Users[code].Password
	if err := svc.UpdateUser(ctx, code, models.UpdateUserRequest{Username: strptr("budi2")}); err != nil {
		t.Fatalf("partial update: %v", err)
	}
	if repo.Users[code].Username != "budi2" {
		t.Fatalf("username not updated: %q", repo.Users[code].Username)
	}
	if repo.Users[code].Password != oldHash {
		t.Fatal("password hash must not change on profile-only update")
	}
	if err := svc.UpdateUser(ctx, "USR-TIDAKADA", models.UpdateUserRequest{Username: strptr("x")}); !errors.Is(err, ErrUserNotFound) {
		t.Fatalf("expected ErrUserNotFound, got %v", err)
	}
}

func TestDeleteUser_Success(t *testing.T) {
	svc, repo, _ := newTestService(t)
	ctx := context.Background()
	code, _ := svc.CreateUser(ctx, "budi", "budi@example.com", "password123", "")
	if err := svc.DeleteUser(ctx, code); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if len(repo.Users) != 0 {
		t.Fatalf("expected 0 users after delete, got %d", len(repo.Users))
	}
	if err := svc.DeleteUser(ctx, code); !errors.Is(err, ErrUserNotFound) {
		t.Fatalf("expected ErrUserNotFound, got %v", err)
	}
}

func TestRefresh_RotatesAndInvalidatesOld(t *testing.T) {
	svc, _, _ := newTestService(t)
	ctx := context.Background()
	if _, err := svc.CreateUser(ctx, "budi", "budi@example.com", "password123", ""); err != nil {
		t.Fatalf("seed: %v", err)
	}
	_, refresh, _ := svc.Login(ctx, "budi", "password123")
	newAccess, newRefresh, err := svc.RefreshAccessToken(ctx, refresh)
	if err != nil || newAccess == "" || newRefresh == "" {
		t.Fatalf("refresh: %v", err)
	}
	if _, _, err := svc.RefreshAccessToken(ctx, refresh); !errors.Is(err, ErrInvalidRefresh) {
		t.Fatalf("expected old token invalidated, got %v", err)
	}
	if _, _, err := svc.RefreshAccessToken(ctx, "token-yang-tidak-ada"); !errors.Is(err, ErrInvalidRefresh) {
		t.Fatalf("expected ErrInvalidRefresh, got %v", err)
	}
}

func TestRefresh_Expired(t *testing.T) {
	svc, _, refreshRepo := newTestService(t)
	ctx := context.Background()
	if _, err := svc.CreateUser(ctx, "budi", "budi@example.com", "password123", ""); err != nil {
		t.Fatalf("seed: %v", err)
	}
	_, refresh, _ := svc.Login(ctx, "budi", "password123")
	// Force-expire the stored hash.
	for _, rt := range refreshRepo.Tokens {
		rt.ExpiresAt = time.Now().Add(-time.Hour)
	}
	if _, _, err := svc.RefreshAccessToken(ctx, refresh); !errors.Is(err, ErrInvalidRefresh) {
		t.Fatalf("expected expired -> ErrInvalidRefresh, got %v", err)
	}
}

func TestLogout(t *testing.T) {
	svc, _, _ := newTestService(t)
	ctx := context.Background()
	if _, err := svc.CreateUser(ctx, "budi", "budi@example.com", "password123", ""); err != nil {
		t.Fatalf("seed: %v", err)
	}
	_, refresh, _ := svc.Login(ctx, "budi", "password123")
	if err := svc.Logout(ctx, refresh); err != nil {
		t.Fatalf("logout: %v", err)
	}
	if _, _, err := svc.RefreshAccessToken(ctx, refresh); !errors.Is(err, ErrInvalidRefresh) {
		t.Fatalf("expected invalid after logout, got %v", err)
	}
	if err := svc.Logout(ctx, "ngawur"); !errors.Is(err, ErrInvalidRefresh) {
		t.Fatalf("expected logout unknown -> ErrInvalidRefresh, got %v", err)
	}
}
