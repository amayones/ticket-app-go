// View Laporan — contoh menu BIASA statis (satu folder = satu halaman).
// MCONTROL=laporan -> app/laporan/. Tanpa store/GRID (seperti stub grafik).
import { useEffect } from 'react'
import { StandardPage } from '../shared/all.js'
import { controller } from './controller.js'

export const meta = { label: 'Laporan', icon: 'list', order: 1 }

export default function Laporan({ nvdata }) {
  useEffect(() => {
    controller.init()
  }, [])

  return (
    <StandardPage
      title={nvdata?.label || 'Laporan'}
      description="Contoh menu biasa di modul REPORT (permission MENU_LAPORAN). Belum ada data bisnis — halaman ini hanya untuk membandingkan tampilan dengan menu bersarang."
      items={[]}
      emptyTitle="Menu contoh"
      emptyDescription="Isi halaman ini dengan tabel/grafik laporan mengikuti pola menu CRUD (index + controller + GRID + FRM + api)."
    />
  )
}
