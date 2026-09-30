// View Organizer Dashboard — padanan modul dashboard_utama langit_v2
// (filter + kartu KPI + grafik + daftar). Read-only kecuali approve/reject
// refund (dikonfirmasi via MessageBox, tercatat di audit log backend).
// Controller: ./controller.js (6 fungsi). Store: read_data ringkasan.
import { useEffect, useState } from 'react'
import {
  Alert,
  Badge,
  Button,
  Card,
  CardTitle,
  EmptyState,
  Icon,
  Skeleton,
  StandardGrid,
  StandardPage,
  Tooltip,
  formatAmount,
  formatTime,
  useDashboard,
  useMessageBox,
  useToast,
  SELECT_CLASS,
  BarList,
  ProgressBar,
  TrendChart,
} from '../shared/all.js'
import { checkinLive, exportSummary, process_approve, process_reject, read_data } from './api.js'
import { controller } from './controller.js'

export const meta = { label: 'Organizer Dashboard', icon: 'terminal', order: 1 }

const PRESETS = [
  { value: 'today', label: 'Hari ini' },
  { value: '7days', label: '7 Hari' },
  { value: '30days', label: '30 Hari' },
  { value: 'month', label: 'Bulan ini' },
  { value: 'custom', label: 'Custom' },
]

function todayLocal() {
  const d = new Date()
  const p = (n) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${p(d.getMonth() + 1)}-${p(d.getDate())}`
}

function rupiah(value) {
  const s = formatAmount(value)
  return s === '' ? 'Rp 0' : `Rp ${s}`
}

function shortDate(iso) {
  try {
    return new Date(`${String(iso).slice(0, 10)}T00:00:00`).toLocaleDateString('id-ID', {
      day: 'numeric',
      month: 'short',
    })
  } catch {
    return String(iso).slice(0, 10)
  }
}

// Badge naik/turun dibanding periode sebelumnya (hijau naik, merah turun).
function DeltaBadge({ delta }) {
  const d = Number(delta) || 0
  if (d === 0) return <Badge tone="neutral">±0%</Badge>
  const up = d > 0
  return (
    <span
      className={`inline-flex items-center gap-1 rounded-full px-2 py-0.5 text-xs font-semibold ${
        up
          ? 'bg-emerald-100 text-emerald-700 dark:bg-emerald-950 dark:text-emerald-300'
          : 'bg-rose-100 text-rose-700 dark:bg-rose-950 dark:text-rose-300'
      }`}
    >
      <Icon name={up ? 'arrowUp' : 'arrowDown'} className="h-3 w-3" />
      {`${up ? '+' : ''}${d.toLocaleString('id-ID', { maximumFractionDigits: 1 })}%`}
    </span>
  )
}

function KpiCard({ icon, label, value, sub, delta }) {
  return (
    <div className="flex min-w-0 flex-col gap-1.5 rounded-2xl border border-zinc-200 bg-white p-4 dark:border-zinc-800 dark:bg-zinc-900">
      <div className="flex items-center justify-between gap-2">
        <span className="flex h-9 w-9 shrink-0 items-center justify-center rounded-xl bg-violet-100 text-violet-600 dark:bg-violet-950 dark:text-violet-300">
          <Icon name={icon} className="h-4 w-4" />
        </span>
        <DeltaBadge delta={delta} />
      </div>
      <p className="truncate text-xl font-bold text-zinc-900 dark:text-zinc-50">{value}</p>
      <p className="truncate text-xs text-zinc-500 dark:text-zinc-400">{label}</p>
      {sub ? <p className="truncate text-[11px] text-zinc-400 dark:text-zinc-500">{sub}</p> : null}
    </div>
  )
}

const ALERT_STYLE = {
  QUOTA: { icon: 'warning', tone: 'warning' },
  DRAFT: { icon: 'info', tone: 'info' },
  PROMO: { icon: 'clock', tone: 'warning' },
  PAYOUT: { icon: 'check', tone: 'success' },
}

export default function OrganizerDashboard({ nvdata }) {
  const toast = useToast()
  const { ask, dialog } = useMessageBox()

  const [preset, setPreset] = useState('30days')
  const [from, setFrom] = useState('')
  const [to, setTo] = useState('')
  const [event, setEvent] = useState('')
  const [city, setCity] = useState('')
  const [live, setLive] = useState([])

  const filters = { preset, from: preset === 'custom' ? from : '', to: preset === 'custom' ? to : '', event, city }
  const { data, loading, showLoading, error, setError, load } = useDashboard(read_data, filters)

  useEffect(() => {
    controller.init()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  // Live check-in: hanya untuk event yang mulai hari ini. Fetch terpisah
  // agar gagal di sini tidak merusak ringkasan utama.
  useEffect(() => {
    const today = todayLocal()
    const todays = (data?.upcoming_events || []).filter((u) => String(u.start_at || '').slice(0, 10) === today)
    if (todays.length === 0) {
      setLive([])
      return
    }
    let alive = true
    ;(async () => {
      try {
        const rows = await Promise.all(todays.map((t) => checkinLive(t.code)))
        if (alive) setLive(rows.filter(Boolean))
      } catch {
        if (alive) setLive([])
      }
    })()
    return () => {
      alive = false
    }
  }, [data])

  const sum = data || null
  const items = sum ? [sum] : []
  const events = sum?.upcoming_events || []
  // Opsi filter event diambil dari event milik sendiri (agar tak bisa
  // memilih event organizer lain); opsi kota dari ringkasan backend.
  const eventOptions = (sum?.upcoming_events || []).map((u) => ({ value: u.code, label: u.title }))
  const cityOptions = sum?.cities || []

  function handler_preset_click(value) {
    setPreset(value)
    if (value !== 'custom') {
      setFrom('')
      setTo('')
    } else if (!from || !to) {
      const t = todayLocal()
      const d = new Date()
      d.setDate(d.getDate() - 29)
      const p = (n) => String(n).padStart(2, '0')
      setFrom(`${d.getFullYear()}-${p(d.getMonth() + 1)}-${p(d.getDate())}`)
      setTo(t)
    }
  }

  async function handler_export_click(format) {
    try {
      const blob = await exportSummary(format, filters)
      const url = URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = url
      a.download = `organizer-summary.${format}`
      a.click()
      URL.revokeObjectURL(url)
      toast.success(`Ringkasan ${format.toUpperCase()} diunduh.`)
    } catch (ex) {
      toast.error(ex.message, { title: 'Export gagal' })
    }
  }

  async function handler_refund_click(refund, approve) {
    try {
      const ok = await ask({
        headline: approve ? 'Setujui refund?' : 'Tolak refund?',
        details: [
          { label: 'Kode', value: refund.code },
          { label: 'Order', value: refund.order_code },
          { label: 'Nominal', value: rupiah(refund.amount) },
        ],
        confirmLabel: approve ? 'Ya, setujui' : 'Ya, tolak',
        danger: !approve,
      })
      if (!ok) return
      const res = approve ? await process_approve(refund.code) : await process_reject(refund.code)
      toast.success(`Refund ${refund.code} ${String(res.status || '').toLowerCase()}.`)
      load()
    } catch (ex) {
      toast.error(ex.message, { title: 'Gagal memproses refund' })
    }
  }

  const kpis = sum
    ? [
        { icon: 'users', label: 'Tiket terjual', value: Number(sum.tickets_sold.value).toLocaleString('id-ID'), delta: sum.tickets_sold.delta_pct },
        { icon: 'list', label: 'Gross revenue', value: rupiah(sum.gross_revenue.value), delta: sum.gross_revenue.delta_pct },
        { icon: 'check', label: 'Net revenue', value: rupiah(sum.net_revenue.value), delta: sum.net_revenue.delta_pct },
        { icon: 'folder', label: 'Event aktif', value: Number(sum.active_events.value).toLocaleString('id-ID'), delta: sum.active_events.delta_pct },
        {
          icon: 'eye',
          label: 'Check-in rate',
          value: `${Number(sum.checkin_rate.value).toLocaleString('id-ID', { maximumFractionDigits: 1 })}%`,
          delta: sum.checkin_rate.delta_pct,
        },
        {
          icon: 'refresh',
          label: 'Refund',
          value: Number(sum.refund_count.value).toLocaleString('id-ID'),
          sub: `${Number(sum.refund_pct.value).toLocaleString('id-ID', { maximumFractionDigits: 1 })}% dari order`,
          delta: sum.refund_count.delta_pct,
        },
        {
          icon: 'bell',
          label: 'Pakai promo',
          value: Number(sum.promo_orders.value).toLocaleString('id-ID'),
          sub: `Diskon ${rupiah(sum.promo_discount.value)}`,
          delta: sum.promo_orders.delta_pct,
        },
        {
          icon: 'send',
          label: 'Iklan (impresi)',
          value: Number(sum.ad_impressions.value).toLocaleString('id-ID'),
          sub: `${Number(sum.ad_clicks.value).toLocaleString('id-ID')} klik • ${rupiah(sum.ad_cost.value)}`,
          delta: sum.ad_impressions.delta_pct,
        },
      ]
    : []

  const ORDER_COLUMNS = [
    { header: 'Kode', dataIndex: 'code', width: 120, render: (o) => <span className="font-mono">{o.code}</span> },
    { header: 'Event', dataIndex: 'event', width: 180 },
    { header: 'Pembeli', dataIndex: 'buyer', width: 110, render: (o) => <span className="font-mono">{o.buyer}</span> },
    { header: 'Tipe', dataIndex: 'ticket_type', width: 110 },
    { header: 'Qty', dataIndex: 'qty', width: 60 },
    {
      header: 'Net',
      dataIndex: 'net',
      width: 130,
      render: (o) => <span className="font-mono">{rupiah(o.net)}</span>,
    },
    { header: 'Status', dataIndex: 'status', width: 100, render: (o) => <Badge tone="success">{o.status}</Badge> },
    {
      header: 'Waktu',
      dataIndex: 'order_date',
      width: 150,
      render: (o) => <span className="whitespace-nowrap text-zinc-500">{formatTime(o.order_date)}</span>,
    },
  ]

  const REFUND_COLUMNS = [
    { header: 'Kode', dataIndex: 'code', width: 120, render: (r) => <span className="font-mono">{r.code}</span> },
    { header: 'Order', dataIndex: 'order_code', width: 120, render: (r) => <span className="font-mono">{r.order_code}</span> },
    { header: 'Event', dataIndex: 'event', width: 160 },
    {
      header: 'Nominal',
      dataIndex: 'amount',
      width: 130,
      render: (r) => <span className="font-mono">{rupiah(r.amount)}</span>,
    },
    { header: 'Alasan', dataIndex: 'reason', width: 200 },
    {
      header: 'Aksi',
      dataIndex: 'code',
      width: 190,
      render: (r) => (
        <span className="flex gap-1.5">
          <Button size="sm" onClick={() => handler_refund_click(r, true)}>
            Setujui
          </Button>
          <Button size="sm" variant="danger" onClick={() => handler_refund_click(r, false)}>
            Tolak
          </Button>
        </span>
      ),
    },
  ]

  return (
    <div className="flex flex-col gap-4">
      <StandardPage
        title={nvdata?.label || 'Organizer Dashboard'}
        description="Ringkasan bisnis event milik Anda. Data diperbarui otomatis tiap 60 detik."
        loading={loading}
        showLoading={showLoading}
        error={error}
        errorTitle="Gagal memuat dashboard"
        onClearError={() => setError('')}
        onRefresh={() => controller.btrefresh_click(load)}
        extraActions={
          <>
            <Button variant="secondary" size="sm" onClick={() => handler_export_click('csv')}>
              <Icon name="download" className="h-4 w-4" />
              CSV
            </Button>
            <Button variant="secondary" size="sm" onClick={() => handler_export_click('pdf')}>
              <Icon name="download" className="h-4 w-4" />
              PDF
            </Button>
          </>
        }
        filterBar={
          <form onSubmit={(e) => e.preventDefault()} className="mb-4 flex flex-wrap items-end gap-2">
            <span className="flex flex-wrap gap-1.5">
              {PRESETS.map((p) => (
                <Button
                  key={p.value}
                  size="sm"
                  variant={preset === p.value ? 'primary' : 'ghost'}
                  onClick={() => handler_preset_click(p.value)}
                >
                  {p.label}
                </Button>
              ))}
            </span>
            {preset === 'custom' && (
              <>
                <label className="flex flex-col gap-1 text-xs font-medium text-zinc-500">
                  Dari
                  <input type="date" value={from} onChange={(e) => setFrom(e.target.value)} className={SELECT_CLASS} />
                </label>
                <label className="flex flex-col gap-1 text-xs font-medium text-zinc-500">
                  Sampai
                  <input type="date" value={to} onChange={(e) => setTo(e.target.value)} className={SELECT_CLASS} />
                </label>
              </>
            )}
            <label className="flex flex-col gap-1 text-xs font-medium text-zinc-500">
              Event
              <select value={event} onChange={(e) => setEvent(e.target.value)} className={SELECT_CLASS}>
                <option value="">Semua event</option>
                {eventOptions.map((o) => (
                  <option key={o.value} value={o.value}>
                    {o.label}
                  </option>
                ))}
              </select>
            </label>
            <label className="flex flex-col gap-1 text-xs font-medium text-zinc-500">
              Kota
              <select value={city} onChange={(e) => setCity(e.target.value)} className={SELECT_CLASS}>
                <option value="">Semua kota</option>
                {cityOptions.map((c) => (
                  <option key={c.code} value={c.code}>
                    {c.name}
                  </option>
                ))}
              </select>
            </label>
          </form>
        }
        items={items}
        emptyTitle="Belum ada event"
        emptyDescription="Buat event pertama Anda di menu Event agar dashboard terisi."
        emptyAction={
          <Button variant="secondary" size="sm" onClick={() => controller.btrefresh_click(load)}>
            Muat ulang
          </Button>
        }
      >
        <div className="grid gap-3 sm:grid-cols-2 xl:grid-cols-4">
          {kpis.map((k) => (
            <KpiCard key={k.label} icon={k.icon} label={k.label} value={k.value} sub={k.sub} delta={k.delta} />
          ))}
        </div>
      </StandardPage>

      {live.length > 0 && (
        <Card>
          <CardTitle description="Hanya tampil saat ada event berlangsung hari ini.">Live check-in hari ini</CardTitle>
          <ul className="flex flex-col gap-3">
            {live.map((l) => (
              <li key={l.event_code}>
                <div className="mb-1 flex items-baseline justify-between gap-2 text-sm">
                  <span className="min-w-0 truncate font-medium text-zinc-800 dark:text-zinc-100">{l.title}</span>
                  <span className="shrink-0 font-mono text-xs text-zinc-500">
                    {l.checked_in}/{l.total} (
                    {Number(l.rate).toLocaleString('id-ID', { maximumFractionDigits: 1 })}%)
                  </span>
                </div>
                <ProgressBar value={l.checked_in} max={Math.max(1, l.total)} tone="emerald" />
              </li>
            ))}
          </ul>
        </Card>
      )}

      <Card>
        <CardTitle description="Penjualan per hari pada rentang filter.">Tren penjualan</CardTitle>
        {showLoading ? (
          <Skeleton className="h-44" />
        ) : (sum?.sales_trend || []).length === 0 ? (
          <p className="text-sm text-zinc-500">Belum ada penjualan pada rentang ini.</p>
        ) : (
          <TrendChart
            points={(sum.sales_trend || []).map((p) => ({ label: shortDate(p.date), value: p.revenue }))}
            formatValue={(v) => rupiah(v)}
          />
        )}
      </Card>

      <div className="grid gap-4 lg:grid-cols-3">
        {[
          { title: 'Per tipe tiket', desc: 'Pendapatan per tipe tiket.', rows: sum?.by_ticket_type || [] },
          { title: 'Per kota', desc: 'Pendapatan per kota event.', rows: sum?.by_city || [] },
          { title: 'Per kode promo', desc: 'Order dan diskon per kode promo.', rows: sum?.by_promo || [] },
        ].map((b) => (
          <Card key={b.title}>
            <CardTitle description={b.desc}>{b.title}</CardTitle>
            {showLoading ? (
              <Skeleton className="h-24" />
            ) : b.rows.length === 0 ? (
              <p className="text-sm text-zinc-500">Belum ada data.</p>
            ) : (
              <BarList
                items={b.rows.map((r) => ({
                  label: r.name,
                  value: r.revenue || r.orders,
                  sub: r.orders ? `${r.orders} order` : '',
                }))}
                formatValue={(v) => rupiah(v)}
              />
            )}
          </Card>
        ))}
      </div>

      <Card>
        <CardTitle description="Terjual vs kuota. Klik untuk detail.">Event mendatang</CardTitle>
        {showLoading ? (
          <Skeleton className="h-24" />
        ) : events.length === 0 ? (
          <EmptyState
            icon="folder"
            title="Belum ada event"
            description="Buat event pertama Anda di menu Event agar muncul di sini."
          />
        ) : (
          <ul className="flex flex-col gap-3">
            {events.map((u) => (
              <li key={u.code}>
                <button
                  type="button"
                  onClick={() => toast.info(`${u.title} — ${u.venue || 'tanpa venue'} — mulai ${formatTime(u.start_at)}`, { title: u.code })}
                  className="mb-1 flex w-full items-baseline justify-between gap-2 text-left text-sm"
                >
                  <span className="min-w-0 truncate font-medium text-zinc-800 hover:underline dark:text-zinc-100">
                    {u.title}
                  </span>
                  <span className="shrink-0 font-mono text-xs text-zinc-500">
                    {u.sold}/{u.quota}
                  </span>
                </button>
                <ProgressBar value={u.sold} max={Math.max(1, u.quota)} tone={u.quota > 0 && u.sold / u.quota >= 0.9 ? 'amber' : 'violet'} />
              </li>
            ))}
          </ul>
        )}
      </Card>

      <Card>
        <CardTitle description="10 order terbaru sesuai filter.">Order terbaru</CardTitle>
        {showLoading ? (
          <Skeleton className="h-24" />
        ) : (sum?.recent_orders || []).length === 0 ? (
          <p className="text-sm text-zinc-500">Belum ada order pada rentang ini.</p>
        ) : (
          <StandardGrid columns={ORDER_COLUMNS} rows={sum.recent_orders} minWidth={760} />
        )}
      </Card>

      <Card>
        <CardTitle description="Setujui atau tolak pengajuan refund dari pembeli.">Refund menunggu keputusan</CardTitle>
        {showLoading ? (
          <Skeleton className="h-24" />
        ) : (sum?.pending_refunds || []).length === 0 ? (
          <p className="text-sm text-zinc-500">Tidak ada refund yang menunggu.</p>
        ) : (
          <StandardGrid columns={REFUND_COLUMNS} rows={sum.pending_refunds} minWidth={760} />
        )}
      </Card>

      <Card>
        <CardTitle description="Klik untuk detail.">Perlu perhatian</CardTitle>
        {showLoading ? (
          <Skeleton className="h-24" />
        ) : (sum?.alerts || []).length === 0 ? (
          <p className="text-sm text-zinc-500">Semua aman, tidak ada peringatan.</p>
        ) : (
          <ul className="flex flex-col gap-2">
            {(sum.alerts || []).map((a, i) => {
              const st = ALERT_STYLE[a.kind] || { icon: 'info', tone: 'neutral' }
              return (
                <li key={`${a.kind}-${i}`}>
                  <button
                    type="button"
                    onClick={() => toast.info(a.detail, { title: a.title })}
                    className="flex w-full items-center gap-3 rounded-xl bg-zinc-50 px-3 py-2 text-left text-sm hover:bg-zinc-100 dark:bg-zinc-800/50 dark:hover:bg-zinc-800"
                  >
                    <Icon name={st.icon} className="h-4 w-4 shrink-0 text-zinc-500" />
                    <span className="min-w-0 flex-1 truncate font-medium text-zinc-800 dark:text-zinc-100">
                      {a.title}
                      <span className="ml-2 font-normal text-zinc-500">{a.detail}</span>
                    </span>
                    <Badge tone={st.tone}>{a.kind}</Badge>
                  </button>
                </li>
              )
            })}
          </ul>
        )}
      </Card>

      {error && (
        <Alert tone="error" title="Gagal memuat dashboard" closable onClose={() => setError('')}>
          {error}
        </Alert>
      )}
      {dialog}
    </div>
  )
}
