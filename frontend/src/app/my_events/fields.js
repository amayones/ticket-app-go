// Items form ala FRM<mod>.js langit_v2: yang beda per tabel hanya bagian ini.
// Dipakai StandardForm pada FRM Create Event (bagian Informasi + S&K).
// Tanggal dan poster memakai input khusus di FRM (datetime-local + file).
export function eventBasicFields(categories = [], cities = []) {
  return [
    { name: 'title', label: 'Judul event', type: 'text', placeholder: 'Contoh: Konser Merdeka', required: true },
    {
      name: 'category_code',
      label: 'Kategori',
      type: 'select',
      options: [{ value: '', label: '— Pilih kategori —' }].concat(
        categories.map((c) => ({ value: c.code, label: c.name })),
      ),
    },
    { name: 'description', label: 'Deskripsi', type: 'textarea', rows: 4, placeholder: 'Ceritakan event Anda…' },
    {
      name: 'city_code',
      label: 'Kota',
      type: 'select',
      options: [{ value: '', label: '— Pilih kota —' }].concat(cities.map((c) => ({ value: c.code, label: c.name }))),
    },
    { name: 'venue', label: 'Venue', type: 'text', placeholder: 'Contoh: GBK Senayan', required: true },
    { name: 'address', label: 'Alamat', type: 'text', placeholder: 'Jalan, nomor, patokan' },
    { name: 'terms', label: 'Syarat & Ketentuan (opsional)', type: 'textarea', rows: 3, placeholder: 'Aturan pengunjung…' },
  ]
}
