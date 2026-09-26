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
	// Registry menu (CPMENU + CPMATRIX view).
	ListMenus(ctx context.Context) ([]models.Menu, error)
	GetMenu(ctx context.Context, code string) (*models.Menu, error)
	CreateMenu(ctx context.Context, input models.MenuInput) (*models.Menu, error)
	DeleteMenu(ctx context.Context, code string) error
	GetMatrix(ctx context.Context, roleFilter string) ([]models.MatrixRow, error)
	MyMenus(ctx context.Context, userCode string) ([]models.MenuEntry, error)
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
			return nil, services.ErrRoleNotFound
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
		return fmt.Errorf("%w: role code must be 1-20 characters", services.ErrInvalidRole)
	}
	for _, r := range code {
		if !(r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '_') {
			return fmt.Errorf("%w: role code must be uppercase letters, digits, or underscore", services.ErrInvalidRole)
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
		return nil, fmt.Errorf("%w: role name is required", services.ErrInvalidRole)
	}
	role := &models.Role{Code: code, Name: name}
	if err := s.roles.Create(ctx, role); err != nil {
		msg := strings.ToLower(err.Error())
		if strings.Contains(msg, "2627") || strings.Contains(msg, "2601") ||
			strings.Contains(msg, "unique constraint failed") ||
			strings.Contains(msg, "duplicate key") {
			return nil, services.ErrRoleExists
		}
		return nil, fmt.Errorf("create role: %w", err)
	}
	return role, nil
}

func (s *Service) DeleteRole(ctx context.Context, code string) error {
	code = strings.TrimSpace(strings.ToUpper(code))
	if code == models.RoleAdmin || code == models.RoleUser {
		return services.ErrRoleProtected
	}
	if _, err := s.roles.GetByCode(ctx, code); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return services.ErrRoleNotFound
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
		return services.ErrRoleInUse
	}
	if err := s.roles.Delete(ctx, code); err != nil {
		// FK dari CPUSER sebagai backstop (mssql 547, pg foreign key, sqlite FK).
		msg := strings.ToLower(err.Error())
		if strings.Contains(msg, "fk_") || strings.Contains(msg, "547") ||
			strings.Contains(msg, "foreign key") {
			return services.ErrRoleInUse
		}
		return fmt.Errorf("delete role: %w", err)
	}
	return nil
}

func (s *Service) SetRolePermissions(ctx context.Context, roleCode string, permCodes []string) error {
	roleCode = strings.TrimSpace(strings.ToUpper(roleCode))
	if _, err := s.roles.GetByCode(ctx, roleCode); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return services.ErrRoleNotFound
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
			return fmt.Errorf("%w: unknown menu permission: %s", services.ErrInvalidRole, pc)
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
			return services.ErrRoleNotFound
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

// --- Registry menu (CPMENU) ---------------------------------------------------

// validMenuCode memastikan kode permission menu: MENU_ + huruf/digit/underscore.
func validMenuCode(code string) error {
	code = strings.TrimSpace(strings.ToUpper(code))
	if !strings.HasPrefix(code, "MENU_") || len(code) > 40 || len(code) <= 5 {
		return fmt.Errorf("%w: menu code must look like MENU_<NAME> (max 40 chars)", services.ErrInvalidMenu)
	}
	for _, r := range code[5:] {
		if !(r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '_') {
			return fmt.Errorf("%w: menu code must be uppercase letters, digits, or underscore", services.ErrInvalidMenu)
		}
	}
	return nil
}

// validModule memastikan MCONTROL = nama folder modul (UPPERCASE persis).
func validModule(m string) error {
	m = strings.TrimSpace(strings.ToUpper(m))
	if m == "" || len(m) > 40 {
		return fmt.Errorf("%w: module must be 1-40 characters", services.ErrInvalidMenu)
	}
	for _, r := range m {
		if !(r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '_') {
			return fmt.Errorf("%w: module must be uppercase letters, digits, or underscore", services.ErrInvalidMenu)
		}
	}
	return nil
}

func isDuplicateErr(err error) bool {
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "2627") || strings.Contains(msg, "2601") ||
		strings.Contains(msg, "unique constraint failed") ||
		strings.Contains(msg, "duplicate key")
}

func (s *Service) ListMenus(ctx context.Context) ([]models.Menu, error) {
	menus, err := s.roles.ListMenus(ctx)
	if err != nil {
		return nil, fmt.Errorf("list menus: %w", err)
	}
	return menus, nil
}

func (s *Service) GetMenu(ctx context.Context, code string) (*models.Menu, error) {
	code = strings.TrimSpace(strings.ToUpper(code))
	m, err := s.roles.GetMenu(ctx, code)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, services.ErrMenuNotFound
		}
		return nil, fmt.Errorf("get menu: %w", err)
	}
	return m, nil
}

