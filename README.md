# Go Core

Aplikasi web **Go + React** untuk login, manajemen user, role, permission menu, audit, system log, dan notifikasi.

- Backend: Go + chi
- Frontend: React + Vite
- Database utama saat ini: **SQL Server**
- Development: backend `http://localhost:1067`, frontend `http://localhost:5173`
- Production: frontend dan API disajikan oleh satu `app.exe`

```text
Browser :5173 ── Vite proxy ──► app.exe :1067 ──► SQL Server database Go
Browser :1067 ─────────────────► app.exe :1067 ──► SQL Server database Go
```

## Fitur Utama

- Login tanpa registrasi publik
- User hanya dibuat oleh admin
- JWT access 15 menit
- Refresh token 7 hari dengan rotasi
- Maksimal 5 sesi per user
- RBAC sederhana: **satu permission untuk satu menu**
- Semua role dapat memakai menu bila role tersebut diberi akses `MENU_*`
- Module `ACCOUNT` untuk dashboard, module `SYSTEM` untuk menu operasional
- Tidak ada pembatasan berdasarkan nama role atau folder admin/user
- Role & Permission: daftar role di kiri, matriks akses menu di kanan
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
DB_DATABASE=Go
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

Buka SQL Server Management Studio atau gunakan `sqlcmd`.

### Membuat database kosong

```sql
IF DB_ID(N'Go') IS NULL
BEGIN
  CREATE DATABASE [Go];
END
GO
```

### Menjalankan migrasi aplikasi

Dari folder root repository:

```bash
task migrate
```

Task ini menjalankan:

1. `scripts/migrate.sql`
2. `scripts/migrate2_rbac.sql`

Tanpa Task CLI:

```bash
sqlcmd -S localhost,1433 -U may -P "password-database-anda" -d Go -C -i scripts/migrate.sql
sqlcmd -S localhost,1433 -U may -P "password-database-anda" -d Go -C -i scripts/migrate2_rbac.sql
```

Migrasi aman diulang. `migrate2_rbac.sql` juga membersihkan permission lama per-fitur dan membuat permission menu baru.

### Membuat akun awal

```bash
sqlcmd -S localhost,1433 -U may -P "password-database-anda" -d Go -C -i scripts/seed-admin.sql
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

Harus ada 9 tabel:

```text
CPAUDITLOG
CPNOTIFLOG
CPNOTIFTEMPLATE
CPPERMISSION
CPREFRESHTOKEN
CPROLE
CPROLEPERMISSION
CPSYSLOG
CPUSER
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

### Verifikasi permission menu

```sql
SELECT CODE, NAME, PERMGROUP
FROM dbo.CPPERMISSION
ORDER BY CODE;
```

Hasil yang benar:

```text
MENU_DASHBOARD  ACCOUNT
MENU_AUDIT      SYSTEM
MENU_NOTIFICATIONS SYSTEM
MENU_ROLES      SYSTEM
MENU_SECURITY   SYSTEM
MENU_SESSIONS   SYSTEM
MENU_SYSLOG     SYSTEM
MENU_USERS      SYSTEM
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
ADMIN  8
USER   1
```

`USER` hanya mendapat `MENU_DASHBOARD`. Setelah admin mencentang menu lain
untuk role USER, query yang sama dapat menunjukkan `USER 8`.

### Verifikasi user

```sql
SELECT CODE, USERNAME, EMAIL, ROLE_CODE
FROM dbo.CPUSER
ORDER BY USERNAME;
```

### Verifikasi template notifikasi

```sql
SELECT CODE, NAME, CHANNEL
FROM dbo.CPNOTIFTEMPLATE
ORDER BY CODE;
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

Setelah login:

1. Dashboard harus tampil.
2. Buka **Role & Permission**.
3. Role `ADMIN` dan `USER` harus terlihat.
4. Matriks harus menampilkan 8 menu, bukan permission per fungsi.
5. Centang menu untuk role yang membutuhkan.
6. Klik **Simpan permission**.
7. User dengan role tersebut harus logout/login ulang agar permission terbaru dimuat.

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

Folder pertama adalah module/kategori, bukan batas role:

```text
frontend/src/menus/account/<menu>/  -> MODULE ACCOUNT
frontend/src/menus/system/<menu>/   -> MODULE SYSTEM
```

Frontend otomatis memindai kedua module melalui `menus/registry.js`.
Module hanya mengelompokkan menu di matriks. Akses tetap ditentukan oleh
permission menu.

```text
menus/account/dashboard -> MENU_DASHBOARD
menus/system/users      -> MENU_USERS
menus/system/roles      -> MENU_ROLES
```

Tidak ada lagi folder `menus/admin` atau `menus/user` sebagai pembatas akses.
Semua role—ADMIN, USER, dan role custom—boleh memakai menu yang sama bila
role tersebut memiliki `MENU_*` yang sesuai.

## 2.2. Alur Memberi Akses Menu

1. Admin membuat atau memilih role.
2. Matriks menampilkan module `ACCOUNT` dan `SYSTEM`.
3. Admin centang menu yang boleh diakses role tersebut.
4. Admin klik **Simpan permission**.
5. Permission tersimpan di `CPROLEPERMISSION`.
6. Saat login, frontend mengambil permission user dari `GET /api/users/me`.
7. Sidebar hanya menampilkan menu yang permission-nya dimiliki user.
8. Backend juga memeriksa permission yang sama pada endpoint menu.

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
├── models/                 # konstanta permission MENU_* dan DTO
├── features/               # service, repository, handler backend
├── frontend/
│   ├── src/api/client.js   # request, login, session, getMe
│   ├── src/menus/
│   │   ├── registry.js     # auto-scan menu per module
│   │   ├── account/        # MODULE ACCOUNT
│   │   └── system/         # MODULE SYSTEM
│   ├── src/components/     # UI kit
│   └── src/pages/          # LoginForm
├── scripts/                # migrate*.sql dan build script
├── tutorial/               # tutorial menu baru
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
| Dashboard `user not found` | Restart backend dengan binary terbaru, logout/login ulang, dan pastikan database `Go` berisi user. |
| Login berhasil tetapi sidebar kosong | Permission role belum diberikan atau user belum login ulang → buka Role & Permission, centang menu, simpan, lalu login ulang. |
| Menu baru tidak muncul | Folder bukan `menus/<module>/<menu>/`, `index.jsx` tidak punya `export default`, `meta` belum diisi, atau `MENU_<MENU>` belum di-seed. |
| Menu baru muncul untuk semua role | Permission menu belum diberikan/di-filter dengan benar; cek `CPROLEPERMISSION` role tersebut. |
| Permission endpoint 403 | User belum memiliki `MENU_<MENU>` atau request dikirim ke role/menu yang salah. |
| `WARN frontend/dist missing` | Normal saat development; build frontend dengan `task build-frontend` jika ingin menghapus warning. |
| `localhost:1067` tidak bisa dibuka | Backend belum jalan → `task start` atau `go run .`. |
| `go:embed no matching files` | Jalankan `task build-frontend`, pastikan `frontend/dist/.gitignore` ada. |

---

# 6. Tutorial Menambah Menu

Panduan lengkap ada di:

```text
tutorial/README.md
```

Tutorial terbaru menjelaskan:

1. Menambah module `ACCOUNT` atau `SYSTEM`.
2. Menambah menu di `menus/<module>/<menu>/`.
3. Membuat permission `MENU_<MENU>`.
4. Seed permission ke database.
5. Mendaftarkan route backend dengan permission menu yang sama.
6. Mengatur akses dari halaman Role & Permission.
7. Login ulang dan memastikan menu otomatis muncul di sidebar.
