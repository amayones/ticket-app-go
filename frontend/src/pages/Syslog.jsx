import { useCallback, useEffect, useState } from 'react'
import { api } from '../api/client.js'
import {
  Alert,
  Badge,
  Button,
  Card,
  CardTitle,
  ConfirmDialog,
  EmptyState,
  Icon,
  Pagination,
  SkeletonRows,
  useToast,
} from '../components'

const PAGE_SIZE = 20
const LEVELS = ['', 'ERROR', 'WARN', 'INFO']

function toneFor(level) {
  if (level === 'ERROR') return 'danger'
  if (level === 'WARN') return 'warning'
  return 'info'
}

function formatTime(iso) {
  try {
    return new Date(iso).toLocaleString('id-ID', { dateStyle: 'medium', timeStyle: 'short' })
  } catch {
    return iso
  }
}

export default function Syslog() {
  const toast = useToast()
  const [logs, setLogs] = useState([])
  const [offset, setOffset] = useState(0)
  const [level, setLevel] = useState('')
  const [days, setDays] = useState(30)
  const [confirmPrune, setConfirmPrune] = useState(false)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  const load = useCallback(async () => {
    setLoading(true)
    setError('')
    try {
      setLogs(await api.listSyslogs({ level, limit: PAGE_SIZE, offset }))
    } catch (err) {
      setError(err.message)
    } finally {
      setLoading(false)
    }
  }, [level, offset])

  useEffect(() => {
    load()
  }, [load])

  async function prune() {
    try {
      const res = await api.pruneSyslogs(days)
      toast.success(`${res.deleted ?? 0} baris log lebih tua dari ${days} hari dihapus.`)
      setConfirmPrune(false)
      load()
    } catch (err) {
      toast.error(err.message, { title: 'Gagal menghapus' })
    }
  }

  return (
    <Card>
      <div className="mb-4 flex flex-wrap items-center justify-between gap-3">
        <CardTitle description="Error & kejadian sistem yang ditangkap backend (pengganti mengintip file log di server).">
          Error / System Log
        </CardTitle>
        <div className="flex items-center gap-2">
          <Button variant="secondary" size="sm" onClick={load} loading={loading}>
            <Icon name="refresh" className="h-4 w-4" />
            Muat ulang
          </Button>
          <Button variant="danger" size="sm" onClick={() => setConfirmPrune(true)}>
            <Icon name="trash" className="h-4 w-4" />
            Bersihkan lama
          </Button>
        </div>
      </div>

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

      {error && (
        <Alert tone="error" title="Gagal memuat log" closable onClose={() => setError('')} className="mb-4">
          {error}
        </Alert>
      )}

      {loading ? (
        <SkeletonRows rows={5} />
      ) : logs.length === 0 ? (
        <EmptyState
          icon="terminal"
          title="Log bersih"
          description="Tidak ada catatan pada level ini. Sistem berjalan tanpa error tercatat."
        />
      ) : (
        <div className="overflow-x-auto rounded-xl border border-zinc-100 dark:border-zinc-800">
          <table className="w-full min-w-[560px] text-left text-sm">
            <thead>
              <tr className="bg-zinc-50 text-xs uppercase tracking-wide text-zinc-500 dark:bg-zinc-800/60 dark:text-zinc-400">
                <th className="px-3 py-2.5">Waktu</th>
                <th className="px-3 py-2.5">Level</th>
                <th className="px-3 py-2.5">Sumber</th>
                <th className="px-3 py-2.5">Pesan</th>
              </tr>
            </thead>
            <tbody>
              {logs.map((l) => (
                <tr key={l.code} className="border-t border-zinc-100 dark:border-zinc-800">
                  <td className="whitespace-nowrap px-3 py-2.5 font-mono text-xs text-zinc-500">{formatTime(l.created_at)}</td>
                  <td className="px-3 py-2.5">
                    <Badge tone={toneFor(l.level)}>{l.level}</Badge>
                  </td>
                  <td className="px-3 py-2.5 font-mono text-xs text-zinc-600 dark:text-zinc-300">{l.source}</td>
                  <td className="max-w-[320px] truncate px-3 py-2.5 text-xs text-zinc-600 dark:text-zinc-300" title={l.message}>
                    {l.message}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      {!loading && logs.length > 0 && (
        <div className="mt-4 border-t border-zinc-100 pt-4 dark:border-zinc-800">
          <Pagination
            offset={offset}
            limit={PAGE_SIZE}
            count={logs.length}
            hasMore={logs.length === PAGE_SIZE}
            loading={loading}
            onPage={setOffset}
          />
        </div>
      )}

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
    </Card>
  )
}
