import { useState } from 'react'
import Icon from './icons.jsx'

// Banner inline untuk error/info di dalam halaman atau form.
// <Alert tone="error" title="Gagal masuk" closable>pesan…</Alert>
const TONES = {
  success: {
    box: 'border-emerald-200 bg-emerald-50 text-emerald-900 dark:border-emerald-800 dark:bg-emerald-950 dark:text-emerald-100',
    icon: 'text-emerald-500',
    iconName: 'check',
  },
  error: {
    box: 'border-rose-200 bg-rose-50 text-rose-900 dark:border-rose-800 dark:bg-rose-950 dark:text-rose-100',
    icon: 'text-rose-500',
    iconName: 'x',
  },
  warning: {
    box: 'border-amber-200 bg-amber-50 text-amber-900 dark:border-amber-800 dark:bg-amber-950 dark:text-amber-100',
    icon: 'text-amber-500',
    iconName: 'warning',
  },
  info: {
    box: 'border-sky-200 bg-sky-50 text-sky-900 dark:border-sky-800 dark:bg-sky-950 dark:text-sky-100',
    icon: 'text-sky-500',
    iconName: 'info',
  },
}

export default function Alert({ tone = 'info', title, children, closable = false, onClose, className = '' }) {
  const [visible, setVisible] = useState(true)
  if (!visible) return null
  const style = TONES[tone] || TONES.info

  function handleClose() {
    setVisible(false)
    onClose?.()
  }

  return (
    <div role="alert" className={`flex items-start gap-3 rounded-xl border px-4 py-3 text-sm ${style.box} ${className}`}>
      <span className={`mt-0.5 shrink-0 ${style.icon}`}>
        <Icon name={style.iconName} className="h-5 w-5" />
      </span>
      <div className="min-w-0 flex-1">
        {title && <p className="font-semibold">{title}</p>}
        <div className="break-words opacity-90">{children}</div>
      </div>
      {closable && (
        <button
          type="button"
          onClick={handleClose}
          aria-label="Tutup peringatan"
          className="shrink-0 rounded-md p-1 opacity-60 transition hover:opacity-100"
        >
          <Icon name="x" className="h-4 w-4" />
        </button>
      )}
    </div>
  )
}
