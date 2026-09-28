// FRM User Account — padanan FRM<mod>.js langit_v2 (window form + tbar save).
// Dipakai untuk New Input (btnew Varian A delegasi ke GRID.handler_btnew_click
// yang membuka FRM ini). Edit profil tetap via EditUserModal (FRM kedua).
import { useEffect, useState } from 'react'
import { Alert, Button, Modal, StandardForm, validateInput, useMessageBox, useToast } from '../shared/all.js'
import { process_create } from './api.js'
import { userCreateFields } from './fields.js'

const validate_field = [
  { field: 'username', type: 'text', msg: 'Username tidak boleh kosong' },
  { field: 'email', type: 'text', msg: 'Email tidak boleh kosong' },
  { field: 'password', type: 'text', msg: 'Password tidak boleh kosong' },
]

export default function FRM({ open, roles, onClose, onSaved }) {
  const toast = useToast()
  const { MessageBox, dialog } = useMessageBox()
  const [values, setValues] = useState({ username: '', email: '', password: '', role: 'USER' })
  const [formError, setFormError] = useState('')
  const [saving, setSaving] = useState(false)

  useEffect(() => {
    if (open) {
      setValues({ username: '', email: '', password: '', role: 'USER' })
      setFormError('')
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [open])

  function handler_validasi_input() {
    const { valid, message } = validateInput(values, validate_field)
    if (!valid) {
      setFormError(message)
      return false
    }
    return true
  }

  async function handler_btsave() {
    try {
      if (handler_validasi_input() === false) return false
      setSaving(true)
      const username = values.username.trim()
      const result = await MessageBox.create({
        headline: 'Konfirmasi Simpan Data',
        details: [
          { label: 'Username', value: `@${username}` },
          { label: 'Email', value: values.email.trim().toLowerCase() },
          { label: 'Role', value: values.role },
        ],
        request: () => process_create(values),
      })
      if (!result) return
      toast.success(`Akun @${username} dibuat.`, { title: 'User dibuat' })
      onClose()
      onSaved()
    } catch (ex) {
      setFormError(ex.message)
    } finally {
      setSaving(false)
    }
  }

  return (
    <>
      <Modal
        open={open}
        onClose={saving ? undefined : onClose}
        title="Tambah user baru"
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
          <StandardForm fields={userCreateFields(roles)} values={values} onChange={setValues} />
        </div>
      </Modal>
      {dialog}
    </>
  )
}
