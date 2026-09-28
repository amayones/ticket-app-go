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
- JWT access 15 menit (default; atur via `ACCESS_TOKEN_MINUTES` di `.env`)
- Refresh token 7 hari dengan rotasi
- Maksimal 5 sesi per user
- RBAC sederhana: **satu permission untuk satu menu**
- Semua role dapat memakai menu bila role tersebut diberi akses `MENU_*`
- Master modul di tabel **`CPMODULE`**, registry menu di tabel **`CPMENU`**
  (kolom `MODULE` = folder modul UPPERCASE), grant role→menu di
  **`CPPERMISSION`**; matriks dibaca dari JOIN ketiga tabel
- Sidebar dikelompokkan per modul (tombol +/−); menu **parent/child**
  ditentukan database lewat `CPMENU.MENU_KIND` (`PARENT`/`CHILD`) +
  `CPMENU.PARENT_CODE`, bukan dari nama folder
- Di matriks akses, menu **PARENT** tampil sebagai header **tanpa centang**:
  yang dicentang hanya menu CHILD (grant ke kode PARENT ditolak 400), dan
  header induknya ikut ter-include otomatis bila minimal satu anaknya dicentang
- Menu terdaftar tapi folder belum dibuat tampil sebagai **halaman 404
  pemandu** (menunjukkan path persis), bukan hilang diam-diam
- Menu baru dibuat via UI Modul & Menu (tanpa auto-grant), dan menu yang
  sudah ada bisa diedit (label, urutan, modul, parent) tanpa dihapus-buat
- Hapus role = cascade: seluruh user pada role itu ikut terhapus, grant
  menunya hilang, refresh token dibersihkan
- Role tanpa akses apa pun (mis. `USER` baru) mendapat halaman kosong
- Tidak ada pembatasan berdasarkan nama role atau folder admin/user
- Role & Permission: daftar role di kiri, matriks akses menu di kanan
- Audit log, system log, dan notifikasi
- Popup login ulang ketika sesi habis tanpa pindah halaman
- Build frontend dan backend dalam satu binary
- Struktur menu mirror langit_v2: tiap folder `app/<mcontrol>/` berisi
  5 file (`index.jsx` shell, `controller.js` 6 fungsi, `GRID.jsx`,
  `FRM.jsx`, `api.js`); tambah menu = copy folder + ganti nama + isi
  kolom/field — tanpa menyentuh file lain

