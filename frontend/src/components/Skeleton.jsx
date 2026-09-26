// Placeholder saat data dimuat: satu baris (Skeleton) dan manyak baris
// tabel/kartu (SkeletonRows). Dipakai seluruh menu bergaya tabel.
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
