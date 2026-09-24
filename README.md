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
- Role `ADMIN` hanya dapat memakai menu admin yang diberi aksesnya
- Role non-ADMIN dapat memakai menu `user/` bila role diberi akses menu tersebut
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

Harus menghasilkan 8 permission:

```text
MENU_AUDIT
MENU_DASHBOARD
MENU_NOTIFICATIONS
MENU_ROLES
MENU_SECURITY
MENU_SESSIONS
MENU_SYSLOG
MENU_USERS
```

### Verifikasi akses admin dan user

```sql
SELECT ROLE_CODE, COUNT(*) AS MENU_COUNT
FROM dbo.CPROLEPERMISSION
WHERE PERMISSION_CODE LIKE 'MENU[_]%'
GROUP BY ROLE_CODE
ORDER BY ROLE_CODE;
```

Hasil database development:

```text
ADMIN  8
USER   1
```

User awal hanya mendapat `MENU_DASHBOARD`.

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

## 2.1. Dua Tipe Menu

```text
frontend/src/menus/admin/<menu>/  -> khusus role ADMIN
frontend/src/menus/user/<menu>/   -> semua role yang diberi akses menu
```

Frontend otomatis memindai folder tersebut melalui:

```text
frontend/src/menus/registry.js
```

Nama key menu otomatis dipetakan ke permission:

```text
laporan       -> MENU_LAPORAN
user-profile  -> MENU_USER_PROFILE
```

## 2.2. Alur Memberi Akses Menu

1. Admin membuat atau memilih role.
2. Admin_centang menu yang boleh diakses role tersebut.
3. Admin klik **Simpan permission**.
4. Permission tersimpan di `CPROLEPERMISSION`.
5. Saat login, frontend mengambil permission user dari `GET /api/users/me`.
6. Sidebar hanya menampilkan menu yang permission-nya dimiliki user.
7. Backend juga memeriksa permission yang sama pada endpoint menu.

Jadi, untuk menu `user/laporan`:

```text
Role EDITOR + MENU_LAPORAN  -> menu tampil
Role EDITOR tanpa akses     -> menu tidak tampil
ADMIN                       -> melihat menu admin yang diberi akses
```

Admin dapat langsung memberi akses menu ke role mana pun dari halaman
**Role & Permission**. Tidak perlu membuat folder menu per role.

## 2.3. Kenapa User Harus Login Ulang?

Permission user diambil saat login. Jika admin baru memberi atau mencabut
akses menu, user yang sedang login perlu:

1. Logout.
2. Login ulang.

Setelah login ulang, sidebar dan permission user baru dimuat.

---

# 3. Struktur Proyek Penting

```text
go-core/
├── main.go
├── routes/                 # registrasi endpoint
├── middleware/              # JWT, role ADMIN, permission menu
├── models/                  # konstanta permission MENU_* dan DTO
├── features/               # service, repository, handler backend
├── frontend/
│   ├── src/api/client.js   # request, login, session, getMe
│   ├── src/menus/
│   │   ├── registry.js     # auto-scan menu
│   │   ├── admin/          # menu khusus ADMIN
│   │   └── user/           # menu semua role sesuai permission
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
menus/admin|user/<menu>/api.js
    ↓
apiRequest dari api/client.js
    ↓
Authorization: Bearer <access_token>
    ↓
middleware NewAuth
    ↓
middleware RequireRole ADMIN (jika menu admin)
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
| Menu baru tidak muncul | Folder bukan `menus/user/<menu>/`, `index.jsx` tidak punya `export default`, `meta` belum diisi, atau `MENU_<MENU>` belum di-seed. |
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

1. Menambah menu `user/` untuk semua role.
2. Membuat permission `MENU_<MENU>`.
3. Seed permission ke database.
4. Mendaftarkan route backend dengan permission menu yang sama.
5. Mengatur akses dari halaman Role & Permission.
6. Login ulang dan memastikan menu otomatis muncul di sidebar.
7. Menghapus akses dan memastikan menu hilang kembali.
