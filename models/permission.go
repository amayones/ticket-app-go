package models

import "time"

// Permission codes (seed di scripts/migrate2_rbac.sql). Tambah permission
// baru = tambah seed + mapping role di SQL, lalu pakai konstantanya di sini.
const (
	PermUserRead        = "USER_READ"
	PermUserCreate      = "USER_CREATE"
	PermUserUpdate      = "USER_UPDATE"
	PermUserDelete      = "USER_DELETE"
	PermUserRoleAssign  = "USER_ROLE_ASSIGN"
	PermRoleRead        = "ROLE_READ"
	PermRoleManage      = "ROLE_MANAGE"
	PermPermissionAssign = "PERMISSION_ASSIGN"
	PermSessionRead     = "SESSION_READ"
	PermSessionRevoke   = "SESSION_REVOKE"
	PermSessionManage   = "SESSION_MANAGE"
	PermAuditRead       = "AUDIT_READ"
	PermSecurityRead    = "SECURITY_READ"
	PermSyslogRead      = "SYSLOG_READ"
	PermSyslogManage    = "SYSLOG_MANAGE"
	PermNotifRead       = "NOTIF_READ"
	PermNotifManage     = "NOTIF_MANAGE"
	PermNotifSend       = "NOTIF_SEND"
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
