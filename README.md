# Go Core

Aplikasi web **Go + React** untuk login, manajemen user, role, permission menu, audit, system log, dan notifikasi.

- Backend: Go + chi
- Frontend: React + Vite
- Database utama saat ini: **SQL Server**
- Development: backend `http://localhost:1067`, frontend `http://localhost:5173`
- Production: frontend dan API disajikan oleh satu `app.exe`

```text
Browser :5173 ── Vite proxy ──► app.exe :1067 ──► SQL Server database <NAMA_DB>
Browser :1067 ─────────────────► app.exe :1067 ──► SQL Server database <NAMA_DB>
```

> Nama database **bebas** (mis. `GoCore`, `AppDb`, `PerusahaanDb`) — yang
> penting sama persis di `DB_DATABASE` (.env) dan database yang Anda buat
> di SQL Server. Tidak harus bernama `Go`.

## Fitur Utama

- Login tanpa registrasi publik
- User hanya dibuat oleh admin
- JWT access 15 menit
- Refresh token 7 hari dengan rotasi
- Maksimal 5 sesi per user
- RBAC sederhana: **satu permission untuk satu menu**
- Semua role dapat memakai menu bila role tersebut diberi akses `MENU_*`
- Master modul di tabel **`CPMATRIX`**, registry menu di tabel **`CPMENU`**
  (`MCONTROL` = folder modul UPPERCASE), grant role→menu di
  **`CPPERMISSION`**; matriks dibaca dari JOIN ketiga tabel
- Sidebar dikelompokkan per modul (tombol +/−), mendukung grup visual
  (folder perantara tanpa `index.jsx`, bukan menu tersendiri)
- Menu terdaftar tapi folder belum dibuat tampil sebagai **halaman 404
  pemandu** (menunjukkan path persis), bukan hilang diam-diam
- Menu baru dibuat via UI Modul & Menu (tanpa auto-grant)
- Role tanpa akses apa pun (mis. `USER` baru) mendapat halaman kosong
- Tidak ada pembatasan berdasarkan nama role atau folder admin/user
- Role & Permission: daftar role di kiri, matriks akses menu di kanan
- Audit log, system log, dan notifikasi
- Popup login ulang ketika sesi habis tanpa pindah halaman
- Build frontend dan backend dalam satu binary

