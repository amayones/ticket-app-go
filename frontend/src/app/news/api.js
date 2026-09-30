// Fungsi menu News (discovery publik, tanpa login).
import { apiRequest as request } from '../../api/client.js'

function qs(params = {}) {
  const q = new URLSearchParams()
  for (const [k, v] of Object.entries(params)) {
    if (v !== undefined && v !== null && String(v) !== '') q.append(k, String(v))
  }
  const s = q.toString()
  return s ? `?${s}` : ''
}

export async function newsCategories(signal) {
  const cats = await request('/api/discovery/news/categories', { signal })
  return Array.isArray(cats) ? cats : []
}

export async function newsList({ category, q, sort, limit, offset } = {}, signal) {
  const items = await request(`/api/discovery/news${qs({ category, q, sort, limit, offset })}`, { signal })
  return Array.isArray(items) ? items : []
}

export async function newsDetail(slug, signal) {
  return request(`/api/discovery/news/${encodeURIComponent(slug)}`, { signal })
}

// ===== Padanan store proxy langit_v2 (method read_data) =====
// read_data: dipakai load daftar berita (kategori + halaman list).
export async function read_data({ category, q, sort, limit, offset } = {}, signal) {
  const [cats, items] = await Promise.all([
    newsCategories(signal),
    newsList({ category, q, sort, limit, offset }, signal),
  ])
  return { categories: cats, items }
}
