package models

import "time"

// Permission codes (seed di scripts/migrate2_rbac.sql). Satu permission
// per menu: jika role punya permission ini, menu-nya tampil di sidebar.
// Tambah menu = tambah konstanta di sini + seed di SQL.
const (
	MenuDashboard     = "MENU_DASHBOARD"
	MenuUsers         = "MENU_USERS"
	MenuRoles         = "MENU_ROLES"
	MenuSessions      = "MENU_SESSIONS"
	MenuAudit         = "MENU_AUDIT"
	MenuSecurity      = "MENU_SECURITY"
	MenuSyslog        = "MENU_SYSLOG"
	MenuNotifications = "MENU_NOTIFICATIONS"
)

// Permission adalah baris tabel CPPERMISSION.
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
