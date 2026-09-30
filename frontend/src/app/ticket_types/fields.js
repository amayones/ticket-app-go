// Items form ala FRM<mod>.js langit_v2: yang beda per tabel hanya bagian ini.
// Dipakai StandardForm pada modal tambah/ubah tipe tiket.
export function ticketTypeFields() {
  return [
    { name: 'name', label: 'Nama tipe', type: 'text', placeholder: 'Regular, VIP…' },
    { name: 'description', label: 'Benefit', type: 'textarea', rows: 2, placeholder: 'Yang didapat pembeli…' },
    { name: 'price', label: 'Harga (0 = gratis)', type: 'number', placeholder: '150000' },
    { name: 'quota', label: 'Kuota', type: 'number', placeholder: '100' },
    { name: 'max_per_person', label: 'Batas per orang (kosong = bebas)', type: 'number', placeholder: '' },
    { name: 'sale_start_at', label: 'Penjualan mulai (opsional)', type: 'text', placeholder: 'YYYY-MM-DD HH:MM' },
    { name: 'sale_end_at', label: 'Penjualan selesai (opsional)', type: 'text', placeholder: 'YYYY-MM-DD HH:MM' },
  ]
}
