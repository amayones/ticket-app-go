-- Migration 11: fold Create Event into My Events (FRM modal pattern).
--
--   Following the SYSTEM-module pattern (one menu page = StandardPage +
--   GRID + FRM modal), Create Event stops being its own menu and becomes
--   the FRM modal of My Events. Deletes MENU_CREATE_EVENT grants (all
--   roles) + the menu row. Module EVENT keeps 3 menus.
--
-- Idempotent (safe to re-run). Rollback: scripts/migrate11_fold_create_event_rollback.sql
-- Run:
--   sqlcmd -S <host>,<port> -U <user> -P <pass> -d <DB_NAME> -C -b -i scripts/migrate11_fold_create_event.sql
-- Or: task migrate
SET ANSI_NULLS ON;
SET QUOTED_IDENTIFIER ON;
SET XACT_ABORT ON;
BEGIN TRAN;

DELETE FROM dbo.CPPERMISSION WHERE MENU_CODE = N'MENU_CREATE_EVENT';
DELETE FROM dbo.CPMENU WHERE CODE = N'MENU_CREATE_EVENT';

COMMIT TRAN;
