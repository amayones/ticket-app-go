# Tutorial: Menambah Modul, Menu, Role, dan User

Tutorial ini berjalan dari nol: buat **modul**, buat **menu** di dalamnya,
isi halamannya, lalu beri **role** akses ke menu itu. Admin tidak perlu
membuat folder atau kode khusus untuk setiap role.

Setelah menu dibuat, admin hanya perlu memberi centang akses pada role yang
diizinkan. Setelah user login ulang, menu otomatis muncul di sidebar.

## Resep 5 Menit (ringkasan)

Contoh: modul `TOKO`, menu `Stok`, role `EDITOR`, user `editor`.

| # | Apa yang dilakukan | Di mana | Hasil |
|---|---|---|---|
| 1 | Buat modul `TOKO` | UI **Modul & Menu** → **Modul baru** | Baris `TOKO` di `CPMATRIX` |
| 2 | Buat menu `MENU_STOK` (modul `TOKO`) | UI **Modul & Menu** → **Menu baru** | Baris `MENU_STOK` di `CPMENU`, belum ada grant |
| 3 | Siapkan halaman | `cp -r tutorial/templates/frontend-menu frontend/src/menus/TOKO/stok` | Folder menu + `index.jsx` + `api.js` |
| 4 | Isi halaman | `index.jsx` + `api.js` | Tabel stok (list, tambah, ubah, hapus) |
| 5 | (opsional) Endpoint backend | `features/stok/` + `routes/routes.go` | `GET/POST/PUT/DELETE /api/stok` |
| 6 | Build | `npm run build` + restart backend | Sidebar punya grup `TOKO` |
| 7 | Buat role `EDITOR` | UI **Role & Permission** → **Role baru** | Baris `EDITOR` di `CPROLE` |
| 8 | Beri akses | UI **Role & Permission** → centang `MENU_STOK` → **Simpan** | Baris di `CPPERMISSION` |
| 9 | Buat user | UI **User Account** → **Tambah user** (role `EDITOR`) | User bisa login, menu Stok terlihat |

Langkah 1–4 wajib. Langkah 5 boleh dilewati bila memakai endpoint yang sudah
ada. Detail tiap langkah ada di bawah.

## Aturan Utama

Struktur folder menu (modul UPPERCASE, bebas tambah modul baru):

```text
frontend/src/menus/<MCONTROL>/[<grup>/]<menu>/  -> semua role yang diberi akses menu
```

Contoh bawaan:

```text
menus/SYSTEM/users/         -> modul SYSTEM, menu users
menus/REPORT/keu/arus-kas/  -> modul REPORT, grup visual "keu", menu arus-kas
menus/REPORT/laporan/       -> modul REPORT, menu laporan (tanpa grup)
```

Modul = kolom `MCONTROL` tabel `CPMENU`, wajib sama persis dengan nama
folder (UPPERCASE). Folder perantara tanpa `index.jsx` = grup visual saja
(bisa dibuka-tutup di sidebar dengan tombol +/−), **bukan** menu tersendiri.

Tutorial ini memakai modul baru `TOKO` dan menu `stok`.

Satu permission mewakili satu menu:

```text
MENU_<NAMA_MENU>
```

Contoh folder `stok` otomatis menggunakan permission:

```text
MENU_STOK
```

Satu checkbox `MENU_STOK` memberi akses ke **seluruh fungsi** menu
Stok. Tidak ada permission terpisah untuk list, create, edit, atau delete.

Tiga aturan yang paling sering membuat orang tersesat:

1. Kode menu harus persis `MENU_` + nama folder menu dalam huruf besar.
   `stok` → `MENU_STOK`, `arus-kas` → `MENU_ARUS_KAS`.
2. Nama folder modul harus sama dengan `MCONTROL` dan huruf besar.
   `TOKO` ↔ `menus/TOKO/`.
3. Menu baru **tidak** otomatis bisa diakses siapa pun. Sampai admin mencentang
   di halaman Role, menu itu tidak muncul di sidebar siapa pun.

## Istilah Penting

| Istilah | Arti |
|---|---|
| Modul | Kelompok menu di sidebar, mis. `SYSTEM`, `REPORT`. Disimpan di `CPMATRIX`. |
| Menu | Satu halaman di sidebar, mis. User Account atau Stok. Disimpan di `CPMENU`. |
| Grup visual | Folder perantara untuk mengelompokkan menu (tombol +/−), bukan menu. |
| Role | Kelompok user, misalnya ADMIN, USER, EDITOR. Disimpan di `CPROLE`. |
| Permission | Akses ke satu menu, misalnya `MENU_STOK`. |
| Grant | Baris `CPPERMISSION` yang memberi role akses ke menu. |
| Checkbox | Centang di halaman Role & Permission. |
| Sidebar | Daftar menu yang otomatis dibuat dari folder menu. |

