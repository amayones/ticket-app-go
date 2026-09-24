# Tutorial: Menambah Menu untuk Semua Role

Tutorial ini menjelaskan cara membuat satu menu yang dapat dipakai oleh semua
role. Admin tidak perlu membuat folder atau kode khusus untuk setiap role.

Setelah menu dibuat, admin hanya perlu memberi centang akses pada role yang
diizinkan. Setelah user login ulang, menu otomatis muncul di sidebar.

## Aturan Utama

Ada dua tipe folder menu:

```text
frontend/src/menus/admin/<menu>/  -> hanya role ADMIN
frontend/src/menus/user/<menu>/   -> semua role yang diberi akses menu
```

Tutorial ini memakai `menus/user/`.

Satu permission mewakili satu menu:

```text
MENU_<NAMA_MENU>
```

Contoh folder `laporan` otomatis menggunakan permission:

```text
MENU_LAPORAN
```

Satu checkbox `MENU_LAPORAN` memberi akses ke **seluruh fungsi** menu
Laporan. Tidak ada permission terpisah untuk list, create, edit, atau delete.

## Istilah Penting

| Istilah | Arti |
|---|---|
| Menu | Halaman di sidebar, misalnya User Account atau Laporan. |
| Role | Kelompok user, misalnya ADMIN, USER, EDITOR. |
| Permission | Akses ke satu menu, misalnya `MENU_LAPORAN`. |
| Checkbox | Centang di halaman Role & Permission. |
| Sidebar | Daftar menu yang otomatis dibuat dari folder menu. |

Semua perintah `cp`, `ls`, dan `grep` dijalankan dari folder root
repository `go-core` menggunakan Git Bash.

---

# bagian 1 — Tentukan Permission Menu

Misalnya kita membuat menu `Laporan`.

## 1.1. Tambahkan Konstanta Permission

Buka `models/permission.go`, lalu tambahkan satu konstanta:

```go
const (
  // Konstanta lain yang sudah ada...
  MenuLaporan = "MENU_LAPORAN"
)
```

✅ **Checkpoint 1.1**

```bash
grep -n "MenuLaporan" models/permission.go
```

Harus ada satu hasil, misalnya:

```text
17:  MenuLaporan = "MENU_LAPORAN"
```

Nama permission harus sama dengan nama folder menu:

```text
laporan       -> MENU_LAPORAN
laporan-penjualan -> MENU_LAPORAN_PENJUALAN
```

## 1.2. Tambahkan Seed Permission

Buka `scripts/migrate2_rbac.sql`. Pada daftar `@perms`, tambahkan:

```sql
(N'MENU_LAPORAN', N'Akses menu Laporan', N'MENU', N'Seluruh fungsi menu Laporan'),
```

Contoh posisi di dalam `INSERT INTO @perms VALUES`:

```sql
INSERT INTO @perms VALUES
  (N'MENU_DASHBOARD', N'Akses menu Dashboard', N'MENU', N'Seluruh fungsi dashboard'),
  (N'MENU_LAPORAN', N'Akses menu Laporan', N'MENU', N'Seluruh fungsi menu Laporan');
```

✅ **Checkpoint 1.2**

```bash
grep -n "MENU_LAPORAN" scripts/migrate2_rbac.sql
```

 Harus ada minimal satu hasil.

## 1.3. Jalankan Migrasi Database

Jalankan dari folder root:

```bash
task migrate
```

Tanpa Task CLI:

```bash
sqlcmd -S localhost,1433 -U may -P "password-database-anda" -d Go -C -i scripts/migrate.sql
sqlcmd -S localhost,1433 -U may -P "password-database-anda" -d Go -C -i scripts/migrate2_rbac.sql
```

✅ **Checkpoint 1.3**

```sql
SELECT CODE, NAME
FROM dbo.CPPERMISSION
WHERE CODE = 'MENU_LAPORAN';
```

Harus menghasilkan satu baris.

---

# Bagian 2 — Membuat Frontend Menu

## 2.1. Copy Template

```bash
cp -r tutorial/templates/frontend-menu frontend/src/menus/user/laporan
```

✅ **Checkpoint 2.1**

```bash
ls frontend/src/menus/user/laporan
```

Harus ada:

```text
api.js
index.jsx
```

