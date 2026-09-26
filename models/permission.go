package models

import "time"

// Permission codes (seed di scripts/migrate2_rbac.sql). Satu permission
// per menu: jika role punya permission ini (baris CPPERMISSION), role dapat
// memakai seluruh fungsi menu tersebut. Tambah menu = tambah baris CPMENU
// (via POST /api/admin/menus). Role tanpa akses apa pun mendapat halaman kosong.
const (
	MenuUsers         = "MENU_USERS"
	MenuRoles         = "MENU_ROLES"
	MenuSessions      = "MENU_SESSIONS"
	MenuAudit         = "MENU_AUDIT"
	MenuSecurity      = "MENU_SECURITY"
	MenuSyslog        = "MENU_SYSLOG"
	MenuNotifications = "MENU_NOTIFICATIONS"
	// MenuModul hanya mengatur tampilnya halaman Modul & Menu di sidebar;
	// API manajemen menu tetap di bawah MENU_ROLES.
	MenuModul = "MENU_MODUL"
)

// Permission adalah permission tampil satu menu (kode MENU_* dari CPMENU);
// grant per role tercatat di tabel CPPERMISSION.
type Permission struct {
	ID          int       `json:"-"`
	Code        string    `json:"code"`
	Name        string    `json:"name"`
	Group       string    `json:"group"`
	Description string    `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

// RoleDetail adalah role + daftar kode permission miliknya (untuk matriks RBAC).
type RoleDetail struct {
	Role
	Permissions []string `json:"permissions"`
}
