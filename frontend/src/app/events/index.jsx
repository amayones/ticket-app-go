// View Events — daftar event published yang belum berakhir (publik).
// Filter tersimpan di hash (#/events?...) supaya bisa dibagikan.
// Controller: ./controller.js (6 fungsi). Store: read_data.
import { useEffect, useState } from 'react'
import {
  Alert,
  Button,
  StandardPage,
  useDashboard,
  SELECT_CLASS,
  EventCard,
  hashParam,
  hashQuery,
  replaceHash,
} from '../shared/all.js'
import { read_data } from './api.js'
import { controller } from './controller.js'
import EventDetailModal from './components/EventDetailModal.jsx'

export const meta = { label: 'Events', icon: 'folder', order: 3 }

const SORTS = [
  { value: 'nearest', label: 'Terdekat' },
  { value: 'popular', label: 'Terpopuler' },
  { value: 'price_asc', label: 'Harga terendah' },
  { value: 'price_desc', label: 'Harga tertinggi' },
]

const PAGE_LIMIT = 12

function useDebounced(value, delay = 400) {
  const [applied, setApplied] = useState(value)
  useEffect(() => {
    const id = setTimeout(() => setApplied(value), delay)
    return () => clearTimeout(id)
  }, [value, delay])
  return applied
}

function initialFilters() {
  const q = hashQuery()
  return {
    categories: (q.get('categories') || '').split(',').map((s) => s.trim()).filter(Boolean),
    city: q.get('city') || '',
    from: q.get('from') || '',
    to: q.get('to') || '',
    min_price: q.get('min_price') || '',
    max_price: q.get('max_price') || '',
    q: q.get('q') || '',
    sort: q.get('sort') || 'nearest',
  }
}

