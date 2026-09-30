-- Rollback for scripts/migrate13_ticketing_menus.sql (TICKETING menus).
-- Removes grants first, then menu rows, then the module.
-- Idempotent (safe to re-run). Run:
--   sqlcmd -S <host>,<port> -U <user> -P <pass> -d <DB_NAME> -C -b -i scripts/migrate13_ticketing_menus_rollback.sql
SET XACT_ABORT ON;
BEGIN TRAN;

DELETE FROM dbo.CPPERMISSION
WHERE MENU_CODE IN (N'MENU_CHECKOUT', N'MENU_MY_TICKETS', N'MENU_REFUNDS',
                    N'MENU_SCAN_TICKET', N'MENU_SCAN_HISTORY');
DELETE FROM dbo.CPMENU
WHERE CODE IN (N'MENU_CHECKOUT', N'MENU_MY_TICKETS', N'MENU_REFUNDS',
               N'MENU_SCAN_TICKET', N'MENU_SCAN_HISTORY');
DELETE FROM dbo.CPMODULE WHERE CODE = N'TICKETING';

COMMIT TRAN;
