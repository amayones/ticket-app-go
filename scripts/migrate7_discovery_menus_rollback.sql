-- Rollback for scripts/migrate7_discovery_menus.sql (DISCOVERY menus).
-- Removes grants first, then menu rows, then the module.
-- Idempotent (safe to re-run). Run:
--   sqlcmd -S <host>,<port> -U <user> -P <pass> -d <DB_NAME> -C -b -i scripts/migrate7_discovery_menus_rollback.sql
SET XACT_ABORT ON;
BEGIN TRAN;

DELETE FROM dbo.CPPERMISSION
WHERE MENU_CODE IN (N'MENU_HOME', N'MENU_NEWS', N'MENU_EVENTS',
                    N'MENU_EVENT_DETAIL', N'MENU_PROMOTIONS');
DELETE FROM dbo.CPMENU
WHERE CODE IN (N'MENU_HOME', N'MENU_NEWS', N'MENU_EVENTS',
               N'MENU_EVENT_DETAIL', N'MENU_PROMOTIONS');
DELETE FROM dbo.CPMODULE WHERE CODE = N'DISCOVERY';

COMMIT TRAN;
