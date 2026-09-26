import { Card, CardTitle, EmptyState } from '../../components'

export const meta = { label: 'Arus Kas', icon: 'list', order: 3 }

// Contoh menu CHILD: baris CPMENU MENU_ARUS_KAS (MCONTROL=arus_kas,
// PARENT_CODE = MENU_KEUANGAN), jadi di sidebar tampil menjorok di bawah
// header parent "Keuangan". Foldernya datar: app/arus_kas/.
export default function ArusKas() {
  return (
    <Card>
      <div className="mb-4">
        <CardTitle description="Contoh menu child di bawah menu parent (permission MENU_ARUS_KAS, parent MENU_KEUANGAN).">
          Arus Kas
        </CardTitle>
      </div>
      <EmptyState
        title="Menu contoh (child)"
        description="Hierarki sidebar ini berasal dari CPMENU: MENU_KIND = CHILD dan PARENT_CODE = MENU_KEUANGAN."
      />
    </Card>
  )
}
