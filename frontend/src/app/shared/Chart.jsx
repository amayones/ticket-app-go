// Grafik SVG ringan tanpa dependensi (repo belum memakai lib grafik):
// TrendChart (garis + area), BarList (bar horizontal), ProgressBar.
// Warna mengikuti palet Tailwind yang dipakai menu lain (violet/emerald/rose).
export function TrendChart({ points = [], height = 180, formatValue }) {
  const W = 600
  const H = 200
  const PAD = 8
  const values = points.map((p) => Number(p.value) || 0)
  const max = Math.max(1, ...values)
  const step = points.length > 1 ? (W - PAD * 2) / (points.length - 1) : 0
  const xy = values.map((v, i) => [PAD + i * step, H - PAD - (v / max) * (H - PAD * 2)])
  const line = xy.map(([x, y]) => `${x.toFixed(1)},${y.toFixed(1)}`).join(' ')
  const area = `${PAD},${H - PAD} ${line} ${(W - PAD).toFixed(1)},${H - PAD}`
  const fmt = typeof formatValue === 'function' ? formatValue : (v) => String(v)
  return (
    <div className="w-full overflow-x-auto">
      <svg viewBox={`0 0 ${W} ${H}`} style={{ height }} className="w-full min-w-[420px]" role="img">
        {[0.25, 0.5, 0.75].map((f) => (
          <line
            key={f}
            x1={PAD}
            x2={W - PAD}
            y1={H * f}
            y2={H * f}
            className="stroke-zinc-200 dark:stroke-zinc-800"
            strokeWidth="1"
          />
        ))}
        <polygon points={area} className="fill-violet-500/10" />
        <polyline
          points={line}
          fill="none"
          className="stroke-violet-500"
          strokeWidth="2.5"
          strokeLinecap="round"
          strokeLinejoin="round"
        />
        {xy.map(([x, y], i) => (
          <circle key={i} cx={x} cy={y} r="4" className="fill-violet-500">
            <title>{`${points[i].label}: ${fmt(values[i])}`}</title>
          </circle>
        ))}
      </svg>
      {points.length > 0 && (
        <div className="mt-1 flex justify-between text-[11px] text-zinc-400 dark:text-zinc-500">
          <span>{points[0].label}</span>
          <span>{points[points.length - 1].label}</span>
        </div>
      )}
    </div>
  )
}

// BarList: bar horizontal + label + nilai (untuk per tipe tiket/kota/promo/produk).
export function BarList({ items = [], formatValue }) {
  const fmt = typeof formatValue === 'function' ? formatValue : (v) => String(v)
  const max = Math.max(1, ...items.map((it) => Number(it.value) || 0))
  if (items.length === 0) return null
  return (
    <ul className="flex flex-col gap-2.5">
      {items.map((it, i) => (
        <li key={it.label || i}>
          <div className="mb-1 flex items-baseline justify-between gap-2 text-xs">
            <span className="min-w-0 truncate font-medium text-zinc-700 dark:text-zinc-200">{it.label}</span>
            <span className="shrink-0 font-mono text-zinc-500 dark:text-zinc-400">
              {fmt(it.value)}
              {it.sub ? <span className="ml-1.5 opacity-70">{it.sub}</span> : null}
            </span>
          </div>
          <div className="h-2 overflow-hidden rounded-full bg-zinc-100 dark:bg-zinc-800">
            <div
              className="h-full rounded-full bg-violet-500 transition-all dark:bg-violet-400"
              style={{ width: `${(Math.max(0, Number(it.value) || 0) / max) * 100}%` }}
            />
          </div>
        </li>
      ))}
    </ul>
  )
}

// ProgressBar: progres 0..max (untuk terjual vs kuota, check-in).
export function ProgressBar({ value = 0, max = 1, tone = 'violet' }) {
  const tones = {
    violet: 'bg-violet-500 dark:bg-violet-400',
    emerald: 'bg-emerald-500 dark:bg-emerald-400',
    rose: 'bg-rose-500 dark:bg-rose-400',
    amber: 'bg-amber-500 dark:bg-amber-400',
  }
  const pct = max > 0 ? Math.min(100, (Math.max(0, value) / max) * 100) : 0
  return (
    <div className="h-2 overflow-hidden rounded-full bg-zinc-100 dark:bg-zinc-800">
      <div className={`h-full rounded-full transition-all ${tones[tone] || tones.violet}`} style={{ width: `${pct}%` }} />
    </div>
  )
}
