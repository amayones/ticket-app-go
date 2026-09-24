package roles

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"golang-backend/models"
	"golang-backend/services"
)

// RBAC: role, permission, dan penugasan role ke user.
//
// Konvensi menu: query SQL selalu di variabel `query`, lalu di-run,
// lalu hasilnya dipetakan ke response (lihat roles_repository.go).
type ServiceInterface interface {
	CheckPermission(ctx context.Context, userCode, permCode string) error
	ListRoles(ctx context.Context) ([]models.Role, error)
	ListPermissions(ctx context.Context) ([]models.Permission, error)
	GetRoleDetail(ctx context.Context, code string) (*models.RoleDetail, error)
	GetRolePermissions(ctx context.Context, roleCode string) ([]string, error)
	CreateRole(ctx context.Context, code, name string) (*models.Role, error)
	DeleteRole(ctx context.Context, code string) error
	SetRolePermissions(ctx context.Context, roleCode string, permCodes []string) error
	UpdateUserRole(ctx context.Context, userCode, roleCode string) error
	CountRoles(ctx context.Context) (int, error)
}

// userStore dipenuhi users.Repository (tanpa import antar-fitur).
type userStore interface {
	GetByCode(ctx context.Context, code string) (*models.User, error)
	UpdateRole(ctx context.Context, code, roleCode string) error
	CountByRole(ctx context.Context, roleCode string) (int, error)
}

type Service struct {
	roles RepositoryInterface
	users userStore
}

func NewService(roles RepositoryInterface, users userStore) *Service {
	return &Service{roles: roles, users: users}
}

// CheckPermission memastikan user berhak atas permission (ADMIN selalu lolos).
func (s *Service) CheckPermission(ctx context.Context, userCode, permCode string) error {
	user, err := s.users.GetByCode(ctx, userCode)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return services.ErrUserNotFound
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
		return services.ErrForbidden
	}
	return nil
}

func (s *Service) ListRoles(ctx context.Context) ([]models.Role, error) {
	roles, err := s.roles.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("list roles: %w", err)
	}
	return roles, nil
}

func (s *Service) CountRoles(ctx context.Context) (int, error) {
	return s.roles.Count(ctx)
}

func (s *Service) ListPermissions(ctx context.Context) ([]models.Permission, error) {
	perms, err := s.roles.ListPermissions(ctx)
	if err != nil {
		return nil, fmt.Errorf("list permissions: %w", err)
	}
	return perms, nil
}

func (s *Service) GetRolePermissions(ctx context.Context, roleCode string) ([]string, error) {
	return s.roles.GetRolePermissions(ctx, roleCode)
}

// GetRoleDetail mengembalikan role + kode permission miliknya (matriks RBAC).
func (s *Service) GetRoleDetail(ctx context.Context, code string) (*models.RoleDetail, error) {
	role, err := s.roles.GetByCode(ctx, code)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, services.ErrUserNotFound
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

func (s *Service) CreateRole(ctx context.Context, code, name string) (*models.Role, error) {
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

func (s *Service) DeleteRole(ctx context.Context, code string) error {
	code = strings.TrimSpace(strings.ToUpper(code))
	if code == models.RoleAdmin || code == models.RoleUser {
		return errors.New("system roles cannot be deleted")
	}
	if _, err := s.roles.GetByCode(ctx, code); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return services.ErrUserNotFound
		}
		return fmt.Errorf("get role: %w", err)
	}
	// Portabel lintas engine: cek pemakaian eksplisit (beberapa engine
	// tidak menegakkan FK, mis. sqlite tanpa pragma foreign_keys).
	used, err := s.users.CountByRole(ctx, code)
	if err != nil {
		return fmt.Errorf("check role usage: %w", err)
	}
	if used > 0 {
		return errors.New("role is still assigned to users")
	}
	if err := s.roles.Delete(ctx, code); err != nil {
		// FK dari CPUSER sebagai backstop (mssql 547, pg foreign key, sqlite FK).
		msg := strings.ToLower(err.Error())
		if strings.Contains(msg, "fk_") || strings.Contains(msg, "547") ||
			strings.Contains(msg, "foreign key") {
			return errors.New("role is still assigned to users")
		}
		return fmt.Errorf("delete role: %w", err)
	}
	return nil
}

func (s *Service) SetRolePermissions(ctx context.Context, roleCode string, permCodes []string) error {
	roleCode = strings.TrimSpace(strings.ToUpper(roleCode))
	if _, err := s.roles.GetByCode(ctx, roleCode); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return services.ErrUserNotFound
		}
		return fmt.Errorf("get role: %w", err)
	}
	all, err := s.roles.ListPermissions(ctx)
	if err != nil {
		return fmt.Errorf("list permissions: %w", err)
	}
	known := make(map[string]bool, len(all))
	for _, p := range all {
		if strings.HasPrefix(p.Code, "MENU_") {
			known[p.Code] = true
		}
	}
	cleaned := make([]string, 0, len(permCodes))
	seen := make(map[string]bool, len(permCodes))
	for _, pc := range permCodes {
		pc = strings.TrimSpace(strings.ToUpper(pc))
		if !strings.HasPrefix(pc, "MENU_") || !known[pc] {
			return fmt.Errorf("unknown menu permission: %s", pc)
		}
		if seen[pc] {
			continue
		}
		seen[pc] = true
		cleaned = append(cleaned, pc)
	}
	if err := s.roles.SetRolePermissions(ctx, roleCode, cleaned); err != nil {
		return fmt.Errorf("set role permissions: %w", err)
	}
	return nil
}

// UpdateUserRole mengganti role akun (butuh MENU_USERS di handler).
func (s *Service) UpdateUserRole(ctx context.Context, userCode, roleCode string) error {
	roleCode = strings.TrimSpace(strings.ToUpper(roleCode))
	if _, err := s.roles.GetByCode(ctx, roleCode); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("role not found")
		}
		return fmt.Errorf("get role: %w", err)
	}
	if err := s.users.UpdateRole(ctx, userCode, roleCode); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return services.ErrUserNotFound
		}
		return fmt.Errorf("update user role: %w", err)
	}
	return nil
}
