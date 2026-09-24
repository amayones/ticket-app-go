-- Migrasi 2: RBAC + Audit Log + System Log + Notifikasi + sesi admin.
--   CPPERMISSION      master permission (CODE, GROUP)
--   CPROLEPERMISSION  mapping ROLE_CODE -> PERMISSION_CODE
--   CPAUDITLOG        jejak aksi user (siapa, apa, kapan, dari IP mana)
--   CPSYSLOG          log error/sistem aplikasi (pengganti baca file log)
--   CPNOTIFTEMPLATE   template notifikasi (email/push/in-app, dukung {{var}})
--   CPNOTIFLOG        riwayat pengiriman notifikasi
-- Sesi memakai CPREFRESHTOKEN yang sudah ada (read/revoke via API).
--
-- Aman diulang (idempotent). Jalankan:
--   sqlcmd -S localhost,1433 -U <user> -P <pass> -d Go -C -i scripts/migrate2_rbac.sql
-- Atau: task migrate (menjalankan migrate.sql lalu file ini berurutan)
SET XACT_ABORT ON;
BEGIN TRAN;

-- 1. CPPERMISSION ------------------------------------------------------------
IF OBJECT_ID(N'dbo.CPPERMISSION', N'U') IS NULL
CREATE TABLE dbo.CPPERMISSION (
  ID INT IDENTITY(1,1) NOT NULL,
  CODE NVARCHAR(40) NOT NULL,
  NAME NVARCHAR(100) NOT NULL,
  PERMGROUP NVARCHAR(40) NOT NULL,
  DESCRIPTION NVARCHAR(255) NULL,
  CREATED_AT DATETIME NOT NULL CONSTRAINT DF_CPPERM_CREATED DEFAULT GETDATE(),
  CONSTRAINT PK_CPPERMISSION PRIMARY KEY CLUSTERED (ID),
  CONSTRAINT UQ_CPPERM_CODE UNIQUE NONCLUSTERED (CODE)
);

-- Seed permission (CODE, NAME, GROUP, DESCRIPTION)
DECLARE @perms TABLE (CODE NVARCHAR(40), NAME NVARCHAR(100), PERMGROUP NVARCHAR(40), DESCRIPTION NVARCHAR(255));
INSERT INTO @perms VALUES
  (N'USER_READ',        N'Lihat user',            N'USER',         N'Melihat daftar & detail akun'),
  (N'USER_CREATE',      N'Buat user',             N'USER',         N'Mendaftarkan akun baru'),
  (N'USER_UPDATE',      N'Ubah user lain',        N'USER',         N'Mengubah akun milik user lain (akun sendiri selalu boleh)'),
  (N'USER_DELETE',      N'Hapus user lain',       N'USER',         N'Menghapus akun milik user lain (akun sendiri selalu boleh)'),
  (N'USER_ROLE_ASSIGN', N'Atur role user',        N'USER',         N'Mengganti role akun'),
  (N'ROLE_READ',        N'Lihat role',            N'ROLE',         N'Melihat daftar role & permission'),
  (N'ROLE_MANAGE',      N'Kelola role',           N'ROLE',         N'Membuat & menghapus role'),
  (N'PERMISSION_ASSIGN',N'Atur permission role',  N'ROLE',         N'Mencentang permission milik role'),
  (N'SESSION_READ',     N'Lihat sesi sendiri',    N'SESSION',      N'Melihat sesi login milik sendiri'),
  (N'SESSION_REVOKE',   N'Cabut sesi sendiri',    N'SESSION',      N'Mencabut sesi login milik sendiri'),
  (N'SESSION_MANAGE',   N'Kelola semua sesi',     N'SESSION',      N'Melihat & mencabut sesi semua user'),
  (N'AUDIT_READ',       N'Lihat audit log',       N'AUDIT',        N'Melihat jejak aksi user'),
  (N'SECURITY_READ',    N'Lihat security center', N'SECURITY',     N'Melihat ringkasan keamanan'),
  (N'SYSLOG_READ',      N'Lihat system log',      N'SYSLOG',       N'Melihat log error/sistem'),
  (N'SYSLOG_MANAGE',    N'Hapus system log',      N'SYSLOG',       N'Menghapus log lama'),
  (N'NOTIF_READ',       N'Lihat notifikasi',      N'NOTIFICATION', N'Melihat template & riwayat notifikasi'),
  (N'NOTIF_MANAGE',     N'Kelola template',       N'NOTIFICATION', N'Membuat, mengubah, menonaktifkan template'),
  (N'NOTIF_SEND',       N'Kirim notifikasi',      N'NOTIFICATION', N'Mengirim / test-kirim notifikasi');

INSERT INTO dbo.CPPERMISSION (CODE, NAME, PERMGROUP, DESCRIPTION)
SELECT CODE, NAME, PERMGROUP, DESCRIPTION FROM @perms p
WHERE NOT EXISTS (SELECT 1 FROM dbo.CPPERMISSION x WHERE x.CODE = p.CODE);

