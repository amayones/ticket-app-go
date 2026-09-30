// Fungsi menu Promotions (discovery publik, tanpa login).
import { apiRequest as request } from '../../api/client.js'

function qs(params = {}) {
  const q = new URLSearchParams()
  for (const [k, v] of Object.entries(params)) {
    if (v !== undefined && v !== null && String(v) !== '') q.append(k, String(v))
  }
  const s = q.toString()
  return s ? `?${s}` : ''
}

export async function promotionsCategories(signal) {
  const cats = await request('/api/discovery/categories', { signal })
  return Array.isArray(cats) ? cats : []
}

export async function promotionsList({ mode, category, q, limit, offset } = {}, signal) {
  const items = await request(`/api/discovery/promotions${qs({ mode, category, q, limit, offset })}`, { signal })
  return Array.isArray(items) ? items : []
}

// ===== Padanan store proxy langit_v2 (method read_data) =====
// read_data: dipakai load daftar promo (kategori + halaman list).
export async function read_data({ mode, category, q, limit, offset } = {}, signal) {
  const [cats, items] = await Promise.all([
    promotionsCategories(signal),
    promotionsList({ mode, category, q, limit, offset }, signal),
  ])
  return { categories: cats, items }
}
