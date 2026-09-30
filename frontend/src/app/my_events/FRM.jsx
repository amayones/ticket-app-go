// FRM Event — padanan FRM<mod>.js langit_v2 (window form + tbar save).
// Satu form dua mode: initial null = baru, initial kode = ubah.
// Dipakai untuk New Input (btnew Varian A delegasi ke GRID.handler_btnew_click
// yang membuka FRM ini) dan tombol Ubah (openEdit -> FRM edit).
import { useEffect, useState } from 'react'
import { Alert, Button, Modal, SafeImage, StandardForm, useToast, INPUT_CLASS } from '../shared/all.js'
import { getEvent, process_create, process_update, uploadPoster } from './api.js'
import { eventBasicFields } from './fields.js'

const POSTER_TYPES = ['image/jpeg', 'image/png', 'image/webp', 'image/gif']
const POSTER_MAX = 2 * 1024 * 1024

function toLocalInput(iso) {
  if (!iso) return ''
  const d = new Date(String(iso).replace(' ', 'T'))
  if (Number.isNaN(d.getTime())) return ''
  const p = (n) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${p(d.getMonth() + 1)}-${p(d.getDate())}T${p(d.getHours())}:${p(d.getMinutes())}`
}

function fromLocalInput(value) {
  if (!value) return ''
  return `${value.slice(0, 10)} ${value.slice(11, 16)}`
}

const EMPTY = {
  title: '',
  category_code: '',
  description: '',
  city_code: '',
  venue: '',
  address: '',
  latitude: '',
  longitude: '',
  start_at: '',
  end_at: '',
  poster_url: '',
  terms: '',
}

export default function FRM({ open, initial, categories, cities, onClose, onSaved, onContinue }) {
  const toast = useToast()
  const isNew = !initial
  const [values, setValues] = useState(EMPTY)
  const [status, setStatus] = useState('')
  const [missing, setMissing] = useState([])
  const [errors, setErrors] = useState({})
  const [formError, setFormError] = useState('')
  const [saving, setSaving] = useState(false)
  const [dirty, setDirty] = useState(false)
  const [posterFile, setPosterFile] = useState(null)
  const [posterPreview, setPosterPreview] = useState('')
  const [uploading, setUploading] = useState(false)

  useEffect(() => {
    if (open) {
      setValues(EMPTY)
      setStatus('')
      setMissing([])
      setErrors({})
      setFormError('')
      setDirty(false)
      setPosterFile(null)
      setPosterPreview('')
      if (initial) {
        getEvent(initial)
          .then((e) => {
            setValues({
              title: e.title || '',
              category_code: e.category_code || '',
              description: e.description || '',
              city_code: e.city_code || '',
              venue: e.venue || '',
              address: e.address || '',
              latitude: e.latitude ? String(e.latitude) : '',
              longitude: e.longitude ? String(e.longitude) : '',
              start_at: toLocalInput(e.start_at),
              end_at: toLocalInput(e.end_at),
              poster_url: e.poster_url || '',
              terms: e.terms || '',
            })
            setPosterPreview(e.poster_url || '')
            setStatus(e.status || '')
            setMissing(e.missing || [])
          })
          .catch((ex) => setFormError(ex.message))
      }
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [open])

  // Cegah kehilangan data: peringatan browser saat tab ditutup/refresh.
  useEffect(() => {
    if (!open || !dirty) return undefined
    function onBefore(e) {
      e.preventDefault()
    }
    window.addEventListener('beforeunload', onBefore)
    return () => window.removeEventListener('beforeunload', onBefore)
  }, [open, dirty])

  function handler_change(next) {
    setValues(next)
    setDirty(true)
  }

  function handler_close() {
    if (dirty && !window.confirm('Form belum disimpan. Tutup?')) return
    onClose()
  }

  function handler_poster_change(file) {
    if (!file) {
      setPosterFile(null)
      return
    }
    if (!POSTER_TYPES.includes(file.type)) {
      setErrors((p) => ({ ...p, poster_url: 'Tipe file harus JPG, PNG, WebP, atau GIF' }))
      return
    }
    if (file.size > POSTER_MAX) {
      setErrors((p) => ({ ...p, poster_url: 'Ukuran maksimal 2 MB' }))
      return
    }
    setErrors((p) => ({ ...p, poster_url: '' }))
    setPosterFile(file)
    setDirty(true)
    try {
      const reader = new FileReader()
      reader.onload = () => setPosterPreview(String(reader.result || ''))
      reader.readAsDataURL(file)
    } catch {
      setPosterPreview('')
    }
  }

  function handler_validasi_input() {
    const errs = {}
    if (values.title.trim().length < 3) errs.title = 'Judul minimal 3 karakter'
    else if (values.title.trim().length > 200) errs.title = 'Judul maksimal 200 karakter'
    if (!values.category_code) errs.category_code = 'Kategori wajib dipilih'
    if (!values.city_code) errs.city_code = 'Kota wajib dipilih'
    if (!values.venue.trim()) errs.venue = 'Venue wajib diisi'
    if (!values.start_at) errs.start_at = 'Tanggal mulai wajib diisi'
    if (values.start_at && values.end_at && values.end_at <= values.start_at) {
      errs.end_at = 'Tanggal selesai harus setelah mulai'
    }
    if (values.latitude !== '' && (Number.isNaN(Number(values.latitude)) || Number(values.latitude) < -90 || Number(values.latitude) > 90)) {
      errs.latitude = 'Latitude harus -90 sampai 90'
    }
    if (values.longitude !== '' && (Number.isNaN(Number(values.longitude)) || Number(values.longitude) < -180 || Number(values.longitude) > 180)) {
      errs.longitude = 'Longitude harus -180 sampai 180'
    }
    setErrors(errs)
    return Object.keys(errs).filter((k) => errs[k]).length === 0
  }

  function buildPayload(posterUrl) {
    return {
      title: values.title.trim(),
      category_code: values.category_code,
      description: values.description,
      city_code: values.city_code,
      venue: values.venue.trim(),
      address: values.address.trim(),
      latitude: values.latitude === '' ? 0 : Number(values.latitude),
      has_latitude: values.latitude !== '',
      longitude: values.longitude === '' ? 0 : Number(values.longitude),
      has_longitude: values.longitude !== '',
      start_at: fromLocalInput(values.start_at),
      end_at: values.end_at ? fromLocalInput(values.end_at) : '',
      poster_url: posterUrl,
      terms: values.terms,
    }
  }

  async function handler_btsave(goTypes) {
    try {
      if (saving || uploading) return false
      if (handler_validasi_input() === false) {
        setFormError('Periksa kolom bertanda merah.')
        return false
      }
      setSaving(true)
      let posterUrl = values.poster_url
      if (posterFile) {
        setUploading(true)
        try {
          const res = await uploadPoster(posterFile)
          posterUrl = res.url || posterUrl
        } finally {
          setUploading(false)
        }
      }
      if (isNew) {
        const res = await process_create(buildPayload(posterUrl))
        toast.success('Event dibuat sebagai draft.', { title: 'Event dibuat' })
        onClose()
        onSaved()
        if (goTypes && res && res.code && onContinue) onContinue(res.code)
      } else {
        const res = await process_update(initial, buildPayload(posterUrl))
        if (res && res.warning) toast.warning(res.warning, { title: 'Perhatian' })
        toast.success('Event diperbarui.')
        onClose()
        onSaved()
        if (goTypes && onContinue) onContinue(initial)
      }
    } catch (ex) {
      setFormError(ex.message)
    } finally {
      setSaving(false)
    }
  }

  const fields = eventBasicFields(categories, cities)

  return (
    <Modal
      open={open}
      onClose={saving || uploading ? undefined : handler_close}
      title={isNew ? 'Create Event' : 'Ubah Event'}
      size="lg"
      closeOnBackdrop={!(saving || uploading)}
      footer={
        <>
          <Button variant="secondary" onClick={handler_close} disabled={saving || uploading}>
            Batal
          </Button>
          <Button variant="secondary" onClick={() => handler_btsave(true)} loading={saving || uploading}>
            Save & Lanjut Tiket
          </Button>
          <Button onClick={() => handler_btsave(false)} loading={saving || uploading}>
            {isNew ? 'Save as Draft' : 'Save'}
          </Button>
        </>
      }
    >
      <div className="flex flex-col gap-5">
        {formError && (
          <Alert tone="error" closable onClose={() => setFormError('')}>
            {formError}
          </Alert>
        )}
        {status ? (
          <p className="text-xs font-semibold text-zinc-500">
            Status: {status}
            {missing.length > 0 ? ` · kurang: ${missing.join('; ')}` : ''}
          </p>
        ) : null}

        <section>
          <h3 className="mb-3 text-sm font-bold uppercase tracking-wide text-zinc-500">Informasi dasar</h3>
          <StandardForm fields={fields.slice(0, 3)} values={values} onChange={handler_change} />
          {(errors.title || errors.category_code) && (
            <p className="mt-1.5 text-xs font-medium text-rose-600 dark:text-rose-400">
              {[errors.title, errors.category_code].filter(Boolean).join(' · ')}
            </p>
          )}
        </section>

        <section>
          <h3 className="mb-3 text-sm font-bold uppercase tracking-wide text-zinc-500">Lokasi</h3>
          <StandardForm fields={fields.slice(3, 6)} values={values} onChange={handler_change} />
          {(errors.city_code || errors.venue) && (
            <p className="mt-1.5 text-xs font-medium text-rose-600 dark:text-rose-400">
              {[errors.city_code, errors.venue].filter(Boolean).join(' · ')}
            </p>
          )}
          <div className="mt-4 grid gap-4 sm:grid-cols-2">
            <label className="block">
              <span className="mb-1.5 block text-sm font-medium text-zinc-700 dark:text-zinc-200">Latitude (opsional)</span>
              <input
                value={values.latitude}
                onChange={(e) => handler_change({ ...values, latitude: e.target.value })}
                placeholder="-6.2"
                inputMode="decimal"
                className={INPUT_CLASS}
              />
              {errors.latitude && <span className="mt-1.5 block text-xs font-medium text-rose-600 dark:text-rose-400">{errors.latitude}</span>}
            </label>
            <label className="block">
              <span className="mb-1.5 block text-sm font-medium text-zinc-700 dark:text-zinc-200">Longitude (opsional)</span>
              <input
                value={values.longitude}
                onChange={(e) => handler_change({ ...values, longitude: e.target.value })}
                placeholder="106.8"
                inputMode="decimal"
                className={INPUT_CLASS}
              />
              {errors.longitude && <span className="mt-1.5 block text-xs font-medium text-rose-600 dark:text-rose-400">{errors.longitude}</span>}
            </label>
          </div>
        </section>

        <section>
          <h3 className="mb-3 text-sm font-bold uppercase tracking-wide text-zinc-500">Jadwal</h3>
          <div className="grid gap-4 sm:grid-cols-2">
            <label className="block">
              <span className="mb-1.5 block text-sm font-medium text-zinc-700 dark:text-zinc-200">Mulai</span>
              <input
                type="datetime-local"
                value={values.start_at}
                onChange={(e) => handler_change({ ...values, start_at: e.target.value })}
                className={INPUT_CLASS}
              />
              {errors.start_at && <span className="mt-1.5 block text-xs font-medium text-rose-600 dark:text-rose-400">{errors.start_at}</span>}
            </label>
            <label className="block">
              <span className="mb-1.5 block text-sm font-medium text-zinc-700 dark:text-zinc-200">Selesai</span>
              <input
                type="datetime-local"
                value={values.end_at}
                onChange={(e) => handler_change({ ...values, end_at: e.target.value })}
                className={INPUT_CLASS}
              />
              {errors.end_at && <span className="mt-1.5 block text-xs font-medium text-rose-600 dark:text-rose-400">{errors.end_at}</span>}
            </label>
          </div>
        </section>

        <section>
          <h3 className="mb-3 text-sm font-bold uppercase tracking-wide text-zinc-500">Media</h3>
          {(posterPreview || values.poster_url) && (
            <SafeImage src={posterPreview || values.poster_url} alt={values.title} className="mb-3 w-full max-w-md rounded-xl object-cover" />
          )}
          <label className="block max-w-md">
            <span className="mb-1.5 block text-sm font-medium text-zinc-700 dark:text-zinc-200">Poster/banner (JPG/PNG/WebP/GIF, maks 2 MB)</span>
            <input
              type="file"
              accept="image/jpeg,image/png,image/webp,image/gif"
              onChange={(e) => handler_poster_change(e.target.files && e.target.files[0])}
              className="w-full text-sm text-zinc-600 file:mr-3 file:rounded-lg file:border file:border-zinc-300 file:bg-white file:px-3 file:py-1.5 file:text-sm file:font-semibold file:text-zinc-700 hover:file:bg-zinc-50 dark:text-zinc-300 dark:file:border-zinc-700 dark:file:bg-zinc-900 dark:file:text-zinc-200"
            />
            {errors.poster_url && <span className="mt-1.5 block text-xs font-medium text-rose-600 dark:text-rose-400">{errors.poster_url}</span>}
          </label>
        </section>

        <section>
          <h3 className="mb-3 text-sm font-bold uppercase tracking-wide text-zinc-500">Syarat & ketentuan</h3>
          <StandardForm fields={fields.slice(6)} values={values} onChange={handler_change} />
        </section>
      </div>
    </Modal>
  )
}
