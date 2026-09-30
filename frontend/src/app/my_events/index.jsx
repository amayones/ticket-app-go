// View My Events — daftar event milik penyelenggara yang login.
// Filter + pencarian + urutan + pagination di server; aksi baris di GRID.
// Controller: ./controller.js (6 fungsi). Store: read_data.
import { useEffect, useRef, useState } from 'react'
import {
  Alert,
  Button,
  StandardPage,
  useStandardController,
  SELECT_CLASS,
  goDetail,
} from '../shared/all.js'
import { eventsCategories, eventsCities } from '../events/api.js'
import { read_data } from './api.js'
import { controller } from './controller.js'
import GRID from './GRID.jsx'
import FRM from './FRM.jsx'

export const meta = { label: 'My Events', icon: 'folder', order: 1 }

const STATUSES = [
  { value: '', label: 'Semua status' },
  { value: 'DRAFT', label: 'Draft' },
  { value: 'PUBLISHED', label: 'Published' },
  { value: 'ENDED', label: 'Ended' },
  { value: 'CANCELLED', label: 'Cancelled' },
]

const SORTS = [
  { value: 'nearest', label: 'Terdekat' },
  { value: 'newest', label: 'Terbaru dibuat' },
  { value: 'popular', label: 'Terpopuler' },
]

function useDebounced(value, delay = 400) {
  const [applied, setApplied] = useState(value)
  useEffect(() => {
    const id = setTimeout(() => setApplied(value), delay)
    return () => clearTimeout(id)
  }, [value, delay])
  return applied
}

export default function MyEvents({ nvdata, onNavigate }) {
  const [status, setStatus] = useState('')
  const [category, setCategory] = useState('')
  const [city, setCity] = useState('')
  const [from, setFrom] = useState('')
  const [to, setTo] = useState('')
  const [query, setQuery] = useState('')
  const [sort, setSort] = useState('nearest')
  const [categories, setCategories] = useState([])
  const [cities, setCities] = useState([])
  const [showFRM, setShowFRM] = useState(null) // null | 'new' | kode event
  const gridRef = useRef(null)
  const appliedQuery = useDebounced(query)

  const params = { status, category, city, from, to, q: appliedQuery, sort }
  const { rows: events, offset, setOffset, loading, showLoading, error, setError, load } =
    useStandardController(read_data, params)

  useEffect(() => {
    controller.init()
    eventsCategories().then(setCategories).catch(() => setCategories([]))
    eventsCities().then(setCities).catch(() => setCities([]))
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  function handler_create_continue(code) {
    setShowFRM(null)
    goDetail(onNavigate, 'ticket_types', code)
    if (onNavigate) onNavigate('ticket_types')
  }

  const items = [{ events }]

  return (
    <div className="flex flex-col gap-4">
      <StandardPage
        title={nvdata?.label || 'My Events'}
        description="Event milik Anda. Kelola status, duplikat, tiket, dan peserta."
        loading={loading}
        showLoading={showLoading}
        error={error}
        errorTitle="Gagal memuat event"
        onClearError={() => setError('')}
        onRefresh={() => controller.btrefresh_click(load)}
        onCreate={() => controller.btnew_click(() => gridRef.current?.handler_btnew_click())}
        createLabel="Create Event"
        filterBar={
          <form onSubmit={(e) => e.preventDefault()} className="mb-4 flex flex-wrap items-end gap-2">
            <label className="flex flex-col gap-1 text-xs font-medium text-zinc-500">
              Status
              <select value={status} onChange={(e) => { setStatus(e.target.value); setOffset(0) }} className={SELECT_CLASS}>
                {STATUSES.map((s) => (
                  <option key={s.label} value={s.value}>
                    {s.label}
                  </option>
                ))}
              </select>
            </label>
            <label className="flex flex-col gap-1 text-xs font-medium text-zinc-500">
              Kategori
              <select value={category} onChange={(e) => { setCategory(e.target.value); setOffset(0) }} className={SELECT_CLASS}>
                <option value="">Semua kategori</option>
                {categories.map((c) => (
                  <option key={c.code} value={c.code}>
                    {c.name}
                  </option>
                ))}
              </select>
            </label>
            <label className="flex flex-col gap-1 text-xs font-medium text-zinc-500">
              Kota
              <select value={city} onChange={(e) => { setCity(e.target.value); setOffset(0) }} className={SELECT_CLASS}>
                <option value="">Semua kota</option>
                {cities.map((c) => (
                  <option key={c.code} value={c.code}>
                    {c.name}
                  </option>
                ))}
              </select>
            </label>
            <label className="flex flex-col gap-1 text-xs font-medium text-zinc-500">
              Dari tanggal
              <input
                type="date"
                value={from}
                onChange={(e) => { setFrom(e.target.value); setOffset(0) }}
                className={SELECT_CLASS}
              />
            </label>
            <label className="flex flex-col gap-1 text-xs font-medium text-zinc-500">
              Sampai tanggal
              <input
                type="date"
                value={to}
                onChange={(e) => { setTo(e.target.value); setOffset(0) }}
                className={SELECT_CLASS}
              />
            </label>
            <label className="flex min-w-[140px] flex-1 flex-col gap-1 text-xs font-medium text-zinc-500">
              Cari judul
              <input
                value={query}
                onChange={(e) => { setQuery(e.target.value); setOffset(0) }}
                placeholder="Ketik judul…"
                className={SELECT_CLASS}
              />
            </label>
            <label className="flex flex-col gap-1 text-xs font-medium text-zinc-500">
              Urutan
              <select value={sort} onChange={(e) => setSort(e.target.value)} className={SELECT_CLASS}>
                {SORTS.map((s) => (
                  <option key={s.value} value={s.value}>
                    {s.label}
                  </option>
                ))}
              </select>
            </label>
          </form>
        }
        items={items}
        emptyTitle="Belum ada event"
        emptyDescription="Buat event pertama Anda, lalu atur tiketnya."
        emptyAction={
          <Button size="sm" onClick={() => setShowFRM('new')}>
            Create Event
          </Button>
        }
        offset={offset}
        onPage={(next) => next >= 0 && setOffset(next)}
      >
        <GRID
          ref={gridRef}
          rows={events}
          reload={load}
          openFRM={() => setShowFRM('new')}
          openEdit={(code) => setShowFRM(code)}
          onOpenTypes={(code) => goDetail(onNavigate, 'ticket_types', code)}
          onOpenAttendees={(code) => goDetail(onNavigate, 'attendees', code)}
        />
      </StandardPage>

      <FRM
        open={!!showFRM}
        initial={showFRM === 'new' ? null : showFRM}
        categories={categories}
        cities={cities}
        onClose={() => setShowFRM(null)}
        onSaved={load}
        onContinue={handler_create_continue}
      />

      {error && (
        <Alert tone="error" title="Gagal memuat event" closable onClose={() => setError('')}>
          {error}
        </Alert>
      )}
    </div>
  )
}
