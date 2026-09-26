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
// Satu baris = satu menu; MODULE = kode modul (FK ke CPMODULE.CODE) dan
// menentukan section sidebar level-1.
// MCONTROL = nama folder frontend (snake_case, mis. user_account) sehingga
// halaman menu tinggal di frontend/src/app/<mcontrol>/index.jsx — datar,
// tanpa folder modul perantara. Judul tampil = LABEL.
// MENU_KIND menentukan peran baris: MenuKindParent (header buka-tutup di
// sidebar, TANPA mcontrol dan TANPA folder) atau MenuKindChild (item biasa,
// mcontrol wajib). PARENT_CODE = kode menu parent (wajib se-modul, dan
// parent-nya harus bertipe PARENT).
type Menu struct {
	ID          int       `json:"-"`
	Code        string    `json:"code"`
	Module      string    `json:"module"`
	ModuleLabel string    `json:"module_label,omitempty"`
	ModuleSort  int       `json:"module_sort,omitempty"`
	Label       string    `json:"label"`
	Mcontrol    string    `json:"mcontrol,omitempty"`
	Kind        string    `json:"kind"`
	SortOrder   int       `json:"sort_order"`
	Parent      string    `json:"parent_code"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// IsParent melaporkan apakah baris ini berperan sebagai menu parent.
func (m Menu) IsParent() bool { return m.Kind == MenuKindParent }

// MenuInput adalah payload POST /api/admin/menus (buat menu).
// Kind opsional; kosong berarti CHILD. Mcontrol wajib untuk CHILD (nama
// folder frontend/src/app/<mcontrol>/, snake_case) dan harus kosong untuk
// PARENT (parent tampil sebagai header buka-tutup, tanpa halaman sendiri).
// Name diterima untuk kompatibilitas API tapi tidak disimpan.
type MenuInput struct {
	Code      string `json:"code"`
	Name      string `json:"name"`
	Module    string `json:"module"`
	Label     string `json:"label"`
	Mcontrol  string `json:"mcontrol"`
	Kind      string `json:"kind"`
	SortOrder int    `json:"sort_order"`
	Parent    string `json:"parent_code"`
}

// MenuUpdateInput adalah payload PUT /api/admin/menus/{code} (patch semantik).
// Field yang tidak dikirim = tidak diubah. Parent, SortOrder, dan Mcontrol
// berupa pointer agar "tidak dikirim" (tetap) berbeda dari "dikirim kosong"
// (lepas parent / urutan 0 / kosongkan mcontrol — khusus PARENT).
type MenuUpdateInput struct {
	Module    string  `json:"module"`
	Label     string  `json:"label"`
	Mcontrol  *string `json:"mcontrol"`
	Kind      string  `json:"kind"`
	SortOrder *int    `json:"sort_order"`
	Parent    *string `json:"parent_code"`
}

// MenuEntry adalah menu milik user (ada grant di CPPERMISSION, plus header
// parent yang menaungi menu ter-grant) untuk sidebar.
// Frontend mencocokkan entri CHILD dengan folder lokal via MCONTROL:
// yang tidak punya folder dirender sebagai halaman 404 pemandu.
// Entri PARENT (kind=PARENT, tanpa mcontrol) tidak punya halaman — hanya
// header buka-tutup. ModuleLabel/ModuleSort dibawa agar sidebar bisa
// dikelompokkan per section modul tanpa request tambahan.
type MenuEntry struct {
	Code        string `json:"code"`
	Module      string `json:"module"`
	ModuleLabel string `json:"module_label,omitempty"`
	ModuleSort  int    `json:"module_sort,omitempty"`
	Label       string `json:"label"`
	Mcontrol    string `json:"mcontrol,omitempty"`
	Kind        string `json:"kind"`
	SortOrder   int    `json:"sort_order"`
	Parent      string `json:"parent_code"`
}

// MatrixRow adalah 1 baris matriks: 1 role x 1 menu + flag akses.
// Dibaca via JOIN (CPROLE x CPMENU LEFT JOIN CPPERMISSION); tulis grant
// tetap ke CPPERMISSION (tidak ada sinkron ganda).
// Hanya baris CHILD yang bisa di-grant (Grantable); baris PARENT adalah header
// buka-tutup: HasAccess-nya diturunkan dari keturunan yang ter-grant.
type MatrixRow struct {
	RoleCode    string `json:"role_code"`
	RoleName    string `json:"role_name"`
	Module      string `json:"module"`
	ModuleLabel string `json:"module_label,omitempty"`
	MenuCode    string `json:"menu_code"`
	MenuLabel   string `json:"menu_label"`
	Mcontrol    string `json:"mcontrol,omitempty"`
	Kind        string `json:"kind"`
	SortOrder   int    `json:"sort_order"`
	Parent      string `json:"parent_code"`
	Grantable   bool   `json:"grantable"`
	HasAccess   bool   `json:"has_access"`
}
