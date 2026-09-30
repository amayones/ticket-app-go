// GRID My Events — padanan GRID<mod>.js langit_v2 (kartu per event).
// Handler nama SAMA: handler_rowbtn_edit_click, handler_publish_click,
// handler_unpublish_click, handler_cancel_click, handler_delete_click,
// handler_duplicate_click. Pintasan: tiket & peserta via onOpen...(code).
import { forwardRef, useImperativeHandle, useState } from 'react'
import { Alert, Badge, Button, Icon, Modal, TextField, Tooltip, useMessageBox, useToast } from '../shared/all.js'
import SafeImage from '../shared/SafeImage.jsx'
import { ProgressBar } from '../shared/Chart.jsx'
import { formatAmount, formatTime } from '../../components/format.js'
import { process_cancel, process_delete, process_duplicate, process_publish, process_unpublish } from './api.js'

function rupiah(value) {
  const s = formatAmount(value)
  return s === '' ? 'Rp 0' : `Rp ${s}`
}

function statusTone(status) {
  switch (status) {
    case 'PUBLISHED':
      return 'success'
    case 'DRAFT':
      return 'warning'
    case 'CANCELLED':
      return 'danger'
    default:
      return 'neutral'
  }
}

export default forwardRef(function GRID({ rows, reload, openFRM, openEdit, onOpenTypes, onOpenAttendees }, ref) {
  const toast = useToast()
  const { MessageBox, dialog } = useMessageBox()
  const [cancelTarget, setCancelTarget] = useState(null)
  const [cancelReason, setCancelReason] = useState('')
  const [cancelError, setCancelError] = useState('')
  const [cancelBusy, setCancelBusy] = useState(false)

  useImperativeHandle(ref, () => ({
    // Varian A: controller.btnew_click delegasi ke sini — buka FRM input baru.
    handler_btnew_click: () => openFRM(),
  }))

  async function confirmAction({ headline, details, danger, request, success }) {
    try {
      const result = await MessageBox.delete({
        headline,
        details,
        confirmLabel: 'Ya, lanjutkan',
        danger: !!danger,
        request,
      })
      if (!result) return
      toast.success(success)
      reload()
    } catch (ex) {
      // Backend 409 publish membawa missing[] di error.data (client.js).
      const missing = ex && ex.data && Array.isArray(ex.data.missing) ? ex.data.missing : null
      if (missing) {
        toast.warning(`Belum bisa publish: ${missing.join('; ')}`, { title: 'Belum lengkap' })
        return
      }
      toast.error(ex.message, { title: 'Error' })
    }
  }

  function handler_publish_click(event) {
    confirmAction({
      headline: 'Publish event?',
      details: [
        { label: 'Event', value: event.title },
        { label: 'Mulai', value: formatTime(event.start_at) },
      ],
      request: () => process_publish(event.code),
      success: `Event ${event.title} tayang.`,
    })
  }

  function handler_unpublish_click(event) {
    confirmAction({
      headline: 'Kembalikan ke draft?',
      details: [
        { label: 'Event', value: event.title },
        { label: 'Terjual', value: `${event.sold} tiket` },
      ],
      danger: true,
      request: () => process_unpublish(event.code),
      success: `Event ${event.title} kembali ke draft.`,
    })
  }

  function handler_cancel_click(event) {
    setCancelTarget(event)
    setCancelReason('')
    setCancelError('')
  }

  async function handler_cancel_save() {
    if (!cancelTarget) return
    if (!cancelReason.trim()) {
      setCancelError('Alasan wajib diisi.')
      return
    }
    setCancelBusy(true)
    try {
      const result = await MessageBox.delete({
        headline: 'Batalkan event?',
        details: [
          { label: 'Event', value: cancelTarget.title },
          { label: 'Alasan', value: cancelReason.trim() },
        ],
        confirmLabel: 'Ya, batalkan',
        danger: true,
        request: () => process_cancel(cancelTarget.code, cancelReason.trim()),
      })
      if (!result) return
      const affected = result.data && result.data.affected_buyers
      setCancelTarget(null)
      toast.warning(`Event dibatalkan.${affected ? ` ${affected} pembeli terdampak (refund di module lain).` : ''}`, {
        title: 'Event dibatalkan',
      })
      reload()
    } catch (ex) {
      setCancelError(ex.message)
    } finally {
      setCancelBusy(false)
    }
  }

  function handler_delete_click(event) {
    confirmAction({
      headline: 'Hapus event?',
      details: [
        { label: 'Event', value: event.title },
        { label: 'Status', value: 'draft, belum pernah tayang' },
      ],
      danger: true,
      request: () => process_delete(event.code),
      success: `Event ${event.title} dihapus.`,
    })
  }

  function handler_duplicate_click(event) {
    confirmAction({
      headline: 'Duplikat event?',
      details: [
        { label: 'Event', value: event.title },
        { label: 'Hasil', value: 'draft baru + salinan tipe tiket' },
      ],
      request: () => process_duplicate(event.code),
      success: `Duplikat ${event.title} dibuat sebagai draft.`,
    })
  }

  return (
    <>
      <ul className="flex flex-col gap-2.5">
        {rows.map((e) => {
          const quota = Math.max(1, e.quota || 0)
          return (
            <li
              key={e.code}
              className="flex flex-col gap-3 rounded-xl border border-zinc-100 bg-zinc-50/60 p-3 transition hover:border-violet-200 hover:bg-violet-50/50 sm:flex-row sm:items-center dark:border-zinc-800 dark:bg-zinc-900/40 dark:hover:border-violet-800"
            >
              <SafeImage src={e.poster_url} alt={e.title} className="h-20 w-full shrink-0 rounded-lg object-cover sm:w-32" />
              <div className="min-w-0 flex-1">
                <p className="flex flex-wrap items-center gap-2 text-sm font-semibold text-zinc-900 dark:text-zinc-50">
                  <span className="truncate">{e.title}</span>
                  <Badge tone={statusTone(e.status)}>{e.status}</Badge>
                </p>
                <p className="truncate text-xs text-zinc-500 dark:text-zinc-400">
                  {[e.category, e.city].filter(Boolean).join(' · ')} · {formatTime(e.start_at)}
                </p>
                <div className="mt-1.5 flex items-center gap-2">
                  <div className="min-w-0 flex-1">
                    <ProgressBar value={e.sold} max={quota} tone={e.quota > 0 && e.sold / e.quota >= 0.9 ? 'amber' : 'violet'} />
                  </div>
                  <span className="shrink-0 font-mono text-[11px] text-zinc-500">
                    {e.sold}/{e.quota} · {rupiah(e.revenue)}
                  </span>
                </div>
              </div>
              <div className="flex shrink-0 flex-wrap items-center gap-1">
                <Tooltip label="Ubah">
                  <button
                    type="button"
                    aria-label={`Ubah ${e.title}`}
                    onClick={() => openEdit(e.code)}
                    className="rounded-lg p-2 text-zinc-500 transition-colors hover:bg-white hover:text-zinc-900 hover:shadow-sm dark:hover:bg-zinc-800 dark:hover:text-zinc-100"
                  >
                    <Icon name="pencil" className="h-4 w-4" />
                  </button>
                </Tooltip>
                <Tooltip label="Tipe tiket">
                  <button
                    type="button"
                    aria-label={`Tipe tiket ${e.title}`}
                    onClick={() => onOpenTypes(e.code)}
                    className="rounded-lg p-2 text-zinc-500 transition-colors hover:bg-white hover:text-zinc-900 hover:shadow-sm dark:hover:bg-zinc-800 dark:hover:text-zinc-100"
                  >
                    <Icon name="list" className="h-4 w-4" />
                  </button>
                </Tooltip>
                <Tooltip label="Peserta">
                  <button
                    type="button"
                    aria-label={`Peserta ${e.title}`}
                    onClick={() => onOpenAttendees(e.code)}
                    className="rounded-lg p-2 text-zinc-500 transition-colors hover:bg-white hover:text-zinc-900 hover:shadow-sm dark:hover:bg-zinc-800 dark:hover:text-zinc-100"
                  >
                    <Icon name="users" className="h-4 w-4" />
                  </button>
                </Tooltip>
                <Tooltip label="Duplikat">
                  <button
                    type="button"
                    aria-label={`Duplikat ${e.title}`}
                    onClick={() => handler_duplicate_click(e)}
                    className="rounded-lg p-2 text-zinc-500 transition-colors hover:bg-white hover:text-zinc-900 hover:shadow-sm dark:hover:bg-zinc-800 dark:hover:text-zinc-100"
                  >
                    <Icon name="folder" className="h-4 w-4" />
                  </button>
                </Tooltip>
                {e.status === 'DRAFT' && (
                  <Button size="sm" onClick={() => handler_publish_click(e)}>
                    Publish
                  </Button>
                )}
                {e.status === 'PUBLISHED' && (
                  <Button size="sm" variant="secondary" onClick={() => handler_unpublish_click(e)}>
                    Unpublish
                  </Button>
                )}
                {(e.status === 'DRAFT' || e.status === 'PUBLISHED') && (
                  <Button size="sm" variant="secondary" onClick={() => handler_cancel_click(e)}>
                    Batal
                  </Button>
                )}
                {e.status === 'DRAFT' && (
                  <Tooltip label="Hapus (draft belum tayang)">
                    <button
                      type="button"
                      aria-label={`Hapus ${e.title}`}
                      onClick={() => handler_delete_click(e)}
                      className="rounded-lg p-2 text-zinc-500 transition-colors hover:bg-rose-50 hover:text-rose-600 hover:shadow-sm dark:hover:bg-rose-950 dark:hover:text-rose-300"
                    >
                      <Icon name="trash" className="h-4 w-4" />
                    </button>
                  </Tooltip>
                )}
              </div>
            </li>
          )
        })}
      </ul>
      {dialog}
      <Modal
        open={!!cancelTarget}
        onClose={cancelBusy ? undefined : () => setCancelTarget(null)}
        title="Alasan pembatalan"
        size="sm"
        closeOnBackdrop={!cancelBusy}
        footer={
          <>
            <Button variant="secondary" onClick={() => setCancelTarget(null)} disabled={cancelBusy}>
              Batal
            </Button>
            <Button variant="danger" onClick={handler_cancel_save} loading={cancelBusy}>
              Lanjut
            </Button>
          </>
        }
      >
        <div className="flex flex-col gap-3">
          {cancelError && (
            <Alert tone="error" closable onClose={() => setCancelError('')}>
              {cancelError}
            </Alert>
          )}
          <TextField
            label="Alasan"
            value={cancelReason}
            onChange={(e) => setCancelReason(e.target.value)}
            placeholder="Contoh: izin keramaian tidak turun"
            hint={cancelTarget ? `Event: ${cancelTarget.title}` : ''}
          />
        </div>
      </Modal>
    </>
  )
})
