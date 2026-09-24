import { useCallback, useMemo, useState } from 'react'
import Icon from './icons.jsx'
import { ToastContext } from './useToast.js'

const TONE_STYLE = {
  success: {
    bar: 'bg-emerald-500',
    icon: 'text-emerald-500',
    iconName: 'check',
  },
  error: {
    bar: 'bg-rose-500',
    icon: 'text-rose-500',
    iconName: 'x',
  },
  warning: {
    bar: 'bg-amber-500',
    icon: 'text-amber-500',
    iconName: 'warning',
  },
  info: {
    bar: 'bg-sky-500',
    icon: 'text-sky-500',
    iconName: 'info',
  },
}

const MAX_VISIBLE = 5
let toastSeq = 1

function ToastItem({ toast, onDismiss }) {
  const style = TONE_STYLE[toast.tone] || TONE_STYLE.info
  return (
    <div
      role="status"
      className="anim-toast-in pointer-events-auto relative w-full overflow-hidden rounded-xl border border-zinc-200 bg-white shadow-xl dark:border-zinc-700 dark:bg-zinc-900"
    >
      <div className="flex items-start gap-3 p-3.5">
        <span className={`mt-0.5 shrink-0 ${style.icon}`}>
          <Icon name={style.iconName} className="h-5 w-5" />
        </span>
        <div className="min-w-0 flex-1">
          {toast.title && (
            <p className="text-sm font-semibold text-zinc-900 dark:text-zinc-50">{toast.title}</p>
          )}
          <p className="break-words text-sm text-zinc-600 dark:text-zinc-300">{toast.message}</p>
        </div>
        <button
          type="button"
          onClick={() => onDismiss(toast.id)}
          aria-label="Tutup notifikasi"
          className="shrink-0 rounded-md p-1 text-zinc-400 transition hover:bg-zinc-100 hover:text-zinc-700 dark:hover:bg-zinc-800 dark:hover:text-zinc-200"
        >
          <Icon name="x" className="h-4 w-4" />
        </button>
      </div>
      {toast.duration > 0 && (
        <div className="h-1 w-full bg-zinc-100 dark:bg-zinc-800">
          <div
            className={`toast-progress h-full ${style.bar}`}
            style={{ animationDuration: `${toast.duration}ms` }}
          />
        </div>
      )}
    </div>
  )
}

function Toaster({ toasts, onDismiss }) {
  return (
    <div
      aria-live="polite"
      className="pointer-events-none fixed right-4 top-4 z-[100] flex w-[min(92vw,360px)] flex-col gap-2"
    >
      {toasts.map((t) => (
        <ToastItem key={t.id} toast={t} onDismiss={onDismiss} />
      ))}
    </div>
  )
}

export default function ToastProvider({ children }) {
  const [toasts, setToasts] = useState([])

  const dismiss = useCallback((id) => {
    setToasts((prev) => prev.filter((t) => t.id !== id))
  }, [])

  const push = useCallback(
    (tone, message, opts = {}) => {
      const id = toastSeq++
      const duration = opts.duration ?? (tone === 'error' ? 6000 : 4000)
      setToasts((prev) => [...prev.slice(-(MAX_VISIBLE - 1)), { id, tone, message, title: opts.title, duration }])
      if (duration > 0) setTimeout(() => dismiss(id), duration)
      return id
    },
    [dismiss],
  )

  const value = useMemo(
    () => ({
      notify: push,
      success: (message, opts) => push('success', message, opts),
      error: (message, opts) => push('error', message, opts),
      warning: (message, opts) => push('warning', message, opts),
      info: (message, opts) => push('info', message, opts),
      dismiss,
    }),
    [push, dismiss],
  )

  return (
    <ToastContext.Provider value={value}>
      {children}
      <Toaster toasts={toasts} onDismiss={dismiss} />
    </ToastContext.Provider>
  )
}
