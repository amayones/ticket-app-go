package roles

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"golang-backend/models"
	"golang-backend/services"
)

var errFakeDuplicate = errors.New("duplicate key value violates unique constraint")

type fakeRoleRepo struct {
	roles   map[string]*models.Role
	perms   map[string][]string
	menus   map[string]*models.Menu
	modules map[string]*models.Module
}

func newFakeRoleRepo() *fakeRoleRepo {
	return &fakeRoleRepo{
		roles: map[string]*models.Role{
			models.RoleAdmin: {Code: models.RoleAdmin, Name: "Administrator"},
			models.RoleUser:  {Code: models.RoleUser, Name: "Pengguna"},
		},
		perms: map[string][]string{
			models.RoleAdmin: {models.MenuUsers, models.MenuRoles, models.MenuSessions, models.MenuAudit, models.MenuSecurity, models.MenuSyslog, models.MenuNotifications},
			models.RoleUser:  {},
		},
		menus: map[string]*models.Menu{
			models.MenuUsers: {Code: models.MenuUsers, MControl: "SYSTEM", Label: "User Account", SortOrder: 1},
			models.MenuRoles: {Code: models.MenuRoles, MControl: "SYSTEM", Label: "Role & Permission", SortOrder: 2},
		},
		modules: map[string]*models.Module{
			"SYSTEM": {Code: "SYSTEM", Label: "System", SortOrder: 1},
		},
	}
}

func (f *fakeRoleRepo) List(ctx context.Context) ([]models.Role, error) {
	out := make([]models.Role, 0, len(f.roles))
	for _, r := range f.roles {
		out = append(out, *r)
	}
	return out, nil
}

func (f *fakeRoleRepo) GetByCode(ctx context.Context, code string) (*models.Role, error) {
	if r, ok := f.roles[code]; ok {
		cp := *r
		return &cp, nil
	}
	return nil, sql.ErrNoRows
}

func (f *fakeRoleRepo) RoleExists(ctx context.Context, code string) (bool, error) {
	_, ok := f.roles[code]
	return ok, nil
}

func (f *fakeRoleRepo) Create(ctx context.Context, role *models.Role) error {
	f.roles[role.Code] = role
	return nil
}

func (f *fakeRoleRepo) Delete(ctx context.Context, code string) error {
	delete(f.roles, code)
	return nil
}

func (f *fakeRoleRepo) Count(ctx context.Context) (int, error) { return len(f.roles), nil }

func (f *fakeRoleRepo) ListPermissions(ctx context.Context) ([]models.Permission, error) {
	out := make([]models.Permission, 0, len(f.menus))
	for _, m := range f.menus {
		out = append(out, models.Permission{Code: m.Code, Name: "Akses " + m.Label, Group: m.MControl})
	}
	return out, nil
}

func (f *fakeRoleRepo) GetRolePermissions(ctx context.Context, roleCode string) ([]string, error) {
	return f.perms[roleCode], nil
}

func (f *fakeRoleRepo) SetRolePermissions(ctx context.Context, roleCode string, permCodes []string) error {
	f.perms[roleCode] = permCodes
	return nil
}

func (f *fakeRoleRepo) HasPermission(ctx context.Context, roleCode, permCode string) (bool, error) {
	for _, p := range f.perms[roleCode] {
		if p == permCode {
			return true, nil
		}
	}
	return false, nil
}

func (f *fakeRoleRepo) ListMenus(ctx context.Context) ([]models.Menu, error) {
	out := make([]models.Menu, 0, len(f.menus))
	for _, m := range f.menus {
		out = append(out, *m)
	}
	return out, nil
}

func (f *fakeRoleRepo) GetMenu(ctx context.Context, code string) (*models.Menu, error) {
	if m, ok := f.menus[code]; ok {
		cp := *m
		return &cp, nil
	}
	return nil, sql.ErrNoRows
}

