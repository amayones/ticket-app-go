import { Card, CardTitle, EmptyState } from '../../components'

export const meta = { label: 'Laporan', icon: 'list', order: 1 }

// Contoh menu BIASA: satu folder = satu halaman, tanpa parent.
// MCONTROL=laporan -> app/laporan/. Bandingkan dengan header PARENT
// MENU_KEUANGAN (tanpa folder) dan app/arus_kas/ (CHILD dari MENU_KEUANGAN).
// Lihat tutorial/README.md Bagian 2 untuk cara membuat menu seperti ini.
export default function Laporan() {
  return (
    <Card>
      <div className="mb-4">
        <CardTitle description="Contoh menu biasa di modul REPORT (permission MENU_LAPORAN). Belum ada data bisnis — halaman ini hanya untuk membandingkan tampilan dengan menu bersarang.">
          Laporan
        </CardTitle>
      </div>
      <EmptyState
        title="Menu contoh"
        description="Isi halaman ini dengan tabel/grafik laporan mengikuti pola menu lain (load → SkeletonRows → Alert → EmptyState → Pagination)."
      />
    </Card>
  )
}
