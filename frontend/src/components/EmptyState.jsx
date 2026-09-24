import Icon from './icons.jsx'

// Tampilan data kosong: ikon + judul + deskripsi + aksi opsional.
export default function EmptyState({ icon = 'users', title = 'Belum ada data', description, action }) {
  return (
    <div className="flex flex-col items-center gap-2 rounded-2xl border border-dashed border-zinc-300 px-6 py-12 text-center dark:border-zinc-700">
      <span className="mb-1 flex h-12 w-12 items-center justify-center rounded-full bg-zinc-100 text-zinc-500 dark:bg-zinc-800 dark:text-zinc-400">
        <Icon name={icon} className="h-6 w-6" />
      </span>
      <p className="text-base font-semibold text-zinc-900 dark:text-zinc-50">{title}</p>
      {description && <p className="max-w-sm text-sm text-zinc-500 dark:text-zinc-400">{description}</p>}
      {action && <div className="mt-3">{action}</div>}
    </div>
  )
}