Semua perintah `cp`, `ls`, dan `grep` dijalankan dari folder root
repository `go-core` menggunakan Git Bash.

---

# Bagian 0 — Nama Database dan Nama Aplikasi

Clone baru selalu mulai dari sini. Tidak ada kode yang perlu diedit untuk
mengganti nama database.

## 0.1. Nama database bebas

Buat database dengan nama apa pun, lalu samakan di `.env`:

```sql
IF DB_ID(N'TokoDb') IS NULL
BEGIN
  CREATE DATABASE [TokoDb];
END
GO
```

```env
DB_CONNECTION=sqlserver
DB_HOST=localhost
DB_PORT=1433
DB_DATABASE=TokoDb
DB_USERNAME=<user-sqlserver-anda>
DB_PASSWORD=<password-anda>
```

Lalu jalankan migrasi ke database **itu** (nama di `-d` harus sama):

```bash
task migrate
sqlcmd -S localhost,1433 -U <user> -P "<password>" -d TokoDb -C -i scripts/seed-admin.sql
```

Semua query aplikasi memakai nama tabel `CP*` tanpa prefix database, jadi
nama database tidak pernah muncul di kode.

## 0.2. Mengganti nama aplikasi

Cukup ubah satu baris di `.env`, lalu build ulang frontend:

```env
APP_NAME=Toko Saya
```

Ikut berubah: judul tab browser, nama di sidebar, nama di kartu login, dan
teks footer.

Logo/favicon diatur terpisah lewat `APP_LOGO` di `.env` yang menunjuk file di
`frontend/public/`:

```env
APP_LOGO=/favicon.svg
```

Ganti `frontend/public/favicon.svg` dengan logo Anda (rasio 1:1), atau simpan
file lain di folder yang sama lalu arahkan `APP_LOGO` ke sana. Favicon tab,
logo sidebar, dan logo kartu login memakai file yang sama. Yang **tidak**
ikut berubah: `frontend/index.html` (`<title>` fallback) dan isi email di
`CPNOTIFTEMPLATE` (bisa diedit dari menu **Notifikasi**).

