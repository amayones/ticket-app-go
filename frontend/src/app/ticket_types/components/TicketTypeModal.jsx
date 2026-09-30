// FRM Ticket Type — padanan FRM<mod>.js langit_v2 (window form + tbar save).
// Dipakai tambah (btnew) dan ubah. Validasi ringan di klien, lengkap di backend.
import { useEffect, useState } from 'react'
import { Alert, Button, Modal, StandardForm, useToast } from '../../shared/all.js'
import { process_create, process_update } from '../api.js'
import { ticketTypeFields } from '../fields.js'

const EMPTY = { name: '', description: '', price: '', quota: '', max_per_person: '', sale_start_at: '', sale_end_at: '' }

export default function TicketTypeModal({ open, eventCode, editing, onClose, onSaved }) {
  const toast = useToast()
  const [values, setValues] = useState(EMPTY)
  const [formError, setFormError] = useState('')
  const [saving, setSaving] = useState(false)

  useEffect(() => {
    if (open) {
      if (editing) {
        setValues({
          name: editing.name || '',
          description: editing.description || '',
          price: editing.price ?? '',
          quota: editing.quota ?? '',
          max_per_person: editing.max_per_person || '',
          sale_start_at: editing.sale_start_at || '',
          sale_end_at: editing.sale_end_at || '',
        })
      } else {
        setValues(EMPTY)
      }
      setFormError('')
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [open])

  function handler_validasi_input() {
    if (!String(values.name || '').trim()) {
      setFormError('Nama tipe wajib diisi.')
      return false
    }
    if (values.price === '' || Number(values.price) < 0) {
      setFormError('Harga tidak boleh negatif (0 = gratis).')
      return false
    }
    if (values.quota === '' || Number(values.quota) < 1) {
      setFormError('Kuota minimal 1.')
      return false
    }
    return true
  }

  function toPayload() {
    const num = (v) => (v === '' || v === null ? null : Number(v))
    return {
      name: String(values.name).trim(),
      description: String(values.description || ''),
      price: Number(values.price),
      quota: Number(values.quota),
      max_per_person: num(values.max_per_person),
      sale_start_at: String(values.sale_start_at || '').trim(),
      sale_end_at: String(values.sale_end_at || '').trim(),
    }
  }

  async function handler_btsave() {
    try {
      if (handler_validasi_input() === false) return false
      setSaving(true)
      if (editing) {
        await process_update(eventCode, editing.code, toPayload())
        toast.success(`Tipe ${values.name} diperbarui.`)
      } else {
        await process_create(eventCode, toPayload())
        toast.success(`Tipe ${values.name} dibuat.`)
      }
      onClose()
      onSaved()
    } catch (ex) {
      setFormError(ex.message)
    } finally {
      setSaving(false)
    }
  }

  return (
    <Modal
      open={open}
      onClose={saving ? undefined : onClose}
      title={editing ? 'Ubah tipe tiket' : 'Tambah tipe tiket'}
      size="sm"
      closeOnBackdrop={!saving}
      footer={
        <>
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
        <StandardForm fields={ticketTypeFields()} values={values} onChange={setValues} />
      </div>
    </Modal>
  )
}
