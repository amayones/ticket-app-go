// View Refunds — daftar pengajuan + form pengajuan baru (milik sendiri).
// Controller: ./controller.js (6 fungsi). Store: read_data.
import { useEffect, useState } from 'react'
import {
  Alert,
  Badge,
  Button,
  Modal,
  StandardGrid,
  StandardPage,
  TextField,
  formatAmount,
  formatTime,
  useDashboard,
  useToast,
} from '../shared/all.js'
import { process_create, read_data } from './api.js'
import { controller } from './controller.js'

export const meta = { label: 'Refunds', icon: 'refresh', order: 3 }

function rupiah(value) {
  const s = formatAmount(value)
  return s === '' ? 'Rp 0' : `Rp ${s}`
}

function statusTone(status) {
  switch (status) {
    case 'APPROVED':
      return 'success'
    case 'REJECTED':
      return 'danger'
    default:
      return 'warning'
  }
}

export default function Refunds({ nvdata }) {
  const toast = useToast()
  const [showFRM, setShowFRM] = useState(false)
  const [orderCode, setOrderCode] = useState('')
  const [reason, setReason] = useState('')
  const [formError, setFormError] = useState('')
  const [saving, setSaving] = useState(false)
  const { data, loading, showLoading, error, setError, load } = useDashboard(read_data, {})

  useEffect(() => {
    controller.init()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  const items = data || []
  const pageItems = [{ items }]

  function handler_validasi_input() {
    if (!orderCode.trim()) {
      setFormError('Kode order wajib diisi (lihat di tiket Anda).')
      return false
    }
    if (!reason.trim()) {
      setFormError('Alasan wajib diisi.')
      return false
    }
    return true
  }

  async function handler_btsave() {
    try {
      if (handler_validasi_input() === false) return false
      setSaving(true)
      await process_create({ order_code: orderCode.trim(), reason: reason.trim() })
      toast.success('Pengajuan dikirim, menunggu keputusan organizer.')
      setShowFRM(false)
      setOrderCode('')
      setReason('')
      load()
    } catch (ex) {
      setFormError(ex.message)
    } finally {
      setSaving(false)
    }
  }

  const COLUMNS = [
    { header: 'Kode', dataIndex: 'code', width: 110, render: (r) => <span className="font-mono">{r.code}</span> },
    { header: 'Order', dataIndex: 'order_code', width: 120, render: (r) => <span className="font-mono">{r.order_code}</span> },
    { header: 'Event', dataIndex: 'event_title', width: 170 },
    {
      header: 'Nominal',
      dataIndex: 'amount',
      width: 120,
      render: (r) => <span className="font-mono">{rupiah(r.amount)}</span>,
    },
    { header: 'Status', dataIndex: 'status', width: 100, render: (r) => <Badge tone={statusTone(r.status)}>{r.status}</Badge> },
    {
      header: 'Diajukan',
      dataIndex: 'created_at',
      width: 150,
      render: (r) => <span className="whitespace-nowrap text-zinc-500">{formatTime(r.created_at)}</span>,
    },
  ]

  return (
    <div className="flex flex-col gap-4">
      <StandardPage
        title={nvdata?.label || 'Refunds'}
        description="Pengajuan refund Anda. Dana kembali diproses module FINANCE setelah disetujui."
        loading={loading}
        showLoading={showLoading}
        error={error}
        errorTitle="Gagal memuat refund"
        onClearError={() => setError('')}
        onRefresh={() => controller.btrefresh_click(load)}
        onCreate={() => {
          setOrderCode('')
          setReason('')
          setFormError('')
          setShowFRM(true)
        }}
        createLabel="Ajukan Refund"
        items={pageItems}
        emptyTitle="Belum ada pengajuan"
        emptyDescription="Refund hanya untuk order lunas yang tiketnya belum dipakai dan event belum mulai."
        offset={0}
        onPage={null}
      >
        <StandardGrid columns={COLUMNS} rows={items} minWidth={720} />
      </StandardPage>

      <Modal
        open={showFRM}
        onClose={saving ? undefined : () => setShowFRM(false)}
        title="Ajukan refund"
        size="sm"
        closeOnBackdrop={!saving}
        footer={
          <>
            <Button variant="secondary" onClick={() => setShowFRM(false)} disabled={saving}>
              Batal
            </Button>
            <Button onClick={handler_btsave} loading={saving}>
              Kirim
            </Button>
          </>
        }
      >
        <div className="flex flex-col gap-4">
          {formError && (
            <Alert tone="error" closable onClose={() => setFormError('')}>
              {formError}
            </Alert>
          )}
          <TextField
            label="Kode order"
            value={orderCode}
            onChange={(e) => setOrderCode(e.target.value)}
            placeholder="ORD-XXXXXXXX"
            hint="Lihat kode di detail tiket Anda."
          />
          <TextField
            label="Alasan"
            value={reason}
            onChange={(e) => setReason(e.target.value)}
            placeholder="Contoh: jadwal bentrok"
          />
        </div>
      </Modal>

      {error && (
        <Alert tone="error" title="Gagal memuat refund" closable onClose={() => setError('')}>
          {error}
        </Alert>
      )}
    </div>
  )
}
