// Pembungkus kartu konten yang konsisten.
export default function Card({ children, className = '' }) {
  return (
    <section
      className={`rounded-2xl border border-zinc-200 bg-white p-5 shadow-sm sm:p-6 dark:border-zinc-800 dark:bg-zinc-900 ${className}`}
    >
      {children}
    </section>
  )
}

export function CardTitle({ children, description }) {
  return (
    <div className="mb-4">
      <h2 className="m-0 text-lg font-semibold text-zinc-900 dark:text-zinc-50">{children}</h2>
      {description && <p className="mt-1 text-sm text-zinc-500 dark:text-zinc-400">{description}</p>}
    </div>
  )
}
