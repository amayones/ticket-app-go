import { useEffect, useRef, useState } from 'react'
import {
  Alert,
  Badge,
  Button,
  Card,
  CardTitle,
  Modal,
  TextField,
  StandardPage,
  StandardGrid,
  useStandardController,
  useMessageBox,
  useToast,
  SELECT_CLASS,
  INPUT_CLASS,
} from '../shared/all.js'
import { formatAmount, formatTime } from '../../components/format.js'
import { generatePromoCode, promoOptions, read_data, togglePromo, process_delete } from './api.js'
import { controller } from './controller.js'
import PromoModal from './FRM.jsx'

export const meta = { label: 'Promo Codes', icon: 'ticket', order: 1 }

const STATUSES = [
  { value: '', label: 'Semua status' },
  { value: 'aktif', label: 'Aktif' },
  { value: 'akan_datang', label: 'Akan datang' },
  { value: 'kedaluwarsa', label: 'Kedaluwarsa' },
  { value: 'kuota_habis', label: 'Kuota habis' },
  { value: 'nonaktif', label: 'Nonaktif' },
]

function rupiah(v) {
  const s = formatAmount(v)
  return s === '' ? 'Rp 0' : `Rp ${s}`
}

function statusTone(s) {
  switch (s) {
    case 'aktif':
      return 'success'
    case 'akan_datang':
      return 'info'
    case 'kedaluwarsa':
      return 'neutral'
    case 'kuota_habis':
      return 'warning'
    case 'nonaktif':
      return 'danger'
    default:
      return 'neutral'
  }
}

function useDebounced(value, delay = 400) {
  const [applied, setApplied] = useState(value)
  useEffect(() => {
    const id = setTimeout(() => setApplied(value), delay)
    return () => clearTimeout(id)
  }, [value, delay])
  return applied
}

