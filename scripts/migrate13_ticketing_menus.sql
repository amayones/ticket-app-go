-- Migration 13: TICKETING module + 5 menus (CPMODULE/CPMENU rows only).
--
--   Same precedent as migrate5/7/10: menus are registered with ZERO grants
--   to any role (not even ADMIN). Access is assigned by the admin via the
--   Role & Permission matrix UI, where these rows now appear.
--   (Ticket Detail is a modal inside My Tickets, not a menu.)
--
--   Menus (mcontrol = frontend/src/app/<mcontrol>/ folder):
--     MENU_CHECKOUT      checkout       Checkout
--     MENU_MY_TICKETS    my_tickets     My Tickets
--     MENU_REFUNDS       refunds        Refunds
--     MENU_SCAN_TICKET   scan_ticket    Scan Ticket
--     MENU_SCAN_HISTORY  scan_history   Scan History
--
-- Idempotent (safe to re-run). Rollback: scripts/migrate13_ticketing_menus_rollback.sql
-- Run:
--   sqlcmd -S <host>,<port> -U <user> -P <pass> -d <DB_NAME> -C -b -i scripts/migrate13_ticketing_menus.sql
-- Or: task migrate
SET ANSI_NULLS ON;
SET QUOTED_IDENTIFIER ON;
SET XACT_ABORT ON;
BEGIN TRAN;

INSERT INTO dbo.CPMODULE (CODE, LABEL, SORT_ORDER)
SELECT N'TICKETING', N'Ticketing', 5
WHERE NOT EXISTS (SELECT 1 FROM dbo.CPMODULE WHERE CODE = N'TICKETING');

IF COL_LENGTH(N'dbo.CPMENU', N'MCONTROL') IS NULL
  EXEC(N'INSERT INTO dbo.CPMENU (CODE, MODULE, LABEL, MENU_KIND, SORT_ORDER, PARENT_CODE)
  SELECT v.CODE, v.MODULE, v.LABEL, v.KIND, v.SORT, v.PARENT FROM (VALUES
    (N''MENU_CHECKOUT'', N''TICKETING'', N''Checkout'', N''CHILD'', 1, NULL),
    (N''MENU_MY_TICKETS'', N''TICKETING'', N''My Tickets'', N''CHILD'', 2, NULL),
    (N''MENU_REFUNDS'', N''TICKETING'', N''Refunds'', N''CHILD'', 3, NULL),
    (N''MENU_SCAN_TICKET'', N''TICKETING'', N''Scan Ticket'', N''CHILD'', 4, NULL),
    (N''MENU_SCAN_HISTORY'', N''TICKETING'', N''Scan History'', N''CHILD'', 5, NULL)
  ) v(CODE, MODULE, LABEL, KIND, SORT, PARENT)
  WHERE NOT EXISTS (SELECT 1 FROM dbo.CPMENU x WHERE x.CODE = v.CODE);');
ELSE
  EXEC(N'INSERT INTO dbo.CPMENU (CODE, MODULE, LABEL, MCONTROL, MENU_KIND, SORT_ORDER, PARENT_CODE)
  SELECT v.CODE, v.MODULE, v.LABEL, v.MCONTROL, v.KIND, v.SORT, v.PARENT FROM (VALUES
    (N''MENU_CHECKOUT'', N''TICKETING'', N''Checkout'', N''checkout'', N''CHILD'', 1, NULL),
    (N''MENU_MY_TICKETS'', N''TICKETING'', N''My Tickets'', N''my_tickets'', N''CHILD'', 2, NULL),
    (N''MENU_REFUNDS'', N''TICKETING'', N''Refunds'', N''refunds'', N''CHILD'', 3, NULL),
    (N''MENU_SCAN_TICKET'', N''TICKETING'', N''Scan Ticket'', N''scan_ticket'', N''CHILD'', 4, NULL),
    (N''MENU_SCAN_HISTORY'', N''TICKETING'', N''Scan History'', N''scan_history'', N''CHILD'', 5, NULL)
  ) v(CODE, MODULE, LABEL, MCONTROL, KIND, SORT, PARENT)
  WHERE NOT EXISTS (SELECT 1 FROM dbo.CPMENU x WHERE x.CODE = v.CODE);');

-- Guard: keep zero grants on ticketing menus (re-clears if the legacy
-- "ADMIN = all CHILD" seed ever runs again after this file).
DELETE FROM dbo.CPPERMISSION
WHERE MENU_CODE IN (N'MENU_CHECKOUT', N'MENU_MY_TICKETS', N'MENU_REFUNDS',
                    N'MENU_SCAN_TICKET', N'MENU_SCAN_HISTORY');

COMMIT TRAN;
