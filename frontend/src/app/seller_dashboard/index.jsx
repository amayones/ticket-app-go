// View Seller Dashboard — padanan modul dashboard_utama langit_v2
// (filter + kartu KPI + grafik + daftar). Read-only penuh.
// Controller: ./controller.js (6 fungsi). Store: read_data ringkasan.
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
  Tooltip,
  formatAmount,
  formatTime,
  useDashboard,
  useToast,
  SELECT_CLASS,
  BarList,
  TrendChart,
} from '../shared/all.js'
import { exportSummary, read_data } from './api.js'
import { controller } from './controller.js'

export const meta = { label: 'Seller Dashboard', icon: 'folder', order: 2 }

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

function statusTone(status) {
  switch (status) {
    case 'COMPLETED':
      return 'success'
    case 'CANCELLED':
    case 'RETURNED':
      return 'danger'
    case 'SHIPPED':
      return 'info'
    case 'PROCESSING':
      return 'warning'
    default:
      return 'brand'
  }
}

const ALERT_STYLE = {
  SHIP: { icon: 'clock', tone: 'warning' },
  MODERATION: { icon: 'warning', tone: 'danger' },
  PAYOUT: { icon: 'check', tone: 'success' },
}

export default function SellerDashboard({ nvdata }) {
  const toast = useToast()

  const [preset, setPreset] = useState('30days')
  const [from, setFrom] = useState('')
  const [to, setTo] = useState('')
  const [category, setCategory] = useState('')

  const filters = { preset, from: preset === 'custom' ? from : '', to: preset === 'custom' ? to : '', category }
  const { data, loading, showLoading, error, setError, load } = useDashboard(read_data, filters)

  useEffect(() => {
    controller.init()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  const sum = data || null
  const items = sum ? [sum] : []
  const categories = sum?.categories || []

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

  async function handler_export_click() {
    try {
      const blob = await exportSummary(filters)
      const url = URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = url
      a.download = 'seller-summary.csv'
      a.click()
      URL.revokeObjectURL(url)
      toast.success('Ringkasan CSV diunduh.')
    } catch (ex) {
      toast.error(ex.message, { title: 'Export gagal' })
    }
  }

  const kpis = sum
    ? [
        { icon: 'list', label: 'Total penjualan', value: rupiah(sum.total_sales.value), delta: sum.total_sales.delta_pct },
        {
          icon: 'folder',
          label: 'Total order',
          value: Number(sum.total_orders.value).toLocaleString('id-ID'),
          delta: sum.total_orders.delta_pct,
        },
        { icon: 'clock', label: 'Saldo tertahan', value: rupiah(sum.escrow.value), delta: sum.escrow.delta_pct },
        { icon: 'check', label: 'Saldo tersedia', value: rupiah(sum.available.value), delta: sum.available.delta_pct },
        {
          icon: 'box',
          label: 'Produk aktif',
          value: Number(sum.active_products.value).toLocaleString('id-ID'),
          delta: sum.active_products.delta_pct,
        },
        {
          icon: 'warning',
          label: 'Stok menipis',
          value: Number(sum.low_stock.value).toLocaleString('id-ID'),
          delta: sum.low_stock.delta_pct,
        },
        {
          icon: 'bell',
          label: 'Keluhan terbuka',
          value: Number(sum.open_complaints.value).toLocaleString('id-ID'),
          delta: sum.open_complaints.delta_pct,
        },
        {
          icon: 'eye',
          label: 'Rating toko',
          value: Number(sum.rating.value).toLocaleString('id-ID', { maximumFractionDigits: 2 }),
          delta: sum.rating.delta_pct,
        },
      ]
    : []

  const ORDER_COLUMNS = [
    { header: 'Kode', dataIndex: 'code', width: 120, render: (o) => <span className="font-mono">{o.code}</span> },
    { header: 'Produk', dataIndex: 'product', width: 160 },
    { header: 'Pembeli', dataIndex: 'buyer', width: 110, render: (o) => <span className="font-mono">{o.buyer}</span> },
    { header: 'Qty', dataIndex: 'qty', width: 60 },
    {
      header: 'Total',
      dataIndex: 'total',
      width: 130,
      render: (o) => <span className="font-mono">{rupiah(o.total)}</span>,
    },
    { header: 'Status', dataIndex: 'status', width: 110, render: (o) => <Badge tone={statusTone(o.status)}>{o.status}</Badge> },
    {
      header: 'Waktu',
      dataIndex: 'order_date',
      width: 150,
      render: (o) => <span className="whitespace-nowrap text-zinc-500">{formatTime(o.order_date)}</span>,
    },
  ]

  const ACTION_COLUMNS = [
    ...ORDER_COLUMNS.slice(0, 6),
    {
      header: 'Batas kirim',
      dataIndex: 'ship_deadline',
      width: 150,
      render: (o) => (
        <span className="whitespace-nowrap text-amber-700 dark:text-amber-300">
          {o.ship_deadline ? formatTime(o.ship_deadline) : '—'}
        </span>
      ),
    },
  ]

  const STOCK_COLUMNS = [
    { header: 'Kode', dataIndex: 'code', width: 120, render: (p) => <span className="font-mono">{p.code}</span> },
    { header: 'Produk', dataIndex: 'name', width: 180 },
    { header: 'Kategori', dataIndex: 'category', width: 120 },
    {
      header: 'Harga',
      dataIndex: 'price',
      width: 130,
      render: (p) => <span className="font-mono">{rupiah(p.price)}</span>,
    },
    {
      header: 'Stok',
      dataIndex: 'stock',
      width: 110,
      render: (p) => (
        <Badge tone={p.stock <= 0 ? 'danger' : 'warning'}>
          {p.stock}/{p.threshold}
        </Badge>
      ),
    },
  ]

  const COMPLAINT_COLUMNS = [
    { header: 'Kode', dataIndex: 'code', width: 110, render: (c) => <span className="font-mono">{c.code}</span> },
    { header: 'Order', dataIndex: 'order_code', width: 120, render: (c) => <span className="font-mono">{c.order_code}</span> },
    { header: 'Keluhan', dataIndex: 'message', width: 280 },
    {
      header: 'Dilaporkan',
      dataIndex: 'created_at',
      width: 150,
      render: (c) => <span className="whitespace-nowrap text-zinc-500">{formatTime(c.created_at)}</span>,
    },
  ]

  return (
    <div className="flex flex-col gap-4">
      <StandardPage
        title={nvdata?.label || 'Seller Dashboard'}
        description="Ringkasan toko Anda. Data diperbarui otomatis tiap 60 detik."
        loading={loading}
        showLoading={showLoading}
        error={error}
        errorTitle="Gagal memuat dashboard"
        onClearError={() => setError('')}
        onRefresh={() => controller.btrefresh_click(load)}
        extraActions={
          <Button variant="secondary" size="sm" onClick={handler_export_click}>
            <Icon name="download" className="h-4 w-4" />
            CSV
          </Button>
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
              Kategori
              <select value={category} onChange={(e) => setCategory(e.target.value)} className={SELECT_CLASS}>
                <option value="">Semua kategori</option>
                {categories.map((c) => (
                  <option key={c} value={c}>
                    {c}
                  </option>
                ))}
              </select>
            </label>
          </form>
        }
        items={items}
        emptyTitle="Belum ada produk"
        emptyDescription="Tambahkan produk pertama Anda di menu Merchandise agar dashboard terisi."
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

      <Card>
        <CardTitle description="Jumlah pesanan per status pada rentang filter.">Pesanan per status</CardTitle>
        {showLoading ? (
          <Skeleton className="h-24" />
        ) : (sum?.orders_by_status || []).length === 0 ? (
          <p className="text-sm text-zinc-500">Belum ada pesanan pada rentang ini.</p>
        ) : (
          <div className="flex flex-wrap gap-2">
            {(sum.orders_by_status || []).map((s) => (
              <Badge key={s.name} tone={statusTone(s.name)}>
                {s.name} · {s.orders}
              </Badge>
            ))}
          </div>
        )}
      </Card>

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

      <Card>
        <CardTitle description="5 produk dengan pendapatan tertinggi.">Produk terlaris</CardTitle>
        {showLoading ? (
          <Skeleton className="h-24" />
        ) : (sum?.top_products || []).length === 0 ? (
          <p className="text-sm text-zinc-500">Belum ada penjualan produk.</p>
        ) : (
          <BarList
            items={(sum.top_products || []).map((p) => ({ label: p.name, value: p.revenue, sub: `${p.tickets} terjual` }))}
            formatValue={(v) => rupiah(v)}
          />
        )}
      </Card>

      <div className="grid gap-4 lg:grid-cols-2">
        <Card>
          <CardTitle description="Pesanan NEW/PROCESSING dengan batas waktu kirim.">Perlu tindakan</CardTitle>
          {showLoading ? (
            <Skeleton className="h-24" />
          ) : (sum?.actionable_orders || []).length === 0 ? (
            <p className="text-sm text-zinc-500">Tidak ada pesanan yang menunggu tindakan.</p>
          ) : (
            <StandardGrid columns={ACTION_COLUMNS} rows={sum.actionable_orders} minWidth={640} />
          )}
        </Card>
        <Card>
          <CardTitle description="Stok di bawah atau sama dengan batas.">Stok menipis</CardTitle>
          {showLoading ? (
            <Skeleton className="h-24" />
          ) : (sum?.low_stock_products || []).length === 0 ? (
            <p className="text-sm text-zinc-500">Semua stok aman.</p>
          ) : (
            <StandardGrid columns={STOCK_COLUMNS} rows={sum.low_stock_products} minWidth={560} />
          )}
        </Card>
      </div>

      <Card>
        <CardTitle description="10 pesanan terbaru sesuai filter.">Pesanan terbaru</CardTitle>
        {showLoading ? (
          <Skeleton className="h-24" />
        ) : (sum?.recent_orders || []).length === 0 ? (
          <p className="text-sm text-zinc-500">Belum ada pesanan pada rentang ini.</p>
        ) : (
          <StandardGrid columns={ORDER_COLUMNS} rows={sum.recent_orders} minWidth={760} />
        )}
      </Card>

      <Card>
        <CardTitle description="Klik untuk detail.">Keluhan belum selesai</CardTitle>
        {showLoading ? (
          <Skeleton className="h-24" />
        ) : (sum?.complaints || []).length === 0 ? (
          <p className="text-sm text-zinc-500">Tidak ada keluhan terbuka.</p>
        ) : (
          <ul className="flex flex-col gap-2">
            {(sum.complaints || []).map((c) => (
              <li key={c.code}>
                <button
                  type="button"
                  onClick={() => toast.info(c.message, { title: `Keluhan ${c.code} — order ${c.order_code}` })}
                  className="flex w-full items-center gap-3 rounded-xl bg-zinc-50 px-3 py-2 text-left text-sm hover:bg-zinc-100 dark:bg-zinc-800/50 dark:hover:bg-zinc-800"
                >
                  <Icon name="bell" className="h-4 w-4 shrink-0 text-amber-600 dark:text-amber-300" />
                  <span className="min-w-0 flex-1 truncate text-zinc-700 dark:text-zinc-200">{c.message}</span>
                  <Badge tone="warning">{c.order_code}</Badge>
                </button>
              </li>
            ))}
          </ul>
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
    </div>
  )
}
