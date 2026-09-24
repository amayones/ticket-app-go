const VARIANTS = {
  primary:
    'bg-violet-600 text-white shadow-sm hover:bg-violet-500 focus-visible:ring-violet-500 disabled:hover:bg-violet-600',
  secondary:
    'border border-zinc-300 bg-white text-zinc-700 hover:bg-zinc-50 focus-visible:ring-violet-500 dark:border-zinc-700 dark:bg-zinc-900 dark:text-zinc-200 dark:hover:bg-zinc-800',
  danger:
    'bg-rose-600 text-white shadow-sm hover:bg-rose-500 focus-visible:ring-rose-500 disabled:hover:bg-rose-600',
  ghost:
    'text-zinc-600 hover:bg-zinc-100 hover:text-zinc-900 focus-visible:ring-violet-500 dark:text-zinc-300 dark:hover:bg-zinc-800 dark:hover:text-zinc-50',
}

const SIZES = {
  sm: 'px-3 py-1.5 text-sm rounded-lg gap-1.5',
  md: 'px-4 py-2.5 text-sm rounded-xl gap-2',
  lg: 'px-5 py-3 text-base rounded-xl gap-2',
}

function SpinnerGlyph({ className = 'h-4 w-4' }) {
  return (
    <svg className={`animate-spin ${className}`} viewBox="0 0 24 24" fill="none" aria-hidden="true">
      <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" />
      <path className="opacity-90" fill="currentColor" d="M4 12a8 8 0 018-8v4a4 4 0 00-4 4H4z" />
    </svg>
  )
}

// Tombol modern: varian, ukuran, state loading.
// <Button variant="danger" loading onClick={…}>Hapus</Button>
export default function Button({
  variant = 'primary',
  size = 'md',
  loading = false,
  fullWidth = false,
  className = '',
  children,
  disabled,
  ...rest
}) {
  return (
    <button
      type="button"
      disabled={disabled || loading}
      className={`inline-flex items-center justify-center font-semibold transition active:scale-[0.98] disabled:cursor-not-allowed disabled:opacity-60 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-offset-2 dark:focus-visible:ring-offset-zinc-900 ${
        VARIANTS[variant] || VARIANTS.primary
      } ${SIZES[size] || SIZES.md} ${fullWidth ? 'w-full' : ''} ${className}`}
      {...rest}
    >
      {loading && <SpinnerGlyph />}
      {children}
    </button>
  )
}
