// Items grid ala GRID<mod>.js langit_v2: yang beda per tabel hanya bagian ini.
// Tabel DB: CPSYSLOG (via ./api.js listSyslogs).
import { Badge, formatTime } from '../shared/all.js'

export const LEVELS = ['', 'ERROR', 'WARN', 'INFO']

export function toneFor(level) {
  if (level === 'ERROR') return 'danger'
  if (level === 'WARN') return 'warning'
  return 'info'
}

export const SYSLOG_COLUMNS = [
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
