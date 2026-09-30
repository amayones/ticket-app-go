// View Promotions — promo berlaku + segera hadir (publik, tanpa login).
// Discovery hanya menampilkan; pembuatan dan validasi di module lain.
// Controller: ./controller.js (6 fungsi). Store: read_data.
import { useEffect, useState } from 'react'
import {
  Alert,
  Button,
  Card,
  CardTitle,
  StandardPage,
  useDashboard,
  useToast,
  SELECT_CLASS,
  PromoCard,
} from '../shared/all.js'
import { read_data } from './api.js'
import { controller } from './controller.js'

export const meta = { label: 'Promotions', icon: 'bell', order: 5 }

const MODES = [
  { value: 'active', label: 'Berlaku' },
  { value: 'upcoming', label: 'Segera hadir' },
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

export default function Promotions({ nvdata }) {
  const toast = useToast()
  const [mode, setMode] = useState('active')
  const [category, setCategory] = useState('')
  const [query, setQuery] = useState('')
  const [limit, setLimit] = useState(PAGE_LIMIT)
  const appliedQuery = useDebounced(query)

  const filters = { mode, category, q: appliedQuery, limit, offset: 0 }
  const { data, loading, showLoading, error, setError, load } = useDashboard(read_data, filters)

  useEffect(() => {
    controller.init()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  function handler_copy_code(promo) {
    const done = () => toast.success(`Kode ${promo.promo_code} disalin.`)
    try {
      if (navigator.clipboard && navigator.clipboard.writeText) {
        navigator.clipboard.writeText(promo.promo_code).then(done).catch(() => toast.info(promo.promo_code, { title: 'Kode promo' }))
      } else {
        toast.info(promo.promo_code, { title: 'Kode promo' })
      }
    } catch {
      toast.info(promo.promo_code, { title: 'Kode promo' })
    }
  }

  function handler_more() {
    setLimit((l) => l + PAGE_LIMIT)
  }

  const categories = data?.categories || []
  const items = data?.items || []
  const pageItems = [{ items }]

  return (
    <div className="flex flex-col gap-4">
      <StandardPage
        title={nvdata?.label || 'Promotions'}
        description="Kode promo yang bisa dipakai. Berlaku sesuai syarat."
        loading={loading}
        showLoading={showLoading}
        error={error}
        errorTitle="Gagal memuat promo"
        onClearError={() => setError('')}
        onRefresh={() => controller.btrefresh_click(load)}
        filterBar={
          <form onSubmit={(e) => e.preventDefault()} className="mb-4 flex flex-wrap items-end gap-2">
            <span className="flex flex-wrap gap-1.5">
              {MODES.map((m) => (
                <Button
                  key={m.value}
                  size="sm"
                  variant={mode === m.value ? 'primary' : 'ghost'}
                  onClick={() => {
                    setMode(m.value)
                    setLimit(PAGE_LIMIT)
                  }}
                >
                  {m.label}
                </Button>
              ))}
            </span>
            <label className="flex flex-col gap-1 text-xs font-medium text-zinc-500">
              Kategori event
              <select
                value={category}
                onChange={(e) => {
                  setCategory(e.target.value)
                  setLimit(PAGE_LIMIT)
                }}
                className={SELECT_CLASS}
              >
                <option value="">Semua kategori</option>
                {categories.map((c) => (
                  <option key={c.code} value={c.code}>
                    {c.name}
                  </option>
                ))}
              </select>
            </label>
            <label className="flex min-w-[160px] flex-1 flex-col gap-1 text-xs font-medium text-zinc-500">
              Cari kode
              <input
                value={query}
                onChange={(e) => {
                  setQuery(e.target.value)
                  setLimit(PAGE_LIMIT)
                }}
                placeholder="Ketik kode…"
                className={SELECT_CLASS}
              />
            </label>
          </form>
        }
        items={pageItems}
        emptyTitle={mode === 'upcoming' ? 'Belum ada promo segera hadir' : 'Belum ada promo'}
        emptyDescription="Coba kategori atau kata kunci lain."
        emptyAction={
          <Button variant="secondary" size="sm" onClick={() => controller.btrefresh_click(load)}>
            Muat ulang
          </Button>
        }
      >
        <div className="grid gap-3 sm:grid-cols-2">
          {items.map((p) => (
            <PromoCard key={p.code} promo={p} onCopy={handler_copy_code} />
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

      <Card>
        <CardTitle description="Promo dipakai saat pembelian di module ticketing.">Catatan pemakaian</CardTitle>
        <p className="text-sm text-zinc-600 dark:text-zinc-300">
          Halaman ini hanya menampilkan promo. Pembuatan promo dan validasi pemakaiannya dikerjakan di module lain.
        </p>
      </Card>

      {error && (
        <Alert tone="error" title="Gagal memuat promo" closable onClose={() => setError('')}>
          {error}
        </Alert>
      )}
    </div>
  )
}
