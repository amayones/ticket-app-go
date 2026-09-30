-- Migration 4: ticket-app roles + starter accounts (English only).
--
--   Roles (CPROLE):
--     ADMIN     = system operator (keeps the 8 SYSTEM menus).
--     AUDIENCE  = end user / ticket buyer.
--     ORGANIZER = event organizer.
--     SELLER    = merchandise seller.
--     OFFICER   = check-in officer.
--     USER      = legacy from old installs (left untouched).
--
--   New roles get ZERO menu grants on purpose: they see an empty page
--   ("no menus assigned - contact admin") until an admin assigns menus
--   via the Role & Permission matrix. ADMIN keeps SYSTEM only.
--
--   Starter accounts (password = username + "123"):
--     admin/admin123         (ADMIN)
--     audience/audience123   (AUDIENCE)
--     organizer/organizer123 (ORGANIZER)
--     seller/seller123       (SELLER)
--     officer/officer123     (OFFICER)
--
-- Idempotent (safe to re-run). Run:
--   sqlcmd -S <host>,<port> -U <user> -P <pass> -d <DB_NAME> -C -b -i scripts/migrate4_ticket_roles.sql
-- Or: task migrate
SET ANSI_NULLS ON;
SET QUOTED_IDENTIFIER ON;
SET XACT_ABORT ON;
BEGIN TRAN;

-- 1. Roles --------------------------------------------------------------------
IF NOT EXISTS (SELECT 1 FROM dbo.CPROLE WHERE CODE = N'ADMIN')
  INSERT INTO dbo.CPROLE (CODE, NAME) VALUES (N'ADMIN', N'Administrator');
IF NOT EXISTS (SELECT 1 FROM dbo.CPROLE WHERE CODE = N'AUDIENCE')
  INSERT INTO dbo.CPROLE (CODE, NAME) VALUES (N'AUDIENCE', N'Audience');
IF NOT EXISTS (SELECT 1 FROM dbo.CPROLE WHERE CODE = N'ORGANIZER')
  INSERT INTO dbo.CPROLE (CODE, NAME) VALUES (N'ORGANIZER', N'Event Organizer');
IF NOT EXISTS (SELECT 1 FROM dbo.CPROLE WHERE CODE = N'SELLER')
  INSERT INTO dbo.CPROLE (CODE, NAME) VALUES (N'SELLER', N'Seller');
IF NOT EXISTS (SELECT 1 FROM dbo.CPROLE WHERE CODE = N'OFFICER')
  INSERT INTO dbo.CPROLE (CODE, NAME) VALUES (N'OFFICER', N'Check-in Officer');

-- 2. Grants -------------------------------------------------------------------
-- ADMIN keeps the 8 SYSTEM menus (repairs old installs missing them).
INSERT INTO dbo.CPPERMISSION (ROLE_CODE, MENU_CODE)
SELECT N'ADMIN', CODE FROM dbo.CPMENU
WHERE MODULE = N'SYSTEM' AND MENU_KIND = N'CHILD'
AND NOT EXISTS (
  SELECT 1 FROM dbo.CPPERMISSION x WHERE x.ROLE_CODE = N'ADMIN' AND x.MENU_CODE = CPMENU.CODE
);
-- New roles: no grants (empty sidebar until assigned via matrix UI).

-- 3. Starter accounts (password = username + "123", bcrypt) --------------------
-- admin/admin123 (ADMIN)
IF NOT EXISTS (SELECT 1 FROM dbo.CPUSER WHERE USERNAME = 'admin')
  INSERT INTO dbo.CPUSER (CODE, USERNAME, EMAIL, PASSWORD, ROLE_CODE)
  VALUES ('USR-ADMIN', 'admin', 'admin@ticket.app',
    '$2a$10$pPedrlyKaNDNl.uZMHK.F.jSkchAcyDcz2kgU1qOQKmLXO/KECB3.', 'ADMIN');
