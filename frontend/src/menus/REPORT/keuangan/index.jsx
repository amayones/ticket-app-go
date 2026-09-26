import { Card, CardTitle, EmptyState } from '../../../components'

export const meta = { label: 'Keuangan', icon: 'list', order: 2 }

// Contoh menu PARENT: baris CPMENU ini MENU_KIND = PARENT, sehingga di
// sidebar bisa dibuka-turut dan memuat anak-anaknya (MENU_ARUS_KAS punya
// PARENT_CODE = MENU_KEUANGAN). Menu parent juga halaman biasa: punya
// folder sendiri seperti menu lain (menus/REPORT/keuangan/).
export default function Keuangan() {
  return (
    <Card>
      <div className="mb-4">
        <CardTitle description="Contoh menu parent (permission MENU_KEUANGAN, MENU_KIND = PARENT).">
          Keuangan
        </CardTitle>
      </div>
      <EmptyState
        title="Menu contoh (parent)"
        description="Klik chevron di samping nama menu ini untuk melihat dan menyembunyikan menu anak (Arus Kas)."
      />
    </Card>
  )
}
