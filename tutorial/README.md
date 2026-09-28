# Tutorial: Dari Clone sampai Aplikasi Sendiri

Tutorial ini berjalan dari nol, dalam dua fase:

| Fase | Isi | Tujuan |
|---|---|---|
| **A — Pengerasan** | [Bagian 0](#bagian-0--dari-clone-sampai-aplikasi-jalan) | Clone repo, aplikasi jalan, ganti nama & favicon, pilih database |
| **B — Penambahan fitur** | [Bagian 1](#bagian-1--daftarkan-modul-dan-menu-di-database) s.d. [Bagian 7](#bagian-7--menghapus-menu) | Tambah modul, menu, role, user, lalu beri akses |

Fase B adalah inti tutorial ini: buat **modul**, buat **menu** di dalamnya,
isi halamannya, lalu beri **role** akses ke menu itu. Admin tidak perlu membuat
folder atau kode khusus untuk setiap role.

Setelah menu dibuat, admin hanya perlu memberi centang akses pada role yang
diizinkan. Setelah user login ulang, menu otomatis muncul di sidebar.

Kalau repo sudah jalan dan Anda hanya ingin menambah fitur, langsung ke
[Bagian 1](#bagian-1--daftarkan-modul-dan-menu-di-database).

## Resep 5 Menit (ringkasan)

Contoh: modul `TOKO`, menu `Stok`, role `EDITOR`, user `editor`.

| # | Apa yang dilakukan | Di mana | Hasil |
|---|---|---|---|
| 1 | Buat modul `TOKO` | UI **Modul & Menu** → **Modul baru** | Baris `TOKO` di `CPMODULE` |
| 2 | Buat menu `MENU_STOK` (modul `TOKO`, jenis `CHILD`) | UI **Modul & Menu** → **Menu baru** | Baris `MENU_STOK` di `CPMENU`, belum ada grant |
| 3 | Siapkan halaman | `cp -r tutorial/templates/frontend-menu frontend/src/app/stok` | Folder menu 5 file (`index.jsx`, `controller.js`, `GRID.jsx`, `FRM.jsx`, `api.js`) |
| 4 | Isi halaman | `GRID.jsx` (`COLUMNS_ITEMS`) + `FRM.jsx` (`validate_field`) + `api.js` (`read_data`/`process_*`) | Tabel stok (list, tambah, ubah, hapus) |
| 5 | (opsional) Endpoint backend | `features/stok/` + `routes/routes.go` | `GET/POST/PUT/DELETE /api/stok` |
| 6 | Build | `npm run build` + restart backend | Sidebar punya grup `TOKO` |
| 7 | Buat role `EDITOR` | UI **Role & Permission** → **Role baru** | Baris `EDITOR` di `CPROLE` |
| 8 | Beri akses | UI **Role & Permission** → centang `MENU_STOK` → **Simpan** | Baris di `CPPERMISSION` |
| 9 | Buat user | UI **User Account** → **Tambah user** (role `EDITOR`) | User bisa login, menu Stok terlihat |

Langkah 1–4 wajib. Langkah 5 boleh dilewati bila memakai endpoint yang sudah
ada. Detail tiap langkah ada di bawah.

## Aturan Utama

Struktur folder menu DATAR (tanpa folder modul):

```text
frontend/src/app/<mcontrol>/  -> satu menu CHILD (MCONTROL snake_case)
```

Contoh bawaan:

```text
app/users/          -> MODULE SYSTEM, MENU_USERS (MCONTROL users)
app/laporan/        -> MODULE REPORT, MENU_LAPORAN (MCONTROL laporan, CHILD tanpa parent)
app/arus_kas/       -> MODULE REPORT, MENU_ARUS_KAS (MCONTROL arus_kas, CHILD dari MENU_KEUANGAN)
(menu PARENT MENU_KEUANGAN tanpa folder — hanya header buka-tutup)
```

`MODULE` = kolom `CPMENU`.MODULE (section sidebar, FK ke `CPMODULE.CODE`).
`MCONTROL` = kolom `CPMENU`.MCONTROL (nama folder, snake_case).
Judul tampil = `LABEL`.

Menu parent/child ditentukan **database**, bukan folder:

| Kolom `CPMENU` | Isi |
|---|---|
| `MODULE` | Section sidebar (FK ke `CPMODULE.CODE`); label section = `CPMODULE.LABEL` |
| `MCONTROL` | Nama folder `app/<mcontrol>/` (snake_case); CHILD wajib isi, PARENT wajib NULL |
| `MENU_KIND` | `PARENT` (header buka-tutup, tanpa folder) atau `CHILD` (item biasa) |
| `PARENT_CODE` | Kode menu parent; harus menunjuk menu `PARENT` satu modul yang sama |

Tutorial ini memakai modul baru `TOKO` dan menu `stok`.

Satu permission mewakili satu menu:

```text
MENU_<NAMA_MENU>
```

Contoh MCONTROL `stok` -> folder `app/stok/`, permission:

```text
MENU_STOK
```

Satu checkbox `MENU_STOK` memberi akses ke **seluruh fungsi** menu
Stok. Tidak ada permission terpisah untuk list, create, edit, atau delete.

Tiga aturan yang paling sering membuat orang tersesat:

1. MCONTROL snake_case = nama folder datar `app/<mcontrol>/`.
   `stok` → folder `app/stok/` (permission bebas, mis. `MENU_STOK`).
2. MODULE = section sidebar (FK ke CPMODULE, mis. `TOKO`); PARENT tanpa
   folder tampil sebagai header buka-tutup bila keturunannya ter-grant.
3. Menu baru **tidak** otomatis bisa diakses siapa pun. Sampai admin mencentang
   di halaman Role, menu itu tidak muncul di sidebar siapa pun.

## Istilah Penting

| Istilah | Arti |
|---|---|
| Modul | Kelompok menu di sidebar (section level-1), mis. `SYSTEM`, `REPORT`. Disimpan di `CPMODULE`. |
| Menu | Satu halaman di sidebar, mis. User Account atau Stok. Disimpan di `CPMENU` (CHILD, punya MCONTROL/folder). |
| MCONTROL | Nama folder menu (`app/<mcontrol>/`, snake_case). CHILD wajib isi, PARENT wajib NULL. |
| Menu parent | Header buka-tutup bertipe `PARENT`: tanpa folder/halaman, tanpa centang di matriks, dan tampil otomatis bila minimal satu keturunannya ter-grant. |
| Menu child | Menu bertipe `CHILD`: item biasa dengan folder, boleh punya `PARENT_CODE`. Hanya menu inilah yang dicentang di matriks. |
| Role | Kelompok user, misalnya ADMIN, USER, EDITOR. Disimpan di `CPROLE`. |
| Permission | Akses ke satu menu, misalnya `MENU_STOK`. |
| Grant | Baris `CPPERMISSION` yang memberi role akses ke menu. |
| Checkbox | Centang di halaman Role & Permission. |
| Sidebar | Daftar menu 3 level (MODULE section -> PARENT header -> CHILD item) yang otomatis dibuat dari database + folder menu. |

Semua perintah `cp`, `ls`, dan `grep` dijalankan dari folder root
repository `go-core` menggunakan Git Bash.

---

# Bagian 0 — Dari Clone sampai Aplikasi Jalan

Bagian ini khusus **clone baru**: dari mengunduh repo sampai aplikasinya jalan
dengan nama, logo, dan database milik Anda. Tidak ada kode yang perlu diedit
manual. Setelah bagian ini selesai, lanjut ke Bagian 1 (tambah modul & menu)
sampai Bagian 7.

| # | Yang dikerjakan | Selesai bila |
|---|---|---|
| 0.1 | Clone repository | folder `go-core/` ada di komputer |
| 0.2 | Pasang dependency frontend | `frontend/node_modules/` terisi |
| 0.3 | Membuat `.env` | nama database + `JWT_SECRET` terisi |
| 0.4 | Buat database, migrasi, akun awal | query tabel `CP%` mengembalikan 10 |
| 0.5 | Jalankan aplikasi | halaman login terbuka di browser |
| 0.6 | Verifikasi hasil clone | sidebar 2 modul, matriks 11 menu |
| 0.7 | Nama database bebas | nama DB di `.env` = nama DB di SQL Server |
| 0.8 | Mengganti nama aplikasi | judul tab & kartu login memakai nama Anda |
| 0.9 | Mengganti favicon & logo | logo Anda muncul di sidebar + kartu login |
| 0.10 | Engine lain (opsional) | hanya bila memilih PostgreSQL/SQLite |
| 0.11 | Kalau mentok saat setup | pesan error ketemu solusinya |

## 0.1. Clone repository

Prasyarat: Go 1.27+, Node.js 20+, npm 9+, SQL Server 2019+ (dengan `sqlcmd`),
dan Task CLI (opsional — semua perintah ada padanannya di `scripts/*.sh`).

```bash
git clone <url-repo>.git
cd go-core
```

Cek prasyarat, semuanya harus keluar versinya:

```bash
go version
node --version
npm --version
sqlcmd -?
```

## 0.2. Pasang dependency frontend

```bash
task install
```

Tanpa Task CLI:

```bash
cd frontend
npm ci
cd ..
```

Selesai bila folder `frontend/node_modules/` sudah ada.

## 0.3. Membuat file `.env`

```bash
copy .env.example .env   # Windows CMD
# cp .env.example .env   # Git Bash / Linux / macOS
notepad .env
```

Isi minimal (sesuaikan nama database, user, dan password SQL Server Anda):

```env
APP_NAME=Toko Saya
APP_LOGO=/favicon.svg
APP_PORT=1067
DB_CONNECTION=sqlserver
DB_HOST=localhost
DB_PORT=1433
DB_DATABASE=TokoDb
DB_USERNAME=<user-sqlserver-anda>
DB_PASSWORD=<password-anda>
JWT_SECRET=<string-acak-minimal-32-karakter>
ACCESS_TOKEN_MINUTES=15
REFRESH_TOKEN_DAYS=7
```

Buat `JWT_SECRET` acak dengan `openssl rand -hex 32`. Jangan pernah commit
`.env` (sudah masuk `.gitignore`).

## 0.4. Buat database, migrasi, dan akun awal

1. Buat database (nama bebas, contoh `TokoDb`):

```sql
IF DB_ID(N'TokoDb') IS NULL
BEGIN
  CREATE DATABASE [TokoDb];
END
GO
```

2. Dari folder root, jalankan tiga skrip migrasi (`task migrate` membaca `.env`
   sehingga nama database takenya dari `DB_DATABASE`):

```bash
task migrate
```

3. Buat dua akun awal:

```bash
sqlcmd -S localhost,1433 -U <user> -P "<password>" -d TokoDb -C -i scripts/seed-admin.sql
```

Akun yang terbentuk: `admin`/`admin` (role `ADMIN`) dan `user`/`user`
(role `USER`, nol menu).

Selesai bila:

```sql
SELECT COUNT(*) AS TABEL
FROM INFORMATION_SCHEMA.TABLES
WHERE TABLE_TYPE = 'BASE TABLE' AND TABLE_NAME LIKE 'CP%';
-- TABEL = 10
```

> Jangan lewatkan `seed-admin.sql`: tanpa itu tidak ada satu pun akun yang
> bisa login.

## 0.5. Jalankan aplikasi

Development (backend + Vite HMR dalam satu terminal, `Ctrl+C` menghentikan
keduanya):

```bash
task start
```

Buka `http://localhost:5173`, lalu login `admin` / `admin`.

Production (satu binary; frontend di-embed ke dalam `app.exe`):

```bash
task build
./app.exe
```

Buka `http://localhost:1067`. Cek backend:

```bash
curl http://localhost:1067/healthz
# {"status":"ok"}
```

Mode tersembunyi di Windows: `./app.exe --hide` (berhenti dengan `./stop.exe`).

> `go:embed no matching files`? Folder `frontend/dist` wajib ada karena
> di-embed ke binary. Jalankan sekali `task build-frontend` (atau `task build`)
> lalu ulangi.

## 0.6. Verifikasi hasil clone

Login `admin`/`admin`, lalu pastikan semua poin ini benar:

- [ ] Sidebar punya 2 section modul: `SYSTEM` (8 menu) dan `REPORT`
      (2 menu contoh `Laporan` + `Arus Kas` di bawah header `Keuangan`).
- [ ] Menu **Modul & Menu** menampilkan 2 modul + 11 baris menu.
- [ ] **Role & Permission** menampilkan matriks 11 menu; `MENU_KEUANGAN`
      tampil sebagai header **tanpa centang**, `MENU_ARUS_KAS` di bawahnya
      punya checkbox.
- [ ] Role `ADMIN` tercentang 8 menu `SYSTEM`; role `USER` nol centang.
- [ ] Menu `Laporan` dan `Arus Kas` belum tampil di sidebar (belum di-grant).
- [ ] Logout, lalu login `user`/`user` → halaman kosong "hubungi admin".

Acuan query lengkap ada di
[Lampiran A](#lampiran-a--hasil-akhir-database-fresh-clone-acuan).

## 0.7. Nama database bebas

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

## 0.8. Mengganti nama aplikasi

Cukup ubah satu baris di `.env`, lalu build ulang frontend:

```env
APP_NAME=Toko Saya
```

Yang ikut berubah otomatis:

| Tempat | Keterangan |
|---|---|
| Judul tab browser | Di-set runtime oleh `main.jsx` |
| Nama di sidebar (desktop + topbar mobile) | `App.jsx` |
| Nama di kartu login | `pages/Auth.jsx` |
| Teks "hubungi admin" & footer | `App.jsx` |
| `Config.AppName` backend | Dibaca `config/env.go` |

Setelah ubah `.env`:

```bash
task build-frontend      # development
# atau
task build               # production (frontend + backend)
```

lalu restart backend. `frontend/index.html` masih berisi `<title>Go Core</title>`
sebagai fallback sebelum JavaScript berjalan — ganti juga file itu bila ingin
judulnya benar sejak tab pertama dibuka.

## 0.9. Mengganti favicon dan logo

Satu file untuk dua kegunaan: **favicon tab browser** dan **logo di sidebar,
topbar mobile, serta kartu login**.

Langkah:

1. Siapkan file gambar (rasio 1:1, persegi, latar transparan atau putih,
   minimal 64×64 px; format bebas: SVG, PNG, WebP).
2. Simpan di `frontend/public/`, misalnya `frontend/public/logo-toko.png`.
   Folder `public/` disalin apa adanya ke `dist/`, jadi file di sana bisa
   dipakai tanpa import.
3. Arahkan `APP_LOGO` di `.env` ke nama file itu (harus diawali `/`):

```env
APP_LOGO=/logo-toko.png
```

4. Build ulang frontend lalu restart backend:

```bash
task build-frontend
```

5. Refresh browser (hard refresh `Ctrl+Shift+R`).

Bila ingin memakai nama file bawaan, cukup ganti isi
`frontend/public/favicon.svg` dan biarkan `APP_LOGO=/favicon.svg`.

Cek cepat:

- [ ] Judul tab memakai `APP_NAME`.
- [ ] Ikon tab memakai gambar baru.
- [ ] Logo muncul di sidebar, topbar mobile, dan kartu login.
- [ ] Gambar tidak gepeng (rasio 1:1) dan tidak pecah di layar kecil.
- [ ] Tidak ada huruf logo sebagai pengganti: kalau file gambar rusak atau
      hilang, `AppLogo` otomatis jatuh ke huruf pertama `APP_NAME`.

Masalah umum:

| Gejala | Penyebab & solusi |
|---|---|
| Ikon tab tetap logo lama | Cache browser agresif → pakai nama file baru (`favicon-toko.svg`) atau hard refresh |
| Gambar tidak muncul, hanya huruf | `APP_LOGO` salah path (harus `/nama-file` di `frontend/public/`) atau belum rebuild frontend |
| Logo pecah/gedang | Rasio gambar bukan 1:1 → crop dulu ke persegi |
| Judul tab masih "Go Core" | `APP_NAME` belum diubah, atau `frontend/index.html` masih memakai `<title>` lama |

Detail lengkap ada di README bagian
[1.9. Mengganti Nama Aplikasi](../README.md#19-mengganti-nama-aplikasi).

## 0.10. Engine lain (opsional)

PostgreSQL: `DB_CONNECTION=postgres` lalu `task migrate-postgres`
(skema `scripts/schema.postgres.sql`).

SQLite: `DB_CONNECTION=sqlite` + `DB_DATABASE=./data/tokodb.db` lalu
`task migrate-sqlite` (skema `scripts/schema.sqlite.sql`).

Ketiganya menghasilkan state awal yang sama: 10 tabel, 2 modul, 11 menu,
`ADMIN` 8 grant, 2 akun (`admin`/`admin` dan `user`/`user`), 3 template
notifikasi.

## 0.11. Kalau mentok saat setup

| Gejala | Penyebab & solusi |
|---|---|
| `sqlcmd` tidak dikenali | Install SQL Server Client Tools, atau jalankan skrip lewat SSMS |
| `Login failed for user` saat `task migrate` | `DB_USERNAME`/`DB_PASSWORD` di `.env` tidak sama dengan login SQL Server |
| Database tidak ditemukan | Nama di `DB_DATABASE` beda dengan database yang dibuat — keduanya harus sama persis |
| `Msg 1934 ... QUOTED_IDENTIFIER` saat query manual ke `CPMENU` | Tambahkan flag `-I` pada `sqlcmd` (skrip di `scripts/` sudah men-set sendiri) |
| `go:embed no matching files` | `frontend/dist` belum ada → `task build-frontend` sekali |
| Port 1067/5173 sudah dipakai | Ubah `APP_PORT` di `.env` (backend) atau `vite.config.js` (frontend dev) |
| Login ditolak padahal `seed-admin.sql` sudah dijalankan | Password `admin`/`admin` hanya untuk development; ganti lewat menu **User Account** → edit → Password baru |
| Sidebar kosong setelah login `admin` | Grant menu belum ada → buka **Role & Permission**, centang menu role `ADMIN`, **Simpan permission**, lalu login ulang |
| Halaman 404 pemandu saat klik menu | Folder `frontend/src/app/<mcontrol>/` belum dibuat → ikuti petunjuk di halaman itu |

Daftar lengkap ada di README bagian
[5. Troubleshooting](../README.md#5-troubleshooting).

---

# Bagian 1 — Daftarkan Modul dan Menu di Database

Urutannya wajib: **modul dulu, baru menu** — `CPMENU.MODULE` menunjuk ke
`CPMODULE.CODE`.

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
INSERT INTO dbo.CPMODULE (CODE, LABEL, SORT_ORDER)
SELECT N'TOKO', N'Toko', 3
WHERE NOT EXISTS (SELECT 1 FROM dbo.CPMODULE WHERE CODE = N'TOKO');
```

Modul gagal dihapus selama masih ada menu di dalamnya (UI memberi pesan
`409 Module is still used by menus`).

✅ **Checkpoint 1.1**

```sql
SELECT CODE, LABEL, SORT_ORDER FROM dbo.CPMODULE ORDER BY SORT_ORDER;
```

`TOKO` harus muncul di antara `SYSTEM` dan `REPORT`.

## 1.2. Buat menu di modul itu

via UI:

1. Login `admin`, buka menu **Modul & Menu**.
2. Klik **Menu baru**, isi:
   - Kode permission: `MENU_STOK` (wajib prefix `MENU_`, maks 40 karakter)
   - Modul (MODULE): pilih `TOKO` dari dropdown
   - Label tampil: `Stok`, Urutan: `1`
   - Jenis menu: `CHILD` (menu biasa) atau `PARENT` (bisa punya anak)
   - Parent: kosongkan (atau pilih menu `PARENT` kalau menu ini mau berada di bawahnya)
3. Simpan. Menu tercatat di `CPMENU`
   **tanpa auto-grant ke role mana pun** — admin mencentang manual di
   halaman Role (Bagian 5).

Menu yang sudah ada bisa diubah kapan saja: di daftar menu, klik ikon
pensil untuk mengubah **label tampil**, **urutan**, **modul**, atau
**parent**. Kode permission (`MENU_STOK`) tidak bisa diubah karena jadi
acuan grant role (folder terpisah: `app/stok/`) — untuk mengganti kode,
buat menu baru lalu hapus yang lama.

✅ **Checkpoint 1.2**

```sql
SELECT CODE, MODULE, LABEL, MCONTROL, MENU_KIND, PARENT_CODE FROM dbo.CPMENU WHERE CODE = 'MENU_STOK';
```

Harus menghasilkan satu baris dengan `MODULE = TOKO`, `MCONTROL = stok`, `MENU_KIND = CHILD`.

## 1.3. Alternatif via SQL

Satu permission mewakili satu menu; MCONTROL = nama folder datar. Cukup satu INSERT:

```sql
INSERT INTO dbo.CPMENU (CODE, MODULE, LABEL, MCONTROL, MENU_KIND, SORT_ORDER, PARENT_CODE)
VALUES (N'MENU_STOK', N'TOKO', N'Stok', N'stok', N'CHILD', 1, NULL);
```

Grant ke role tetap via matriks UI (atau `INSERT INTO dbo.CPPERMISSION
(ROLE_CODE, MENU_CODE) ...`) — hanya untuk menu `CHILD`; grant ke kode
`PARENT` ditolak `400` karena header tidak punya halaman sendiri.

✅ **Checkpoint 1.3**

```sql
SELECT CODE, MODULE, MCONTROL, MENU_KIND FROM dbo.CPMENU WHERE CODE = 'MENU_STOK';
```

Harus menghasilkan satu baris. (Jalur UI di 1.2 tidak menulis seed file —
cek ke database, bukan ke `migrate2_rbac.sql`.)

Untuk menu permanen bawaan, tambahkan seed yang sama di
`scripts/migrate2_rbac.sql` (proyek ini khusus SQL Server). Menu yang dibuat
lewat UI tidak perlu seed.

> Catatan: kolom **Parent** hanya muncul saat jenis menu `CHILD`, dan
> hanya berisi menu bertipe `PARENT` di modul yang sama. Menu `PARENT` tidak
> boleh punya parent, tidak boleh punya MCONTROL, dan tidak bisa diubah jadi
> `CHILD` selama masih punya anak. Menu parent = header buka-tutup, tanpa
> folder/halaman.

---

# Bagian 2 — Membuat Frontend Menu

Folder menu DATAR: `app/<mcontrol>/` (MCONTROL snake_case). Judul = LABEL di CPMENU.

## 2.1. Copy Template

```bash
cp -r tutorial/templates/frontend-menu frontend/src/app/stok
```

✅ **Checkpoint 2.1**

```bash
ls frontend/src/app/stok
```

Harus ada (5 file mirror langit_v2):

```text
api.js
controller.js
FRM.jsx
GRID.jsx
index.jsx
```

| File | Padanan langit_v2 | Yang diganti |
|---|---|---|
| `index.jsx` | `<mod>.js` (shell) | `MCONTROL`, `meta`, judul |
| `controller.js` | `C<mod>.js` (6 fungsi) | `mcontrol` (1 baris) |
| `GRID.jsx` | `GRID<mod>.js` | `COLUMNS_ITEMS`, `ID_FIELD`, details |
| `FRM.jsx` | `FRM<mod>.js` | `FORM_FIELDS`, `validate_field` |
| `api.js` | store proxy | path endpoint `<menu>` |

## 2.2. Isi 5 file mirror (cukup 4 titik)

Template sudah jadi shell + controller + GRID + FRM + api yang saling
tersambung. Anda hanya mengisi 4 titik (mirip clone modul di langit_v2):

| # | File | Titik yang diisi | Contoh stok |
|---|---|---|---|
| 1 | `index.jsx` | `MCONTROL` + `meta` + judul | `stok`, `Stok` |
| 2 | `controller.js` | `mcontrol` (1 baris) | `stok` |
| 3 | `GRID.jsx` | `COLUMNS_ITEMS` + `ID_FIELD` | kolom kode/nama/qty |
| 4 | `FRM.jsx` | `FORM_FIELDS` + `validate_field` | field kode/nama/qty |
| 5 | `api.js` | path endpoint | `/api/stok` |

**Titik 1 — `index.jsx`**: sesuaikan `MCONTROL` dan `meta`:

```jsx
const MCONTROL = 'stok'

export const meta = { label: 'Stok', icon: 'list', order: 1 }
```

`icon` memakai nama ikon dari `components/icons.jsx` (`list`, `users`,
`shield`, `bell`, `key`, `terminal`, `download`, ...). `order` menentukan
urutan dalam sidebar modul (makin kecil makin atas). Judul yang tampil
di halaman diambil dari `nvdata.label` (kolom `CPMENU.LABEL`) — `meta`
hanya fallback.

**Titik 2 — `controller.js`**: satu baris (6 fungsi lainnya jangan diubah):

```js
export const controller = createController({ mcontrol: 'stok' })
```

**Titik 3 — `GRID.jsx`**: isi kolom tabel (satu-satunya sumber nama field —
inilah yang mencegah bug kelas PPH di langit_v2, di mana handler masih
menyebut field modul lama):

```jsx
const COLUMNS_ITEMS = [
  { header: 'Kode', dataIndex: 'code', width: 110, editor: { xtype: 'textfield', allowBlank: false, maxLength: 40 } },
  { header: 'Nama', dataIndex: 'name', width: 200, editor: { xtype: 'textfield', allowBlank: false, maxLength: 100 } },
  { header: 'Qty', dataIndex: 'qty', width: 80, editor: { xtype: 'numberfield', allowBlank: false } },
]

const ID_FIELD = 'code'
```

`editor.allowBlank: false` otomatis menjadi aturan wajib-isi di
`handler_validasi_inline` — jangan tulis ulang nama field di validasi.

**Titik 4 — `FRM.jsx`**: isi field form + aturan validasi:

```jsx
const FORM_FIELDS = [
  { name: 'code', label: 'Kode', type: 'text', placeholder: 'mis. STK-001' },
  { name: 'name', label: 'Nama', type: 'text', placeholder: 'mis. Kaos polos' },
  { name: 'qty', label: 'Qty', type: 'number', placeholder: '0' },
]

const validate_field = [
  { field: 'code', type: 'text', msg: 'Kode tidak boleh kosong' },
  { field: 'name', type: 'text', msg: 'Nama tidak boleh kosong' },
  { field: 'qty', type: 'number', msg: 'Qty harus angka lebih dari 0' },
]
```

Pola yang dijaga di semua menu (mirror langit_v2):

| File | Isi |
|---|---|
| `index.jsx` | shell: toolbar (`btrefresh`/`New Input`/`Download`) + `items:[<GRID/>]` + `meta` |
| `controller.js` | 6 fungsi: `init`, `renderpage`, `formatAmount`, `formatDate`, `btrefresh_click`, `btnew_click` |
| `GRID.jsx` | `COLUMNS_ITEMS` + 7 handler `handler_rowbtn_*` / `handler_validasi_*` |
| `FRM.jsx` | `FORM_FIELDS` + `validate_field` + `handler_btsave/btdelete/validasi_input` |
| `api.js` | `read_data` / `process_create` / `process_update` / `process_delete` |

Semua komponen UI diambil dari barrel `app/shared/all.js` — jangan
meng-`import` file komponen satu per satu.

✅ **Checkpoint 2.2**

```bash
grep -n "export default\|export const meta" frontend/src/app/stok/index.jsx
```

Harus ada dua baris export:

```text
export const meta = ...
export default function ...
```

## 2.3. Isi `api.js` (4 method proxy)

Buka `frontend/src/app/stok/api.js`.

Satu file ini milik satu menu — padanan store proxy langit_v2 (satu URL +
method `read_data`/`process_*`). Ganti `<menu>` dengan path endpoint:

```js
// Fungsi menu Stok (modul TOKO, permission MENU_STOK).
import { apiRequest as request } from '../../api/client.js'

// read_data: GET list (dipakai store/load GRID). Selalu kembalikan array.
export async function read_data({ limit = 20, offset = 0 } = {}) {
  const data = await request(`/api/stok?limit=${limit}&offset=${offset}`, { auth: true })
  return Array.isArray(data) ? data : []
}

// process_create: POST tambah (dipakai FRM handler_btsave mode new).
export async function process_create(dtval) {
  return request('/api/stok', { method: 'POST', body: dtval, auth: true })
}

// process_update: PUT ubah (dipakai GRID handler_rowbtn_save + FRM edit).
export async function process_update(code, dtval) {
  return request(`/api/stok/${code}`, { method: 'PUT', body: dtval, auth: true })
}

// process_delete: DELETE hapus (dipakai GRID + FRM).
export async function process_delete(code) {
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
grep -rn "<menu>" frontend/src/app/stok || echo OK-tidak-ada-placeholder
```

Harus menghasilkan:

```text
OK-tidak-ada-placeholder
```

## 2.4. Pahami Cara Menu Ditemukan (DB + folder)

Akses menu **selalu dari database**, bukan dari nama folder:

```text
CPMENU.CODE = MENU_STOK  →  grant di CPPERMISSION (ROLE EDITOR ✓)
CPMENU.MCONTROL = stok   →  folder frontend/src/app/stok/ (halaman)
CPMENU.MODULE = TOKO     →  section sidebar
```

Alurnya saat user login:

1. `GET /api/users/me` mengembalikan `menus` (entri `CPMENU` yang
   ter-grant ke role user, plus header PARENT otomatis).
2. `registry.js` mencocokkan tiap entri CHILD dengan folder lokal via
   `MCONTROL`. Cocok → halaman asli; tidak cocok → halaman 404 pemandu.
3. Sidebar dikelompokkan per `MODULE`.

Jadi tidak perlu menambah daftar menu manual di `registry.js`. Folder menu
selalu datar satu level: `app/<mcontrol>/` (tanpa folder modul/perantara).
Penempatan anak di bawah parent diatur lewat `CPMENU.PARENT_CODE`, bukan
lewat folder.

✅ **Checkpoint 2.4**

Setelah Bagian 4 (build + grant + login ulang), cek di DevTools:

```bash
# entri MENU_STOK ikut dari /me bila sudah di-grant
```

Bila menu 404: folder `frontend/src/app/stok/` belum ada atau
`MCONTROL`-nya beda huruf. Bila menu hilang total: grant belum dicentang
atau user belum login ulang.

## 2.5. Pahami Alur 404 Pemandu

Urutan yang benar: **daftar di database dulu (Bagian 1), folder belakangan
(Bagian 2)**. Di jeda itu, bila menu sudah dicentang ke suatu role, user
role tersebut melihat halaman **404 pemandu** (bukan menu hilang diam-diam):

- Permission yang hilang + path folder persis yang harus dibuat, mis.
  `frontend/src/app/stok/`
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

> ✅ **Checkpoint 4.2** — setelah restart, `curl
> http://localhost:1067/healthz` balas `{"status":"ok"}` dan log backend
> menulis `database connected` tanpa error. Bila health gagal: binary lama
> masih jalan (hentikan dulu) atau port dipakai proses lain.

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
   banyak menu sekaligus. Baris menu `PARENT` (bila ada) tampil sebagai header
   tanpa centang — centang saja menu CHILD di bawahnya, header-nya ikut
   ter-include otomatis.
6. Klik **Simpan permission**.

Matriks menampilkan **semua** menu `CHILD` di `CPMENU`, termasuk menu contoh
yang belum pernah di-grant. Jadi administrator bebas memberi akses ke menu mana
pun — begitulah cara memberi akses ke menu contoh `REPORT` bila diinginkan.
Header `PARENT` hanya tampil sebagai judul section dan tidak bisa di-grant.

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

> ✅ **Checkpoint 5.1**
>
> ```sql
> SELECT CODE, NAME FROM dbo.CPROLE WHERE CODE = 'EDITOR';  -- harus 1 baris
> ```

## 5.2. Beri Akses Menu ke Role Itu

Centang menu yang boleh dipakai role `EDITOR` di halaman Role & Permission
(langkah sama seperti 4.3), lalu **Simpan permission**. Contoh centang:

```text
MENU_USERS
MENU_STOK
```

Role tanpa satu pun centang = user akan melihat halaman kosong
"hubungi admin".

> ✅ **Checkpoint 5.2**
>
> ```sql
> SELECT MENU_CODE FROM dbo.CPPERMISSION
> WHERE ROLE_CODE = 'EDITOR' ORDER BY MENU_CODE;
> ```
>
> Menu yang dicentang harus muncul di hasil query.

## 5.3. Buat User Baru

via UI (disarankan):

1. Login admin.
2. Buka menu **User Account**.
3. Klik **New Input** (tombol tambah user).
4. Isi username, email, dan password (minimal 8 karakter).
5. Pilih role `EDITOR` pada dropdown, lalu **Save**.

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

## 5.5. Admin Mengganti Password User

Tidak ada halaman "ubah password" mandiri untuk user biasa. Password user
diubah oleh admin dari menu **User Account**:

1. Login admin.
2. Buka menu **User Account**.
3. Klik ikon edit pada user yang dikehendaki.
4. Isi kolom **Password baru** (kosongkan bila tidak ingin mengganti).
5. Simpan.

✅ **Checkpoint 5.5**

```sql
-- Password tidak pernah ditampilkan; yang bisa dicek hanya usernya masih aktif:
SELECT USERNAME, EMAIL, ROLE_CODE, UPDATED_AT
FROM dbo.CPUSER
WHERE USERNAME = 'editor';
```

Password diganti **tidak** otomatis logout user dari device yang sedang login.
Untuk memaksa user keluar, di halaman **User Account** klik tombol
**Keluarkan semua sesi** (ikon tooltip) pada user tersebut — semua refresh
token-nya dihapus sehingga harus login ulang.

## 5.6. Aturan Validasi User (biar tidak ditolak backend)

| Field | Aturan | Pesan error dari backend |
|---|---|---|
| Username | 3–50 karakter, tanpa karakter kontrol/newline; unik secara case-insensitive, disimpan apa adanya | `username must be at least 3 characters` / `username already taken` |
| Email | Format email valid, maksimal 254 karakter, tanpa `..`; otomatis lowercase dan unik | `invalid email format` / `email already taken` |
| Password | Minimal **8** karakter, maksimal 72 byte (batas bcrypt) | `password must be at least 8 characters` / `password must not exceed 72 bytes` |
| Kode user | Dibuat otomatis backend (`USR-XXXXXXXX`), tidak perlu diisi | — |
| Role | Wajib ada di `CPROLE`; role `ADMIN`/`USER` bawaan tidak bisa dihapus | `role not found` |

Kolom edit user **tidak** mengubah role — pindah role lewat dropdown role di
halaman **User Account** (atau `PUT /api/users/{code}/role`).
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

> ✅ **Checkpoint 6** — `SELECT * FROM dbo.CPPERMISSION WHERE MENU_CODE =
> 'MENU_STOK' AND ROLE_CODE = 'EDITOR'` kosong, dan sidebar user `EDITOR`
> tidak lagi menampilkan grup/menu tersebut.

---

# Bagian 7 — Menghapus Menu

Jika menu tidak diperlukan, hapus via UI **Modul & Menu**
(tombol hapus; gagal bila masih dipakai role atau punya menu anak). Itu
menghapus baris `CPMENU` sekaligus grant-nya (ikut CASCADE).

Pembersihan manual (bila perlu):

1. Hapus folder frontend:

```bash
rm -r frontend/src/app/stok
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
   UI **Modul & Menu** atau `DELETE FROM dbo.CPMODULE WHERE CODE = N'TOKO';`.
6. Jalankan `task build-frontend`.
7. Refresh/login ulang.

> ✅ **Checkpoint 7** — `SELECT COUNT(*) FROM dbo.CPMENU WHERE CODE =
> 'MENU_STOK'` = 0, folder `frontend/src/app/stok` sudah tidak ada, dan
> `npm run build` tetap `✓ built` (tidak ada import menggantung).

---

# Checklist Akhir

Setup (khusus clone baru):

- [ ] `DB_DATABASE` di `.env` sama persis dengan nama database yang dibuat.
- [ ] `task migrate` + `scripts/seed-admin.sql` sudah dijalankan (10 tabel `CP%`).
- [ ] Login `admin` / `admin` berhasil (ganti passwordnya setelah itu).
- [ ] Login `user` / `user` menunjukkan halaman kosong "hubungi admin".
- [ ] Sidebar menampilkan 2 section modul dan 11 menu di matriks Role & Permission.
- [ ] `APP_NAME` di `.env` disesuaikan bila nama aplikasi diganti (judul tab,
      sidebar, kartu login).
- [ ] `APP_LOGO` di `.env` + file di `frontend/public/` sudah diganti bila
      favicon/logo sudah dikustomisasi; `task build-frontend` dijalankan ulang.
- [ ] `task build` + `./app.exe` melayani aplikasi di `http://localhost:1067` dan
      `curl http://localhost:1067/healthz` balas `{"status":"ok"}`.
- [ ] `npm run lint` dan `task test` (go vet + test) tidak error.

Modul & menu:

- [ ] Modul dibuat dulu di `CPMODULE` (kode UPPERCASE), baru menu di `CPMENU`.
- [ ] `CPMENU.MODULE` = section sidebar (FK ke `CPMODULE`); `CPMENU.MCONTROL` = nama folder.
- [ ] MCONTROL snake_case = nama folder datar (`stok` ↔ `app/stok/`); permission bebas (mis. `MENU_STOK`).
- [ ] Menu CHILD diletakkan di `frontend/src/app/<mcontrol>/` (datar); PARENT tanpa folder.
- [ ] `index.jsx` memiliki `export const meta` dan `export default`;
      `controller.js` 6 fungsi; `GRID.jsx` punya `COLUMNS_ITEMS`;
      `FRM.jsx` punya `validate_field`.
- [ ] `api.js` memakai `read_data`/`process_*` + `auth: true` untuk endpoint
      privat, tanpa placeholder `<menu>`.
- [ ] Menu bisa diedit dari **Modul & Menu** (label/urutan/modul/jenis/parent) tanpa dihapus-buat.
- [ ] Menu child punya `PARENT_CODE` yang menunjuk menu `PARENT` se-modul; menu `PARENT` tanpa parent dan tanpa mcontrol.
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
-- CPAUDITLOG, CPMENU, CPMODULE, CPNOTIFLOG, CPNOTIFTEMPLATE,
-- CPPERMISSION, CPREFRESHTOKEN, CPROLE, CPSYSLOG, CPUSER

-- 2 modul bawaan (SYSTEM operasional + REPORT contoh tes tampilan):
SELECT CODE, LABEL, SORT_ORDER FROM dbo.CPMODULE ORDER BY SORT_ORDER;
-- SYSTEM | System | 1
-- REPORT | Report | 2

-- 11 menu (8 SYSTEM + 3 contoh REPORT); permission = menu (1:1):
SELECT COUNT(*) FROM dbo.CPMENU;  -- 11
-- Contoh parent-child: MENU_KEUANGAN = PARENT, MENU_ARUS_KAS = CHILD
-- dengan PARENT_CODE = MENU_KEUANGAN.
SELECT CODE, MODULE, MENU_KIND, PARENT_CODE FROM dbo.CPMENU
WHERE MODULE = 'REPORT' ORDER BY SORT_ORDER;

-- Grant: ADMIN 8, USER 0 (akses USER selalu manual via matriks).
-- Tiga menu REPORT sengaja tanpa akses role mana pun.
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
