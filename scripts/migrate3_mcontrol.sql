-- Migrasi 3: MCONTROL = nama folder frontend per menu CHILD.
--   CPMENU.MCONTROL   NVARCHAR(40) NULL, snake_case (mis. user_account)
--                     = nama folder frontend/src/app/<mcontrol>/ (datar,
--                     tanpa folder modul). NULL = menu PARENT (header
--                     buka-tutup di sidebar, tanpa halaman/folder).
--   Aturan: CHILD wajib punya MCONTROL (unik), PARENT wajib NULL.
--   Sidebar 3 level: MODULE (section) -> PARENT (header) -> CHILD (item).
--   PARENT ikut tampil otomatis bila minimal satu keturunannya ter-grant
--   (ancestor dinaikkan di MyMenusByRole, tanpa grant sendiri).
--
-- Aman diulang (idempotent). Jalankan:
--   sqlcmd -S localhost,1433 -U <user> -P <pass> -d <NAMA_DB> -C -i scripts/migrate3_mcontrol.sql
-- Atau: task migrate (menjalankan migrate.sql, migrate2_rbac.sql, lalu file ini)
--
-- WAJIB: filtered index/constraint di bawah (UQ_CPMENU_MCONTROL) menuntut
-- QUOTED_IDENTIFIER ON. sqlcmd default-nya OFF, jadi diset eksplisit di sini
-- (tanpa ini: "CREATE INDEX failed because the following SET options have
-- incorrect settings: 'QUOTED_IDENTIFIER'").
SET ANSI_NULLS ON;
SET QUOTED_IDENTIFIER ON;
GO

IF NOT EXISTS (SELECT 1 FROM sys.columns
               WHERE object_id = OBJECT_ID(N'dbo.CPMENU') AND name = N'MCONTROL')
  ALTER TABLE dbo.CPMENU ADD MCONTROL NVARCHAR(40) NULL;
GO

-- Backfill menu bawaan: snake_case dari LABEL (User Account -> user_account).
-- PARENT (MENU_KEUANGAN) tetap NULL — tanpa folder.
UPDATE dbo.CPMENU SET MCONTROL = N'users'            WHERE CODE = N'MENU_USERS'         AND MCONTROL IS NULL;
UPDATE dbo.CPMENU SET MCONTROL = N'modul_menu'       WHERE CODE = N'MENU_MODUL'         AND MCONTROL IS NULL;
UPDATE dbo.CPMENU SET MCONTROL = N'role_permission'  WHERE CODE = N'MENU_ROLES'         AND MCONTROL IS NULL;
UPDATE dbo.CPMENU SET MCONTROL = N'sesi_auth'        WHERE CODE = N'MENU_SESSIONS'      AND MCONTROL IS NULL;
UPDATE dbo.CPMENU SET MCONTROL = N'audit_log'        WHERE CODE = N'MENU_AUDIT'         AND MCONTROL IS NULL;
UPDATE dbo.CPMENU SET MCONTROL = N'security_center'  WHERE CODE = N'MENU_SECURITY'      AND MCONTROL IS NULL;
UPDATE dbo.CPMENU SET MCONTROL = N'system_log'       WHERE CODE = N'MENU_SYSLOG'        AND MCONTROL IS NULL;
UPDATE dbo.CPMENU SET MCONTROL = N'notifikasi'       WHERE CODE = N'MENU_NOTIFICATIONS' AND MCONTROL IS NULL;
UPDATE dbo.CPMENU SET MCONTROL = N'laporan'          WHERE CODE = N'MENU_LAPORAN'        AND MCONTROL IS NULL;
UPDATE dbo.CPMENU SET MCONTROL = N'arus_kas'         WHERE CODE = N'MENU_ARUS_KAS'       AND MCONTROL IS NULL;
-- Menu CHILD buatan sendiri yang belum punya MCONTROL: turunkan dari CODE
-- (MENU_STOK -> stok, MENU_ARUS_KAS -> arus_kas).
UPDATE dbo.CPMENU
SET MCONTROL = LOWER(REPLACE(SUBSTRING(CODE, 6, 35), N'_', N'_'))
WHERE MENU_KIND = N'CHILD' AND MCONTROL IS NULL AND CODE LIKE N'MENU[_]%';
GO
-- Normalisasi: huruf kecil semua (folder case-sensitive di Linux).
UPDATE dbo.CPMENU SET MCONTROL = LOWER(MCONTROL) WHERE MCONTROL <> LOWER(MCONTROL);
GO
-- Baris PARENT tidak boleh punya folder (menu dengan anak yang baru dipromosikan
-- jadi PARENT oleh migrate2_rbac.sql bisa masih membawa MCONTROL lama).
UPDATE dbo.CPMENU SET MCONTROL = NULL WHERE MENU_KIND = N'PARENT' AND MCONTROL IS NOT NULL;
GO

-- Unik per folder (NULL dikecualikan otomatis di SQL Server: boleh banyak NULL).
IF NOT EXISTS (SELECT 1 FROM sys.indexes WHERE name = N'UQ_CPMENU_MCONTROL' AND object_id = OBJECT_ID(N'dbo.CPMENU'))
  CREATE UNIQUE NONCLUSTERED INDEX UQ_CPMENU_MCONTROL ON dbo.CPMENU (MCONTROL) WHERE MCONTROL IS NOT NULL;
GO

-- Aturan folder: PARENT tanpa mcontrol, CHILD wajib punya mcontrol.
IF NOT EXISTS (SELECT 1 FROM sys.check_constraints WHERE name = N'CK_CPMENU_MCONTROL')
  ALTER TABLE dbo.CPMENU ADD CONSTRAINT CK_CPMENU_MCONTROL
    CHECK ((MENU_KIND = N'PARENT' AND MCONTROL IS NULL) OR (MENU_KIND = N'CHILD' AND MCONTROL IS NOT NULL));
GO
