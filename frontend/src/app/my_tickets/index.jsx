// View My Tickets — tiket milik pembeli yang login (filter status,
// cari, pagination). Klik kartu = modal detail (QR + refund).
// Controller: ./controller.js (6 fungsi). Store: read_data.
import { useEffect, useState } from 'react'
import {
  Alert,
  Badge,
  Button,
  StandardPage,
  formatTime,
  useDashboard,
  SELECT_CLASS,
} from '../shared/all.js'
import { read_data } from './api.js'
import { controller } from './controller.js'
import TicketDetailModal from './components/TicketDetailModal.jsx'

export const meta = { label: 'My Tickets', icon: 'list', order: 2 }

const FILTERS = [
  { value: '', label: 'Aktif' },
  { value: 'used', label: 'Sudah Dipakai' },
  { value: 'refunded', label: 'Refund' },
  { value: 'expired', label: 'Kedaluwarsa' },
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

function statusTone(status) {
  switch (status) {
    case 'ACTIVE':
      return 'success'
    case 'USED':
      return 'neutral'
    case 'REFUNDED':
      return 'warning'
    default:
      return 'neutral'
  }
}

export default function MyTickets({ nvdata, onNavigate }) {
  const [status, setStatus] = useState('')
  const [query, setQuery] = useState('')
  const [limit, setLimit] = useState(PAGE_LIMIT)
  const [detailCode, setDetailCode] = useState('')
  const appliedQuery = useDebounced(query)

  const filters = { status, q: appliedQuery, limit, offset: 0 }
  const { data, loading, showLoading, error, setError, load } = useDashboard(read_data, filters)

  useEffect(() => {
    controller.init()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  const items = data || []
  const pageItems = [{ items }]

  return (
    <div className="flex flex-col gap-4">
      <StandardPage
        title={nvdata?.label || 'My Tickets'}
        description="Tiket milik Anda, event terdekat lebih dulu."
        loading={loading}
        showLoading={showLoading}
        error={error}
        errorTitle="Gagal memuat tiket"
        onClearError={() => setError('')}
        onRefresh={() => controller.btrefresh_click(load)}
        filterBar={
          <form onSubmit={(e) => e.preventDefault()} className="mb-4 flex flex-wrap items-end gap-2">
            <span className="flex flex-wrap gap-1.5">
              {FILTERS.map((f) => (
                <Button
                  key={f.label}
                  size="sm"
                  variant={status === f.value ? 'primary' : 'ghost'}
                  onClick={() => {
                    setStatus(f.value)
                    setLimit(PAGE_LIMIT)
                  }}
                >
                  {f.label}
                </Button>
              ))}
            </span>
            <label className="flex min-w-[160px] flex-1 flex-col gap-1 text-xs font-medium text-zinc-500">
              Cari event/tipe/kode
              <input
                value={query}
                onChange={(e) => {
                  setQuery(e.target.value)
                  setLimit(PAGE_LIMIT)
                }}
                placeholder="Ketik…"
                className={SELECT_CLASS}
              />
            </label>
          </form>
        }
        items={pageItems}
        emptyTitle="Belum ada tiket"
        emptyDescription={
          status === ''
            ? 'Beli tiket event favorit Anda.'
            : 'Tidak ada tiket pada filter ini.'
        }
        emptyAction={
          <Button size="sm" onClick={() => onNavigate && onNavigate('events')}>
            Ke Events
          </Button>
        }
      >
        <div className="grid gap-3 sm:grid-cols-2">
          {items.map((t) => (
            <button
              key={t.ticket_code}
              type="button"
              onClick={() => setDetailCode(t.ticket_code)}
              className="flex min-w-0 flex-col gap-1 rounded-2xl border border-zinc-200 bg-white p-4 text-left shadow-sm transition hover:border-violet-300 hover:shadow-md dark:border-zinc-800 dark:bg-zinc-900"
            >
              <span className="flex items-center justify-between gap-2">
                <span className="truncate text-sm font-semibold text-zinc-900 dark:text-zinc-50">{t.event_title}</span>
                <Badge tone={statusTone(t.status)}>{t.status}</Badge>
              </span>
              <span className="truncate text-xs text-zinc-500 dark:text-zinc-400">
                {[t.type_name, t.venue].filter(Boolean).join(' · ')}
              </span>
              <span className="truncate text-xs text-zinc-500 dark:text-zinc-400">{formatTime(t.start_at)}</span>
              <span className="font-mono text-[11px] text-zinc-400">{t.ticket_code}</span>
            </button>
          ))}
        </div>
        {items.length >= limit && (
          <div className="mt-4 text-center">
            <Button variant="secondary" size="sm" onClick={() => setLimit((l) => l + PAGE_LIMIT)} loading={loading}>
              Muat lagi
            </Button>
          </div>
        )}
      </StandardPage>

      {detailCode && (
        <TicketDetailModal key={detailCode} code={detailCode} onClose={() => setDetailCode('')} onRefunded={load} />
      )}

      {error && (
        <Alert tone="error" title="Gagal memuat tiket" closable onClose={() => setError('')}>
          {error}
        </Alert>
      )}
    </div>
  )
}