Detail lengkap ada di README bagian
[1.9. Mengganti Nama Aplikasi](../README.md#19-mengganti-nama-aplikasi).

## 0.3. Engine lain (opsional)

PostgreSQL: `DB_CONNECTION=postgres` lalu `task migrate-postgres`
(skema `scripts/schema.postgres.sql`).

SQLite: `DB_CONNECTION=sqlite` + `DB_DATABASE=./data/tokodb.db` lalu
`task migrate-sqlite` (skema `scripts/schema.sqlite.sql`).

Ketiganya menghasilkan state awal yang sama: 10 tabel, 2 modul, 10 menu,
`ADMIN` 8 grant, 2 akun (`admin`/`admin` dan `user`/`user`), 3 template
notifikasi.

---

# Bagian 1 — Daftarkan Modul dan Menu di Database

Urutannya wajib: **modul dulu, baru menu** — `CPMENU.MCONTROL` menunjuk ke
`CPMATRIX.CODE`.

## 1.1. Buat modul baru

Lewati bagian ini bila modul yang Anda butuh sudah ada (`SYSTEM` dan
`REPORT` bawaan).

via UI (disarankan):

1. Login `admin`, buka menu **Modul & Menu**.
2. Klik **Modul baru**.
3. Isi: kode modul `TOKO` (harus huruf besar, maks 40 karakter), label
   `Toko`, urutan `3`.
4. Simpan.

via SQL:

```sql
INSERT INTO dbo.CPMATRIX (CODE, LABEL, SORT_ORDER)
SELECT N'TOKO', N'Toko', 3
WHERE NOT EXISTS (SELECT 1 FROM dbo.CPMATRIX WHERE CODE = N'TOKO');
```

Modul gagal dihapus selama masih ada menu di dalamnya (UI memberi pesan
`409 Module is still used by menus`).

✅ **Checkpoint 1.1**

```sql
SELECT CODE, LABEL, SORT_ORDER FROM dbo.CPMATRIX ORDER BY SORT_ORDER;
```

`TOKO` harus muncul di antara `SYSTEM` dan `REPORT`.

## 1.2. Buat menu di modul itu

via UI:

1. Login `admin`, buka menu **Modul & Menu**.
2. Klik **Menu baru**, isi:
   - Kode permission: `MENU_STOK` (wajib prefix `MENU_`, maks 40 karakter)
   - Modul (MCONTROL): pilih `TOKO` dari dropdown
   - Label tampil: `Stok`, Urutan: `1`, Parent: kosongkan
3. Simpan. Menu tercatat di `CPMENU`
   **tanpa auto-grant ke role mana pun** — admin mencentang manual di
   halaman Role (Bagian 5).

Menu yang sudah ada bisa diubah kapan saja: di daftar menu, klik ikon
pensil untuk mengubah **label tampil**, **urutan**, **modul**, atau
**parent**. Kode permission (`MENU_STOK`) tidak bisa diubah karena jadi
acuan folder `menus/TOKO/stok/` dan grant role — untuk mengganti kode,
buat menu baru lalu hapus yang lama.

✅ **Checkpoint 1.2**

```sql
SELECT CODE, MCONTROL, LABEL FROM dbo.CPMENU WHERE CODE = 'MENU_STOK';
```

Harus menghasilkan satu baris dengan `MCONTROL = TOKO`.

## 1.3. Alternatif via SQL

Satu permission mewakili satu menu (`MENU_<NAMA_MENU>`); modul = nama
folder (`TOKO` → `menus/TOKO/`). Cukup satu INSERT:

```sql
INSERT INTO dbo.CPMENU (CODE, MCONTROL, LABEL, SORT_ORDER)
VALUES (N'MENU_STOK', N'TOKO', N'Stok', 1);
```

Grant ke role tetap via matriks UI (atau `INSERT INTO dbo.CPPERMISSION
(ROLE_CODE, MENU_CODE) ...`).

✅ **Checkpoint 1.3**

```sql
SELECT CODE, MCONTROL FROM dbo.CPMENU WHERE CODE = 'MENU_STOK';
```

Harus menghasilkan satu baris. (Jalur UI di 1.2 tidak menulis seed file —
cek ke database, bukan ke `migrate2_rbac.sql`.)

Untuk menu permanen bawaan, tambahkan seed yang sama di
`scripts/migrate2_rbac.sql` (proyek ini khusus SQL Server). Menu yang dibuat
lewat UI tidak perlu seed.

> Catatan: kolom `Parent (opsional)` di form menu biasanya **dikosongkan**.
> Grup visual di sidebar datang dari folder perantara tanpa `index.jsx`,
> bukan dari menu induk.

---

# Bagian 2 — Membuat Frontend Menu

Folder menu wajib huruf besar untuk segmen modul, dan nama folder terakhir
menentukan permission.

## 2.1. Copy Template

```bash
cp -r tutorial/templates/frontend-menu frontend/src/menus/TOKO/stok
```

✅ **Checkpoint 2.1**

```bash
ls frontend/src/menus/TOKO/stok
```

Harus ada:

```text
api.js
index.jsx
```

## 2.2. Isi `index.jsx`

Buka `frontend/src/menus/TOKO/stok/index.jsx`.

Template sudah punya `export const meta` di atas komponen. Sesuaikan label,
ikon, dan urutan:

```jsx
export const meta = { label: 'Stok', icon: 'list', order: 1 }
```

`icon` memakai nama ikon dari `components/icons.jsx` (`list`, `users`,
`shield`, `bell`, `key`, `terminal`, ...). `order` menentukan urutan dalam
sidebar modul (makin kecil makin atas).

Isi halaman mengikuti pola menu tabel yang sudah ada
(`menus/SYSTEM/syslog/index.jsx` adalah contoh paling ringkas):

```jsx
import { useCallback, useEffect, useState } from 'react'
import { deleteStok, listStok, saveStok } from './api.js'
import {
  Alert,
  Button,
  Card,
  CardTitle,
  ConfirmDialog,
  EmptyState,
  Icon,
  Pagination,
  SkeletonRows,
  TextField,
  useSmoothLoading,
  useToast,
} from '../../../components'

const PAGE_SIZE = 20

export const meta = { label: 'Stok', icon: 'list', order: 1 }

export default function Stok() {
  const toast = useToast()
  const [rows, setRows] = useState([])
  const [loading, setLoading] = useState(true)
  const showLoading = useSmoothLoading(loading)
  const [error, setError] = useState('')
  const [offset, setOffset] = useState(0)
  const [saving, setSaving] = useState(false)
  const [deleting, setDeleting] = useState(null)
  const [form, setForm] = useState({ code: '', name: '', qty: '' })

  const load = useCallback(async () => {
    setLoading(true)
    setError('')
    try {
      setRows(await listStok(PAGE_SIZE, offset))
    } catch (err) {
      setError(err.message)
    } finally {
      setLoading(false)
    }
  }, [offset])

  useEffect(() => {
    load()
  }, [load])

  async function save() {
    setSaving(true)
    try {
      await saveStok(form)
      toast.success('Stok disimpan.')
      setForm({ code: '', name: '', qty: '' })
      load()
    } catch (err) {
      toast.error(err.message, { title: 'Gagal menyimpan' })
    } finally {
      setSaving(false)
    }
  }

  async function remove() {
    if (!deleting) return
    try {
      await deleteStok(deleting.code)
      toast.success(`Stok ${deleting.code} dihapus.`)
      setDeleting(null)
      load()
    } catch (err) {
      toast.error(err.message, { title: 'Gagal menghapus' })
    }
  }

  return (
    <div className="flex flex-col gap-4">
      <Card>
        <div className="mb-4 flex flex-wrap items-center justify-between gap-3">
          <CardTitle description="Contoh menu CRUD di modul TOKO.">Stok</CardTitle>
          <Button size="sm" onClick={save} loading={saving}>Simpan stok</Button>
        </div>

        {error && (
          <Alert tone="error" title="Gagal memuat" closable onClose={() => setError('')} className="mb-4">
            {error}
          </Alert>
        )}

        {showLoading ? (
          <SkeletonRows rows={4} />
        ) : rows.length === 0 ? (
          <EmptyState title="Belum ada data" description="Tambahkan stok pertama lewat form di atas." />
        ) : (
          <table className="w-full text-left text-sm">
            <thead className="text-xs uppercase tracking-wide text-zinc-500">
              <tr>
                <th className="px-3 py-2">Kode</th>
                <th className="px-3 py-2">Nama</th>
                <th className="px-3 py-2">Qty</th>
                <th className="px-3 py-2" />
              </tr>
            </thead>
            <tbody>
              {rows.map((r) => (
                <tr key={r.code} className="border-t border-zinc-100 dark:border-zinc-800">
                  <td className="px-3 py-2.5 font-mono">{r.code}</td>
                  <td className="px-3 py-2.5">{r.name}</td>
                  <td className="px-3 py-2.5">{r.qty}</td>
                  <td className="px-3 py-2.5 text-right">
                    <Button variant="danger" size="sm" onClick={() => setDeleting(r)}>
                      <Icon name="trash" className="h-4 w-4" />
                      Hapus
                    </Button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}

        <div className="mt-4 flex flex-col gap-3 sm:flex-row">
          <TextField
            label="Kode"
            value={form.code}
            onChange={(e) => setForm({ ...form, code: e.target.value.toUpperCase() })}
            placeholder="mis. STK-001"
          />
          <TextField
            label="Nama"
            value={form.name}
            onChange={(e) => setForm({ ...form, name: e.target.value })}
            placeholder="mis. Kaos polos"
          />
          <TextField
            label="Qty"
            value={form.qty}
            onChange={(e) => setForm({ ...form, qty: e.target.value })}
            placeholder="0"
          />
        </div>
      </Card>

      <Pagination
        offset={offset}
        limit={PAGE_SIZE}
        count={rows.length}
        hasMore={rows.length === PAGE_SIZE}
        loading={showLoading}
        onPage={setOffset}
      />

      <ConfirmDialog
        open={!!deleting}
        title={`Hapus stok ${deleting?.code}?`}
        message="Data stok akan dihapus permanen."
        confirmLabel="Ya, hapus"
        danger
        onConfirm={remove}
        onCancel={() => setDeleting(null)}
      />
    </div>
  )
}
```

Pola yang dijaga di semua menu:

| Urutan | Isi |
|---|---|
| 1 | `import` (react, `./api.js`, lalu komponen dari barrel `../../../components`) |
| 2 | `export const meta` (wajib: label, icon, order) |
| 3 | `export default function <Nama>()` |
| 4 | State: `loading` + `useSmoothLoading`, `error`, lalu state domain |
| 5 | `load` dengan `useCallback` + `useEffect(() => load(), [load])` |
| 6 | Tampilan: `Card` + `CardTitle`, `Alert` error, `SkeletonRows`, tabel, `EmptyState`, `Pagination` |
| 7 | `ConfirmDialog` untuk aksi hapus |

Semua komponen UI diambil dari barrel `components/index.js` — jangan
meng-`import` file komponen satu per satu.

✅ **Checkpoint 2.2**

```bash
grep -n "export default\|export const meta" frontend/src/menus/TOKO/stok/index.jsx
```

Harus ada dua baris export:

```text
export const meta = ...
export default function ...
```

## 2.3. Isi `api.js`

Buka `frontend/src/menus/TOKO/stok/api.js`.

Ganti placeholder `<menu>` dengan endpoint yang benar. Pola file ini: satu
fungsi per aksi, `auth: true` untuk endpoint privat, dan list dinormalisasi
ke array.

```js
// Fungsi menu Stok (modul TOKO, permission MENU_STOK).
import { apiRequest as request } from '../../../api/client.js'

// GET list -> selalu kembalikan array (paginasi lewat ?limit=&offset=).
export async function listStok(limit = 20, offset = 0) {
  const data = await request(`/api/stok?limit=${limit}&offset=${offset}`, { auth: true })
  return Array.isArray(data) ? data : []
}

// POST/PUT satu data (PATCH parsial: kirim hanya field yang berubah).
export async function saveStok(payload) {
  return request('/api/stok', { method: 'POST', body: payload, auth: true })
}

export async function deleteStok(code) {
  return request(`/api/stok/${code}`, { method: 'DELETE', auth: true })
}
```

Kalau endpoint yang Anda butuh **sudah ada** di backend, jangan buat
endpoint baru — panggil saja dari `api.js` menu ini:

```js
// Menu ini memakai endpoint yang sudah ada (daftar template notifikasi).
import { apiRequest as request } from '../../../api/client.js'

export async function listTemplateNotif() {
  const data = await request('/api/admin/notifications/templates', { auth: true })
  return Array.isArray(data) ? data : []
}
```

Tetap jadikan `api.js` milik menu sendiri: satu menu satu file, supaya
independen saat backend berubah.

✅ **Checkpoint 2.3**

```bash
grep -rn "<menu>" frontend/src/menus/TOKO/stok || echo OK-tidak-ada-placeholder
```

Harus menghasilkan:

```text
OK-tidak-ada-placeholder
```

## 2.4. Pastikan Permission Mapping Otomatis

Frontend otomatis mengubah nama folder menjadi permission:

```text
frontend/src/menus/TOKO/stok/
              ↓
MENU_STOK   (modul TOKO dari nama folder)
```

Folder perantara tanpa `index.jsx` = grup visual, mis.
`menus/TOKO/gudang/stok/` tetap memakai permission `MENU_STOK` di bawah grup
`gudang` (bisa dibuka-tutup di sidebar).

Jadi tidak perlu menambah daftar menu manual di `registry.js`.

✅ **Checkpoint 2.4**

```bash
grep -n "import.meta.glob" frontend/src/menus/registry.js
```

Pastikan glob `./*/**/index.jsx` memindai semua modul (termasuk modul baru
seperti `TOKO`).

## 2.5. Pahami Alur 404 Pemandu

Urutan yang benar: **daftar di database dulu (Bagian 1), folder belakangan
(Bagian 2)**. Di jeda itu, bila menu sudah dicentang ke suatu role, user
role tersebut melihat halaman **404 pemandu** (bukan menu hilang diam-diam):

- Permission yang hilang + path folder persis yang harus dibuat, mis.
  `frontend/src/menus/TOKO/stok/`
- Perintah copy template siap salin
- Checklist rebuild (`npm run build` + restart backend)

Setelah folder dibuat + rebuild, menu asli otomatis menggantikan halaman itu.
Gunakan halaman 404 sebagai kompas: ia selalu menunjuk lokasi yang benar.
---

# Bagian 3 — Menambahkan Backend Jika Dibutuhkan

Jika menu hanya memakai endpoint yang sudah ada, **lewati bagian ini** dan
langsung ke Bagian 4.

Contoh di bawah memakai menu `Stok` (modul `TOKO`, permission `MENU_STOK`).

## 3.1. Tabel di database

Buat tabel dengan awalan `CP` seperti tabel lain, lalu daftarkan di
`scripts/migrate2_rbac.sql` (atau `schema.postgres.sql` / `schema.sqlite.sql`
bila memakai engine tersebut):

```sql
IF OBJECT_ID(N'dbo.CPSTOK', N'U') IS NULL
BEGIN
  CREATE TABLE dbo.CPSTOK (
    ID        INT IDENTITY(1,1) PRIMARY KEY,
    CODE      NVARCHAR(20)  NOT NULL UNIQUE,
    NAME      NVARCHAR(100) NOT NULL,
    QTY       INT           NOT NULL DEFAULT 0,
    CREATED_AT DATETIME2(0) NOT NULL DEFAULT SYSDATETIME(),
    UPDATED_AT DATETIME2(0) NOT NULL DEFAULT SYSDATETIME()
  );
END
GO
```

## 3.2. Model + konstanta permission

Tambah model di `models/stok.go`:

```go
package models

type Stok struct {
    ID        int       `json:"-"`
    Code      string    `json:"code"`
    Name      string    `json:"name"`
    Qty       int       `json:"qty"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}
