package models

import "time"

// Aksi yang dicatat ke CPAUDITLOG.
const (
	AuditRegister         = "REGISTER"
	AuditLogin            = "LOGIN"
	AuditLoginFailed      = "LOGIN_FAILED"
	AuditLogout           = "LOGOUT"
	AuditLogoutAll        = "LOGOUT_ALL"
	AuditRefresh          = "REFRESH"
	AuditUpdateUser       = "UPDATE_USER"
	AuditDeleteUser       = "DELETE_USER"
	AuditRoleAssign       = "ROLE_ASSIGN"
	AuditRoleCreate       = "ROLE_CREATE"
	AuditRoleDelete       = "ROLE_DELETE"
	AuditPermissionAssign = "PERMISSION_ASSIGN"
	AuditSessionRevoke    = "SESSION_REVOKE"
	AuditNotifSend        = "NOTIF_SEND"
	AuditTemplateCreate   = "TEMPLATE_CREATE"
	AuditTemplateUpdate   = "TEMPLATE_UPDATE"
	AuditTemplateDelete   = "TEMPLATE_DELETE"
	AuditSyslogPrune      = "SYSLOG_PRUNE"
	AuditMenuCreate       = "MENU_CREATE"
	AuditMenuUpdate       = "MENU_UPDATE"
	AuditMenuDelete       = "MENU_DELETE"
	AuditModuleCreate     = "MODULE_CREATE"
	AuditModuleDelete     = "MODULE_DELETE"
)

// Entity yang dikenai aksi audit.
const (
	EntityUser         = "USER"
	EntityRole         = "ROLE"
	EntitySession      = "SESSION"
	EntityNotification = "NOTIFICATION"
	EntityTemplate     = "TEMPLATE"
	EntitySystem       = "SYSTEM"
	EntityAuth         = "AUTH"
)

// AuditLog adalah baris tabel CPAUDITLOG.
type AuditLog struct {
	ID         int       `json:"-"`
	Code       string    `json:"code"`
	ActorCode  string    `json:"actor_code,omitempty"`
	Action     string    `json:"action"`
	Entity     string    `json:"entity"`
	EntityCode string    `json:"entity_code,omitempty"`
	Detail     string    `json:"detail,omitempty"`
	IPAddress  string    `json:"ip_address,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
}

// AuditFilter menyaring daftar audit log.
type AuditFilter struct {
	Action string
	Entity string
	Actor  string
	Limit  int
	Offset int
}
