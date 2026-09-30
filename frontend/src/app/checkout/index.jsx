// View Checkout — beli tiket: qty per tipe, data pemesan, promo,
// rincian harga (backend), hitung mundur bayar, tombol bayar.
// Masuk dari tombol Buy di Event Detail (hash #/checkout/<eventCode>).
import { useEffect, useRef, useState } from 'react'
import {
  Alert,
  Badge,
  Button,
  Card,
  CardTitle,
  EmptyState,
  Skeleton,
  StandardPage,
  formatAmount,
  formatTime,
  useToast,
  INPUT_CLASS,
  clearHash,
  hashParam,
} from '../shared/all.js'
import { cancelOrder, checkoutEvent, checkoutOrder, payOrder } from './api.js'
import { controller } from './controller.js'

export const meta = { label: 'Checkout', icon: 'list', order: 1 }

function rupiah(value) {
  const s = formatAmount(value)
  return s === '' ? 'Rp 0' : `Rp ${s}`
}

function leftSeconds(expiresAt) {
  const end = new Date(String(expiresAt).replace(' ', 'T')).getTime()
  if (Number.isNaN(end)) return 0
  return Math.max(0, Math.floor((end - Date.now()) / 1000))
}

function clockText(total) {
  const m = Math.floor(total / 60)
  const s = total % 60
  const p = (n) => String(n).padStart(2, '0')
  return `${p(m)}:${p(s)}`
}