```

Tambah konstanta permission di `models/permission.go` (pola sama dengan
`MenuUsers`):

```go
const (
    // ... konstanta yang sudah ada
    MenuStok = "MENU_STOK"
)
```

Konstanta ini hanya dipakai route backend. Baris `CPMENU` tetap dibuat lewat
UI atau seed.

## 3.3. Feature: repository, service, handler

```text
features/stok/
├── stok_repository.go
├── stok_service.go
└── stok_handler.go
```

Ikuti pola feature yang sudah ada (paling mirip: `features/users/`):

- **repository**: `SELECT/INSERT/UPDATE/DELETE`, selalu lewat
  `r.dialect.Table(...)` dan `r.dialect.Bind(query)` supaya jalan di SQL
  Server, PostgreSQL, dan SQLite sekaligus. Pakai `repositories.WithTimeout`.
- **service**: validasi input (kode uppercase, panjang, qty >= 0) dan
  `services.Err...` yang sudah ada; error baru cukup ditambah di
  `services/errors.go` + dipetakan di `internal/web/web.go`.
- **handler**: parse input lewat `web.PathCode`/`web.PathInt`, tulis JSON
  lewat `web.WriteJSON`/`web.WriteError`, dan petakan error dengan
  `web.ServiceError(w, err)`.

## 3.4. Daftarkan route + permission

Tambahkan blok route di `routes/routes.go`. Path boleh `/api/stok` atau
`/api/admin/stok`; yang membatasi hanya permission menunya:

```go
r.With(auth, need(models.MenuStok)).Get("/stok", d.Stok.ListStok)
r.With(auth, need(models.MenuStok)).Post("/stok", d.Stok.CreateStok)
r.With(auth, need(models.MenuStok)).Put("/stok/{code}", d.Stok.UpdateStok)
r.With(auth, need(models.MenuStok)).Delete("/stok/{code}", d.Stok.DeleteStok)
```

Semua fungsi menu **wajib** memakai permission yang sama (`MENU_STOK`).
Jangan membuat `MenuStokRead`, `MenuStokCreate`, atau permission per fungsi
lainnya — role yang boleh membuka menu Stok sudah otomatis boleh memakai
seluruh aksinya.

Tambahkan juga field handler di `routes.Deps` lalu rakit di `main.go`
mengikuti pola `Users`:

```go
// main.go
stokRepo := fstok.NewRepository(db, dialect)
stokSvc, err := fstok.NewService(stokRepo)
if err != nil {
    fail("service init failed", err)
}
// ...
deps := routes.Deps{
    // ... yang sudah ada
    Stok: fstok.NewHandler(stokSvc, auditSvc),
}
```

## 3.5. Verifikasi Backend

```bash
go vet ./...
go test -race ./...
```

✅ **Checkpoint 3.5**

Semua test harus selesai tanpa error, lalu lanjut ke Bagian 4.
---

# Bagian 4 — Build dan Login Ulang

## 4.1. Build Frontend

```bash
cd frontend
npm run lint
npm run build
cd ..
```

Jika memakai Task:

```bash
task build-frontend
```

✅ **Checkpoint 4.1**

- `npm run lint` tidak memiliki error baru.
- `npm run build` selesai dengan `✓ built`.

## 4.2. Restart Backend

Jika memakai development:

```bash
task start
```

Jika sudah memakai binary production:

```bash
task build
./stop.exe
./app.exe
```

Restart backend karena permission dan route baru harus dimuat oleh binary
terbaru.

## 4.3. Beri Akses Menu ke Role

Inilah langkah yang membuat menu benar-benar bisa dipakai. Grant disimpan di
`CPPERMISSION` dan **tidak** otomatis diberikan saat menu dibuat.

1. Buka `http://localhost:5173` pada development.
2. Login `admin` / `admin`.
3. Buka menu **Role & Permission**.
4. Daftar role tampil di sebelah kiri. Pilih role yang boleh memakai menu
   Stok, misalnya `EDITOR`.
