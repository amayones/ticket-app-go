-- Migrasi 2: RBAC + Audit Log + System Log + Notifikasi + sesi admin.
--   CPMODULE          master modul (CODE, LABEL, SORT_ORDER)
--   CPMENU            registry menu (MODULE -> CPMODULE = section sidebar,
--                     MCONTROL = folder frontend app/<mcontrol>/ untuk CHILD,
--                     MENU_KIND + PARENT_CODE = header buka-tutup)
--   CPPERMISSION      grant role -> menu (ROLE_CODE, MENU_CODE)
--   CPAUDITLOG        jejak aksi user (siapa, apa, kapan, dari IP mana)
--   CPSYSLOG          log error/sistem aplikasi (pengganti baca file log)
--   CPNOTIFTEMPLATE   template notifikasi (email/push/in-app, dukung {{var}})
--   CPNOTIFLOG        riwayat pengiriman notifikasi
-- Sesi memakai CPREFRESHTOKEN yang sudah ada (read/revoke via API).
--
-- Aman diulang (idempotent). Jalankan:
--   sqlcmd -S localhost,1433 -U <user> -P <pass> -d <NAMA_DB> -C -i scripts/migrate2_rbac.sql
-- Atau: task migrate (menjalankan migrate.sql, file ini, lalu migrate3_mcontrol.sql)
--
-- WAJIB: DML ke dbo.CPMENU (backfill MENU_KIND + seed menu) di bawah menuntut
-- QUOTED_IDENTIFIER ON karena migrasi 3 menambah filtered unique index
-- UQ_CPMENU_MCONTROL. Default sqlcmd = OFF, jadi diset eksplisit di sini
-- (tanpa ini: "UPDATE failed because the following SET options have incorrect
-- settings: 'QUOTED_IDENTIFIER'").
SET ANSI_NULLS ON;
SET QUOTED_IDENTIFIER ON;
SET XACT_ABORT ON;
BEGIN TRAN;

-- 1. (Dihapus) Tabel definisi CPPERMISSION lama (CODE/NAME/PERMGROUP) tidak
-- lagi dibuat; definisi menu tinggal di CPMENU. Database lama dibersihkan
-- di blok 2d (data grant ikut pindah, tabel lama di-drop).

-- 2. (Dihapus) Tabel grant CPERMISSION lama diganti CPPERMISSION (blok 2d).

-- 2b. CPMODULE (master modul) --------------------------------------------------
-- Satu baris = satu modul = satu section sidebar di frontend (MODULE di
-- CPMENU ber-FK ke sini sehingga modul wajib dibuat dulu sebelum menunya;
-- urut: modul -> menu -> grant role).
--
-- Riwayat nama: tabel ini dulu CPMATRIX, sekarang CPMODULE (kolom
-- CPMENU.MCONTROL waktu itu ikut jadi CPMENU.MODULE). Blok di bawah me-rename
-- objek lama bila masih ada, lalu membuat tabel bila belum ada sama sekali.
-- Catatan: sejak migrasi 3, MCONTROL dipakai lagi untuk arti lain (nama folder
-- frontend per menu CHILD), lihat migrate3_mcontrol.sql.
IF OBJECT_ID(N'dbo.CPMATRIX', N'V') IS NOT NULL DROP VIEW dbo.CPMATRIX;
GO
-- Rename dulu (instalasi lama), baru buat kalau memang belum ada. Urutan ini
-- penting: kalau tabel baru dibuat lebih dulu, sp_rename berikutnya bentrok.
IF OBJECT_ID(N'dbo.CPMODULE', N'U') IS NULL AND OBJECT_ID(N'dbo.CPMATRIX', N'U') IS NOT NULL
BEGIN
  EXEC sp_rename N'dbo.CPMATRIX', N'CPMODULE';
  IF EXISTS (SELECT 1 FROM sys.key_constraints WHERE name = N'PK_CPMATRIX')
    EXEC sp_rename N'dbo.CPMODULE.PK_CPMATRIX', N'PK_CPMODULE', N'INDEX';
  IF EXISTS (SELECT 1 FROM sys.indexes WHERE name = N'UQ_CPMATRIX_CODE')
    EXEC sp_rename N'dbo.CPMODULE.UQ_CPMATRIX_CODE', N'UQ_CPMODULE_CODE', N'INDEX';
END
GO
IF OBJECT_ID(N'dbo.CPMODULE', N'U') IS NULL
CREATE TABLE dbo.CPMODULE (
  ID INT IDENTITY(1,1) NOT NULL,
  CODE NVARCHAR(40) NOT NULL,
  LABEL NVARCHAR(100) NOT NULL,
  SORT_ORDER INT NOT NULL CONSTRAINT DF_CPMODULE_SORT DEFAULT 99,
  CREATED_AT DATETIME NOT NULL CONSTRAINT DF_CPMODULE_CREATED DEFAULT GETDATE(),
  UPDATED_AT DATETIME NOT NULL CONSTRAINT DF_CPMODULE_UPDATED DEFAULT GETDATE(),
  CONSTRAINT PK_CPMODULE PRIMARY KEY CLUSTERED (ID),
  CONSTRAINT UQ_CPMODULE_CODE UNIQUE NONCLUSTERED (CODE)
);
GO

-- Seed modul bawaan (SYSTEM untuk operasional).
INSERT INTO dbo.CPMODULE (CODE, LABEL, SORT_ORDER)
SELECT N'SYSTEM', N'System', 1
WHERE NOT EXISTS (SELECT 1 FROM dbo.CPMODULE WHERE CODE = N'SYSTEM');
GO

-- 2b. CPMENU (registry menu) ---------------------------------------------------
-- Satu baris = satu menu. MODULE = section sidebar (FK ke CPMODULE; label
-- section dari CPMODULE.LABEL), bukan nama folder.
-- MENU_KIND menentukan peran baris: 'PARENT' (header buka-tutup di sidebar,
-- tanpa halaman/folder) atau 'CHILD' (item biasa, folder frontend dari
-- MCONTROL). PARENT_CODE = kode menu parent; wajib menunjuk menu bertipe
-- PARENT, satu modul yang sama, dan tidak boleh menimbulkan siklus.
IF OBJECT_ID(N'dbo.CPMENU', N'U') IS NULL
CREATE TABLE dbo.CPMENU (
  ID INT IDENTITY(1,1) NOT NULL,
  CODE NVARCHAR(40) NOT NULL,
  MODULE NVARCHAR(40) NOT NULL,
  LABEL NVARCHAR(100) NOT NULL,
  MENU_KIND NVARCHAR(10) NOT NULL CONSTRAINT DF_CPMENU_KIND DEFAULT N'CHILD',
  SORT_ORDER INT NOT NULL CONSTRAINT DF_CPMENU_SORT DEFAULT 99,
  PARENT_CODE NVARCHAR(40) NULL,
  CREATED_AT DATETIME NOT NULL CONSTRAINT DF_CPMENU_CREATED DEFAULT GETDATE(),
  UPDATED_AT DATETIME NOT NULL CONSTRAINT DF_CPMENU_UPDATED DEFAULT GETDATE(),
  CONSTRAINT PK_CPMENU PRIMARY KEY CLUSTERED (ID),
  CONSTRAINT UQ_CPMENU_CODE UNIQUE NONCLUSTERED (CODE),
  CONSTRAINT UQ_CPMENU_MODULE_CODE UNIQUE NONCLUSTERED (MODULE, CODE),
  CONSTRAINT FK_CPMENU_MODULE FOREIGN KEY (MODULE) REFERENCES dbo.CPMODULE (CODE),
  CONSTRAINT FK_CPMENU_PARENT FOREIGN KEY (PARENT_CODE) REFERENCES dbo.CPMENU (CODE),
  CONSTRAINT CK_CPMENU_KIND CHECK (MENU_KIND IN (N'PARENT', N'CHILD'))
);
GO
-- Rename kolom lama MCONTROL -> MODULE (instalasi sebelum rename). Hanya bila
-- MODULE belum ada: setelah migrate3_mcontrol.sql, MCONTROL dipakai lagi untuk
-- nama folder frontend sehingga rename tidak boleh diulang.
IF EXISTS (SELECT 1 FROM sys.columns
           WHERE object_id = OBJECT_ID(N'dbo.CPMENU') AND name = N'MCONTROL')
   AND NOT EXISTS (SELECT 1 FROM sys.columns
           WHERE object_id = OBJECT_ID(N'dbo.CPMENU') AND name = N'MODULE')
  EXEC sp_rename N'dbo.CPMENU.MCONTROL', N'MODULE', N'COLUMN';
GO
-- Tambah kolom MENU_KIND bila belum ada. ALTER dipisah dari UPDATE karena
-- SQL Server meng-compile satu batch sebelum menjalankannya, jadi kolom baru
-- tidak boleh dirujuk di batch yang sama dengan ADD COLUMN.
IF NOT EXISTS (SELECT 1 FROM sys.columns
               WHERE object_id = OBJECT_ID(N'dbo.CPMENU') AND name = N'MENU_KIND')
  ALTER TABLE dbo.CPMENU ADD MENU_KIND NVARCHAR(10) NOT NULL
    CONSTRAINT DF_CPMENU_KIND DEFAULT N'CHILD';
GO
-- Backfill: menu yang sudah punya anak = PARENT, selain itu CHILD.
UPDATE dbo.CPMENU
SET MENU_KIND = N'PARENT'
WHERE MENU_KIND = N'CHILD'
  AND EXISTS (SELECT 1 FROM dbo.CPMENU c WHERE c.PARENT_CODE = CPMENU.CODE);
GO
IF NOT EXISTS (SELECT 1 FROM sys.check_constraints WHERE name = N'CK_CPMENU_KIND')
  ALTER TABLE dbo.CPMENU ADD CONSTRAINT CK_CPMENU_KIND
    CHECK (MENU_KIND IN (N'PARENT', N'CHILD'));
GO
-- FK lama ke tabel definisi (sudah tidak ada) dicabut bila masih menempel.
IF EXISTS (SELECT 1 FROM sys.foreign_keys WHERE name = N'FK_CPMENU_PERM')
  ALTER TABLE dbo.CPMENU DROP CONSTRAINT FK_CPMENU_PERM;
-- Tabel lama (sebelum FK modul ada) dilengkapi secara kondisional.
IF NOT EXISTS (SELECT 1 FROM sys.foreign_keys WHERE name = N'FK_CPMENU_MODULE')
  ALTER TABLE dbo.CPMENU ADD CONSTRAINT FK_CPMENU_MODULE
    FOREIGN KEY (MODULE) REFERENCES dbo.CPMODULE (CODE);

-- Seed registry menu bawaan (urut sesuai sidebar).
-- MENU_MODUL = halaman manajemen modul & menu. MCONTROL = nama folder
-- frontend app/<mcontrol>/ (kolomnya ditambahkan migrate3_mcontrol.sql;
-- lihat percabangan di bawah).
IF OBJECT_ID(N'tempdb..#menus') IS NOT NULL DROP TABLE #menus;
CREATE TABLE #menus (CODE NVARCHAR(40), MODULE NVARCHAR(40), LABEL NVARCHAR(100), MCONTROL NVARCHAR(40) NULL, MENU_KIND NVARCHAR(10), SORT_ORDER INT, PARENT_CODE NVARCHAR(40) NULL);
INSERT INTO #menus VALUES
  (N'MENU_USERS',         N'SYSTEM', N'User Account',      N'users',           N'CHILD',  1, NULL),
  (N'MENU_MODUL',         N'SYSTEM', N'Modul & Menu',      N'modul_menu',      N'CHILD',  2, NULL),
  (N'MENU_ROLES',         N'SYSTEM', N'Role & Permission', N'role_permission', N'CHILD',  3, NULL),
  (N'MENU_SESSIONS',      N'SYSTEM', N'Sesi & Auth',       N'sesi_auth',       N'CHILD',  4, NULL),
  (N'MENU_AUDIT',         N'SYSTEM', N'Audit Log',         N'audit_log',       N'CHILD',  5, NULL),
  (N'MENU_SECURITY',      N'SYSTEM', N'Security Center',   N'security_center', N'CHILD',  6, NULL),
  (N'MENU_SYSLOG',        N'SYSTEM', N'System Log',        N'system_log',      N'CHILD',  7, NULL),
  (N'MENU_NOTIFICATIONS', N'SYSTEM', N'Notifikasi',        N'notifikasi',      N'CHILD',  8, NULL);

-- Dua varian INSERT: tanpa MCONTROL (kolom belum ada) atau dengan MCONTROL
-- (kolom sudah ada, CK_CPMENU_MCONTROL menuntut CHILD punya folder). Diambil
-- lewat EXEC karena nama kolom di-hardcode per varian.
IF COL_LENGTH(N'dbo.CPMENU', N'MCONTROL') IS NULL
  EXEC(N'INSERT INTO dbo.CPMENU (CODE, MODULE, LABEL, MENU_KIND, SORT_ORDER, PARENT_CODE)
  SELECT CODE, MODULE, LABEL, MENU_KIND, SORT_ORDER, PARENT_CODE FROM #menus m
  WHERE NOT EXISTS (SELECT 1 FROM dbo.CPMENU x WHERE x.CODE = m.CODE);');
ELSE
  EXEC(N'INSERT INTO dbo.CPMENU (CODE, MODULE, LABEL, MCONTROL, MENU_KIND, SORT_ORDER, PARENT_CODE)
  SELECT CODE, MODULE, LABEL, MCONTROL, MENU_KIND, SORT_ORDER, PARENT_CODE FROM #menus m
  WHERE NOT EXISTS (SELECT 1 FROM dbo.CPMENU x WHERE x.CODE = m.CODE);');

-- Sinkronkan urutan + label menu bawaan (penting untuk instalasi lama yang
-- sudah menjalankan migrasi ini sebelum urutan_sidebar berubah). Hanya 8 kode
-- bawaan di SYSTEM yang disentuh; menu buatan sendiri tidak terpengaruh.
UPDATE m
SET m.SORT_ORDER = s.SORT_ORDER, m.LABEL = s.LABEL
FROM dbo.CPMENU m
JOIN #menus s ON s.CODE = m.CODE
WHERE m.MODULE = N'SYSTEM'
  AND (m.SORT_ORDER <> s.SORT_ORDER OR m.LABEL <> s.LABEL);
DROP TABLE #menus;
GO

-- 2d. CPPERMISSION (grant role -> menu) --------------------------------------------
-- Satu-satunya tabel relasi: ROLE_CODE -> MENU_CODE (role boleh tampil menu
-- apa). Definisi menu tinggal di CPMENU; tidak ada tabel definisi terpisah.
-- Migrasi dari format lama dipecah batch GO kecil (tiap langkah idempoten;
-- dengan -b, batch yang gagal berhenti dengan pesan jelas).
-- Langkah 1: tabel penampung.
IF OBJECT_ID(N'dbo.CPPERMISSION', N'U') IS NULL
   OR COL_LENGTH(N'dbo.CPPERMISSION', N'MENU_CODE') IS NULL
BEGIN
  IF OBJECT_ID(N'dbo.CPPERMISSION_NEW', N'U') IS NULL
  CREATE TABLE dbo.CPPERMISSION_NEW (
    ROLE_CODE NVARCHAR(20) NOT NULL,
    MENU_CODE NVARCHAR(40) NOT NULL,
    CREATED_AT DATETIME NOT NULL CONSTRAINT DF_CPP_NEW_CREATED DEFAULT GETDATE(),
    CONSTRAINT PK_CPP_NEW PRIMARY KEY CLUSTERED (ROLE_CODE, MENU_CODE),
    CONSTRAINT FK_CPP_NEW_ROLE FOREIGN KEY (ROLE_CODE) REFERENCES dbo.CPROLE (CODE) ON DELETE CASCADE,
    CONSTRAINT FK_CPP_NEW_MENU FOREIGN KEY (MENU_CODE) REFERENCES dbo.CPMENU (CODE) ON DELETE CASCADE
  );
END
GO
-- Langkah 2: salin grant lama (hanya yang menunjuk ke menu yang ada).
-- CATATAN: query ke tabel legacy dibungkus EXEC agar SQL Server hanya
-- mengikat namanya saat benar-benar dijalankan (guard IF di atas).
IF OBJECT_ID(N'dbo.CPPERMISSION_NEW', N'U') IS NOT NULL
  AND OBJECT_ID(N'dbo.CPERMISSION', N'U') IS NOT NULL
  EXEC(N'INSERT INTO dbo.CPPERMISSION_NEW (ROLE_CODE, MENU_CODE, CREATED_AT)
  SELECT g.ROLE_CODE, g.PERMISSION_CODE, g.CREATED_AT FROM dbo.CPERMISSION g
  JOIN dbo.CPMENU m ON m.CODE = g.PERMISSION_CODE
  WHERE NOT EXISTS (
    SELECT 1 FROM dbo.CPPERMISSION_NEW x
    WHERE x.ROLE_CODE = g.ROLE_CODE AND x.MENU_CODE = g.PERMISSION_CODE
  )');
GO
IF OBJECT_ID(N'dbo.CPPERMISSION_NEW', N'U') IS NOT NULL
  AND OBJECT_ID(N'dbo.CPROLEPERMISSION', N'U') IS NOT NULL
  EXEC(N'INSERT INTO dbo.CPPERMISSION_NEW (ROLE_CODE, MENU_CODE)
  SELECT g.ROLE_CODE, g.PERMISSION_CODE FROM dbo.CPROLEPERMISSION g
  JOIN dbo.CPMENU m ON m.CODE = g.PERMISSION_CODE
  WHERE NOT EXISTS (
    SELECT 1 FROM dbo.CPPERMISSION_NEW x
    WHERE x.ROLE_CODE = g.ROLE_CODE AND x.MENU_CODE = g.PERMISSION_CODE
  )');
GO
-- Langkah 3: buang tabel format lama (hanya bila isi sudah pindah).
IF OBJECT_ID(N'dbo.CPPERMISSION_NEW', N'U') IS NOT NULL
BEGIN
  IF OBJECT_ID(N'dbo.CPERMISSION', N'U') IS NOT NULL
    EXEC(N'IF NOT EXISTS (
      SELECT 1 FROM dbo.CPERMISSION g JOIN dbo.CPMENU m ON m.CODE = g.PERMISSION_CODE
      WHERE NOT EXISTS (
        SELECT 1 FROM dbo.CPPERMISSION_NEW x
        WHERE x.ROLE_CODE = g.ROLE_CODE AND x.MENU_CODE = g.PERMISSION_CODE
      )
    )
    DROP TABLE dbo.CPERMISSION;');
  IF OBJECT_ID(N'dbo.CPROLEPERMISSION', N'U') IS NOT NULL
    EXEC(N'IF NOT EXISTS (
      SELECT 1 FROM dbo.CPROLEPERMISSION g JOIN dbo.CPMENU m ON m.CODE = g.PERMISSION_CODE
      WHERE NOT EXISTS (
        SELECT 1 FROM dbo.CPPERMISSION_NEW x
        WHERE x.ROLE_CODE = g.ROLE_CODE AND x.MENU_CODE = g.PERMISSION_CODE
      )
    )
    DROP TABLE dbo.CPROLEPERMISSION;');
  IF EXISTS (SELECT 1 FROM sys.foreign_keys WHERE name = N'FK_CPMENU_PERM')
    ALTER TABLE dbo.CPMENU DROP CONSTRAINT FK_CPMENU_PERM;
  IF OBJECT_ID(N'dbo.CPPERMISSION', N'U') IS NOT NULL
    AND COL_LENGTH(N'dbo.CPPERMISSION', N'MENU_CODE') IS NULL
    DROP TABLE dbo.CPPERMISSION;
END
GO
-- Langkah 4: tukar nama (batch sendiri agar SQL Server mengikat nama baru).
IF OBJECT_ID(N'dbo.CPPERMISSION_NEW', N'U') IS NOT NULL
  AND OBJECT_ID(N'dbo.CPPERMISSION', N'U') IS NULL
BEGIN
  EXEC sp_rename N'dbo.CPPERMISSION_NEW', N'CPPERMISSION';
  EXEC sp_rename N'PK_CPP_NEW', N'PK_CPPERMISSION';
  EXEC sp_rename N'FK_CPP_NEW_ROLE', N'FK_CPP_ROLE';
  EXEC sp_rename N'FK_CPP_NEW_MENU', N'FK_CPP_MENU';
  EXEC sp_rename N'DF_CPP_NEW_CREATED', N'DF_CPP_CREATED';
END
GO
-- Langkah 5: bersih + seed (jalan di format final).
-- Bersihkan kode non-menu sisa era permission per-fitur.
DELETE FROM dbo.CPPERMISSION
WHERE MENU_CODE IN (
  N'USER_READ', N'USER_CREATE', N'USER_UPDATE', N'USER_DELETE', N'USER_ROLE_ASSIGN',
  N'ROLE_READ', N'ROLE_MANAGE', N'PERMISSION_ASSIGN',
  N'SESSION_READ', N'SESSION_REVOKE', N'SESSION_MANAGE',
  N'AUDIT_READ', N'SECURITY_READ', N'SYSLOG_READ', N'SYSLOG_MANAGE',
  N'NOTIF_READ', N'NOTIF_MANAGE', N'NOTIF_SEND'
);
-- ADMIN mendapat semua menu CHILD secara default (dari CPMENU).
-- Menu PARENT tidak di-grant: header buka-tutup tanpa halaman sendiri, aksesnya
-- menyusul otomatis dari menu CHILD di bawahnya (lihat MyMenusByRole).
INSERT INTO dbo.CPPERMISSION (ROLE_CODE, MENU_CODE)
SELECT N'ADMIN', CODE FROM dbo.CPMENU
WHERE MENU_KIND = N'CHILD'
AND NOT EXISTS (
  SELECT 1 FROM dbo.CPPERMISSION x WHERE x.ROLE_CODE = N'ADMIN' AND x.MENU_CODE = CPMENU.CODE
);
-- Bersihkan sisa modul REPORT yang sudah dihapus (instalasi lama yang pernah
-- menjalankan seed REPORT): urut hapus grant -> child -> parent -> modul.
DELETE FROM dbo.CPPERMISSION WHERE MENU_CODE IN (N'MENU_LAPORAN', N'MENU_ARUS_KAS', N'MENU_KEUANGAN');
DELETE FROM dbo.CPMENU WHERE CODE = N'MENU_ARUS_KAS';
DELETE FROM dbo.CPMENU WHERE CODE = N'MENU_LAPORAN';
DELETE FROM dbo.CPMENU WHERE CODE = N'MENU_KEUANGAN';
DELETE FROM dbo.CPMODULE WHERE CODE = N'REPORT';
-- Bersihkan grant PARENT sisa instalasi lama: matriks Role & Permission kini
-- hanya mencentang menu CHILD (header PARENT tampil otomatis, tanpa grant).
DELETE g
FROM dbo.CPPERMISSION g
JOIN dbo.CPMENU m ON m.CODE = g.MENU_CODE
WHERE m.MENU_KIND = N'PARENT';
-- USER tidak diberi menu apa pun (nol mapping). Akses diberikan manual
-- oleh admin lewat matriks Role & Permission.
-- (Sengaja tidak ada INSERT untuk USER di sini.)
GO

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
