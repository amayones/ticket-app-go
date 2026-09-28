// Template seragam semua halaman app/*.
// Satu sumber PAGE_SIZE + class + flags default agar tiap index.jsx
// tidak menulis ulang konstanta yang sama. Nilai disepakati: 10 semua.
export const PAGE_SIZE = 10
export const SKELETON_ROWS = 5

export const SELECT_CLASS =
  'rounded-lg border border-zinc-300 bg-white px-2.5 py-1.5 text-sm text-zinc-700 dark:border-zinc-700 dark:bg-zinc-900 dark:text-zinc-200'

export const INPUT_CLASS =
  'w-full rounded-xl border border-zinc-300 bg-white px-3.5 py-2.5 text-sm text-zinc-900 focus:border-violet-500 focus:outline-none focus:ring-2 focus:ring-violet-500/30 dark:border-zinc-700 dark:bg-zinc-900 dark:text-zinc-50'

// Flags default: tiap index.jsx menyalin objek ini ke const FEATURES
// lalu tinggal set true/false per blok tanpa ubah return.
export const FEATURES_DEFAULT = {
  header: true,
  refresh: true,
  filter: true,
  tabs: true,
  create: true,
  edit: true,
  remove: true,
  pagination: true,
  empty: true,
  error: true,
  confirmDialog: true,
  extraActions: true,
}