export default function Events({ nvdata, onNavigate }) {
  const [init] = useState(initialFilters)
  const [categories, setCategories] = useState(init.categories)
  const [city, setCity] = useState(init.city)
  const [from, setFrom] = useState(init.from)
  const [to, setTo] = useState(init.to)
  const [minPrice, setMinPrice] = useState(init.min_price)
  const [maxPrice, setMaxPrice] = useState(init.max_price)
  const [query, setQuery] = useState(init.q)
  const [sort, setSort] = useState(init.sort)
  const [limit, setLimit] = useState(PAGE_LIMIT)
  const appliedQuery = useDebounced(query)
  // FRM detail event: kode dari hash (#/events/<code>) agar tautan modal
  // bisa dibagikan dan dibuka langsung.
  const [detailCode, setDetailCode] = useState(() => hashParam('events'))

  const filters = {
    categories: categories.join(','),
    city,
    from,
    to,
    min_price: minPrice,
    max_price: maxPrice,
    q: appliedQuery,
    sort,
    limit,
    offset: 0,
  }
  const { data, loading, showLoading, error, setError, load } = useDashboard(read_data, filters)

  useEffect(() => {
    controller.init()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  // Tulis filter ke hash (replace, tanpa menumpuk riwayat) agar bisa dibagikan.
  // Lewati saat modal detail terbuka (hash milik tautan modal).
  useEffect(() => {
    if (detailCode) return
    const q = new URLSearchParams()
    if (categories.length > 0) q.append('categories', categories.join(','))
    if (city) q.append('city', city)
    if (from) q.append('from', from)
    if (to) q.append('to', to)
    if (minPrice !== '') q.append('min_price', minPrice)
    if (maxPrice !== '') q.append('max_price', maxPrice)
    if (appliedQuery) q.append('q', appliedQuery)
    if (sort && sort !== 'nearest') q.append('sort', sort)
    const s = q.toString()
    replaceHash(`#/events${s ? `?${s}` : ''}`)
  }, [categories, city, from, to, minPrice, maxPrice, appliedQuery, sort, detailCode])

  // Tombol kembali browser: hash tanpa kode = tutup modal.
  useEffect(() => {
    function onHash() {
      const c = hashParam('events')
      setDetailCode((prev) => (prev === c ? prev : c))
    }
    window.addEventListener('hashchange', onHash)
    return () => window.removeEventListener('hashchange', onHash)
  }, [])

  function handler_open_event(code) {
    setDetailCode(code)
    replaceHash(`#/events/${encodeURIComponent(code)}`)
  }

  function handler_close_modal() {
    setDetailCode('')
  }

  function handler_toggle_category(code) {
    setCategories((prev) => (prev.includes(code) ? prev.filter((c) => c !== code) : [...prev, code]))
    setLimit(PAGE_LIMIT)
  }

  function handler_reset() {
    setCategories([])
    setCity('')
    setFrom('')
    setTo('')
    setMinPrice('')
    setMaxPrice('')
    setQuery('')
    setSort('nearest')
    setLimit(PAGE_LIMIT)
  }

  function handler_more() {
    setLimit((l) => l + PAGE_LIMIT)
  }

  const cats = data?.categories || []
  const cities = data?.cities || []
  const items = data?.items || []
  const pageItems = [{ items }]

  return (
    <div className="flex flex-col gap-4">
      <StandardPage
        title={nvdata?.label || 'Events'}
        description="Event yang sedang dibuka dan belum berakhir."
        loading={loading}
        showLoading={showLoading}
        error={error}
        errorTitle="Gagal memuat event"
        onClearError={() => setError('')}
        onRefresh={() => controller.btrefresh_click(load)}
        filterBar={
          <div className="mb-4 flex flex-col gap-2.5">
            <div className="flex flex-wrap gap-1.5">
              {cats.map((c) => (
                <Button
                  key={c.code}
                  size="sm"
                  variant={categories.includes(c.code) ? 'primary' : 'secondary'}
                  onClick={() => handler_toggle_category(c.code)}
                >
                  {c.name}
                </Button>
              ))}
            </div>
            <div className="flex flex-wrap items-end gap-2">
              <label className="flex flex-col gap-1 text-xs font-medium text-zinc-500">
                Kota
                <select
                  value={city}
                  onChange={(e) => {
                    setCity(e.target.value)
                    setLimit(PAGE_LIMIT)
                  }}
                  className={SELECT_CLASS}
                >
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
                  onChange={(e) => {
                    setFrom(e.target.value)
                    setLimit(PAGE_LIMIT)
                  }}
                  className={SELECT_CLASS}
                />
              </label>
              <label className="flex flex-col gap-1 text-xs font-medium text-zinc-500">
                Sampai tanggal
                <input
                  type="date"
                  value={to}
                  onChange={(e) => {
                    setTo(e.target.value)
                    setLimit(PAGE_LIMIT)
                  }}
                  className={SELECT_CLASS}
                />
              </label>
              <label className="flex flex-col gap-1 text-xs font-medium text-zinc-500">
                Harga min
                <input
                  type="number"
                  min="0"
                  value={minPrice}
                  onChange={(e) => {
                    setMinPrice(e.target.value)
                    setLimit(PAGE_LIMIT)
                  }}
                  placeholder="0"
                  className={`${SELECT_CLASS} w-28`}
                />
              </label>
              <label className="flex flex-col gap-1 text-xs font-medium text-zinc-500">
                Harga maks
                <input
                  type="number"
                  min="0"
                  value={maxPrice}
                  onChange={(e) => {
                    setMaxPrice(e.target.value)
                    setLimit(PAGE_LIMIT)
                  }}
                  placeholder="—"
                  className={`${SELECT_CLASS} w-28`}
                />
              </label>
              <label className="flex min-w-[160px] flex-1 flex-col gap-1 text-xs font-medium text-zinc-500">
                Cari event / venue
                <input
                  value={query}
                  onChange={(e) => {
                    setQuery(e.target.value)
                    setLimit(PAGE_LIMIT)
                  }}
                  placeholder="Ketik nama…"
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
              <Button size="sm" variant="ghost" onClick={handler_reset}>
                Reset
              </Button>
            </div>
          </div>
        }
        items={pageItems}
        emptyTitle={city ? 'Belum ada event di kota ini' : 'Belum ada event'}
        emptyDescription="Ubah filter atau kata kunci, lalu coba lagi."
        emptyAction={
          <Button variant="secondary" size="sm" onClick={handler_reset}>
            Reset filter
          </Button>
        }
      >
        <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
          {items.map((e) => (
            <EventCard key={e.code} event={e} onOpen={(item) => handler_open_event(item.code)} />
          ))}
        </div>
        {items.length >= limit && (
          <div className="mt-4 text-center">
            <Button variant="secondary" size="sm" onClick={handler_more} loading={loading}>
              Muat lagi
            </Button>
          </div>
        )}
      </StandardPage>

      {error && (
        <Alert tone="error" title="Gagal memuat event" closable onClose={() => setError('')}>
          {error}
        </Alert>
      )}

      {detailCode && (
        <EventDetailModal
          key={detailCode}
          code={detailCode}
          onClose={handler_close_modal}
          onOpenEvent={handler_open_event}
          onBuy={(eventCode) => {
            try {
              window.location.hash = `#/checkout/${encodeURIComponent(eventCode)}`
            } catch {
              // abaikan (non-browser)
            }
            if (onNavigate) onNavigate('checkout')
          }}
        />
      )}
    </div>
  )
}
