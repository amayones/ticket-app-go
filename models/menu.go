package models

import "time"

// Nilai kolom CPMENU.MENU_KIND: menentukan apakah baris itu menu parent
// (dapat punya anak, tampil expandable di sidebar) atau menu child.
const (
	MenuKindParent = "PARENT"
	MenuKindChild  = "CHILD"
)

// Module adalah baris tabel CPMODULE (master modul).
// Satu baris = satu section sidebar. CPMENU.MODULE ber-FK ke CODE,
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
// Satu baris = satu menu; MODULE = nama folder modul (UPPERCASE), wajib
// sama persis dengan folder frontend menus/<MODULE>/<menu>/ (tanpa folder
// perantara: semua menu satu level di dalam folder modul).
// MENU_KIND menentukan peran baris: MenuKindParent (punya anak, expandable)
// atau MenuKindChild. PARENT_CODE = kode menu parent (wajib se-modul, dan
// parent-nya harus bertipe PARENT).
type Menu struct {
	ID        int       `json:"-"`
	Code      string    `json:"code"`
	Module    string    `json:"module"`
	Label     string    `json:"label"`
	Kind      string    `json:"kind"`
	SortOrder int       `json:"sort_order"`
	Parent    string    `json:"parent_code"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// IsParent melaporkan apakah baris ini berperan sebagai menu parent.
func (m Menu) IsParent() bool { return m.Kind == MenuKindParent }

// MenuInput adalah payload POST /api/admin/menus (buat menu).
// Kind opsional; kosong berarti CHILD. Name diterima untuk kompatibilitas
// API tapi tidak disimpan (nama permission diambil dari Label).
type MenuInput struct {
	Code      string `json:"code"`
	Name      string `json:"name"`
	Module    string `json:"module"`
	Label     string `json:"label"`
	Kind      string `json:"kind"`
	SortOrder int    `json:"sort_order"`
	Parent    string `json:"parent_code"`
}

// MenuUpdateInput adalah payload PUT /api/admin/menus/{code} (patch semantik).
// Field yang tidak dikirim = tidak diubah. Parent dan SortOrder berupa
// pointer agar "tidak dikirim" (tetap) berbeda dari "dikirim kosong"
// (lepas parent / urutan 0).
type MenuUpdateInput struct {
	Module    string  `json:"module"`
	Label     string  `json:"label"`
	Kind      string  `json:"kind"`
	SortOrder *int    `json:"sort_order"`
	Parent    *string `json:"parent_code"`
}

// MenuEntry adalah menu milik user (ada grant di CPPERMISSION) untuk sidebar.
// Frontend membandingkan dengan folder registry: yang tidak punya folder
// dirender sebagai halaman 404 pemandu (tahu harus bikin di mana).
// Parent/Kind berasal dari CPMENU sehingga hierarki sidebar mengikuti DB.
type MenuEntry struct {
	Code      string `json:"code"`
	Module    string `json:"module"`
	Label     string `json:"label"`
	Kind      string `json:"kind"`
	SortOrder int    `json:"sort_order"`
	Parent    string `json:"parent_code"`
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
	Kind      string `json:"kind"`
	SortOrder int    `json:"sort_order"`
	Parent    string `json:"parent_code"`
	HasAccess bool   `json:"has_access"`
}
