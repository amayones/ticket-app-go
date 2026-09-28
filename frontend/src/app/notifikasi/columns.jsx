// Items grid + form ala GRID/FRM<mod>.js langit_v2: yang beda per tabel hanya bagian ini.
// Tabel DB: CPNOTIFTEMPLATE + CPNOTIFLOG (via ./api.js).
import { Badge, formatTime } from '../shared/all.js'

export const CHANNELS = ['EMAIL', 'PUSH', 'INAPP']

export const NOTIF_LOG_COLUMNS = [
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

// Items form template (dipakai StandardForm).
export const TEMPLATE_FIELDS = [
  { name: 'name', label: 'Nama template', type: 'text', placeholder: 'mis. Promo Akhir Tahun' },
  { name: 'channel', label: 'Channel', type: 'select', options: CHANNELS.map((c) => ({ value: c, label: c })) },
  { name: 'subject', label: 'Subjek', type: 'text', placeholder: 'dipakai untuk EMAIL' },
  {
    name: 'body',
    label: 'Isi pesan',
    type: 'textarea',
    rows: 5,
    placeholder: 'Halo {{nama}}, …',
    hint: 'Variabel: {{nama}} {{kode}} {{role}} {{detail}}',
  },
  { name: 'is_active', label: 'Template aktif (nonaktif = tidak bisa dikirim)', type: 'checkbox' },
]

// Items form kirim/test (dipakai StandardForm; opsi template diisi dinamis).
export function sendFields(activeTemplates = []) {
  return [
    {
      name: 'template_code',
      label: 'Template',
      type: 'select',
      options: [
        { value: '', label: '— pilih —' },
        ...activeTemplates.map((t) => ({ value: t.code, label: `${t.name} (${t.channel})` })),
      ],
    },
    { name: 'recipient', label: 'Penerima', type: 'text', placeholder: 'email / id perangkat / kode user' },
    { name: 'nama', label: 'Variabel {{nama}}', type: 'text', placeholder: 'nama penerima' },
    { name: 'detail', label: 'Variabel {{detail}}', type: 'text', placeholder: 'detail tambahan' },
  ]
}