## 2.2. Isi `index.jsx`

Buka `frontend/src/menus/user/laporan/index.jsx`.

Atur metadata:

```jsx
export const meta = {
  label: 'Laporan',
  icon: 'list',
  order: 8,
}
```

Ganti isi halaman dengan komponen yang dibutuhkan.

✅ **Checkpoint 2.2**

```bash
grep -n "export default\|export const meta" frontend/src/menus/user/laporan/index.jsx
```

Harus ada dua baris export:

```text
export const meta = ...
export default function ...
```

Jangan lupa membuat komponen utama dengan `export default`.

## 2.3. Isi `api.js`

Buka `frontend/src/menus/user/laporan/api.js`.

Ganti placeholder `<menu>` dengan endpoint yang benar.

Contoh jika memakai endpoint yang sudah ada:

```js
import { apiRequest as request } from '../../../api/client.js'

export async function listLaporan() {
  return request('/api/laporan', { auth: true })
}
```

Contoh membuat endpoint di frontend:

```js
import { apiRequest as request } from '../../../api/client.js'

export async function listLaporan() {
  return request('/api/laporan', { auth: true })
}

export async function createLaporan(payload) {
  return request('/api/laporan', {
    method: 'POST',
    body: payload,
    auth: true,
  })
}
```

✅ **Checkpoint 2.3**

```bash
grep -rn "<menu>" frontend/src/menus/user/laporan || echo OK-tidak-ada-placeholder
```

Harus menghasilkan:

```text
OK-tidak-ada-placeholder
```

## 2.4. Pastikan Permission Mapping Otomatis

Frontend otomatis mengubah nama folder menjadi permission:

```text
frontend/src/menus/user/laporan/
             ↓
MENU_LAPORAN
```

Jadi tidak perlu menambah daftar menu manual di `registry.js`.

✅ **Checkpoint 2.4**

```bash
grep -n "permissionFor\|menus/user" frontend/src/menus/registry.js
```

Pastikan `menus/user/` dipindai oleh `import.meta.glob`.

---

# Bagian 3 — Menambahkan Backend Jika Dibutuhkan

Jika menu hanya memakai endpoint yang sudah ada, lewati bagian ini.

Jika menu membutuhkan endpoint baru, buat folder:

```text
features/laporan/
├── laporan_repository.go
├── laporan_service.go
└── laporan_handler.go
```

Ikuti pola feature yang sudah ada. Query SQL harus memakai variabel `query`
dan `r.dialect.Bind(query)`.

## 3.1. Tambahkan Permission di Route

Karena menu ini dapat dipakai semua role, jangan masukkan route ke blok
`/api/admin` yang hanya untuk ADMIN.

Contoh route di `routes/routes.go`:

```go
r.With(auth, need(models.MenuLaporan)).Get("/laporan", deps.Laporan.ListLaporan)
r.With(auth, need(models.MenuLaporan)).Post("/laporan", deps.Laporan.CreateLaporan)
r.With(auth, need(models.MenuLaporan)).Put("/laporan/{code}", deps.Laporan.UpdateLaporan)
r.With(auth, need(models.MenuLaporan)).Delete("/laporan/{code}", deps.Laporan.DeleteLaporan)
```

Semua fungsi menu wajib memakai permission yang sama:

```text
models.MenuLaporan
```

Jangan membuat `MenuLaporanRead`, `MenuLaporanCreate`, atau permission
per fungsi lainnya.

## 3.2. Daftarkan Handler

Tambahkan handler di `routes.Deps` dan rakit di `main.go` mengikuti pola
menu yang sudah ada.

## 3.3. Verifikasi Backend

```bash
go vet ./...
go test -race ./...
```

✅ **Checkpoint 3.3**

Semua test harus selesai tanpa error.

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

## 4.3. Login sebagai Admin

1. Buka `http://localhost:5173` pada development.
2. Login `admin` / `admin`.
3. Buka menu **Role & Permission**.
4. Daftar role akan tampil di sebelah kiri.
5. Pilih role, misalnya `USER` atau `EDITOR`.
6. Cari checkbox `MENU_LAPORAN` di matriks sebelah kanan.
7. Centang menu Laporan.
8. Klik **Simpan permission**.

✅ **Checkpoint 4.3**