// CreateMenu menulis permission + registry dalam 1 transaksi.
// Tanpa auto-grant ke role mana pun (admin mencentang manual via matriks).
func (s *Service) CreateMenu(ctx context.Context, input models.MenuInput) (*models.Menu, error) {
	code := strings.TrimSpace(strings.ToUpper(input.Code))
	name := strings.TrimSpace(input.Name)
	module := strings.TrimSpace(strings.ToUpper(input.Module))
	label := strings.TrimSpace(input.Label)
	if err := validMenuCode(code); err != nil {
		return nil, err
	}
	if err := validModule(module); err != nil {
		return nil, err
	}
	if name == "" || label == "" {
		return nil, fmt.Errorf("%w: menu name and label are required", services.ErrInvalidMenu)
	}
	if len(name) > 100 || len(label) > 100 {
		return nil, fmt.Errorf("%w: name/label must be at most 100 characters", services.ErrInvalidMenu)
	}
	sortOrder := input.SortOrder
	if sortOrder < 0 || sortOrder > 9999 {
		return nil, fmt.Errorf("%w: sort order must be 0-9999", services.ErrInvalidMenu)
	}
	parent := strings.TrimSpace(strings.ToUpper(input.Parent))
	if parent != "" {
		pm, err := s.roles.GetMenu(ctx, parent)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, fmt.Errorf("%w: parent menu not found: %s", services.ErrInvalidMenu, parent)
			}
			return nil, fmt.Errorf("get parent menu: %w", err)
		}
		if pm.MControl != module {
			return nil, fmt.Errorf("%w: parent must be in the same module", services.ErrInvalidMenu)
		}
	}
	menu := &models.Menu{Code: code, MControl: module, Label: label, SortOrder: sortOrder, Parent: parent}
	desc := "Seluruh fungsi " + label
	if err := s.roles.CreateMenuFull(ctx, code, name, module, desc, menu); err != nil {
		if isDuplicateErr(err) {
			return nil, services.ErrPermissionExists
		}
		return nil, fmt.Errorf("create menu: %w", err)
	}
	return menu, nil
}

// DeleteMenu menghapus menu + permission (mapping role ikut CASCADE).
// Ditolak bila masih dipakai role atau masih punya anak.
func (s *Service) DeleteMenu(ctx context.Context, code string) error {
	code = strings.TrimSpace(strings.ToUpper(code))
	if _, err := s.GetMenu(ctx, code); err != nil {
		return err
	}
	children, err := s.roles.ListChildren(ctx, code)
	if err != nil {
		return fmt.Errorf("list child menus: %w", err)
	}
	if len(children) > 0 {
		return services.ErrMenuHasChildren
	}
	used, err := s.roles.CountMenuUsage(ctx, code)
	if err != nil {
		return fmt.Errorf("check menu usage: %w", err)
	}
	if used > 0 {
		return services.ErrMenuInUse
	}
	if err := s.roles.DeleteMenu(ctx, code); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return services.ErrMenuNotFound
		}
		// FK anak sebagai backstop lintas engine.
		msg := strings.ToLower(err.Error())
		if strings.Contains(msg, "fk_") || strings.Contains(msg, "547") ||
			strings.Contains(msg, "foreign key") {
			return services.ErrMenuHasChildren
		}
		return fmt.Errorf("delete menu: %w", err)
	}
	return nil
}

// GetMatrix membaca view CPMATRIX (roleFilter kosong = semua role).
func (s *Service) GetMatrix(ctx context.Context, roleFilter string) ([]models.MatrixRow, error) {
	roleFilter = strings.TrimSpace(strings.ToUpper(roleFilter))
	if roleFilter != "" {
		if _, err := s.roles.GetByCode(ctx, roleFilter); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, services.ErrRoleNotFound
			}
			return nil, fmt.Errorf("get role: %w", err)
		}
	}
	rows, err := s.roles.QueryMatrix(ctx, roleFilter)
	if err != nil {
		return nil, fmt.Errorf("query matrix: %w", err)
	}
	return rows, nil
}

// MyMenus mengembalikan menu milik user (untuk sidebar + halaman kosong).
func (s *Service) MyMenus(ctx context.Context, userCode string) ([]models.MenuEntry, error) {
	user, err := s.users.GetByCode(ctx, userCode)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, services.ErrUserNotFound
		}
		return nil, fmt.Errorf("menu user lookup: %w", err)
	}
	entries, err := s.roles.MyMenusByRole(ctx, user.RoleCode)
	if err != nil {
		return nil, fmt.Errorf("list my menus: %w", err)
	}
	return entries, nil
}
