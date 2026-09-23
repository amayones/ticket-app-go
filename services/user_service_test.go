package services

import (
	"database/sql"
	"errors"
	"os"
	"testing"
)

var sqlErrNoRows = sql.ErrNoRows
var errUsernameConstraint = errors.New(`mssql: Violation of UNIQUE KEY constraint 'UQ_users_username'`)
var errEmailConstraint = errors.New(`mssql: Violation of UNIQUE KEY constraint 'UQ_users_email'`)

func init() {
	_ = os.Setenv("JWT_SECRET", "test-secret-key-for-unit-tests-32-chars")
}

func TestCreateUser_Success(t *testing.T) {
	repo := NewMockUserRepository()
	refreshRepo := NewMockRefreshTokenRepository()
	service := NewUserService(repo, refreshRepo)
	err := service.CreateUser("budi", "budi@example.com", "password123")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(repo.Users) != 1 {
		t.Fatalf("expected 1 user in repo, got %d", len(repo.Users))
	}
}

func TestCreateUser_InvalidEmail(t *testing.T) {
	repo := NewMockUserRepository()
	refreshRepo := NewMockRefreshTokenRepository()
	service := NewUserService(repo, refreshRepo)
	err := service.CreateUser("budi", "bukan-email", "password123")
	if !errors.Is(err, ErrInvalidEmail) {
		t.Fatalf("expected ErrInvalidEmail, got %v", err)
	}
}

func TestCreateUser_PasswordTooShort(t *testing.T) {
	repo := NewMockUserRepository()
	refreshRepo := NewMockRefreshTokenRepository()
	service := NewUserService(repo, refreshRepo)
	err := service.CreateUser("budi", "budi@example.com", "123")
	if !errors.Is(err, ErrPasswordTooShort) {
		t.Fatalf("expected ErrPasswordTooShort, got %v", err)
	}
}

func TestCreateUser_UsernameTooShort(t *testing.T) {
	repo := NewMockUserRepository()
	refreshRepo := NewMockRefreshTokenRepository()
	service := NewUserService(repo, refreshRepo)
	err := service.CreateUser("ab", "budi@example.com", "password123")
	if !errors.Is(err, ErrUsernameTooShort) {
		t.Fatalf("expected ErrUsernameTooShort, got %v", err)
	}
}

func TestCreateUser_DuplicateUsername(t *testing.T) {
	repo := NewMockUserRepository()
	refreshRepo := NewMockRefreshTokenRepository()
	service := NewUserService(repo, refreshRepo)
	_ = service.CreateUser("budi", "budi@example.com", "password123")
	err := service.CreateUser("budi", "lain@example.com", "password123")
	if !errors.Is(err, ErrUsernameTaken) {
		t.Fatalf("expected ErrUsernameTaken, got %v", err)
	}
}

func TestCreateUser_DuplicateEmail(t *testing.T) {
	repo := NewMockUserRepository()
	refreshRepo := NewMockRefreshTokenRepository()
	service := NewUserService(repo, refreshRepo)
	_ = service.CreateUser("budi", "budi@example.com", "password123")
	err := service.CreateUser("lain", "budi@example.com", "password123")
	if !errors.Is(err, ErrEmailTaken) {
		t.Fatalf("expected ErrEmailTaken, got %v", err)
	}
}

func TestLogin_Success(t *testing.T) {
	repo := NewMockUserRepository()
	refreshRepo := NewMockRefreshTokenRepository()
	service := NewUserService(repo, refreshRepo)
	_ = service.CreateUser("budi", "budi@example.com", "password123")
	accessToken, refreshToken, err := service.Login("budi", "password123")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if accessToken == "" {
		t.Fatal("expected non-empty access token")
	}
	if refreshToken == "" {
		t.Fatal("expected non-empty refresh token")
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	repo := NewMockUserRepository()
	refreshRepo := NewMockRefreshTokenRepository()
	service := NewUserService(repo, refreshRepo)
	_ = service.CreateUser("budi", "budi@example.com", "password123")
	_, _, err := service.Login("budi", "passwordsalah")
	if !errors.Is(err, ErrInvalidLogin) {
		t.Fatalf("expected ErrInvalidLogin, got %v", err)
	}
}

func TestLogin_UserNotFound(t *testing.T) {
	repo := NewMockUserRepository()
	refreshRepo := NewMockRefreshTokenRepository()
	service := NewUserService(repo, refreshRepo)
	_, _, err := service.Login("tidakada", "password123")
	if !errors.Is(err, ErrInvalidLogin) {
		t.Fatalf("expected ErrInvalidLogin, got %v", err)
	}
}

func TestGetUserByID_NotFound(t *testing.T) {
	repo := NewMockUserRepository()
	refreshRepo := NewMockRefreshTokenRepository()
	service := NewUserService(repo, refreshRepo)
	_, err := service.GetUserByID(999)
	if !errors.Is(err, ErrUserNotFound) {
		t.Fatalf("expected ErrUserNotFound, got %v", err)
	}
}

func TestDeleteUser_Success(t *testing.T) {
	repo := NewMockUserRepository()
	refreshRepo := NewMockRefreshTokenRepository()
	service := NewUserService(repo, refreshRepo)
	_ = service.CreateUser("budi", "budi@example.com", "password123")
	err := service.DeleteUser(1)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(repo.Users) != 0 {
		t.Fatalf("expected 0 users after delete, got %d", len(repo.Users))
	}
}

func TestRefreshAccessToken_Success(t *testing.T) {
	repo := NewMockUserRepository()
	refreshRepo := NewMockRefreshTokenRepository()
	service := NewUserService(repo, refreshRepo)
	_ = service.CreateUser("budi", "budi@example.com", "password123")
	_, refreshToken, _ := service.Login("budi", "password123")
	newAccessToken, newRefreshToken, err := service.RefreshAccessToken(refreshToken)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if newAccessToken == "" {
		t.Fatal("expected non-empty access token")
	}
	if newRefreshToken == "" {
		t.Fatal("expected non-empty rotated refresh token")
	}
	if _, _, err := service.RefreshAccessToken(refreshToken); !errors.Is(err, ErrInvalidRefresh) {
		t.Fatalf("expected old refresh token invalidated after rotation, got %v", err)
	}
}

func TestRefreshAccessToken_InvalidToken(t *testing.T) {
	repo := NewMockUserRepository()
	refreshRepo := NewMockRefreshTokenRepository()
	service := NewUserService(repo, refreshRepo)
	_, _, err := service.RefreshAccessToken("token-yang-tidak-ada")
	if !errors.Is(err, ErrInvalidRefresh) {
		t.Fatalf("expected ErrInvalidRefresh, got %v", err)
	}
}

func TestLogout_Success(t *testing.T) {
	repo := NewMockUserRepository()
	refreshRepo := NewMockRefreshTokenRepository()
	service := NewUserService(repo, refreshRepo)
	_ = service.CreateUser("budi", "budi@example.com", "password123")
	_, refreshToken, _ := service.Login("budi", "password123")
	err := service.Logout(refreshToken)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	_, _, err = service.RefreshAccessToken(refreshToken)
	if !errors.Is(err, ErrInvalidRefresh) {
		t.Fatalf("expected token to be invalid after logout, got %v", err)
	}
}
