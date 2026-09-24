import { useCallback, useEffect, useState } from 'react'
import { api } from '../../../api/client.js'
import { listAllSessions, listMySessions, revokeSession } from './api.js'
import { logoutAll } from '../users/api.js'

export const meta = { label: 'Sesi & Auth', icon: 'key', order: 3 }
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
  useSmoothLoading,
  useToast,
} from '../../../components'

const PAGE_SIZE = 15

function formatTime(iso) {
  try {
    return new Date(iso).toLocaleString('id-ID', { dateStyle: 'medium', timeStyle: 'short' })
  } catch {
    return iso
  }
}

export default function Sessions() {
  const toast = useToast()
  const me = api.currentUser()
  const canViewAll = true
  const [tab, setTab] = useState('mine')
  const [sessions, setSessions] = useState([])
  const [offset, setOffset] = useState(0)
  const [loading, setLoading] = useState(true)
  const showLoading = useSmoothLoading(loading)
  const [error, setError] = useState('')
  const [revoking, setRevoking] = useState(null)

  const load = useCallback(async () => {
    setLoading(true)
    setError('')
    try {
      if (tab === 'all' && canViewAll) {
        setSessions(await listAllSessions(PAGE_SIZE, offset))
      } else {
        setSessions(await listMySessions())
      }
    } catch (err) {
      setError(err.message)
    } finally {
      setLoading(false)
    }
  }, [tab, offset, canViewAll])

  useEffect(() => {
    load()
  }, [load])

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

  async function logoutAll() {
    try {
      await logoutAll(me.code)
      toast.warning('Semua sesi Anda dicabut. Silakan login ulang.', { title: 'Sesi berakhir' })
      load()
    } catch (err) {
      toast.error(err.message)
    }
  }

  return (
    <div className="flex flex-col gap-4">
      <Card>
        <div className="mb-4 flex flex-wrap items-center justify-between gap-3">
          <CardTitle description="Setiap login dari perangkat/browser tercatat sebagai 1 sesi (refresh token). Cabut sesi yang tidak dikenal.">
            Authentication & Session Management
          </CardTitle>
          <Button variant="secondary" size="sm" onClick={load} loading={loading}>
            <Icon name="refresh" className="h-4 w-4" />
            Muat ulang
          </Button>
        </div>

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

        {error && (
          <Alert tone="error" title="Gagal memuat sesi" closable onClose={() => setError('')} className="mb-4">
            {error}
          </Alert>
        )}

        {showLoading ? (
          <SkeletonRows rows={3} />
        ) : sessions.length === 0 ? (
          <EmptyState
            title="Tidak ada sesi aktif"
            description="Semua sesi sudah kedaluwarsa atau dicabut. Login ulang untuk membuat sesi baru."
          />
        ) : (
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
                  <Icon name="trash" className="h-4.5 w-4.5" />
                </button>
              </li>
            ))}
          </ul>
        )}

        {tab === 'all' && !showLoading && sessions.length > 0 && (
          <div className="mt-4 border-t border-zinc-100 pt-4 dark:border-zinc-800">
            <Pagination
              offset={offset}
              limit={PAGE_SIZE}
              count={sessions.length}
              hasMore={sessions.length === PAGE_SIZE}
              loading={loading}
              onPage={setOffset}
            />
          </div>
        )}

        {tab === 'mine' && !showLoading && sessions.length > 0 && (
          <div className="mt-4 border-t border-zinc-100 pt-4 dark:border-zinc-800">
            <Button variant="danger" size="sm" onClick={logoutAll}>
              <Icon name="logout" className="h-4 w-4" />
              Cabut semua sesi saya
            </Button>
          </div>
        )}
      </Card>

      <ConfirmDialog
        open={!!revoking}
        title={`Cabut sesi #${revoking?.id}?`}
        message="Perangkat pemilik sesi ini akan langsung dikeluarkan dan harus login ulang."
        confirmLabel="Ya, cabut"
        danger
        onConfirm={revoke}
        onCancel={() => setRevoking(null)}
      />
    </div>
  )
}
