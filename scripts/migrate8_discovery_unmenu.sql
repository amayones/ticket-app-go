-- Migration 8: unlist MENU_EVENT_DETAIL (event detail becomes a modal
-- inside the Events menu instead of its own menu).
--
--   Deletes MENU_EVENT_DETAIL grants (all roles) + the menu row.
--   Module DISCOVERY keeps 4 menus (HOME, NEWS, EVENTS, PROMOTIONS).
--   The GET /api/discovery/events/:code backend stays as-is (used by modal).
--
-- Idempotent (safe to re-run). Rollback: scripts/migrate8_discovery_unmenu_rollback.sql
-- Run:
--   sqlcmd -S <host>,<port> -U <user> -P <pass> -d <DB_NAME> -C -b -i scripts/migrate8_discovery_unmenu.sql
-- Or: task migrate
SET ANSI_NULLS ON;
SET QUOTED_IDENTIFIER ON;
SET XACT_ABORT ON;
BEGIN TRAN;

DELETE FROM dbo.CPPERMISSION WHERE MENU_CODE = N'MENU_EVENT_DETAIL';
DELETE FROM dbo.CPMENU WHERE CODE = N'MENU_EVENT_DETAIL';

COMMIT TRAN;
