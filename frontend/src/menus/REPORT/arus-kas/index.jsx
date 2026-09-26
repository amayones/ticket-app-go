import { Card, CardTitle, EmptyState } from '../../../components'

export const meta = { label: 'Arus Kas', icon: 'list', order: 3 }

// Contoh menu CHILD: baris CPMENU ini MENU_KIND = CHILD dengan
// PARENT_CODE = MENU_KEUANGAN, jadi di sidebar tampil menjorok di bawah
// menu parent "Keuangan". Foldernya tetap datar satu level di dalam
// folder modul: menus/REPORT/arus-kas/ (tidak ada folder perantara).
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
