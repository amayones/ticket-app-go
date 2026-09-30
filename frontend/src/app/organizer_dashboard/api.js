// Fungsi menu Organizer Dashboard.
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

export async function organizerSummary({ preset, from, to, event, city } = {}, signal) {
  return request(`/api/dashboard/organizer/summary${qs({ preset, from, to, event, city })}`, { auth: true, signal })
}

export async function organizerOrders({ preset, from, to, event, city, limit } = {}, signal) {
  const orders = await request(`/api/dashboard/organizer/orders${qs({ preset, from, to, event, city, limit })}`, {
    auth: true,
    signal,
  })
  return Array.isArray(orders) ? orders : []
}

export async function organizerRefunds(signal) {
  const refunds = await request('/api/dashboard/organizer/refunds', { auth: true, signal })
  return Array.isArray(refunds) ? refunds : []
}

export async function checkinLive(event) {
  return request(`/api/dashboard/organizer/checkin-live${qs({ event })}`, { auth: true })
}

// Export biner (CSV/PDF): pakai fetch mentah + blob karena apiRequest
// selalu mem-parsing JSON. Token dibaca seperti store client.js.
export async function exportSummary(format, params = {}) {
  const res = await fetch(`/api/dashboard/organizer/export${qs({ format, ...params })}`, {
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

// ===== Padanan store proxy langit_v2 (method read_data/process_*) =====
// read_data: dipakai load dashboard (ringkasan + refund + check-in hari ini
// menyusul di halaman agar satu bagian gagal tidak merusak yang lain).
export async function read_data(params = {}, signal) {
  const summary = await organizerSummary(params, signal)
  return summary || null
}

// process_approve: dipakai tombol Setujui refund (handler_rowbtn_save).
export async function process_approve(code) {
  return request(`/api/dashboard/organizer/refunds/${code}/approve`, { method: 'POST', auth: true })
}

// process_reject: dipakai tombol Tolak refund (handler_rowbtn_delete).
export async function process_reject(code) {
  return request(`/api/dashboard/organizer/refunds/${code}/reject`, { method: 'POST', auth: true })
}
