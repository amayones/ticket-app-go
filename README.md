# go-core

Aplikasi web **Go + React dalam satu binary**: backend Go (chi) + frontend React (Vite)
digabung via `//go:embed`, sehingga deploy production cukup membawa
**3 file**: `app.exe` + `stop.exe` + `.env`.

```
Browser ──► app.exe :1067 ──┬──► /               (frontend React, hasil build Vite)
                             └──► /api/*          (REST API JSON)
                             └──► /healthz        (health check)
                                      │
                                      ▼
                               SQL Server (database Go)
```

**Fitur utama:** register/login user, JWT access 15 menit + refresh token rotasi
7 hari (disimpan sebagai hash SHA-256), rate-limit, single-binary embed,
skrip build satu pintu (`Taskfile.yml`), CI + Dockerfile siap pakai.

---

## Daftar Isi

0. [Panduan Memakai Aplikasi — Klik per Klik](#0-panduan-memakai-aplikasi--klik-per-klik)
1. [Mulai Cepat — Langkah demi Langkah](#1-mulai-cepat--langkah-demi-langkah)
2. [Memahami Isi Proyek (Tur Folder)](#2-memahami-isi-proyek-tur-folder)
3. [Cara Kerja Sistem (Alur Penting)](#3-cara-kerja-sistem-alur-penting)
4. [Referensi API](#4-referensi-api)
5. [Referensi Perintah & Konfigurasi](#5-referensi-perintah--konfigurasi)
6. [Deploy Production](#6-deploy-production)
7. [Troubleshooting & FAQ](#7-troubleshooting--faq)

---

## 0. Panduan Memakai Aplikasi — Klik per Klik

Bagian ini untuk yang **baru pertama kali membuka aplikasi**: dari browser kosong
sampai bisa login, daftar, dan mengelola user. Baca bagian ini saja sampai bisa.

### 0.1 Buka aplikasinya di browser

Pastikan backend + frontend sudah jalan (cara menjalankannya ada di
[Bagian 1](#1-mulai-cepat--langkah-demi-langkah)). Lalu buka alamat berikut:

| Mode | Alamat yang dibuka | Keterangan |
|------|--------------------|------------|
| Development | `http://localhost:5173` | Frontend Vite (HMR, otomatis reload tiap simpan file) |
| Production | `http://localhost:1067/` | Satu binary `app.exe` menyajikan frontend + API sekaligus |

> Port `1067` bisa berbeda bila `APP_PORT` di `.env` diubah. Vite otomatis
> mengikuti port tersebut untuk proxy `/api`, jadi tidak perlu setting manual.

Saat halaman terbuka Anda melihat **header "Go Core"** di atas dan sebuah kartu
form di tengah. Ada 2 tab di kanan header: **Login** dan **Register**.

### 0.2 Daftar akun baru (tab Register)

1. Klik tab **Register** di kanan atas.
2. Isi 3 kolom:
   - **Username** — minimal 3 karakter, contoh: `budi`
   - **Email** — harus format email valid, contoh: `budi@example.com`
   - **Password** — minimal 8 karakter (maks 72). Klik **ikon mata** di kanan
     kolom untuk mengintip/menyembunyikan ketikan.
3. Klik tombol **Register** (tombol menampilkan animasi loading saat diproses).
4. Hasilnya:
   - **Berhasil** → muncul **notifikasi toast hijau** "Akun dibuat" di kanan
     atas layar, lalu otomatis pindah ke tab **Login**.
   - **Gagal** (mis. username sudah dipakai) → muncul **kotak peringatan merah**
     di dalam form. Kotak ini bisa ditutup dengan tombol **×** di kanannya.

### 0.3 Masuk ke aplikasi (tab Login)

1. Isi **Username** dan **Password** akun yang tadi didaftarkan.
2. Klik **Login**.
3. **Berhasil** → muncul toast hijau "Selamat datang kembali!" dan halaman
   berganti ke **Daftar Pengguna**. Header kini menampilkan tombol
   **Users** dan **Logout**.
4. **Gagal** (salah password) → kotak merah "Login gagal" muncul di form dan
   bisa di-close dengan tombol ×.

### 0.4 Halaman Daftar Pengguna (Users)

Halaman ini menampilkan semua akun dalam bentuk **kartu modern**:

- Setiap baris punya **avatar lingkaran** (inisial username, warna gradien
  berbeda tiap nama), username dengan awalan `@`, dan email di bawahnya.
- Baris milik **akun Anda sendiri** ditandai lencana ungu **"Anda"** dan punya
  3 tombol aksi di kanan:
  | Tombol | Fungsi |
  |--------|--------|
  | ✏️ Pensil | **Edit profil** — membuka dialog: ubah username/email, atau isi password baru (kosongkan bila tidak diganti). Klik **Simpan perubahan**. |
  | 🚪 Pintu | **Keluarkan semua sesi** — minta konfirmasi, lalu semua perangkat yang login sebagai Anda dikeluarkan (harus login ulang). |
  | 🗑️ Sampah | **Hapus akun** — dialog konfirmasi merah. Bila menghapus **akun sendiri**, Anda otomatis keluar dan kembali ke halaman Login. |
- **Pagination** — data tampil 10 per halaman. Pakai tombol **Sebelumnya /
  Berikutnya** di bawah daftar. Teks "Menampilkan 1–10" menunjukkan posisi.
- **Muat ulang** — tombol di kanan judul untuk memuat ulang data (ada animasi
  *skeleton* saat memuat dan tampilan khusus bila data kosong).
- Semua aksi penting menampilkan **toast**: hijau bila sukses, merah bila gagal
  (toast hilang sendiri setelah ±4 detik, atau klik **×** untuk menutup manual).

### 0.5 Keluar dari aplikasi (Logout)

Klik tombol **Logout** di header → sesi berakhir, muncul toast "Anda telah
keluar", dan kembali ke halaman Login.

### 0.6 Menu-menu admin (setelah login)

Setelah login Anda masuk ke aplikasi **sidebar** (layar lebar) atau navigasi
atas (layar HP). Menu yang tampil tergantung role:

| Menu | Untuk | Isi & cara pakai |
|------|-------|------------------|
| **User Account** | semua | Kartu user + avatar + badge role (lihat 0.4). Admin bisa mengganti role lewat dialog Edit (dropdown Role). |
| **Role & Permission** | ADMIN | Pilih role (tombol kiri) → centang permission per grup di matriks → **Simpan permission**. **Role baru** (kode huruf besar, mis. `EDITOR`) → atur permission-nya → user bisa dipindah ke role itu. Role `ADMIN`/`USER` bawaan tidak bisa dihapus; role yang masih dipakai user tidak bisa dihapus. |
| **Sesi & Auth** | semua | Tab **Sesi saya**: daftar perangkat login + tombol sampah untuk mencabut satu sesi + tombol cabut semua. Tab **Semua sesi** (ADMIN): semua user + pagination. |
| **Audit Log** | ADMIN | Tabel siapa–apa–kapan–IP. Filter: aksi (LOGIN, DELETE_USER, …), entitas, kode pelaku + tombol Filter/Reset + pagination. |
| **Security Center** | ADMIN | 6 kartu ringkasan (user, role, sesi aktif, audit 24 jam, error 24 jam, status), aktivitas terkini, dan daftar kebijakan keamanan aktif. |
| **System Log** | ADMIN | Filter level SEMUA/ERROR/WARN/INFO + tabel + tombol **Bersihkan lama** (hapus log > N hari, tercatat di audit). |
| **Notifikasi** | ADMIN | Tab **Template**: buat/edit/nonaktifkan/hapus template (channel EMAIL/PUSH/INAPP, variabel `{{nama}}` `{{kode}}` `{{role}}` `{{detail}}`). Tombol **Kirim/Test**: pilih template + penerima + isi variabel → tercatat di tab **Riwayat kirim**. |

### 0.7 Hal-hal modern yang bisa dicoba

- **Toast** bisa ditutup manual (tombol ×) atau biarkan hilang otomatis
  (ada garis progres di bawahnya).
- **Dialog** (edit/konfirmasi) bisa ditutup dengan tombol ×, klik area gelap di
  luar dialog, atau tombol **Esc** di keyboard.
- **Dark mode** mengikuti pengaturan sistem operasi/laptop Anda secara otomatis
  (coba ubah Windows ke Dark mode → tampilan ikut gelap).
- Tampilan responsif: buka di HP, kartu dan tombol menyesuaikan layar kecil.

---

## 1. Mulai Cepat — Langkah demi Langkah

Ikuti langkah 1–6 berurutan. Estimasi: ±15 menit untuk pemula.

### Langkah 0 — Siapkan prasyarat

| Kebutuhan | Versi minimal | Cara cek |
|-----------|---------------|----------|
| Go | 1.27 | `go version` |
| Node.js | 20 (lihat `.nvmrc`) | `node --version` |
| npm | 9+ | `npm --version` |
| SQL Server | 2019+ / Express juga bisa | SSMS / `sqlcmd` konek |
| Git | bebas | `git --version` |
| `task` (opsional) | 3.x | `task --version` — kalau belum ada, semua perintah `task x` bisa diganti `pwsh ./scripts/x.ps1` |

> SQL Server harus **sudah berjalan** sebelum backend dijalankan
> (cek via `services.msc` → `SQL Server (MSSQLSERVER)` → Running).

### Langkah 1 — Clone dan masuk folder

```powershell
git clone <url-repo>.git
cd go-core
```

### Langkah 2 — Buat file `.env`

```powershell
copy .env.example .env
notepad .env
```

Isi yang **wajib diisi** (sisanya boleh default):

```ini
DB_USERNAME=...        # user SQL Server kamu
DB_PASSWORD=...        # password SQL Server kamu
JWT_SECRET=...         # minimal 32 karakter acak (lihat bawah)
```

Buat `JWT_SECRET` acak (pilih salah satu):

```powershell
openssl rand -hex 32
```

Daftar lengkap variabel ada di [tabel konfigurasi](#51-variabel-environment).

### Langkah 3 — Buat database + tabel (sekali saja)

Buka SSMS / `sqlcmd`, jalankan:

Skema memakai tabel **huruf besar prefix `CP`** (core app) dengan relasi via
**CODE** (kolom `ID` tetap ada sebagai IDENTITY tapi tidak dipakai relasi/API):

```sql
CREATE DATABASE Go;
GO
USE Go;
GO
CREATE TABLE CPROLE (
  ID INT IDENTITY(1,1) PRIMARY KEY,
  CODE NVARCHAR(20) NOT NULL UNIQUE,   -- 'ADMIN', 'USER'
  NAME NVARCHAR(100) NOT NULL,
  CREATED_AT DATETIME NOT NULL DEFAULT GETDATE(),
  UPDATED_AT DATETIME NOT NULL DEFAULT GETDATE()
);
INSERT INTO CPROLE (CODE, NAME) VALUES ('ADMIN', 'Administrator'), ('USER', 'Pengguna');
CREATE TABLE CPUSER (
  ID INT IDENTITY(1,1) PRIMARY KEY,
  CODE NVARCHAR(20) NOT NULL UNIQUE,   -- 'USR-XXXXXXXX', identitas publik
  USERNAME NVARCHAR(50) NOT NULL UNIQUE,
  EMAIL NVARCHAR(255) NOT NULL UNIQUE,
  PASSWORD NVARCHAR(255) NOT NULL,
  ROLE_CODE NVARCHAR(20) NOT NULL DEFAULT 'USER' FOREIGN KEY REFERENCES CPROLE(CODE),
  CREATED_AT DATETIME NOT NULL DEFAULT GETDATE(),
  UPDATED_AT DATETIME NOT NULL DEFAULT GETDATE()
);
CREATE TABLE CPREFRESHTOKEN (
  ID INT IDENTITY(1,1) PRIMARY KEY,
  USER_CODE NVARCHAR(20) NOT NULL FOREIGN KEY REFERENCES CPUSER(CODE) ON DELETE CASCADE,
  TOKEN NVARCHAR(512) NOT NULL UNIQUE, -- berisi SHA-256 hex, BUKAN token asli
  EXPIRES_AT DATETIME NOT NULL,
  CREATED_AT DATETIME NOT NULL DEFAULT GETDATE()
);
```

> Punya database lama (`users`/`refresh_tokens`)? Jangan buat manual — jalankan
> migrasi otomatis yang memindahkan data + membuat kode + menghapus tabel lama:
> `task migrate` (butuh `sqlcmd`) atau
> `sqlcmd -S localhost,1433 -U <user> -P <pass> -d Go -C -i scripts/migrate.sql`.
> Jadikan user pertama sebagai admin:
> `UPDATE CPUSER SET ROLE_CODE='ADMIN' WHERE USERNAME='budi';`
> (lalu login ulang agar klaim `role` di JWT terbarui).
>
> Migrasi lanjutan (`scripts/migrate2_rbac.sql`, otomatis ikut via `task migrate`):
> tabel `CPPERMISSION` (18 permission seed) + `CPROLEPERMISSION` (ADMIN=semua,
> USER=hak dasar) + `CPAUDITLOG` + `CPSYSLOG` + `CPNOTIFTEMPLATE` (3 template
> bawaan) + `CPNOTIFLOG`. Aman diulang (idempotent).

### Langkah 4 — Jalankan mode development (2 terminal)

**Terminal 1 — backend:**

```powershell
go run .
# -> API di http://localhost:1067/api
# -> health di http://localhost:1067/healthz
```

**Terminal 2 — frontend:**

```powershell
npm --prefix frontend install   # hanya pertama kali
npm --prefix frontend run dev
# -> http://localhost:5173 (otomatis proxy /api ke :1067)
```

Atau sekaligus dengan satu perintah: `task dev`.

> Catatan: log `WARN frontend/dist missing` saat `go run .` itu **normal** di mode
> dev (frontend belum di-build). API tetap jalan.

### Langkah 5 — Verifikasi instalasi

```powershell
# 1. health check
curl http://localhost:1067/healthz
# -> {"status":"ok"}

# 2. register user pertama
curl -X POST http://localhost:1067/api/users `
  -H "Content-Type: application/json" `
  -d '{"username":"budi","email":"budi@example.com","password":"password123"}'
# -> {"code":"USR-XXXXXX","message":"User created successfully"}

# 3. login
curl -X POST http://localhost:1067/api/login `
  -H "Content-Type: application/json" `
  -d '{"username":"budi","password":"password123"}'
# -> {"access_token":"...","refresh_token":"...","message":"Login successful"}

# 4. akses endpoint privat (ganti <token>)
curl http://localhost:1067/api/users -H "Authorization: Bearer <token>"
```

Lalu buka `http://localhost:5173` → halaman Login/Register/Users harus bisa dipakai
end-to-end (daftar → login → muat daftar user → logout).

### Langkah 6 — Build production (single binary)

```powershell
task build            # atau: pwsh ./scripts/build.ps1
.\app.exe             # jalan di foreground, buka http://localhost:1067/
```

Mode background (Windows) + cara berhenti:

```powershell
.\app.exe --hide      # console hilang, PID di %TEMP%\go-core.pid
.\stop.exe            # berhenti graceful
```

Selamat — aplikasi sudah jalan. Bagian 2 menjelaskan **apa isi setiap folder**,
bagian 3 menjelaskan **cara kerjanya**.

---

## 2. Memahami Isi Proyek (Tur Folder)

Struktur root (hasil `ls`):

```
go-core/
├── main.go                 # titik masuk: rakit config→db→service→route→server
├── hide_windows.go         # sembunyikan console (--hide) khusus Windows
├── hide_other.go           # versi no-op untuk Linux/macOS
│
├── config/                 # baca & validasi konfigurasi
├── models/                 # entity DB + DTO API
├── repositories/           # query SQL (mentah, parameterized)
├── services/               # logika bisnis + aturan auth
├── handlers/               # HTTP: parse request → panggil service → tulis JSON
├── middleware/             # auth JWT + rate-limit
├── routes/                 # daftarkan semua endpoint + middleware global
├── internal/pidfile/       # helper PID file (dipakai app & stop)
├── cmd/stop/               # program kecil penghenti app background
│
├── frontend/               # aplikasi React (Vite)
├── scripts/                # build.ps1, watch.ps1, dev.ps1, migrate.sql (skema CP*)
├── Taskfile.yml            # satu pintu semua perintah (task build/test/...)
│
├── Dockerfile              # image production multi-stage
├── .github/workflows/ci.yml# CI: vet+test Go, lint+build frontend
├── .env.example            # template konfigurasi (commit)
├── .env                    # konfigurasi asli (JANGAN commit, di-ignore)
├── .nvmrc                  # pin Node 20
└── app.exe + stop.exe      # hasil build (di-ignore, dibuat ulang via task build)
```

### 2.1 Lapisan backend — alur satu request

Setiap request API mengalir **dari luar ke dalam** seperti ini:

```
routes/  →  middleware/  →  handlers/  →  services/  →  repositories/  →  SQL Server
(daftar     (cek JWT,       (terima JSON,  (aturan bisnis:  (query         (tabel
 endpoint)   rate-limit)     tulis JSON)    validasi, hash)   parameterized) users/refresh_tokens)
                              ▲                  │
                           models/dto.go ────────┘
                           (bentuk data request/response)
```

Peran tiap folder:

| Folder | Isi file | Tugasnya dalam bahasa sederhana |
|--------|----------|----------------------------------|
| `config/` | `env.go`, `database.go` | Baca `.env` **sekali** saat start (`Load()` → struct `Config`), validasi (JWT ≥32 char, port numerik), buka koneksi DB dengan timeout. Tidak pernah `log.Fatal` — selalu kembalikan `error`. |
| `models/` | `user.go`, `refresh_token.go`, `role.go`, `permission.go`, `audit.go`, `syslog.go`, `notification.go`, `session.go`, `dto.go` | `User` = baris `CPUSER`; `Role`/`Permission`/`RoleDetail` (matriks RBAC); `AuditLog`+`AuditFilter`; `SysLog`; `NotifTemplate`/`NotifLog`/`NotifSendRequest`; `Session`+`SecuritySummary`. `ID` selalu disembunyikan dari JSON. |
| `repositories/` | `user/refresh_token/role/audit/syslog/notification_repository.go` | Satu-satunya tempat berisi SQL (`CPUSER`, `CPREFRESHTOKEN`, `CPROLE`, `CPPERMISSION`, `CPROLEPERMISSION`, `CPAUDITLOG`, `CPSYSLOG`, `CPNOTIF*`; relasi via `CODE`). Query parameterized, timeout 5 detik, paginated. Token = **hash SHA-256**. |
| `services/` | `user_service.go` + `rbac/session/audit/syslog/notification_service.go` | RBAC (`CheckPermission`, role CRUD, matriks permission, `UpdateUserRole`), sesi (list/revoke), audit & syslog (best-effort), notifikasi (render `{{var}}` + catat log). |
| `handlers/` | `user_handler.go`, `admin_handler.go`, `deps.go` | User: audit otomatis (register/login/update/delete/sesi) + syslog untuk error 500 + `requireSelfOrPerm` (pemilik atau pemegang permission). Admin: 20 endpoint menu (role, sesi, audit, security, syslog, notifikasi). |
| `middleware/` | `auth.go`, `ratelimit.go`, `rbac.go` | `RequirePermission(...)` → 403 JSON `forbidden: missing X` bila role tak punya permission (ADMIN selalu lolos). |
| `services/` | `user_service.go` (+ `*_test.go`) | Otak aplikasi: normalisasi email (`trim+lowercase`), validasi, bcrypt, buat kode `USR-XXXXXXXX` + role default `USER`, buat/cek JWT (`user_code` + `role`), rotasi refresh token, batasi 5 sesi/user, petakan error DB ke error bermakna (`ErrUsernameTaken`, …). Punya unit test (`go test ./services/`). |
| `handlers/` | `user_handler.go` | Penerjemah HTTP↔service: batasi body 1 MB, tolak field asing, parse `code`, cek "hanya pemilik data" (`requireSelf` bandingkan `user_code` JWT), tulis sukses/error **selalu JSON** `{...}` / `{error: ...}`. |
| `middleware/` | `auth.go`, `ratelimit.go` | `NewAuth(secret)` = satpam JWT (cek `Bearer`, pin HS256, cek `iss/aud/exp`); `RateLimiter` = pembatas request/menit per IP (anti-spoof XFF, kirim header `Retry-After`). |
| `routes/` | `routes.go` | Daftar endpoint `/api/*` + `/healthz`, pasang middleware global (`RequestID`, `Logger`, `Recoverer`, `Timeout`), dan rate-limit berbeda per endpoint (login 5/mnt, register 10/mnt, refresh 30/mnt). |
| `internal/pidfile/` | `pidfile.go` | Satu-satunya penentu lokasi PID file (`%TEMP%\go-core.pid` + fallback nama lama) agar app dan stop.exe **tidak pernah beda path**. |
| `cmd/stop/` | `main.go`, `signal_*.go` | `stop.exe`: kirim SIGINT (graceful, tunggu 8 dtk) → baru force-kill **PID itu saja**; sengaja **tidak** kill by-name agar tak salah bunuh proses lain. Windows + Unix. |
| `main.go` | — | Lem: load config → tulis PID → konek DB → rakit service → pasang route → tempel frontend embed → jalan + graceful shutdown. Tiap jam bersihkan refresh token kedaluwarsa. |

### 2.2 Frontend (`frontend/`)

```
frontend/
├── index.html              # judul "Go Core", muat /src/main.jsx
├── vite.config.js          # plugin React + Tailwind; baca APP_PORT dari root
│                           # .env → proxy /api otomatis sinkron
├── .env.example            # contoh VITE_API_URL (dev)
├── public/favicon.svg      # ikon (disajikan apa adanya)
├── src/
│   ├── main.jsx            # entry React (StrictMode) + import index.css & ui.css
│   ├── App.jsx             # shell: dibungkus ToastProvider; header, navigasi,
│   │                       # footer; state login + view
│   ├── index.css           # token CSS + `@import "tailwindcss"`
│   ├── api/client.js       # SATU-SATUNYA yang fetch ke backend:
│   │                       # simpan token di localStorage, auto-refresh 1x saat
│   │                       # 401; register/login/logout/listUsers/getUser/
│   │                       # updateUser/deleteUser/logoutAll + currentUser()
│   │                       # (baca user_id dari klaim JWT, tanpa request)
│   ├── components/         # UI KIT modern (lihat 2.5): Toast, Alert, Modal,
│   │   │                   # ConfirmDialog, Button, TextField/PasswordInput,
│   │   │                   # Spinner/Skeleton, EmptyState, Badge, Pagination,
│   │   │                   # Avatar, Card, Icon + index.js (barrel export)
│   │   ├── ui.css          # keyframes: toast slide-in, modal pop, fade,
│   │   │                   # progress bar, skeleton shimmer
│   │   └── index.js        # `import { Button, Modal } from '../components'`
│   └── pages/
│       ├── Auth.jsx        # LoginForm, RegisterForm
│       ├── Users.jsx       # User Account Management (kartu, role badge,
│       │                   # pagination, edit/logout-all/hapus)
│       ├── EditUserModal.jsx # dialog edit + dropdown role (khusus admin)
│       ├── Roles.jsx       # Role & Permission: matriks checkbox per grup,
│       │                   # buat/hapus role, simpan permission
│       ├── Sessions.jsx    # tab Sesi saya / Semua sesi + revoke per sesi
│       ├── Audit.jsx       # filter aksi/entitas/pelaku + tabel + pagination
│       ├── Security.jsx    # 6 kartu ringkasan + aktivitas + kebijakan aktif
│       ├── Syslog.jsx      # filter level + tabel + bersihkan log lama
│       └── Notifications.jsx # tab Template (CRUD + variabel) / Kirim-Test /
│                             # Riwayat kirim
└── dist/                   # HASIL build (di-ignore, jangan edit manual)
    └── .gitignore          # placeholder agar go:embed tetap compile di fresh clone
```

Alur data frontend: `pages/*.jsx` → `api/client.js` → `fetch(${VITE_API_URL}/api/...)`.
Saat dev (`npm run dev`), `VITE_API_URL` kosong → request relatif `/api/...` →
diproxy Vite ke backend. Saat production (di-embed), frontend disajikan dari
binary yang sama → request relatif otomatis benar.

### 2.5 UI Kit Frontend — komponen modern (Tailwind + custom)

Styling memakai **Tailwind CSS v4** (via plugin `@tailwindcss/vite`, nol file
konfig — cukup `@import "tailwindcss"` di `index.css`) dipadu komponen custom
di `src/components/`. Prinsip "modern" yang dipakai: varian warna semantik
(success/error/warning/info/brand), ikon SVG inline sendiri (tanpa emoji,
tanpa library ikon), animasi halus (`ui.css`), bisa di-close (tombol ×, klik
backdrop, tombol Esc), responsif, dark-mode otomatis (`dark:` mengikuti
`prefers-color-scheme`), dan aksesibel (`role="alert/dialog/status"`,
`aria-label`).

| Komponen | File | Cara pakai singkat |
|----------|------|--------------------|
| Toast (notifikasi global) | `Toast.jsx` + `useToast.js` | Bungkus app dengan `<ToastProvider>`, lalu `const t = useToast(); t.success('…')` / `t.error('…', { title })`. Auto-hilang 4 dtk (error 6 dtk), ada progres bar + tombol ×. |
| Alert (banner inline) | `Alert.jsx` | `<Alert tone="error" title="…" closable>pesan</Alert>` — untuk error form. |
| Modal (dialog) | `Modal.jsx` | `<Modal open title onClose footer size="sm\|md\|lg">` — portal ke body, kunci scroll, tutup via Esc/backdrop/×. |
| ConfirmDialog | `ConfirmDialog.jsx` | `<ConfirmDialog open danger loading onConfirm onCancel message>` — untuk hapus & logout-all. |
| Button | `Button.jsx` | `<Button variant="primary\|secondary\|danger\|ghost" size="sm\|md\|lg" loading fullWidth>` |
| TextField / PasswordInput | `TextField.jsx` | `<TextField label error hint>`; password punya toggle intip (ikon mata). |
| Spinner / PageLoader / Skeleton | `Spinner.jsx` | `<SkeletonRows rows={4}/>` untuk loading daftar. |
| EmptyState | `EmptyState.jsx` | `<EmptyState title description action={<Button…/>}>` saat data kosong. |
| Badge | `Badge.jsx` | `<Badge tone="brand">Anda</Badge>` |
| Pagination | `Pagination.jsx` | `<Pagination offset limit count hasMore onPage>` — offset-based, cocok dengan `limit/offset` backend. |
| Avatar | `Avatar.jsx` | `<Avatar name size="sm\|md\|lg"/>` — inisial + gradien stabil per nama. |
| Card / CardTitle | `Card.jsx` | Pembungkus section konten yang konsisten. |
| Icon | `icons.jsx` | `<Icon name="check\|x\|info\|warning\|eye\|pencil\|trash\|logout\|…"/>` |

Contoh menambah komponen baru yang konsisten: buat `Baru.jsx` + pakai kelas
Tailwind + varian `dark:` + ikon dari `icons.jsx`, lalu daftarkan di
`components/index.js` agar bisa diimpor lewat barrel.

### 2.3 Build & skrip (`Taskfile.yml`, `scripts/`)

```
Taskfile.yml  →  task build | test | watch | dev | clean | run | tray | stop | ...
     │                │
     │                └── memanggil ./scripts/*.ps1 (logika asli hanya di sini)
     │
     └── butuh CLI `task`? Kalau belum install, langsung: pwsh ./scripts/build.ps1
```

| File | Perintah | Isi kerjanya |
|------|----------|--------------|
| `scripts/build.ps1` | `task build` | `npm ci` → `vite build` → buat ulang `dist/.gitignore` → `go vet` → `go build -trimpath -ldflags "-s -w"` → `app.exe` + `stop.exe` |
| `scripts/watch.ps1` | `task watch` | Pantau `*.go/js/jsx/css/html` → rebuild + restart via **PID file** (bukan kill by-name) |
| `scripts/dev.ps1` | `task dev` | Buka 2 jendela: Vite HMR + `go run .` |
| `Taskfile.yml` | semua `task *` | Definisi task lintas-fungsi: `install`, `build-frontend`, `build-backend`, `test` (`go vet` + `go test -race`), `lint-frontend`, `clean` (aman Windows) |

### 2.4 File konfigurasi & operasional

| File | Untuk apa | Boleh diedit? |
|------|-----------|---------------|
| `.env` | Konfigurasi asli (DB, JWT, port). Di-ignore git | ✅ wajib (tidak ikut commit) |
| `.env.example` | Template `.env` + dokumentasi default | ✅ jika tambah variabel baru |
| `frontend/.env.example` | Contoh `VITE_API_URL` untuk dev terpisah | ✅ |
| `.nvmrc` | Pin Node 20 (`nvm use`) | ❌ kecuali upgrade Node |
| `Dockerfile` | Build image: node→build frontend, go→binary, distroless→jalan | ✅ jika ubah port/proses build |
| `.github/workflows/ci.yml` | CI 3 job: backend, frontend, single-binary Windows | ✅ |
| `go.mod` / `go.sum` | Dependensi Go (chi, jwt, mssqldb, crypto, godotenv) | via `go get`, jangan manual |
| `frontend/package.json` | Dependensi React + Tailwind v4 + script `dev/build/lint/preview` | via `npm install <pkg>` (devDeps: `@tailwindcss/vite`, `tailwindcss`) |

---

## 3. Cara Kerja Sistem (Alur Penting)

### 3.1 Alur login → akses → refresh (auth)

```
REGISTER
  UI ──POST /api/users {username,email,password}──► bcrypt hash ──► INSERT users

LOGIN
  UI ──POST /api/login {username,password}──► cek bcrypt ──► access JWT (15 mnt)
                                                              + refresh acak (256-bit)
                                                              + simpan SHA256(refresh) di DB
  UI simpan keduanya di localStorage.

AKSES PRIVAT
  UI ──GET /api/users + Header "Authorization: Bearer <access>"──► middleware cek JWT
      ──► 200 JSON / 401 {"error":...}

TOKEN KEDALUWARSA (otomatis di client.js)
  401 ──► POST /api/refresh {refresh_token} ──► hapus token lama + terbitkan pasangan
  baru ──► ulangi request awal. Gagal refresh → logout (token dibuang).
```

Aturan keamanan: 1 user maksimal **5 sesi** (login ke-6 menghapus sesi tertua);
`PUT/DELETE /users/{code}` dan `logout-all` hanya oleh **pemilik code** (403 jika bukan);
logout token yang tidak dikenal → `401 invalid or expired refresh token`.

### 3.2 Contoh request/response

```powershell
# Header auth untuk endpoint privat:
$h = @{ Authorization = "Bearer <access_token>" }

# List paginated (default limit 50, maks 200):
curl "http://localhost:1067/api/users?limit=10&offset=0" -H "Authorization: Bearer <token>"

# Update partial — kirim HANYA field yang berubah (tanpa password = password tetap).
# Ganti USR-000002 dengan code milik Anda (lihat dari respons register / daftar users):
curl -X PUT http://localhost:1067/api/users/USR-000002 -H "Authorization: Bearer <token>" `
  -H "Content-Type: application/json" -d '{"username":"budi2"}'

# Cabut semua sesi user tersebut:
curl -X POST http://localhost:1067/api/users/USR-000002/logout-all -H "Authorization: Bearer <token>"

# Lihat master role:
curl http://localhost:1067/api/roles -H "Authorization: Bearer <token>"
```

### 3.3 Cara kerja single binary (embed)

1. `task build` → `vite build` menghasilkan `frontend/dist/` (index.html + assets).
2. `go build` membaca `main.go: //go:embed all:frontend/dist` → frontend **masuk ke dalam** `app.exe` (~10 MB).
3. Saat jalan, request `/` + file statis dilayani dari embed (index.html di-cache di memori);
   request `/api/*` + `/healthz` yang tak dikenal → JSON `{"error":"Not found"}` (bukan HTML),
   sehingga refresh halaman SPA tidak 404.
4. Fresh clone tanpa `dist/`: embed tetap compile berkat placeholder
   `frontend/dist/.gitignore` → app jalan **mode API-only** + log peringatan.

---

## 4. Referensi API

Base URL dev: `http://localhost:5173` (frontend) / `http://localhost:1067` (langsung).
Base URL prod: `http://localhost:1067/` (keduanya satu origin).

| Method | Path | Auth | Rate-limit | Body | Sukses |
|--------|------|------|------------|------|--------|
| GET | `/healthz` | — | — | — | `{"status":"ok"}` |
| POST | `/api/users` | — | 10/mnt | `{username, email, password}` (role otomatis `USER`) | `201 {"code","message"}` (`code` = `USR-XXXXXXXX`) |
| POST | `/api/login` | — | 5/mnt | `{username, password}` | `200 {access_token, refresh_token}` (JWT berisi `user_code` + `role`) |
| POST | `/api/refresh` | — | 30/mnt | `{refresh_token}` | `200 {access_token, refresh_token}` (lama hangus) |
| POST | `/api/logout` | — | 30/mnt | `{refresh_token}` | `200 {message}` |
| GET | `/api/roles` | Bearer | — | — | `200 [{code,name,...}]` (master `CPROLE`) |
| GET | `/api/users?limit=&offset=` | Bearer | — | — | `200 [...]` (array item `{code,username,email,role_code,...}`, `[]` jika kosong) |
| GET | `/api/users/{code}` | Bearer | — | — | `200 {code,username,email,role_code,...}` |
| PUT | `/api/users/{code}` | Bearer + owner | — | partial `{username?, email?, password?}` | `200 {message}` |
| DELETE | `/api/users/{code}` | Bearer + owner | — | — | `200 {message}` (+ sesi dibersihkan via CASCADE) |
| POST | `/api/users/{code}/logout-all` | Bearer + owner | — | — | `200 {message}` |
| PUT | `/api/users/{code}/role` | Bearer + `USER_ROLE_ASSIGN` | — | `{role_code}` | `200 {message}` |
| POST | `/api/admin/roles` | Bearer + `ROLE_MANAGE` | — | `{code, name}` | `201 role` |
| GET | `/api/admin/roles/{code}` | Bearer + `ROLE_READ` | — | — | `200 {role, permissions[]}` (matriks) |
| DELETE | `/api/admin/roles/{code}` | Bearer + `ROLE_MANAGE` | — | — | `200` (gagal bila role dipakai user) |
| GET | `/api/admin/permissions` | Bearer + `ROLE_READ` | — | — | `200 [...]` (18 permission per grup) |
| PUT | `/api/admin/roles/{code}/permissions` | Bearer + `PERMISSION_ASSIGN` | — | `{permissions:[...]}` | `200` (replace atomik) |
| GET | `/api/admin/sessions` | Bearer | — | — | sesi login milik sendiri |
| GET | `/api/admin/sessions/all?limit=&offset=` | Bearer + `SESSION_MANAGE` | — | — | semua sesi aktif |
| DELETE | `/api/admin/sessions/{id}` | Bearer (pemilik/`SESSION_MANAGE`) | — | — | `200` |
| GET | `/api/admin/audit?action=&entity=&actor=` | Bearer + `AUDIT_READ` | — | — | jejak aksi + IP |
| GET | `/api/admin/security/summary` | Bearer + `SECURITY_READ` | — | — | 6 angka ringkasan |
| GET | `/api/admin/syslogs?level=` | Bearer + `SYSLOG_READ` | — | — | `ERROR/WARN/INFO` |
| DELETE | `/api/admin/syslogs?days=` | Bearer + `SYSLOG_MANAGE` | — | — | `{deleted}` |
| GET | `/api/admin/notifications/templates` | Bearer + `NOTIF_READ` | — | — | template + `{{var}}` |
| POST/PUT/DELETE | `/api/admin/notifications/templates…` | Bearer + `NOTIF_MANAGE` | — | `{name,channel,subject,body,is_active}` | CRUD template |
| POST | `/api/admin/notifications/send` | Bearer + `NOTIF_SEND` | — | `{template_code,recipient,variables}` | `201` + tercatat di log |
| GET | `/api/admin/notifications/logs` | Bearer + `NOTIF_READ` | — | — | riwayat kirim |

Aturan validasi: username ≥3 (maks 50, tanpa karakter kontrol), email valid
(maks 254, disimpan lowercase), password 8–72 byte. Semua error: JSON
`{"error": "..."}` dengan status `400/401/403/404/409/429/500` yang sesuai.

---

## 5. Referensi Perintah & Konfigurasi

### 5.1 Semua perintah (`task` = `Taskfile.yml`)

| Perintah | Artinya | Kapan dipakai |
|----------|---------|---------------|
| `task dev` | Buka Vite HMR + `go run .` | Kerja harian |
| `task build` | Build penuh → `app.exe` + `stop.exe` | Rilis / test prod lokal |
| `task build-frontend` / `task build-backend` | Build salah satu sisi | Hemat waktu |
| `task run` | Build + jalan foreground | Coba prod cepat |
| `task tray` | Build + jalan background (`--hide`) | Pakai harian di Windows |
| `task stop` | Hentikan app background | — |
| `task watch` | Auto-rebuild tiap ada file berubah | Demo / iterasi prod-like |
| `task migrate` | Terapkan `scripts/migrate.sql` + `migrate2_rbac.sql` berurutan | Setelah pull / untuk DB lama (CP* + RBAC + audit + notif seed) |
| `task test` | `go vet` + `go test -race ./...` | Sebelum commit |
| `task lint-frontend` | `oxlint` | Sebelum commit |
| `task install` | `npm ci` di frontend | Sinkron dep frontend |
| `task clean` | Hapus `app.exe`, `stop.exe`, output `dist` | Mulai bersih |

> Belum install CLI `task`? Ganti `task build` → `pwsh ./scripts/build.ps1`
> (dan seterusnya). Go + Node tetap wajib.

### 5.1 Variabel environment

| Variabel | Wajib | Default | Keterangan |
|----------|-------|---------|------------|
| `APP_NAME` / `APP_ENV` | tidak | `GoBackend` / `development` | Label saja |
| `APP_PORT` | tidak | `1067` | Port HTTP; Vite proxy ikut otomatis |
| `DB_HOST` / `DB_PORT` / `DB_DATABASE` | ya | — | Contoh: `localhost` / `1433` / `Go` |
| `DB_USERNAME` / `DB_PASSWORD` | ya | — | Kredensial SQL Server |
| `DB_MAX_OPEN_CONNS` / `DB_MAX_IDLE_CONNS` | tidak | `25` / `10` | Tuning pool koneksi |
| `JWT_SECRET` | ya | — | **≥32 karakter acak** (`openssl rand -hex 32`) |
| `VITE_API_URL` (frontend) | tidak | kosong (relatif) | Isi `http://localhost:1067` hanya jika frontend & backend beda origin |

### 5.2 Testing

```powershell
task test                       # vet + seluruh test Go (race detector)
go test ./services/ -run TestLogin -v   # satu grup test
go test -race ./...             # eksplisit
npm --prefix frontend run lint  # lint React
```

Test ada di `services/user_service_test.go` (table-driven: validasi, duplikat,
cap 5 sesi, rotasi, expired, logout) dengan mock thread-safe
(`user_service_mock_test.go`, tanpa DB).

---

## 6. Deploy Production

**Opsi A — Windows (file copy):**

```powershell
task build
# copy ke server: app.exe + stop.exe + .env
.\app.exe --hide
.\stop.exe   # untuk berhenti
```

**Opsi B — Docker (Linux):**

```powershell
docker build -t go-core .
docker run -p 1067:1067 --env-file .env go-core
# Catatan: butuh SQL Server yang reachable dari container (bukan localhost container)
```

Checklist sebelum live: `JWT_SECRET` acak ≥32 char & beda dari dev,
`DB_PASSWORD` kuat, `.env` tidak ikut repo (`git status` bersih),
`task test` hijau, `task build` dari clone bersih berhasil.

---

## 7. Troubleshooting & FAQ

| Gejala | Penyebab umum → solusi |
|--------|------------------------|
| `missing required env: ...` saat start | `.env` belum dibuat/diisi → `copy .env.example .env`, isi 6 variabel wajib |
| `ping database: ...` | SQL Server mati / kredensial salah / firewall → cek service SQL, login via SSMS, port 1433 |
| `WARN frontend/dist missing` (dev) | Normal — build frontend hanya untuk prod → `task build-frontend` jika ingin hilangkan |
| Kelas Tailwind tidak berefek di browser | Pastikan `npm run dev` jalan dari folder `frontend` dan file sudah disimpan (Tailwind v4 mendeteksi class otomatis, tanpa restart). Coba hard refresh `Ctrl+Shift+R` |
| Toast/modal tidak muncul | Pastikan halaman dibungkus `<ToastProvider>` (sudah di `App.jsx`) dan panggil `useToast()` di dalam provider |
| `localhost:1067` 404 di browser | Buka `http://localhost:1067/` (bukan `/api`); hard refresh `Ctrl+Shift+R` |
| `Missing authorization header` / 401 | Endpoint privat butuh `Authorization: Bearer <access_token>` → login dulu; jika expired, client auto-refresh |
| `Too many requests` (429) | Kena rate-limit → tunggu sesuai header `Retry-After`, jangan spam retry |
| Port bentrok | Ganti `APP_PORT` di `.env` → restart backend; Vite ikut otomatis |
| `stop.exe` → "PID file not found" | App tidak jalan via PID (mungkin crash) → cek Task Manager; hapus `%TEMP%\go-core.pid` jika stale |
| `go:embed ... no matching files` | `frontend/dist/` kosong total → `task build-frontend` (atau `dist/.gitignore` hilang → kembalikan) |
| Cross-compile Linux gagal (cgo/gcc) | Tambahkan `CGO_ENABLED=0`: `CGO_ENABLED=0 GOOS=linux go build ./...` |
| Token lama invalid setelah update server | Wajar — JWT kini berisi `user_code` (bukan `user_id`) + tabel jadi `CP*` → semua user wajib login ulang |

**FAQ singkat:**
- *Data user dari mana?* SQL Server, tabel `users` + `refresh_tokens` (lihat Langkah 3).
- *Ganti React dengan framework lain?* Bisa — yang di-embed hanya isi `frontend/dist/`; backend tidak peduli isinya.
- *Tambah endpoint baru?* Urutan: SQL di `repositories/` → aturan di `services/` (+test) → JSON di `handlers/` → daftar di `routes/` → pakai dari `api/client.js`.
- *Frontend & backend beda server?* Isi `VITE_API_URL` di frontend + rebuild; backend tetap sama (tambah CORS bila perlu).
