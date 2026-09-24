import { useState } from 'react'

const POSITIONS = {
  top: 'bottom-full left-1/2 mb-2 -translate-x-1/2',
  bottom: 'top-full left-1/2 mt-2 -translate-x-1/2',
  left: 'right-full top-1/2 mr-2 -translate-y-1/2',
  right: 'left-full top-1/2 ml-2 -translate-y-1/2',
}

// Tooltip modern pengganti title bawaan browser.
// <Tooltip label="Tutup sidebar" position="right"><button>…</button></Tooltip>
export default function Tooltip({ label, position = 'top', children, className = '' }) {
  const [visible, setVisible] = useState(false)
  if (!label) return children
  return (
    <span
      className={`relative inline-flex ${className}`}
      onMouseEnter={() => setVisible(true)}
      onMouseLeave={() => setVisible(false)}
      onFocus={() => setVisible(true)}
      onBlur={() => setVisible(false)}
    >
      {children}
      <span
        role="tooltip"
        aria-hidden={!visible}
        className={`anim-tooltip pointer-events-none absolute z-[120] whitespace-nowrap rounded-lg bg-zinc-900 px-2.5 py-1.5 text-xs font-medium text-zinc-50 shadow-lg transition-opacity duration-150 dark:bg-zinc-100 dark:text-zinc-900 ${
          POSITIONS[position] || POSITIONS.top
        } ${visible ? 'opacity-100' : 'opacity-0'}`}
      >
        {label}
      </span>
    </span>
  )
}
