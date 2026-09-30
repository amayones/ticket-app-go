-- Migration 9: EVENT module columns (no new tables - all reused).
--
--   Verified against live ticket-db: T_EVENT, T_TICKET_TYPE, T_ORDER and
--   T_TICKET already exist (migrate5/6) and are reused as-is. Only columns
--   required by the EVENT module are added here; CP_* untouched.
--     T_EVENT: SLUG (unique, backfilled), TERMS, CANCEL_REASON,
--              IS_DELETED (soft delete), PUBLISHED_AT (first publish time).
--     T_TICKET_TYPE: DESCRIPTION (benefit), MAX_PER_PERSON, IS_ACTIVE,
--                    SORT_ORDER (display order).
--     T_TICKET: EMAIL, PHONE (attendee contact; TICKETING owns check-in).
--   T_ORDER doubles as payment status (PENDING/PAID/CANCELLED/REFUNDED).
--
-- Idempotent (safe to re-run). Rollback: scripts/migrate9_event_rollback.sql
-- Run:
--   sqlcmd -S <host>,<port> -U <user> -P <pass> -d <DB_NAME> -C -b -i scripts/migrate9_event.sql
-- Or: task migrate
SET ANSI_NULLS ON;
SET QUOTED_IDENTIFIER ON;
SET XACT_ABORT ON;
GO

-- NOTE: file ini memakai batch GO (sesi sqlcmd yang sama, satu transaksi
-- logis tetap dijaga XACT_ABORT): SQL Server meng-compile satu batch sebelum
-- menjalankannya, jadi kolom baru (ADD COLUMN) tidak boleh dirujuk di batch
-- yang sama dengan ALTER-nya (pola yang sama dipakai migrate2_rbac.sql).
BEGIN TRAN;

-- 1. T_EVENT ----------------------------------------------------------------------
IF COL_LENGTH(N'dbo.T_EVENT', N'SLUG') IS NULL
  ALTER TABLE dbo.T_EVENT ADD SLUG NVARCHAR(200) NULL;
IF COL_LENGTH(N'dbo.T_EVENT', N'TERMS') IS NULL
  ALTER TABLE dbo.T_EVENT ADD TERMS NVARCHAR(MAX) NULL;
IF COL_LENGTH(N'dbo.T_EVENT', N'CANCEL_REASON') IS NULL
  ALTER TABLE dbo.T_EVENT ADD CANCEL_REASON NVARCHAR(500) NULL;
IF COL_LENGTH(N'dbo.T_EVENT', N'IS_DELETED') IS NULL
  ALTER TABLE dbo.T_EVENT ADD IS_DELETED BIT NOT NULL
    CONSTRAINT DF_T_EVENT_DELETED DEFAULT 0;
IF COL_LENGTH(N'dbo.T_EVENT', N'PUBLISHED_AT') IS NULL
  ALTER TABLE dbo.T_EVENT ADD PUBLISHED_AT DATETIME NULL;
GO
-- Backfill slugs for existing rows: slugified title + CODE (unique by design).
UPDATE dbo.T_EVENT
SET SLUG = LEFT(LOWER(REPLACE(REPLACE(REPLACE(REPLACE(REPLACE(
  TITLE, N' ', N'-'), N',', N''), N'.', N''), N'''', N''), N'/', N'-')) + N'-' + LOWER(CODE), 200)
WHERE SLUG IS NULL;
GO
IF NOT EXISTS (SELECT 1 FROM sys.indexes WHERE name = N'UQ_T_EVENT_SLUG' AND object_id = OBJECT_ID(N'dbo.T_EVENT'))
  CREATE UNIQUE NONCLUSTERED INDEX UQ_T_EVENT_SLUG ON dbo.T_EVENT (SLUG);
GO

-- 2. T_TICKET_TYPE -------------------------------------------------------------------
IF COL_LENGTH(N'dbo.T_TICKET_TYPE', N'DESCRIPTION') IS NULL
  ALTER TABLE dbo.T_TICKET_TYPE ADD DESCRIPTION NVARCHAR(500) NULL;
IF COL_LENGTH(N'dbo.T_TICKET_TYPE', N'MAX_PER_PERSON') IS NULL
  ALTER TABLE dbo.T_TICKET_TYPE ADD MAX_PER_PERSON INT NULL;
IF COL_LENGTH(N'dbo.T_TICKET_TYPE', N'IS_ACTIVE') IS NULL
  ALTER TABLE dbo.T_TICKET_TYPE ADD IS_ACTIVE BIT NOT NULL
    CONSTRAINT DF_T_TT_ACTIVE DEFAULT 1;
IF COL_LENGTH(N'dbo.T_TICKET_TYPE', N'SORT_ORDER') IS NULL
  ALTER TABLE dbo.T_TICKET_TYPE ADD SORT_ORDER INT NOT NULL
    CONSTRAINT DF_T_TT_SORT DEFAULT 99;

-- 3. T_TICKET (attendee contact; payment lives in T_ORDER, check-in in TICKETING) ------
IF COL_LENGTH(N'dbo.T_TICKET', N'EMAIL') IS NULL
  ALTER TABLE dbo.T_TICKET ADD EMAIL NVARCHAR(255) NULL;
IF COL_LENGTH(N'dbo.T_TICKET', N'PHONE') IS NULL
  ALTER TABLE dbo.T_TICKET ADD PHONE NVARCHAR(30) NULL;
GO

-- 4. Index kolom filter My Events ----------------------------------------------------------
IF NOT EXISTS (SELECT 1 FROM sys.indexes WHERE name = N'IX_T_EVENT_MINE' AND object_id = OBJECT_ID(N'dbo.T_EVENT'))
  CREATE NONCLUSTERED INDEX IX_T_EVENT_MINE ON dbo.T_EVENT (ORGANIZER_CODE, IS_DELETED, STATUS, START_AT);
IF NOT EXISTS (SELECT 1 FROM sys.indexes WHERE name = N'IX_T_TT_EVENT_ACTIVE' AND object_id = OBJECT_ID(N'dbo.T_TICKET_TYPE'))
  CREATE NONCLUSTERED INDEX IX_T_TT_EVENT_ACTIVE ON dbo.T_TICKET_TYPE (EVENT_CODE, IS_ACTIVE, SORT_ORDER);

COMMIT TRAN;
