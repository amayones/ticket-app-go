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
	DeleteRole(ctx context.Context, code string) (int, error)
	SetRolePermissions(ctx context.Context, roleCode string, permCodes []string) error
	UpdateUserRole(ctx context.Context, userCode, roleCode string) error
	CountRoles(ctx context.Context) (int, error)
	// Registry menu (CPMENU + CPMODULE tabel + grant CPPERMISSION).
	ListMenus(ctx context.Context) ([]models.Menu, error)
	GetMenu(ctx context.Context, code string) (*models.Menu, error)
	CreateMenu(ctx context.Context, input models.MenuInput) (*models.Menu, error)
	UpdateMenu(ctx context.Context, code string, input models.MenuUpdateInput) (*models.Menu, error)
	DeleteMenu(ctx context.Context, code string) error
	GetMatrix(ctx context.Context, roleFilter string) ([]models.MatrixRow, error)
	MyMenus(ctx context.Context, userCode string) ([]models.MenuEntry, error)
	// Master modul (CPMODULE tabel).
	ListModules(ctx context.Context) ([]models.Module, error)
	CreateModule(ctx context.Context, code, label string, sortOrder int) (*models.Module, error)
	DeleteModule(ctx context.Context, code string) error
}

// userStore dipenuhi users.Repository (tanpa import antar-fitur).
type userStore interface {
	GetByCode(ctx context.Context, code string) (*models.User, error)
	UpdateRole(ctx context.Context, code, roleCode string) error
	CountByRole(ctx context.Context, roleCode string) (int, error)
	// DeleteByRole menghapus semua user pada satu role dan mengembalikan
	// jumlah yang terhapus (dipakai cascade hapus role).
	DeleteByRole(ctx context.Context, roleCode string) (int, error)
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

// DeleteRole menghapus role beserta seluruh user yang memakainya.
// Grant role->menu (CPPERMISSION) dan refresh token ikut terhapus
// (FK ON DELETE CASCADE; penghapusan eksplisit dipakai sebagai fallback untuk
// engine yang tidak menegakkan FK, mis. sqlite tanpa pragma foreign_keys).
// Return: jumlah user yang ikut terhapus (0 bila role memang tak terpakai).
func (s *Service) DeleteRole(ctx context.Context, code string) (int, error) {
	code = strings.TrimSpace(strings.ToUpper(code))
	if code == models.RoleAdmin || code == models.RoleUser {
		return 0, services.ErrRoleProtected
	}
	if _, err := s.roles.GetByCode(ctx, code); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, services.ErrRoleNotFound
		}
		return 0, fmt.Errorf("get role: %w", err)
	}
	// Urutan: user dulu (CPUSER.ROLE_CODE FK ke CPROLE), baru role-nya.
	// Sesi/refresh token user ikut terhapus lewat CASCADE CPREFRESHTOKEN.
	deleted, err := s.users.DeleteByRole(ctx, code)
	if err != nil {
		return 0, fmt.Errorf("delete role users: %w", err)
	}
	if err := s.roles.DeleteGrants(ctx, code); err != nil {
		return 0, fmt.Errorf("delete role grants: %w", err)
	}
	if err := s.roles.Delete(ctx, code); err != nil {
		msg := strings.ToLower(err.Error())
		if strings.Contains(msg, "fk_") || strings.Contains(msg, "547") ||
			strings.Contains(msg, "foreign key") {
			return 0, services.ErrRoleInUse
		}
		return 0, fmt.Errorf("delete role: %w", err)
	}
	return deleted, nil
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
	known := make(map[string]models.Permission, len(all))
	for _, p := range all {
		if strings.HasPrefix(p.Code, "MENU_") {
			known[p.Code] = p
		}
	}
	cleaned := make([]string, 0, len(permCodes))
	seen := make(map[string]bool, len(permCodes))
	for _, pc := range permCodes {
		pc = strings.TrimSpace(strings.ToUpper(pc))
		p, ok := known[pc]
		if !ok {
			return fmt.Errorf("%w: unknown menu permission: %s", services.ErrInvalidRole, pc)
		}
		// Menu PARENT = header buka-tutup, bukan menu yang bisa dibuka: aksesnya
		// menyusul otomatis dari menu CHILD di bawahnya (lihat MyMenusByRole),
		// jadi grant untuk kode PARENT ditolak — bukan disimpan diam-diam.
		if p.Kind == models.MenuKindParent {
			return fmt.Errorf("%w: PARENT menu is a header without its own page, grant its child menus instead: %s",
				services.ErrInvalidRole, pc)
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
		return fmt.Errorf("%w: module must be 1-40 characters", services.ErrInvalidModule)
	}
	for _, r := range m {
		if !(r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '_') {
			return fmt.Errorf("%w: module must be uppercase letters, digits, or underscore", services.ErrInvalidModule)
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

// CreateMenu menulis registry menu (permission + tampilan) dalam 1 baris.
// Tanpa auto-grant ke role mana pun (admin mencentang manual via matriks).
//
// Kind menentukan peran baris: MenuKindParent (header buka-tutup di sidebar,
// TANPA mcontrol/folder, tidak boleh punya parent) atau MenuKindChild
// (item biasa: mcontrol wajib = nama folder frontend/src/app/<mcontrol>/,
// boleh punya PARENT_CODE). Judul tampil = LABEL.
func (s *Service) CreateMenu(ctx context.Context, input models.MenuInput) (*models.Menu, error) {
	code := strings.TrimSpace(strings.ToUpper(input.Code))
	name := strings.TrimSpace(input.Name)
	module := strings.TrimSpace(strings.ToUpper(input.Module))
	label := strings.TrimSpace(input.Label)
	mcontrol := strings.ToLower(strings.TrimSpace(input.Mcontrol))
	if err := validMenuCode(code); err != nil {
		return nil, err
	}
	if err := validModule(module); err != nil {
		return nil, err
	}
	// Modul wajib sudah terdaftar di CPMODULE (urut: modul -> menu).
	if _, err := s.roles.GetModule(ctx, module); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%w: module not found in CPMODULE: %s (buat modul dulu)", services.ErrInvalidMenu, module)
		}
		return nil, fmt.Errorf("get module: %w", err)
	}
	// Catatan: input.Name tidak disimpan. Nama permission di matriks diambil
	// dari LABEL (lihat ListPermissions), jadi cukup label yang wajib.
	_ = name
	if label == "" {
		return nil, fmt.Errorf("%w: menu label is required", services.ErrInvalidMenu)
	}
	if len(name) > 100 || len(label) > 100 {
		return nil, fmt.Errorf("%w: name/label must be at most 100 characters", services.ErrInvalidMenu)
	}
	sortOrder := input.SortOrder
	if sortOrder < 0 || sortOrder > 9999 {
		return nil, fmt.Errorf("%w: sort order must be 0-9999", services.ErrInvalidMenu)
	}
	kind, err := normalizeKind(input.Kind)
	if err != nil {
		return nil, err
	}
	if err := validMcontrol(kind, mcontrol); err != nil {
		return nil, err
	}
	parent := strings.TrimSpace(strings.ToUpper(input.Parent))
	if err := s.validateParent(ctx, &models.Menu{
		Code: code, Module: module, Kind: kind, Parent: parent,
	}); err != nil {
		return nil, err
	}
	menu := &models.Menu{
		Code: code, Module: module, Label: label, Mcontrol: mcontrol,
		Kind: kind, SortOrder: sortOrder, Parent: parent,
	}
	if err := s.roles.CreateMenu(ctx, menu); err != nil {
		if isDuplicateErr(err) {
			return nil, services.ErrPermissionExists
		}
		return nil, fmt.Errorf("create menu: %w", err)
	}
	return menu, nil
}

// validMcontrol memastikan aturan folder: CHILD wajib punya mcontrol
// (snake_case, dipakai sebagai nama folder frontend/src/app/<mcontrol>/),
// PARENT wajib kosong (header buka-tutup tanpa halaman).
func validMcontrol(kind, mcontrol string) error {
	if kind == models.MenuKindParent {
		if mcontrol != "" {
			return fmt.Errorf("%w: a PARENT menu must not have mcontrol (no folder)", services.ErrInvalidMenu)
		}
		return nil
	}
	if mcontrol == "" {
		return fmt.Errorf("%w: mcontrol is required for CHILD menu (folder name)", services.ErrInvalidMenu)
	}
	if len(mcontrol) > 40 {
		return fmt.Errorf("%w: mcontrol must be at most 40 characters", services.ErrInvalidMenu)
	}
	for _, r := range mcontrol {
		if !(r >= 'a' && r <= 'z' || r >= '0' && r <= '9' || r == '_') {
			return fmt.Errorf("%w: mcontrol must be snake_case (a-z, 0-9, _)", services.ErrInvalidMenu)
		}
	}
	if strings.HasPrefix(mcontrol, "_") || strings.HasSuffix(mcontrol, "_") || strings.Contains(mcontrol, "__") {
		return fmt.Errorf("%w: mcontrol must not start/end with _ or contain __", services.ErrInvalidMenu)
	}
	return nil
}

// normalizeKind menerima PARENT/CHILD (case-insensitive); kosong = CHILD.
func normalizeKind(kind string) (string, error) {
	k := strings.TrimSpace(strings.ToUpper(kind))
	switch k {
	case "":
		return models.MenuKindChild, nil
	case models.MenuKindParent:
		return models.MenuKindParent, nil
	case models.MenuKindChild:
		return models.MenuKindChild, nil
	default:
		return "", fmt.Errorf("%w: menu kind must be PARENT or CHILD", services.ErrInvalidMenu)
	}
}

// validateParent memastikan aturan parent-child pada menu (dipakai create
// dan update):
//   - menu PARENT tidak boleh punya parent;
//   - PARENT_CODE harus menunjuk menu bertipe PARENT, se-modul, bukan diri
//     sendiri, dan tidak menimbulkan siklus;
//   - parent tidak boleh keturunan menu yang sedang disetel.
func (s *Service) validateParent(ctx context.Context, m *models.Menu) error {
	if m.Kind == models.MenuKindParent {
		if m.Parent != "" {
			return fmt.Errorf("%w: a PARENT menu cannot have a parent", services.ErrInvalidMenu)
		}
		return nil
	}
	if m.Parent == "" {
		return nil
	}
	if m.Parent == m.Code {
		return fmt.Errorf("%w: menu cannot be its own parent", services.ErrInvalidMenu)
	}
	pm, err := s.roles.GetMenu(ctx, m.Parent)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("%w: parent menu not found: %s", services.ErrInvalidMenu, m.Parent)
		}
		return fmt.Errorf("get parent menu: %w", err)
	}
	if !pm.IsParent() {
		return fmt.Errorf("%w: parent menu must be kind PARENT: %s", services.ErrInvalidMenu, m.Parent)
	}
	if pm.Module != m.Module {
		return fmt.Errorf("%w: parent must be in the same module", services.ErrInvalidMenu)
	}
	// Cegah siklus: parent tidak boleh keturunan menu ini.
	descendants, err := s.descendantsOf(ctx, m.Code)
	if err != nil {
		return err
	}
	if descendants[m.Parent] {
		return fmt.Errorf("%w: parent would create a cycle", services.ErrInvalidMenu)
	}
	return nil
}

