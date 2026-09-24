import { useState } from 'react'
import Icon from './icons.jsx'

const INPUT_CLASS =
  'w-full rounded-xl border bg-white px-3.5 py-2.5 text-sm text-zinc-900 placeholder:text-zinc-400 transition focus:outline-none focus:ring-2 dark:bg-zinc-900 dark:text-zinc-50 dark:placeholder:text-zinc-500'
const NORMAL_RING = 'border-zinc-300 focus:border-violet-500 focus:ring-violet-500/30 dark:border-zinc-700'
const ERROR_RING = 'border-rose-400 focus:border-rose-500 focus:ring-rose-500/30 dark:border-rose-600'

// Input teks berlabel dengan pesan error & hint.
// <TextField label="Email" error={err} hint="…" {...register} />
export default function TextField({ label, id, error, hint, className = '', ...props }) {
  const inputId = id || `field-${label?.toLowerCase().replace(/\s+/g, '-')}`
  return (
    <label htmlFor={inputId} className={`block ${className}`}>
      {label && (
        <span className="mb-1.5 block text-sm font-medium text-zinc-700 dark:text-zinc-200">{label}</span>
      )}
      <input id={inputId} className={`${INPUT_CLASS} ${error ? ERROR_RING : NORMAL_RING}`} {...props} />
      {error ? (
        <span className="mt-1.5 block text-xs font-medium text-rose-600 dark:text-rose-400">{error}</span>
      ) : (
        hint && <span className="mt-1.5 block text-xs text-zinc-500 dark:text-zinc-400">{hint}</span>
      )}
    </label>
  )
}

// Input password dengan tombol intip (eye / eye-off).
export function PasswordInput({ label, id, error, hint, className = '', ...props }) {
  const [visible, setVisible] = useState(false)
  const inputId = id || `field-${label?.toLowerCase().replace(/\s+/g, '-')}`
  return (
    <label htmlFor={inputId} className={`block ${className}`}>
      {label && (
        <span className="mb-1.5 block text-sm font-medium text-zinc-700 dark:text-zinc-200">{label}</span>
      )}
      <span className="relative block">
        <input
          id={inputId}
          type={visible ? 'text' : 'password'}
          className={`${INPUT_CLASS} pr-11 ${error ? ERROR_RING : NORMAL_RING}`}
          {...props}
        />
        <button
          type="button"
          onClick={() => setVisible((v) => !v)}
          aria-label={visible ? 'Sembunyikan password' : 'Tampilkan password'}
          className="absolute right-2.5 top-1/2 -translate-y-1/2 rounded-md p-1 text-zinc-400 transition hover:bg-zinc-100 hover:text-zinc-700 dark:hover:bg-zinc-800 dark:hover:text-zinc-200"
        >
          <Icon name={visible ? 'eyeOff' : 'eye'} className="h-5 w-5" />
        </button>
      </span>
      {error ? (
        <span className="mt-1.5 block text-xs font-medium text-rose-600 dark:text-rose-400">{error}</span>
      ) : (
        hint && <span className="mt-1.5 block text-xs text-zinc-500 dark:text-zinc-400">{hint}</span>
      )}
    </label>
  )
}
