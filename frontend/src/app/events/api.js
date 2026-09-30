// Fungsi menu Events (discovery publik, tanpa login).
import { apiRequest as request } from '../../api/client.js'

function qs(params = {}) {
  const q = new URLSearchParams()
  for (const [k, v] of Object.entries(params)) {
    if (v !== undefined && v !== null && String(v) !== '') q.append(k, String(v))
  }
  const s = q.toString()
  return s ? `?${s}` : ''
}

export async function eventsCategories(signal) {
  const cats = await request('/api/discovery/categories', { signal })
  return Array.isArray(cats) ? cats : []
}

export async function eventsCities(signal) {
  const cities = await request('/api/discovery/cities', { signal })
  return Array.isArray(cities) ? cities : []
}

export async function eventsList({ categories, city, from, to, min_price, max_price, q, sort, limit, offset } = {}, signal) {
  const items = await request(
    `/api/discovery/events${qs({ categories, city, from, to, min_price, max_price, q, sort, limit, offset })}`,
    { signal },
  )
  return Array.isArray(items) ? items : []
}

export async function eventDetail(code, signal) {
  return request(`/api/discovery/events/${encodeURIComponent(code)}`, { signal })
}

// ===== Padanan store proxy langit_v2 (method read_data) =====
// read_data: dipakai load daftar event (opsi filter + halaman list).
export async function read_data(params = {}, signal) {
  const [cats, cities, items] = await Promise.all([
    eventsCategories(signal),
    eventsCities(signal),
    eventsList(params, signal),
  ])
  return { categories: cats, cities, items }
}
