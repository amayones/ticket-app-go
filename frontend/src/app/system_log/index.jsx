// View System Log — padanan modul viewer doc_*/mapp_* langit_v2
// (toolbar refresh saja, tanpa btnew/FRM; aksi khusus prune via extraActions).
// Controller: ./controller.js (6 fungsi). Store: read_data. Tabel: CPSYSLOG.
import { useEffect, useState } from 'react'
import { Badge, Button, ConfirmDialog, Icon, StandardGrid, StandardPage, DownloadButton, formatTime, useStandardController, useToast } from '../shared/all.js'
import { read_data, pruneSyslogs } from './api.js'
import { controller } from './controller.js'

export const meta = { label: 'System Log', icon: 'terminal', order: 7 }

const LEVELS = ['', 'ERROR', 'WARN', 'INFO']

function toneFor(level) {
  if (level === 'ERROR') return 'danger'
  if (level === 'WARN') return 'warning'
  return 'info'
}

const COLUMNS_ITEMS = [
  {
    header: 'Waktu',
    dataIndex: 'created_at',
    width: 130,
    render: (l) => <span className="whitespace-nowrap font-mono text-zinc-500">{formatTime(l.created_at)}</span>,
  },
  {
    header: 'Level',
    dataIndex: 'level',
    width: 90,
    render: (l) => <Badge tone={toneFor(l.level)}>{l.level}</Badge>,
  },
  {
    header: 'Sumber',
    dataIndex: 'source',
    width: 140,
    render: (l) => <span className="font-mono">{l.source}</span>,
  },
  {
    header: 'Pesan',
    dataIndex: 'message',
    width: 320,
    render: (l) => (
      <span className="block max-w-[320px] truncate" title={l.message}>
        {l.message}
      </span>
    ),
  },
]

export default function Syslog({ nvdata }) {
  const toast = useToast()
  const [level, setLevel] = useState('')
  const [days, setDays] = useState(30)
  const [confirmPrune, setConfirmPrune] = useState(false)

  const { rows: logs, offset, setOffset, loading, showLoading, error, setError, load } =
    useStandardController(read_data, { level })

  useEffect(() => {
    controller.init()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  async function prune() {
    try {
      const res = await pruneSyslogs(days)
      toast.success(`${res.deleted ?? 0} baris log lebih tua dari ${days} hari dihapus.`)
      setConfirmPrune(false)
      controller.btrefresh_click(load)
    } catch (err) {
      toast.error(err.message, { title: 'Gagal menghapus' })
    }
  }

  return (
    <div className="flex flex-col gap-4">
      <StandardPage
        title={nvdata?.label || 'Error / System Log'}
        description="Error & kejadian sistem yang ditangkap backend (pengganti mengintip file log di server)."
        loading={loading}
        showLoading={showLoading}
        error={error}
        errorTitle="Gagal memuat log"
        onClearError={() => setError('')}
        onRefresh={() => controller.btrefresh_click(load)}
        extraActions={
          <>
            <DownloadButton rows={logs} columns={COLUMNS_ITEMS} filename="system_log" />
            <Button variant="danger" size="sm" onClick={() => setConfirmPrune(true)}>
              <Icon name="trash" className="h-4 w-4" />
              Bersihkan lama
            </Button>
          </>
        }
        filterBar={
          <div className="mb-4 flex gap-1.5">
            {LEVELS.map((l) => (
              <button
                key={l}
                type="button"
                onClick={() => { setLevel(l); setOffset(0) }}
                className={`rounded-lg px-3 py-1.5 text-xs font-bold transition ${
                  level === l
                    ? 'bg-zinc-900 text-white dark:bg-zinc-100 dark:text-zinc-900'
                    : 'bg-zinc-100 text-zinc-500 hover:bg-zinc-200 dark:bg-zinc-800 dark:text-zinc-400'
                }`}
              >
                {l === '' ? 'SEMUA' : l}
              </button>
            ))}
          </div>
        }
        items={logs}
        emptyTitle="Log bersih"
        emptyDescription="Tidak ada catatan pada level ini. Sistem berjalan tanpa error tercatat."
        emptyIcon="terminal"
        offset={offset}
        onPage={setOffset}
      >
        <StandardGrid columns={COLUMNS_ITEMS} rows={logs} minWidth={560} />
      </StandardPage>

      <ConfirmDialog
        open={confirmPrune}
        title="Hapus log lama?"
        message={
          <span className="flex flex-col gap-3">
            <span>Hapus permanen semua log lebih tua dari jumlah hari berikut:</span>
            <input
              type="number"
              min={1}
              max={365}
              value={days}
              onChange={(e) => setDays(Number(e.target.value) || 30)}
              className="w-28 rounded-lg border border-zinc-300 px-2.5 py-1.5 text-sm dark:border-zinc-700 dark:bg-zinc-900"
            />
          </span>
        }
        confirmLabel="Ya, hapus"
        danger
        onConfirm={prune}
        onCancel={() => setConfirmPrune(false)}
      />
    </div>
  )
}
