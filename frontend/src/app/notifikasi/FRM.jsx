// FRM Template Notifikasi — padanan FRM<mod>.js langit_v2 (window form).
// Items form + validate_field inline (satu-satunya yang beda per tabel).
import { useEffect, useState } from 'react'
import { Alert, Button, Modal, StandardForm, validateInput, useMessageBox, useToast } from '../shared/all.js'
import { process_create, process_update, process_delete } from './api.js'

const CHANNELS = ['EMAIL', 'PUSH', 'INAPP']

const FORM_FIELDS = [
  { name: 'name', label: 'Nama template', type: 'text', placeholder: 'mis. Promo Akhir Tahun' },
  { name: 'channel', label: 'Channel', type: 'select', options: CHANNELS.map((c) => ({ value: c, label: c })) },
  { name: 'subject', label: 'Subjek', type: 'text', placeholder: 'dipakai untuk EMAIL' },
  {
    name: 'body',
    label: 'Isi pesan',
    type: 'textarea',
    rows: 5,
    placeholder: 'Halo {{nama}}, …',
    hint: 'Variabel: {{nama}} {{kode}} {{role}} {{detail}}',
  },
  { name: 'is_active', label: 'Template aktif (nonaktif = tidak bisa dikirim)', type: 'checkbox' },
]

const validate_field = [
  { field: 'name', type: 'text', msg: 'Nama template tidak boleh kosong' },
  { field: 'body', type: 'text', msg: 'Isi pesan tidak boleh kosong' },
]

export default function FRM({ open, initial, onClose, onSaved }) {
  const toast = useToast()
  const { MessageBox, dialog } = useMessageBox()
  const isNew = !initial || initial === 'new'
  const [values, setValues] = useState({ name: '', channel: 'EMAIL', subject: '', body: '', is_active: true })
  const [formError, setFormError] = useState('')
  const [saving, setSaving] = useState(false)

  useEffect(() => {
    if (open) {
      const src = isNew ? null : initial
      setValues({
        name: src?.name || '',
        channel: src?.channel || 'EMAIL',
        subject: src?.subject || '',
        body: src?.body || '',
        is_active: src ? !!src.is_active : true,
      })
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
            { label: 'Nama', value: values.name },
            { label: 'Channel', value: values.channel },
          ],
          request: () => process_create(values),
        })
        if (!result) return
        toast.success('Template dibuat.', { title: 'Success' })
      } else {
        const result = await MessageBox.update({
          headline: 'Konfirmasi Update Data',
          details: [
            { label: 'Nama', value: values.name },
            { label: 'Channel', value: values.channel },
          ],
          request: () => process_update(initial.code, values),
        })
        if (!result) return
        toast.success('Template diperbarui.', { title: 'Success' })
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
      const result = await MessageBox.delete({
        headline: `Hapus template ${initial?.name}?`,
        description: 'Template dihapus permanen. Riwayat pengiriman yang sudah tercatat tidak ikut terhapus.',
        request: () => process_delete(initial.code),
      })
      if (!result) return
      toast.success(`Template ${initial.name} dihapus.`)
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
        title={isNew ? 'Template baru' : `Edit ${initial?.name || ''}`}
        size="lg"
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