5. Cari di matriks sebelah kanan grup **`MODULE TOKO`**, centang `MENU_STOK`.
   Tiap grup punya tombol **Pilih semua** / **Hapus semua** untuk mengelola
   banyak menu sekaligus.
6. Klik **Simpan permission**.

Matriks menampilkan **semua** menu di `CPMENU`, termasuk menu contoh yang
belum pernah di-grant. Jadi administrator bebas memberi akses ke menu mana
pun — begitulah cara memberi akses ke menu contoh `REPORT` bila diinginkan.

✅ **Checkpoint 4.3**

```sql
SELECT ROLE_CODE, MENU_CODE
FROM dbo.CPPERMISSION
WHERE MENU_CODE = 'MENU_STOK'
ORDER BY ROLE_CODE;
```

Pastikan role yang dipilih muncul di hasil query.

## 4.4. Login User yang Diizinkan

1. Logout dari admin bila perlu.
2. Login sebagai user dengan role yang tadi diberi akses.
3. Periksa sidebar: grup **`TOKO`** harus tampil, dan di dalamnya menu
   **Stok**. Buka untuk memastikan halamannya termuat.
4. Request backend-nya juga harus berhasil:

```bash
curl http://localhost:1067/api/stok \
  -H "Authorization: Bearer <token-user-dengan-akses>"
```

