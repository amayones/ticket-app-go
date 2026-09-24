# Tutorial: Tambah Role & Menu Baru

Semua yang perlu diketahui untuk menambah menu yang langsung tampil di
sidebar — khusus admin maupun untuk semua role. File template siap
copy ada di `tutorial/templates/`.

Hanya ada **2 tipe folder menu** (dipindai otomatis oleh
`frontend/src/menus/registry.js`):

| Folder | Dilihat oleh |
|--------|--------------|
| `menus/admin/<menu>/` | role `ADMIN` saja |
| `menus/user/<menu>/` | semua role yang memiliki permission menunya |

> Menu `user/` hanya tampil jika role memiliki permission
> `MENU_<NAMA_MENU>`. **Satu permission menu otomatis memberi akses ke
> seluruh fungsi di menu itu** — tidak ada checkbox/permission per fungsi.
> Permission yang sama dipakai backend pada seluruh route menu sebagai
> `RequirePermission`, sehingga request langsung tanpa akses mendapat `403`.
> Menu `admin/` tetap khusus role `ADMIN`, meskipun role tersebut
> memiliki permission menu yang sama.

Daftar ikon valid: `frontend/src/components/icons.jsx` (objek `PATHS`).

Semua perintah `cp`/`ls` di bawah dijalankan dari **folder root repo**
(`go-core/`), memakai Git Bash (Windows) atau terminal Linux/macOS.
Setiap langkah diakhiri ✅ **checkpoint** — cek dulu sebelum lanjut,
jangan dilewati agar tidak error di langkah berikutnya.

---

## Kasus A — Menu baru khusus ADMIN (5 menit)

Contoh: menu **Laporan** yang hanya boleh dibuka admin.

### A1. Copy template

```bash
cp -r tutorial/templates/frontend-menu frontend/src/menus/admin/laporan
```

✅ **Checkpoint A1** — folder dan 2 file wajib ada:

```bash
ls frontend/src/menus/admin/laporan/
# harus tampil: api.js  index.jsx
```

### A2. Sesuaikan `meta` + isi halaman

Buka `frontend/src/menus/admin/laporan/index.jsx`:

- Ubah `meta` → `{ label: 'Laporan', icon: 'list', order: 8 }`
  (`order` = urutan di sidebar; Dashboard 0, Users 1, … Notifikasi 7).
- Ganti judul, deskripsi, dan isi `<Card>` dengan UI-mu.

✅ **Checkpoint A2** — dua baris ini wajib ada (satu `export default`,
satu `export const meta`):

```bash
grep -n "export default\|export const meta" frontend/src/menus/admin/laporan/index.jsx
# harus tampil 2 baris, mis:
# 22:export const meta = { label: 'Laporan', icon: 'list', order: 8 }
# 24:export default function ...
```

Kalau hanya tampil 1 baris → menu **tidak akan terdaftar** di sidebar.
Perbaiki dulu sebelum lanjut.

### A3. Isi `api.js`

Buka `frontend/src/menus/admin/laporan/api.js`:

- Ganti `<menu>` dengan path endpoint backend-mu (atau endpoint yang
  sudah ada, mis. `/api/admin/audit`).
- Tambah/kurangi fungsi mengikuti kebutuhan halaman.

✅ **Checkpoint A3** — tidak boleh ada sisa placeholder:

```bash
grep -rn "<menu>" frontend/src/menus/admin/laporan/ || echo OK-tidak-ada-placeholder
# harus tampil: OK-tidak-ada-placeholder
```

### A4. Lint + build

```bash
cd frontend
npm run lint
npm run build
cd ..
```

✅ **Checkpoint A4**:

- `npm run lint` → tidak ada baris `error` (warning lama boleh).
- `npm run build` → baris terakhir `✓ built in ...`.

### A5. Cek di browser

1. Login sebagai **admin** → menu **Laporan** ada di sidebar. Selesai —
   tidak ada file lain yang perlu disentuh.
2. Login sebagai **user** biasa → menu **Laporan** tidak ada.

✅ **Checkpoint A5** — admin melihat, user tidak melihat. Kalau user ikut
melihat → folder salah tempat (harus di `menus/admin/`, bukan `menus/user/`).

