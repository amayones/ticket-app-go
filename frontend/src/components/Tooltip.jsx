import { useEffect, useRef, useState } from 'react'

const POSITIONS = {
  top: 'bottom-full left-1/2 mb-2 -translate-x-1/2',
  bottom: 'top-full left-1/2 mt-2 -translate-x-1/2',
  left: 'right-full top-1/2 mr-2 -translate-y-1/2',
  right: 'left-full top-1/2 ml-2 -translate-y-1/2',
}

// Tooltip modern pengganti title bawaan browser.
// Aturan tampil ketat: hanya saat hover menetap (delay), fokus keyboard,
// atau sentuh — mati saat leave/blur/klik/scroll/Esc.
// <Tooltip label="Tutup sidebar" position="right"><button>…</button></Tooltip>
export default function Tooltip({ label, position = 'top', openDelay = 200, children, className = '' }) {
  const [visible, setVisible] = useState(false)
  const timer = useRef(null)

  function clear() {
    if (timer.current) {
      clearTimeout(timer.current)
      timer.current = null
    }
  }

  useEffect(() => clear, [])

  if (!label) return children

  function enter() {
    clear()
    timer.current = setTimeout(() => setVisible(true), openDelay)
  }

  function leave() {
    clear()
    setVisible(false)
  }

  // Hanya tampil saat fokus keyboard asli, BUKAN fokus sisa klik mouse.
  function focus(e) {
    if (e.target.matches && e.target.matches(':focus-visible')) {
      clear()
      setVisible(true)
    }
  }

  function key(e) {
    if (e.key === 'Escape') leave()
  }

  // Bila pemanggil sudah memberi position sendiri (absolute/fixed/sticky),
  // jangan timpa dengan relative agar tidak konflik.
  const positioned = /(^|\s)(absolute|fixed|sticky)(\s|$)/.test(className)

  return (
    <span
      className={`${positioned ? '' : 'relative '}inline-flex ${className}`}
      onMouseEnter={enter}
      onMouseLeave={leave}
      onFocus={focus}
      onBlur={leave}
      onClick={leave}
      onKeyDown={key}
    >
      {children}
      {visible && (
        <span
          role="tooltip"
          className={`anim-tooltip pointer-events-none absolute z-[120] whitespace-nowrap rounded-lg bg-zinc-900 px-2.5 py-1.5 text-xs font-medium text-zinc-50 shadow-lg dark:bg-zinc-100 dark:text-zinc-900 ${
            POSITIONS[position] || POSITIONS.top
          }`}
        >
          {label}
        </span>
      )}
    </span>
  )
}
