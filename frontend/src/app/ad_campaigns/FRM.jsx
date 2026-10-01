import { useEffect, useState } from 'react'
import { Alert, Button, Modal, SafeImage, useToast, INPUT_CLASS, SELECT_CLASS } from '../shared/all.js'
import { uploadBanner } from './api.js'
import { formatAmount } from '../../components/format.js'

function toLocal(v) {
  if (!v) return ''
  const d = new Date(String(v).replace(' ', 'T'))
  if (Number.isNaN(d.getTime())) return ''
  const p = (n) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${p(d.getMonth() + 1)}-${p(d.getDate())}T${p(d.getHours())}:${p(d.getMinutes())}`
}
function fromLocal(v) {
  if (!v) return ''
  return `${v.slice(0, 10)} ${v.slice(11, 16)}:00`
}

export default function AdModal({ open, editing, positions, events, onClose, onSaved }) {
  const toast = useToast()
  const isNew = !editing
  const [values, setValues] = useState({ title: '', position: '', start_at: '', end_at: '', image_url: '', target_type: 'url', target_url: '', target_event: '' })
  const [preview, setPreview] = useState('')
  const [errors, setErrors] = useState({})
  const [formError, setFormError] = useState('')
  const [saving, setSaving] = useState(false)
  const [uploading, setUploading] = useState(false)
  const [dirty, setDirty] = useState(false)

  useEffect(() => {
    if (!open) return
    setFormError(''); setErrors({}); setDirty(false)
    if (editing) {
      const isEvent = editing.target_url && editing.target_url.startsWith('/events/')
      setValues({
        title: editing.title || '',
        position: editing.position || '',
        start_at: toLocal(editing.start_at),
        end_at: toLocal(editing.end_at),
        image_url: editing.image_url || '',
        target_type: isEvent ? 'event' : 'url',
        target_url: isEvent ? '' : editing.target_url || '',
        target_event: isEvent ? editing.target_url.replace('/events/', '') : '',
      })
      setPreview(editing.image_url || '')
    } else {
      setValues({ title: '', position: '', start_at: '', end_at: '', image_url: '', target_type: 'url', target_url: '', target_event: '' })
      setPreview('')
    }
  }, [open, editing])

  function change(next) { setValues(next); setDirty(true) }
  function handleClose() {
    if (dirty && !window.confirm('Form belum disimpan. Tutup?')) return
    onClose()
  }

  async function handleBanner(file) {
    if (!file) return
    if (!['image/jpeg', 'image/png', 'image/webp', 'image/gif'].includes(file.type)) { setErrors((p) => ({ ...p, image_url: 'Tipe harus JPG/PNG/WebP/GIF' })); return }
    if (file.size > 2 * 1024 * 1024) { setErrors((p) => ({ ...p, image_url: 'Maks 2 MB' })); return }
    setUploading(true)
    try {
      const res = await uploadBanner(file)
      change({ ...values, image_url: res.url })
      setPreview(res.url)
      setErrors((p) => ({ ...p, image_url: '' }))
    } catch (ex) { setFormError(ex.message) } finally { setUploading(false) }
  }

  function validate() {
    const errs = {}
    if (values.title.trim().length < 3) errs.title = 'Judul minimal 3 karakter'
    if (!values.position) errs.position = 'Wajib'
    if (!values.start_at) errs.start_at = 'Wajib'
    if (!values.end_at) errs.end_at = 'Wajib'
    if (values.start_at && values.end_at && values.end_at <= values.start_at) errs.end_at = 'Harus setelah mulai'
    if (!values.image_url.trim()) errs.image_url = 'Gambar wajib'
    if (values.target_type === 'url') {
      if (!values.target_url.trim()) errs.target_url = 'URL wajib'
      else if (!/^https?:\/\//i.test(values.target_url.trim())) errs.target_url = 'Harus http/https'
    } else if (!values.target_event) errs.target_event = 'Pilih event'
    setErrors(errs)
    return Object.keys(errs).filter((k) => errs[k]).length === 0
  }

  function costPreview() {
    const pos = positions.find((p) => p.code === values.position)
    if (!pos || !values.start_at || !values.end_at) return '—'
    const s = new Date(values.start_at)
    const e = new Date(values.end_at)
    if (Number.isNaN(s.getTime()) || Number.isNaN(e.getTime()) || e <= s) return '—'
    const days = Math.max(1, Math.ceil((e - s) / (1000 * 60 * 60 * 24)) + 1)
    const total = pos.price_per_day * days
    return `Rp ${formatAmount(total)} (${days} hari × Rp ${formatAmount(pos.price_per_day)}/hari)`
  }

  async function handleSave() {
    if (saving || uploading) return
    if (!validate()) { setFormError('Periksa kolom bertanda merah.'); return }
    setSaving(true)
    try {
      const payload = {
        title: values.title.trim(),
        position: values.position,
        start_at: fromLocal(values.start_at),
        end_at: fromLocal(values.end_at),
        image_url: values.image_url.trim(),
        target_url: values.target_type === 'url' ? values.target_url.trim() : '',
        target_event_code: values.target_type === 'event' ? values.target_event : '',
      }
      const { process_create, process_update } = await import('./api.js')
      if (isNew) await process_create(payload)
      else await process_update(editing.code, payload)
      toast.success(isNew ? 'Kampanye dibuat.' : 'Kampanye diperbarui.')
      onClose(); onSaved()
    } catch (ex) { setFormError(ex.message) } finally { setSaving(false) }
  }

  return (
    <Modal open={open} onClose={saving || uploading ? undefined : handleClose} title={isNew ? 'Create Campaign' : 'Ubah Campaign'} size="lg" closeOnBackdrop={!(saving || uploading)} footer={<><Button variant="secondary" onClick={handleClose} disabled={saving || uploading}>Batal</Button><Button onClick={handleSave} loading={saving || uploading}>{isNew ? 'Create' : 'Save'}</Button></>}>
      <div className="flex flex-col gap-4">
        {formError && <Alert tone="error" closable onClose={() => setFormError('')}>{formError}</Alert>}
        <label className="block">
          <span className="mb-1.5 block text-sm font-medium">Judul *</span>
          <input value={values.title} onChange={(e) => change({ ...values, title: e.target.value })} className={INPUT_CLASS} />
          {errors.title && <span className="mt-1 block text-xs text-rose-600">{errors.title}</span>}
        </label>
        <label className="block">
          <span className="mb-1.5 block text-sm font-medium">Gambar banner *</span>
          {preview && <SafeImage src={preview} alt={values.title} className="mb-2 h-32 w-full rounded-xl object-cover" />}
          <input type="file" accept="image/jpeg,image/png,image/webp,image/gif" onChange={(e) => handleBanner(e.target.files?.[0])} className="w-full text-sm file:mr-3 file:rounded-lg file:border file:border-zinc-300 file:bg-white file:px-3 file:py-1.5 file:text-sm file:font-semibold" />
          <span className="mt-1 block text-xs text-zinc-500">Validasi tipe, ukuran 2MB, rasio bebas. Upload via /api/marketing/ads/upload-banner.</span>
          {errors.image_url && <span className="mt-1 block text-xs text-rose-600">{errors.image_url}</span>}
          <input value={values.image_url} onChange={(e) => change({ ...values, image_url: e.target.value })} placeholder="atau tempel URL gambar" className={`${INPUT_CLASS} mt-2`} />
        </label>
        <div className="grid gap-4 sm:grid-cols-2">
          <label className="block">
            <span className="mb-1.5 block text-sm font-medium">Tautan tujuan</span>
            <select value={values.target_type} onChange={(e) => change({ ...values, target_type: e.target.value })} className={SELECT_CLASS}>
              <option value="url">URL luar</option>
              <option value="event">Event milik sendiri</option>
            </select>
          </label>
          {values.target_type === 'url' ? (
            <label className="block">
              <span className="mb-1.5 block text-sm font-medium">URL *</span>
              <input value={values.target_url} onChange={(e) => change({ ...values, target_url: e.target.value })} placeholder="https://..." className={INPUT_CLASS} />
              {errors.target_url && <span className="mt-1 block text-xs text-rose-600">{errors.target_url}</span>}
            </label>
          ) : (
            <label className="block">
              <span className="mb-1.5 block text-sm font-medium">Event *</span>
              <select value={values.target_event} onChange={(e) => change({ ...values, target_event: e.target.value })} className={SELECT_CLASS}>
                <option value="">— Pilih event —</option>
                {events.map((ev) => <option key={ev.code} value={ev.code}>{ev.title} ({ev.code})</option>)}
              </select>
              {errors.target_event && <span className="mt-1 block text-xs text-rose-600">{errors.target_event}</span>}
            </label>
          )}
        </div>
        <div className="grid gap-4 sm:grid-cols-2">
          <label className="block">
            <span className="mb-1.5 block text-sm font-medium">Posisi *</span>
            <select value={values.position} onChange={(e) => change({ ...values, position: e.target.value })} className={SELECT_CLASS}>
              <option value="">— Pilih posisi —</option>
              {positions.map((p) => <option key={p.code} value={p.code}>{p.name} — Rp {formatAmount(p.price_per_day)}/hari</option>)}
            </select>
            {errors.position && <span className="mt-1 block text-xs text-rose-600">{errors.position}</span>}
          </label>
          <div className="rounded-xl bg-zinc-50 p-3 text-sm dark:bg-zinc-800/50">
            <p className="text-xs text-zinc-500">Ringkasan biaya (backend hitung ulang)</p>
            <p className="font-mono font-semibold">{costPreview()}</p>
          </div>
        </div>
        <div className="grid gap-4 sm:grid-cols-2">
          <label className="block">
            <span className="mb-1.5 block text-sm font-medium">Mulai *</span>
            <input type="datetime-local" value={values.start_at} onChange={(e) => change({ ...values, start_at: e.target.value })} className={INPUT_CLASS} />
            {errors.start_at && <span className="mt-1 block text-xs text-rose-600">{errors.start_at}</span>}
          </label>
          <label className="block">
            <span className="mb-1.5 block text-sm font-medium">Selesai *</span>
            <input type="datetime-local" value={values.end_at} onChange={(e) => change({ ...values, end_at: e.target.value })} className={INPUT_CLASS} />
            {errors.end_at && <span className="mt-1 block text-xs text-rose-600">{errors.end_at}</span>}
          </label>
        </div>
        {preview && (
          <div>
            <p className="mb-2 text-xs font-bold uppercase tracking-wide text-zinc-500">Pratinjau seperti di Home</p>
            <div className="relative overflow-hidden rounded-xl border border-zinc-200 dark:border-zinc-700">
              <SafeImage src={preview} alt={values.title} className="h-36 w-full object-cover" />
              <span className="absolute left-2 top-2 rounded-full bg-black/60 px-2 py-0.5 text-[11px] font-bold text-white">Iklan</span>
            </div>
          </div>
        )}
      </div>
    </Modal>
  )
}
