-- Rollback for scripts/migrate8_discovery_unmenu.sql.
-- Re-registers MENU_EVENT_DETAIL with ZERO grants (assign via matrix UI).
-- Idempotent (safe to re-run). Run:
--   sqlcmd -S <host>,<port> -U <user> -P <pass> -d <DB_NAME> -C -b -i scripts/migrate8_discovery_unmenu_rollback.sql
SET ANSI_NULLS ON;
SET QUOTED_IDENTIFIER ON;
SET XACT_ABORT ON;
BEGIN TRAN;

IF COL_LENGTH(N'dbo.CPMENU', N'MCONTROL') IS NULL
  EXEC(N'INSERT INTO dbo.CPMENU (CODE, MODULE, LABEL, MENU_KIND, SORT_ORDER, PARENT_CODE)
  SELECT N''MENU_EVENT_DETAIL'', N''DISCOVERY'', N''Event Detail'', N''CHILD'', 4, NULL
  WHERE NOT EXISTS (SELECT 1 FROM dbo.CPMENU x WHERE x.CODE = N''MENU_EVENT_DETAIL'');');
ELSE
  EXEC(N'INSERT INTO dbo.CPMENU (CODE, MODULE, LABEL, MCONTROL, MENU_KIND, SORT_ORDER, PARENT_CODE)
  SELECT N''MENU_EVENT_DETAIL'', N''DISCOVERY'', N''Event Detail'', N''event_detail'', N''CHILD'', 4, NULL
  WHERE NOT EXISTS (SELECT 1 FROM dbo.CPMENU x WHERE x.CODE = N''MENU_EVENT_DETAIL'');');

COMMIT TRAN;
