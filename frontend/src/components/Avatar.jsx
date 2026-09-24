// Lingkaran inisial username — satu gradien netral agar tidak warna-warni.
// Variasi kecil per nama (terang/gelap) tanpa mengubah hue.
const SHADES = ['from-zinc-500 to-zinc-700', 'from-zinc-600 to-zinc-800', 'from-slate-500 to-slate-700']

function shadeFor(name) {
  let hash = 0
  for (let i = 0; i < name.length; i++) hash = (hash * 31 + name.charCodeAt(i)) >>> 0
  return SHADES[hash % SHADES.length]
}

export default function Avatar({ name = '?', size = 'md', className = '' }) {
  const initials = name.trim().slice(0, 2).toUpperCase() || '?'
  const sizes = {
    sm: 'h-8 w-8 text-xs',
    md: 'h-10 w-10 text-sm',
    lg: 'h-12 w-12 text-base',
  }
  return (
    <span
      aria-hidden="true"
      className={`inline-flex shrink-0 items-center justify-center rounded-full bg-gradient-to-br font-bold text-white ${shadeFor(name)} ${sizes[size] || sizes.md} ${className}`}
    >
      {initials}
    </span>
  )
}
