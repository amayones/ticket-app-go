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
			models.MenuUsers: {Code: models.MenuUsers, Module: "SYSTEM", Label: "User Account", Mcontrol: "users", Kind: models.MenuKindChild, SortOrder: 1},
			models.MenuRoles: {Code: models.MenuRoles, Module: "SYSTEM", Label: "Role & Permission", Mcontrol: "role_permission", Kind: models.MenuKindChild, SortOrder: 2},
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

func (f *fakeRoleRepo) DeleteGrants(ctx context.Context, roleCode string) error {
	delete(f.perms, roleCode)
	return nil
}

func (f *fakeRoleRepo) Count(ctx context.Context) (int, error) { return len(f.roles), nil }

func (f *fakeRoleRepo) ListPermissions(ctx context.Context) ([]models.Permission, error) {
	out := make([]models.Permission, 0, len(f.menus))
	for _, m := range f.menus {
		out = append(out, models.Permission{
			Code: m.Code, Name: "Akses " + m.Label, Group: m.Module,
			Kind: m.Kind, Parent: m.Parent,
		})
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

func (f *fakeRoleRepo) UpdateMenu(ctx context.Context, menu *models.Menu) error {
	cur, ok := f.menus[menu.Code]
	if !ok {
		return sql.ErrNoRows
	}
	cur.Module = menu.Module
	cur.Label = menu.Label
	cur.Mcontrol = menu.Mcontrol
	cur.Kind = menu.Kind
	cur.SortOrder = menu.SortOrder
	cur.Parent = menu.Parent
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
			out = append(out, models.MatrixRow{
				RoleCode: rc, Module: m.Module, MenuCode: m.Code, MenuLabel: m.Label,
				Kind: m.Kind, Parent: m.Parent, HasAccess: has,
			})
		}
	}
	return out, nil
}

