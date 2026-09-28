// ===== TEMPLATE SERAGAM app/* (ala <mod>.js langit_v2) =====
// Shell + controller standar; body custom (tabs mine/all) seperti modul
// non-grid di langit_v2. Tabel DB: CPREFRESHTOKEN via ./api.js.
import { Badge, Button, ConfirmDialog, Icon, StandardPage, api, formatTime, useStandardController, useState, useToast } from '../shared/all.js'
import { listAllSessions, listMySessions, revokeSession } from './api.js'
import { logoutAll as logoutAllUser } from '../users/api.js'

const FEATURES = { header: true, refresh: true, filter: false, tabs: true, create: false, edit: false, remove: true, pagination: true, empty: true, error: true, confirmDialog: true, extraActions: true }

const CONFIG = {
  title: 'Authentication & Session Management',
  description: 'Setiap login dari perangkat/browser tercatat sebagai 1 sesi (refresh token). Cabut sesi yang tidak dikenal.',
  errorTitle: 'Gagal memuat sesi',
  emptyTitle: 'Tidak ada sesi aktif',
  emptyDescription: 'Semua sesi sudah kedaluwarsa atau dicabut. Login ulang untuk membuat sesi baru.',
}

export const meta = { label: 'Sesi & Auth', icon: 'key', order: 4 }

export default function Sessions() {
  const toast = useToast()
  const me = api.currentUser()
  const canViewAll = true
  const [tab, setTab] = useState('mine')
  const [revoking, setRevoking] = useState(null)

  const { rows: sessions, offset, setOffset, loading, showLoading, error, setError, load } =
    useStandardController(
      ({ limit, offset }) => (tab === 'all' && canViewAll ? listAllSessions(limit, offset) : listMySessions()),
      { tab }
    )

  function switchTab(t) {
    setTab(t)
    setOffset(0)
  }

  async function revoke() {
    if (!revoking) return
    try {
      await revokeSession(revoking.id)
      toast.success('Sesi berhasil dicabut.')
      setRevoking(null)
      load()
    } catch (err) {
      toast.error(err.message, { title: 'Gagal mencabut sesi' })
    }
  }

  async function logoutAllMine() {
    if (!me?.code) {
      toast.error('Profil pengguna belum dimuat.')
      return
    }
    try {
      await logoutAllUser(me.code)
      toast.warning('Semua sesi Anda dicabut. Silakan login ulang.', { title: 'Sesi berakhir' })
      load()
    } catch (err) {
      toast.error(err.message)
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
        tabsBar={
          <div className="mb-4 flex gap-1.5 rounded-xl bg-zinc-100 p-1 dark:bg-zinc-800">
            {['mine', ...(canViewAll ? ['all'] : [])].map((t) => (
              <button
                key={t}
                type="button"
                onClick={() => switchTab(t)}
                className={`flex-1 rounded-lg px-3 py-1.5 text-sm font-semibold transition ${
                  tab === t
                    ? 'bg-white text-zinc-900 shadow-sm dark:bg-zinc-900 dark:text-zinc-50'
                    : 'text-zinc-500 dark:text-zinc-400'
                }`}
              >
                {t === 'mine' ? 'Sesi saya' : 'Semua sesi'}
              </button>
            ))}
          </div>
        }
        items={sessions}
        emptyTitle={CONFIG.emptyTitle}
        emptyDescription={CONFIG.emptyDescription}
        offset={offset}
        onPage={tab === 'all' ? setOffset : null}
        footer={
          tab === 'mine' && !showLoading && sessions.length > 0 ? (
            <div className="mt-4 border-t border-zinc-100 pt-4 dark:border-zinc-800">
              <Button variant="danger" size="sm" onClick={logoutAllMine}>
                <Icon name="logout" className="h-4 w-4" />
                Cabut semua sesi saya
              </Button>
            </div>
          ) : null
        }
      >
        <ul className="flex flex-col gap-2.5">
          {sessions.map((s) => (
            <li
              key={s.id}
              className="flex items-center gap-3 rounded-xl border border-zinc-100 bg-zinc-50/60 p-3 dark:border-zinc-800 dark:bg-zinc-900/40"
            >
              <span className="flex h-10 w-10 shrink-0 items-center justify-center rounded-full bg-zinc-100 text-zinc-500 dark:bg-zinc-800 dark:text-zinc-400">
                <Icon name="key" className="h-5 w-5" />
              </span>
              <div className="min-w-0 flex-1">
                <p className="flex flex-wrap items-center gap-2 text-sm font-semibold text-zinc-900 dark:text-zinc-50">
                  <span className="font-mono">Sesi #{s.id}</span>
                  <Badge tone="success">Aktif</Badge>
                  {tab === 'all' && <span className="truncate text-xs font-normal opacity-70">@{s.username}</span>}
                </p>
                <p className="text-xs text-zinc-500 dark:text-zinc-400">
                  Dibuat {formatTime(s.created_at)} · kedaluwarsa {formatTime(s.expires_at)}
                </p>
              </div>
              <button
                type="button"
                title="Cabut sesi ini"
                onClick={() => setRevoking(s)}
                className="shrink-0 rounded-lg p-2 text-zinc-500 transition hover:bg-rose-50 hover:text-rose-600 dark:hover:bg-rose-950 dark:hover:text-rose-300"
              >
                <Icon name="trash" className="h-4 w-4" />
              </button>
            </li>
          ))}
        </ul>
      </StandardPage>

      {FEATURES.confirmDialog && (
        <ConfirmDialog
          open={!!revoking}
          title={`Cabut sesi #${revoking?.id}?`}
          message="Perangkat pemilik sesi ini akan langsung dikeluarkan dan harus login ulang."
          confirmLabel="Ya, cabut"
          danger
          onConfirm={revoke}
          onCancel={() => setRevoking(null)}
        />
      )}
    </div>
  )
}
