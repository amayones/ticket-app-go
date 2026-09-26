package models

import "time"

// Menu adalah baris tabel CPMENU (registry menu).
// Satu baris = satu menu; MCONTROL = nama folder modul (UPPERCASE),
// wajib sama persis dengan folder frontend menus/<MCONTROL>/... .
// PARENT_CODE = menu induk untuk grup visual bersarang (boleh NULL).
type Menu struct {
	ID        int       `json:"-"`
	Code      string    `json:"code"`
	MControl  string    `json:"mcontrol"`
	Label     string    `json:"label"`
	SortOrder int       `json:"sort_order"`
	Parent    string    `json:"parent_code,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// MenuInput adalah payload POST /api/admin/menus (buat menu + permission).
type MenuInput struct {
	Code      string `json:"code"`
	Name      string `json:"name"`
	Module    string `json:"module"`
	Label     string `json:"label"`
	SortOrder int    `json:"sort_order"`
	Parent    string `json:"parent_code,omitempty"`
}

// MenuEntry adalah menu milik user (dari CPMATRIX, HAS_ACCESS = 1).
// Frontend membandingkan dengan folder registry: yang tidak punya folder
// dirender sebagai halaman 404 pemandu (tahu harus bikin di mana).
type MenuEntry struct {
	Code      string `json:"code"`
	Module    string `json:"module"`
	Label     string `json:"label"`
	SortOrder int    `json:"sort_order"`
	Parent    string `json:"parent_code,omitempty"`
}

// MatrixRow adalah 1 baris view CPMATRIX: 1 role x 1 menu + flag akses.
// Satu-satunya bacaan matriks (tulis tetap lewat CPROLEPERMISSION).
type MatrixRow struct {
	RoleCode  string `json:"role_code"`
	RoleName  string `json:"role_name"`
	Module    string `json:"module"`
	MenuCode  string `json:"menu_code"`
	MenuLabel string `json:"menu_label"`
	SortOrder int    `json:"sort_order"`
	Parent    string `json:"parent_code,omitempty"`
	HasAccess bool   `json:"has_access"`
}