✅ **Checkpoint 4.4**

Menu **Stok** muncul di sidebar modul `TOKO` dan endpoint `/api/stok`
mengembalikan data (atau `200` dengan array kosong), bukan `403`.

## 4.5. Login User yang Tidak Diizinkan

1. Logout.
2. Login sebagai user dengan role yang tidak diberi `MENU_STOK`.
3. Periksa sidebar.

✅ **Checkpoint 4.5**

Menu **Stok** tidak muncul, dan request langsung ditolak backend:

```bash
curl http://localhost:1067/api/stok \
  -H "Authorization: Bearer <token-user-tanpa-akses>"
```

Hasilnya:

```json
{"error":"forbidden: missing MENU_STOK"}
```

---

# Bagian 5 — Menambah Role dan User Baru

Role tidak membutuhkan folder menu baru; user tidak membutuhkan kode khusus.
Keduanya hanya mengatur kontrol akses lewat tabel yang sama.

## 5.1. Buat Role

via UI:

1. Login admin.
2. Buka **Role & Permission**.
3. Klik **Role baru**.
4. Isi kode `EDITOR` (huruf besar/angka/underscore, maks 20 karakter) dan
   nama `Editor`.
5. Simpan.

Role `ADMIN` dan `USER` bawaan tidak bisa dihapus. Role custom bisa dihapus
kapan saja, dan penghapusan bersifat **cascade**: seluruh user yang memakai
role itu ikut terhapus, begitu juga grant menunya dan refresh token. Jadi
pindah dulu user bila ingin menyimpan them.

