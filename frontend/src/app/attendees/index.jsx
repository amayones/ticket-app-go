// View Attendees — daftar pembeli tiket satu event (read-only penuh).
// Masuk dari My Events (hash #/attendees/<eventCode>).
// Controller: ./controller.js (6 fungsi). Store: read_data.
import { useEffect, useState } from 'react'
import {
  Alert,
  Badge,
  Button,
  Card,
  CardTitle,
  Icon,
  Skeleton,
  StandardGrid,
  StandardPage,
  formatTime,
  useDashboard,
  useToast,
  SELECT_CLASS,
  clearHash,
  hashParam,
} from '../shared/all.js'
import { exportAttendees, read_data } from './api.js'
import { controller } from './controller.js'
import AttendeeModal from './components/AttendeeModal.jsx'

export const meta = { label: 'Attendees', icon: 'users', order: 4 }

const PAGE_LIMIT = 20

function useDebounced(value, delay = 400) {
  const [applied, setApplied] = useState(value)
  useEffect(() => {
    const id = setTimeout(() => setApplied(value), delay)
    return () => clearTimeout(id)
  }, [value, delay])
  return applied
}

export default function Attendees({ nvdata, onNavigate }) {
  const toast = useToast()
  const [eventCode] = useState(() => hashParam('attendees'))
  const [type, setType] = useState('')
  const [attendance, setAttendance] = useState('')
  const [payment, setPayment] = useState('')
  const [query, setQuery] = useState('')
  const [limit, setLimit] = useState(PAGE_LIMIT)
  const [detail, setDetail] = useState(null)
  const appliedQuery = useDebounced(query)

  const filters = { eventCode, type, attendance, payment, q: appliedQuery, limit, offset: 0 }
  const { data, loading, showLoading, error, setError, load } = useDashboard(read_data, filters)

  useEffect(() => {
    controller.init()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  function handler_back() {
    clearHash()
    if (onNavigate) onNavigate('my_events')
  }

  async function handler_export_click() {
    try {
      const blob = await exportAttendees({ eventCode, type, attendance, payment, q: appliedQuery })
      const url = URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = url
      a.download = `attendees-${eventCode}.csv`
      a.click()
      URL.revokeObjectURL(url)
      toast.success('Daftar peserta diunduh.')
    } catch (ex) {
      toast.error(ex.message, { title: 'Export gagal' })
    }
  }

  if (!eventCode) {
    return (
      <Card>
        <CardTitle description="Pilih event dulu dari My Events.">Tidak ada event dipilih</CardTitle>
        <Button variant="secondary" size="sm" onClick={handler_back}>
          Ke My Events
        </Button>
      </Card>
    )
  }

  const rows = data?.attendees || []
  const sum = data?.summary || {}
  const types = data?.types || []
  const items = data ? [data] : []

  const COLUMNS = [
    { header: 'Nama', dataIndex: 'buyer_name', width: 150, render: (a) => <span className="font-semibold">{a.buyer_name}</span> },
    { header: 'Tipe', dataIndex: 'ticket_type', width: 110 },
    { header: 'Kode tiket', dataIndex: 'ticket_code', width: 110, render: (a) => <span className="font-mono">{a.ticket_code}</span> },
    { header: 'Bayar', dataIndex: 'payment_status', width: 100, render: (a) => <Badge tone={a.payment_status === 'PAID' ? 'success' : 'warning'}>{a.payment_status}</Badge> },
    { header: 'Hadir', dataIndex: 'attendance', width: 90, render: (a) => <Badge tone={a.attendance === 'HADIR' ? 'success' : 'neutral'}>{a.attendance}</Badge> },
    {
      header: 'Detail',
      dataIndex: 'ticket_code',
      width: 70,
      render: (a) => (
        <button
          type="button"
          aria-label={`Detail ${a.buyer_name}`}
          onClick={() => setDetail(a)}
          className="rounded-lg p-1.5 text-zinc-500 hover:bg-zinc-100 dark:hover:bg-zinc-800"
        >
          <Icon name="eye" className="h-4 w-4" />
        </button>
      ),
    },
  ]

  return (
    <div className="flex flex-col gap-4">
      <StandardPage
        title={nvdata?.label || 'Attendees'}
        description="Hanya melihat. Check-in dikerjakan di module ticketing."
        loading={loading}
        showLoading={showLoading}
        error={error}
        errorTitle="Gagal memuat peserta"
        onClearError={() => setError('')}
        onRefresh={load}
        extraActions={
          <>
            <Button variant="secondary" size="sm" onClick={handler_back}>
              Kembali
            </Button>
            <Button variant="secondary" size="sm" onClick={handler_export_click}>
              <Icon name="download" className="h-4 w-4" />
              CSV
            </Button>
          </>
        }
        filterBar={
          <form onSubmit={(e) => e.preventDefault()} className="mb-4 flex flex-wrap items-end gap-2">
            <label className="flex flex-col gap-1 text-xs font-medium text-zinc-500">
              Tipe tiket
              <select value={type} onChange={(e) => { setType(e.target.value); setLimit(PAGE_LIMIT) }} className={SELECT_CLASS}>
                <option value="">Semua tipe</option>
                {types.map((t) => (
                  <option key={t.code} value={t.code}>
                    {t.name}
                  </option>
                ))}
              </select>
            </label>
            <label className="flex flex-col gap-1 text-xs font-medium text-zinc-500">
              Kehadiran
              <select value={attendance} onChange={(e) => { setAttendance(e.target.value); setLimit(PAGE_LIMIT) }} className={SELECT_CLASS}>
                <option value="">Semua</option>
                <option value="HADIR">Sudah hadir</option>
                <option value="BELUM">Belum hadir</option>
              </select>
            </label>
            <label className="flex flex-col gap-1 text-xs font-medium text-zinc-500">
              Pembayaran
              <select value={payment} onChange={(e) => { setPayment(e.target.value); setLimit(PAGE_LIMIT) }} className={SELECT_CLASS}>
                <option value="">Semua</option>
                <option value="PAID">Paid</option>
                <option value="PENDING">Pending</option>
                <option value="CANCELLED">Cancelled</option>
                <option value="REFUNDED">Refunded</option>
              </select>
            </label>
            <label className="flex min-w-[160px] flex-1 flex-col gap-1 text-xs font-medium text-zinc-500">
              Cari nama/email/kode
              <input
                value={query}
                onChange={(e) => { setQuery(e.target.value); setLimit(PAGE_LIMIT) }}
                placeholder="Ketik…"
                className={SELECT_CLASS}
              />
            </label>
          </form>
        }
        items={items}
        emptyTitle="Belum ada tiket terjual"
        emptyDescription="Peserta muncul di sini setelah ada pembelian."
        offset={0}
        onPage={null}
      >
        <div className="mb-4 grid gap-3 sm:grid-cols-4">
          {[
            ['Terjual', sum.total],
            ['Sudah hadir', sum.checked_in],
            ['Belum hadir', sum.pending],
            ['Kehadiran', `${Number(sum.rate || 0).toLocaleString('id-ID', { maximumFractionDigits: 1 })}%`],
          ].map(([label, value]) => (
            <div key={label} className="rounded-xl bg-zinc-50 p-3 text-sm dark:bg-zinc-800/50">
              <p className="text-xs text-zinc-500">{label}</p>
              <p className="text-lg font-bold">{value ?? 0}</p>
            </div>
          ))}
        </div>
        {showLoading ? (
          <Skeleton className="h-24" />
        ) : rows.length === 0 ? (
          <p className="text-sm text-zinc-500">Tidak ada peserta cocok dengan filter.</p>
        ) : (
          <>
            <StandardGrid columns={COLUMNS} rows={rows} minWidth={640} />
            {rows.length >= limit && (
              <div className="mt-4 text-center">
                <Button variant="secondary" size="sm" onClick={() => setLimit((l) => l + PAGE_LIMIT)} loading={loading}>
                  Muat lagi
                </Button>
              </div>
            )}
          </>
        )}
      </StandardPage>

      <AttendeeModal attendee={detail} onClose={() => setDetail(null)} />

      {error && (
        <Alert tone="error" title="Gagal memuat peserta" closable onClose={() => setError('')}>
          {error}
        </Alert>
      )}
    </div>
  )
}
