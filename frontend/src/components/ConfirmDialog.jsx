import Button from './Button.jsx'
import Modal from './Modal.jsx'

// Dialog konfirmasi siap pakai untuk aksi penting/destruktif.
export default function ConfirmDialog({
  open,
  title = 'Yakin lanjutkan?',
  message,
  confirmLabel = 'Ya, lanjutkan',
  cancelLabel = 'Batal',
  danger = false,
  loading = false,
  onConfirm,
  onCancel,
}) {
  return (
    <Modal
      open={open}
      onClose={loading ? undefined : onCancel}
      title={title}
      size="sm"
      closeOnBackdrop={!loading}
      footer={
        <>
          <Button variant="secondary" onClick={onCancel} disabled={loading}>
            {cancelLabel}
          </Button>
          <Button variant={danger ? 'danger' : 'primary'} onClick={onConfirm} loading={loading}>
            {confirmLabel}
          </Button>
        </>
      }
    >
      <p>{message}</p>
    </Modal>
  )
}
