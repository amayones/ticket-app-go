// Fungsi menu Seller Dashboard.
import { apiRequest as request } from '../../api/client.js'

function qs(params = {}) {
  const q = new URLSearchParams()
  for (const [k, v] of Object.entries(params)) {
    if (v !== undefined && v !== null && String(v) !== '') q.append(k, String(v))
  }
  const s = q.toString()
  return s ? `?${s}` : ''
}

function token() {
  try {
    return localStorage.getItem('access_token') || ''
  } catch {
    return ''
  }
}

export async function sellerSummary({ preset, from, to, category } = {}, signal) {
  return request(`/api/dashboard/seller/summary${qs({ preset, from, to, category })}`, { auth: true, signal })
}

export async function sellerOrders({ preset, from, to, category, limit } = {}, signal) {
  const orders = await request(`/api/dashboard/seller/orders${qs({ preset, from, to, category, limit })}`, {
    auth: true,
    signal,
  })
  return Array.isArray(orders) ? orders : []
}

export async function sellerActionable(signal) {
  const orders = await request('/api/dashboard/seller/orders/actionable', { auth: true, signal })
  return Array.isArray(orders) ? orders : []
}

export async function sellerLowStock(signal) {
  const products = await request('/api/dashboard/seller/products/low-stock', { auth: true, signal })
  return Array.isArray(products) ? products : []
}

export async function sellerComplaints(signal) {
  const complaints = await request('/api/dashboard/seller/complaints', { auth: true, signal })
  return Array.isArray(complaints) ? complaints : []
}

// Export CSV: fetch mentah + blob (apiRequest selalu mem-parsing JSON).
export async function exportSummary(params = {}) {
  const res = await fetch(`/api/dashboard/seller/export${qs({ format: 'csv', ...params })}`, {
    headers: { Authorization: `Bearer ${token()}` },
  })
  if (!res.ok) {
    let message = `Export gagal (${res.status})`
    try {
      const data = await res.json()
      if (data && data.error) message = data.error
    } catch {
      // abaikan (bukan JSON)
    }
    throw new Error(message)
  }
  return res.blob()
}

// ===== Padanan store proxy langit_v2 (method read_data) =====
// read_data: dipakai load dashboard (satu ringkasan untuk semua bagian).
export async function read_data(params = {}, signal) {
  const summary = await sellerSummary(params, signal)
  return summary || null
}
