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
- Registry menu di tabel **`CPMENU`** (`MCONTROL` = folder modul UPPERCASE);
  matriks role × modul × menu dibaca dari **view `CPMATRIX`**
- Sidebar dikelompokkan per modul (tombol +/−), mendukung parent bersarang
- Menu terdaftar tapi folder belum dibuat tampil sebagai **halaman 404
  pemandu** (menunjukkan path persis), bukan hilang diam-diam
- Menu baru dibuat via UI Role & Permission → Registry Menu (tanpa auto-grant)
- Role tanpa akses apa pun (mis. `USER` baru) mendapat halaman kosong
- Tidak ada pembatasan berdasarkan nama role atau folder admin/user
- Role & Permission: daftar role di kiri, matriks akses menu di kanan
- Audit log, system log, dan notifikasi
- Popup login ulang ketika sesi habis tanpa pindah halaman
- Build frontend dan backend dalam satu binary

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
```

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

Migrasi aman diulang. `migrate2_rbac.sql` juga membersihkan permission lama per-fitur dan membuat permission menu baru.

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
CPMENU
CPNOTIFLOG
CPNOTIFTEMPLATE
CPPERMISSION
CPREFRESHTOKEN
CPROLE
CPROLEPERMISSION
CPSYSLOG
CPUSER
```

Plus 1 view matriks:

```sql
SELECT TABLE_NAME
FROM INFORMATION_SCHEMA.VIEWS
WHERE TABLE_NAME = 'CPMATRIX';
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

Hasil yang benar (7 menu bawaan, modul `SYSTEM`):

```text
MENU_USERS          SYSTEM  User Account       1
MENU_ROLES          SYSTEM  Role & Permission  2
MENU_SESSIONS       SYSTEM  Sesi & Auth        3
MENU_AUDIT          SYSTEM  Audit Log          4
MENU_SECURITY       SYSTEM  Security Center    5
MENU_SYSLOG         SYSTEM  System Log         6
MENU_NOTIFICATIONS  SYSTEM  Notifikasi         7
```

### Verifikasi permission menu

```sql
SELECT CODE, NAME, PERMGROUP
FROM dbo.CPPERMISSION
ORDER BY CODE;
```

Hasil yang benar (7 baris, tanpa `MENU_DASHBOARD`):

```text
MENU_AUDIT        SYSTEM
MENU_NOTIFICATIONS SYSTEM
MENU_ROLES        SYSTEM
MENU_SECURITY     SYSTEM
MENU_SESSIONS     SYSTEM
MENU_SYSLOG       SYSTEM
MENU_USERS        SYSTEM
```

### Verifikasi akses admin dan user

```sql
SELECT ROLE_CODE, COUNT(*) AS MENU_COUNT
FROM dbo.CPROLEPERMISSION
WHERE PERMISSION_CODE LIKE 'MENU[_]%'
GROUP BY ROLE_CODE
ORDER BY ROLE_CODE;
```

Pada fresh install, hasil default:

```text
ADMIN  7
```

`USER` tidak memiliki baris (nol menu) — akses diberikan manual oleh admin
via matriks. Setelah admin mencentang menu untuk suatu role, view `CPMATRIX`
menampilkannya:

```sql
SELECT ROLE_CODE, MODULE, MENU_CODE, HAS_ACCESS
FROM dbo.CPMATRIX
ORDER BY ROLE_CODE, SORT_ORDER;
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

Setelah login (admin masih memegang 7 menu):

1. Menu pertama otomatis terbuka, sidebar dikelompokkan per modul (SYSTEM).
2. Buka **Role & Permission**.
3. Role `ADMIN` dan `USER` harus terlihat.
4. Matriks harus menampilkan 7 menu, bukan permission per fungsi.
5. Kartu **Registry Menu** menampilkan 7 baris `CPMENU`.
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

1. Admin membuat menu via **Registry Menu** (atau memilih menu bawaan).
2. Admin membuat atau memilih role.
3. Matriks menampilkan menu per module (`SYSTEM`, `REPORT`, ...).
4. Admin centang menu yang boleh diakses role tersebut.
5. Admin klik **Simpan permission**.
6. Permission tersimpan di `CPROLEPERMISSION` (terbaca via view `CPMATRIX`).
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
│   │   └── SYSTEM/         # MODULE SYSTEM (7 menu bawaan)
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
| Menu baru muncul untuk semua role | Permission menu belum diberikan/di-filter dengan benar; cek `CPROLEPERMISSION` role tersebut. |
| Permission endpoint 403 | User belum memiliki `MENU_<MENU>` atau request dikirim ke role/menu yang salah. |
| `WARN frontend/dist missing` | Normal saat development; build frontend dengan `task build-frontend` jika ingin menghapus warning. |
| `localhost:1067` tidak bisa dibuka | Backend belum jalan → `task start` atau `go run .`. |
| `go:embed no matching files` | Jalankan `task build-frontend`, pastikan `frontend/dist/.gitignore` ada. |
| Koneksi DB gagal saat start | `DB_DATABASE` di `.env` tidak sama dengan nama database di SQL Server, atau kredensial salah → samakan ketiganya (`DB_HOST/DB_PORT/DB_DATABASE` + perintah `-d`). |

---

# 6. Tutorial Menambah Menu

Panduan lengkap ada di:

```text
tutorial/README.md
```

Tutorial terbaru menjelaskan:

1. Mendaftarkan menu di `CPMENU` via UI Registry Menu (tanpa auto-grant).
2. Menambah folder `menus/<MCONTROL>/[<parent>/]<menu>/` (modul UPPERCASE).
3. Memahami halaman 404 pemandu sebagai kompas lokasi folder.
4. Membuat permission `MENU_<MENU>` (satu permission per menu).
5. Mendaftarkan route backend dengan permission menu yang sama.
6. Mengatur akses dari halaman Role & Permission (matriks per modul).
7. Login ulang dan memastikan menu otomatis muncul di grup modulnya.