func (f *fakeRoleRepo) CreateMenu(ctx context.Context, menu *models.Menu) error {
	if _, ok := f.menus[menu.Code]; ok {
		return errFakeDuplicate
	}
	cp := *menu
	f.menus[menu.Code] = &cp
	return nil
}

func (f *fakeRoleRepo) DeleteMenu(ctx context.Context, code string) error {
	if _, ok := f.menus[code]; !ok {
		return sql.ErrNoRows
	}
	delete(f.menus, code)
	return nil
}

func (f *fakeRoleRepo) CountMenuUsage(ctx context.Context, code string) (int, error) {
	n := 0
	for _, perms := range f.perms {
		for _, p := range perms {
			if p == code {
				n++
			}
		}
	}
	return n, nil
}

func (f *fakeRoleRepo) ListChildren(ctx context.Context, code string) ([]models.Menu, error) {
	out := make([]models.Menu, 0)
	for _, m := range f.menus {
		if m.Parent == code {
			out = append(out, *m)
		}
	}
	return out, nil
}

func (f *fakeRoleRepo) QueryMatrix(ctx context.Context, roleFilter string) ([]models.MatrixRow, error) {
	out := make([]models.MatrixRow, 0)
	roles := []string{models.RoleAdmin, models.RoleUser}
	if roleFilter != "" {
		roles = []string{roleFilter}
	}
	for _, rc := range roles {
		for _, m := range f.menus {
			has := false
			for _, p := range f.perms[rc] {
				if p == m.Code {
					has = true
				}
			}
			out = append(out, models.MatrixRow{RoleCode: rc, Module: m.MControl, MenuCode: m.Code, MenuLabel: m.Label, HasAccess: has})
		}
	}
	return out, nil
}

func (f *fakeRoleRepo) MyMenusByRole(ctx context.Context, roleCode string) ([]models.MenuEntry, error) {
	out := make([]models.MenuEntry, 0)
	for _, p := range f.perms[roleCode] {
		if m, ok := f.menus[p]; ok {
			out = append(out, models.MenuEntry{Code: m.Code, Module: m.MControl, Label: m.Label, SortOrder: m.SortOrder})
		}
	}
	return out, nil
}

func (f *fakeRoleRepo) ListModules(ctx context.Context) ([]models.Module, error) {
	out := make([]models.Module, 0, len(f.modules))
	for _, m := range f.modules {
		out = append(out, *m)
	}
	return out, nil
}

func (f *fakeRoleRepo) GetModule(ctx context.Context, code string) (*models.Module, error) {
	if m, ok := f.modules[code]; ok {
		cp := *m
		return &cp, nil
	}
	return nil, sql.ErrNoRows
}

func (f *fakeRoleRepo) CreateModule(ctx context.Context, m *models.Module) error {
	if _, ok := f.modules[m.Code]; ok {
		return errFakeDuplicate
	}
	cp := *m
	f.modules[m.Code] = &cp
	return nil
}

func (f *fakeRoleRepo) DeleteModule(ctx context.Context, code string) error {
	if _, ok := f.modules[code]; !ok {
		return sql.ErrNoRows
	}
	delete(f.modules, code)
	return nil
}

func (f *fakeRoleRepo) CountModuleMenus(ctx context.Context, code string) (int, error) {
	n := 0
	for _, m := range f.menus {
		if m.MControl == code {
			n++
		}
	}
	return n, nil
}

type fakeUserStore struct {
	users map[string]*models.User
}

func (f *fakeUserStore) GetByCode(ctx context.Context, code string) (*models.User, error) {
	if u, ok := f.users[code]; ok {
		cp := *u
		return &cp, nil
	}
	return nil, sql.ErrNoRows
}

func (f *fakeUserStore) UpdateRole(ctx context.Context, code, roleCode string) error {
	f.users[code].RoleCode = roleCode
	return nil
}

func (f *fakeUserStore) CountByRole(ctx context.Context, roleCode string) (int, error) {
	n := 0
	for _, u := range f.users {
		if u.RoleCode == roleCode {
			n++
		}
	}
	return n, nil
}

