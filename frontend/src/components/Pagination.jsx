import Button from './Button.jsx'
import Icon from './icons.jsx'

// Navigasi halaman berbasis offset (backend: limit/offset).
// hasMore = true bila jumlah data yang kembali == limit.
export default function Pagination({ offset, limit, count, hasMore, loading = false, onPage }) {
  const from = count === 0 ? 0 : offset + 1
  const to = offset + count
  return (
    <div className="flex flex-wrap items-center justify-between gap-3">
      <p className="text-xs text-zinc-500 dark:text-zinc-400">
        Menampilkan <span className="font-semibold text-zinc-700 dark:text-zinc-200">{from}–{to}</span>
      </p>
      <div className="flex items-center gap-2">
        <Button variant="secondary" size="sm" disabled={offset === 0 || loading} onClick={() => onPage(offset - limit)}>
          <Icon name="chevronLeft" className="h-4 w-4" />
          Sebelumnya
        </Button>
        <Button variant="secondary" size="sm" disabled={!hasMore || loading} onClick={() => onPage(offset + limit)}>
          Berikutnya
          <Icon name="chevronRight" className="h-4 w-4" />
        </Button>
      </div>
    </div>
  )
}
