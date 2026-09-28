// View Audit Log — padanan modul viewer doc_*/mapp_* langit_v2
// (toolbar refresh saja, tanpa btnew/FRM).
// Controller: ./controller.js (6 fungsi). Store: read_data. Tabel: CPAUDITLOG.
import { useEffect, useState } from 'react'
import { Badge, Button, StandardGrid, StandardPage, DownloadButton, SELECT_CLASS, formatTime, useStandardController } from '../shared/all.js'
import { read_data } from './api.js'
import { controller } from './controller.js'

export const meta = { label: 'Audit Log', icon: 'list', order: 5 }

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

const COLUMNS_ITEMS = [
  {
    header: 'Waktu',
    dataIndex: 'created_at',
    width: 130,
    render: (l) => <span className="whitespace-nowrap text-zinc-500">{formatTime(l.created_at)}</span>,
  },
  {
    header: 'Pelaku',
    dataIndex: 'actor_code',
    width: 110,
    render: (l) => <span className="font-mono">{l.actor_code || '—'}</span>,
  },
  {
    header: 'Aksi',
    dataIndex: 'action',
    width: 200,
    render: (l) => (
      <span>
        <Badge tone={toneFor(l.action)}>{l.action}</Badge>
        <span className="ml-1.5 text-zinc-500">{l.entity}{l.entity_code ? ` · ${l.entity_code}` : ''}</span>
      </span>
    ),
  },
  {
    header: 'Detail',
    dataIndex: 'detail',
    width: 260,
    render: (l) => (
      <span className="block max-w-[260px] truncate" title={l.detail}>
        {l.detail || '—'}
      </span>
    ),
  },
  {
    header: 'IP',
    dataIndex: 'ip_address',
    width: 110,
    render: (l) => <span className="font-mono text-zinc-500">{l.ip_address || '—'}</span>,
  },
]

export default function Audit({ nvdata }) {
  const [action, setAction] = useState('')
  const [entity, setEntity] = useState('')
  const [actor, setActor] = useState('')
  const [actorInput, setActorInput] = useState('')

  const { rows: logs, offset, setOffset, loading, showLoading, error, setError, load } =
    useStandardController(read_data, { action, entity, actor })

  useEffect(() => {
    controller.init()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

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
    <StandardPage
      title={nvdata?.label || 'Audit Log'}
      description="Siapa melakukan apa, kapan, dari IP mana. Tercatat otomatis untuk login, CRUD user/role, sesi, dan notifikasi."
      loading={loading}
      showLoading={showLoading}
      error={error}
      errorTitle="Gagal memuat audit"
      onClearError={() => setError('')}
      onRefresh={() => controller.btrefresh_click(load)}
      extraActions={<DownloadButton rows={logs} columns={COLUMNS_ITEMS} filename="audit_log" />}
      filterBar={
        <form onSubmit={applyFilter} className="mb-4 flex flex-wrap items-end gap-2">
          <label className="flex flex-col gap-1 text-xs font-medium text-zinc-500">
            Aksi
            <select value={action} onChange={(e) => { setAction(e.target.value); setOffset(0) }} className={SELECT_CLASS}>
              {ACTIONS.map((a) => (
                <option key={a} value={a}>{a === '' ? 'Semua' : a}</option>
              ))}
            </select>
          </label>
          <label className="flex flex-col gap-1 text-xs font-medium text-zinc-500">
            Entitas
            <select value={entity} onChange={(e) => { setEntity(e.target.value); setOffset(0) }} className={SELECT_CLASS}>
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
              className={`${SELECT_CLASS} w-32 font-mono`}
            />
          </label>
          <Button size="sm" type="submit">
            Filter
          </Button>
          <Button size="sm" variant="ghost" type="button" onClick={reset}>
            Reset
          </Button>
        </form>
      }
      items={logs}
      emptyTitle="Belum ada jejak audit"
      emptyDescription="Sesuaikan filter atau lakukan aksi (mis. login) agar tercatat di sini."
      offset={offset}
      onPage={setOffset}
    >
      <StandardGrid columns={COLUMNS_ITEMS} rows={logs} minWidth={640} />
    </StandardPage>
  )
}
