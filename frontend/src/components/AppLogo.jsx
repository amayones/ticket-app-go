import { useState } from 'react'
import { APP_LOGO, APP_NAME } from '../brand.js'

// Logo aplikasi untuk sidebar, topbar mobile, dan kartu login.
// Sumber gambarnya APP_LOGO (default /favicon.svg di frontend/public/),
// jadi mengganti favicon ikut mengganti logo di seluruh UI.
// Kalau gambar gagal dimuat, jatuh ke huruf pertama APP_NAME.
export default function AppLogo({ className = 'h-8 w-8', alt = '' }) {
  const [broken, setBroken] = useState(false)

  if (broken) {
    return (
      <span
        aria-hidden="true"
        className={`flex shrink-0 items-center justify-center rounded-lg bg-zinc-900 text-sm font-bold text-white dark:bg-zinc-100 dark:text-zinc-900 ${className}`}
      >
        {APP_NAME.charAt(0).toUpperCase()}
      </span>
    )
  }

  return (
    <img
      src={APP_LOGO}
      alt={alt}
      aria-hidden={alt ? undefined : true}
      onError={() => setBroken(true)}
      className={`shrink-0 rounded-lg object-contain ${className}`}
    />
  )
}
