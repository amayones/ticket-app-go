# Tutorial: Menambah Menu untuk Semua Role

Tutorial ini menjelaskan cara membuat satu menu yang dapat dipakai oleh semua
role. Admin tidak perlu membuat folder atau kode khusus untuk setiap role.

Setelah menu dibuat, admin hanya perlu memberi centang akses pada role yang
diizinkan. Setelah user login ulang, menu otomatis muncul di sidebar.

## Aturan Utama

Struktur folder menu (modul UPPERCASE, bebas tambah modul baru):

```text
frontend/src/menus/<MCONTROL>/[<parent>/]<menu>/  -> semua role yang diberi akses menu
```

Contoh bawaan:

```text
menus/SYSTEM/users/            -> modul SYSTEM, menu users
menus/REPORT/keu/stok/      -> modul REPORT, parent grup visual "keu", menu stok
```

Modul = kolom `MCONTROL` tabel `CPMENU`, wajib sama persis dengan nama
folder (UPPERCASE). Folder perantara tanpa `index.jsx` = grup visual saja
(di sidebar bisa dibuka-tutup dengan tombol +/−).

Tutorial ini memakai modul `REPORT` dan menu `stok`.

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

## Istilah Penting

| Istilah | Arti |
|---|---|
| Menu | Halaman di sidebar, misalnya User Account atau Stok. |
| Role | Kelompok user, misalnya ADMIN, USER, EDITOR. |
| Permission | Akses ke satu menu, misalnya `MENU_STOK`. |
| Checkbox | Centang di halaman Role & Permission. |
| Sidebar | Daftar menu yang otomatis dibuat dari folder menu. |

Semua perintah `cp`, `ls`, dan `grep` dijalankan dari folder root
repository `go-core` menggunakan Git Bash.

---

# bagian 1 — Daftarkan Menu di Database

Misalnya kita membuat menu `Stok` di modul `REPORT`.

## 1.1. Buat via UI (disarankan)

1. Login `admin`, buka menu **Modul & Menu** (modul `REPORT` sudah ada sebagai contoh).
2. Klik **Menu baru**, isi:
   - Kode permission: `MENU_STOK` (wajib prefix `MENU_`, maks 40 karakter)
   - Nama permission: `Akses menu Stok`
   - Modul (MCONTROL): pilih `REPORT` dari dropdown
   - Label tampil: `Stok`, Urutan: `4`, Parent: kosongkan
3. Simpan. Menu langsung tercatat di `CPMENU`
   **tanpa auto-grant ke role mana pun** — admin mencentang manual di halaman Role.

✅ **Checkpoint 1.1**

```sql
SELECT CODE, MCONTROL, LABEL FROM dbo.CPMENU WHERE CODE = 'MENU_STOK';
```

Harus menghasilkan satu baris (`REPORT`).

## 1.2. Alternatif via SQL

Satu permission mewakili satu menu (`MENU_<NAMA_MENU>`); modul = nama
folder (`REPORT` → `menus/REPORT/`). Cukup satu INSERT (permission = menu):

```sql
INSERT INTO dbo.CPMENU (CODE, MCONTROL, LABEL, SORT_ORDER)
VALUES (N'MENU_STOK', N'REPORT', N'Stok', 4);
```

Grant ke role tetap via matriks UI (atau `INSERT INTO dbo.CPPERMISSION
(ROLE_CODE, MENU_CODE) ...`).

✅ **Checkpoint 1.2**

```sql
SELECT CODE, MCONTROL FROM dbo.CPMENU WHERE CODE = 'MENU_STOK';
```

Harus menghasilkan satu baris. (Jalur UI di 1.1 tidak menulis seed file —
cek ke database, bukan ke `migrate2_rbac.sql`.)

Untuk menu permanen bawaan, tambahkan seed yang sama di
`scripts/migrate2_rbac.sql` (proyek ini khusus SQL Server).

---

# Bagian 2 — Membuat Frontend Menu

## 2.1. Copy Template

```bash
cp -r tutorial/templates/frontend-menu frontend/src/menus/REPORT/stok
```

✅ **Checkpoint 2.1**

```bash
ls frontend/src/menus/REPORT/stok
```

Harus ada:

```text
api.js
index.jsx
```

## 2.2. Isi `index.jsx`

Buka `frontend/src/menus/REPORT/stok/index.jsx`.

Atur metadata:

```jsx
export const meta = {
  label: 'Stok',
  icon: 'list',
  order: 4,
}
```

Ganti isi halaman dengan komponen yang dibutuhkan.

✅ **Checkpoint 2.2**

```bash
grep -n "export default\|export const meta" frontend/src/menus/REPORT/stok/index.jsx
```

Harus ada dua baris export:

```text
export const meta = ...
export default function ...
```

Jangan lupa membuat komponen utama dengan `export default`.

## 2.3. Isi `api.js`

Buka `frontend/src/menus/REPORT/stok/api.js`.

Ganti placeholder `<menu>` dengan endpoint yang benar.

Contoh jika memakai endpoint yang sudah ada:

```js
import { apiRequest as request } from '../../../api/client.js'

export async function listStok() {
  return request('/api/stok', { auth: true })
}
```

Contoh membuat endpoint di frontend:

```js
import { apiRequest as request } from '../../../api/client.js'

export async function listStok() {
  return request('/api/stok', { auth: true })
}

export async function createStok(payload) {
  return request('/api/stok', {
    method: 'POST',
    body: payload,
    auth: true,
  })
}
```

✅ **Checkpoint 2.3**

