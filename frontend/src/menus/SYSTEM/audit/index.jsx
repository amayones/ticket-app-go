import { useCallback, useEffect, useState } from 'react'
import { listAudit } from './api.js'

export const meta = { label: 'Audit Log', icon: 'list', order: 4 }
import {
  Alert,
  Badge,
  Button,
  Card,
  CardTitle,
  EmptyState,
  Icon,
  Pagination,
  SkeletonRows,
  useSmoothLoading,
  formatTime,
} from '../../../components'

const PAGE_SIZE = 20
const ACTIONS = [
  '',
  'REGISTER',
  'LOGIN',
  'LOGIN_FAILED',
  'LOGOUT',
  'LOGOUT_ALL',
  'UPDATE_USER',
  'DELETE_USER',
  'ROLE_ASSIGN',
  'ROLE_CREATE',
  'ROLE_DELETE',
  'PERMISSION_ASSIGN',
  'SESSION_REVOKE',
  'NOTIF_SEND',
  'TEMPLATE_CREATE',
  'TEMPLATE_UPDATE',
  'TEMPLATE_DELETE',
  'SYSLOG_PRUNE',
]
const ENTITIES = ['', 'AUTH', 'USER', 'ROLE', 'SESSION', 'NOTIFICATION', 'TEMPLATE', 'SYSTEM']

function toneFor(action) {
  if (action.includes('DELETE') || action.includes('FAILED') || action.includes('REVOKE')) return 'danger'
  if (action.includes('CREATE') || (action.includes('LOGIN') && !action.includes('FAILED'))) return 'success'
  if (action.includes('UPDATE') || action.includes('ASSIGN') || action.includes('SEND')) return 'info'
  return 'neutral'
}

const selectClass =
  'rounded-lg border border-zinc-300 bg-white px-2.5 py-1.5 text-sm text-zinc-700 dark:border-zinc-700 dark:bg-zinc-900 dark:text-zinc-200'

export default function Audit() {
  const [logs, setLogs] = useState([])
  const [offset, setOffset] = useState(0)
  const [action, setAction] = useState('')
  const [entity, setEntity] = useState('')
  const [actor, setActor] = useState('')
  const [actorInput, setActorInput] = useState('')
  const [loading, setLoading] = useState(true)
  const showLoading = useSmoothLoading(loading)
  const [error, setError] = useState('')

  const load = useCallback(async () => {
    setLoading(true)
    setError('')
    try {
      setLogs(await listAudit({ action, entity, actor, limit: PAGE_SIZE, offset }))
    } catch (err) {
      setError(err.message)
    } finally {
      setLoading(false)
    }
  }, [action, entity, actor, offset])

  useEffect(() => {
    load()
  }, [load])

  function applyFilter(e) {
    e.preventDefault()
    setOffset(0)
    setActor(actorInput.trim())
  }

  function reset() {
    setAction('')
    setEntity('')
    setActor('')
    setActorInput('')
    setOffset(0)
  }

  return (
    <Card>
      <div className="mb-4 flex flex-wrap items-center justify-between gap-3">
        <CardTitle description="Siapa melakukan apa, kapan, dari IP mana. Tercatat otomatis untuk login, CRUD user/role, sesi, dan notifikasi.">
          Audit Log
        </CardTitle>
        <Button variant="secondary" size="sm" onClick={load} loading={loading}>
          <Icon name="refresh" className="h-4 w-4" />
          Muat ulang
        </Button>
      </div>

      <form onSubmit={applyFilter} className="mb-4 flex flex-wrap items-end gap-2">
        <label className="flex flex-col gap-1 text-xs font-medium text-zinc-500">
          Aksi
          <select value={action} onChange={(e) => { setAction(e.target.value); setOffset(0) }} className={selectClass}>
            {ACTIONS.map((a) => (
              <option key={a} value={a}>{a === '' ? 'Semua' : a}</option>
            ))}
          </select>
        </label>
        <label className="flex flex-col gap-1 text-xs font-medium text-zinc-500">
          Entitas
          <select value={entity} onChange={(e) => { setEntity(e.target.value); setOffset(0) }} className={selectClass}>
            {ENTITIES.map((a) => (
              <option key={a} value={a}>{a === '' ? 'Semua' : a}</option>
            ))}
          </select>
        </label>
        <label className="flex flex-col gap-1 text-xs font-medium text-zinc-500">
          Kode pelaku
          <input
            value={actorInput}
            onChange={(e) => setActorInput(e.target.value)}
            placeholder="USR-…"
            className={`${selectClass} w-32 font-mono`}
          />
        </label>
        <Button size="sm" type="submit">
          Filter
        </Button>
        <Button size="sm" variant="ghost" type="button" onClick={reset}>
          Reset
        </Button>
      </form>

      {error && (
        <Alert tone="error" title="Gagal memuat audit" closable onClose={() => setError('')} className="mb-4">
          {error}
        </Alert>
      )}

      {showLoading ? (
        <SkeletonRows rows={5} />
      ) : logs.length === 0 ? (
        <EmptyState
          title="Belum ada jejak audit"
          description="Sesuaikan filter atau lakukan aksi (mis. login) agar tercatat di sini."
        />
      ) : (
        <div className="overflow-x-auto rounded-xl border border-zinc-100 dark:border-zinc-800">
          <table className="w-full min-w-[640px] text-left text-sm">
            <thead>
              <tr className="bg-zinc-50 text-xs uppercase tracking-wide text-zinc-500 dark:bg-zinc-800/60 dark:text-zinc-400">
                <th className="px-3 py-2.5">Waktu</th>
                <th className="px-3 py-2.5">Pelaku</th>
                <th className="px-3 py-2.5">Aksi</th>
                <th className="px-3 py-2.5">Detail</th>
                <th className="px-3 py-2.5">IP</th>
              </tr>
            </thead>
            <tbody>
              {logs.map((l) => (
                <tr key={l.code} className="border-t border-zinc-100 dark:border-zinc-800">
                  <td className="whitespace-nowrap px-3 py-2.5 text-xs text-zinc-500">{formatTime(l.created_at)}</td>
                  <td className="px-3 py-2.5 font-mono text-xs">{l.actor_code || '—'}</td>
                  <td className="px-3 py-2.5">
                    <Badge tone={toneFor(l.action)}>{l.action}</Badge>
                    <span className="ml-1.5 text-xs text-zinc-500">{l.entity}{l.entity_code ? ` · ${l.entity_code}` : ''}</span>
                  </td>
                  <td className="max-w-[260px] truncate px-3 py-2.5 text-xs text-zinc-600 dark:text-zinc-300" title={l.detail}>
                    {l.detail || '—'}
                  </td>
                  <td className="px-3 py-2.5 font-mono text-xs text-zinc-500">{l.ip_address || '—'}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      {!showLoading && logs.length > 0 && (
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
    </Card>
  )
}
