import { useState } from 'react'
import Icon from './icons.jsx'
import { applyTheme, currentTheme } from './theme.js'

// Tombol ikon bulan/matahari untuk ganti tema terang-gelap.
export default function ThemeToggle({ className = '' }) {
  const [theme, setTheme] = useState(currentTheme)
  const dark = theme === 'dark'

  function toggle() {
    setTheme(applyTheme(dark ? 'light' : 'dark'))
  }

  return (
    <button
      type="button"
      onClick={toggle}
      aria-label={dark ? 'Ganti ke mode terang' : 'Ganti ke mode gelap'}
      className={`rounded-lg p-2 text-zinc-500 transition hover:bg-zinc-100 hover:text-zinc-800 dark:text-zinc-400 dark:hover:bg-zinc-800 dark:hover:text-zinc-100 ${className}`}
    >
      <Icon name={dark ? 'sun' : 'moon'} className="h-5 w-5" />
    </button>
  )
}
