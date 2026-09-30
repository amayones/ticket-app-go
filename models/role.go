package models

import "time"

// Known role codes (seed rows in scripts/migrate.sql + migrate4_ticket_roles.sql).
//   ADMIN     = system operator (SYSTEM menu group only).
//   AUDIENCE  = end user / ticket buyer.
//   ORGANIZER = event organizer.
//   SELLER    = merchandise seller.
//   OFFICER   = check-in officer.
//   USER      = legacy (old installs); zero menus.
// New roles intentionally have zero menu grants: they get an empty page
// until an admin assigns menus via the Role & Permission matrix.
const (
	RoleAdmin     = "ADMIN"
	RoleUser      = "USER"
	RoleAudience  = "AUDIENCE"
	RoleOrganizer = "ORGANIZER"
	RoleSeller    = "SELLER"
	RoleOfficer   = "OFFICER"
	DefaultRoleCode = RoleAudience
)

// Role adalah baris tabel CPROLE.
type Role struct {
	ID        int       `json:"-"`
	Code      string    `json:"code"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
