import { useEffect, useState } from 'react'
import {
  Alert,
  Badge,
  Button,
  Card,
  CardTitle,
  Modal,
  StandardPage,
  StandardGrid,
  SafeImage,
  TrendChart,
  useStandardController,
  useMessageBox,
  useToast,
  SELECT_CLASS,
} from '../shared/all.js'
import { formatAmount, formatTime } from '../../components/format.js'
import { adDetail, adPositions, adStatus, ownedEvents, read_data, simulatePay } from './api.js'
import { controller } from './controller.js'
import AdModal from './FRM.jsx'

export const meta = { label: 'Ad Campaigns', icon: 'image', order: 2 }

const STATUSES = [
  { value: '', label: 'Semua status' },
  { value: 'draft', label: 'Draft' },
  { value: 'menunggu_pembayaran', label: 'Menunggu pembayaran' },
  { value: 'menunggu_persetujuan', label: 'Menunggu persetujuan' },
  { value: 'tayang', label: 'Tayang' },
  { value: 'selesai', label: 'Selesai' },
  { value: 'ditolak', label: 'Ditolak' },
  { value: 'dijeda', label: 'Dijeda' },
]

function rupiah(v) {
  const s = formatAmount(v)
  return s === '' ? 'Rp 0' : `Rp ${s}`
}
function toneFor(s) {
  switch (s) {
    case 'tayang':
      return 'success'
    case 'draft':
      return 'neutral'
    case 'menunggu_pembayaran':
      return 'warning'
    case 'menunggu_persetujuan':
      return 'info'
    case 'selesai':
      return 'neutral'
    case 'ditolak':
      return 'danger'
    case 'dijeda':
      return 'warning'
    default:
      return 'neutral'
  }
}

