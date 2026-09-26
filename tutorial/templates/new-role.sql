-- TEMPLATE role baru. Ganti EDITOR + permission sesuai kebutuhan, jalankan
-- di SQL Server (sqlcmd / SSMS).
-- Menu frontend untuk role ini: taruh di frontend/src/menus/<MCONTROL>/...
-- (modul UPPERCASE sesuai CPMENU.MCONTROL); menu hanya tampil bila
-- permission MENU_<NAMA_MENU> di bawah diberikan.
-- Permission yang sama dipakai RequirePermission di routes.

-- 1. Daftarkan role (kode huruf besar, maks 20 karakter):
INSERT INTO CPROLE (CODE, NAME)
SELECT 'EDITOR', 'Editor'
WHERE NOT EXISTS (SELECT 1 FROM CPROLE WHERE CODE = 'EDITOR');

-- 2. Beri permission menu (contoh: users dan sesi).
--    Role baru tanpa centang apa pun = halaman kosong (hubungi admin).
INSERT INTO CPPERMISSION (ROLE_CODE, MENU_CODE)
SELECT 'EDITOR', v.CODE FROM (VALUES
  ('MENU_USERS'),
  ('MENU_SESSIONS')
) v(CODE)
WHERE NOT EXISTS (
  SELECT 1 FROM CPPERMISSION x
  WHERE x.ROLE_CODE = 'EDITOR' AND x.MENU_CODE = v.CODE
);

-- 3. Pindahkan user ke role baru (ganti USR-XXXXXX dengan kode user):
-- UPDATE CPUSER SET ROLE_CODE = 'EDITOR' WHERE CODE = 'USR-XXXXXX';

-- 4. Hapus role bila salah (gagal bila masih dipakai user):
-- DELETE FROM CPROLE WHERE CODE = 'EDITOR';
