# Tutorial: Tambah Role & Menu Baru

Semua yang perlu diketahui untuk menambah menu yang langsung tampil di
sidebar — untuk role yang sudah ada maupun role baru. File template siap
copy ada di `tutorial/templates/`.

Aturan folder (dipindai otomatis oleh `frontend/src/menus/registry.js`):

| Folder | Dilihat oleh |
|--------|--------------|
| `menus/shared/<menu>/` | SEMUA role |
| `menus/admin/<menu>/` | role `ADMIN` |
| `menus/user/<menu>/` | role `USER` |
| `menus/<nama-role-lowercase>/<menu>/` | role custom (mis. `menus/editor/` untuk `EDITOR`) |

Daftar ikon valid: `frontend/src/components/icons.jsx` (objek `PATHS`).

---

## Kasus A — Menu baru di role yang sudah ada (5 menit)

Contoh: menu **Laporan** khusus admin.

1. Copy template:
   ```bash
   cp -r tutorial/templates/frontend-menu frontend/src/menus/admin/laporan
   ```
2. Buka `frontend/src/menus/admin/laporan/index.jsx`:
   - Ubah `meta` → `{ label: 'Laporan', icon: 'list', order: 8 }`
     (`order` = urutan di sidebar; Dashboard 0, Users 1, … Notifikasi 7).
   - Ganti judul, deskripsi, dan isi `<Card>` dengan UI-mu.
3. Buka `frontend/src/menus/admin/laporan/api.js`:
   - Ganti `<menu>` dengan path endpoint backend-mu (atau endpoint yang
     sudah ada, mis. `/api/admin/audit`).
   - Tambah/kurangi fungsi mengikuti kebutuhan halaman.
4. Verifikasi:
   ```bash
   npm run lint   # dari folder frontend/
   npm run build
   ```
5. Login sebagai admin → menu **Laporan** sudah ada di sidebar. Selesai —
   tidak ada file lain yang perlu disentuh.

Hapus menu = hapus foldernya (atau dari git). Menu hilang total dari
bundle setelah `npm run build` berikutnya.

## Kasus B — Role baru + menu (10 menit)

Contoh: role `EDITOR` dengan halaman sendiri.

1. Buat role + permission di database (edit dulu file template ini):
   `tutorial/templates/new-role.sql` → jalankan di DB-mu.
2. Buat folder menu untuk role itu (nama folder role = lowercase):
   ```bash
   cp -r tutorial/templates/frontend-menu frontend/src/menus/editor/beranda
   ```
   Sesuaikan `meta` seperti Kasus A. Role `EDITOR` otomatis melihat menu
   `shared/` + `editor/`. Role lain tidak terpengaruh.
3. Pindahkan user ke role baru (via UI: User Account → pensil → dropdown
   Role, atau SQL `UPDATE CPUSER SET ROLE_CODE='EDITOR' WHERE CODE='…'`).
   User harus **login ulang** agar klaim role di JWT terbarui.
4. Verifikasi seperti Kasus A + login sebagai user tersebut.

## Backend untuk menu baru (bila butuh endpoint sendiri)

Satu menu = satu folder `features/<menu>/` berisi 3 file + daftarkan 1 baris.
Konvensi nama (sama untuk semua menu): `Repository`/`Service`/`Handler` +
`NewRepository`/`NewService`/`NewHandler`, method repo
`List/GetByX/Create/Update/Delete/Count`, SQL selalu di variabel `query`.

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

**4. Daftarkan (2 tempat, masing-masing 1–3 baris):**

- `main.go` (ikuti blok `// Wiring per menu`):
  ```go
  laporanRepo := laporan.NewRepository(db, dialect)
  laporanSvc := laporan.NewService(laporanRepo)
  ```
  tambah ke `deps := routes.Deps{ ... Laporan: laporan.NewHandler(laporanSvc), }`
  (tambah juga field `Laporan` di struct `routes.Deps` di `routes/routes.go`).
- `routes/routes.go` (di dalam `r.Route("/admin", …)`, tambah permission
  dulu bila perlu di `models/permission.go` + seed `CPPERMISSION`):
  ```go
  r.With(auth, need(models.PermLaporanRead)).Get("/laporan", deps.Laporan.ListLaporan)
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

**6. Verifikasi backend:** `go vet ./...`, `go test -race ./...`,
`curl` endpoint dengan token admin (200) dan user biasa (403 bila
diproteksi permission).

## Checklist akhir (frontend + backend)

- [ ] Folder di `menus/<role>/` yang benar (cek tabel aturan di atas).
- [ ] `index.jsx` punya `export default` + `export const meta`.
- [ ] `api.js` hanya memanggil endpoint yang ada (cek 404 di console browser).
- [ ] `lint` + `build` hijau; menu tampil setelah login role yang tepat,
  tidak tampil di role lain.
- [ ] Backend: `vet` + `test` hijau; endpoint 200 untuk pemilik permission,
  403 untuk yang tidak.
- [ ] Tulis audit log bila aksinya penting (contoh di handler lain:
  `h.Audit.Log(...)`).
