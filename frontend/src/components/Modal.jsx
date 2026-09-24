import { useEffect } from 'react'
import { createPortal } from 'react-dom'
import Icon from './icons.jsx'

const SIZES = {
  sm: 'max-w-sm',
  md: 'max-w-md',
  lg: 'max-w-lg',
}

// Dialog modal: tutup via ESC, klik backdrop, atau tombol ×.
// <Modal open title="…" onClose footer={<…/>}>isi…</Modal>
// Set showClose={false} + closeOnBackdrop={false} + onClose={undefined}
// untuk dialog wajib (misal sesi habis) yang tidak boleh di-skip.
export default function Modal({
  open,
  onClose,
  title,
  children,
  footer,
  size = 'md',
  closeOnBackdrop = true,
  showClose = true,
}) {
  useEffect(() => {
    if (!open || !onClose) return
    function onKey(e) {
      if (e.key === 'Escape') onClose?.()
    }
    document.addEventListener('keydown', onKey)
    const prev = document.body.style.overflow
    document.body.style.overflow = 'hidden'
    return () => {
      document.removeEventListener('keydown', onKey)
      document.body.style.overflow = prev
    }
  }, [open, onClose])

  if (!open) return null

  return createPortal(
    <div
      className="anim-fade-in fixed inset-0 z-[90] flex items-center justify-center bg-black/50 p-4 backdrop-blur-sm"
      onMouseDown={(e) => {
        if (closeOnBackdrop && e.target === e.currentTarget) onClose?.()
      }}
    >
      <div
        role="dialog"
        aria-modal="true"
        aria-label={typeof title === 'string' ? title : 'Dialog'}
        className={`anim-modal-pop w-full ${SIZES[size] || SIZES.md} overflow-hidden rounded-2xl bg-white shadow-2xl dark:bg-zinc-900`}
      >
        <div className="flex items-center justify-between gap-3 border-b border-zinc-100 px-5 py-4 dark:border-zinc-800">
          <h3 className="truncate text-base font-semibold text-zinc-900 dark:text-zinc-50">{title}</h3>
          {showClose && (
            <button
              type="button"
              onClick={onClose}
              aria-label="Tutup dialog"
              className="shrink-0 rounded-lg p-1.5 text-zinc-400 transition hover:bg-zinc-100 hover:text-zinc-700 dark:hover:bg-zinc-800 dark:hover:text-zinc-200"
            >
              <Icon name="x" className="h-5 w-5" />
            </button>
          )}
        </div>
        <div className="max-h-[70vh] overflow-y-auto px-5 py-4 text-sm text-zinc-600 dark:text-zinc-300">
          {children}
        </div>
        {footer && (
          <div className="flex justify-end gap-2 border-t border-zinc-100 bg-zinc-50 px-5 py-3.5 dark:border-zinc-800 dark:bg-zinc-900/60">
            {footer}
          </div>
        )}
      </div>
    </div>,
    document.body,
  )
}