// UpdateMenu mengubah data menu yang boleh diubah: label, urutan, modul,
// mcontrol, jenis (PARENT/CHILD), dan parent. CODE (permission) tidak bisa
// diubah karena jadi acuan grant CPPERMISSION; untuk mengganti permission,
// buat menu baru lalu hapus yang lama.
//
// Field yang dikosongkan pada input tidak di-update (patch semantik),
// kecuali Parent dan Mcontrol: pointer non-nil berarti "setel ke nilai ini"
// (string kosong = lepas parent / kosongkan mcontrol — khusus PARENT).
func (s *Service) UpdateMenu(ctx context.Context, code string, input models.MenuUpdateInput) (*models.Menu, error) {
	code = strings.TrimSpace(strings.ToUpper(code))
	current, err := s.GetMenu(ctx, code)
	if err != nil {
		return nil, err
	}
	updated := *current
	if v := strings.TrimSpace(input.Label); v != "" {
		updated.Label = v
	}
	if input.Mcontrol != nil {
		updated.Mcontrol = strings.ToLower(strings.TrimSpace(*input.Mcontrol))
	}
	if v := strings.TrimSpace(strings.ToUpper(input.Module)); v != "" {
		updated.Module = v
	}
	if input.SortOrder != nil {
		updated.SortOrder = *input.SortOrder
	}
	if v := strings.TrimSpace(input.Kind); v != "" {
		kind, err := normalizeKind(v)
		if err != nil {
			return nil, err
		}
		updated.Kind = kind
	}
	if updated.Kind == "" {
		updated.Kind = models.MenuKindChild
	}
	// Parent hanya berubah bila dikirim: string kosong = lepas parent.
	if input.Parent != nil {
		updated.Parent = strings.TrimSpace(strings.ToUpper(*input.Parent))
	}
	if err := validMcontrol(updated.Kind, updated.Mcontrol); err != nil {
		return nil, err
	}
	if updated.Label == "" {
		return nil, fmt.Errorf("%w: menu label is required", services.ErrInvalidMenu)
	}
	if len(updated.Label) > 100 {
		return nil, fmt.Errorf("%w: label must be at most 100 characters", services.ErrInvalidMenu)
	}
	if updated.SortOrder < 0 || updated.SortOrder > 9999 {
		return nil, fmt.Errorf("%w: sort order must be 0-9999", services.ErrInvalidMenu)
	}
	if err := validModule(updated.Module); err != nil {
		return nil, err
	}
	if _, err := s.roles.GetModule(ctx, updated.Module); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%w: module not found in CPMODULE: %s (buat modul dulu)", services.ErrInvalidMenu, updated.Module)
		}
		return nil, fmt.Errorf("get module: %w", err)
	}
	// Menu yang sudah punya anak tidak boleh diturunkan jadi CHILD, karena
	// anak-anaknya butuh parent bertipe PARENT.
	if updated.Kind == models.MenuKindChild {
		children, err := s.roles.ListChildren(ctx, updated.Code)
		if err != nil {
			return nil, fmt.Errorf("list children: %w", err)
		}
		if len(children) > 0 {
			return nil, fmt.Errorf("%w: menu still has %d child menu(s)", services.ErrMenuHasChildren, len(children))
		}
	}
	if err := s.validateParent(ctx, &updated); err != nil {
		return nil, err
	}
	if err := s.roles.UpdateMenu(ctx, &updated); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, services.ErrMenuNotFound
		}
		return nil, fmt.Errorf("update menu: %w", err)
	}
	return &updated, nil
}