Hapus menu = hapus foldernya (atau dari git). Menu hilang total dari
bundle setelah `npm run build` berikutnya.

---

## Kasus B — Menu baru untuk SEMUA role + permission (10 menit)

Contoh: menu **Laporan** yang boleh dibuka semua role, tetapi setiap
role hanya melihatnya bila diberi permission `MENU_LAPORAN`.

### B1. Copy template ke `menus/user/`

```bash
cp -r tutorial/templates/frontend-menu frontend/src/menus/user/laporan
```

✅ **Checkpoint B1**:

```bash
ls frontend/src/menus/user/laporan/
# harus tampil: api.js  index.jsx
```

### B2. Sesuaikan `meta` + `api.js`

Sama seperti A2–A3 (meta, isi `<Card>`, ganti endpoint, tanpa sisa
`<menu>`).

✅ **Checkpoint B2** — gabungan A2 + A3 untuk folder `user/laporan`:

```bash
grep -n "export default\|export const meta" frontend/src/menus/user/laporan/index.jsx
grep -rn "<menu>" frontend/src/menus/user/laporan/ || echo OK-tidak-ada-placeholder
# harus tampil 2 baris export + OK-tidak-ada-placeholder
```

### B3. Lint + build

Sama seperti A4 dari folder `frontend/`.

✅ **Checkpoint B3** — `✓ built in ...`, tanpa baris `error` di lint.

### B4. Atur permission menu per role

Setiap menu memiliki **satu checkbox**. Jika dicentang, role tersebut
mendapat seluruh fungsi menu itu; jika tidak dicentang, menu tidak tampil
 dan endpoint menu tersebut mendapat `403`. Tidak perlu membuat permission
terpisah untuk create, update, delete, atau aksi lain.

**Cara 1 — via UI (tanpa SQL):**

1. Login admin → **Role & Permission** → pilih role di sebelah kiri →
   centang `MENU_LAPORAN` di matriks sebelah kanan → **Simpan**.
   Tambahkan konstanta `MenuLaporan = "MENU_LAPORAN"` di
   `models/permission.go` dan seed `CPPERMISSION` bila menu baru.
2. Ulangi untuk setiap role yang boleh / tidak boleh akses.

**Cara 2 — via SQL** (edit dulu `tutorial/templates/new-role.sql`
sesuai kebutuhan, lalu jalankan di DB-mu):

```bash
# contoh SQL Server:
sqlcmd -S localhost,1433 -U <user> -P <pass> -d Go -C -i tutorial/templates/new-role.sql
```

✅ **Checkpoint B4** — permission tercatat di DB:

```sql
SELECT ROLE_CODE, PERMISSION_CODE FROM CPROLEPERMISSION
WHERE PERMISSION_CODE = 'MENU_LAPORAN' ORDER BY ROLE_CODE;
-- harus tampil baris untuk role yang diberi hak, dan TIDAK ada
-- baris untuk role yang tidak diberi hak
```

### B5. Cek di browser + API

1. Login sebagai user yang **punya** permission → menu **Laporan** tampil
   dan data termuat.
2. Login sebagai user yang **tidak punya** permission → menu **Laporan**
   tidak tampil di sidebar. Request endpoint yang dipanggil langsung
   tetap mendapat `403` dari backend.

✅ **Checkpoint B5** — via `curl` (ganti `<token>` dengan access token
masing-masing user, lihat README utama Bagian 1 Langkah 5 cara login):

```bash
curl http://localhost:1067/api/admin/laporan -H "Authorization: Bearer <token-punya-hak>"
# -> 200 + data
curl http://localhost:1067/api/admin/laporan -H "Authorization: Bearer <token-tanpa-hak>"
# -> 403 {"error":"forbidden: missing MENU_LAPORAN"}
```

---

## Kasus C — Role baru (mis. `EDITOR`)

Role baru **tidak butuh folder menu sendiri** — menu `user/` otomatis
terlihat oleh role apa pun. Yang perlu disiapkan hanya role + permission.

### C1. Buat role + permission di database

Edit `tutorial/templates/new-role.sql` (ganti `EDITOR` + daftar permission),
lalu jalankan di DB-mu.

✅ **Checkpoint C1**:

