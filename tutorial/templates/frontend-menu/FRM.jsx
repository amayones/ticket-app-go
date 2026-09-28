// TEMPLATE FRM menu baru — padanan FRM<mod>.js langit_v2 (window form).
// GANTI: TITLE, FORM_FIELDS (name/label/type), validate_field, details,
// process_create/process_update/process_delete.
import { useEffect, useState } from 'react'
import { Alert, Button, Modal, StandardForm, validateInput, useMessageBox, useToast } from '../../../app/shared/all.js'
import { process_create, process_update, process_delete } from './api.js'

// GANTI: field form (name = kunci payload, label tampil)
const FORM_FIELDS = [
  { name: 'code', label: 'Kode', type: 'text', placeholder: 'mis. STK-001' },
  { name: 'name', label: 'Nama', type: 'text', placeholder: 'nama item' },
]

// GANTI: validasi (field/type/msg) — loop-nya tetap, jangan diubah
const validate_field = [
  { field: 'code', type: 'text', msg: 'Kode tidak boleh kosong' },
  { field: 'name', type: 'text', msg: 'Nama tidak boleh kosong' },
]

export default function FRM({ open, initial, onClose, onSaved }) {
  const toast = useToast()
  const { MessageBox, dialog } = useMessageBox()
  const isNew = !initial?.code
  const [values, setValues] = useState({ code: '', name: '' })
  const [formError, setFormError] = useState('')
  const [saving, setSaving] = useState(false)

  useEffect(() => {
    if (open) {
      setValues({ code: initial?.code || '', name: initial?.name || '' })
      setFormError('')
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [open])

  function handler_validasi_input() {
    const { valid, message } = validateInput(values, validate_field)
    if (!valid) {
      setFormError(message)
      toast.error(message, { title: 'Validasi' })
      return false
    }
    return true
  }

  async function handler_btsave() {
    try {
      if (handler_validasi_input() === false) return false
      setSaving(true)
      if (isNew) {
        const result = await MessageBox.create({
          headline: 'Konfirmasi Simpan Data',
          details: [
            { label: 'Kode', value: values.code },
            { label: 'Nama', value: values.name },
          ],
          request: () => process_create(values),
        })
        if (!result) return
        toast.success('Data berhasil disimpan.', { title: 'Success' })
      } else {
        const result = await MessageBox.update({
          headline: 'Konfirmasi Update Data',
          details: [
            { label: 'Kode', value: values.code },
            { label: 'Nama', value: values.name },
          ],
          request: () => process_update(values.code, values),
        })
        if (!result) return
        toast.success('Data berhasil diupdate.', { title: 'Success' })
      }
      onClose()
      onSaved()
    } catch (ex) {
      setFormError(ex.message)
    } finally {
      setSaving(false)
    }
  }

  async function handler_btdelete() {
    try {
      if (!values.code) return
      const result = await MessageBox.delete({
        headline: 'Konfirmasi Delete Data',
        details: [
          { label: 'Kode', value: values.code },
          { label: 'Nama', value: values.name },
        ],
        request: () => process_delete(values.code),
      })
      if (!result) return
      toast.success('Data berhasil dihapus.', { title: 'Success' })
      onClose()
      onSaved()
    } catch (ex) {
      setFormError(ex.message)
    }
  }

  return (
    <>
      <Modal
        open={open}
        onClose={saving ? undefined : onClose}
        title={isNew ? 'Create <judul>' : 'Edit <judul>'}
        size="sm"
        closeOnBackdrop={!saving}
        footer={
          <>
            {!isNew && (
              <Button variant="danger" onClick={handler_btdelete} disabled={saving}>
                Delete
              </Button>
            )}
            <span className="flex-1" />
            <Button variant="secondary" onClick={onClose} disabled={saving}>
              Batal
            </Button>
            <Button onClick={handler_btsave} loading={saving}>
              Save
            </Button>
          </>
        }
      >
        <div className="flex flex-col gap-4">
          {formError && (
            <Alert tone="error" closable onClose={() => setFormError('')}>
              {formError}
            </Alert>
          )}
          <StandardForm fields={FORM_FIELDS} values={values} onChange={setValues} />
        </div>
      </Modal>
      {dialog}
    </>
  )
}
