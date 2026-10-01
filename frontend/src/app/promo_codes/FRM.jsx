import { useEffect, useState } from 'react'
import { Alert, Button, Modal, useToast, INPUT_CLASS, SELECT_CLASS } from '../shared/all.js'

const EMPTY = {
  promo_code: '',
  description: '',
  discount_type: 'percent',
  discount_percent: '10',
  discount_amount: '',
  max_discount: '0',
  min_purchase: '0',
  max_uses: '',
  max_per_user: '',
  valid_from: '',
  valid_to: '',
  scope_type: 'ALL',
  category_codes: [],
  event_codes: [],
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

export default function PromoModal({ open, editing, opts, onClose, onSaved, onGenerate }) {
  const toast = useToast()
  const isNew = !editing
  const [values, setValues] = useState(EMPTY)
  const [errors, setErrors] = useState({})
  const [formError, setFormError] = useState('')
  const [saving, setSaving] = useState(false)
  const [dirty, setDirty] = useState(false)

  useEffect(() => {
    if (!open) return
    setFormError('')
    setErrors({})
    setDirty(false)
    if (editing) {
      setValues({
        promo_code: editing.promo_code || '',
        description: editing.description || '',
        discount_type: editing.discount_percent > 0 ? 'percent' : 'amount',
        discount_percent: String(editing.discount_percent || '10'),
        discount_amount: String(editing.discount_amount || ''),
        max_discount: String(editing.max_discount || '0'),
        min_purchase: String(editing.min_purchase || '0'),
        max_uses: editing.max_uses != null ? String(editing.max_uses) : '',
        max_per_user: editing.max_per_user != null ? String(editing.max_per_user) : '',
        valid_from: toLocal(editing.valid_from),
        valid_to: toLocal(editing.valid_to),
        scope_type: editing.scope_type || 'ALL',
        category_codes: editing.category_codes || [],
        event_codes: editing.event_codes || [],
      })
    } else {
      setValues(EMPTY)
    }
  }, [open, editing])

  useEffect(() => {
    if (!open || !dirty) return
    function onBefore(e) { e.preventDefault() }
    window.addEventListener('beforeunload', onBefore)
    return () => window.removeEventListener('beforeunload', onBefore)
  }, [open, dirty])

  function handler_change(next) { setValues(next); setDirty(true) }

  function handler_close() {
    if (dirty && !window.confirm('Form belum disimpan. Tutup?')) return
    onClose()
  }

  function handler_validasi_input() {
    const errs = {}
    if (!values.promo_code.trim()) errs.promo_code = 'Kode wajib diisi'
    else if (!/^[A-Z0-9_-]+$/.test(values.promo_code.trim())) errs.promo_code = 'Huruf besar / angka / _ / -'
    if (values.discount_type === 'percent') {
      const p = Number(values.discount_percent)
      if (!(p >= 1 && p <= 100)) errs.discount_percent = '1-100'
      if (Number(values.max_discount) < 0) errs.max_discount = 'Tidak negatif'
    } else if (!(Number(values.discount_amount) > 0)) errs.discount_amount = 'Harus > 0'
    if (Number(values.min_purchase) < 0) errs.min_purchase = 'Tidak negatif'
    if (values.max_uses !== '' && !(Number(values.max_uses) >= 1)) errs.max_uses = 'Minimal 1'
    if (values.max_per_user !== '' && !(Number(values.max_per_user) >= 1)) errs.max_per_user = 'Minimal 1'
    if (!values.valid_from) errs.valid_from = 'Wajib'
    if (!values.valid_to) errs.valid_to = 'Wajib'
    if (values.valid_from && values.valid_to && values.valid_to <= values.valid_from) errs.valid_to = 'Harus setelah mulai'
    if (values.scope_type === 'CATEGORY' && values.category_codes.length === 0) errs.category_codes = 'Pilih kategori'
    if (values.scope_type === 'EVENT' && values.event_codes.length === 0) errs.event_codes = 'Pilih event'
    setErrors(errs)
    return Object.keys(errs).filter((k) => errs[k]).length === 0
  }

  function buildPayload() {
    return {
      promo_code: values.promo_code.trim().toUpperCase(),
      description: values.description,
      discount_percent: values.discount_type === 'percent' ? Number(values.discount_percent) : 0,
      discount_amount: values.discount_type === 'amount' ? Number(values.discount_amount) : 0,
      max_discount: values.discount_type === 'percent' ? Number(values.max_discount) : 0,
      min_purchase: Number(values.min_purchase) || 0,
      max_uses: values.max_uses === '' ? null : Number(values.max_uses),
      max_per_user: values.max_per_user === '' ? null : Number(values.max_per_user),
      valid_from: fromLocal(values.valid_from),
      valid_to: fromLocal(values.valid_to),
      scope_type: values.scope_type,
      category_codes: values.category_codes,
      event_codes: values.event_codes,
    }
  }

  async function handler_generate() {
    try {
      const code = await onGenerate()
      if (code) handler_change({ ...values, promo_code: code })
    } catch (ex) { setFormError(ex.message) }
  }

  async function handler_btsave() {
    if (saving) return
    if (!handler_validasi_input()) { setFormError('Periksa kolom bertanda merah.'); return }
    setSaving(true)
    try {
      const payload = buildPayload()
      const { process_create, process_update } = await import('./api.js')
      if (isNew) await process_create(payload)
      else await process_update(editing.code, payload)
      toast.success(isNew ? 'Promo dibuat.' : 'Promo diperbarui.')
      onClose(); onSaved()
    } catch (ex) { setFormError(ex.message) } finally { setSaving(false) }
  }

  return (
    <Modal open={open} onClose={saving ? undefined : handler_close} title={isNew ? 'Create Promo' : 'Ubah Promo'} size="lg" closeOnBackdrop={!saving} footer={<><Button variant="secondary" onClick={handler_close} disabled={saving}>Batal</Button><Button onClick={handler_btsave} loading={saving}>{isNew ? 'Create' : 'Save'}</Button></>}>
      <div className="flex flex-col gap-4">
        {formError && <Alert tone="error" closable onClose={() => setFormError('')}>{formError}</Alert>}
        <label className="block">
          <span className="mb-1.5 block text-sm font-medium text-zinc-700 dark:text-zinc-200">Kode *</span>
          <span className="flex gap-2">
            <input value={values.promo_code} onChange={(e) => handler_change({ ...values, promo_code: e.target.value.toUpperCase() })} placeholder="HEMAT20" className={INPUT_CLASS} style={{ flex: 1 }} />
            <Button variant="secondary" size="sm" onClick={handler_generate}>Generate</Button>
          </span>
          {errors.promo_code && <span className="mt-1 block text-xs text-rose-600">{errors.promo_code}</span>}
        </label>
        <label className="block">
          <span className="mb-1.5 block text-sm font-medium text-zinc-700 dark:text-zinc-200">Deskripsi</span>
          <textarea value={values.description} onChange={(e) => handler_change({ ...values, description: e.target.value })} rows={2} className={INPUT_CLASS} />
        </label>
        <div className="grid gap-4 sm:grid-cols-2">
          <label className="block">
            <span className="mb-1.5 block text-sm font-medium">Jenis potongan</span>
            <select value={values.discount_type} onChange={(e) => handler_change({ ...values, discount_type: e.target.value })} className={SELECT_CLASS}>
              <option value="percent">Persen</option>
              <option value="amount">Nominal</option>
            </select>
          </label>
          {values.discount_type === 'percent' ? (
            <label className="block">
              <span className="mb-1.5 block text-sm font-medium">Persen (1-100) *</span>
              <input type="number" value={values.discount_percent} onChange={(e) => handler_change({ ...values, discount_percent: e.target.value })} className={INPUT_CLASS} />
              {errors.discount_percent && <span className="mt-1 block text-xs text-rose-600">{errors.discount_percent}</span>}
            </label>
          ) : (
            <label className="block">
              <span className="mb-1.5 block text-sm font-medium">Nominal *</span>
              <input type="number" value={values.discount_amount} onChange={(e) => handler_change({ ...values, discount_amount: e.target.value })} className={INPUT_CLASS} />
              {errors.discount_amount && <span className="mt-1 block text-xs text-rose-600">{errors.discount_amount}</span>}
            </label>
          )}
        </div>
        {values.discount_type === 'percent' && (
          <label className="block">
            <span className="mb-1.5 block text-sm font-medium">Maks potongan (untuk persen)</span>
            <input type="number" value={values.max_discount} onChange={(e) => handler_change({ ...values, max_discount: e.target.value })} className={INPUT_CLASS} />
            {errors.max_discount && <span className="mt-1 block text-xs text-rose-600">{errors.max_discount}</span>}
          </label>
        )}
        <div className="grid gap-4 sm:grid-cols-2">
          <label className="block">
            <span className="mb-1.5 block text-sm font-medium">Min. pembelian</span>
            <input type="number" value={values.min_purchase} onChange={(e) => handler_change({ ...values, min_purchase: e.target.value })} className={INPUT_CLASS} />
          </label>
          <label className="block">
            <span className="mb-1.5 block text-sm font-medium">Kuota total</span>
            <input type="number" value={values.max_uses} onChange={(e) => handler_change({ ...values, max_uses: e.target.value })} placeholder="Kosong = tanpa batas" className={INPUT_CLASS} />
            {errors.max_uses && <span className="mt-1 block text-xs text-rose-600">{errors.max_uses}</span>}
          </label>
        </div>
        <label className="block">
          <span className="mb-1.5 block text-sm font-medium">Batas per pengguna</span>
          <input type="number" value={values.max_per_user} onChange={(e) => handler_change({ ...values, max_per_user: e.target.value })} placeholder="Kosong = tanpa batas" className={INPUT_CLASS} />
          {errors.max_per_user && <span className="mt-1 block text-xs text-rose-600">{errors.max_per_user}</span>}
        </label>
        <div className="grid gap-4 sm:grid-cols-2">
          <label className="block">
            <span className="mb-1.5 block text-sm font-medium">Mulai *</span>
            <input type="datetime-local" value={values.valid_from} onChange={(e) => handler_change({ ...values, valid_from: e.target.value })} className={INPUT_CLASS} />
            {errors.valid_from && <span className="mt-1 block text-xs text-rose-600">{errors.valid_from}</span>}
          </label>
          <label className="block">
            <span className="mb-1.5 block text-sm font-medium">Selesai *</span>
            <input type="datetime-local" value={values.valid_to} onChange={(e) => handler_change({ ...values, valid_to: e.target.value })} className={INPUT_CLASS} />
            {errors.valid_to && <span className="mt-1 block text-xs text-rose-600">{errors.valid_to}</span>}
          </label>
        </div>
        <label className="block">
          <span className="mb-1.5 block text-sm font-medium">Cakupan</span>
          <select value={values.scope_type} onChange={(e) => handler_change({ ...values, scope_type: e.target.value })} className={SELECT_CLASS}>
            <option value="ALL">Semua event milik saya</option>
            <option value="CATEGORY">Kategori tertentu</option>
            <option value="EVENT">Event tertentu</option>
          </select>
        </label>
        {values.scope_type === 'CATEGORY' && (
          <div>
            <span className="mb-1.5 block text-sm font-medium">Kategori</span>
            <div className="flex flex-wrap gap-2">
              {(opts.categories || []).map((c) => (
                <label key={c.code} className="flex items-center gap-1.5 text-sm">
                  <input type="checkbox" checked={values.category_codes.includes(c.code)} onChange={(e) => handler_change({ ...values, category_codes: e.target.checked ? [...values.category_codes, c.code] : values.category_codes.filter((x) => x !== c.code) })} />
                  {c.name}
                </label>
              ))}
            </div>
            {errors.category_codes && <span className="mt-1 block text-xs text-rose-600">{errors.category_codes}</span>}
          </div>
        )}
        {values.scope_type === 'EVENT' && (
          <div>
            <span className="mb-1.5 block text-sm font-medium">Event</span>
            <div className="flex max-h-40 flex-col gap-1 overflow-y-auto rounded-xl border border-zinc-200 p-3 dark:border-zinc-700">
              {(opts.events || []).map((ev) => (
                <label key={ev.code} className="flex items-center gap-1.5 text-sm">
                  <input type="checkbox" checked={values.event_codes.includes(ev.code)} onChange={(e) => handler_change({ ...values, event_codes: e.target.checked ? [...values.event_codes, ev.code] : values.event_codes.filter((x) => x !== ev.code) })} />
                  {ev.title} <span className="text-xs text-zinc-500">({ev.code})</span>
                </label>
              ))}
              {(opts.events || []).length === 0 && <span className="text-sm text-zinc-500">Belum ada event.</span>}
            </div>
            {errors.event_codes && <span className="mt-1 block text-xs text-rose-600">{errors.event_codes}</span>}
          </div>
        )}
      </div>
    </Modal>
  )
}
