// Helper format waktu bersama (konsisten id-ID di semua menu).
export function formatTime(iso) {
  try {
    return new Date(iso).toLocaleString('id-ID', { dateStyle: 'medium', timeStyle: 'short' })
  } catch {
    return iso
  }
}