-- 2. CPROLEPERMISSION ----------------------------------------------------------
IF OBJECT_ID(N'dbo.CPROLEPERMISSION', N'U') IS NULL
CREATE TABLE dbo.CPROLEPERMISSION (
  ROLE_CODE NVARCHAR(20) NOT NULL,
  PERMISSION_CODE NVARCHAR(40) NOT NULL,
  CREATED_AT DATETIME NOT NULL CONSTRAINT DF_CRP_CREATED DEFAULT GETDATE(),
  CONSTRAINT PK_CPROLEPERMISSION PRIMARY KEY CLUSTERED (ROLE_CODE, PERMISSION_CODE),
  CONSTRAINT FK_CRP_ROLE FOREIGN KEY (ROLE_CODE) REFERENCES dbo.CPROLE (CODE) ON DELETE CASCADE,
  CONSTRAINT FK_CRP_PERM FOREIGN KEY (PERMISSION_CODE) REFERENCES dbo.CPPERMISSION (CODE) ON DELETE CASCADE
);

-- ADMIN = semua permission
INSERT INTO dbo.CPROLEPERMISSION (ROLE_CODE, PERMISSION_CODE)
SELECT N'ADMIN', CODE FROM dbo.CPPERMISSION
WHERE NOT EXISTS (
  SELECT 1 FROM dbo.CPROLEPERMISSION x WHERE x.ROLE_CODE = N'ADMIN' AND x.PERMISSION_CODE = CPPERMISSION.CODE
);

-- USER = hak dasar (aksi milik sendiri tetap dijaga requireSelf di API)
INSERT INTO dbo.CPROLEPERMISSION (ROLE_CODE, PERMISSION_CODE)
SELECT N'USER', v.CODE FROM (VALUES
  (N'USER_READ'), (N'USER_UPDATE'), (N'USER_DELETE'),
  (N'SESSION_READ'), (N'SESSION_REVOKE')
) v(CODE)
WHERE NOT EXISTS (
  SELECT 1 FROM dbo.CPROLEPERMISSION x WHERE x.ROLE_CODE = N'USER' AND x.PERMISSION_CODE = v.CODE
);

-- 3. CPAUDITLOG ------------------------------------------------------------------
IF OBJECT_ID(N'dbo.CPAUDITLOG', N'U') IS NULL
CREATE TABLE dbo.CPAUDITLOG (
  ID INT IDENTITY(1,1) NOT NULL,
  CODE NVARCHAR(20) NOT NULL,
  ACTOR_CODE NVARCHAR(20) NULL,          -- CODE CPUSER pelaku (NULL = sistem/login-gagal)
  ACTION NVARCHAR(60) NOT NULL,          -- LOGIN, REGISTER, UPDATE_USER, ...
  ENTITY NVARCHAR(40) NOT NULL,          -- USER, ROLE, SESSION, NOTIFICATION, ...
  ENTITY_CODE NVARCHAR(60) NULL,         -- kode objek yang dikenai aksi
  DETAIL NVARCHAR(1000) NULL,
  IP_ADDRESS NVARCHAR(60) NULL,
  CREATED_AT DATETIME NOT NULL CONSTRAINT DF_CPAUDIT_CREATED DEFAULT GETDATE(),
  CONSTRAINT PK_CPAUDITLOG PRIMARY KEY CLUSTERED (ID),
  CONSTRAINT UQ_CPAUDIT_CODE UNIQUE NONCLUSTERED (CODE)
);
IF NOT EXISTS (SELECT 1 FROM sys.indexes WHERE name = N'IX_CPAUDIT_ACTION' AND object_id = OBJECT_ID(N'dbo.CPAUDITLOG'))
  CREATE NONCLUSTERED INDEX IX_CPAUDIT_ACTION ON dbo.CPAUDITLOG (ACTION, CREATED_AT DESC);
IF NOT EXISTS (SELECT 1 FROM sys.indexes WHERE name = N'IX_CPAUDIT_ACTOR' AND object_id = OBJECT_ID(N'dbo.CPAUDITLOG'))
  CREATE NONCLUSTERED INDEX IX_CPAUDIT_ACTOR ON dbo.CPAUDITLOG (ACTOR_CODE, CREATED_AT DESC);

-- 4. CPSYSLOG ---------------------------------------------------------------------
IF OBJECT_ID(N'dbo.CPSYSLOG', N'U') IS NULL
CREATE TABLE dbo.CPSYSLOG (
  ID INT IDENTITY(1,1) NOT NULL,
  CODE NVARCHAR(20) NOT NULL,
  LEVEL NVARCHAR(10) NOT NULL,           -- ERROR, WARN, INFO
  SOURCE NVARCHAR(100) NOT NULL,         -- handler/service asal error
  MESSAGE NVARCHAR(1000) NOT NULL,
  CREATED_AT DATETIME NOT NULL CONSTRAINT DF_CPSYS_CREATED DEFAULT GETDATE(),
  CONSTRAINT PK_CPSYSLOG PRIMARY KEY CLUSTERED (ID),
  CONSTRAINT UQ_CPSYS_CODE UNIQUE NONCLUSTERED (CODE)
);
IF NOT EXISTS (SELECT 1 FROM sys.indexes WHERE name = N'IX_CPSYS_LEVEL' AND object_id = OBJECT_ID(N'dbo.CPSYSLOG'))
  CREATE NONCLUSTERED INDEX IX_CPSYS_LEVEL ON dbo.CPSYSLOG (LEVEL, CREATED_AT DESC);