> Mau memakai untuk proyek Anda sendiri? Baca sesuai kebutuhan:
>
> | Tujuan | Langsung ke |
> |---|---|
> | Dari clone sampai aplikasi jalan | [1. Setup dari Nol](#1-setup-dari-nol-sampai-aplikasi-jalan) |
> | Mengganti nama aplikasi | [1.9](#19-mengganti-nama-aplikasi) |
> | Mengganti favicon & logo | [1.9 Mengganti favicon / logo](#19-mengganti-nama-aplikasi) |
> | Menambah role | [7. Resep Cepat](#7-resep-cepat-role-user-modul-menu) |
> | Menambah user | [7. Resep Cepat](#7-resep-cepat-role-user-modul-menu) |
> | Menambah modul | [tutorial 1.1](tutorial/README.md#bagian-0--dari-clone-sampai-aplikasi-jalan) |
> | Menambah menu (folder + halaman + akses) | [tutorial Bagian 1–4](tutorial/README.md#bagian-1--daftarkan-modul-dan-menu-di-database) |
> | Step-by-step lengkap dari clone sampai jadi | [tutorial lengkap](tutorial/README.md) |

---

# 1. Setup dari Nol sampai Aplikasi Jalan

Ikuti bagian ini secara berurutan.

## 1.0. Peta Jalan (Clone → Aplikasi Jadi)

| # | Langkah | Perintah utama | Selesai bila |
|---|---|---|---|
| 1 | [1.1](#11-prepare-komputer) Prepare komputer + clone | `git clone` | folder `go-core/` ada |
| 2 | [1.2](#12-install-dependency-frontend) Install dependency frontend | `task install` | `frontend/node_modules/` ada |
| 3 | [1.3](#13-membuat-file-env) Membuat `.env` | `copy .env.example .env` | `DB_DATABASE` + `JWT_SECRET` terisi |
| 4 | [1.4](#14-membuat-database-sql-server) Buat DB + migrasi + akun awal | `task migrate`, `sqlcmd -i scripts/seed-admin.sql` | 10 tabel `CP%` |
| 5 | [1.5](#15-verifikasi-database) Verifikasi data | query `CPMENU`/`CPPERMISSION` | 1 modul, 8 menu, 8 grant |
| 6 | [1.6](#16-menjalankan-development) Jalankan development | `task start` | login di `localhost:5173` |
| 7 | [1.7](#17-menjalankan-production) Build + jalankan production | `task build`, `./app.exe` | aplikasi di `localhost:1067` |
| 8 | [1.8](#18-health-check) Health check | `curl .../healthz` | `{"status":"ok"}` |
| 9 | [1.9](#19-mengganti-nama-aplikasi) Ganti nama & favicon | ubah `APP_NAME`/`APP_LOGO` | judul tab & logo ikut berubah |
| 10 | [7](#7-resep-cepat-role-user-modul-menu) Tambah role/user/modul/menu | UI atau SQL | menu baru tampil setelah login ulang |

Tutorial versi panjang (termasuk clone dari nol, ganti favicon, tambah
role/user/modul/menu, dan troubleshooting) ada di
[tutorial/README.md](tutorial/README.md).

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

> ✅ **Checkpoint 1.1** — `go version` ≥ 1.27, `node --version` ≥ 20,
> `sqlcmd -?` jalan. Folder `go-core/` ada.

### Task CLI tidak ditemukan (`task: command not found`)

`task` **bukan** perintah bawaan Git Bash — itu [Task CLI](https://taskfile.dev)
yang membaca `Taskfile.yml`. Kalau `task --version` gagal di Git Bash, pilih
salah satu:

**A. Install Task CLI (disarankan, sekali saja):**

```bash
winget install Task.Task          # Windows 10/11 (paling mudah)
# choco install go-task           # alternatif via Chocolatey
# scoop install task              # alternatif via Scoop
# npm i -g @go-task/cli           # alternatif via npm
task --version                    # tutup + buka ulang Git Bash bila masih gagal
```

**B. Tanpa install — pakai skrip langsung** (`Taskfile.yml` hanya alias
tipis ke `scripts/*.sh`, jadi hasilnya sama persis):

| Perintah `task` | Padanan tanpa Task CLI |
|---|---|
| `task install` | `cd frontend && npm ci && cd ..` |
| `task build-frontend` | `bash ./scripts/build-frontend.sh` |
| `task build` | `bash ./scripts/build-frontend.sh` lalu `bash ./scripts/build-backend.sh` |
| `task start` / `task dev` | `bash ./scripts/start.sh` |
| `task run` | `bash ./scripts/run.sh` |
| `task stop` | `bash ./scripts/stop.sh` |
| `task test` | `go vet ./...` lalu `go test -race ./...` |
| `task lint-frontend` | `cd frontend && npm run lint` |
| `task clean` | `bash ./scripts/clean.sh` |
| `task migrate` | tiga perintah `sqlcmd ... -i scripts/migrate.sql`, `migrate2_rbac.sql`, `migrate3_mcontrol.sql` (lihat 1.4) |

Dokumen ini menulis bentuk `task ...` agar ringkas; setiap ada tulisan
"Tanpa Task CLI" di bawahnya adalah perintah kolom kanan tabel di atas.

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

> ✅ **Checkpoint 1.2** — folder `frontend/node_modules/` ada dan perintah
> selesai tanpa error `EBADENGINE` (versi Node sesuai).

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
3. `scripts/migrate3_mcontrol.sql`

Tanpa Task CLI (ganti `<USER>`, `<PASSWORD>`, `<NAMA_DB>`):

```bash
sqlcmd -S localhost,1433 -U <USER> -P "<PASSWORD>" -d <NAMA_DB> -C -i scripts/migrate.sql
sqlcmd -S localhost,1433 -U <USER> -P "<PASSWORD>" -d <NAMA_DB> -C -i scripts/migrate2_rbac.sql
sqlcmd -S localhost,1433 -U <USER> -P "<PASSWORD>" -d <NAMA_DB> -C -i scripts/migrate3_mcontrol.sql
```

Migrasi aman diulang (boleh dijalankan berkali-kali). `migrate2_rbac.sql`
 membersihkan definisi permission lama per-fitur, lalu membuat master modul
(`CPMODULE`), registry menu (`CPMENU`), dan grant default `CPPERMISSION`.
`migrate3_mcontrol.sql` menambah kolom `CPMENU.MCONTROL` (nama folder frontend
per menu CHILD) berikut unique index dan CHECK `MENU_KIND`/`MCONTROL`.

> Database lama yang masih punya modul `REPORT` (`MENU_LAPORAN`,
> `MENU_ARUS_KAS`, `MENU_KEUANGAN`): jalankan `task migrate` sekali —
> migrasi otomatis menghapus grant, menu, lalu modulnya (idempoten).
> Setelah itu rebuild frontend (`task build-frontend`) karena folder
> `app/laporan/` dan `app/arus_kas/` sudah dihapus dari repo.

Catatan `QUOTED_IDENTIFIER`: karena `MCONTROL` memakai *filtered unique index*
(`UQ_CPMENU_MCONTROL`), SQL Server menolak INSERT/UPDATE ke `CPMENU` bila
`QUOTED_IDENTIFIER` OFF — dan default `sqlcmd` memang OFF. Skrip di `scripts/`
sudah men-set opsinya sendiri; hanya query manual ke `CPMENU` yang perlu flag
`-I`:

```bash
sqlcmd -S localhost,1433 -U <USER> -P "<PASSWORD>" -d <NAMA_DB> -C -I -Q "UPDATE dbo.CPMENU SET SORT_ORDER = 5 WHERE CODE = N'MENU_USERS'"
```

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

> ✅ **Checkpoint 1.4** — query berikut wajib 10 baris (lolos bila migrasi
> lengkap dan seed masuk):
>
> ```sql
> SELECT COUNT(*) AS TABEL_CP FROM INFORMATION_SCHEMA.TABLES
> WHERE TABLE_NAME LIKE 'CP%';                      -- harus 10
> SELECT COUNT(*) AS MENU FROM dbo.CPMENU;           -- fresh install: 8
> SELECT COUNT(*) AS AKUN FROM dbo.CPUSER;           -- harus 2 (admin, user)
> ```

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
CPMODULE
CPNOTIFLOG
CPNOTIFTEMPLATE
CPPERMISSION
CPREFRESHTOKEN
CPROLE
CPSYSLOG
CPUSER
```

### Verifikasi modul (CPMODULE)

```sql
SELECT CODE, LABEL, SORT_ORDER
FROM dbo.CPMODULE
ORDER BY SORT_ORDER;
```

Hasil yang benar (1 modul bawaan; tambah modul baru via UI **Modul & Menu**):

```text
SYSTEM  System  1
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
SELECT CODE, MODULE, LABEL, MENU_KIND, SORT_ORDER, PARENT_CODE
FROM dbo.CPMENU
ORDER BY MODULE, SORT_ORDER;
```

Hasil yang benar (8 baris, semuanya `SYSTEM`):

```text
CODE                MODULE  LABEL            KIND    URUTAN  PARENT_CODE
MENU_USERS          SYSTEM  User Account     CHILD   1       NULL
MENU_MODUL          SYSTEM  Modul & Menu     CHILD   2       NULL
MENU_ROLES          SYSTEM  Role & Permission CHILD  3       NULL
MENU_SESSIONS       SYSTEM  Sesi & Auth      CHILD   4       NULL
MENU_AUDIT          SYSTEM  Audit Log        CHILD   5       NULL
MENU_SECURITY       SYSTEM  Security Center  CHILD   6       NULL
MENU_SYSLOG         SYSTEM  System Log       CHILD   7       NULL
MENU_NOTIFICATIONS  SYSTEM  Notifikasi       CHILD   8       NULL
```

Tiga kolom penting di `CPMENU`:

| Kolom | Isi |
|---|---|
| `MODULE` | Section sidebar (FK ke `CPMODULE.CODE`); label section = `CPMODULE.LABEL` |
| `MCONTROL` | Nama folder frontend `app/<mcontrol>/` (snake_case); CHILD wajib isi, PARENT wajib NULL |
| `MENU_KIND` | `PARENT` (header buka-tutup, tanpa folder/halaman) atau `CHILD` (item biasa) |
| `PARENT_CODE` | Kode menu parent; hanya boleh menunjuk menu `PARENT` satu modul yang sama |

### Verifikasi grant role -> menu (CPPERMISSION)

Satu-satunya tabel relasi: role boleh tampil menu apa. Definisi menu
tinggal di `CPMENU` (tidak ada tabel definisi terpisah).

```sql
SELECT ROLE_CODE, MENU_CODE
FROM dbo.CPPERMISSION
ORDER BY ROLE_CODE, MENU_CODE;
```

Hasil yang benar (8 baris `ADMIN`):

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
via matriks. Hanya menu **CHILD** yang di-grant: header `PARENT` tidak punya
baris di sini karena tidak punya halaman sendiri. Setelah admin mencentang
menu CHILD untuk suatu role, query JOIN berikut menampilkannya (modul
otomatis ketahuan dari `CPMENU`):

```sql
SELECT g.ROLE_CODE, m.MODULE, g.MENU_CODE
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

Setelah login (admin bawaan memegang 8 menu `SYSTEM` — perhatikan bahwa
mencentang di matriks lalu **Simpan permission** akan memberi akses ke role itu):

1. Menu pertama otomatis terbuka, sidebar dikelompokkan per modul (SYSTEM).
2. Buka **Role & Permission**.
3. Role `ADMIN` dan `USER` harus terlihat.
4. Matriks menampilkan 8 menu, bukan permission per fungsi. Menu `PARENT`
   (bila Anda membuatnya nanti) tampil sebagai **header tanpa centang**;
   hanya menu `CHILD` yang punya checkbox.
5. Menu **Modul & Menu** menampilkan 1 modul + 8 baris `CPMENU`.
6. Centang menu CHILD untuk role yang membutuhkan (tanpa auto-grant).
7. Klik **Simpan permission**.
8. User dengan role tersebut harus logout/login ulang agar permission terbaru dimuat.
9. Login sebagai `user`/`user` (nol menu) → halaman kosong "hubungi admin".

> ✅ **Checkpoint 1.6** — login `admin`/`admin` berhasil, sidebar
> menampilkan grup **System** (8 menu). `user`/`user` login berhasil tapi
> sidebar kosong + pesan "hubungi admin" (artinya permission bekerja).

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

> ✅ **Checkpoint 1.7** — halaman login tampil di `:1067` tanpa akses
> internet (semua aset lokal dari binary), dan menu-menu bawaan (User
> Account, Modul & Menu, Role & Permission, Sesi & Auth, Audit Log,
> Security Center, System Log, Notifikasi) semuanya bisa dibuka.

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

> ✅ **Checkpoint 1.8** — `curl http://localhost:1067/healthz` persis
> `{"status":"ok"}`. Bila gagal: backend belum jalan atau `APP_PORT`
> di `.env` beda dengan URL yang dibuka.

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

Satu file untuk dua tempat: **favicon tab browser** dan **logo di sidebar,
topbar mobile, serta kartu login**.

Langkah praktis:

1. Siapkan file gambar: rasio **1:1** (persis persegi), minimal 64×64 px,
   latar transparan atau putih. Format bebas (SVG, PNG, WebP).
2. Simpan di `frontend/public/`, misalnya `frontend/public/logo-toko.png`.
   Folder `public/` disalin apa adanya ke `dist/`, jadi file di sana bisa
   dipakai langsung tanpa import.
3. Arahkan `APP_LOGO` di `.env` ke nama file tersebut (harus diawali `/`):

```env
APP_LOGO=/logo-toko.png
```

4. Build ulang frontend lalu restart backend:

```bash
task build-frontend
```

5. Hard refresh browser (`Ctrl+Shift+R`).

Cek cepat: judul tab, logo sidebar, topbar mobile, dan kartu login semuanya
sudah memakai gambar/nama Anda. Kalau file gambar rusak atau hilang,
`AppLogo` otomatis jatuh ke huruf pertama `APP_NAME`.

Cache browser untuk favicon sangat agresif. Kalau logo tidak berubah setelah
ganti file, pakai nama file baru (mis. `favicon-toko.svg`) atau hard refresh
(`Ctrl+Shift+R`).

Langkah nomor + tabel masalah umum ada di tutorial bagian
[0.9. Mengganti favicon dan logo](tutorial/README.md#09-mengganti-favicon-dan-logo).

### Branding lain yang tidak ikut `.env`

| File | Isi |
|---|---|
| `frontend/index.html` | `<title>` + `<link rel="icon">` fallback |
| `README.md` | Judul & deskripsi proyek |
| `Dockerfile`, `Taskfile.yml` | Nama binary bila ingin `toko.exe` |
| Tabel `CPNOTIFTEMPLATE` | Teks email (bisa diedit dari menu **Notifikasi**) |

Nama modul bawaan (`SYSTEM`) dan kode tabel `CP*` tidak perlu
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

Ketiganya menghasilkan state awal yang sama: 10 tabel, 1 modul, 8 menu,
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

Sidebar 100% dari database (MODULE section -> PARENT header -> CHILD item).
Folder frontend DATAR: satu folder = satu menu CHILD, tanpa folder modul:

```text
frontend/src/app/users/           -> MCONTROL users           (MENU_USERS)
frontend/src/app/role_permission/ -> MCONTROL role_permission (MENU_ROLES)
frontend/src/app/stok/            -> MCONTROL stok            (MENU_STOK, contoh menu baru)
```

`CPMENU.MODULE` hanya menentukan section sidebar (FK ke `CPMENU` ->
`CPMODULE.CODE`; label section = `CPMODULE.LABEL`). Menu PARENT
TANPA folder/mcontrol — hanya header buka-tutup.

Frontend memindai folder datar via `app/registry.js` (glob `./*/index.jsx`)
lalu menggabungkannya dengan entri DB (`buildSidebar`). Akses tetap
ditentukan grant `CPPERMISSION` + baris `CPMENU`; PARENT ikut tampil otomatis
bila minimal satu keturunannya ter-grant (tanpa grant sendiri).

```text
MENU_USERS         (SYSTEM, MCONTROL users)            -> app/users/
MENU_ROLES         (SYSTEM, MCONTROL role_permission)  -> app/role_permission/
MENU_STOK          (TOKO, MCONTROL stok)               -> app/stok/ (contoh menu baru)
MENU_KEUANGAN      (TOKO, PARENT, tanpa mcontrol)      -> header saja (contoh, bila dibuat)
```

Tidak ada lagi folder `app/account` atau menu dashboard. Semua
role—ADMIN, USER, dan role custom—boleh memakai menu yang sama bila role
tersebut memiliki `MENU_*` yang sesuai; role tanpa akses mendapat halaman
kosong.

## 2.2. Alur Memberi Akses Menu

1. Admin membuat menu via **Modul & Menu** (atau memilih menu bawaan).
   Menu yang sudah ada bisa diedit di halaman yang sama: label, mcontrol,
   urutan, modul, dan parent. Kode permission tidak bisa diubah — buat menu baru
   bila perlu.
2. Admin membuat atau memilih role.
3. Matriks menampilkan menu per module (`SYSTEM`, `TOKO`, ...). Menu
   `PARENT` tampil sebagai **header tanpa centang**; hanya menu `CHILD` di
   bawahnya yang punya checkbox.
4. Admin centang menu CHILD yang boleh diakses role tersebut.
5. Admin klik **Simpan permission**.
6. Grant tersimpan di `CPPERMISSION` (ROLE_CODE -> MENU_CODE) — hanya untuk
   menu CHILD. Grant ke kode `PARENT` ditolak `400` (`invalid role: PARENT menu
   is a header without its own page, grant its child menus instead`), dan
   baris grant PARENT sisa instalasi lama diabaikan/dibersihkan migrasi.
7. Saat login, frontend mengambil permission + entri menu user dari `GET /api/users/me`.
8. Sidebar hanya menampilkan menu yang permission-nya dimiliki user; parent
   header ikut ter-include otomatis dari anaknya, sehingga yang tampil tetap
   hanya anak yang dicentang. Menu terdaftar tapi folder belum dibuat tampil
   sebagai halaman 404 pemandu.
9. Backend juga memeriksa permission yang sama pada endpoint menu.

Endpoint matriks (`GET /api/admin/matrix`) menandai aturan itu: baris `PARENT`
punya `"grantable": false` dan `has_access`-nya `true` begitu ada anak/keturunan
yang ter-grant.

Contoh:

```text
Role EDITOR + MENU_STOK           -> menu Stok tampil
Role EDITOR + MENU_ARUS_KAS       -> header Keuangan + menu Arus Kas tampil (contoh parent-child)
Role EDITOR + MENU_KEUANGAN saja  -> ditolak 400 (header tidak bisa di-grant;
                                     centang menu anaknya)
Role EDITOR tanpa akses           -> menu Stok tidak tampil
Role ADMIN                        -> semua menu CHILD yang dicentang
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
│   ├── src/app/
│   │   ├── registry.js     # scan datar app/<mcontrol>/ + buildSidebar 3 level + 404 pemandu
│   │   ├── shared/         # padanan COMP.* : controller.js, messageBox.jsx,
│   │   │                   # validateInput.js, DownloadButton.jsx, StandardPage/
│   │   │                   # Grid/Form, useStandardController.js
│   │   ├── users/          # contoh menu CRUD penuh:
│   │   │   ├── index.jsx     # shell (toolbar + items) = <mod>.js
│   │   │   ├── controller.js # 6 fungsi (init/renderpage/btrefresh/btnew/...) = C<mod>.js
│   │   │   ├── GRID.jsx      # kolom + handler_rowbtn_* = GRID<mod>.js
│   │   │   ├── FRM.jsx       # form + validate_field = FRM<mod>.js
│   │   │   └── api.js        # read_data/process_* = store proxy
│   │   ├── ...             # tiap CHILD = 1 folder 5 file; PARENT tanpa folder
│   ├── src/components/     # UI kit (termasuk MissingMenu)
│   └── src/pages/          # LoginForm
├── scripts/                # migrate*.sql (SQL Server) dan build script
├── tutorial/               # tutorial menu baru (SQL Server)
├── Taskfile.yml
└── README.md
```

## 3.1. Alur Request

```text
Frontend menu (index.jsx shell → controller.btrefresh_click/btnew_click)
    ↓
app/<mcontrol>/api.js  (read_data / process_create / process_update / process_delete
                        = padanan method proxy langit_v2)
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

> ✅ **Checkpoint alur**: buka DevTools → tab Network → klik menu →
> pastikan request `GET /api/...?limit=20&offset=0` kembali `200` dengan
> array JSON. Bila `401`, token habis (login ulang). Bila `403`,
> role belum di-grant `MENU_*` (beri akses di matriks, login ulang).

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
| `Tercatat 0 dari 0 menu CHILD` | `CPPERMISSION` masih kosong/permission lama belum dimigrasi → jalankan `task migrate`. |
| Login berhasil tetapi sidebar kosong | Wajar bila role memang nol menu (mis. `USER` baru) → buka Role & Permission, centang menu, simpan, lalu login ulang. |
| `Msg 1934 ... incorrect settings: 'QUOTED_IDENTIFIER'` | Query manual ke `CPMENU` (punya filtered unique index `UQ_CPMENU_MCONTROL`) dijalankan dari `sqlcmd` yang default `QUOTED_IDENTIFIER` OFF → tambahkan flag `-I` (atau `SET QUOTED_IDENTIFIER ON;`). Skrip di `scripts/` sudah men-set sendiri. |
| Menu terdaftar tapi tampil 404 | Folder `frontend/src/app/<mcontrol>/` belum dibuat → ikuti petunjuk di halaman 404 (copy template, `npm run build`, restart). |
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

1. Membuat modul baru di `CPMODULE` (UI **Modul & Menu**).
2. Mendaftarkan menu di `CPMENU` via UI yang sama (tanpa auto-grant).
3. Menambah folder `app/<mcontrol>/` (datar, MCONTROL snake_case).
4. Mengisi 5 file mirror (`index.jsx`, `controller.js`, `GRID.jsx`,
   `FRM.jsx`, `api.js`) — cukup isi `COLUMNS_ITEMS` + `validate_field`.
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
| 2 | User baru | **User Account** → **New Input** → pilih role `EDITOR` | `SELECT USERNAME, ROLE_CODE FROM dbo.CPUSER` |
| 3 | Modul baru | **Modul & Menu** → **Modul baru** → kode `TOKO` | `SELECT * FROM dbo.CPMODULE` |
| 4 | Menu baru | **Modul & Menu** → **Menu baru** → `MENU_STOK`, modul `TOKO`, mcontrol `stok`, jenis `CHILD`/`PARENT` | `SELECT * FROM dbo.CPMENU WHERE CODE='MENU_STOK'` |
| 5 | Isi menu | `cp -r tutorial/templates/frontend-menu frontend/src/app/stok`, ganti `<mcontrol>` di 5 file, isi `COLUMNS_ITEMS` + `validate_field` | `npm run lint && npm run build` tanpa error |
| 6 | Beri akses | **Role & Permission** → pilih `EDITOR` → centang `MENU_STOK` → **Simpan permission** | `SELECT * FROM dbo.CPPERMISSION WHERE MENU_CODE='MENU_STOK'` |
| 7 | Cek hasil | `task build` + restart, login user role `EDITOR` | Menu **Stok** muncul di sidebar modul `TOKO` |

Catatan penting:

- Kode menu **wajib** `MENU_` + huruf besar/angka/underscore, maksimal 40 karakter.
- Folder frontend **datar** dan mengikuti `CPMENU.MCONTROL` (snake_case):
  `MENU_STOK` + mcontrol `stok` ↔ `frontend/src/app/stok/`. Tidak ada folder
  modul/perantara; `MODULE` hanya section sidebar.
- Menu `PARENT` = header buka-tutup: tanpa `MCONTROL`, tidak butuh folder, dan
  **tidak punya centang** di matriks (yang dicentang hanya CHILD); header itu
  ikut ter-include otomatis bila minimal satu anaknya dicentang.
- Menu child butuh `PARENT_CODE` yang menunjuk menu `PARENT` di modul yang
  sama; menu `PARENT` tidak boleh punya parent dan tidak bisa diubah jadi
  `CHILD` selama masih punya anak.
- Menu baru **tidak** dapat diakses role mana pun sampai dicentang di langkah 6.
- User harus **logout/login ulang** setelah akses berubah.
- Ganti password user: **User Account** → ikon edit → kolom **Password baru**
  (kosongkan bila tidak diganti). Untuk memaksa user keluar, klik tombol
  **K mengeluarkan semua sesi**. Detail + aturan validasi ada di
  [tutorial 5.5–5.6](tutorial/README.md#55-admin-mengganti-password-user).
- Menambah menu tanpa endpoint backend sendiri? Lewati saja — pakai endpoint
  yang sudah ada (mis. `/api/admin/menus`).
- Menghapus menu: UI **Modul & Menu** (grant ikut terhapus), atau lihat
  [tutorial bagian 7](tutorial/README.md).
