// GRID Notifikasi — padanan GRID<mod>.js langit_v2 (Tipe viewer: read-only,
// seperti grid approval doc_*/mapp_*: tanpa rowediting, tanpa FRM inline).
// Kolom items inline di sini (satu-satunya yang beda per tabel).
import { Badge, StandardGrid, formatTime } from '../shared/all.js'

const COLUMNS_ITEMS = [
  {
    header: 'Waktu',
    dataIndex: 'created_at',
    width: 130,
    render: (l) => <span className="whitespace-nowrap text-zinc-500">{formatTime(l.created_at)}</span>,
  },
  {
    header: 'Channel',
    dataIndex: 'channel',
    width: 90,
    render: (l) => <Badge tone="info">{l.channel}</Badge>,
  },
  {
    header: 'Penerima',
    dataIndex: 'recipient',
    width: 180,
    render: (l) => <span className="block max-w-[180px] truncate">{l.recipient}</span>,
  },
  {
    header: 'Subjek',
    dataIndex: 'subject',
    width: 220,
    render: (l) => (
      <span className="block max-w-[220px] truncate" title={l.body}>
        {l.subject || l.body}
      </span>
    ),
  },
  {
    header: 'Status',
    dataIndex: 'status',
    width: 90,
    render: (l) => <Badge tone={l.status === 'SENT' ? 'success' : 'danger'}>{l.status}</Badge>,
  },
]

export default function GRID({ rows }) {
  return <StandardGrid columns={COLUMNS_ITEMS} rows={rows} minWidth={560} />
}

GRID.columns = COLUMNS_ITEMS
