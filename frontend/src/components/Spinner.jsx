// Indikator loading: inline, overlay fullscreen, dan skeleton.
export function Spinner({ className = 'h-5 w-5', label = 'Memuat…' }) {
  return (
    <span role="status" aria-label={label} className="inline-flex items-center">
      <svg className={`animate-spin text-violet-600 dark:text-violet-400 ${className}`} viewBox="0 0 24 24" fill="none" aria-hidden="true">
        <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" />
        <path className="opacity-90" fill="currentColor" d="M4 12a8 8 0 018-8v4a4 4 0 00-4 4H4z" />
      </svg>
    </span>
  )
}

export function PageLoader({ label = 'Memuat…' }) {
  return (
    <div className="flex items-center justify-center gap-3 rounded-2xl border border-zinc-200 bg-white px-6 py-10 dark:border-zinc-800 dark:bg-zinc-900">
      <Spinner className="h-6 w-6" label={label} />
      <p className="text-sm text-zinc-500 dark:text-zinc-400">{label}</p>
    </div>
  )
}

export function Skeleton({ className = 'h-4 w-full' }) {
  return <div aria-hidden="true" className={`skeleton-shimmer rounded-lg ${className}`} />
}

// Placeholder baris tabel/kartu saat data dimuat.
export function SkeletonRows({ rows = 4 }) {
  return (
    <div className="flex flex-col gap-2.5" aria-hidden="true">
      {Array.from({ length: rows }).map((_, i) => (
        <div key={i} className="flex items-center gap-3 rounded-xl border border-zinc-100 p-3 dark:border-zinc-800">
          <div className="skeleton-shimmer h-10 w-10 shrink-0 rounded-full" />
          <div className="flex flex-1 flex-col gap-2">
            <div className="skeleton-shimmer h-3.5 w-1/3 rounded" />
            <div className="skeleton-shimmer h-3 w-1/2 rounded" />
          </div>
        </div>
      ))}
    </div>
  )
}
