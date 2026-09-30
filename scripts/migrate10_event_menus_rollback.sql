-- Rollback for scripts/migrate10_event_menus.sql (EVENT menus).
-- Removes grants first, then menu rows, then the module.
-- Idempotent (safe to re-run). Run:
--   sqlcmd -S <host>,<port> -U <user> -P <pass> -d <DB_NAME> -C -b -i scripts/migrate10_event_menus_rollback.sql
SET XACT_ABORT ON;
BEGIN TRAN;

DELETE FROM dbo.CPPERMISSION
WHERE MENU_CODE IN (N'MENU_MY_EVENTS', N'MENU_CREATE_EVENT',
                    N'MENU_TICKET_TYPES', N'MENU_ATTENDEES');
DELETE FROM dbo.CPMENU
WHERE CODE IN (N'MENU_MY_EVENTS', N'MENU_CREATE_EVENT',
               N'MENU_TICKET_TYPES', N'MENU_ATTENDEES');
DELETE FROM dbo.CPMODULE WHERE CODE = N'EVENT';

COMMIT TRAN;
