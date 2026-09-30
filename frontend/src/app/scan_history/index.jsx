// View Scan History — tabel riwayat scan + ringkasan + export CSV.
// Controller: ./controller.js (6 fungsi). Store: read_data.
import { useEffect, useState } from 'react'
import {
  Alert,
  Badge,
  Button,
  Skeleton,
  StandardGrid,
  StandardPage,
  formatTime,
  useDashboard,
  useToast,
  SELECT_CLASS,
} from '../shared/all.js'
import { exportHistory, read_data } from './api.js'
import { controller } from './controller.js'

export const meta = { label: 'Scan History', icon: 'clock', order: 5 }

const PAGE_LIMIT = 20

function resultTone(result) {
  switch (result) {
    case 'VALID':
      return 'success'
    case 'USED':
    case 'OUTSIDE_SCHEDULE':
      return 'warning'
    default:
      return 'danger'
  }
}

export default function ScanHistory({ nvdata }) {
  const toast = useToast()
  const [event, setEvent] = useState('')
  const [result, setResult] = useState('')
  const [officer, setOfficer] = useState('')
  const [from, setFrom] = useState('')
  const [to, setTo] = useState('')
  const [limit, setLimit] = useState(PAGE_LIMIT)

  const filters = { event, result, officer, from, to, limit, offset: 0 }
  const { data, loading, showLoading, error, setError, load } = useDashboard(read_data, filters)

  useEffect(() => {
    controller.init()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  async function handler_export_click() {
    try {
      const blob = await exportHistory({ event, result, officer, from, to })
      const url = URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = url
      a.download = 'scan-history.csv'
      a.click()
      URL.revokeObjectURL(url)
      toast.success('Riwayat scan diunduh.')
    } catch (ex) {
      toast.error(ex.message, { title: 'Export gagal' })
    }
  }

  const logs = data?.logs || []
  const sum = data?.summary || {}
  const events = data?.events || []
  const pageItems = [{ logs }]

  const COLUMNS = [
    {
      header: 'Waktu',
      dataIndex: 'created_at',
      width: 150,
      render: (l) => <span className="whitespace-nowrap text-zinc-500">{formatTime(l.created_at)}</span>,
    },
    { header: 'Tiket', dataIndex: 'ticket_code', width: 110, render: (l) => <span className="font-mono">{l.ticket_code || '—'}</span> },
    { header: 'Pemegang', dataIndex: 'holder_name', width: 150 },
    { header: 'Tipe', dataIndex: 'type_name', width: 110 },
    { header: 'Hasil', dataIndex: 'result', width: 130, render: (l) => <Badge tone={resultTone(l.result)}>{l.result}</Badge> },
    { header: 'Petugas', dataIndex: 'officer', width: 110, render: (l) => <span className="font-mono">{l.officer}</span> },
  ]

  return (
    <div className="flex flex-col gap-4">
      <StandardPage
        title={nvdata?.label || 'Scan History'}
        description="Semua percobaan scan tercatat, termasuk yang ditolak."
        loading={loading}
        showLoading={showLoading}
        error={error}
        errorTitle="Gagal memuat riwayat"
        onClearError={() => setError('')}
        onRefresh={() => controller.btrefresh_click(load)}
        extraActions={
          <Button variant="secondary" size="sm" onClick={handler_export_click}>
            CSV
          </Button>
        }
        filterBar={
          <form onSubmit={(e) => e.preventDefault()} className="mb-4 flex flex-wrap items-end gap-2">
            <label className="flex flex-col gap-1 text-xs font-medium text-zinc-500">
              Event
              <select value={event} onChange={(e) => { setEvent(e.target.value); setLimit(PAGE_LIMIT) }} className={SELECT_CLASS}>
                <option value="">Semua event</option>
                {events.map((ev) => (
                  <option key={ev.code} value={ev.code}>
                    {ev.title}
                  </option>
                ))}
              </select>
            </label>
            <label className="flex flex-col gap-1 text-xs font-medium text-zinc-500">
              Hasil
              <select value={result} onChange={(e) => { setResult(e.target.value); setLimit(PAGE_LIMIT) }} className={SELECT_CLASS}>
                <option value="">Semua hasil</option>
                {['VALID', 'USED', 'INVALID', 'WRONG_EVENT', 'UNPAID', 'REFUNDED', 'EXPIRED', 'OUTSIDE_SCHEDULE'].map((r) => (
                  <option key={r} value={r}>
                    {r}
                  </option>
                ))}
              </select>
            </label>
            <label className="flex flex-col gap-1 text-xs font-medium text-zinc-500">
              Petugas
              <input
                value={officer}
                onChange={(e) => { setOfficer(e.target.value); setLimit(PAGE_LIMIT) }}
                placeholder="Kode user…"
                className={`${SELECT_CLASS} w-32 font-mono`}
              />
            </label>
            <label className="flex flex-col gap-1 text-xs font-medium text-zinc-500">
              Dari
              <input type="date" value={from} onChange={(e) => { setFrom(e.target.value); setLimit(PAGE_LIMIT) }} className={SELECT_CLASS} />
            </label>
            <label className="flex flex-col gap-1 text-xs font-medium text-zinc-500">
              Sampai
              <input type="date" value={to} onChange={(e) => { setTo(e.target.value); setLimit(PAGE_LIMIT) }} className={SELECT_CLASS} />
            </label>
          </form>
        }
        items={pageItems}
        emptyTitle="Belum ada scan"
        emptyDescription="Riwayat muncul setelah ada percobaan scan."
        offset={0}
        onPage={null}
      >
        <div className="mb-4 grid gap-3 sm:grid-cols-3">
          {[
            ['Total scan', sum.total],
            ['Valid', sum.valid],
            ['Ditolak', sum.denied],
          ].map(([label, value]) => (
            <div key={label} className="rounded-xl bg-zinc-50 p-3 text-sm dark:bg-zinc-800/50">
              <p className="text-xs text-zinc-500">{label}</p>
              <p className="text-lg font-bold">{Number(value || 0).toLocaleString('id-ID')}</p>
            </div>
          ))}
        </div>
        {showLoading ? (
          <Skeleton className="h-24" />
        ) : logs.length === 0 ? (
          <p className="text-sm text-zinc-500">Tidak ada riwayat cocok dengan filter.</p>
        ) : (
          <>
            <StandardGrid columns={COLUMNS} rows={logs} minWidth={720} />
            {logs.length >= limit && (
              <div className="mt-4 text-center">
                <Button variant="secondary" size="sm" onClick={() => setLimit((l) => l + PAGE_LIMIT)} loading={loading}>
                  Muat lagi
                </Button>
              </div>
            )}
          </>
        )}
      </StandardPage>

      {error && (
        <Alert tone="error" title="Gagal memuat riwayat" closable onClose={() => setError('')}>
          {error}
        </Alert>
      )}
    </div>
  )
}
