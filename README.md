# Golang Backend + React — Single Binary (embed.FS)

Backend Go (chi) + Frontend React (Vite) jadi **satu binary** via `//go:embed all:frontend/dist`. Deploy cukup `app.exe` + `.env`.

## Stack

- **Backend:** Go 1.27, chi v5, golang-jwt v5, MSSQL (go-mssqldb), bcrypt
- **Frontend:** React 19, Vite 8, di `frontend/`
- **Auth:** Access Token 15m (JWT HS256) + Refresh Token rotate 7 hari, limit 5/device per user
- **Port:** `1067` (via `APP_PORT` di `.env`, proxy Vite sudah sinkron)

## Struktur

```
golang-backend/
├── config/          # env + database
├── handlers/        # http handlers
├── middleware/      # auth, rate-limit (5/min login)
├── models/
├── repositories/    # RowsAffected check, refresh limit
├── routes/          # /api/*  (SPA NotFound fallback)
├── services/        # business logic
├── utils/           # jwt (iat/exp + HMAC check), hash, validator
├── frontend/        # React app (Vite)
│   ├── src/
│   ├── dist/        # hasil build (di-ignore, kecuali .gitkeep)
│   └── vite.config.js # base:/, outDir:dist, proxy /api -> :1067
├── main.go          # //go:embed all:frontend/dist + SPA fallback
├── .env             # tidak di-commit (lihat .env.example)
├── .env.example     # template
├── Makefile         # make build / dev / clean
├── build.ps1        # Windows one-command build
└── build.sh         # Linux/macOS one-command build
```

## Prasyarat

- Go >=1.27
- Node >=18, npm >=9
- SQL Server (MSSQL) running, database `Go` bisa diakses
- Git

## Fresh Clone — Langkah Lengkap

### 1. Clone

```powershell
git clone <url-repo-kamu>.git
cd golang-backend
```

### 2. Env

```powershell
copy .env.example .env
# edit .env isi DB_USERNAME, DB_PASSWORD, JWT_SECRET
```

Isi `.env`:

```ini
APP_PORT=1067
DB_HOST=localhost
DB_PORT=1433
DB_DATABASE=Go
DB_USERNAME=RAYSERVER
DB_PASSWORD=rayserver
JWT_SECRET=isi-random-min-32-char
```

Generate JWT_SECRET (pilih satu):

```powershell
# PowerShell
-join ((48..57)+(65..90)+(97..122) | Get-Random -Count 64 | % {[char]$_})
# atau
openssl rand -hex 32
```

> `.env` sudah di `.gitignore`, tidak akan ter-push. Jangan commit `.env`.

### 3. Database

Buat DB + tabel (SSMS / sqlcmd):

```sql
CREATE DATABASE Go;
GO
USE Go;
GO
CREATE TABLE users (
  id INT IDENTITY(1,1) PRIMARY KEY,
  username NVARCHAR(50) NOT NULL,
  email NVARCHAR(255) NOT NULL,
  password NVARCHAR(255) NOT NULL,
  created_at DATETIME NOT NULL DEFAULT GETDATE(),
  updated_at DATETIME NOT NULL DEFAULT GETDATE(),
  CONSTRAINT UQ_users_username UNIQUE (username),
  CONSTRAINT UQ_users_email UNIQUE (email)
);
CREATE TABLE refresh_tokens (
  id INT IDENTITY(1,1) PRIMARY KEY,
  user_id INT NOT NULL FOREIGN KEY REFERENCES users(id) ON DELETE CASCADE,
  token NVARCHAR(512) NOT NULL UNIQUE,
  expires_at DATETIME NOT NULL,
  created_at DATETIME NOT NULL DEFAULT GETDATE()
);
```

### 4. Development (2 terminal, HMR)

Terminal 1 — Backend:

```powershell
go mod tidy   # pertama kali saja (opsional)
go run .
# -> http://localhost:1067
# -> API http://localhost:1067/api
```

Terminal 2 — Frontend:

```powershell
cd frontend
npm install   # pertama kali saja
npm run dev
# -> http://localhost:5173 (proxy /api -> :1067)
```

