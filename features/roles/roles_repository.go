package roles

import (
	"context"
	"database/sql"
	"strings"

	"golang-backend/models"
	"golang-backend/repositories"
)

// RepositoryInterface reads/writes CPROLE + permission mapping.
type RepositoryInterface interface {
	List(ctx context.Context) ([]models.Role, error)
	GetByCode(ctx context.Context, code string) (*models.Role, error)
	RoleExists(ctx context.Context, code string) (bool, error)
	Create(ctx context.Context, role *models.Role) error
	Delete(ctx context.Context, code string) error
	DeleteGrants(ctx context.Context, roleCode string) error
	Count(ctx context.Context) (int, error)
	ListPermissions(ctx context.Context) ([]models.Permission, error)
	GetRolePermissions(ctx context.Context, roleCode string) ([]string, error)
	SetRolePermissions(ctx context.Context, roleCode string, permCodes []string) error
	HasPermission(ctx context.Context, roleCode, permCode string) (bool, error)
	// Registry menu (CPMENU + CPMATRIX tabel + grant CPPERMISSION).
	ListMenus(ctx context.Context) ([]models.Menu, error)
	GetMenu(ctx context.Context, code string) (*models.Menu, error)
	CreateMenu(ctx context.Context, menu *models.Menu) error
	UpdateMenu(ctx context.Context, menu *models.Menu) error
	DeleteMenu(ctx context.Context, code string) error
	CountMenuUsage(ctx context.Context, code string) (int, error)
	ListChildren(ctx context.Context, code string) ([]models.Menu, error)
	QueryMatrix(ctx context.Context, roleFilter string) ([]models.MatrixRow, error)
	MyMenusByRole(ctx context.Context, roleCode string) ([]models.MenuEntry, error)
	// Master modul (CPMODULE tabel).
	ListModules(ctx context.Context) ([]models.Module, error)
	GetModule(ctx context.Context, code string) (*models.Module, error)
	CreateModule(ctx context.Context, m *models.Module) error
	DeleteModule(ctx context.Context, code string) error
	CountModuleMenus(ctx context.Context, code string) (int, error)
}

type Repository struct {
	db      *sql.DB
	dialect repositories.Dialect
}

func NewRepository(db *sql.DB, dialect repositories.Dialect) RepositoryInterface {
	return &Repository{db: db, dialect: dialect}
}

func (r *Repository) roleTable() string { return r.dialect.Table("CPROLE") }

// grantTable adalah satu-satunya tabel relasi grant: CPPERMISSION
// (ROLE_CODE -> MENU_CODE). Definisi menu tinggal di CPMENU.
func (r *Repository) grantTable() string { return r.dialect.Table("CPPERMISSION") }

func (r *Repository) menuTable() string { return r.dialect.Table("CPMENU") }

func (r *Repository) moduleTable() string { return r.dialect.Table("CPMODULE") }

func (r *Repository) List(ctx context.Context) ([]models.Role, error) {
	ctx, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	rows, err := r.db.QueryContext(ctx,
		`SELECT ID, CODE, NAME, CREATED_AT, UPDATED_AT FROM `+r.roleTable()+` ORDER BY CODE ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	roles := make([]models.Role, 0)
	for rows.Next() {
		var role models.Role
		if err := rows.Scan(&role.ID, &role.Code, &role.Name, &role.CreatedAt, &role.UpdatedAt); err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return roles, nil
}

func (r *Repository) RoleExists(ctx context.Context, code string) (bool, error) {
	ctx, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	var n int
	err := r.db.QueryRowContext(ctx, r.dialect.Bind(
		`SELECT COUNT(*) FROM `+r.roleTable()+` WHERE CODE = ?`), code).Scan(&n)
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

func (r *Repository) GetByCode(ctx context.Context, code string) (*models.Role, error) {
	ctx, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	var role models.Role
	err := r.db.QueryRowContext(ctx, r.dialect.Bind(
		`SELECT ID, CODE, NAME, CREATED_AT, UPDATED_AT FROM `+r.roleTable()+` WHERE CODE = ?`), code,
	).Scan(&role.ID, &role.Code, &role.Name, &role.CreatedAt, &role.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *Repository) Create(ctx context.Context, role *models.Role) error {
	ctx, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	_, err := r.db.ExecContext(ctx, r.dialect.Bind(
		`INSERT INTO `+r.roleTable()+` (CODE, NAME) VALUES (?, ?)`), role.Code, role.Name)
	return err
}

func (r *Repository) Delete(ctx context.Context, code string) error {
	ctx, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	res, err := r.db.ExecContext(ctx,
		r.dialect.Bind(`DELETE FROM `+r.roleTable()+` WHERE CODE = ?`), code)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *Repository) Count(ctx context.Context) (int, error) {
	ctx, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	var n int
	err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM `+r.roleTable()).Scan(&n)
	return n, err
}

