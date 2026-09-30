-- Migration 7: DISCOVERY module + 5 menus (CPMODULE/CPMENU rows only).
--
--   Module DASHBOARD precedent (migrate5): menus are registered with ZERO
--   grants to any role (not even ADMIN). Access is assigned by the admin
--   via the Role & Permission matrix UI.
--
--   Menus (mcontrol = frontend/src/app/<mcontrol>/ folder):
--     MENU_HOME          home           Home
--     MENU_NEWS          news           News
--     MENU_EVENTS        events         Events
--     MENU_EVENT_DETAIL  event_detail   Event Detail
--     MENU_PROMOTIONS    promotions     Promotions
--
-- Idempotent (safe to re-run). Rollback: scripts/migrate7_discovery_menus_rollback.sql
-- Run:
--   sqlcmd -S <host>,<port> -U <user> -P <pass> -d <DB_NAME> -C -b -i scripts/migrate7_discovery_menus.sql
-- Or: task migrate
SET ANSI_NULLS ON;
SET QUOTED_IDENTIFIER ON;
SET XACT_ABORT ON;
BEGIN TRAN;

INSERT INTO dbo.CPMODULE (CODE, LABEL, SORT_ORDER)
SELECT N'DISCOVERY', N'Discovery', 3
WHERE NOT EXISTS (SELECT 1 FROM dbo.CPMODULE WHERE CODE = N'DISCOVERY');

IF COL_LENGTH(N'dbo.CPMENU', N'MCONTROL') IS NULL
  EXEC(N'INSERT INTO dbo.CPMENU (CODE, MODULE, LABEL, MENU_KIND, SORT_ORDER, PARENT_CODE)
  SELECT v.CODE, v.MODULE, v.LABEL, v.KIND, v.SORT, v.PARENT FROM (VALUES
    (N''MENU_HOME'', N''DISCOVERY'', N''Home'', N''CHILD'', 1, NULL),
    (N''MENU_NEWS'', N''DISCOVERY'', N''News'', N''CHILD'', 2, NULL),
    (N''MENU_EVENTS'', N''DISCOVERY'', N''Events'', N''CHILD'', 3, NULL),
    (N''MENU_EVENT_DETAIL'', N''DISCOVERY'', N''Event Detail'', N''CHILD'', 4, NULL),
    (N''MENU_PROMOTIONS'', N''DISCOVERY'', N''Promotions'', N''CHILD'', 5, NULL)
  ) v(CODE, MODULE, LABEL, KIND, SORT, PARENT)
  WHERE NOT EXISTS (SELECT 1 FROM dbo.CPMENU x WHERE x.CODE = v.CODE);');
ELSE
  EXEC(N'INSERT INTO dbo.CPMENU (CODE, MODULE, LABEL, MCONTROL, MENU_KIND, SORT_ORDER, PARENT_CODE)
  SELECT v.CODE, v.MODULE, v.LABEL, v.MCONTROL, v.KIND, v.SORT, v.PARENT FROM (VALUES
    (N''MENU_HOME'', N''DISCOVERY'', N''Home'', N''home'', N''CHILD'', 1, NULL),
    (N''MENU_NEWS'', N''DISCOVERY'', N''News'', N''news'', N''CHILD'', 2, NULL),
    (N''MENU_EVENTS'', N''DISCOVERY'', N''Events'', N''events'', N''CHILD'', 3, NULL),
    (N''MENU_EVENT_DETAIL'', N''DISCOVERY'', N''Event Detail'', N''event_detail'', N''CHILD'', 4, NULL),
    (N''MENU_PROMOTIONS'', N''DISCOVERY'', N''Promotions'', N''promotions'', N''CHILD'', 5, NULL)
  ) v(CODE, MODULE, LABEL, MCONTROL, KIND, SORT, PARENT)
  WHERE NOT EXISTS (SELECT 1 FROM dbo.CPMENU x WHERE x.CODE = v.CODE);');

-- Guard: keep zero grants on discovery menus (re-clears if the legacy
-- "ADMIN = all CHILD" seed ever runs again after this file).
DELETE FROM dbo.CPPERMISSION
WHERE MENU_CODE IN (N'MENU_HOME', N'MENU_NEWS', N'MENU_EVENTS',
                    N'MENU_EVENT_DETAIL', N'MENU_PROMOTIONS');

COMMIT TRAN;
