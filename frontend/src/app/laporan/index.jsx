// ===== TEMPLATE SERAGAM app/* (ala <mod>.js langit_v2) =====
// Halaman contoh statis: shell standar + EmptyState, tanpa controller.
import { StandardPage } from '../shared/all.js'

const FEATURES = { header: true, refresh: false, filter: false, tabs: false, create: false, edit: false, remove: false, pagination: false, empty: true, error: true, confirmDialog: false, extraActions: false }

const CONFIG = {
  title: 'Laporan',
  description: 'Contoh menu biasa di modul REPORT (permission MENU_LAPORAN). Belum ada data bisnis — halaman ini hanya untuk membandingkan tampilan dengan menu bersarang.',
  emptyTitle: 'Menu contoh',
  emptyDescription: 'Isi halaman ini dengan tabel/grafik laporan mengikuti pola menu lain (load → SkeletonRows → Alert → EmptyState → Pagination).',
}

export const meta = { label: 'Laporan', icon: 'list', order: 1 }

// Contoh menu BIASA: satu folder = satu halaman, tanpa parent.
// MCONTROL=laporan -> app/laporan/. Bandingkan dengan header PARENT
// MENU_KEUANGAN (tanpa folder) dan app/arus_kas/ (CHILD dari MENU_KEUANGAN).
// Lihat tutorial/README.md Bagian 2 untuk cara membuat menu seperti ini.
export default function Laporan() {
  return (
    <StandardPage
      title={CONFIG.title}
      description={CONFIG.description}
      features={FEATURES}
      items={[]}
      emptyTitle={CONFIG.emptyTitle}
      emptyDescription={CONFIG.emptyDescription}
    />
  )
}