func newTestService() (*Service, *fakeRoleRepo, *fakeUserStore) {
	repo := newFakeRoleRepo()
	store := &fakeUserStore{users: map[string]*models.User{
		"USR-ADMIN": {Code: "USR-ADMIN", Username: "admin", RoleCode: models.RoleAdmin},
		"USR-00001": {Code: "USR-00001", Username: "budi", RoleCode: models.RoleUser},
	}}
	return NewService(repo, store), repo, store
}

func TestCheckPermission(t *testing.T) {
	svc, _, _ := newTestService()
	ctx := context.Background()
	if err := svc.CheckPermission(ctx, "USR-ADMIN", models.MenuAudit); err != nil {
		t.Fatalf("admin menu access: %v", err)
	}
	// USER nol menu: semua akses harus ditolak.
	if err := svc.CheckPermission(ctx, "USR-00001", models.MenuUsers); err == nil {
		t.Fatal("expected forbidden for user without menus")
	}
	if err := svc.CheckPermission(ctx, "USR-00001", models.MenuAudit); err == nil {
		t.Fatal("expected forbidden for missing permission")
	}
}

func TestRoleLifecycle(t *testing.T) {
	svc, _, store := newTestService()
	ctx := context.Background()
	if _, err := svc.CreateRole(ctx, "editor", "Editor"); err != nil {
		t.Fatalf("create: %v", err)
	}
	if err := svc.SetRolePermissions(ctx, "EDITOR", []string{models.MenuUsers}); err != nil {
		t.Fatalf("set perms: %v", err)
	}
	if err := svc.UpdateUserRole(ctx, "USR-00001", "EDITOR"); err != nil {
		t.Fatalf("assign: %v", err)
	}
	if store.users["USR-00001"].RoleCode != "EDITOR" {
		t.Fatal("role not assigned")
	}
	if err := svc.DeleteRole(ctx, "EDITOR"); err == nil {
		t.Fatal("expected delete blocked while in use")
	}
	if err := svc.UpdateUserRole(ctx, "USR-00001", models.RoleUser); err != nil {
		t.Fatalf("unassign: %v", err)
	}
	if err := svc.DeleteRole(ctx, "EDITOR"); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if err := svc.DeleteRole(ctx, models.RoleAdmin); err == nil {
		t.Fatal("expected system role protected")
	}
}

func TestListRoles(t *testing.T) {
	svc, _, _ := newTestService()
	roles, err := svc.ListRoles(context.Background())
	if err != nil {
		t.Fatalf("list roles: %v", err)
	}
	if len(roles) != 2 {
		t.Fatalf("expected 2 roles, got %d", len(roles))
	}
}