via SQL:

```sql
INSERT INTO dbo.CPROLE (CODE, NAME)
SELECT 'EDITOR', 'Editor'
WHERE NOT EXISTS (
  SELECT 1 FROM dbo.CPROLE WHERE CODE = 'EDITOR'
);
```

Atau pakai template siap pakai yang sekaligus memuat grant menu, pemindahan
user, dan cara hapus role:

```bash
# edit dulu nilai EDITOR di file, lalu:
sqlcmd -S localhost,1433 -U <user> -P "<password>" -d <NAMA_DB> -C -i tutorial/templates/new-role.sql
```

## 5.2. Beri Akses Menu ke Role Itu

Centang menu yang boleh dipakai role `EDITOR` di halaman Role & Permission
(langkah sama seperti 4.3), lalu **Simpan permission**. Contoh centang:

```text
MENU_USERS
MENU_STOK
```

Role tanpa satu pun centang = user akan melihat halaman kosong
"hubungi admin".

## 5.3. Buat User Baru

via UI (disarankan):

1. Login admin.
2. Buka menu **User Account**.
3. Klik **Tambah user**.
4. Isi username, email, dan password (minimal 8 karakter).
5. Pilih role `EDITOR` pada dropdown, lalu simpan.

Kode user dibuat otomatis oleh backend (`USR-XXXXXXXX`) — Anda tidak perlu
menulisnya.

✅ **Checkpoint 5.3**

```sql
SELECT CODE, USERNAME, EMAIL, ROLE_CODE
FROM dbo.CPUSER
WHERE USERNAME = 'editor';
```

Row tersebut akan **belum punya menu apa pun** sampai role `EDITOR` mendapat
grant dan user login ulang. Itu normal: akses menu menempel pada role, bukan
pada user.

## 5.4. Pindahkan User ke Role Lain

via UI: buka **User Account**, klik ikon edit pada user, ganti dropdown
role, simpan.

via SQL:

```sql
UPDATE dbo.CPUSER
SET ROLE_CODE = 'EDITOR'
WHERE CODE = 'USR-XXXXXX';
```

Verifikasi:

```sql
SELECT USERNAME, ROLE_CODE
FROM dbo.CPUSER
WHERE CODE = 'USR-XXXXXX';
```

User harus logout/login ulang setelah role atau grant berubah.
---

# Bagian 6 — Menghapus Akses Menu

Untuk mencabut akses:

1. Login admin.
2. Buka **Role & Permission**.
3. Pilih role.
4. Hilangkan centang `MENU_STOK`.
5. Klik **Simpan permission**.
6. User logout/login ulang.

Setelah login ulang, menu harus hilang dari sidebar.

---

# Bagian 7 — Menghapus Menu

Jika menu tidak diperlukan, hapus via UI **Modul & Menu**
(tombol hapus; gagal bila masih dipakai role atau punya menu anak). Itu
menghapus baris `CPMENU` sekaligus grant-nya (ikut CASCADE).

Pembersihan manual (bila perlu):

1. Hapus folder frontend:

```bash
rm -r frontend/src/menus/TOKO/stok
```

2. Hapus seed `MENU_STOK` dari `scripts/migrate2_rbac.sql` hanya jika
   tidak ada menu lain yang menggunakan permission tersebut.
3. Konstanta di `models/permission.go` hanya perlu dihapus bila route backend
   sempat memakainya.
4. Jika memakai database yang sudah berjalan tanpa UI (grant ikut CASCADE
   saat menunya dihapus):

```sql
DELETE FROM dbo.CPMENU WHERE CODE = 'MENU_STOK';
```