func (r *Repository) ListPermissions(ctx context.Context) ([]models.Permission, error) {
	ctx, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	// Definisi permission = registry menu (1:1): CODE/Nama/Grup dari CPMENU.
	// Urutan "per section": tiap PARENT langsung diikuti anak-anaknya
	// (GROUP_SORT = urutan parent; menu tanpa parent memakai urutannya
	// sendiri) supaya UI matriks tinggal menyusun header + anaknya berurutan.
	rows, err := r.db.QueryContext(ctx,
		`SELECT m.CODE, m.LABEL, m.MODULE, m.MENU_KIND, m.PARENT_CODE, m.CREATED_AT
		FROM `+r.menuTable()+` m
		LEFT JOIN `+r.menuTable()+` p ON p.CODE = m.PARENT_CODE
		ORDER BY m.MODULE ASC, COALESCE(p.SORT_ORDER, m.SORT_ORDER) ASC,
		  CASE WHEN m.MENU_KIND = 'PARENT' THEN 0 ELSE 1 END ASC,
		  m.SORT_ORDER ASC, m.CODE ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	perms := make([]models.Permission, 0)
	for rows.Next() {
		var p models.Permission
		var parent sql.NullString
		if err := rows.Scan(&p.Code, &p.Name, &p.Group, &p.Kind, &parent, &p.CreatedAt); err != nil {
			return nil, err
		}
		if !strings.HasPrefix(p.Code, "MENU_") {
			continue
		}
		p.Parent = parent.String
		perms = append(perms, p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return perms, nil
}

func (r *Repository) GetRolePermissions(ctx context.Context, roleCode string) ([]string, error) {
	ctx, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	// JOIN ke CPMENU: hanya menu CHILD yang berarti sebagai akses. Grant yang
	// menempel di baris PARENT (sisa data lama — baris header tidak bisa
	// dicentang lagi) disaring agar tidak muncul sebagai permission role.
	rows, err := r.db.QueryContext(ctx, r.dialect.Bind(
		`SELECT g.MENU_CODE FROM `+r.grantTable()+` g
		JOIN `+r.menuTable()+` m ON m.CODE = g.MENU_CODE
		WHERE g.ROLE_CODE = ? AND m.MENU_KIND <> 'PARENT'
		ORDER BY g.MENU_CODE ASC`), roleCode)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	codes := make([]string, 0)
	for rows.Next() {
		var c string
		if err := rows.Scan(&c); err != nil {
			return nil, err
		}
		if strings.HasPrefix(c, "MENU_") {
			codes = append(codes, c)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return codes, nil
}

// SetRolePermissions replaces the permission set atomically.
func (r *Repository) SetRolePermissions(ctx context.Context, roleCode string, permCodes []string) error {
	ctx, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, r.dialect.Bind(
		`DELETE FROM `+r.grantTable()+` WHERE ROLE_CODE = ?`), roleCode); err != nil {
		return err
	}
	for _, pc := range permCodes {
		if _, err := tx.ExecContext(ctx, r.dialect.Bind(
			`INSERT INTO `+r.grantTable()+` (ROLE_CODE, MENU_CODE) VALUES (?, ?)`),
			roleCode, pc); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (r *Repository) HasPermission(ctx context.Context, roleCode, permCode string) (bool, error) {
	ctx, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	var n int
	err := r.db.QueryRowContext(ctx, r.dialect.Bind(
		`SELECT COUNT(*) FROM `+r.grantTable()+` WHERE ROLE_CODE = ? AND MENU_CODE = ?`),
		roleCode, permCode).Scan(&n)
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

// --- Registry menu (CPMENU) + JOIN modul untuk label section -------------

const menuColumns = `m.CODE, m.MODULE, mod.LABEL, mod.SORT_ORDER, m.LABEL, m.MCONTROL, m.MENU_KIND, m.SORT_ORDER, m.PARENT_CODE, m.CREATED_AT, m.UPDATED_AT`

func scanMenu(row interface {
	Scan(dest ...any) error
}, m *models.Menu) error {
	var parent sql.NullString
	var mcontrol sql.NullString
	var moduleLabel sql.NullString
	var moduleSort sql.NullInt64
	err := row.Scan(&m.Code, &m.Module, &moduleLabel, &moduleSort, &m.Label, &mcontrol, &m.Kind, &m.SortOrder, &parent, &m.CreatedAt, &m.UpdatedAt)
	if err != nil {
		return err
	}
	m.Parent = parent.String
	m.Mcontrol = mcontrol.String
	m.ModuleLabel = moduleLabel.String
	if moduleSort.Valid {
		m.ModuleSort = int(moduleSort.Int64)
	}
	return nil
}

func (r *Repository) menuFrom() string {
	return r.menuTable() + ` m LEFT JOIN ` + r.moduleTable() + ` mod ON mod.CODE = m.MODULE`
}

func (r *Repository) ListMenus(ctx context.Context) ([]models.Menu, error) {
	ctx, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	query := `SELECT ` + menuColumns + ` FROM ` + r.menuFrom() + ` ORDER BY mod.SORT_ORDER ASC, m.MODULE ASC, m.SORT_ORDER ASC, m.CODE ASC`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]models.Menu, 0)
	for rows.Next() {
		var m models.Menu
		if err := scanMenu(rows, &m); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *Repository) GetMenu(ctx context.Context, code string) (*models.Menu, error) {
	ctx, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	query := `SELECT ` + menuColumns + ` FROM ` + r.menuFrom() + ` WHERE m.CODE = ?`
	var m models.Menu
	if err := scanMenu(r.db.QueryRowContext(ctx, r.dialect.Bind(query), code), &m); err != nil {
		return nil, err
	}
	return &m, nil
}

// CreateMenu menulis satu baris registry CPMENU. Permission = menu itu
// sendiri (1:1); grant role diatur terpisah via SetRolePermissions.
// MCONTROL NULL untuk PARENT (tanpa folder), wajib isi untuk CHILD.
func (r *Repository) CreateMenu(ctx context.Context, menu *models.Menu) error {
	ctx, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	query := `INSERT INTO ` + r.menuTable() + ` (CODE, MODULE, LABEL, MCONTROL, MENU_KIND, SORT_ORDER, PARENT_CODE) VALUES (?, ?, ?, ?, ?, ?, ?)`
	var parent any
	if menu.Parent != "" {
		parent = menu.Parent
	}
	var mcontrol any
	if strings.TrimSpace(menu.Mcontrol) != "" {
		mcontrol = strings.ToLower(strings.TrimSpace(menu.Mcontrol))
	}
	_, err := r.db.ExecContext(ctx, r.dialect.Bind(query),
		menu.Code, menu.Module, menu.Label, mcontrol, menu.Kind, menu.SortOrder, parent)
	return err
}

// UpdateMenu mengubah label/urutan/modul/mcontrol/parent sebuah menu.
// CODE tidak diubah: ia adalah permission yang jadi acuan grant
// CPPERMISSION, jadi memindahkannya = buat menu baru.
func (r *Repository) UpdateMenu(ctx context.Context, menu *models.Menu) error {
	ctx, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	query := `UPDATE ` + r.menuTable() + ` SET MODULE = ?, LABEL = ?, MCONTROL = ?, MENU_KIND = ?, SORT_ORDER = ?, PARENT_CODE = ? WHERE CODE = ?`
	var parent any
	if menu.Parent != "" {
		parent = menu.Parent
	}
	var mcontrol any
	if strings.TrimSpace(menu.Mcontrol) != "" {
		mcontrol = strings.ToLower(strings.TrimSpace(menu.Mcontrol))
	}
	res, err := r.db.ExecContext(ctx, r.dialect.Bind(query),
		menu.Module, menu.Label, mcontrol, menu.Kind, menu.SortOrder, parent, menu.Code)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// DeleteGrants menghapus semua grant menu milik satu role.
// Dipakai saat hapus role; DB juga punya ON DELETE CASCADE, jadi pernyataan
// ini hanya jaminan untuk engine yang tidak menegakkan FK.
func (r *Repository) DeleteGrants(ctx context.Context, roleCode string) error {
	ctx, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	_, err := r.db.ExecContext(ctx, r.dialect.Bind(
		`DELETE FROM `+r.grantTable()+` WHERE ROLE_CODE = ?`), roleCode)
	return err
}

// DeleteMenu menghapus registry menu (grant role ikut CASCADE).
// Anak (PARENT_CODE) menahan hapus via FK; service memvalidasi lebih dulu.
func (r *Repository) DeleteMenu(ctx context.Context, code string) error {
	ctx, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	res, err := r.db.ExecContext(ctx, r.dialect.Bind(
		`DELETE FROM `+r.menuTable()+` WHERE CODE = ?`), code)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *Repository) CountMenuUsage(ctx context.Context, code string) (int, error) {
	ctx, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	var n int
	err := r.db.QueryRowContext(ctx, r.dialect.Bind(
		`SELECT COUNT(*) FROM `+r.grantTable()+` WHERE MENU_CODE = ?`), code).Scan(&n)
	if err != nil {
		return 0, err
	}
	return n, nil
}

func (r *Repository) ListChildren(ctx context.Context, code string) ([]models.Menu, error) {
	ctx, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	query := `SELECT ` + menuColumns + ` FROM ` + r.menuFrom() + ` WHERE m.PARENT_CODE = ? ORDER BY m.SORT_ORDER ASC, m.CODE ASC`
	rows, err := r.db.QueryContext(ctx, r.dialect.Bind(query), code)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]models.Menu, 0)
	for rows.Next() {
		var m models.Menu
		if err := scanMenu(rows, &m); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

// QueryMatrix membaca matriks via JOIN eksplisit (CPMODULE label modul,
// CPROLE x CPMENU LEFT JOIN CPPERMISSION grant). roleFilter kosong = semua role.
func (r *Repository) QueryMatrix(ctx context.Context, roleFilter string) ([]models.MatrixRow, error) {
	ctx, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	query := `SELECT r.CODE AS ROLE_CODE, r.NAME AS ROLE_NAME, m.MODULE AS MODULE,
		mod.LABEL AS MODULE_LABEL,
		m.CODE AS MENU_CODE, m.LABEL AS MENU_LABEL, m.MCONTROL, m.MENU_KIND, m.SORT_ORDER, m.PARENT_CODE,
		CASE WHEN pm.ROLE_CODE IS NULL THEN 0 ELSE 1 END AS HAS_ACCESS
		FROM ` + r.roleTable() + ` r CROSS JOIN ` + r.menuTable() + ` m
		LEFT JOIN ` + r.moduleTable() + ` mod ON mod.CODE = m.MODULE
		LEFT JOIN ` + r.grantTable() + ` pm
		  ON pm.ROLE_CODE = r.CODE AND pm.MENU_CODE = m.CODE`
	var args []any
	if strings.TrimSpace(roleFilter) != "" {
		query += ` WHERE r.CODE = ?`
		args = append(args, strings.ToUpper(strings.TrimSpace(roleFilter)))
	}
	query += ` ORDER BY ROLE_CODE ASC, mod.SORT_ORDER ASC, m.MODULE ASC, m.SORT_ORDER ASC, MENU_CODE ASC`
	rows, err := r.db.QueryContext(ctx, r.dialect.Bind(query), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]models.MatrixRow, 0)
	for rows.Next() {
		var row models.MatrixRow
		var parent sql.NullString
		var mcontrol sql.NullString
		var moduleLabel sql.NullString
		var access int
		if err := rows.Scan(&row.RoleCode, &row.RoleName, &row.Module, &moduleLabel, &row.MenuCode,
			&row.MenuLabel, &mcontrol, &row.Kind, &row.SortOrder, &parent, &access); err != nil {
			return nil, err
		}
		row.Parent = parent.String
		row.Mcontrol = mcontrol.String
		row.ModuleLabel = moduleLabel.String
		row.HasAccess = access != 0
		out = append(out, row)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

// MyMenusByRole mengembalikan menu milik satu role untuk sidebar:
// semua baris ter-grant + ancestor PARENT dari menu ter-grant (rekursif di
// memori agar jalan identik di mssql/postgres/sqlite tanpa CTE rekursif).
// PARENT tidak butuh grant sendiri: selama minimal satu keturunannya punya
// akses, header-nya ikut tampil.
func (r *Repository) MyMenusByRole(ctx context.Context, roleCode string) ([]models.MenuEntry, error) {
	ctx, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	query := `SELECT m.CODE, m.MODULE, mod.LABEL, mod.SORT_ORDER, m.LABEL, m.MCONTROL, m.MENU_KIND, m.SORT_ORDER, m.PARENT_CODE
		FROM ` + r.menuTable() + ` m JOIN ` + r.grantTable() + ` pm ON pm.MENU_CODE = m.CODE
		LEFT JOIN ` + r.moduleTable() + ` mod ON mod.CODE = m.MODULE
		WHERE pm.ROLE_CODE = ?`
	rows, err := r.db.QueryContext(ctx, r.dialect.Bind(query), roleCode)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	granted := make([]models.MenuEntry, 0)
	for rows.Next() {
		var e models.MenuEntry
		var parent sql.NullString
		var mcontrol sql.NullString
		var moduleLabel sql.NullString
		var moduleSort sql.NullInt64
		if err := rows.Scan(&e.Code, &e.Module, &moduleLabel, &moduleSort, &e.Label, &mcontrol, &e.Kind, &e.SortOrder, &parent); err != nil {
			return nil, err
		}
		// PARENT = header buka-tutup, tidak pernah jadi akar tampilan: grant
		// lama yang menempel di baris PARENT diabaikan (baris itu tetap masuk
		// lewat ancestor-walk di bawah bila ada anak yang ter-grant). Jadi yang
		// tampil = cabang PARENT yang benar-benar punya menu CHILD ter-grant.
		if e.Kind == models.MenuKindParent {
			continue
		}
		e.Parent = parent.String
		e.Mcontrol = mcontrol.String
		e.ModuleLabel = moduleLabel.String
		if moduleSort.Valid {
			e.ModuleSort = int(moduleSort.Int64)
		}
		granted = append(granted, e)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(granted) == 0 {
		return []models.MenuEntry{}, nil
	}
	// Peta semua menu (untuk menaikkan ancestor) — murah: tabel menu kecil.
	all, err := r.ListMenus(ctx)
	if err != nil {
		return nil, err
	}
	byCode := make(map[string]models.Menu, len(all))
	for _, m := range all {
		byCode[m.Code] = m
	}
	out := make([]models.MenuEntry, 0, len(granted)+4)
	seen := make(map[string]bool, len(granted)+4)
	for _, g := range granted {
		if !seen[g.Code] {
			seen[g.Code] = true
			out = append(out, g)
		}
		// Naik ke parent sampai akar; setiap ancestor ikut tampil walau
		// tanpa grant (header buka-tutup, tanpa halaman).
		for p := g.Parent; p != ""; {
			if seen[p] {
				// Lanjut naik walau sudah terlihat (kakeknya mungkin belum).
				if pm, ok := byCode[p]; ok {
					p = pm.Parent
					continue
				}
				break
			}
			pm, ok := byCode[p]
			if !ok {
				break
			}
			seen[p] = true
			out = append(out, models.MenuEntry{
				Code: pm.Code, Module: pm.Module, ModuleLabel: pm.ModuleLabel,
				ModuleSort: pm.ModuleSort, Label: pm.Label, Mcontrol: pm.Mcontrol,
				Kind: pm.Kind, SortOrder: pm.SortOrder, Parent: pm.Parent,
			})
			p = pm.Parent
		}
	}
	return out, nil
}

// --- Master modul (CPMODULE tabel) --------------------------------------------

func scanModule(row interface {
	Scan(dest ...any) error
}, m *models.Module) error {
	return row.Scan(&m.Code, &m.Label, &m.SortOrder, &m.CreatedAt, &m.UpdatedAt)
}

func (r *Repository) ListModules(ctx context.Context) ([]models.Module, error) {
	ctx, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	query := `SELECT CODE, LABEL, SORT_ORDER, CREATED_AT, UPDATED_AT FROM ` +
		r.moduleTable() + ` ORDER BY SORT_ORDER ASC, CODE ASC`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]models.Module, 0)
	for rows.Next() {
		var m models.Module
		if err := scanModule(rows, &m); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *Repository) GetModule(ctx context.Context, code string) (*models.Module, error) {
	ctx, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	query := `SELECT CODE, LABEL, SORT_ORDER, CREATED_AT, UPDATED_AT FROM ` +
		r.moduleTable() + ` WHERE CODE = ?`
	var m models.Module
	if err := scanModule(r.db.QueryRowContext(ctx, r.dialect.Bind(query), code), &m); err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *Repository) CreateModule(ctx context.Context, m *models.Module) error {
	ctx, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	query := `INSERT INTO ` + r.moduleTable() + ` (CODE, LABEL, SORT_ORDER) VALUES (?, ?, ?)`
	_, err := r.db.ExecContext(ctx, r.dialect.Bind(query), m.Code, m.Label, m.SortOrder)
	return err
}

func (r *Repository) DeleteModule(ctx context.Context, code string) error {
	ctx, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	res, err := r.db.ExecContext(ctx, r.dialect.Bind(
		`DELETE FROM `+r.moduleTable()+` WHERE CODE = ?`), code)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// CountModuleMenus menghitung menu yang menunjuk ke modul (FK menahan hapus).
func (r *Repository) CountModuleMenus(ctx context.Context, code string) (int, error) {
	ctx, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	var n int
	err := r.db.QueryRowContext(ctx, r.dialect.Bind(
		`SELECT COUNT(*) FROM `+r.menuTable()+` WHERE MODULE = ?`), code).Scan(&n)
	if err != nil {
		return 0, err
	}
	return n, nil
}
