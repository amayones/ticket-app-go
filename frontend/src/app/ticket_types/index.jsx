// View Ticket Types & Pricing — daftar tipe tiket satu event + ringkasan.
// Masuk dari My Events (hash #/ticket_types/<eventCode>).
// Controller: ./controller.js (6 fungsi). Store: read_data.
import { useEffect, useRef, useState } from 'react'
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
  useMessageBox,
  useSmoothLoading,
  useToast,
  hashParam,
  clearHash,
} from '../shared/all.js'
import { process_delete, publishEvent, read_data, reorderTypes, setTypeActive } from './api.js'
import { controller } from './controller.js'
import TicketTypeModal from './components/TicketTypeModal.jsx'

export const meta = { label: 'Ticket Types', icon: 'list', order: 3 }

function rupiah(value) {
  const s = formatAmount(value)
  return s === '' ? 'Rp 0' : `Rp ${s}`
}

export default function TicketTypes({ nvdata, onNavigate }) {
  const toast = useToast()
  const { MessageBox, dialog } = useMessageBox()
  const [eventCode] = useState(() => hashParam('ticket_types'))
  const [showFRM, setShowFRM] = useState(false)
  const [editing, setEditing] = useState(null)
  const abortRef = useRef(null)
  const [data, setData] = useState(null)
  const [loading, setLoading] = useState(true)
  const showLoading = useSmoothLoading(loading)
  const [error, setError] = useState('')

  async function load() {
    if (!eventCode) {
      setLoading(false)
      return
    }
    abortRef.current?.abort()
    const ctrl = new AbortController()
    abortRef.current = ctrl
    setLoading(true)
    setError('')
    try {
      const result = await read_data({ eventCode }, ctrl.signal)
      if (!ctrl.signal.aborted) setData(result)
    } catch (err) {
      if (err && err.name === 'AbortError') return
      if (!ctrl.signal.aborted) setError(err.message)
    } finally {
      if (!ctrl.signal.aborted) setLoading(false)
    }
  }

  useEffect(() => {
    controller.init()
    load()
    return () => abortRef.current?.abort()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  function handler_back() {
    clearHash()
    if (onNavigate) onNavigate('my_events')
  }

  function handler_move(code, dir) {
    const order = (data?.types || []).map((t) => t.code)
    const i = order.indexOf(code)
    const j = i + dir
    if (i < 0 || j < 0 || j >= order.length) return
    const next = [...order]
    next[i] = order[j]
    next[j] = order[i]
    reorderTypes(eventCode, next)
      .then(() => load())
      .catch((ex) => toast.error(ex.message, { title: 'Error' }))
  }

  async function handler_toggle_active(row) {
    try {
      const ok = await MessageBox.update({
        headline: row.is_active ? 'Nonaktifkan tipe?' : 'Aktifkan tipe?',
        details: [{ label: 'Tipe', value: row.name }],
        confirmLabel: row.is_active ? 'Ya, nonaktifkan' : 'Ya, aktifkan',
        danger: row.is_active,
        request: () => setTypeActive(eventCode, row.code, !row.is_active),
      })
      if (!ok) return
      toast.success(`Tipe ${row.name} ${row.is_active ? 'dinonaktifkan' : 'diaktifkan'}.`)
      load()
    } catch (ex) {
      toast.error(ex.message, { title: 'Error' })
    }
  }

  async function handler_delete_click(row) {
    try {
      const result = await MessageBox.delete({
        headline: 'Hapus tipe tiket?',
        details: [
          { label: 'Tipe', value: row.name },
          { label: 'Terjual', value: `${row.sold} tiket` },
        ],
        request: () => process_delete(eventCode, row.code),
      })
      if (!result) return
      toast.success(`Tipe ${row.name} dihapus.`)
      load()
    } catch (ex) {
      toast.error(ex.message, { title: 'Error' })
    }
  }

  async function handler_publish_click() {
    try {
      await publishEvent(eventCode)
      toast.success('Event dipublish.')
      load()
    } catch (ex) {
      const missing = ex && ex.data && Array.isArray(ex.data.missing) ? ex.data.missing : null
      if (missing) {
        toast.warning(`Belum bisa publish: ${missing.join('; ')}`, { title: 'Belum lengkap' })
        return
      }
      toast.error(ex.message, { title: 'Error' })
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

  const types = data?.types || []
  const sum = data?.summary || {}
  const items = data ? [data] : []

  const COLUMNS = [
    { header: 'Nama', dataIndex: 'name', width: 150, render: (t) => <span className="font-semibold">{t.name}</span> },
    {
      header: 'Harga',
      dataIndex: 'price',
      width: 130,
      render: (t) => <span className="font-mono">{t.price === 0 ? 'Gratis' : rupiah(t.price)}</span>,
    },
    {
      header: 'Kuota',
      dataIndex: 'quota',
      width: 130,
      render: (t) => (
        <span className="font-mono">
          {t.sold}/{t.quota} (sisa {t.remaining})
        </span>
      ),
    },
    { header: 'Batas/orang', dataIndex: 'max_per_person', width: 100, render: (t) => (t.max_per_person > 0 ? t.max_per_person : '—') },
    { header: 'Status', dataIndex: 'is_active', width: 100, render: (t) => <Badge tone={t.is_active ? 'success' : 'neutral'}>{t.is_active ? 'Aktif' : 'Nonaktif'}</Badge> },
    {
      header: 'Aksi',
      dataIndex: 'code',
      width: 200,
      render: (t) => (
        <span className="flex gap-1">
          <Tooltip label="Naik">
            <button type="button" aria-label="Naik" onClick={() => handler_move(t.code, -1)} className="rounded-lg p-1.5 text-zinc-500 hover:bg-zinc-100 dark:hover:bg-zinc-800">
              <Icon name="chevronLeft" className="h-4 w-4 -rotate-90" />
            </button>
          </Tooltip>
          <Tooltip label="Turun">
            <button type="button" aria-label="Turun" onClick={() => handler_move(t.code, 1)} className="rounded-lg p-1.5 text-zinc-500 hover:bg-zinc-100 dark:hover:bg-zinc-800">
              <Icon name="chevronRight" className="h-4 w-4 -rotate-90" />
            </button>
          </Tooltip>
          <Tooltip label="Ubah">
            <button
              type="button"
              aria-label={`Ubah ${t.name}`}
              onClick={() => {
                setEditing(t)
                setShowFRM(true)
              }}
              className="rounded-lg p-1.5 text-zinc-500 hover:bg-zinc-100 dark:hover:bg-zinc-800"
            >
              <Icon name="pencil" className="h-4 w-4" />
            </button>
          </Tooltip>
          <Tooltip label={t.is_active ? 'Nonaktifkan' : 'Aktifkan'}>
            <button
              type="button"
              aria-label="Aktif"
              onClick={() => handler_toggle_active(t)}
              className="rounded-lg p-1.5 text-zinc-500 hover:bg-zinc-100 dark:hover:bg-zinc-800"
            >
              <Icon name={t.is_active ? 'eyeOff' : 'eye'} className="h-4 w-4" />
            </button>
          </Tooltip>
          <Tooltip label="Hapus">
            <button
              type="button"
              aria-label={`Hapus ${t.name}`}
              onClick={() => handler_delete_click(t)}
              className="rounded-lg p-1.5 text-zinc-500 hover:bg-rose-50 hover:text-rose-600 dark:hover:bg-rose-950"
            >
              <Icon name="trash" className="h-4 w-4" />
            </button>
          </Tooltip>
        </span>
      ),
    },
  ]

  return (
    <div className="flex flex-col gap-4">
      <StandardPage
        title={nvdata?.label || 'Ticket Types'}
        description={sum.event_title ? `${sum.event_title} (${sum.event_status || ''})` : 'Tipe tiket event.'}
        loading={loading}
        showLoading={showLoading}
        error={error}
        errorTitle="Gagal memuat tipe tiket"
        onClearError={() => setError('')}
        onRefresh={load}
        onCreate={() => {
          setEditing(null)
          setShowFRM(true)
        }}
        createLabel="Tambah Tipe"
        extraActions={
          <>
            <Button variant="secondary" size="sm" onClick={handler_back}>
              Kembali
            </Button>
            {sum.event_status === 'DRAFT' && (
              <Button size="sm" onClick={handler_publish_click}>
                Publish Event
              </Button>
            )}
          </>
        }
        items={items}
        emptyTitle="Belum ada tipe tiket"
        emptyDescription="Tambahkan minimal 1 tipe aktif agar event bisa publish."
        offset={0}
        onPage={null}
      >
        <div className="mb-4 grid gap-3 sm:grid-cols-3">
          <div className="rounded-xl bg-zinc-50 p-3 text-sm dark:bg-zinc-800/50">
            <p className="text-xs text-zinc-500">Total kuota</p>
            <p className="text-lg font-bold">{Number(sum.total_quota || 0).toLocaleString('id-ID')}</p>
          </div>
          <div className="rounded-xl bg-zinc-50 p-3 text-sm dark:bg-zinc-800/50">
            <p className="text-xs text-zinc-500">Total terjual</p>
            <p className="text-lg font-bold">{Number(sum.total_sold || 0).toLocaleString('id-ID')}</p>
          </div>
          <div className="rounded-xl bg-zinc-50 p-3 text-sm dark:bg-zinc-800/50">
            <p className="text-xs text-zinc-500">Potensi pendapatan</p>
            <p className="text-lg font-bold">{rupiah(sum.potential || 0)}</p>
          </div>
        </div>
        {(sum.missing || []).length > 0 && sum.event_status === 'DRAFT' && (
          <Alert tone="warning" title="Belum bisa publish" className="mb-4">
            {(sum.missing || []).join('; ')}
          </Alert>
        )}
        <StandardGrid columns={COLUMNS} rows={types} minWidth={760} />
      </StandardPage>

      <TicketTypeModal
        open={showFRM}
        eventCode={eventCode}
        editing={editing}
        onClose={() => {
          setShowFRM(false)
          setEditing(null)
        }}
        onSaved={load}
      />
      {dialog}

      {error && (
        <Alert tone="error" title="Gagal memuat tipe tiket" closable onClose={() => setError('')}>
          {error}
        </Alert>
      )}
    </div>
  )
}