5. Modul `TOKO` boleh dihapus setelah tidak ada menu di dalamnya — hapus lewat
   UI **Modul & Menu** atau `DELETE FROM dbo.CPMATRIX WHERE CODE = N'TOKO';`.
6. Jalankan `task build-frontend`.
7. Refresh/login ulang.

---

# Checklist Akhir

Setup (khusus clone baru):

- [ ] `DB_DATABASE` di `.env` sama persis dengan nama database yang dibuat.
- [ ] `APP_NAME` di `.env` disesuaikan bila nama aplikasi diganti.
- [ ] `task migrate` + `scripts/seed-admin.sql` sudah dijalankan.
- [ ] Login `admin` / `admin` berhasil (ganti passwordnya setelah itu).

Modul & menu:

- [ ] Modul dibuat dulu di `CPMATRIX` (kode UPPERCASE), baru menu di `CPMENU`.
- [ ] `CPMENU.MCONTROL` sama persis dengan nama folder modul.
- [ ] Kode menu = `MENU_` + nama folder menu (`MENU_STOK` ↔ `menus/TOKO/stok/`).
- [ ] Menu diletakkan di `frontend/src/menus/<MCONTROL>/[<grup>/]<menu>/`.
- [ ] `index.jsx` memiliki `export const meta` (setelah import) dan `export default`.
- [ ] `api.js` memakai `auth: true` untuk endpoint privat, tanpa placeholder `<menu>`.
- [ ] Menu bisa diedit dari **Modul & Menu** (label/urutan/modul/parent) tanpa dihapus-buat.
- [ ] `npm run lint` dan `npm run build` berhasil tanpa error.

Role, user, dan akses:

- [ ] Role baru ada di `CPROLE` (dibuat via UI **Role & Permission**).
- [ ] Menghapus role custom = user + grant role itu ikut terhapus (cascade).
- [ ] Admin sudah memberi centang menu yang benar untuk role tersebut (tanpa auto-grant).
- [ ] User baru ada di `CPUSER` dengan role yang benar.
- [ ] User yang diizinkan login ulang dan melihat menu di grup modulnya.
- [ ] User yang tidak diizinkan login ulang dan tidak melihat menu.
- [ ] Request endpoint tanpa akses mendapat `403`.
- [ ] Menu tanpa folder tampil sebagai halaman 404 pemandu (bukan hilang diam-diam).

---

# Lampiran A — Hasil Akhir Database Fresh Clone (acuan)

Repo ini baru core aplikasi (belum ada menu alur bisnis). Clone baru yang
mengikuti README bagian 1 (buat DB kosong bebas nama → `task migrate` →
`seed-admin.sql`) wajib menghasilkan state persis ini:

```sql
-- 10 tabel (grant menyatu di CPPERMISSION; tanpa tabel definisi terpisah):
SELECT TABLE_NAME FROM INFORMATION_SCHEMA.TABLES
WHERE TABLE_TYPE = 'BASE TABLE' AND TABLE_NAME LIKE 'CP%'
ORDER BY TABLE_NAME;
-- CPAUDITLOG, CPMATRIX, CPMENU, CPNOTIFLOG, CPNOTIFTEMPLATE,
-- CPPERMISSION, CPREFRESHTOKEN, CPROLE, CPSYSLOG, CPUSER

-- 2 modul bawaan (SYSTEM operasional + REPORT contoh tes tampilan):
SELECT CODE, LABEL, SORT_ORDER FROM dbo.CPMATRIX ORDER BY SORT_ORDER;
-- SYSTEM | System | 1
-- REPORT | Report | 2

-- 10 menu (8 SYSTEM + 2 contoh REPORT); permission = menu (1:1):
SELECT COUNT(*) FROM dbo.CPMENU;  -- 10

-- Grant: ADMIN 8, USER 0 (akses USER selalu manual via matriks).
-- Dua menu REPORT sengaja tanpa akses role mana pun.
SELECT ROLE_CODE, COUNT(*) AS JML FROM dbo.CPPERMISSION GROUP BY ROLE_CODE;
-- ADMIN | 8

-- 2 akun seed (password = username, wajib diganti):
SELECT USERNAME, ROLE_CODE FROM dbo.CPUSER ORDER BY USERNAME;
-- admin | ADMIN
-- user  | USER

-- 3 template notifikasi bawaan:
SELECT COUNT(*) FROM dbo.CPNOTIFTEMPLATE;  -- 3
```

Bila ada angka yang berbeda, ulangi `task migrate` (aman diulang) lalu
bandingkan lagi sebelum lanjut ke menu bisnis.

Untuk engine lain, `scripts/schema.postgres.sql` dan
`scripts/schema.sqlite.sql` menghasilkan state yang sama persis (termasuk 2
akun seed), jadi angka di atas juga berlaku sebagai acuan.
