package models

import "time"

// Kode role yang dikenal (baris seed di scripts/migrate.sql).
const (
	RoleAdmin       = "ADMIN"
	RoleUser        = "USER"
	DefaultRoleCode = RoleUser
)

// Role adalah baris tabel CPROLE.
type Role struct {
	ID        int       `json:"-"`
	Code      string    `json:"code"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