-- 5. CPNOTIFTEMPLATE ----------------------------------------------------------------
IF OBJECT_ID(N'dbo.CPNOTIFTEMPLATE', N'U') IS NULL
CREATE TABLE dbo.CPNOTIFTEMPLATE (
  ID INT IDENTITY(1,1) NOT NULL,
  CODE NVARCHAR(20) NOT NULL,
  NAME NVARCHAR(100) NOT NULL,
  CHANNEL NVARCHAR(10) NOT NULL,         -- EMAIL, PUSH, INAPP
  SUBJECT NVARCHAR(200) NULL,
  BODY NVARCHAR(4000) NOT NULL,          -- dukung variabel {{nama}}, {{kode}}, ...
  IS_ACTIVE BIT NOT NULL CONSTRAINT DF_CPNT_ACTIVE DEFAULT 1,
  CREATED_AT DATETIME NOT NULL CONSTRAINT DF_CPNT_CREATED DEFAULT GETDATE(),
  UPDATED_AT DATETIME NOT NULL CONSTRAINT DF_CPNT_UPDATED DEFAULT GETDATE(),
  CONSTRAINT PK_CPNOTIFTEMPLATE PRIMARY KEY CLUSTERED (ID),
  CONSTRAINT UQ_CPNT_CODE UNIQUE NONCLUSTERED (CODE)
);

-- Seed 3 template bawaan
IF NOT EXISTS (SELECT 1 FROM dbo.CPNOTIFTEMPLATE WHERE CODE = N'NTPL-WELCOME')
  INSERT INTO dbo.CPNOTIFTEMPLATE (CODE, NAME, CHANNEL, SUBJECT, BODY)
  VALUES (N'NTPL-WELCOME', N'Selamat datang', N'EMAIL', N'Selamat datang di Go Core, {{nama}}!',
    N'Halo {{nama}},\n\nAkun Anda ({{kode}}) berhasil dibuat dengan role {{role}}.\nSilakan login untuk mulai menggunakan aplikasi.\n\nSalam,\nTim Go Core');
IF NOT EXISTS (SELECT 1 FROM dbo.CPNOTIFTEMPLATE WHERE CODE = N'NTPL-RESET')
  INSERT INTO dbo.CPNOTIFTEMPLATE (CODE, NAME, CHANNEL, SUBJECT, BODY)
  VALUES (N'NTPL-RESET', N'Reset password', N'EMAIL', N'Permintaan reset password akun {{kode}}',
    N'Halo {{nama}},\n\nKami menerima permintaan reset password untuk akun {{kode}}.\nAbaikan pesan ini bila Anda tidak memintanya.\n\nSalam,\nTim Go Core');
IF NOT EXISTS (SELECT 1 FROM dbo.CPNOTIFTEMPLATE WHERE CODE = N'NTPL-ALERT')
  INSERT INTO dbo.CPNOTIFTEMPLATE (CODE, NAME, CHANNEL, SUBJECT, BODY)
  VALUES (N'NTPL-ALERT', N'Peringatan keamanan', N'INAPP', N'Peringatan keamanan',
    N'Aktivitas penting terdeteksi pada akun {{kode}}: {{detail}}');

-- 6. CPNOTIFLOG -----------------------------------------------------------------------
IF OBJECT_ID(N'dbo.CPNOTIFLOG', N'U') IS NULL
CREATE TABLE dbo.CPNOTIFLOG (
  ID INT IDENTITY(1,1) NOT NULL,
  CODE NVARCHAR(20) NOT NULL,
  TEMPLATE_CODE NVARCHAR(20) NULL,       -- NULL = kirim manual tanpa template
  CHANNEL NVARCHAR(10) NOT NULL,
  RECIPIENT NVARCHAR(255) NOT NULL,
  SUBJECT NVARCHAR(200) NULL,
  BODY NVARCHAR(4000) NOT NULL,
  STATUS NVARCHAR(10) NOT NULL,          -- SENT, FAILED
  ERROR NVARCHAR(500) NULL,
  CREATED_AT DATETIME NOT NULL CONSTRAINT DF_CPNL_CREATED DEFAULT GETDATE(),
  CONSTRAINT PK_CPNOTIFLOG PRIMARY KEY CLUSTERED (ID),
  CONSTRAINT UQ_CPNL_CODE UNIQUE NONCLUSTERED (CODE)
);

COMMIT TRAN;