ELSE
  UPDATE dbo.CPUSER SET PASSWORD = '$2a$10$pPedrlyKaNDNl.uZMHK.F.jSkchAcyDcz2kgU1qOQKmLXO/KECB3.',
    ROLE_CODE = 'ADMIN', EMAIL = 'admin@ticket.app', UPDATED_AT = GETDATE()
  WHERE USERNAME = 'admin';

-- audience/audience123 (AUDIENCE)
IF NOT EXISTS (SELECT 1 FROM dbo.CPUSER WHERE USERNAME = 'audience')
  INSERT INTO dbo.CPUSER (CODE, USERNAME, EMAIL, PASSWORD, ROLE_CODE)
  VALUES ('USR-AUDIENCE', 'audience', 'audience@ticket.app',
    '$2a$10$IZUzHE4snobTnbmKwb.lQ.APsOKdWvYpfermkIwtcMxiPMUUruTnW', 'AUDIENCE');
ELSE
  UPDATE dbo.CPUSER SET PASSWORD = '$2a$10$IZUzHE4snobTnbmKwb.lQ.APsOKdWvYpfermkIwtcMxiPMUUruTnW',
    ROLE_CODE = 'AUDIENCE', UPDATED_AT = GETDATE()
  WHERE USERNAME = 'audience';

-- organizer/organizer123 (ORGANIZER)
IF NOT EXISTS (SELECT 1 FROM dbo.CPUSER WHERE USERNAME = 'organizer')
  INSERT INTO dbo.CPUSER (CODE, USERNAME, EMAIL, PASSWORD, ROLE_CODE)
  VALUES ('USR-ORGANIZER', 'organizer', 'organizer@ticket.app',
    '$2a$10$HFQR9/SSJPcvsDX4HXMXKuvnYz9DV/JTelLL4bc/Decxehdj1d3b6', 'ORGANIZER');
ELSE
  UPDATE dbo.CPUSER SET PASSWORD = '$2a$10$HFQR9/SSJPcvsDX4HXMXKuvnYz9DV/JTelLL4bc/Decxehdj1d3b6',
    ROLE_CODE = 'ORGANIZER', UPDATED_AT = GETDATE()
  WHERE USERNAME = 'organizer';

-- seller/seller123 (SELLER)
IF NOT EXISTS (SELECT 1 FROM dbo.CPUSER WHERE USERNAME = 'seller')
  INSERT INTO dbo.CPUSER (CODE, USERNAME, EMAIL, PASSWORD, ROLE_CODE)
  VALUES ('USR-SELLER', 'seller', 'seller@ticket.app',
    '$2a$10$sZLQKWkONHQ4muRlWVZeqeO8gYIpvBepQuEPy6DHDxA1WpoWMDpre', 'SELLER');
ELSE
  UPDATE dbo.CPUSER SET PASSWORD = '$2a$10$sZLQKWkONHQ4muRlWVZeqeO8gYIpvBepQuEPy6DHDxA1WpoWMDpre',
    ROLE_CODE = 'SELLER', UPDATED_AT = GETDATE()
  WHERE USERNAME = 'seller';

-- officer/officer123 (OFFICER)
IF NOT EXISTS (SELECT 1 FROM dbo.CPUSER WHERE USERNAME = 'officer')
  INSERT INTO dbo.CPUSER (CODE, USERNAME, EMAIL, PASSWORD, ROLE_CODE)
  VALUES ('USR-OFFICER', 'officer', 'officer@ticket.app',
    '$2a$10$nHVgbB3.h39bb4k4yJbBmuzN6Ddmsqhi9t0LXuOOHAtMXTPyWLFde', 'OFFICER');
ELSE
  UPDATE dbo.CPUSER SET PASSWORD = '$2a$10$nHVgbB3.h39bb4k4yJbBmuzN6Ddmsqhi9t0LXuOOHAtMXTPyWLFde',
    ROLE_CODE = 'OFFICER', UPDATED_AT = GETDATE()
  WHERE USERNAME = 'officer';

COMMIT TRAN;
