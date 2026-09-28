// ===== TEMPLATE SERAGAM app/* (ala <mod>.js langit_v2) =====
// Shell + controller standar; yang khas tabel ini hanya COLUMNS (./columns.js).
// Tabel DB: CPSYSLOG via ./api.js listSyslogs/pruneSyslogs.
import { Button, ConfirmDialog, Icon, StandardGrid, StandardPage, useStandardController, useState, useToast } from '../shared/all.js'
import { listSyslogs, pruneSyslogs } from './api.js'
import { LEVELS, SYSLOG_COLUMNS } from './columns.jsx'

const FEATURES = { header: true, refresh: true, filter: true, tabs: false, create: false, edit: false, remove: true, pagination: true, empty: true, error: true, confirmDialog: true, extraActions: true }

const CONFIG = {
  title: 'Error / System Log',
  description: 'Error & kejadian sistem yang ditangkap backend (pengganti mengintip file log di server).',
  errorTitle: 'Gagal memuat log',
  emptyTitle: 'Log bersih',
  emptyDescription: 'Tidak ada catatan pada level ini. Sistem berjalan tanpa error tercatat.',
}

export default function Syslog() {
  const toast = useToast()
  const [level, setLevel] = useState('')
  const [days, setDays] = useState(30)
  const [confirmPrune, setConfirmPrune] = useState(false)

  const { rows: logs, offset, setOffset, loading, showLoading, error, setError, load } =
    useStandardController(listSyslogs, { level })

  async function prune() {
    try {
      const res = await pruneSyslogs(days)
      toast.success(`${res.deleted ?? 0} baris log lebih tua dari ${days} hari dihapus.`)
      setConfirmPrune(false)
      load()
    } catch (err) {
      toast.error(err.message, { title: 'Gagal menghapus' })
    }
  }

  return (
    <div className="flex flex-col gap-4">
      <StandardPage
        title={CONFIG.title}
        description={CONFIG.description}
        features={FEATURES}
        loading={loading}
        showLoading={showLoading}
        error={error}
        errorTitle={CONFIG.errorTitle}
        onClearError={() => setError('')}
        onRefresh={load}
        extraActions={
          <Button variant="danger" size="sm" onClick={() => setConfirmPrune(true)}>
            <Icon name="trash" className="h-4 w-4" />
            Bersihkan lama
          </Button>
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
        emptyTitle={CONFIG.emptyTitle}
        emptyDescription={CONFIG.emptyDescription}
        emptyIcon="terminal"
        offset={offset}
        onPage={setOffset}
      >
        <StandardGrid columns={SYSLOG_COLUMNS} rows={logs} minWidth={560} />
      </StandardPage>

      {FEATURES.confirmDialog && (
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
      )}
    </div>
  )
}
