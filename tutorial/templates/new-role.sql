-- TEMPLATE role baru. Ganti EDITOR + permission sesuai kebutuhan, jalankan
-- di database (sqlcmd / SSMS / psql / sqlite3 — sintaks di bawah standar).
-- Menu frontend untuk role ini: taruh di frontend/src/menus/user/ (otomatis
-- terlihat oleh SEMUA role); yang membedakan antar-role adalah permission
-- di bawah ini + RequirePermission di routes (lihat tutorial/README.md).

-- 1. Daftarkan role (kode huruf besar, maks 20 karakter):
INSERT INTO CPROLE (CODE, NAME)
SELECT 'EDITOR', 'Editor'
WHERE NOT EXISTS (SELECT 1 FROM CPROLE WHERE CODE = 'EDITOR');

-- 2. Beri permission (contoh hak baca dasar; tambah sesuai kebutuhan):
INSERT INTO CPROLEPERMISSION (ROLE_CODE, PERMISSION_CODE)
SELECT 'EDITOR', v.CODE FROM (VALUES
  ('USER_READ'),
  ('SESSION_READ'),
  ('SESSION_REVOKE')
) v(CODE)
WHERE NOT EXISTS (
  SELECT 1 FROM CPROLEPERMISSION x
  WHERE x.ROLE_CODE = 'EDITOR' AND x.PERMISSION_CODE = v.CODE
);

-- 3. Pindahkan user ke role baru (ganti USR-XXXXXX dengan kode user):
-- UPDATE CPUSER SET ROLE_CODE = 'EDITOR' WHERE CODE = 'USR-XXXXXX';

-- 4. Hapus role bila salah (gagal bila masih dipakai user):
-- DELETE FROM CPROLE WHERE CODE = 'EDITOR';
