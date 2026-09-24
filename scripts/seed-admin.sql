-- Seed 2 akun awal (sama seperti database dev): admin/ADMIN + user/USER.
-- Password keduanya sama dengan username-nya (admin/admin, user/user).
-- WAJIB diganti setelah login pertama (Dashboard > Keamanan akun), dan
-- JANGAN dipakai di production tanpa diganti!
--
-- SQL Server: sqlcmd -S localhost,1433 -U <user> -P <pass> -d Go -C -i scripts/seed-admin.sql
-- (Jalankan SETELAH migrate.sql + migrate2_rbac.sql.)
IF NOT EXISTS (SELECT 1 FROM dbo.CPUSER WHERE USERNAME = 'admin')
  INSERT INTO dbo.CPUSER (CODE, USERNAME, EMAIL, PASSWORD, ROLE_CODE)
  VALUES ('USR-000004', 'admin', 'admin@example.com',
    '$2a$10$vq5MqkuG4/mD/twgYMBsxeOwbrmL3xE88rIpOEzMYt./TmL8xKAqO', 'ADMIN');
IF NOT EXISTS (SELECT 1 FROM dbo.CPUSER WHERE USERNAME = 'user')
  INSERT INTO dbo.CPUSER (CODE, USERNAME, EMAIL, PASSWORD, ROLE_CODE)
  VALUES ('USR-USER0001', 'user', 'user@example.com',
    '$2a$10$MpKcDZmTSotcNefe53ODl.3frNUr0E/fuU/cpAKPglu85tG0Niyr6', 'USER');
