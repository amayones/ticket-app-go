// Panel detail peserta — padanan FRM baca (window info, tanpa tbar save).
// Read-only: check-in dikerjakan module TICKETING.
import { Badge, Button, Modal } from '../../shared/all.js'
import { formatTime } from '../../../components/format.js'

export default function AttendeeModal({ attendee, onClose }) {
  return (
    <Modal open={!!attendee} onClose={onClose} title="Detail peserta" size="sm">
      {attendee && (
        <dl className="flex flex-col gap-2.5 text-sm">
          {[
            ['Nama', attendee.buyer_name],
            ['Username', attendee.buyer_username ? `@${attendee.buyer_username}` : '—'],
            ['Email', attendee.email],
            ['Telepon', attendee.phone || '—'],
            ['Tipe tiket', attendee.ticket_type],
            ['Kode tiket', attendee.ticket_code],
            ['Order', attendee.order_code],
            ['Pembayaran', attendee.payment_status],
            ['Kehadiran', attendee.attendance],
            ['Check-in', attendee.checked_in_at ? formatTime(attendee.checked_in_at) : '—'],
            ['Dibeli', formatTime(attendee.purchase_date)],
          ].map(([label, value]) => (
            <div key={label} className="flex items-start justify-between gap-4">
              <dt className="shrink-0 text-zinc-500 dark:text-zinc-400">{label}</dt>
              <dd className="min-w-0 break-words text-right font-medium text-zinc-900 dark:text-zinc-50">
                {label === 'Kehadiran' ? (
                  <Badge tone={value === 'HADIR' ? 'success' : 'neutral'}>{value}</Badge>
                ) : (
                  value || '—'
                )}
              </dd>
            </div>
          ))}
          <div className="mt-1 flex justify-end">
            <Button variant="secondary" size="sm" onClick={onClose}>
              Tutup
            </Button>
          </div>
        </dl>
      )}
    </Modal>
  )
}