```sql
SELECT CODE FROM CPROLE WHERE CODE = 'EDITOR';
-- harus tampil 1 baris: EDITOR
SELECT PERMISSION_CODE FROM CPROLEPERMISSION WHERE ROLE_CODE = 'EDITOR';
-- harus tampil sesuai daftar yang kamu isi (min. 1 baris)
```

(Alternatif tanpa SQL: admin → **Role & Permission** → **Role baru**
(kode huruf besar) → centang permission → **Simpan**.)

### C2. Pindahkan user ke role baru

Via UI: User Account → pensil → dropdown Role. Atau via SQL:

```sql
UPDATE CPUSER SET ROLE_CODE = 'EDITOR' WHERE CODE = 'USR-XXXXXX';
SELECT USERNAME, ROLE_CODE FROM CPUSER WHERE CODE = 'USR-XXXXXX';
-- ROLE_CODE harus sudah EDITOR
```

User harus **login ulang** agar klaim role di JWT terbarui.

### C3. Cek di browser

Login sebagai user tersebut.

✅ **Checkpoint C3**:

- Menu `user/` (Dashboard + menu user lain) tampil.
- Menu `admin/` (User Account, Role & Permission, …) **tidak** tampil,
  kecuali role baru itu memang `ADMIN`.

---

## Backend untuk menu baru (bila butuh endpoint sendiri)

Satu menu = satu folder `features/<menu>/` berisi 3 file + daftarkan 1 baris.
Konvensi nama (sama untuk semua menu): `Repository`/`Service`/`Handler` +
`NewRepository`/`NewService`/`NewHandler`, method repo
`List/GetByX/Create/Update/Delete/Count`, SQL selalu di variabel `query`.

> Menu `user/` yang datanya sensitif **wajib** dipasang permission di route
> (langkah 4 di bawah). Menu `admin/` hidup di dalam blok `/admin` yang
> memang hanya untuk admin.

**1. `features/laporan/laporan_repository.go`** — raw SQL full:

```go
package laporan

import (
	"context"
	"database/sql"

	"golang-backend/models"
	"golang-backend/repositories"
)

type Item struct {
	Code      string `json:"code"`
	Name      string `json:"name"`
	CreatedAt string `json:"created_at"`
}

type RepositoryInterface interface {
	List(ctx context.Context, limit, offset int) ([]Item, error)
}

type Repository struct {
	db      *sql.DB
	dialect repositories.Dialect
}

func NewRepository(db *sql.DB, dialect repositories.Dialect) RepositoryInterface {
	return &Repository{db: db, dialect: dialect}
}

func (r *Repository) List(ctx context.Context, limit, offset int) ([]Item, error) {
	ctx, cancel := repositories.WithTimeout(ctx)
	defer cancel()
	// 1. SQL full di variabel. 2. Bind (? -> @pN/$N). 3. Run. 4. Petakan.
	query := `SELECT CODE, NAME, CREATED_AT FROM ` + r.dialect.Table("CPLAPORAN") + ` ORDER BY ID DESC `
	var args []any
	if r.dialect == repositories.DialectMSSQL {
		query += `OFFSET ? ROWS FETCH NEXT ? ROWS ONLY`
		args = []any{offset, limit}
	} else {
		query += `LIMIT ? OFFSET ?`
		args = []any{limit, offset}
	}
	rows, err := r.db.QueryContext(ctx, r.dialect.Bind(query), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]Item, 0)
	for rows.Next() {
		var it Item
		if err := rows.Scan(&it.Code, &it.Name, &it.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, it)
	}
	return out, rows.Err()
}
```

✅ **Checkpoint backend-1** — nama package = nama folder, ada
`NewRepository`, SQL di variabel `query`:

```bash
grep -n "^package \|func NewRepository\|query :=" features/laporan/laporan_repository.go
# harus tampil 3 baris
```

**2. `features/laporan/laporan_service.go`** — aturan bisnis:

```go
package laporan

import (
	"context"
	"fmt"
)

type ServiceInterface interface {
	ListItems(ctx context.Context, limit, offset int) ([]Item, error)
}

type Service struct {
	repo RepositoryInterface
}

func NewService(repo RepositoryInterface) *Service {
	return &Service{repo: repo}
}

func (s *Service) ListItems(ctx context.Context, limit, offset int) ([]Item, error) {
	items, err := s.repo.List(ctx, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list laporan: %w", err)
	}
	return items, nil
}
```

