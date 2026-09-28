// View Arus Kas — contoh menu CHILD di bawah header PARENT Keuangan
// (MENU_KIND=CHILD, PARENT_CODE=MENU_KEUANGAN). Folder datar: app/arus_kas/.
import { useEffect } from 'react'
import { StandardPage } from '../shared/all.js'
import { controller } from './controller.js'

export const meta = { label: 'Arus Kas', icon: 'list', order: 3 }

export default function ArusKas({ nvdata }) {
  useEffect(() => {
    controller.init()
  }, [])

  return (
    <StandardPage
      title={nvdata?.label || 'Arus Kas'}
      description="Contoh menu child di bawah menu parent (permission MENU_ARUS_KAS, parent MENU_KEUANGAN)."
      items={[]}
      emptyTitle="Menu contoh (child)"
      emptyDescription="Hierarki sidebar ini berasal dari CPMENU: MENU_KIND = CHILD dan PARENT_CODE = MENU_KEUANGAN."
    />
  )
}
