// Fungsi menu Scan History (riwayat + export, read-only).
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

export async function scanEventsList(signal) {
  const events = await request('/api/ticketing/scan/events', { auth: true, signal })
  return Array.isArray(events) ? events : []
}

export async function scanHistory(params = {}, signal) {
  const data = await request(`/api/ticketing/scan/history${qs(params)}`, { auth: true, signal })
  return data || { logs: [], summary: {} }
}

// Export CSV: fetch mentah + blob (apiRequest selalu mem-parsing JSON).
export async function exportHistory(params = {}) {
  const res = await fetch(`/api/ticketing/scan/export${qs(params)}`, {
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
// read_data: dipakai load riwayat + ringkasan + opsi event.
export async function read_data(params = {}, signal) {
  const [history, events] = await Promise.all([
    scanHistory(params, signal),
    scanEventsList(signal),
  ])
  return { logs: history.logs || [], summary: history.summary || {}, events }
}