✅ **Checkpoint backend-2**:

```bash
grep -n "^package \|func NewService" features/laporan/laporan_service.go
# harus tampil 2 baris
```

**3. `features/laporan/laporan_handler.go`** — HTTP JSON:

```go
package laporan

import (
	"net/http"

	"golang-backend/internal/web"
)

type Handler struct {
	Service ServiceInterface
}

func NewHandler(service ServiceInterface) *Handler {
	return &Handler{Service: service}
}

func (h *Handler) ListLaporan(w http.ResponseWriter, r *http.Request) {
	limit, offset := web.Paginate(r)
	items, err := h.Service.ListItems(r.Context(), limit, offset)
	if err != nil {
		web.ServiceError(w, err)
		return
	}
	web.WriteJSON(w, http.StatusOK, items)
}
```

✅ **Checkpoint backend-3**:

```bash
grep -n "^package \|func NewHandler\|func (h \*Handler)" features/laporan/laporan_handler.go
# harus tampil 3 baris
```

**4. Daftarkan (2 tempat, masing-masing 1–3 baris):**

- `main.go` (ikuti blok `// Wiring per menu`):
  ```go
  laporanRepo := laporan.NewRepository(db, dialect)
  laporanSvc := laporan.NewService(laporanRepo)
  ```
  tambah ke `deps := routes.Deps{ ... Laporan: laporan.NewHandler(laporanSvc), }`
  (tambah juga field `Laporan` di struct `routes.Deps` di `routes/routes.go`).
- `routes/routes.go` — untuk menu semua-role, pasang permission agar
  batas antar-role bekerja (tambah permission dulu bila perlu di
  `models/permission.go` + seed `CPPERMISSION`):
  ```go
  r.With(auth, need(models.MenuLaporan)).Get("/laporan", deps.Laporan.ListLaporan)
  ```
  Untuk menu khusus admin, taruh di dalam blok `r.Route("/admin", …)`
  seperti endpoint admin yang sudah ada.

✅ **Checkpoint backend-4** — kompilasi lolos:

```bash
go vet ./...
# harus selesai tanpa output (tanpa error)
```

**5. Buat tabelnya** (contoh SQL Server; sesuaikan untuk postgres/sqlite
mengikuti `scripts/schema.*.sql`):

```sql
CREATE TABLE CPLAPORAN (
  ID INT IDENTITY(1,1) PRIMARY KEY,
  CODE NVARCHAR(20) NOT NULL UNIQUE,
  NAME NVARCHAR(100) NOT NULL,
  CREATED_AT DATETIME NOT NULL DEFAULT GETDATE()
);
```

**6. Verifikasi backend:**

```bash
go test -race ./...
curl http://localhost:1067/api/laporan -H "Authorization: Bearer <token-punya-hak>"
# -> 200 + data
curl http://localhost:1067/api/laporan -H "Authorization: Bearer <token-tanpa-hak>"
# -> 403
```

✅ **Checkpoint backend-5** — `go test` hijau (semua `ok`), `curl` pertama
200 dan kedua 403.

## Checklist akhir (frontend + backend)

- [ ] Folder di `menus/admin/` (khusus ADMIN) atau `menus/user/` (semua
  role) — cek tabel 2 tipe di atas. Tidak ada lagi folder `shared/` atau
  per-role custom.
- [ ] `index.jsx` punya `export default` + `export const meta`
  (cek via `grep` seperti checkpoint A2).
- [ ] `api.js` tanpa sisa `<menu>` (cek via `grep` seperti checkpoint A3).
- [ ] Menu `user/` yang sensitif: endpoint diproteksi permission
  (cek via `curl` 200 vs 403 seperti checkpoint B5).
- [ ] `lint` + `build` hijau; menu tampil setelah login role yang tepat
  (admin melihat semua; non-admin hanya menu `user/`).
- [ ] Backend: `vet` + `test` hijau; endpoint 200 untuk pemilik permission,
  403 untuk yang tidak.
- [ ] Tulis audit log bila aksinya penting (contoh di handler lain:
  `h.Audit.Log(...)`).