// descendantsOf mengembalikan kode menu yang berada di bawah `code`
// (anak, cucu, dst) untuk mendeteksi siklus parent.
func (s *Service) descendantsOf(ctx context.Context, code string) (map[string]bool, error) {
	out := map[string]bool{}
	queue := []string{code}
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		children, err := s.roles.ListChildren(ctx, cur)
		if err != nil {
			return nil, fmt.Errorf("list children: %w", err)
		}
		for _, c := range children {
			if out[c.Code] {
				continue
			}
			out[c.Code] = true
			queue = append(queue, c.Code)
		}
	}
	return out, nil
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

// GetMatrix membaca matriks via JOIN (roleFilter kosong = semua role).
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
	return decorateMatrix(rows), nil
}

// decorateMatrix menerapkan aturan menu PARENT pada matriks:
//
//   - Grantable = false untuk PARENT: baris header tidak punya centang akses
//     (grant ditolak SetRolePermissions), hanya CHILD yang bisa di-grant;
//   - HasAccess PARENT diturunkan dari keturunannya: bila minimal satu menu
//     CHILD di bawahnya ter-grant, header itu ikut ter-include — sama seperti
//     MyMenusByRole yang menaikkan ancestor untuk sidebar.
func decorateMatrix(rows []models.MatrixRow) []models.MatrixRow {
	index := make(map[string]map[string]int, 4)
	for i := range rows {
		if rows[i].Kind != models.MenuKindParent {
			rows[i].Grantable = true
		}
		byCode := index[rows[i].RoleCode]
		if byCode == nil {
			byCode = make(map[string]int, len(rows))
			index[rows[i].RoleCode] = byCode
		}
		byCode[rows[i].MenuCode] = i
	}
	for i := range rows {
		if !rows[i].HasAccess {
			continue
		}
		byCode := index[rows[i].RoleCode]
		// Naik ke atas: anak ter-grant menyalakan header PARENT-nya (berlapis
		// pun aman; guard mencegah siklus bila data di luar aturan).
		for parent, guard := rows[i].Parent, 0; parent != "" && guard < 10; guard++ {
			pi, ok := byCode[parent]
			if !ok {
				break
			}
			rows[pi].HasAccess = true
			parent = rows[pi].Parent
		}
	}
	return rows
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

// --- Master modul (CPMODULE tabel) --------------------------------------------

func (s *Service) ListModules(ctx context.Context) ([]models.Module, error) {
	modules, err := s.roles.ListModules(ctx)
	if err != nil {
		return nil, fmt.Errorf("list modules: %w", err)
	}
	return modules, nil
}

// CreateModule mendaftarkan modul baru (tanpa menu di dalamnya).
func (s *Service) CreateModule(ctx context.Context, code, label string, sortOrder int) (*models.Module, error) {
	code = strings.TrimSpace(strings.ToUpper(code))
	label = strings.TrimSpace(label)
	if err := validModule(code); err != nil {
		return nil, err
	}
	if label == "" || len(label) > 100 {
		return nil, fmt.Errorf("%w: module label is required (max 100 chars)", services.ErrInvalidModule)
	}
	if sortOrder < 0 || sortOrder > 9999 {
		return nil, fmt.Errorf("%w: sort order must be 0-9999", services.ErrInvalidModule)
	}
	m := &models.Module{Code: code, Label: label, SortOrder: sortOrder}
	if err := s.roles.CreateModule(ctx, m); err != nil {
		if isDuplicateErr(err) {
			return nil, services.ErrModuleExists
		}
		return nil, fmt.Errorf("create module: %w", err)
	}
	return m, nil
}

// DeleteModule menghapus modul; ditolak bila masih ada menu di dalamnya.
func (s *Service) DeleteModule(ctx context.Context, code string) error {
	code = strings.TrimSpace(strings.ToUpper(code))
	if _, err := s.roles.GetModule(ctx, code); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return services.ErrModuleNotFound
		}
		return fmt.Errorf("get module: %w", err)
	}
	n, err := s.roles.CountModuleMenus(ctx, code)
	if err != nil {
		return fmt.Errorf("check module menus: %w", err)
	}
	if n > 0 {
		return services.ErrModuleInUse
	}
	if err := s.roles.DeleteModule(ctx, code); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return services.ErrModuleNotFound
		}
		msg := strings.ToLower(err.Error())
		if strings.Contains(msg, "fk_") || strings.Contains(msg, "547") ||
			strings.Contains(msg, "foreign key") {
			return services.ErrModuleInUse
		}
		return fmt.Errorf("delete module: %w", err)
	}
	return nil
}
