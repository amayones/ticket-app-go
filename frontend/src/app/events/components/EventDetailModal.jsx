// FRM Event Detail — padanan FRM<mod>.js langit_v2 (window + tbar).
// Modal baca detail event (read-only): dibuka dari kartu event (klik),
// dipasang dengan key={code} oleh induk agar isi ter-reset tiap ganti event
// (sama seperti EditUserModal di menu users).
import { useEffect, useRef, useState } from 'react'
import { Alert, Badge, Button, Icon, Modal, SafeImage, Skeleton, Tooltip, formatAmount, formatTime, useToast } from '../../shared/all.js'
import { PromoCard } from '../../shared/cards.jsx'
import { eventDetail } from '../api.js'

function rupiah(value) {
  const s = formatAmount(value)
  return s === '' ? 'Rp 0' : `Rp ${s}`
}

function ticketStateBadge(state) {
  switch (state) {
    case 'OPEN':
      return <Badge tone="success">Tersedia</Badge>
    case 'LOW':
      return <Badge tone="warning">Hampir habis</Badge>
    case 'SOLD_OUT':
      return <Badge tone="danger">Sold out</Badge>
    case 'NOT_OPEN':
      return <Badge tone="info">Belum dibuka</Badge>
    case 'CLOSED':
      return <Badge tone="neutral">Ditutup</Badge>
    default:
      return <Badge tone="neutral">{state}</Badge>
  }
}

function mapUrl(detail) {
  const q = detail.latitude
    ? `${detail.latitude},${detail.longitude}`
    : [detail.venue, detail.address, detail.city].filter(Boolean).join(', ')
  return `https://www.google.com/maps/search/?api=1&query=${encodeURIComponent(q)}`
}

