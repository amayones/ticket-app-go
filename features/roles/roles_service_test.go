package roles

import (
	"context"
	"database/sql"
	"testing"

	"golang-backend/models"
)

type fakeRoleRepo struct {
	roles map[string]*models.Role
	perms map[string][]string
}

func newFakeRoleRepo() *fakeRoleRepo {
	return &fakeRoleRepo{
		roles: map[string]*models.Role{
			models.RoleAdmin: {Code: models.RoleAdmin, Name: "Administrator"},
			models.RoleUser:  {Code: models.RoleUser, Name: "Pengguna"},
		},
		perms: map[string][]string{
			models.RoleAdmin: {models.MenuDashboard, models.MenuUsers, models.MenuRoles, models.MenuSessions, models.MenuAudit, models.MenuSecurity, models.MenuSyslog, models.MenuNotifications},
			models.RoleUser:  {models.MenuDashboard},
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
	return []models.Permission{{Code: models.MenuDashboard, Name: "Lihat dashboard", Group: "MENU"}}, nil
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
	if err := svc.CheckPermission(ctx, "USR-ADMIN", "NGAWUR"); err != nil {
		t.Fatalf("admin must pass all: %v", err)
	}
	if err := svc.CheckPermission(ctx, "USR-00001", models.MenuDashboard); err != nil {
		t.Fatalf("user read: %v", err)
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
	if err := svc.SetRolePermissions(ctx, "EDITOR", []string{models.MenuDashboard}); err != nil {
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