func TestMenuLifecycle(t *testing.T) {
	svc, _, _ := newTestService()
	ctx := context.Background()
	// Modul wajib ada dulu (urut: modul -> menu).
	badMod := models.MenuInput{Code: "MENU_LAPORAN", Name: "Akses menu Laporan", Module: "report", Label: "Laporan"}
	if _, err := svc.CreateMenu(ctx, badMod); !errors.Is(err, services.ErrInvalidMenu) {
		t.Fatalf("expected ErrInvalidMenu for unknown module, got %v", err)
	}
	if _, err := svc.CreateModule(ctx, "report", "Report", 10); err != nil {
		t.Fatalf("create module: %v", err)
	}
	in := models.MenuInput{Code: "menu_laporan", Name: "Akses menu Laporan", Module: "report", Label: "Laporan", SortOrder: 8}
	menu, err := svc.CreateMenu(ctx, in)
	if err != nil {
		t.Fatalf("create menu: %v", err)
	}
	if menu.Code != "MENU_LAPORAN" || menu.MControl != "REPORT" {
		t.Fatalf("unexpected normalization: %+v", menu)
	}
	// Tanpa auto-grant: USER tetap nol menu.
	mine, err := svc.MyMenus(ctx, "USR-00001")
	if err != nil {
		t.Fatalf("my menus: %v", err)
	}
	if len(mine) != 0 {
		t.Fatalf("expected no auto-grant, got %d", len(mine))
	}
	// Duplikat → 409 typed.
	if _, err := svc.CreateMenu(ctx, in); !errors.Is(err, services.ErrPermissionExists) {
		t.Fatalf("expected ErrPermissionExists, got %v", err)
	}
	// Kode tanpa prefix → 400 typed.
	bad := in
	bad.Code = "LAPORAN"
	if _, err := svc.CreateMenu(ctx, bad); !errors.Is(err, services.ErrInvalidMenu) {
		t.Fatalf("expected ErrInvalidMenu, got %v", err)
	}
	// Dipakai role → hapus ditolak.
	if err := svc.SetRolePermissions(ctx, models.RoleAdmin, append(
		[]string{models.MenuUsers}, "MENU_LAPORAN")); err != nil {
		t.Fatalf("assign: %v", err)
	}
	if err := svc.DeleteMenu(ctx, "MENU_LAPORAN"); !errors.Is(err, services.ErrMenuInUse) {
		t.Fatalf("expected ErrMenuInUse, got %v", err)
	}
	// Lepas lalu hapus sukses.
	if err := svc.SetRolePermissions(ctx, models.RoleAdmin, []string{models.MenuUsers}); err != nil {
		t.Fatalf("unassign: %v", err)
	}
	if err := svc.DeleteMenu(ctx, "MENU_LAPORAN"); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if _, err := svc.GetMenu(ctx, "MENU_LAPORAN"); !errors.Is(err, services.ErrMenuNotFound) {
		t.Fatalf("expected ErrMenuNotFound, got %v", err)
	}
}

func TestMatrix(t *testing.T) {
	svc, _, _ := newTestService()
	rows, err := svc.GetMatrix(context.Background(), "")
	if err != nil {
		t.Fatalf("matrix: %v", err)
	}
	if len(rows) == 0 {
		t.Fatal("expected matrix rows")
	}
	for _, r := range rows {
		if r.Module == "" || r.MenuCode == "" || r.RoleCode == "" {
			t.Fatalf("incomplete matrix row: %+v", r)
		}
	}
}

func TestModuleLifecycle(t *testing.T) {
	svc, _, _ := newTestService()
	ctx := context.Background()
	m, err := svc.CreateModule(ctx, "report", "Report", 10)
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if m.Code != "REPORT" {
		t.Fatalf("expected uppercase, got %q", m.Code)
	}
	if _, err := svc.CreateModule(ctx, "REPORT", "Duplikat", 11); !errors.Is(err, services.ErrModuleExists) {
		t.Fatalf("expected ErrModuleExists, got %v", err)
	}
	if _, err := svc.CreateModule(ctx, "nope!", "Bad", 1); !errors.Is(err, services.ErrInvalidModule) {
		t.Fatalf("expected ErrInvalidModule, got %v", err)
	}
	// Modul berisi menu tidak boleh dihapus.
	if _, err := svc.CreateMenu(ctx, models.MenuInput{Code: "MENU_X", Name: "X", Module: "REPORT", Label: "X"}); err != nil {
		t.Fatalf("create menu: %v", err)
	}
	if err := svc.DeleteModule(ctx, "REPORT"); !errors.Is(err, services.ErrModuleInUse) {
		t.Fatalf("expected ErrModuleInUse, got %v", err)
	}
	if err := svc.DeleteMenu(ctx, "MENU_X"); err != nil {
		t.Fatalf("delete menu: %v", err)
	}
	if err := svc.DeleteModule(ctx, "REPORT"); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if err := svc.DeleteModule(ctx, "REPORT"); !errors.Is(err, services.ErrModuleNotFound) {
		t.Fatalf("expected ErrModuleNotFound, got %v", err)
	}
}