func (f *fakeRoleRepo) MyMenusByRole(ctx context.Context, roleCode string) ([]models.MenuEntry, error) {
	out := make([]models.MenuEntry, 0)
	for _, p := range f.perms[roleCode] {
		if m, ok := f.menus[p]; ok {
			out = append(out, models.MenuEntry{Code: m.Code, Module: m.Module, Label: m.Label, SortOrder: m.SortOrder})
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
		if m.Module == code {
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

func (f *fakeUserStore) DeleteByRole(ctx context.Context, roleCode string) (int, error) {
	n := 0
	for code, u := range f.users {
		if u.RoleCode == roleCode {
			delete(f.users, code)
			n++
		}
	}
	return n, nil
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
	svc, repo, store := newTestService()
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
	// Hapus role = cascade: user yang memakainya ikut terhapus.
	deleted, err := svc.DeleteRole(ctx, "EDITOR")
	if err != nil {
		t.Fatalf("delete role: %v", err)
	}
	if deleted != 1 {
		t.Fatalf("expected 1 user deleted with role, got %d", deleted)
	}
	if _, ok := store.users["USR-00001"]; ok {
		t.Fatal("user with deleted role must be removed")
	}
	if _, ok := repo.perms["EDITOR"]; ok {
		t.Fatal("grants of deleted role must be removed")
	}
	if _, err := svc.DeleteRole(ctx, "EDITOR"); !errors.Is(err, services.ErrRoleNotFound) {
		t.Fatalf("expected ErrRoleNotFound for missing role, got %v", err)
	}
	if _, err := svc.DeleteRole(ctx, models.RoleAdmin); !errors.Is(err, services.ErrRoleProtected) {
		t.Fatalf("expected system role protected, got %v", err)
	}
	if _, err := svc.DeleteRole(ctx, models.RoleUser); !errors.Is(err, services.ErrRoleProtected) {
		t.Fatalf("expected system role protected, got %v", err)
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
	badMod := models.MenuInput{Code: "MENU_LAPORAN", Name: "Akses menu Laporan", Module: "report", Label: "Laporan", Mcontrol: "laporan"}
	if _, err := svc.CreateMenu(ctx, badMod); !errors.Is(err, services.ErrInvalidMenu) {
		t.Fatalf("expected ErrInvalidMenu for unknown module, got %v", err)
	}
	if _, err := svc.CreateModule(ctx, "report", "Report", 10); err != nil {
		t.Fatalf("create module: %v", err)
	}
	in := models.MenuInput{Code: "menu_laporan", Name: "Akses menu Laporan", Module: "report", Label: "Laporan", Mcontrol: "laporan", SortOrder: 8}
	menu, err := svc.CreateMenu(ctx, in)
	if err != nil {
		t.Fatalf("create menu: %v", err)
	}
	if menu.Code != "MENU_LAPORAN" || menu.Module != "REPORT" {
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

func intPtr(v int) *int       { return &v }
func strPtr(v string) *string { return &v }

func TestUpdateMenu(t *testing.T) {
	svc, _, _ := newTestService()
	ctx := context.Background()
	if _, err := svc.CreateModule(ctx, "report", "Report", 10); err != nil {
		t.Fatalf("create module: %v", err)
	}
	if _, err := svc.CreateModule(ctx, "toko", "Toko", 11); err != nil {
		t.Fatalf("create module: %v", err)
	}
	// MENU_PAJU = menu parent (tanpa mcontrol/folder, hanya header), MENU_STOK = child biasa.
	if _, err := svc.CreateMenu(ctx, models.MenuInput{
		Code: "MENU_PAJU", Module: "report", Label: "Paju", Kind: models.MenuKindParent, SortOrder: 5,
	}); err != nil {
		t.Fatalf("create parent menu: %v", err)
	}
	if _, err := svc.CreateMenu(ctx, models.MenuInput{
		Code: "MENU_STOK", Module: "report", Label: "Stok", Mcontrol: "stok", Kind: models.MenuKindChild, SortOrder: 4,
		Parent: "MENU_PAJU",
	}); err != nil {
		t.Fatalf("create child menu: %v", err)
	}
	// Default tanpa Kind = CHILD.
	if cur, _ := svc.GetMenu(ctx, "MENU_STOK"); cur.Kind != models.MenuKindChild {
		t.Fatalf("jenis menu harus CHILD, got %q", cur.Kind)
	}
	// Menu PARENT tidak boleh punya parent.
	if _, err := svc.CreateMenu(ctx, models.MenuInput{
		Code: "MENU_X", Module: "report", Label: "X", Kind: models.MenuKindParent, Parent: "MENU_PAJU",
	}); !errors.Is(err, services.ErrInvalidMenu) {
		t.Fatalf("PARENT tidak boleh punya parent, got %v", err)
	}
	// PARENT wajib tanpa mcontrol.
	if _, err := svc.CreateMenu(ctx, models.MenuInput{
		Code: "MENU_XP", Module: "report", Label: "XP", Mcontrol: "xp", Kind: models.MenuKindParent,
	}); !errors.Is(err, services.ErrInvalidMenu) {
		t.Fatalf("PARENT tidak boleh punya mcontrol, got %v", err)
	}
	// Parent harus bertipe PARENT.
	if _, err := svc.CreateMenu(ctx, models.MenuInput{
		Code: "MENU_Y", Module: "report", Label: "Y", Mcontrol: "y", Parent: "MENU_STOK",
	}); !errors.Is(err, services.ErrInvalidMenu) {
		t.Fatalf("parent harus kind PARENT, got %v", err)
	}
	// CHILD wajib punya mcontrol.
	if _, err := svc.CreateMenu(ctx, models.MenuInput{
		Code: "MENU_Y2", Module: "report", Label: "Y2",
	}); !errors.Is(err, services.ErrInvalidMenu) {
		t.Fatalf("CHILD tanpa mcontrol harus ditolak, got %v", err)
	}
	// Kind tidak dikenal ditolak.
	if _, err := svc.CreateMenu(ctx, models.MenuInput{
		Code: "MENU_Z", Module: "report", Label: "Z", Mcontrol: "z", Kind: "INDUK",
	}); !errors.Is(err, services.ErrInvalidMenu) {
		t.Fatalf("kind tidak dikenal harus ditolak, got %v", err)
	}

	// Ubah label + urutan (patch: field lain kosong tidak boleh berubah).
	got, err := svc.UpdateMenu(ctx, "menu_stok", models.MenuUpdateInput{Label: "Stok Barang", SortOrder: intPtr(1)})
	if err != nil {
		t.Fatalf("update label/sort: %v", err)
	}
	if got.Label != "Stok Barang" || got.SortOrder != 1 || got.Module != "REPORT" {
		t.Fatalf("unexpected update result: %+v", got)
	}
	if got.Parent != "MENU_PAJU" || got.Kind != models.MenuKindChild {
		t.Fatalf("parent/jenis tidak boleh berubah: %+v", got)
	}

	// Parent harus se-modul.
	if _, err := svc.UpdateMenu(ctx, "MENU_STOK", models.MenuUpdateInput{Module: "toko", Parent: strPtr("MENU_PAJU")}); !errors.Is(err, services.ErrInvalidMenu) {
		t.Fatalf("parent beda modul harus ditolak, got %v", err)
	}
	// Lepas parent dengan string kosong, lalu pasang lagi.
	if _, err := svc.UpdateMenu(ctx, "MENU_STOK", models.MenuUpdateInput{Parent: strPtr("")}); err != nil {
		t.Fatalf("lepas parent: %v", err)
	}
	if cur, _ := svc.GetMenu(ctx, "MENU_STOK"); cur.Parent != "" {
		t.Fatalf("parent belum lepas: %+v", cur)
	}
	if _, err := svc.UpdateMenu(ctx, "MENU_STOK", models.MenuUpdateInput{Parent: strPtr("MENU_PAJU")}); err != nil {
		t.Fatalf("set parent lagi: %v", err)
	}
	// Parent tidak boleh punya anak -> parent.
	if _, err := svc.UpdateMenu(ctx, "MENU_PAJU", models.MenuUpdateInput{Parent: strPtr("MENU_STOK")}); !errors.Is(err, services.ErrInvalidMenu) {
		t.Fatalf("PARENT tidak boleh punya parent, got %v", err)
	}
	// Menu yang punya anak tidak boleh diubah jadi CHILD (butuh mcontrol juga).
	if _, err := svc.UpdateMenu(ctx, "MENU_PAJU", models.MenuUpdateInput{Kind: models.MenuKindChild, Mcontrol: strPtr("paju")}); !errors.Is(err, services.ErrMenuHasChildren) {
		t.Fatalf("parent beranak tidak boleh jadi CHILD, got %v", err)
	}
	// Menu tidak boleh jadi parent dirinya sendiri.
	if _, err := svc.UpdateMenu(ctx, "MENU_STOK", models.MenuUpdateInput{Parent: strPtr("MENU_STOK")}); !errors.Is(err, services.ErrInvalidMenu) {
		t.Fatalf("parent diri sendiri harus ditolak, got %v", err)
	}
	// Siklus: STOK anak PAJU, lalu STOK tidak boleh jadi parent PAJU lewat
	// HPLC -> tidak ada siklus karena PAJU tidak punya parent. Cek siklus
	// dengan mengubah PAJU jadi child STOK harus ditolak (butuh PAJU PARENT).
	if _, err := svc.UpdateMenu(ctx, "MENU_PAJU", models.MenuUpdateInput{Kind: models.MenuKindChild, Parent: strPtr("MENU_STOK")}); err == nil {
		t.Fatalf("siklus / parent beranak harus ditolak, got nil error")
	}
	// Modul tak dikenal dan menu tak ada.
	if _, err := svc.UpdateMenu(ctx, "MENU_STOK", models.MenuUpdateInput{Module: "hantu"}); !errors.Is(err, services.ErrInvalidMenu) {
		t.Fatalf("modul tak dikenal harus ditolak, got %v", err)
	}
	if _, err := svc.UpdateMenu(ctx, "MENU_HANTU", models.MenuUpdateInput{Label: "X"}); !errors.Is(err, services.ErrMenuNotFound) {
		t.Fatalf("menu tak ada harus 404, got %v", err)
	}
	// Patch semantik: label kosong = tidak diubah (bukan dihapus).
	before, _ := svc.GetMenu(ctx, "MENU_STOK")
	if _, err := svc.UpdateMenu(ctx, "MENU_STOK", models.MenuUpdateInput{Label: "   "}); err != nil {
		t.Fatalf("label kosong harus diabaikan, bukan error: %v", err)
	}
	if after, _ := svc.GetMenu(ctx, "MENU_STOK"); after.Label != before.Label {
		t.Fatalf("label berubah padahal tidak dikirim: %q -> %q", before.Label, after.Label)
	}
	// Grant role tidak boleh hilang karena edit menu.
	if err := svc.SetRolePermissions(ctx, models.RoleUser, []string{"MENU_STOK"}); err != nil {
		t.Fatalf("grant: %v", err)
	}
	if _, err := svc.UpdateMenu(ctx, "MENU_STOK", models.MenuUpdateInput{Label: "Stok Akhir"}); err != nil {
		t.Fatalf("update after grant: %v", err)
	}
	if perms, _ := svc.GetRolePermissions(ctx, models.RoleUser); len(perms) != 1 || perms[0] != "MENU_STOK" {
		t.Fatalf("grant hilang setelah edit menu: %v", perms)
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

// Menu PARENT tidak bisa di-grant: hanya CHILD yang boleh dicentang di matriks
// (grant parent ditolak, bukan disimpan diam-diam).
func TestGrantParentMenuRejected(t *testing.T) {
	svc, repo, _ := newTestService()
	ctx := context.Background()
	if _, err := svc.CreateRole(ctx, "editor", "Editor"); err != nil {
		t.Fatalf("create role: %v", err)
	}
	const parent, child = "MENU_KEUANGAN", "MENU_ARUS_KAS"
	repo.menus[parent] = &models.Menu{
		Code: parent, Module: "REPORT", Label: "Keuangan",
		Kind: models.MenuKindParent, SortOrder: 2,
	}
	repo.menus[child] = &models.Menu{
		Code: child, Module: "REPORT", Label: "Arus Kas", Mcontrol: "arus_kas",
		Kind: models.MenuKindChild, SortOrder: 3, Parent: parent,
	}

	err := svc.SetRolePermissions(ctx, "EDITOR", []string{models.MenuUsers, parent})
	if !errors.Is(err, services.ErrInvalidRole) {
		t.Fatalf("expected ErrInvalidRole for PARENT grant, got %v", err)
	}
	if _, ok := repo.perms["EDITOR"]; ok {
		t.Fatalf("grant must not be stored when a PARENT menu is included: %v", repo.perms["EDITOR"])
	}
	if err := svc.SetRolePermissions(ctx, "EDITOR", []string{child}); err != nil {
		t.Fatalf("child grant: %v", err)
	}
	if perms := repo.perms["EDITOR"]; len(perms) != 1 || perms[0] != child {
		t.Fatalf("expected only the child grant, got %v", perms)
	}
	// Menu yang tidak dikenal tetap ditolak seperti sebelumnya.
	if err := svc.SetRolePermissions(ctx, "EDITOR", []string{"MENU_TIDAK_ADA"}); !errors.Is(err, services.ErrInvalidRole) {
		t.Fatalf("expected ErrInvalidRole for unknown menu, got %v", err)
	}
}

// Matriks: baris PARENT bukan target grant (grantable=false) dan aksesnya
// menyala otomatis saat minimal satu anaknya ter-grant.
func TestMatrixParentAccessDerived(t *testing.T) {
	svc, repo, _ := newTestService()
	ctx := context.Background()
	const parent, child = "MENU_KEUANGAN", "MENU_ARUS_KAS"
	repo.menus[parent] = &models.Menu{
		Code: parent, Module: "REPORT", Label: "Keuangan",
		Kind: models.MenuKindParent, SortOrder: 2,
	}
	repo.menus[child] = &models.Menu{
		Code: child, Module: "REPORT", Label: "Arus Kas", Mcontrol: "arus_kas",
		Kind: models.MenuKindChild, SortOrder: 3, Parent: parent,
	}
	// ADMIN ter-grant menu anak itu (meniru centang di UI matriks).
	repo.perms[models.RoleAdmin] = append(repo.perms[models.RoleAdmin], child)

	rows, err := svc.GetMatrix(ctx, models.RoleAdmin)
	if err != nil {
		t.Fatalf("matrix: %v", err)
	}
	byCode := make(map[string]models.MatrixRow, len(rows))
	for _, r := range rows {
		byCode[r.MenuCode] = r
	}
	p := byCode[parent]
	if p.Grantable {
		t.Fatal("PARENT row must not be grantable")
	}
	if !p.HasAccess {
		t.Fatal("PARENT row must inherit access from its granted child")
	}
	if c := byCode[child]; !c.Grantable || !c.HasAccess {
		t.Fatalf("child row must be grantable + accessible: %+v", c)
	}

	// Role tanpa grant: header PARENT tidak menyala (tidak ada anak ter-grant).
	userRows, err := svc.GetMatrix(ctx, models.RoleUser)
	if err != nil {
		t.Fatalf("matrix user: %v", err)
	}
	for _, r := range userRows {
		if r.HasAccess {
			t.Fatalf("USER must have no access, got %+v", r)
		}
		if r.Kind == models.MenuKindParent && r.Grantable {
			t.Fatalf("PARENT must never be grantable: %+v", r)
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
	if _, err := svc.CreateMenu(ctx, models.MenuInput{Code: "MENU_X", Name: "X", Module: "REPORT", Label: "X", Mcontrol: "x"}); err != nil {
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
