-- Skema PostgreSQL (fresh install). Relasi via CODE, tabel huruf besar prefix CP.
-- Jalankan sekali: psql -h localhost -U <user> -d <db> -f scripts/schema.postgres.sql
-- Env: DB_CONNECTION=postgres (+ DB_HOST/DB_PORT/DB_DATABASE/DB_USERNAME/DB_PASSWORD)

CREATE TABLE IF NOT EXISTS CPROLE (
  ID SERIAL PRIMARY KEY,
  CODE TEXT NOT NULL UNIQUE,
  NAME TEXT NOT NULL,
  CREATED_AT TIMESTAMP NOT NULL DEFAULT NOW(),
  UPDATED_AT TIMESTAMP NOT NULL DEFAULT NOW()
);
INSERT INTO CPROLE (CODE, NAME)
SELECT 'ADMIN', 'Administrator'
WHERE NOT EXISTS (SELECT 1 FROM CPROLE WHERE CODE = 'ADMIN');
INSERT INTO CPROLE (CODE, NAME)
SELECT 'USER', 'Pengguna'
WHERE NOT EXISTS (SELECT 1 FROM CPROLE WHERE CODE = 'USER');

CREATE TABLE IF NOT EXISTS CPUSER (
  ID SERIAL PRIMARY KEY,
  CODE TEXT NOT NULL UNIQUE,
  USERNAME VARCHAR(50) NOT NULL UNIQUE,
  EMAIL VARCHAR(255) NOT NULL UNIQUE,
  PASSWORD VARCHAR(255) NOT NULL,
  ROLE_CODE TEXT NOT NULL DEFAULT 'USER' REFERENCES CPROLE (CODE),
  CREATED_AT TIMESTAMP NOT NULL DEFAULT NOW(),
  UPDATED_AT TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS CPREFRESHTOKEN (
  ID SERIAL PRIMARY KEY,
  USER_CODE TEXT NOT NULL REFERENCES CPUSER (CODE) ON DELETE CASCADE,
  TOKEN VARCHAR(512) NOT NULL UNIQUE,
  EXPIRES_AT TIMESTAMP NOT NULL,
  CREATED_AT TIMESTAMP NOT NULL DEFAULT NOW()
);

-- (CPPERMISSION sebagai tabel grant dibuat setelah CPMENU di bawah,
-- karena FK-nya menunjuk ke sana.)

-- CPMODULE: master modul. Satu baris = satu section sidebar + satu folder
-- di frontend/src/app/<MODULE>/. CPMENU.MODULE ber-FK ke sini (modul
-- dibuat dulu, baru menunya).
CREATE TABLE IF NOT EXISTS CPMODULE (
  ID SERIAL PRIMARY KEY,
  CODE VARCHAR(40) NOT NULL UNIQUE,
  LABEL VARCHAR(100) NOT NULL,
  SORT_ORDER INT NOT NULL DEFAULT 99,
  CREATED_AT TIMESTAMP NOT NULL DEFAULT NOW(),
  UPDATED_AT TIMESTAMP NOT NULL DEFAULT NOW()
);
INSERT INTO CPMODULE (CODE, LABEL, SORT_ORDER)
SELECT 'SYSTEM', 'System', 1
WHERE NOT EXISTS (SELECT 1 FROM CPMODULE WHERE CODE = 'SYSTEM');
INSERT INTO CPMODULE (CODE, LABEL, SORT_ORDER)
SELECT 'REPORT', 'Report', 2
WHERE NOT EXISTS (SELECT 1 FROM CPMODULE WHERE CODE = 'REPORT');

-- CPMENU: registry menu. MODULE = kode modul (FK ke CPMODULE.CODE),
-- menentukan section sidebar level-1 (MODULE tampil sebagai section header).
-- MCONTROL = nama folder frontend (snake_case, mis. user_account): halaman
-- menu tinggal di frontend/src/app/<mcontrol>/index.jsx — datar, tanpa
-- folder modul perantara. NULL = menu PARENT (header buka-tutup di sidebar,
-- tanpa halaman/folder). Judul tampil = LABEL.
-- MENU_KIND: 'PARENT' = header buka-tutup (MCONTROL wajib NULL, tidak boleh
-- punya PARENT_CODE), 'CHILD' = item biasa (MCONTROL wajib unik, boleh punya
-- PARENT_CODE yang menunjuk menu PARENT se-modul, tanpa siklus).
CREATE TABLE IF NOT EXISTS CPMENU (
  ID SERIAL PRIMARY KEY,
  CODE VARCHAR(40) NOT NULL UNIQUE,
  MODULE VARCHAR(40) NOT NULL REFERENCES CPMODULE (CODE),
  LABEL VARCHAR(100) NOT NULL,
  MCONTROL VARCHAR(40) UNIQUE,
  MENU_KIND VARCHAR(10) NOT NULL DEFAULT 'CHILD' CHECK (MENU_KIND IN ('PARENT', 'CHILD')),
  SORT_ORDER INT NOT NULL DEFAULT 99,
  PARENT_CODE VARCHAR(40) REFERENCES CPMENU (CODE),
  CREATED_AT TIMESTAMP NOT NULL DEFAULT NOW(),
  UPDATED_AT TIMESTAMP NOT NULL DEFAULT NOW(),
  UNIQUE (MODULE, CODE),
  CHECK ((MENU_KIND = 'PARENT' AND MCONTROL IS NULL) OR (MENU_KIND = 'CHILD' AND MCONTROL IS NOT NULL))
);

-- Seed menu bawaan (urut sesuai sidebar) + contoh parent-child di REPORT.
-- MCONTROL snake_case = nama folder frontend/src/app/<mcontrol>/.
-- PARENT (MENU_KEUANGAN) tanpa mcontrol — header buka-tutup saja.
INSERT INTO CPMENU (CODE, MODULE, LABEL, MCONTROL, MENU_KIND, SORT_ORDER, PARENT_CODE) VALUES
  ('MENU_USERS',         'SYSTEM', 'User Account',       'users',           'CHILD',  1, NULL),
  ('MENU_MODUL',         'SYSTEM', 'Modul & Menu',       'modul_menu',      'CHILD',  2, NULL),
  ('MENU_ROLES',         'SYSTEM', 'Role & Permission',  'role_permission', 'CHILD',  3, NULL),
  ('MENU_SESSIONS',      'SYSTEM', 'Sesi & Auth',        'sesi_auth',       'CHILD',  4, NULL),
  ('MENU_AUDIT',         'SYSTEM', 'Audit Log',          'audit_log',       'CHILD',  5, NULL),
  ('MENU_SECURITY',      'SYSTEM', 'Security Center',    'security_center', 'CHILD',  6, NULL),
  ('MENU_SYSLOG',        'SYSTEM', 'System Log',         'system_log',      'CHILD',  7, NULL),
  ('MENU_NOTIFICATIONS', 'SYSTEM', 'Notifikasi',         'notifikasi',      'CHILD',  8, NULL),
  ('MENU_LAPORAN',       'REPORT', 'Laporan',            'laporan',         'CHILD',  1, NULL),
  ('MENU_KEUANGAN',      'REPORT', 'Keuangan',           NULL,              'PARENT', 2, NULL),
  ('MENU_ARUS_KAS',      'REPORT', 'Arus Kas',           'arus_kas',        'CHILD',  3, 'MENU_KEUANGAN')
ON CONFLICT (CODE) DO NOTHING;

-- CPPERMISSION: satu-satunya tabel relasi grant (ROLE_CODE -> MENU_CODE).
-- Definisi menu tinggal di CPMENU; tidak ada tabel definisi terpisah.
CREATE TABLE IF NOT EXISTS CPPERMISSION (
  ROLE_CODE TEXT NOT NULL REFERENCES CPROLE (CODE) ON DELETE CASCADE,
  MENU_CODE VARCHAR(40) NOT NULL REFERENCES CPMENU (CODE) ON DELETE CASCADE,
  CREATED_AT TIMESTAMP NOT NULL DEFAULT NOW(),
  PRIMARY KEY (ROLE_CODE, MENU_CODE)
);

-- ADMIN mendapat semua menu CHILD kecuali contoh tes tampilan (tanpa akses).
-- Menu PARENT tidak pernah di-grant: header buka-tutup tanpa halaman sendiri,
-- aksesnya menyusul otomatis dari menu CHILD di bawahnya.
-- USER nol mapping (manual via UI).
INSERT INTO CPPERMISSION (ROLE_CODE, MENU_CODE)
SELECT 'ADMIN', CODE FROM CPMENU
WHERE MENU_KIND = 'CHILD'
  AND CODE NOT IN ('MENU_LAPORAN', 'MENU_ARUS_KAS')
ON CONFLICT DO NOTHING;
-- Bersihkan grant PARENT sisa database lama (matriks kini hanya mencentang CHILD).
DELETE FROM CPPERMISSION
WHERE MENU_CODE IN (SELECT CODE FROM CPMENU WHERE MENU_KIND = 'PARENT');

-- Matriks role x modul x menu dibaca via JOIN (CPROLE x CPMENU LEFT JOIN
-- CPPERMISSION); tulis grant tetap ke CPPERMISSION (tidak ada sinkron ganda).

CREATE TABLE IF NOT EXISTS CPAUDITLOG (
  ID SERIAL PRIMARY KEY,
  CODE VARCHAR(20) NOT NULL UNIQUE,
  ACTOR_CODE VARCHAR(20),
  ACTION VARCHAR(60) NOT NULL,
  ENTITY VARCHAR(40) NOT NULL,
  ENTITY_CODE VARCHAR(60),
  DETAIL VARCHAR(1000),
  IP_ADDRESS VARCHAR(60),
  CREATED_AT TIMESTAMP NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS IX_CPAUDIT_ACTION ON CPAUDITLOG (ACTION, CREATED_AT DESC);
CREATE INDEX IF NOT EXISTS IX_CPAUDIT_ACTOR ON CPAUDITLOG (ACTOR_CODE, CREATED_AT DESC);

CREATE TABLE IF NOT EXISTS CPSYSLOG (
  ID SERIAL PRIMARY KEY,
  CODE VARCHAR(20) NOT NULL UNIQUE,
  LEVEL VARCHAR(10) NOT NULL,
  SOURCE VARCHAR(100) NOT NULL,
  MESSAGE VARCHAR(1000) NOT NULL,
  CREATED_AT TIMESTAMP NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS IX_CPSYS_LEVEL ON CPSYSLOG (LEVEL, CREATED_AT DESC);

CREATE TABLE IF NOT EXISTS CPNOTIFTEMPLATE (
  ID SERIAL PRIMARY KEY,
  CODE VARCHAR(20) NOT NULL UNIQUE,
  NAME VARCHAR(100) NOT NULL,
  CHANNEL VARCHAR(10) NOT NULL,
  SUBJECT VARCHAR(200),
  BODY VARCHAR(4000) NOT NULL,
  IS_ACTIVE BOOLEAN NOT NULL DEFAULT TRUE,
  CREATED_AT TIMESTAMP NOT NULL DEFAULT NOW(),
  UPDATED_AT TIMESTAMP NOT NULL DEFAULT NOW()
);

INSERT INTO CPNOTIFTEMPLATE (CODE, NAME, CHANNEL, SUBJECT, BODY) VALUES
  ('NTPL-WELCOME', 'Selamat datang', 'EMAIL', 'Selamat datang di Go Core, {{nama}}!',
   'Halo {{nama}},

Akun Anda ({{kode}}) berhasil dibuat dengan role {{role}}.
Silakan login untuk mulai menggunakan aplikasi.

Salam,
Tim Go Core'),
  ('NTPL-RESET', 'Reset password', 'EMAIL', 'Permintaan reset password akun {{kode}}',
   'Halo {{nama}},

Kami menerima permintaan reset password untuk akun {{kode}}.
Abaikan pesan ini bila Anda tidak memintanya.

Salam,
Tim Go Core'),
  ('NTPL-ALERT', 'Peringatan keamanan', 'INAPP', 'Peringatan keamanan',
   'Aktivitas penting terdeteksi pada akun {{kode}}: {{detail}}')
ON CONFLICT (CODE) DO NOTHING;

CREATE TABLE IF NOT EXISTS CPNOTIFLOG (
  ID SERIAL PRIMARY KEY,
  CODE VARCHAR(20) NOT NULL UNIQUE,
  TEMPLATE_CODE VARCHAR(20),
  CHANNEL VARCHAR(10) NOT NULL,
  RECIPIENT VARCHAR(255) NOT NULL,
  SUBJECT VARCHAR(200),
  BODY VARCHAR(4000) NOT NULL,
  STATUS VARCHAR(10) NOT NULL,
  ERROR VARCHAR(500),
  CREATED_AT TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Seed 2 akun awal (sama seperti scripts/seed-admin.sql untuk SQL Server):
-- admin/admin (ADMIN) + user/user (USER). WAJIB ganti password setelah login
-- pertama, dan jangan dipakai di production apa adanya.
-- Hash di bawah = bcrypt; sama persis dengan seed SQL Server.
INSERT INTO CPUSER (CODE, USERNAME, EMAIL, PASSWORD, ROLE_CODE)
SELECT 'USR-000004', 'admin', 'admin@example.com',
  '$2a$10$vq5MqkuG4/mD/twgYMBsxeOwbrmL3xE88rIpOEzMYt./TmL8xKAqO', 'ADMIN'
WHERE NOT EXISTS (SELECT 1 FROM CPUSER WHERE USERNAME = 'admin');
INSERT INTO CPUSER (CODE, USERNAME, EMAIL, PASSWORD, ROLE_CODE)
SELECT 'USR-USER0001', 'user', 'user@example.com',
  '$2a$10$MpKcDZmTSotcNefe53ODl.3frNUr0E/fuU/cpAKPglu85tG0Niyr6', 'USER'
WHERE NOT EXISTS (SELECT 1 FROM CPUSER WHERE USERNAME = 'user');
