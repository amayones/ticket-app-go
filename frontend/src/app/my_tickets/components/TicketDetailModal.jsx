// FRM baca Ticket Detail — padanan FRM<mod>.js langit_v2 (window info).
// Modal detail tiket: QR (+ unduh), data pemilik, tombol Request Refund
// bila memenuhi syarat (aktif = belum dipakai/refund/kedaluwarsa).
// Dipasang dengan key={code} oleh induk agar ter-reset tiap ganti tiket.
import { useEffect, useRef, useState } from 'react'
import { Alert, Badge, Button, Modal, SafeImage, Skeleton, useToast } from '../../shared/all.js'
import { formatTime } from '../../../components/format.js'
import { process_refund, ticketDetail, ticketQrUrl } from '../api.js'

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

export default function TicketDetailModal({ code, onClose, onRefunded }) {
  const toast = useToast()
  const [data, setData] = useState(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [refundReason, setRefundReason] = useState('')
  const [refundOpen, setRefundOpen] = useState(false)
  const [refundBusy, setRefundBusy] = useState(false)
  const abortRef = useRef(null)

  useEffect(() => {
    if (!code) return
    abortRef.current?.abort()
    const ctrl = new AbortController()
    abortRef.current = ctrl
    setLoading(true)
    setError('')
    setData(null)
    setRefundOpen(false)
    setRefundReason('')
    ticketDetail(code, ctrl.signal)
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

  function handler_download() {
    try {
      const a = document.createElement('a')
      a.href = ticketQrUrl(code)
      a.download = `${code}.png`
      a.click()
    } catch {
      toast.error('Unduhan gagal dimulai.', { title: 'Error' })
    }
  }

  async function handler_refund_save() {
    if (!refundReason.trim()) {
      toast.warning('Alasan wajib diisi.', { title: 'Batal' })
      return
    }
    setRefundBusy(true)
    try {
      await process_refund(data.order_code, refundReason.trim())
      toast.success('Pengajuan refund dikirim, menunggu keputusan organizer.')
      setRefundOpen(false)
      if (onRefunded) onRefunded()
    } catch (ex) {
      toast.error(ex.message, { title: 'Gagal' })
    } finally {
      setRefundBusy(false)
    }
  }

  return (
    <>
      <Modal open={!!code} onClose={onClose} title={data?.event_title || 'Detail tiket'} size="sm">
        {loading ? (
          <div className="flex flex-col gap-2">
            <Skeleton className="h-44" />
            <Skeleton className="h-6 w-2/3" />
          </div>
        ) : error || !data ? (
          <Alert tone="error" title="Tiket tidak ditemukan" closable onClose={() => setError('')}>
            {error || 'Tiket bukan milik Anda atau sudah tidak ada.'}
          </Alert>
        ) : (
          <div className="flex min-w-0 flex-col items-center gap-3">
            {data.active ? (
              <img src={ticketQrUrl(code)} alt={`QR ${code}`} className="h-44 w-44 rounded-xl border border-zinc-200 dark:border-zinc-700" />
            ) : (
              <SafeImage src="" alt={data.event_title} className="h-44 w-44 rounded-xl object-cover opacity-60" />
            )}
            <p className="font-mono text-sm font-bold text-zinc-900 dark:text-zinc-50">{data.ticket_code}</p>
            <p>
              <Badge tone={statusTone(data.status)}>{data.status}</Badge>
              {!data.active && <Badge tone="neutral" className="ml-1.5">QR nonaktif</Badge>}
            </p>
            <dl className="flex w-full flex-col gap-1.5 text-sm">
              {[
                ['Pemilik', data.holder_name],
                ['Tipe', data.type_name],
                ['Event', data.event_title],
                ['Jadwal', formatTime(data.start_at)],
                ['Venue', [data.venue, data.city].filter(Boolean).join(', ')],
                ['Order', data.order_code],
              ].map(([label, value]) => (
                <div key={label} className="flex items-start justify-between gap-4">
                  <dt className="shrink-0 text-zinc-500 dark:text-zinc-400">{label}</dt>
                  <dd className="min-w-0 break-words text-right font-medium text-zinc-900 dark:text-zinc-50">{value || '—'}</dd>
                </div>
              ))}
            </dl>
            <div className="flex w-full flex-wrap gap-2">
              {data.active && (
                <Button size="sm" variant="secondary" onClick={handler_download}>
                  Download Ticket
                </Button>
              )}
              {data.active && !refundOpen && (
                <Button size="sm" variant="secondary" onClick={() => setRefundOpen(true)}>
                  Request Refund
                </Button>
              )}
            </div>
            {refundOpen && (
              <div className="flex w-full flex-col gap-2 rounded-xl bg-zinc-50 p-3 dark:bg-zinc-800/50">
                <label className="block text-xs font-medium text-zinc-500">
                  Alasan refund
                  <input
                    value={refundReason}
                    onChange={(e) => setRefundReason(e.target.value)}
                    placeholder="Contoh: jadwal bentrok"
                    className="mt-1 w-full rounded-lg border border-zinc-300 bg-white px-2.5 py-1.5 text-sm dark:border-zinc-700 dark:bg-zinc-900"
                  />
                </label>
                <div className="flex gap-2">
                  <Button size="sm" onClick={handler_refund_save} loading={refundBusy}>
                    Kirim
                  </Button>
                  <Button size="sm" variant="ghost" onClick={() => setRefundOpen(false)} disabled={refundBusy}>
                    Batal
                  </Button>
                </div>
              </div>
            )}
          </div>
        )}
      </Modal>
    </>
  )
}
