// TEMPLATE GRID menu baru — padanan GRID<mod>.js Tipe A langit_v2.
// GANTI: COLUMNS_ITEMS (header/dataIndex/width/editor), ID_FIELD,
// process_update/process_delete, details konfirmasi.
// Aturan anti-bug PPH: nama field HANYA ditulis di COLUMNS_ITEMS;
// validasi + payload + save dibaca otomatis dari sana.
import { forwardRef, useImperativeHandle, useState } from 'react'
import { EditableGrid, validateInput, useMessageBox, useToast } from '../../../app/shared/all.js'
import { process_update, process_delete } from './api.js'

// GANTI: kolom tabel <tabel> (header tampil, dataIndex = nama kolom DB/JSON)
const COLUMNS_ITEMS = [
  { header: 'Kode', dataIndex: 'code', width: 110, editor: { xtype: 'textfield', allowBlank: false, maxLength: 40 } },
  { header: 'Nama', dataIndex: 'name', width: 200, editor: { xtype: 'textfield', allowBlank: false, maxLength: 100 } },
]

// GANTI: field ID baris (kolom kunci, mis. code)
const ID_FIELD = 'code'

const GRID = forwardRef(function GRID({ rows, reload, openFRM }, ref) {
  const toast = useToast()
  const { MessageBox, dialog } = useMessageBox()
  const [editingKey, setEditingKey] = useState(null)
  const [draft, setDraft] = useState({})

  useImperativeHandle(ref, () => ({
    // Varian A: btnew delegasi ke sini — buka FRM untuk input baru.
    handler_btnew_click: () => openFRM({}),
  }))

  function handler_rowbtn_edit_click(row) {
    setEditingKey(row[ID_FIELD])
    setDraft({ ...row })
  }

  function handler_canceledit() {
    setEditingKey(null)
    setDraft({})
  }

  // Dibaca otomatis dari editor.allowBlank (jangan tulis nama field manual).
  function handler_validasi_inline(record) {
    for (const c of COLUMNS_ITEMS) {
      if (!c.editor || c.editor.allowBlank !== false) continue
      const v = record[c.dataIndex]
      if (v === '' || v == null || String(v).trim() === '') {
        return `${c.header} tidak boleh kosong`
      }
      if (c.editor.maxLength && String(v).length > c.editor.maxLength) {
        return `${c.header} maksimal ${c.editor.maxLength} karakter`
      }
    }
    return ''
  }

  function handler_validasi_payload(record) {
    const dtval = { ...record }
    delete dtval.id
    delete dtval.INLINE_NEW
    return dtval
  }

  async function handler_rowbtn_save() {
    try {
      const pesan = handler_validasi_inline(draft)
      if (pesan !== '') {
        toast.error(pesan, { title: 'Validasi' })
        return false
      }
      const dtval = handler_validasi_payload(draft)
      const result = await MessageBox.update({
        headline: 'Konfirmasi Update Data',
        details: [
          { label: 'Kode', value: dtval[ID_FIELD] },
          { label: 'Nama', value: dtval.name },
        ],
        request: () => process_update(dtval[ID_FIELD], dtval),
      })
      if (!result) return
      handler_canceledit()
      reload()
    } catch (ex) {
      toast.error(ex.message, { title: 'Error' })
    }
  }

  async function handler_rowbtn_delete_click(row) {
    try {
      const dtval = handler_validasi_payload(row)
      const result = await MessageBox.delete({
        headline: 'Konfirmasi Delete Data',
        details: [
          { label: 'Kode', value: dtval[ID_FIELD] },
          { label: 'Nama', value: dtval.name },
        ],
        request: () => process_delete(dtval[ID_FIELD]),
      })
      if (!result) return
      reload()
    } catch (ex) {
      toast.error(ex.message, { title: 'Error' })
    }
  }

  return (
    <>
      <EditableGrid
        columns={COLUMNS_ITEMS}
        rows={rows}
        idOf={(r) => r[ID_FIELD]}
        editingKey={editingKey}
        draft={draft}
        setDraft={setDraft}
        onEdit={handler_rowbtn_edit_click}
        onSave={handler_rowbtn_save}
        onCancelEdit={handler_canceledit}
        onDelete={handler_rowbtn_delete_click}
      />
      {dialog}
    </>
  )
})

GRID.columns = COLUMNS_ITEMS

export default GRID
