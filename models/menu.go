package models

import "time"

// Module adalah baris tabel CPMATRIX (master modul).
// Satu baris = satu section sidebar. CPMENU.MCONTROL ber-FK ke CODE,
// sehingga modul wajib dibuat dulu sebelum menunya.
type Module struct {
	ID        int       `json:"-"`
	Code      string    `json:"code"`
	Label     string    `json:"label"`
	SortOrder int       `json:"sort_order"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

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
// Name diterima untuk kompatibilitas API; label tampil diambil dari Label.
type MenuInput struct {
	Code      string `json:"code"`
	Name      string `json:"name"`
	Module    string `json:"module"`
	Label     string `json:"label"`
	SortOrder int    `json:"sort_order"`
	Parent    string `json:"parent_code,omitempty"`
}

// MenuUpdateInput adalah payload PUT /api/admin/menus/{code} (patch semantik).
// Field kosong berarti tidak diubah, kecuali Parent: string kosong = lepas parent.
// SortOrder pointer supaya "0" (posisi teratas) bisa dikirim dan dibedakan
// dari "tidak diisi".
type MenuUpdateInput struct {
	Module    string `json:"module"`
	Label     string `json:"label"`
	SortOrder *int   `json:"sort_order"`
	Parent    string `json:"parent_code"`
}

// MenuEntry adalah menu milik user (ada grant di CPPERMISSION).
// Frontend membandingkan dengan folder registry: yang tidak punya folder
// dirender sebagai halaman 404 pemandu (tahu harus bikin di mana).
type MenuEntry struct {
	Code      string `json:"code"`
	Module    string `json:"module"`
	Label     string `json:"label"`
	SortOrder int    `json:"sort_order"`
	Parent    string `json:"parent_code,omitempty"`
}

// MatrixRow adalah 1 baris matriks: 1 role x 1 menu + flag akses.
// Dibaca via JOIN (CPROLE x CPMENU LEFT JOIN CPPERMISSION); tulis grant
// tetap ke CPPERMISSION (tidak ada sinkron ganda).
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
