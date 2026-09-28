// Items form ala FRM<mod>.js langit_v2: yang beda per tabel hanya bagian ini.
// Dipakai StandardForm pada modal Tambah user.
export function userCreateFields(roles = []) {
  const list = roles.length > 0 ? roles : [{ code: 'USER', name: '' }]
  return [
    { name: 'username', label: 'Username', type: 'text', placeholder: 'min. 3 karakter' },
    { name: 'email', label: 'Email', type: 'email', placeholder: 'nama@email.com' },
    {
      name: 'password',
      label: 'Password awal',
      type: 'password',
      placeholder: 'min. 8 karakter',
      hint: 'Beritahu password ini ke user. Untuk menggantinya, admin buka User Account → edit → Password baru.',
    },
    {
      name: 'role',
      label: 'Role awal',
      type: 'select',
      options: list.map((r) => ({ value: r.code, label: r.name ? `${r.code} — ${r.name}` : r.code })),
    },
  ]
}