export default function Checkout({ nvdata, onNavigate }) {
  const toast = useToast()
  const [eventCode] = useState(() => hashParam('checkout'))
  const [event, setEvent] = useState(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [qty, setQty] = useState({})
  const [buyer, setBuyer] = useState({ name: '', email: '', phone: '' })
  const [promo, setPromo] = useState('')
  const [order, setOrder] = useState(null)
  const [orderError, setOrderError] = useState('')
  const [ordering, setOrdering] = useState(false)
  const [paying, setPaying] = useState(false)
  const [left, setLeft] = useState(0)
  const abortRef = useRef(null)

  useEffect(() => {
    controller.init()
    if (!eventCode) {
      setLoading(false)
      return
    }
    abortRef.current?.abort()
    const ctrl = new AbortController()
    abortRef.current = ctrl
    setLoading(true)
    setError('')
    checkoutEvent(eventCode, ctrl.signal)
      .then((detail) => {
        if (!ctrl.signal.aborted) setEvent(detail)
      })
      .catch((err) => {
        if (err && err.name === 'AbortError') return
        if (!ctrl.signal.aborted) setError(err.message)
      })
      .finally(() => {
        if (!ctrl.signal.aborted) setLoading(false)
      })
    return () => ctrl.abort()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  useEffect(() => () => abortRef.current?.abort(), [])

  // Hitung mundur batas bayar; order hangus saat nol (backend melepas kuota).
  useEffect(() => {
    if (!order?.expires_at) return undefined
    setLeft(leftSeconds(order.expires_at))
    const id = setInterval(() => setLeft(leftSeconds(order.expires_at)), 1000)
    return () => clearInterval(id)
  }, [order])

  function handler_back() {
    clearHash()
    if (onNavigate) onNavigate('events')
  }

  function handler_qty(code, delta, max) {
    setQty((prev) => {
      const cur = prev[code] || 0
      const next = Math.min(Math.max(0, cur + delta), max)
      return { ...prev, [code]: next }
    })
  }

  async function handler_order_click() {
    const items = (event.ticket_types || [])
      .filter((t) => (qty[t.code] || 0) > 0)
      .map((t) => ({ type_code: t.code, qty: qty[t.code] }))
    if (items.length === 0) {
      setOrderError('Pilih minimal 1 tiket.')
      return
    }
    if (!buyer.name.trim()) {
      setOrderError('Nama pemesan wajib diisi.')
      return
    }
    setOrdering(true)
    setOrderError('')
    try {
      const res = await checkoutOrder({
        event_code: eventCode,
        items,
        buyer_name: buyer.name.trim(),
        buyer_email: buyer.email.trim(),
        buyer_phone: buyer.phone.trim(),
        promo_code: promo.trim(),
      })
      setOrder(res)
      toast.success('Order dibuat, segera bayar sebelum hangus.')
    } catch (ex) {
      setOrderError(ex.message)
    } finally {
      setOrdering(false)
    }
  }

  async function handler_pay_click() {
    if (!order) return
    setPaying(true)
    try {
      await payOrder(order.order_code, { holder: buyer.name.trim(), email: buyer.email.trim(), phone: buyer.phone.trim() })
      toast.success('Pembayaran simulasi berhasil. Tiket terbit.', { title: 'Lunas' })
      clearHash()
      if (onNavigate) onNavigate('my_tickets')
    } catch (ex) {
      toast.error(ex.message, { title: 'Bayar gagal' })
    } finally {
      setPaying(false)
    }
  }

  async function handler_cancel_order() {
    if (!order) return
    try {
      await cancelOrder(order.order_code)
      toast.info('Order dibatalkan, kuota dilepas.')
      setOrder(null)
    } catch (ex) {
      toast.error(ex.message, { title: 'Error' })
    }
  }

  if (!eventCode) {
    return (
      <Card>
        <EmptyState
          icon="list"
          title="Pilih event dulu"
          description="Masuk lewat tombol Buy di halaman event."
          action={
            <Button variant="secondary" size="sm" onClick={handler_back}>
              Ke Events
            </Button>
          }
        />
      </Card>
    )
  }

  const types = (event?.ticket_types || []).filter((t) => t.state === 'OPEN' || t.state === 'LOW')
  const expired = order && left <= 0

  return (
    <div className="flex flex-col gap-4">
      <StandardPage
        title={nvdata?.label || 'Checkout'}
        description={event ? `${event.title} — ${formatTime(event.start_at)}` : 'Ringkasan pembelian.'}
        loading={loading}
        showLoading={loading}
        error={error}
        errorTitle="Gagal memuat event"
        onClearError={() => setError('')}
        onRefresh={() => window.location.reload()}
        items={event ? [event] : []}
        emptyTitle="Event tidak ditemukan"
        emptyDescription="Event belum dipublikasi, sudah berakhir, atau dibatalkan."
      >
        {!order ? (
          <div className="flex flex-col gap-4">
            <Card>
              <CardTitle description="Atur jumlah per tipe tiket.">Pilih tiket</CardTitle>
              {types.length === 0 ? (
                <p className="text-sm text-zinc-500">Tidak ada tipe yang sedang dijual.</p>
              ) : (
                <ul className="flex flex-col gap-2">
                  {types.map((t) => (
                    <li
                      key={t.code}
                      className="flex items-center gap-3 rounded-xl border border-zinc-100 px-3 py-2.5 dark:border-zinc-800"
                    >
                      <div className="min-w-0 flex-1">
                        <p className="truncate text-sm font-semibold text-zinc-900 dark:text-zinc-50">
                          {t.name} {t.state === 'LOW' ? <Badge tone="warning">Hampir habis</Badge> : null}
                        </p>
                        <p className="font-mono text-xs text-zinc-500">
                          {rupiah(t.price)} · sisa {t.remaining}
                        </p>
                      </div>
                      <div className="flex shrink-0 items-center gap-2">
                        <Button size="sm" variant="secondary" onClick={() => handler_qty(t.code, -1, t.remaining)}>
                          −
                        </Button>
                        <span className="w-8 text-center font-mono text-sm font-bold">{qty[t.code] || 0}</span>
                        <Button size="sm" variant="secondary" onClick={() => handler_qty(t.code, 1, t.remaining)}>
                          +
                        </Button>
                      </div>
                    </li>
                  ))}
                </ul>
              )}
            </Card>

            <Card>
              <CardTitle description="Tiket diterbitkan atas nama ini.">Data pemesan</CardTitle>
              <div className="grid gap-3 sm:grid-cols-3">
                <label className="block">
                  <span className="mb-1.5 block text-sm font-medium text-zinc-700 dark:text-zinc-200">Nama</span>
                  <input value={buyer.name} onChange={(e) => setBuyer({ ...buyer, name: e.target.value })} placeholder="Nama lengkap" className={INPUT_CLASS} />
                </label>
                <label className="block">
                  <span className="mb-1.5 block text-sm font-medium text-zinc-700 dark:text-zinc-200">Email</span>
                  <input value={buyer.email} onChange={(e) => setBuyer({ ...buyer, email: e.target.value })} placeholder="nama@email.com" className={INPUT_CLASS} />
                </label>
                <label className="block">
                  <span className="mb-1.5 block text-sm font-medium text-zinc-700 dark:text-zinc-200">Telepon</span>
                  <input value={buyer.phone} onChange={(e) => setBuyer({ ...buyer, phone: e.target.value })} placeholder="08…" className={INPUT_CLASS} />
                </label>
              </div>
              <label className="mt-3 block max-w-xs">
                <span className="mb-1.5 block text-sm font-medium text-zinc-700 dark:text-zinc-200">Kode promo (opsional)</span>
                <input value={promo} onChange={(e) => setPromo(e.target.value)} placeholder="DISC10" className={`${INPUT_CLASS} font-mono uppercase`} />
              </label>
              {orderError && (
                <Alert tone="error" closable onClose={() => setOrderError('')} className="mt-3">
                  {orderError}
                </Alert>
              )}
              <div className="mt-4">
                <Button onClick={handler_order_click} loading={ordering}>
                  Buat Order
                </Button>
              </div>
            </Card>
          </div>
        ) : (
          <Card>
            <CardTitle description={`Order ${order.order_code} · hitung mundur ${clockText(left)}`}>
              {expired ? 'Order hangus' : 'Bayar sebelum hangus'}
            </CardTitle>
            {expired ? (
              <div className="flex flex-col items-start gap-3">
                <p className="text-sm text-zinc-500">Batas bayar lewat, kuota dilepas. Ulangi dari awal.</p>
                <Button variant="secondary" size="sm" onClick={() => setOrder(null)}>
                  Buat order baru
                </Button>
              </div>
            ) : (
              <dl className="flex flex-col gap-2 text-sm">
                {[
                  ['Subtotal', rupiah(order.gross)],
                  ['Diskon', `− ${rupiah(order.discount)}`],
                  ['Biaya layanan', rupiah(order.fee)],
                  ['Total', rupiah(order.net)],
                ].map(([label, value]) => (
                  <div key={label} className="flex items-center justify-between gap-4">
                    <dt className="text-zinc-500 dark:text-zinc-400">{label}</dt>
                    <dd className="font-mono font-semibold text-zinc-900 dark:text-zinc-50">{value}</dd>
                  </div>
                ))}
              </dl>
            )}
            {!expired && (
              <div className="mt-4 flex flex-wrap gap-2">
                <Button onClick={handler_pay_click} loading={paying}>
                  Bayar (Simulasi)
                </Button>
                <Button variant="secondary" onClick={handler_cancel_order} disabled={paying}>
                  Batalkan
                </Button>
              </div>
            )}
            <p className="mt-3 text-xs text-zinc-400">Simulasi bayar hanya aktif di development; FINANCE menyambung di sini.</p>
          </Card>
        )}
      </StandardPage>
    </div>
  )
}
