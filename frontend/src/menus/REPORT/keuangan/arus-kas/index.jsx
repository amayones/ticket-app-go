import { Card, CardTitle, EmptyState } from '../../../../components'

export const meta = { label: 'Arus Kas', icon: 'list', order: 1 }

// Contoh menu DI DALAM GRUP VISUAL: folder frontend
// (menus/REPORT/keuangan/arus-kas/) tidak punya baris menu sendiri di CPMENU,
// jadi "Keuangan" di sidebar murni grup tampilan, bukan menu (tidak bisa
// diklik, tidak punya permission). Menu ini tetap menu biasa di DB.
export default function ArusKas() {
  return (
    <Card>
      <div className="mb-4">
        <CardTitle description="Contoh menu di dalam grup visual (permission MENU_ARUS_KAS, tanpa PARENT_CODE).">
          Arus Kas
        </CardTitle>
      </div>
      <EmptyState
        title="Menu contoh di dalam grup"
        description="Bandingkan dengan menu Laporan (di luar grup) untuk melihat perbedaan tampilan sidebar."
      />
    </Card>
  )
}
