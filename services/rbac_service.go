package services

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"golang-backend/models"
	"golang-backend/utils"
)

// RBAC: role, permission, dan penugasan role ke user.
// Method menempel pada *UserService agar wiring tetap satu service inti.

// CheckPermission memastikan user berhak atas permission (ADMIN selalu lolos).
func (s *UserService) CheckPermission(ctx context.Context, userCode, permCode string) error {
	user, err := s.users.GetByCode(ctx, userCode)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrUserNotFound
		}
		return fmt.Errorf("permission user lookup: %w", err)
	}
	if user.RoleCode == models.RoleAdmin {
		return nil
	}
	ok, err := s.roles.HasPermission(ctx, user.RoleCode, permCode)
	if err != nil {
		return fmt.Errorf("permission check: %w", err)
	}
	if !ok {
		return ErrForbidden
	}
	return nil
}

func (s *UserService) ListPermissions(ctx context.Context) ([]models.Permission, error) {
	perms, err := s.roles.ListPermissions(ctx)
	if err != nil {
		return nil, fmt.Errorf("list permissions: %w", err)
	}
	return perms, nil
}

// GetRoleDetail mengembalikan role + kode permission miliknya (matriks RBAC).
func (s *UserService) GetRoleDetail(ctx context.Context, code string) (*models.RoleDetail, error) {
	role, err := s.roles.GetByCode(ctx, code)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("get role: %w", err)
	}
	perms, err := s.roles.GetRolePermissions(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("get role permissions: %w", err)
	}
	return &models.RoleDetail{Role: *role, Permissions: perms}, nil
}

func validRoleCode(code string) error {
	code = strings.TrimSpace(strings.ToUpper(code))
	if code == "" || len(code) > 20 {
		return errors.New("role code must be 1-20 characters")
	}
	for _, r := range code {
		if !(r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '_') {
			return errors.New("role code must be uppercase letters, digits, or underscore")
		}
	}
	return nil
}

func (s *UserService) CreateRole(ctx context.Context, code, name string) (*models.Role, error) {
	code = strings.TrimSpace(strings.ToUpper(code))
	name = strings.TrimSpace(name)
	if err := validRoleCode(code); err != nil {
		return nil, err
	}
	if name == "" {
		return nil, errors.New("role name is required")
	}
	role := &models.Role{Code: code, Name: name}
	if err := s.roles.Create(ctx, role); err != nil {
		msg := err.Error()
		if strings.Contains(msg, "2627") || strings.Contains(msg, "2601") {
			return nil, errors.New("role code already exists")
		}
		return nil, fmt.Errorf("create role: %w", err)
	}
	return role, nil
}

func (s *UserService) DeleteRole(ctx context.Context, code string) error {
	code = strings.TrimSpace(strings.ToUpper(code))
	if code == models.RoleAdmin || code == models.RoleUser {
		return errors.New("system roles cannot be deleted")
	}
	if _, err := s.roles.GetByCode(ctx, code); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrUserNotFound
		}
		return fmt.Errorf("get role: %w", err)
	}
	if err := s.roles.Delete(ctx, code); err != nil {
		// FK dari CPUSER: tolak bila masih dipakai user.
		if strings.Contains(err.Error(), "FK_") || strings.Contains(err.Error(), "547") {
			return errors.New("role is still assigned to users")
		}
		return fmt.Errorf("delete role: %w", err)
	}
	return nil
}

func (s *UserService) SetRolePermissions(ctx context.Context, roleCode string, permCodes []string) error {
	roleCode = strings.TrimSpace(strings.ToUpper(roleCode))
	if _, err := s.roles.GetByCode(ctx, roleCode); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrUserNotFound
		}
		return fmt.Errorf("get role: %w", err)
	}
	// Validasi semua kode permission ada.
	all, err := s.roles.ListPermissions(ctx)
	if err != nil {
		return fmt.Errorf("list permissions: %w", err)
	}
	known := make(map[string]bool, len(all))
	for _, p := range all {
		known[p.Code] = true
	}
	for _, pc := range permCodes {
		if !known[pc] {
			return fmt.Errorf("unknown permission: %s", pc)
		}
	}
	if err := s.roles.SetRolePermissions(ctx, roleCode, permCodes); err != nil {
		return fmt.Errorf("set role permissions: %w", err)
	}
	return nil
}

// UpdateUserRole mengganti role akun (butuh USER_ROLE_ASSIGN di handler).
func (s *UserService) UpdateUserRole(ctx context.Context, userCode, roleCode string) error {
	roleCode = strings.TrimSpace(strings.ToUpper(roleCode))
	if _, err := s.roles.GetByCode(ctx, roleCode); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("role not found")
		}
		return fmt.Errorf("get role: %w", err)
	}
	if err := s.users.UpdateRole(ctx, userCode, roleCode); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrUserNotFound
		}
		return fmt.Errorf("update user role: %w", err)
	}
	return nil
}

// GenerateAccessTokenFor mengemas ulang pembuatan token (dipakai refresh flow).
func (s *UserService) accessTokenFor(user *models.User) (string, error) {
	return utils.GenerateAccessToken(s.jwtSecret, user.Code, user.Username, user.RoleCode)
}