export default function PromoCodes({ nvdata }) {
  const toast = useToast()
  const { MessageBox, dialog } = useMessageBox()
  const [status, setStatus] = useState('')
  const [query, setQuery] = useState('')
  const appliedQuery = useDebounced(query)
  const [showFRM, setShowFRM] = useState(false)
  const [editing, setEditing] = useState(null)
  const [detail, setDetail] = useState(null)
  const [opts, setOpts] = useState({ categories: [], events: [] })
  const gridRef = useRef(null)
  const params = { status, q: appliedQuery }
  const { rows, offset, setOffset, loading, showLoading, error, setError, load } = useStandardController(read_data, params)

  useEffect(() => {
    controller.init()
    promoOptions().then(setOpts).catch(() => {})
  }, [])

  function handler_open_create() {
    setEditing(null)
    setShowFRM(true)
  }

  function handler_open_edit(row) {
    setEditing(row)
    setShowFRM(true)
  }

  async function handler_toggle(row) {
    const next = row.display_status === 'nonaktif'
    try {
      await togglePromo(row.code, next)
      toast.success(next ? 'Promo diaktifkan.' : 'Promo dinonaktifkan.')
      load()
    } catch (ex) {
      toast.error(ex.message, { title: 'Error' })
    }
  }

  async function handler_row_delete(row) {
    try {
      const ok = await MessageBox.delete({
        headline: 'Hapus promo?',
        details: [{ label: 'Kode', value: row.promo_code }],
        danger: true,
        request: () => process_delete(row.code),
      })
      if (!ok) return
      toast.success('Promo dihapus.')
      load()
    } catch (ex) {
      toast.error(ex.message, { title: 'Error' })
    }
  }

  async function handler_detail(row) {
    try {
      const { promoDetail } = await import('./api.js')
      const data = await promoDetail(row.code)
      setDetail(data)
    } catch (ex) {
      toast.error(ex.message, { title: 'Error' })
    }
  }

  const items = [{ rows }]
  const COLUMNS = [
    { header: 'Kode', dataIndex: 'promo_code', width: 130, render: (r) => <span className="font-mono font-bold">{r.promo_code}</span> },
    {
      header: 'Potongan',
      dataIndex: 'discount_percent',
      width: 120,
      render: (r) => (r.discount_percent > 0 ? `${r.discount_percent}%` : rupiah(r.discount_amount)),
    },
    { header: 'Cakupan', dataIndex: 'scope_type', width: 110, render: (r) => r.scope_type || 'ALL' },
    { header: 'Periode', dataIndex: 'valid_from', width: 180, render: (r) => `${r.valid_from || ''} → ${r.valid_to || ''}` },
    {
      header: 'Pemakaian',
      dataIndex: 'used_count',
      width: 120,
      render: (r) => (r.remaining === -1 ? `${r.used_count} / ∞` : `${r.used_count} / ${r.used_count + r.remaining}`),
    },
    { header: 'Status', dataIndex: 'display_status', width: 110, render: (r) => <Badge tone={statusTone(r.display_status)}>{r.display_status}</Badge> },
    {
      header: 'Aksi',
      dataIndex: 'code',
      width: 200,
      render: (r) => (
        <span className="flex flex-wrap gap-1">
          <Button size="sm" variant="secondary" onClick={() => handler_detail(r)}>
            Detail
          </Button>
          <Button size="sm" variant="secondary" onClick={() => handler_open_edit(r)}>
            Ubah
          </Button>
          <Button size="sm" variant={r.display_status === 'nonaktif' ? 'primary' : 'secondary'} onClick={() => handler_toggle(r)}>
            {r.display_status === 'nonaktif' ? 'Aktifkan' : 'Nonaktifkan'}
          </Button>
          <Button size="sm" variant="ghost" onClick={() => handler_row_delete(r)}>
            Hapus
          </Button>
        </span>
      ),
    },
  ]

  return (
    <div className="flex flex-col gap-4">
      <StandardPage
        title={nvdata?.label || 'Promo Codes'}
        description="Kode promo milik Anda. Kelola potongan dan cakupannya."
        loading={loading}
        showLoading={showLoading}
        error={error}
        errorTitle="Gagal memuat promo"
        onClearError={() => setError('')}
        onRefresh={() => controller.btrefresh_click(load)}
        onCreate={() => controller.btnew_click(handler_open_create)}
        createLabel="Create Promo"
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
            <label className="flex min-w-[160px] flex-1 flex-col gap-1 text-xs font-medium text-zinc-500">
              Cari kode
              <input value={query} onChange={(e) => { setQuery(e.target.value); setOffset(0) }} placeholder="Ketik kode…" className={SELECT_CLASS} />
            </label>
          </form>
        }
        items={items}
        emptyTitle="Belum ada promo"
        emptyDescription="Buat promo pertama Anda."
        emptyAction={<Button size="sm" onClick={handler_open_create}>Create Promo</Button>}
        offset={offset}
        onPage={(next) => next >= 0 && setOffset(next)}
      >
        <StandardGrid columns={COLUMNS} rows={rows} minWidth={900} />
      </StandardPage>

      <PromoModal open={showFRM} editing={editing} opts={opts} onClose={() => { setShowFRM(false); setEditing(null) }} onSaved={load} onGenerate={generatePromoCode} />

      <Modal open={!!detail} onClose={() => setDetail(null)} title={detail?.promo?.promo_code || 'Detail Promo'} size="lg" footer={<Button variant="secondary" onClick={() => setDetail(null)}>Tutup</Button>}>
        {detail && (
          <div className="flex flex-col gap-4">
            <div className="grid gap-2 text-sm">
              <p><span className="font-semibold">Kode:</span> {detail.promo.promo_code}</p>
              <p><span className="font-semibold">Deskripsi:</span> {detail.promo.description || '—'}</p>
              <p><span className="font-semibold">Potongan:</span> {detail.promo.discount_percent > 0 ? `${detail.promo.discount_percent}% (maks ${rupiah(detail.promo.max_discount)})` : rupiah(detail.promo.discount_amount)}</p>
              <p><span className="font-semibold">Min. belanja:</span> {rupiah(detail.promo.min_purchase)}</p>
              <p><span className="font-semibold">Kuota:</span> {detail.promo.remaining === -1 ? `${detail.promo.used_count} / ∞` : `${detail.promo.used_count} / ${detail.promo.used_count + detail.promo.remaining}`}</p>
              <p><span className="font-semibold">Per pengguna:</span> {detail.promo.max_per_user ?? '∞'}</p>
              <p><span className="font-semibold">Periode:</span> {formatTime(detail.promo.valid_from)} → {formatTime(detail.promo.valid_to)}</p>
              <p><span className="font-semibold">Cakupan:</span> {detail.promo.scope_type} {(detail.promo.category_codes || []).join(', ')} {(detail.promo.event_codes || []).join(', ')}</p>
            </div>
            <div>
              <p className="mb-2 text-sm font-bold">Riwayat pemakaian</p>
              {(detail.usage || []).length === 0 ? (
                <p className="text-sm text-zinc-500">Belum ada pemakaian.</p>
              ) : (
                <div className="overflow-x-auto rounded-xl border border-zinc-100 dark:border-zinc-800">
                  <table className="w-full text-left text-sm">
                    <thead>
                      <tr className="bg-zinc-50 text-xs uppercase text-zinc-500 dark:bg-zinc-800/60">
                        <th className="px-3 py-2">Pembeli</th>
                        <th className="px-3 py-2">Order</th>
                        <th className="px-3 py-2">Potongan</th>
                        <th className="px-3 py-2">Tanggal</th>
                      </tr>
                    </thead>
                    <tbody>
                      {detail.usage.map((u) => (
                        <tr key={u.code} className="border-t border-zinc-100 dark:border-zinc-800">
                          <td className="px-3 py-2">{u.username}</td>
                          <td className="px-3 py-2 font-mono text-xs">{u.order_code}</td>
                          <td className="px-3 py-2">{rupiah(u.discount)}</td>
                          <td className="px-3 py-2 text-xs">{formatTime(u.used_at)}</td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              )}
            </div>
          </div>
        )}
      </Modal>
      {dialog}
    </div>
  )
}
