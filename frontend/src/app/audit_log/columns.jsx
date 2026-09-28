// Items grid ala GRID<mod>.js langit_v2: yang beda per tabel hanya bagian ini.
// Tabel DB: CPAUDITLOG (via ./api.js listAudit).
import { Badge, formatTime } from '../shared/all.js'

export const ACTIONS = [
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

export const ENTITIES = ['', 'AUTH', 'USER', 'ROLE', 'SESSION', 'NOTIFICATION', 'TEMPLATE', 'SYSTEM']

export function toneFor(action) {
  if (action.includes('DELETE') || action.includes('FAILED') || action.includes('REVOKE')) return 'danger'
  if (action.includes('CREATE') || (action.includes('LOGIN') && !action.includes('FAILED'))) return 'success'
  if (action.includes('UPDATE') || action.includes('ASSIGN') || action.includes('SEND')) return 'info'
  return 'neutral'
}

export const AUDIT_COLUMNS = [
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