export default function EventDetailModal({ code, onClose, onOpenEvent, onBuy }) {
  const toast = useToast()
  const [data, setData] = useState(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const abortRef = useRef(null)

  useEffect(() => {
    if (!code) return
    abortRef.current?.abort()
    const ctrl = new AbortController()
    abortRef.current = ctrl
    setLoading(true)
    setError('')
    setData(null)
    eventDetail(code, ctrl.signal)
      .then((detail) => {
        if (!ctrl.signal.aborted) setData(detail)
      })
      .catch((err) => {
        if (err && err.name === 'AbortError') return
        if (!ctrl.signal.aborted) setError(err.message)
      })
      .finally(() => {
        if (!ctrl.signal.aborted) setLoading(false)
      })
    return () => ctrl.abort()
  }, [code])

  useEffect(() => () => abortRef.current?.abort(), [])

  function handler_share() {
    const done = () => toast.success('Tautan disalin.')
    try {
      if (navigator.clipboard && navigator.clipboard.writeText) {
        navigator.clipboard.writeText(window.location.href).then(done).catch(() => toast.info(window.location.href, { title: 'Tautan' }))
      } else {
        toast.info(window.location.href, { title: 'Tautan' })
      }
    } catch {
      toast.info(window.location.href, { title: 'Tautan' })
    }
  }

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

  const buyable = !!data && !data.sold_out && (data.ticket_types || []).some((t) => t.state === 'OPEN' || t.state === 'LOW')
  const buyLabel = !data ? 'Buy Ticket' : data.sold_out ? 'Sold Out' : buyable ? 'Buy Ticket' : 'Segera hadir'

  return (
    <Modal
      open={!!code}
      onClose={onClose}
      title={data?.title || 'Detail event'}
      size="lg"
      footer={
        <>
          <Button variant="secondary" onClick={handler_share}>
            <Icon name="send" className="h-4 w-4" />
            Bagikan
          </Button>
          <Tooltip label={buyable ? 'Lanjut ke pembayaran' : buyLabel}>
            <span className="inline-flex">
              <Button disabled={!buyable} onClick={() => buyable && onBuy && onBuy(data.code)}>
                <Icon name="list" className="h-4 w-4" />
                {buyLabel}
              </Button>
            </span>
          </Tooltip>
        </>
      }
    >
      {loading ? (
        <div className="flex flex-col gap-2">
          <Skeleton className="h-44" />
          <Skeleton className="h-6 w-2/3" />
          <Skeleton className="h-20" />
        </div>
      ) : error || !data ? (
        <Alert tone="error" title="Event tidak ditemukan" closable onClose={() => setError('')}>
          {error || 'Event belum dipublikasi, sudah berakhir, atau dibatalkan.'}
        </Alert>
      ) : (
        <div className="flex min-w-0 flex-col gap-4">
          <SafeImage src={data.poster_url} alt={data.title} className="w-full rounded-xl object-cover" />
          <div className="flex min-w-0 flex-col gap-1.5">
            <span className="flex flex-wrap gap-1.5">
              {data.category ? <Badge tone="brand">{data.category}</Badge> : null}
              {data.sold_out ? <Badge tone="danger">Sold Out</Badge> : null}
            </span>
            <p className="text-sm text-zinc-500 dark:text-zinc-400">
              Penyelenggara: <span className="font-mono">{data.organizer || '—'}</span>
            </p>
            <p className="text-sm text-zinc-600 dark:text-zinc-300">
              {formatTime(data.start_at)}
              {data.end_at ? ` — ${formatTime(data.end_at)}` : ''}
            </p>
            <p className="text-sm text-zinc-600 dark:text-zinc-300">
              {[data.venue, data.address, data.city].filter(Boolean).join(', ')}{' '}
              <a
                href={mapUrl(data)}
                target="_blank"
                rel="noreferrer"
                className="font-semibold text-violet-600 hover:underline dark:text-violet-300"
              >
                Lihat peta
              </a>
            </p>
            {data.description ? (
              <p className="text-sm leading-relaxed text-zinc-700 dark:text-zinc-200">{data.description}</p>
            ) : null}
          </div>
          <div>
            <p className="mb-2 text-sm font-semibold text-zinc-900 dark:text-zinc-50">Tipe tiket</p>
            {(data.ticket_types || []).length === 0 ? (
              <p className="text-sm text-zinc-500">Belum ada tipe tiket.</p>
            ) : (
              <ul className="flex flex-col gap-2">
                {data.ticket_types.map((t) => (
                  <li
                    key={t.code}
                    className="flex items-center gap-3 rounded-xl border border-zinc-100 px-3 py-2.5 dark:border-zinc-800"
                  >
                    <div className="min-w-0 flex-1">
                      <p className="truncate text-sm font-semibold text-zinc-900 dark:text-zinc-50">{t.name}</p>
                      <p className="font-mono text-xs text-zinc-500">
                        {rupiah(t.price)} · sisa {t.remaining}/{t.quota}
                      </p>
                    </div>
                    {ticketStateBadge(t.state)}
                  </li>
                ))}
              </ul>
            )}
          </div>
          {(data.promos || []).length > 0 && (
            <div>
              <p className="mb-2 text-sm font-semibold text-zinc-900 dark:text-zinc-50">Promo berlaku</p>
              <div className="grid gap-2 sm:grid-cols-2">
                {data.promos.map((p) => (
                  <PromoCard key={p.code} promo={p} onCopy={handler_copy_code} />
                ))}
              </div>
            </div>
          )}
          {(data.related || []).length > 0 && (
            <div>
              <p className="mb-2 text-sm font-semibold text-zinc-900 dark:text-zinc-50">Event terkait</p>
              <div className="flex flex-col gap-1.5">
                {data.related.map((e) => (
                  <button
                    key={e.code}
                    type="button"
                    onClick={() => onOpenEvent && onOpenEvent(e.code)}
                    className="flex w-full items-center justify-between gap-2 rounded-xl bg-zinc-50 px-3 py-2 text-left text-sm hover:bg-zinc-100 dark:bg-zinc-800/50 dark:hover:bg-zinc-800"
                  >
                    <span className="min-w-0 truncate font-medium text-zinc-800 dark:text-zinc-100">{e.title}</span>
                    <span className="shrink-0 font-mono text-xs text-zinc-500">{formatTime(e.start_at)}</span>
                  </button>
                ))}
              </div>
            </div>
          )}
        </div>
      )}
    </Modal>
  )
}
