// TEMPLATE view menu baru — padanan <mod>.js langit_v2.
// Cara clone (sama seperti clone modul di langit_v2):
// 1. Daftarkan di CPMENU via UI Modul & Menu: MODULE (section sidebar),
//    CODE (mis. MENU_STOK), LABEL (judul), MCONTROL snake_case (mis. stok),
//    KIND=CHILD (+ PARENT_CODE bila anak header). PARENT tanpa folder.
// 2. Copy folder tutorial/templates/frontend-menu/ ke app/<mcontrol>/
//    (mis. app/stok/). Judul = LABEL CPMENU; folder = MCONTROL.
// 3. Ganti <mcontrol> di 5 file ini (index, controller, GRID, FRM, api)
//    + <MENU_...> + <judul> + <tabel>. Titiknya ditandai GANTI.
// 4. Isi columns.items di GRID.jsx + validate_field di FRM.jsx.
// 5. Grant role via Role & Permission, login ulang, npm run build.
import { useEffect, useRef, useState } from 'react'
import { StandardPage, DownloadButton, useStandardController } from '../../../app/shared/all.js'
import { read_data } from './api.js'
import { controller } from './controller.js'
import GRID from './GRID.jsx'
import FRM from './FRM.jsx'

// GANTI: mcontrol folder ini (mis. stok)
const MCONTROL = '<mcontrol>'

export const meta = { label: '<judul>', icon: 'list', order: 50 }

export default function MenuBaru({ nvdata }) {
  const [showFRM, setShowFRM] = useState(null) // null | {} (new) | row (edit via FRM)
  const gridRef = useRef(null)

  const { rows, offset, setOffset, loading, showLoading, error, setError, load } =
    useStandardController(({ limit, offset }) => read_data({ limit, offset }), {})

  useEffect(() => {
    controller.init()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  return (
    <div className="flex flex-col gap-4">
      <StandardPage
        title={nvdata?.label || '<judul>'}
        description="GANTI: deskripsi menu ini."
        loading={loading}
        showLoading={showLoading}
        error={error}
        errorTitle="Gagal memuat data"
        onClearError={() => setError('')}
        onRefresh={() => controller.btrefresh_click(load)}
        onCreate={() => controller.btnew_click(() => gridRef.current?.handler_btnew_click())}
        createLabel="New Input"
        extraActions={<DownloadButton rows={rows} columns={GRID.columns} filename={MCONTROL} />}
        items={rows}
        emptyTitle="Belum ada data"
        emptyDescription="Klik New Input untuk menambah data pertama."
        offset={offset}
        onPage={setOffset}
      >
        <GRID
          ref={gridRef}
          rows={rows}
          reload={load}
          openFRM={(row) => setShowFRM(row || {})}
        />
      </StandardPage>

      <FRM
        open={showFRM !== null}
        initial={showFRM}
        onClose={() => setShowFRM(null)}
        onSaved={load}
      />
    </div>
  )
}