Buka `http://localhost:5173` untuk dev. Edit `frontend/src/App.jsx` auto reload.

> Jika `go run .` log `WARN: frontend/dist not found` itu normal di mode dev (belum build). `/api` tetap jalan. Build frontend hanya untuk production single binary.

### 5. Production — Single Binary

```powershell
# Windows
.\build.ps1
# atau
make build

# Linux/macOS
bash build.sh
# atau
make build
```

Output: `app.exe` (~12-13 MB) sudah embed `frontend/dist`.

Jalankan:

```powershell
.\app.exe
# buka http://localhost:1067/  (frontend)
# buka http://localhost:1067/api/users  (API)
# refresh /dashboard tidak 404 (SPA fallback)
```

Deploy prod cukup copy **2 file**: `app.exe` + `.env` ke server. Folder `frontend/` tidak perlu ikut.

## API

| Method | Path | Auth | Body | Ket |
|--------|------|------|------|-----|
| POST | `/api/users` | - | `{username,email,password}` | Register (201) |
| POST | `/api/login` | - | `{username,password}` | Login -> `access_token` + `refresh_token` |
| POST | `/api/refresh` | - | `{refresh_token}` | Rotate -> token baru, old invalidate |
| POST | `/api/logout` | - | `{refresh_token}` | Hapus refresh token |
| GET | `/api/users` | Bearer | - | List (butuh token) |
| GET | `/api/users/{id}` | Bearer | - | Detail |
| PUT | `/api/users/{id}` | Bearer (owner only) | `{username,email,password}` | 403 jika bukan owner |
| DELETE | `/api/users/{id}` | Bearer (owner only) | - | 403 jika bukan owner |

Contoh:

```powershell
# register
curl -X POST http://localhost:1067/api/users -H "Content-Type: application/json" -d '{"username":"budi","email":"budi@example.com","password":"password123"}'

# login
curl -X POST http://localhost:1067/api/login -H "Content-Type: application/json" -d '{"username":"budi","password":"password123"}'

# pakai token
curl http://localhost:1067/api/users -H "Authorization: Bearer <access_token>"
```

## Script Lain

```powershell
make install         # npm install di frontend
make build-frontend  # npm run build saja
make build-backend   # go build saja (butuh dist sudah ada)
make clean           # hapus app.exe + dist/assets
make dev             # petunjuk 2 terminal
```

## Port

Ubah di `.env` (`APP_PORT`) + `frontend/vite.config.js` (`proxy /api`) harus sama. Default `1067` aman (non-privileged, tidak butuh Administrator, tidak bentrok DHCP port 67). Hindari `8080` jika bentrok, jangan pakai `<1024`.

## Troubleshooting

| Masalah | Solusi |
|---------|--------|
| `go:embed pattern all:frontend/dist: no matching files` | `frontend/dist` belum ada. Buat placeholder sudah ada `.gitkeep`, atau `npm --prefix frontend run build`. |
| `WARN: frontend/dist not found` saat `go run .` | Normal di dev. Jalankan `npm --prefix frontend run build` atau abaikan, `/api` tetap jalan. |
| `localhost:1067` Not Found di browser | Pastikan buka `http://localhost:1067/` (bukan `/api`). Hard refresh `Ctrl+Shift+R`. Cek `netstat -ano | findstr 1067` harus LISTENING. |
| `Missing authorization header` | Endpoint `/api/users` butuh `Authorization: Bearer <token>`. Login dulu. |
| DB `Failed to connect` | Cek SQL Server jalan, `DB_*` di `.env` benar, firewall port 1433. |
| `JWT_SECRET not configured` | Isi `JWT_SECRET` di `.env` (min 32 char). |

## Keamanan

- `.env` tidak di-commit (sudah `.gitignore`). Jika history lama pernah ke-commit secret, **rotate** `JWT_SECRET` + `DB_PASSWORD`.
- Commit template pakai `.env.example` (tanpa secret).
- Rate limit login 5/menit, refresh rotate + limit 5 per user, ownership check 403.

## Lisensi

Internal — sesuaikan kebutuhan.