```bash
grep -rn "<menu>" frontend/src/menus/REPORT/stok || echo OK-tidak-ada-placeholder
```

Harus menghasilkan:

```text
OK-tidak-ada-placeholder
```

## 2.4. Pastikan Permission Mapping Otomatis

Frontend otomatis mengubah nama folder menjadi permission:

```text
frontend/src/menus/REPORT/stok/
              ↓
MENU_STOK   (modul REPORT dari nama folder)
```

Folder perantara tanpa `index.jsx` = grup visual bersarang, mis.
`menus/REPORT/keu/stok/` tetap memakai permission `MENU_STOK`
di bawah grup `keu` (bisa dibuka-tutup di sidebar).

Jadi tidak perlu menambah daftar menu manual di `registry.js`.

✅ **Checkpoint 2.4**

```bash
grep -n "import.meta.glob" frontend/src/menus/registry.js
```

Pastikan glob `./*/**/index.jsx` memindai semua modul (termasuk modul baru
seperti `REPORT`).

## 2.5. Pahami Alur 404 Pemandu

Urutan yang benar: **daftar di database dulu (Bagian 1), folder belakangan
(Bagian 2)**. Di jeda itu, bila menu sudah dicentang ke suatu role, user
role tersebut melihat halaman **404 pemandu** (bukan menu hilang diam-diam):

- Permission yang hilang + path folder persis yang harus dibuat, mis.
  `frontend/src/menus/REPORT/stok/`
- Perintah copy template siap salin
- Checklist rebuild (`npm run build` + restart backend)

Setelah folder dibuat + rebuild, menu asli otomatis menggantikan halaman itu.
Gunakan halaman 404 sebagai kompas: ia selalu menunjuk lokasi yang benar.

---

# Bagian 3 — Menambahkan Backend Jika Dibutuhkan

Jika menu hanya memakai endpoint yang sudah ada, lewati bagian ini.

Jika menu membutuhkan endpoint baru, buat folder:

```text
features/stok/
├── stok_repository.go
├── stok_service.go
└── stok_handler.go
```

Ikuti pola feature yang sudah ada. Query SQL harus memakai variabel `query`
dan `r.dialect.Bind(query)`.

## 3.1. Tambahkan Permission di Route

Route menu dapat berada di path `/api/<menu>` atau `/api/admin/<menu>`.
Nama path tidak membatasi role; yang membatasi hanya permission menu.

Contoh route di `routes/routes.go` (tambah konstanta permission di
`models/permission.go` hanya bila route backend membutuhkannya):

```go
r.With(auth, need("MENU_STOK")).Get("/stok", deps.Stok.ListStok)
r.With(auth, need("MENU_STOK")).Post("/stok", deps.Stok.CreateStok)
r.With(auth, need("MENU_STOK")).Put("/stok/{code}", deps.Stok.UpdateStok)
r.With(auth, need("MENU_STOK")).Delete("/stok/{code}", deps.Stok.DeleteStok)
```

Semua fungsi menu wajib memakai permission yang sama:

```text
MENU_STOK
```

Jangan membuat `MenuStokRead`, `MenuStokCreate`, atau permission
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
6. Cari checkbox `MENU_STOK` di matriks sebelah kanan.
7. Centang menu Stok.
8. Klik **Simpan permission**.

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
3. Tutup dan buka kembali browser bila perlu.
4. Periksa sidebar.

✅ **Checkpoint 4.4**

Menu **Stok** harus otomatis muncul di sidebar.

## 4.5. Login User yang Tidak Diizinkan

1. Logout.
2. Login sebagai user dengan role yang tidak diberi `MENU_STOK`.
3. Periksa sidebar.

✅ **Checkpoint 4.5**

Menu **Stok** harus tidak muncul.

Backend tetap menolak request langsung:

```bash
curl http://localhost:1067/api/stok \
  -H "Authorization: Bearer <token-user-tanpa-akses>"
```

Hasilnya:

```json
{"error":"forbidden: missing MENU_STOK"}
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
MENU_USERS
MENU_STOK
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
rm -r frontend/src/menus/REPORT/stok
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

5. Jalankan `task build-frontend`.
6. Refresh/login ulang.

---

# Checklist Akhir

- [ ] Modul dibuat dulu di `CPMATRIX`, lalu menu di `CPMENU`, lalu centang role.
- [ ] Menu terdaftar di `CPMENU` (`MCONTROL` = nama folder modul, UPPERCASE).
- [ ] Menu diletakkan di `frontend/src/menus/<MCONTROL>/[<parent>/]<menu>/`.
- [ ] `index.jsx` memiliki `export default` dan `export const meta`.
- [ ] `api.js` tidak memakai `apiRequest` tanpa `auth: true` untuk endpoint privat.
- [ ] Tidak ada placeholder `<menu>`.
- [ ] Permission `MENU_<MENU>` terdaftar via UI **Modul & Menu**.
- [ ] Semua route backend memakai permission `MENU_<MENU>` yang sama.
- [ ] `task migrate` sudah dijalankan (untuk seed bawaan).
- [ ] `npm run lint` dan `npm run build` berhasil.
- [ ] Admin sudah memberi centang role yang benar (tanpa auto-grant).
- [ ] User yang diizinkan login ulang dan melihat menu (di grup modulnya).
- [ ] User yang tidak diizinkan login ulang dan tidak melihat menu.
- [ ] Menu tanpa folder tampil sebagai halaman 404 pemandu (bukan hilang diam-diam).
- [ ] Request endpoint tanpa akses mendapat `403`.

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
