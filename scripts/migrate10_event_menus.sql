-- Migration 10: EVENT module + 4 menus (CPMODULE/CPMENU rows only).
--
--   Same precedent as migrate5 (DASHBOARD) and migrate7 (DISCOVERY):
--   menus are registered with ZERO grants to any role (not even ADMIN).
--   Access is assigned by the admin via the Role & Permission matrix UI,
--   where these rows now appear.
--
--   Menus (mcontrol = frontend/src/app/<mcontrol>/ folder):
--     MENU_MY_EVENTS     my_events      My Events
--     MENU_CREATE_EVENT  create_event   Create Event
--     MENU_TICKET_TYPES  ticket_types   Ticket Types
--     MENU_ATTENDEES     attendees      Attendees
--
-- Idempotent (safe to re-run). Rollback: scripts/migrate10_event_menus_rollback.sql
-- Run:
--   sqlcmd -S <host>,<port> -U <user> -P <pass> -d <DB_NAME> -C -b -i scripts/migrate10_event_menus.sql
-- Or: task migrate
SET ANSI_NULLS ON;
SET QUOTED_IDENTIFIER ON;
SET XACT_ABORT ON;
BEGIN TRAN;

INSERT INTO dbo.CPMODULE (CODE, LABEL, SORT_ORDER)
SELECT N'EVENT', N'Event', 4
WHERE NOT EXISTS (SELECT 1 FROM dbo.CPMODULE WHERE CODE = N'EVENT');

IF COL_LENGTH(N'dbo.CPMENU', N'MCONTROL') IS NULL
  EXEC(N'INSERT INTO dbo.CPMENU (CODE, MODULE, LABEL, MENU_KIND, SORT_ORDER, PARENT_CODE)
  SELECT v.CODE, v.MODULE, v.LABEL, v.KIND, v.SORT, v.PARENT FROM (VALUES
    (N''MENU_MY_EVENTS'', N''EVENT'', N''My Events'', N''CHILD'', 1, NULL),
    (N''MENU_CREATE_EVENT'', N''EVENT'', N''Create Event'', N''CHILD'', 2, NULL),
    (N''MENU_TICKET_TYPES'', N''EVENT'', N''Ticket Types'', N''CHILD'', 3, NULL),
    (N''MENU_ATTENDEES'', N''EVENT'', N''Attendees'', N''CHILD'', 4, NULL)
  ) v(CODE, MODULE, LABEL, KIND, SORT, PARENT)
  WHERE NOT EXISTS (SELECT 1 FROM dbo.CPMENU x WHERE x.CODE = v.CODE);');
ELSE
  EXEC(N'INSERT INTO dbo.CPMENU (CODE, MODULE, LABEL, MCONTROL, MENU_KIND, SORT_ORDER, PARENT_CODE)
  SELECT v.CODE, v.MODULE, v.LABEL, v.MCONTROL, v.KIND, v.SORT, v.PARENT FROM (VALUES
    (N''MENU_MY_EVENTS'', N''EVENT'', N''My Events'', N''my_events'', N''CHILD'', 1, NULL),
    (N''MENU_CREATE_EVENT'', N''EVENT'', N''Create Event'', N''create_event'', N''CHILD'', 2, NULL),
    (N''MENU_TICKET_TYPES'', N''EVENT'', N''Ticket Types'', N''ticket_types'', N''CHILD'', 3, NULL),
    (N''MENU_ATTENDEES'', N''EVENT'', N''Attendees'', N''attendees'', N''CHILD'', 4, NULL)
  ) v(CODE, MODULE, LABEL, MCONTROL, KIND, SORT, PARENT)
  WHERE NOT EXISTS (SELECT 1 FROM dbo.CPMENU x WHERE x.CODE = v.CODE);');

-- Guard: keep zero grants on event menus (re-clears if the legacy
-- "ADMIN = all CHILD" seed ever runs again after this file).
DELETE FROM dbo.CPPERMISSION
WHERE MENU_CODE IN (N'MENU_MY_EVENTS', N'MENU_CREATE_EVENT',
                    N'MENU_TICKET_TYPES', N'MENU_ATTENDEES');

COMMIT TRAN;
