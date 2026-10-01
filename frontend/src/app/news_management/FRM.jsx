import { useEffect, useState } from 'react'
import { Alert, Button, Modal, SafeImage, useToast, INPUT_CLASS, SELECT_CLASS } from '../shared/all.js'
import { eventCategories, ownedEvents } from './api.js'

const EMPTY = {
  title: '',
  summary: '',
  body: '',
  image_url: '',
  news_category_code: '',
  event_category_code: '',
  event_code: '',
  scheduled_at: '',
}

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

export default function NewsModal({ open, editing, cats, onClose, onSaved }) {
  const toast = useToast()
  const isNew = !editing
  const [values, setValues] = useState(EMPTY)
  const [eCats, setECats] = useState([])
  const [events, setEvents] = useState([])
  const [errors, setErrors] = useState({})
  const [formError, setFormError] = useState('')
  const [saving, setSaving] = useState(false)
  const [dirty, setDirty] = useState(false)

  useEffect(() => {
    if (open) {
      eventCategories().then(setECats).catch(() => {})
      ownedEvents().then(setEvents).catch(() => {})
    }
  }, [open])

  useEffect(() => {
    if (!open) return
    setFormError(''); setErrors({}); setDirty(false)
    if (editing) {
      setValues({
        title: editing.title || '',
        summary: editing.summary || '',
        body: editing.body || '',
        image_url: editing.image_url || '',
        news_category_code: editing.news_category_code || '',
        event_category_code: editing.event_category_code || '',
        event_code: editing.event_code || '',
        scheduled_at: toLocal(editing.scheduled_at),
      })
    } else setValues(EMPTY)
  }, [open, editing])

  function change(next) { setValues(next); setDirty(true) }
  function handleClose() {
    if (dirty && !window.confirm('Form belum disimpan. Tutup?')) return
    onClose()
  }
  function validate() {
    const errs = {}
    if (values.title.trim().length < 3) errs.title = 'Judul minimal 3 karakter'
    else if (values.title.trim().length > 200) errs.title = 'Maks 200'
    if (!values.news_category_code) errs.news_category_code = 'Wajib'
    if (!values.body.trim()) errs.body = 'Isi wajib'
    if (values.summary.trim().length > 500) errs.summary = 'Maks 500'
    if (values.image_url.trim().length > 500) errs.image_url = 'URL terlalu panjang'
    if (values.scheduled_at && Number.isNaN(new Date(values.scheduled_at).getTime())) errs.scheduled_at = 'Format tidak valid'
    setErrors(errs)
    return Object.keys(errs).filter((k) => errs[k]).length === 0
  }
  async function handleSave() {
    if (saving) return
    if (!validate()) { setFormError('Periksa kolom bertanda merah.'); return }
    setSaving(true)
    try {
      const payload = {
        title: values.title.trim(),
        summary: values.summary.trim(),
        body: values.body,
        image_url: values.image_url.trim(),
        news_category_code: values.news_category_code,
        event_category_code: values.event_category_code || '',
        event_code: values.event_code || '',
        scheduled_at: values.scheduled_at ? fromLocal(values.scheduled_at) : '',
      }
      const { process_create, process_update } = await import('./api.js')
      if (isNew) await process_create(payload)
      else await process_update(editing.code, payload)
      toast.success(isNew ? 'Berita dibuat.' : 'Berita diperbarui.')
      onClose(); onSaved()
    } catch (ex) { setFormError(ex.message) } finally { setSaving(false) }
  }

  return (
    <Modal open={open} onClose={saving ? undefined : handleClose} title={isNew ? 'Create News' : 'Ubah Berita'} size="lg" closeOnBackdrop={!saving} footer={<><Button variant="secondary" onClick={handleClose} disabled={saving}>Batal</Button><Button onClick={handleSave} loading={saving}>{isNew ? 'Create' : 'Save'}</Button></>}>
      <div className="flex flex-col gap-4">
        {formError && <Alert tone="error" closable onClose={() => setFormError('')}>{formError}</Alert>}
        <label className="block">
          <span className="mb-1.5 block text-sm font-medium">Judul *</span>
          <input value={values.title} onChange={(e) => change({ ...values, title: e.target.value })} className={INPUT_CLASS} />
          {errors.title && <span className="mt-1 block text-xs text-rose-600">{errors.title}</span>}
        </label>
        <label className="block">
          <span className="mb-1.5 block text-sm font-medium">Ringkasan</span>
          <textarea value={values.summary} onChange={(e) => change({ ...values, summary: e.target.value })} rows={2} className={INPUT_CLASS} placeholder="Maks 500 karakter" />
          {errors.summary && <span className="mt-1 block text-xs text-rose-600">{errors.summary}</span>}
        </label>
        <label className="block">
          <span className="mb-1.5 block text-sm font-medium">Isi (HTML, disanitasi XSS) *</span>
          <textarea value={values.body} onChange={(e) => change({ ...values, body: e.target.value })} rows={8} className={INPUT_CLASS} placeholder="<p>Tulis berita…</p>" />
          {errors.body && <span className="mt-1 block text-xs text-rose-600">{errors.body}</span>}
        </label>
        <label className="block">
          <span className="mb-1.5 block text-sm font-medium">Gambar utama</span>
          {values.image_url && <SafeImage src={values.image_url} alt={values.title} className="mb-2 h-32 w-full rounded-xl object-cover" />}
          <input value={values.image_url} onChange={(e) => change({ ...values, image_url: e.target.value })} placeholder="https://..." className={INPUT_CLASS} />
          {errors.image_url && <span className="mt-1 block text-xs text-rose-600">{errors.image_url}</span>}
        </label>
        <div className="grid gap-4 sm:grid-cols-2">
          <label className="block">
            <span className="mb-1.5 block text-sm font-medium">Kategori berita *</span>
            <select value={values.news_category_code} onChange={(e) => change({ ...values, news_category_code: e.target.value })} className={SELECT_CLASS}>
              <option value="">— Pilih —</option>
              {cats.map((c) => <option key={c.code} value={c.code}>{c.name}</option>)}
            </select>
            {errors.news_category_code && <span className="mt-1 block text-xs text-rose-600">{errors.news_category_code}</span>}
          </label>
          <label className="block">
            <span className="mb-1.5 block text-sm font-medium">Kategori event terkait (opsional)</span>
            <select value={values.event_category_code} onChange={(e) => change({ ...values, event_category_code: e.target.value })} className={SELECT_CLASS}>
              <option value="">— Tidak ada —</option>
              {eCats.map((c) => <option key={c.code} value={c.code}>{c.name}</option>)}
            </select>
          </label>
        </div>
        <label className="block">
          <span className="mb-1.5 block text-sm font-medium">Event terkait (opsional)</span>
          <select value={values.event_code} onChange={(e) => change({ ...values, event_code: e.target.value })} className={SELECT_CLASS}>
            <option value="">— Tidak ada —</option>
            {events.map((ev) => <option key={ev.code} value={ev.code}>{ev.title} ({ev.code})</option>)}
          </select>
        </label>
        <label className="block">
          <span className="mb-1.5 block text-sm font-medium">Jadwal terbit (opsional)</span>
          <input type="datetime-local" value={values.scheduled_at} onChange={(e) => change({ ...values, scheduled_at: e.target.value })} className={INPUT_CLASS} />
          <span className="mt-1 block text-xs text-zinc-500">Kosong = draft. Isi = terbit otomatis pada waktunya (hanya DRAFT).</span>
        </label>
      </div>
    </Modal>
  )
}
