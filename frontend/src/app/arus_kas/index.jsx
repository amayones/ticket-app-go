// ===== TEMPLATE SERAGAM app/* (ala <mod>.js langit_v2) =====
// Halaman contoh statis: shell standar + EmptyState, tanpa controller.
import { StandardPage } from '../shared/all.js'

const FEATURES = { header: true, refresh: false, filter: false, tabs: false, create: false, edit: false, remove: false, pagination: false, empty: true, error: true, confirmDialog: false, extraActions: false }

const CONFIG = {
  title: 'Arus Kas',
  description: 'Contoh menu child di bawah menu parent (permission MENU_ARUS_KAS, parent MENU_KEUANGAN).',
  emptyTitle: 'Menu contoh (child)',
  emptyDescription: 'Hierarki sidebar ini berasal dari CPMENU: MENU_KIND = CHILD dan PARENT_CODE = MENU_KEUANGAN.',
}

export const meta = { label: 'Arus Kas', icon: 'list', order: 3 }

// Contoh menu CHILD: baris CPMENU MENU_ARUS_KAS (MCONTROL=arus_kas,
// PARENT_CODE = MENU_KEUANGAN), jadi di sidebar tampil menjorok di bawah
// header parent "Keuangan". Foldernya datar: app/arus_kas/.
export default function ArusKas() {
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
