-- Rollback for scripts/migrate11_fold_create_event.sql.
-- Re-registers MENU_CREATE_EVENT with ZERO grants (assign via matrix UI).
-- Idempotent (safe to re-run). Run:
--   sqlcmd -S <host>,<port> -U <user> -P <pass> -d <DB_NAME> -C -b -i scripts/migrate11_fold_create_event_rollback.sql
SET ANSI_NULLS ON;
SET QUOTED_IDENTIFIER ON;
SET XACT_ABORT ON;
BEGIN TRAN;

IF COL_LENGTH(N'dbo.CPMENU', N'MCONTROL') IS NULL
  EXEC(N'INSERT INTO dbo.CPMENU (CODE, MODULE, LABEL, MENU_KIND, SORT_ORDER, PARENT_CODE)
  SELECT N''MENU_CREATE_EVENT'', N''EVENT'', N''Create Event'', N''CHILD'', 2, NULL
  WHERE NOT EXISTS (SELECT 1 FROM dbo.CPMENU x WHERE x.CODE = N''MENU_CREATE_EVENT'');');
ELSE
  EXEC(N'INSERT INTO dbo.CPMENU (CODE, MODULE, LABEL, MCONTROL, MENU_KIND, SORT_ORDER, PARENT_CODE)
  SELECT N''MENU_CREATE_EVENT'', N''EVENT'', N''Create Event'', N''create_event'', N''CHILD'', 2, NULL
  WHERE NOT EXISTS (SELECT 1 FROM dbo.CPMENU x WHERE x.CODE = N''MENU_CREATE_EVENT'');');

COMMIT TRAN;
