// Helper format waktu bersama (konsisten id-ID di semua menu).
export function formatTime(iso) {
  try {
    return new Date(iso).toLocaleString('id-ID', { dateStyle: 'medium', timeStyle: 'short' })
  } catch {
    return iso
  }
}

// Padanan formatAmount/formatDate controller langit_v2
// (dipakai controller.js tiap menu + renderer kolom GRID).
export function formatAmount(value) {
  const num = Number(value)
  if (value == null || value === '' || Number.isNaN(num)) return ''
  return num.toLocaleString('id-ID', { minimumFractionDigits: 2, maximumFractionDigits: 2 })
}

export function formatDate(value) {
  if (value == null || value === '') return ''
  try {
    const d = new Date(value)
    if (Number.isNaN(d.getTime())) return ''
    const p = (n) => String(n).padStart(2, '0')
    return `${d.getFullYear()}-${p(d.getMonth() + 1)}-${p(d.getDate())} ${p(d.getHours())}:${p(d.getMinutes())}:${p(d.getSeconds())}`
  } catch {
    return ''
  }
}