```sql
SELECT ROLE_CODE, PERMISSION_CODE
FROM dbo.CPROLEPERMISSION
WHERE PERMISSION_CODE = 'MENU_LAPORAN'
ORDER BY ROLE_CODE;
```

Pastikan role yang dipilih muncul di hasil query.

## 4.4. Login User yang Diizinkan

1. Logout dari admin bila perlu.
2. Login sebagai user dengan role yang tadi diberi akses.
3. Tutup dan buka kembali browser bila perlu.
4. Periksa sidebar.

✅ **Checkpoint 4.4**

Menu **Laporan** harus otomatis muncul di sidebar.

## 4.5. Login User yang Tidak Diizinkan

1. Logout.
2. Login sebagai user dengan role yang tidak diberi `MENU_LAPORAN`.
3. Periksa sidebar.

✅ **Checkpoint 4.5**

Menu **Laporan** harus tidak muncul.

Backend tetap menolak request langsung:

```bash
curl http://localhost:1067/api/laporan \
  -H "Authorization: Bearer <token-user-tanpa-akses>"
```

Hasilnya:

```json
{"error":"forbidden: missing MENU_LAPORAN"}
```

---

# Bagian 5 — Menambah Role Baru

Role baru tidak membutuhkan folder menu baru.

## 5.1. Buat Role

via UI:

1. Login admin.
2. Buka **Role & Permission**.
3. Klik **Role baru**.
4. Isi kode, misalnya `EDITOR`.
5. Simpan.

Atau via SQL:

```sql
INSERT INTO dbo.CPROLE (CODE, NAME)
SELECT 'EDITOR', 'Editor'
WHERE NOT EXISTS (
  SELECT 1 FROM dbo.CPROLE WHERE CODE = 'EDITOR'
);
```

## 5.2. Beri Akses Menu

Centang menu yang boleh dipakai role `EDITOR` di halaman Role & Permission.
Contohnya centang:

```text
MENU_DASHBOARD
MENU_LAPORAN
```

Lalu simpan.

## 5.3. Pindahkan User ke Role

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

User harus logout/login ulang.

---

# Bagian 6 — Menghapus Akses Menu

Untuk mencabut akses:

1. Login admin.
2. Buka **Role & Permission**.
3. Pilih role.
4. Hilangkan centang `MENU_LAPORAN`.
5. Klik **Simpan permission**.
6. User logout/login ulang.

Setelah login ulang, menu harus hilang dari sidebar.

---

# Bagian 7 — Menghapus Menu

Jika menu tidak diperlukan:

1. Hapus folder frontend:

```bash
rm -r frontend/src/menus/user/laporan
```

2. Hapus konstanta `MenuLaporan` dari `models/permission.go`.
3. Hapus seed `MENU_LAPORAN` dari `scripts/migrate2_rbac.sql` hanya jika
   tidak ada menu lain yang menggunakan permission tersebut.
4. Jika memakai database yang sudah berjalan, bersihkan mapping dengan SQL:

```sql
DELETE FROM dbo.CPROLEPERMISSION
WHERE PERMISSION_CODE = 'MENU_LAPORAN';

DELETE FROM dbo.CPPERMISSION
WHERE CODE = 'MENU_LAPORAN';
```

5. Jalankan `task build-frontend`.
6. Refresh/login ulang.

---

# Checklist Akhir

- [ ] Menu diletakkan di `frontend/src/menus/user/<menu>/`.
- [ ] `index.jsx` memiliki `export default` dan `export const meta`.
- [ ] `api.js` tidak memakai `apiRequest` tanpa `auth: true` untuk endpoint privat.
- [ ] Tidak ada placeholder `<menu>`.
- [ ] Permission `MENU_<MENU>` ada di `models/permission.go`.
- [ ] Permission yang sama ada di seed `migrate2_rbac.sql`.
- [ ] Semua route backend memakai `models.Menu<MENU>` yang sama.
- [ ] `task migrate` sudah dijalankan.
- [ ] `npm run lint` dan `npm run build` berhasil.
- [ ] Admin sudah memberi centang role yang benar.
- [ ] User yang diizinkan login ulang dan melihat menu.
- [ ] User yang tidak diizinkan login ulang dan tidak melihat menu.
- [ ] Request endpoint tanpa akses mendapat `403`.