> Mau memakai untuk proyek Anda sendiri? Urutan yang benar:
> [ganti nama aplikasi & database](#19-mengganti-nama-aplikasi) →
> [tambah role, user, modul, menu](#7-resep-cepat-role-user-modul-menu) →
> [tutorial lengkap](tutorial/README.md).

---

# 1. Setup dari Nol sampai Aplikasi Jalan

Ikuti bagian ini secara berurutan.

## 1.1. Prepare Komputer

Prasyarat:

| Kebutuhan | Versi | Cek |
|---|---:|---|
| Go | 1.27+ | `go version` |
| Node.js | 20+ | `node --version` |
| npm | 9+ | `npm --version` |
| SQL Server | 2019+ | `sqlcmd -?` |
| Task CLI | opsional | `task --version` |

Clone repository:

```bash
git clone <url-repo>.git
cd go-core
```

## 1.2. Install Dependency Frontend

```bash
task install
```

Tanpa Task CLI:

```bash
cd frontend
npm ci
cd ..
```

## 1.3. Membuat File `.env`

Windows CMD:

```cmd
copy .env.example .env
notepad .env
```

PowerShell:

```powershell
Copy-Item .env.example .env
notepad .env
```

Isi minimal variabel ini:

```env
APP_PORT=1067
DB_CONNECTION=sqlserver
DB_HOST=localhost
DB_PORT=1433
DB_DATABASE=<NAMA_DB-bebas-mis-GoCore>
DB_USERNAME=may
DB_PASSWORD=password-database-anda
JWT_SECRET=ganti-dengan-string-acak-minimal-32-karakter
ACCESS_TOKEN_MINUTES=15
REFRESH_TOKEN_DAYS=7
```

Umur sesi dapat diubah tanpa rebuild: `ACCESS_TOKEN_MINUTES` dalam menit
(default 15), `REFRESH_TOKEN_DAYS` dalam hari (default 7). Untuk mengecek
modal "Sesi habis", set `ACCESS_TOKEN_MINUTES=1` lalu restart backend.

Buat `JWT_SECRET` acak:

```bash
openssl rand -hex 32
```

Jangan commit `.env`. File ini hanya untuk konfigurasi lokal.

## 1.4. Membuat Database SQL Server

Buka SQL Server Management Studio atau gunakan `sqlcmd`. Ganti
`<NAMA_DB>` dengan nama pilihan Anda (bebas, mis. `GoCore`) dan pakai nama
yang **sama** di `.env` (`DB_DATABASE`) dan semua perintah `-d` di bawah.

### Membuat database kosong

```sql
IF DB_ID(N'<NAMA_DB>') IS NULL
BEGIN
  CREATE DATABASE [<NAMA_DB>];
END
GO
```

### Menjalankan migrasi aplikasi

Dari folder root repository:

```bash
task migrate
```

`task migrate` membaca `.env`, jadi pastikan `DB_DATABASE` sudah diisi nama
yang sama. Task ini menjalankan:

1. `scripts/migrate.sql`
2. `scripts/migrate2_rbac.sql`

Tanpa Task CLI (ganti `<USER>`, `<PASSWORD>`, `<NAMA_DB>`):

```bash
sqlcmd -S localhost,1433 -U <USER> -P "<PASSWORD>" -d <NAMA_DB> -C -i scripts/migrate.sql
sqlcmd -S localhost,1433 -U <USER> -P "<PASSWORD>" -d <NAMA_DB> -C -i scripts/migrate2_rbac.sql
```

Migrasi aman diulang (boleh dijalankan berkali-kali). `migrate2_rbac.sql`
 membersihkan definisi permission lama per-fitur, lalu membuat master modul
(`CPMATRIX`), registry menu (`CPMENU`), dan grant default `CPPERMISSION`.

### Membuat akun awal

```bash
sqlcmd -S localhost,1433 -U <USER> -P "<PASSWORD>" -d <NAMA_DB> -C -i scripts/seed-admin.sql
```

Akun development yang dibuat:

| Username | Password | Role | Keterangan |
|---|---|---|---|
| `admin` | `admin` | `ADMIN` | Ganti password setelah login |
| `user` | `user` | `USER` | Ganti password setelah login |

Password di atas hanya untuk development. Jangan digunakan di production.

## 1.5. Verifikasi Database

Pastikan tabel dan data utama sudah benar.

### Checklist tabel

```sql
SELECT TABLE_NAME
FROM INFORMATION_SCHEMA.TABLES
WHERE TABLE_TYPE = 'BASE TABLE'
ORDER BY TABLE_NAME;
```

Harus ada 10 tabel:

```text
CPAUDITLOG
CPMATRIX
CPMENU
CPNOTIFLOG
CPNOTIFTEMPLATE
CPPERMISSION
CPREFRESHTOKEN
CPROLE
CPSYSLOG
CPUSER
```

### Verifikasi modul (CPMATRIX)

```sql
SELECT CODE, LABEL, SORT_ORDER
FROM dbo.CPMATRIX
ORDER BY SORT_ORDER;
```

Hasil yang benar (2 modul bawaan; tambah modul baru via UI **Modul & Menu**):

```text
SYSTEM  System  1
REPORT  Report  2
```

### Verifikasi role

```sql
SELECT CODE, NAME
FROM dbo.CPROLE
ORDER BY CODE;
```

Hasil minimum:

```text
ADMIN  Administrator
USER   Pengguna
```

### Verifikasi registry menu (CPMENU)

```sql
SELECT CODE, MCONTROL, LABEL, SORT_ORDER
FROM dbo.CPMENU
ORDER BY SORT_ORDER;
```

Hasil yang benar (10 baris: 8 menu `SYSTEM` + 2 contoh `REPORT`):

```text
MENU_USERS          SYSTEM  User Account       1
MENU_ROLES          SYSTEM  Role & Permission  2
MENU_SESSIONS       SYSTEM  Sesi & Auth        3
MENU_AUDIT          SYSTEM  Audit Log          4
MENU_SECURITY       SYSTEM  Security Center    5
MENU_SYSLOG         SYSTEM  System Log         6
MENU_NOTIFICATIONS  SYSTEM  Notifikasi         7
MENU_MODUL          SYSTEM  Modul & Menu       8
MENU_LAPORAN        REPORT  Laporan            1
MENU_ARUS_KAS       REPORT  Arus Kas           2
```

Dua menu `REPORT` adalah contoh tes tampilan (menu biasa vs menu di dalam grup
visual) yang sengaja tanpa akses role mana pun.

### Verifikasi grant role -> menu (CPPERMISSION)

Satu-satunya tabel relasi: role boleh tampil menu apa. Definisi menu
tinggal di `CPMENU` (tidak ada tabel definisi terpisah).

```sql
SELECT ROLE_CODE, MENU_CODE
FROM dbo.CPPERMISSION
ORDER BY ROLE_CODE, MENU_CODE;
```

Hasil yang benar (8 baris `ADMIN`; contoh `REPORT` tanpa akses):

```text
ADMIN  MENU_AUDIT
ADMIN  MENU_MODUL
ADMIN  MENU_NOTIFICATIONS
ADMIN  MENU_ROLES
ADMIN  MENU_SECURITY
ADMIN  MENU_SESSIONS
ADMIN  MENU_SYSLOG
ADMIN  MENU_USERS
```

### Verifikasi akses admin dan user

```sql
SELECT ROLE_CODE, COUNT(*) AS MENU_COUNT
FROM dbo.CPPERMISSION
WHERE MENU_CODE LIKE 'MENU[_]%'
GROUP BY ROLE_CODE
ORDER BY ROLE_CODE;
```

Pada fresh install, hasil default:

```text
ADMIN  8
```

`USER` tidak memiliki baris (nol menu) — akses diberikan manual oleh admin
via matriks. Setelah admin mencentang menu untuk suatu role, query JOIN
berikut menampilkannya (modul otomatis ketahuan dari `CPMENU`):

```sql
SELECT g.ROLE_CODE, m.MCONTROL AS MODULE, g.MENU_CODE
FROM dbo.CPPERMISSION g
JOIN dbo.CPMENU m ON m.CODE = g.MENU_CODE
ORDER BY g.ROLE_CODE, m.SORT_ORDER;
```

### Verifikasi user

```sql
SELECT CODE, USERNAME, EMAIL, ROLE_CODE
FROM dbo.CPUSER
ORDER BY USERNAME;
```

Fresh install berisi 2 akun (`admin`/`ADMIN`, `user`/`USER`).

### Verifikasi template notifikasi

```sql
SELECT CODE, NAME, CHANNEL
FROM dbo.CPNOTIFTEMPLATE
ORDER BY CODE;
```

Hasil yang benar (3 template bawaan):

```text
NTPL-ALERT   Peringatan keamanan  INAPP
NTPL-RESET   Reset password       EMAIL
NTPL-WELCOME Selamat datang       EMAIL
```

## 1.6. Menjalankan Development

Jalankan backend dan frontend sekaligus:

```bash
task start
```

Tanpa Task CLI:

```bash
bash ./scripts/start.sh
```

Buka:

```text
http://localhost:5173
```

Login:

```text
admin / admin
```

Setelah login (admin bawaan memegang 8 menu `SYSTEM`; menu contoh `REPORT`
sengaja belum di-grant — perhatikan bahwa mencentang di matriks lalu
**Simpan permission** akan memberi akses ke role itu):

1. Menu pertama otomatis terbuka, sidebar dikelompokkan per modul (SYSTEM).
2. Buka **Role & Permission**.
3. Role `ADMIN` dan `USER` harus terlihat.
4. Matriks harus menampilkan 10 menu, bukan permission per fungsi.
5. Menu **Modul & Menu** menampilkan 2 modul + 10 baris `CPMENU`.
6. Centang menu untuk role yang membutuhkan (tanpa auto-grant).
7. Klik **Simpan permission**.
8. User dengan role tersebut harus logout/login ulang agar permission terbaru dimuat.
9. Login sebagai `user`/`user` (nol menu) → halaman kosong "hubungi admin".

## 1.7. Menjalankan Production

Build frontend dan backend:

```bash
task build
```

Jalankan:

```bash
./app.exe
```

Buka:

```text
http://localhost:1067/
```

Mode background Windows:

```bash
./app.exe --hide
./stop.exe
```

## 1.8. Health Check

```bash
curl http://localhost:1067/healthz
```

Hasil yang benar:

```json
{"status":"ok"}
```

Jika memakai development, frontend berjalan di port `5173`, tetapi API tetap diproxy ke backend `1067`.

## 1.9. Mengganti Nama Aplikasi

Satu sumber nama aplikasi: `APP_NAME` di `.env`.

```env
APP_NAME=Toko Saya
```

Yang otomatis mengikuti `APP_NAME`:

| Yang berubah | Keterangan |
|---|---|
| Judul tab browser | Di-set runtime oleh `main.jsx` |
| Nama di sidebar aplikasi | `App.jsx` (desktop + topbar mobile) |
| Nama di kartu login | `pages/Auth.jsx` |
| Teks "hubungi admin" & footer | `App.jsx` |
| `Config.AppName` backend | Dibaca `config/env.go` |

Setelah mengubah `.env`, **build ulang frontend** (`task build-frontend`),
lalu restart backend. Clone lama yang masih punya `APP_NAME=GoBackend` (nilai
default lama) ikut diganti ke nama yang Anda pilih. `frontend/index.html`
masih berisi `<title>Go Core</title>` sebagai fallback sebelum JavaScript
berjalan — ganti juga file itu bila ingin judulnya benar sejak awal.

### Mengganti favicon / logo

Favicon berada di:

```text
frontend/public/favicon.svg
```

Folder `public/` disalin apa adanya ke `dist/`, jadi file di sana bisa
dipakai langsung tanpa import. Ganti file `favicon.svg` (atau tambah file
lain, mis. `logo-toko.png`) lalu arahkan `APP_LOGO` di `.env`:

```env
APP_LOGO=/favicon.svg
# atau
APP_LOGO=/logo-toko.png
```

`APP_LOGO` dipakai di dua tempat sekaligus:

| Tempat | Keterangan |
|---|---|
| Favicon tab browser | Di-set runtime oleh `main.jsx` |
| Logo di sidebar, topbar mobile, dan kartu login | `components/AppLogo.jsx` |

Jadi **tidak ada huruf logo lagi** — semuanya gambar favicon. Kalau file
gambar rusak/hilang, `AppLogo` otomatis jatuh ke huruf pertama `APP_NAME`.

Saran gambar: rasio 1:1 (persis), bentuk persegi, latar transparan atau
putih, minimal 64×64 px. Format bebas (SVG, PNG, WebP).

Cache browser untuk favicon sangat agresif. Kalau logo tidak berubah setelah
ganti file, pakai nama file baru (mis. `favicon-toko.svg`) atau hard refresh
(`Ctrl+Shift+R`).

### Branding lain yang tidak ikut `.env`

| File | Isi |
|---|---|
| `frontend/index.html` | `<title>` + `<link rel="icon">` fallback |
| `README.md` | Judul & deskripsi proyek |
| `Dockerfile`, `Taskfile.yml` | Nama binary bila ingin `toko.exe` |
| Tabel `CPNOTIFTEMPLATE` | Teks email (bisa diedit dari menu **Notifikasi**) |

Nama modul bawaan (`SYSTEM`, `REPORT`) dan kode tabel `CP*` tidak perlu
diubah — keduanya hanya internal. Module baru cukup ditambah lewat UI
**Modul & Menu**, bukan mengganti folder yang ada.

## 1.10. Memakai Database Lain

### Nama database berbeda (SQL Server)

Nama database **bebas**. Yang wajib sama persis:

1. Nama database yang dibuat di SQL Server.
2. `DB_DATABASE` di `.env`.
3. Argumen `-d` pada perintah `sqlcmd` manual.

Contoh memakai nama `TokoDb`:

```sql
IF DB_ID(N'TokoDb') IS NULL
BEGIN
  CREATE DATABASE [TokoDb];
END
GO
```

```env
DB_DATABASE=TokoDb
```

```bash
task migrate
sqlcmd -S localhost,1433 -U <USER> -P "<PASSWORD>" -d TokoDb -C -i scripts/seed-admin.sql
```

Tidak ada kode aplikasi yang perlu diedit. Nama tersebut tidak muncul di
query aplikasi — semua query memakai `CP*` tanpa prefix database.

### Engine lain: PostgreSQL / SQLite

Repository ini menyediakan skema fresh-install untuk keduanya. SQL Server
adalah jalur utama yang dipakai harian; dua skema lain tersedia bila Anda
tidak memakai SQL Server.

PostgreSQL:

```env
DB_CONNECTION=postgres
DB_HOST=localhost
DB_PORT=5432
DB_DATABASE=tokodb
DB_USERNAME=postgres
DB_PASSWORD=password-postgres-anda
```

```bash
task migrate-postgres
# atau manual:
psql -h localhost -U postgres -d tokodb -v ON_ERROR_STOP=1 -f scripts/schema.postgres.sql
```

SQLite (cukup satu file, tanpa server):

```env
DB_CONNECTION=sqlite
DB_DATABASE=./data/tokodb.db
```

```bash
task migrate-sqlite
# atau manual:
mkdir -p ./data && sqlite3 ./data/tokodb.db < scripts/schema.sqlite.sql
```

Ketiganya menghasilkan state awal yang sama: 10 tabel, 2 modul, 10 menu,
`ADMIN` 8 grant, 2 akun (`admin`/`admin`, `user`/`user`), 3 template
notifikasi. Skema ikut `IF NOT EXISTS` / `WHERE NOT EXISTS` sehingga aman
dijalankan ulang. Untuk PostgreSQL, buat database dulu dengan
`CREATE DATABASE tokodb;` sebelum menjalankan skemanya.

Username/password pada `DB_USERNAME`/`DB_PASSWORD` diabaikan saat
`DB_CONNECTION=sqlite`.

---

# 2. Cara Kerja Permission Menu

Sistem sekarang tidak memakai permission per fungsi.

Contoh:

```text
MENU_USERS
```

Satu permission tersebut sudah berarti role boleh memakai **seluruh fungsi**
di menu User Account: melihat daftar, membuat, mengubah, menghapus, mengganti
role, dan aksi lain yang berada di menu tersebut.

Tidak ada lagi permission seperti:

```text
USER_CREATE
USER_UPDATE
USER_DELETE
ROLE_READ
SESSION_MANAGE
NOTIF_SEND
```

## 2.1. Struktur Module Menu

Folder pertama adalah module/kategori (UPPERCASE, bebas tambah modul baru),
bukan batas role:

```text
frontend/src/menus/SYSTEM/<menu>/            -> MODULE SYSTEM
frontend/src/menus/REPORT/keu/<menu>/        -> MODULE REPORT, parent grup visual "keu"
```

Frontend otomatis memindai semua module melalui `menus/registry.js`
(glob `./*/**/index.jsx`). Module hanya mengelompokkan menu di sidebar dan
matriks. Akses tetap ditentukan oleh permission menu + baris `CPMENU`.

```text
menus/SYSTEM/users   -> MENU_USERS  (MCONTROL SYSTEM)
menus/SYSTEM/roles   -> MENU_ROLES  (MCONTROL SYSTEM)
menus/REPORT/laporan -> MENU_LAPORAN (MCONTROL REPORT)
```

Tidak ada lagi folder `menus/account` atau menu dashboard. Semua
role—ADMIN, USER, dan role custom—boleh memakai menu yang sama bila role
tersebut memiliki `MENU_*` yang sesuai; role tanpa akses mendapat halaman
kosong.

## 2.2. Alur Memberi Akses Menu

1. Admin membuat menu via **Modul & Menu** (atau memilih menu bawaan).
2. Admin membuat atau memilih role.
3. Matriks menampilkan menu per module (`SYSTEM`, `REPORT`, ...).
4. Admin centang menu yang boleh diakses role tersebut.
5. Admin klik **Simpan permission**.
6. Grant tersimpan di `CPPERMISSION` (ROLE_CODE -> MENU_CODE).
7. Saat login, frontend mengambil permission + entri menu user dari `GET /api/users/me`.
8. Sidebar hanya menampilkan menu yang permission-nya dimiliki user; menu
   terdaftar tapi folder belum dibuat tampil sebagai halaman 404 pemandu.
9. Backend juga memeriksa permission yang sama pada endpoint menu.

Contoh:

```text
Role EDITOR + MENU_LAPORAN -> menu Laporan tampil
Role EDITOR tanpa akses    -> menu Laporan tidak tampil
Role ADMIN                 -> semua menu yang dicentang
```

## 2.3. Kenapa User Harus Login Ulang?

Permission user diambil saat login. Jika admin baru memberi atau mencabut
akses menu, user yang sedang login perlu logout/login ulang. Setelah itu,
sidebar dimuat ulang sesuai permission terbaru.

---

# 3. Struktur Proyek Penting

```text
go-core/
├── main.go
├── routes/                 # registrasi endpoint
├── middleware/             # JWT + permission menu
├── models/                 # konstanta permission MENU_*, menu, dan DTO
├── features/               # service, repository, handler backend
├── frontend/
│   ├── src/api/client.js   # request, login, session, getMe
│   ├── src/menus/
│   │   ├── registry.js     # auto-scan menu nested per module + 404 pemandu
│   │   └── SYSTEM/         # MODULE SYSTEM (8 menu: operasional + Modul & Menu)
│   │   └── REPORT/         # MODULE REPORT (2 contoh tes: biasa vs di dalam grup)
│   ├── src/components/     # UI kit (termasuk MissingMenu)
│   └── src/pages/          # LoginForm
├── scripts/                # migrate*.sql (SQL Server) dan build script
├── tutorial/               # tutorial menu baru (SQL Server)
├── Taskfile.yml
└── README.md
```

## 3.1. Alur Request

```text
Frontend menu
    ↓
menus/<module>/<menu>/api.js
    ↓
apiRequest dari api/client.js
    ↓
Authorization: Bearer <access_token>
    ↓
middleware NewAuth
    ↓
middleware RequirePermission MENU_<MENU>
    ↓
handler → service → repository → SQL Server
```

---

# 4. Menjalankan Test dan Build

Backend:

```bash
go vet ./...
go test -race ./...
```

Frontend:

```bash
cd frontend
npm run lint
npm run build
cd ..
```

Build production:

```bash
task build
```

---

# 5. Troubleshooting

| Gejala | Penyebab dan solusi |
|---|---|
| `Matriks akses menu kosong` | Database belum menjalankan `migrate2_rbac.sql` → jalankan `task migrate`, restart backend, lalu refresh browser. |
| `Tercatat 0 dari 0 permission` | `CPPERMISSION` masih kosong/permission lama belum dimigrasi → jalankan `task migrate`. |
| Login berhasil tetapi sidebar kosong | Wajar bila role memang nol menu (mis. `USER` baru) → buka Role & Permission, centang menu, simpan, lalu login ulang. |
| Menu terdaftar tapi tampil 404 | Folder `frontend/src/menus/<MCONTROL>/[<parent>/]<key>/` belum dibuat → ikuti petunjuk di halaman 404 (copy template, `npm run build`, restart). |
| Menu baru tidak muncul sama sekali | Permission belum dicentang ke role tersebut, atau user belum login ulang. |
| Menu baru muncul untuk semua role | Permission menu belum diberikan/di-filter dengan benar; cek `CPPERMISSION` role tersebut. |
| Permission endpoint 403 | User belum memiliki `MENU_<MENU>` atau request dikirim ke role/menu yang salah. |
| `WARN frontend/dist missing` | Normal saat development; build frontend dengan `task build-frontend` jika ingin menghapus warning. |
| `localhost:1067` tidak bisa dibuka | Backend belum jalan → `task start` atau `go run .`. |
| `go:embed no matching files` | Jalankan `task build-frontend`, pastikan `frontend/dist/.gitignore` ada. |
| Koneksi DB gagal saat start | `DB_DATABASE` di `.env` tidak sama dengan nama database di SQL Server, atau kredensial salah → samakan ketiganya (`DB_HOST/DB_PORT/DB_DATABASE` + perintah `-d`). |
| `sqlite3: unable to open database file` | Folder file SQLite belum ada → `mkdir -p ./data` lalu jalankan ulang skemanya. |
| Judul tab masih "Go Core" | `APP_NAME` diubah di `.env` tapi frontend belum di-build ulang → `task build-frontend` + restart. |

---

# 6. Tutorial Menambah Menu

Panduan lengkap ada di:

```text
tutorial/README.md
```

Tutorial terbaru menjelaskan:

1. Membuat modul baru di `CPMATRIX` (UI **Modul & Menu**).
2. Mendaftarkan menu di `CPMENU` via UI yang sama (tanpa auto-grant).
3. Menambah folder `menus/<MCONTROL>/[<grup>/]<menu>/` (modul UPPERCASE).
4. Mengisi `index.jsx` + `api.js` (termasuk pola tabel, skeleton, paginasi).
5. Memahami halaman 404 pemandu sebagai kompas lokasi folder.
6. Menambah endpoint backend (opsional) dengan permission menu yang sama.
7. Membuat role baru, memindahkan user ke role itu, lalu memberi akses menu.
8. Login ulang dan memastikan menu otomatis muncul di grup modulnya.

---

# 7. Resep Cepat: Role, User, Modul, Menu

Semua langkah bisa lewat **UI** (tanpa SQL). Urutan tetap: modul → menu →
folder → role → grant.

| # | Tujuan | Langkah | Verifikasi |
|---|---|---|---|
| 1 | Role baru | **Role & Permission** → **Role baru** → kode `EDITOR` | `SELECT * FROM dbo.CPROLE WHERE CODE='EDITOR'` |
| 2 | User baru | **User Account** → **Tambah user** → pilih role `EDITOR` | `SELECT USERNAME, ROLE_CODE FROM dbo.CPUSER` |
| 3 | Modul baru | **Modul & Menu** → **Modul baru** → kode `TOKO` | `SELECT * FROM dbo.CPMATRIX` |
| 4 | Menu baru | **Modul & Menu** → **Menu baru** → `MENU_STOK`, modul `TOKO` | `SELECT * FROM dbo.CPMENU WHERE CODE='MENU_STOK'` |
| 5 | Isi menu | `cp -r tutorial/templates/frontend-menu frontend/src/menus/TOKO/stok`, lalu isi `index.jsx` + `api.js` | `npm run lint && npm run build` |
| 6 | Beri akses | **Role & Permission** → pilih `EDITOR` → centang `MENU_STOK` → **Simpan permission** | `SELECT * FROM dbo.CPPERMISSION WHERE MENU_CODE='MENU_STOK'` |
| 7 | Cek hasil | `task build` + restart, login user role `EDITOR` | Menu **Stok** muncul di sidebar modul `TOKO` |

Catatan penting:

- Kode menu **wajib** `MENU_` + huruf besar/angka/underscore, maksimal 40 karakter.
- Nama folder frontend **wajib sama** dengan kode menu setelah `MENU_`:
  `MENU_STOK` ↔ `menus/TOKO/stok/`.
- Menu baru **tidak** dapat diakses role mana pun sampai dicentang di langkah 6.
- User harus **logout/login ulang** setelah akses berubah.
- Menambah menu tanpa endpoint backend sendiri? Lewati saja — pakai endpoint
  yang sudah ada (mis. `/api/admin/menus`).
- Menghapus menu: UI **Modul & Menu** (grant ikut terhapus), atau lihat
  [tutorial bagian 7](tutorial/README.md).
