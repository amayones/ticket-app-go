package models

import "time"

// Session adalah tampilan sesi login aktif (dari CPREFRESHTOKEN).
// Hash token TIDAK PERNAH diekspos; ID dipakai untuk revoke per-sesi.
type Session struct {
	ID        int       `json:"id"`
	UserCode  string    `json:"user_code,omitempty"`
	Username  string    `json:"username,omitempty"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

// SecuritySummary adalah ringkasan menu Security Center.
type SecuritySummary struct {
	TotalUsers      int `json:"total_users"`
	TotalRoles      int `json:"total_roles"`
	ActiveSessions  int `json:"active_sessions"`
	AuditLast24h    int `json:"audit_last_24h"`
	ErrorsLast24h   int `json:"errors_last_24h"`
	NotifSentLast24h int `json:"notif_sent_last_24h"`
}