export default function AdCampaigns({ nvdata }) {
  const toast = useToast()
  const { MessageBox, dialog } = useMessageBox()
  const [status, setStatus] = useState('')
  const [query, setQuery] = useState('')
  const [positions, setPositions] = useState([])
  const [events, setEvents] = useState([])
  const [showFRM, setShowFRM] = useState(false)
  const [editing, setEditing] = useState(null)
  const [detail, setDetail] = useState(null)
  const [detailStats, setDetailStats] = useState([])
  const params = { status, q: query }
  const { rows, offset, setOffset, loading, showLoading, error, setError, load } = useStandardController(read_data, params)

  useEffect(() => {
    controller.init()
    adPositions().then(setPositions).catch(() => {})
    ownedEvents().then(setEvents).catch(() => {})
  }, [])

  function handler_create() {
    setEditing(null)
    setShowFRM(true)
  }
  function handler_edit(row) {
    setEditing(row)
    setShowFRM(true)
  }
  async function handler_detail(row) {
    try {
      const data = await adDetail(row.code)
      setDetail(data.campaign || data)
      setDetailStats(data.stats || [])
    } catch (ex) {
      toast.error(ex.message, { title: 'Error' })
    }
  }
  async function handler_action(row, action) {
    try {
      await adStatus(row.code, action)
      toast.success('Status diperbarui.')
      load()
    } catch (ex) {
      toast.error(ex.message, { title: 'Error' })
    }
  }
  async function handler_simulate(row) {
    try {
      await simulatePay(row.code)
      toast.success('Pembayaran disimulasi (dev).')
      load()
    } catch (ex) {
      toast.error(ex.message, { title: 'Error' })
    }
  }

  const items = [{ rows }]
  const COLUMNS = [
    { header: 'Judul', dataIndex: 'title', width: 180, render: (r) => <span className="font-semibold">{r.title}</span> },
    { header: 'Gambar', dataIndex: 'image_url', width: 70, render: (r) => <SafeImage src={r.image_url} alt={r.title} className="h-10 w-16 rounded object-cover" /> },
    { header: 'Posisi', dataIndex: 'position', width: 110, render: (r) => r.position_name || r.position },
    { header: 'Periode', dataIndex: 'start_at', width: 160, render: (r) => `${(r.start_at || '').slice(0, 10)} → ${(r.end_at || '').slice(0, 10)}` },
    { header: 'Biaya', dataIndex: 'cost', width: 110, render: (r) => rupiah(r.cost) },
    { header: 'Status', dataIndex: 'display_status', width: 130, render: (r) => <Badge tone={toneFor(r.display_status)}>{r.display_status}</Badge> },
    { header: 'Imp.', dataIndex: 'impressions', width: 70, render: (r) => r.impressions },
    { header: 'Klik', dataIndex: 'clicks', width: 60, render: (r) => r.clicks },
    { header: 'CTR', dataIndex: 'ctr', width: 70, render: (r) => `${(Number(r.ctr) || 0).toFixed(1)}%` },
    {
      header: 'Aksi',
      dataIndex: 'code',
      width: 220,
      render: (r) => (
        <span className="flex flex-wrap gap-1">
          <Button size="sm" variant="secondary" onClick={() => handler_detail(r)}>Detail</Button>
          {r.display_status === 'draft' && <Button size="sm" variant="secondary" onClick={() => handler_edit(r)}>Ubah</Button>}
          {r.display_status === 'draft' && <Button size="sm" variant="primary" onClick={() => handler_action(r, 'submit')}>Submit</Button>}
          {r.display_status === 'menunggu_pembayaran' && <Button size="sm" variant="primary" onClick={() => handler_simulate(r)}>Bayar (dev)</Button>}
          {r.display_status === 'tayang' && <Button size="sm" variant="secondary" onClick={() => handler_action(r, 'pause')}>Jeda</Button>}
          {r.display_status === 'dijeda' && <Button size="sm" variant="primary" onClick={() => handler_action(r, 'resume')}>Lanjut</Button>}
          {(r.display_status === 'tayang' || r.display_status === 'dijeda') && <Button size="sm" variant="ghost" onClick={() => handler_action(r, 'end')}>Selesai</Button>}
        </span>
      ),
    },
  ]

  return (
    <div className="flex flex-col gap-4">
      <StandardPage
        title={nvdata?.label || 'Ad Campaigns'}
        description="Kampanye iklan milik Anda. Biaya dari tarif posisi × durasi."
        loading={loading}
        showLoading={showLoading}
        error={error}
        errorTitle="Gagal memuat kampanye"
        onClearError={() => setError('')}
        onRefresh={() => controller.btrefresh_click(load)}
        onCreate={() => controller.btnew_click(handler_create)}
        createLabel="Create Campaign"
        filterBar={
          <form onSubmit={(e) => e.preventDefault()} className="mb-4 flex flex-wrap items-end gap-2">
            <label className="flex flex-col gap-1 text-xs font-medium text-zinc-500">
              Status
              <select value={status} onChange={(e) => { setStatus(e.target.value); setOffset(0) }} className={SELECT_CLASS}>
                {STATUSES.map((s) => (
                  <option key={s.label} value={s.value}>{s.label}</option>
                ))}
              </select>
            </label>
            <label className="flex min-w-[160px] flex-1 flex-col gap-1 text-xs font-medium text-zinc-500">
              Cari judul
              <input value={query} onChange={(e) => { setQuery(e.target.value); setOffset(0) }} placeholder="Ketik judul…" className={SELECT_CLASS} />
            </label>
          </form>
        }
        items={items}
        emptyTitle="Belum ada kampanye"
        emptyDescription="Buat kampanye pertama Anda."
        emptyAction={<Button size="sm" onClick={handler_create}>Create Campaign</Button>}
        offset={offset}
        onPage={(next) => next >= 0 && setOffset(next)}
      >
        <StandardGrid columns={COLUMNS} rows={rows} minWidth={1100} />
      </StandardPage>

      <AdModal open={showFRM} editing={editing} positions={positions} events={events} onClose={() => { setShowFRM(false); setEditing(null) }} onSaved={load} />

      <Modal open={!!detail} onClose={() => setDetail(null)} title={detail?.title || 'Detail Kampanye'} size="lg" footer={<Button variant="secondary" onClick={() => setDetail(null)}>Tutup</Button>}>
        {detail && (
          <div className="flex flex-col gap-4">
            <div className="rounded-xl border border-zinc-200 p-3 dark:border-zinc-700">
              <p className="mb-2 text-xs font-bold uppercase tracking-wide text-zinc-500">Pratinjau Home (label Iklan)</p>
              <div className="relative overflow-hidden rounded-xl">
                <SafeImage src={detail.image_url} alt={detail.title} className="h-40 w-full object-cover" />
                <span className="absolute left-2 top-2 rounded-full bg-black/60 px-2 py-0.5 text-[11px] font-bold text-white">Iklan</span>
              </div>
              <p className="mt-2 text-sm font-semibold">{detail.title}</p>
              <p className="text-xs text-zinc-500">{detail.position_name || detail.position} · {detail.start_at?.slice(0, 10)} → {detail.end_at?.slice(0, 10)} · {rupiah(detail.cost)} · {detail.display_status}</p>
              {detail.target_url && <a href={detail.target_url} target="_blank" rel="noreferrer" className="text-xs text-violet-600 underline">{detail.target_url}</a>}
            </div>
            <div className="grid grid-cols-3 gap-3">
              <div className="rounded-xl bg-zinc-50 p-3 text-center dark:bg-zinc-800/50"><p className="text-xs text-zinc-500">Impressions</p><p className="text-lg font-bold">{detail.impressions}</p></div>
              <div className="rounded-xl bg-zinc-50 p-3 text-center dark:bg-zinc-800/50"><p className="text-xs text-zinc-500">Klik</p><p className="text-lg font-bold">{detail.clicks}</p></div>
              <div className="rounded-xl bg-zinc-50 p-3 text-center dark:bg-zinc-800/50"><p className="text-xs text-zinc-500">CTR</p><p className="text-lg font-bold">{(Number(detail.ctr) || 0).toFixed(2)}%</p></div>
            </div>
            {detailStats.length > 0 ? (
              <>
                <CardTitle description="Impressions per hari">Grafik</CardTitle>
                <TrendChart points={detailStats.map((p) => ({ label: p.date, value: p.impressions }))} height={160} />
                <TrendChart points={detailStats.map((p) => ({ label: p.date, value: p.clicks }))} height={120} />
              </>
            ) : (
              <p className="text-sm text-zinc-500">Belum ada data harian.</p>
            )}
            <p className="text-xs text-zinc-500">Pembayaran: {detail.payment_status} · Persetujuan: {detail.approval_status}{detail.reject_note ? ` · ${detail.reject_note}` : ''}</p>
            <div className="rounded-xl bg-amber-50 p-3 text-xs text-amber-800 dark:bg-amber-950 dark:text-amber-200">
              Titik sambung FINANCE: pembayaran via module FINANCE nanti. Simulasi bayar hanya aktif di development.
            </div>
          </div>
        )}
      </Modal>
      {dialog}
    </div>
  )
}
