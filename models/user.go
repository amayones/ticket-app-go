package models

import "time"

// User adalah baris tabel CPUSER. Identitas luar memakai Code
// (relasi antar tabel via CODE); ID hanya PK fisik, tidak diekspos ke API.
type User struct {
	ID        int       `json:"-"`
	Code      string    `json:"code"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Password  string    `json:"-"`
	RoleCode  string    `json:"role_code"`
	RoleName  string    `json:"role_name,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ToResponse converts the DB entity to the public API shape
// (never leaks password hash or internal ID).
func (u User) ToResponse() UserResponse {
	return UserResponse{
		Code:      u.Code,
		Username:  u.Username,
		Email:     u.Email,
		RoleCode:  u.RoleCode,
		RoleName:  u.RoleName,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}

// MeResponse adalah response GET /api/users/me: user + permission codes
// + entri menu miliknya (dari CPMATRIX). Frontend memakai Menus untuk
// sidebar (termasuk placeholder 404 bila folder belum dibuat).
type MeResponse struct {
	Code        string      `json:"code"`
	Username    string      `json:"username"`
	Email       string      `json:"email"`
	Role        string      `json:"role"`
	RoleCode    string      `json:"role_code"`
	RoleName    string      `json:"role_name,omitempty"`
	Permissions []string    `json:"permissions"`
	Menus       []MenuEntry `json:"menus"`
}

// UserResponse is the API DTO. Keep separate from User so DB schema
// changes don't silently become API breaking changes.
type UserResponse struct {
	Code      string    `json:"code"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	RoleCode  string    `json:"role_code"`
	RoleName  string    `json:"role_name,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
