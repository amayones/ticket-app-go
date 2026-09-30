// Fungsi menu My Tickets (tiket milik pembeli yang login).
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

export async function myTickets({ status, q, limit, offset } = {}, signal) {
  const items = await request(`/api/ticketing/tickets${qs({ status, q, limit, offset })}`, { auth: true, signal })
  return Array.isArray(items) ? items : []
}

export async function ticketDetail(code, signal) {
  return request(`/api/ticketing/tickets/${code}`, { auth: true, signal })
}

// URL gambar QR untuk <img> (token via query karena <img> tak bisa header).
export function ticketQrUrl(code) {
  return `/api/ticketing/tickets/${code}/qr?access_token=${encodeURIComponent(token())}`
}

export async function requestRefund(orderCode, reason) {
  return request('/api/ticketing/refunds', { method: 'POST', body: { order_code: orderCode, reason }, auth: true })
}

// ===== Padanan store proxy langit_v2 (method read_data/process_*) =====
// read_data: dipakai load daftar tiket milik sendiri.
export async function read_data(params = {}, signal) {
  return myTickets(params, signal)
}

// process_refund: dipakai tombol Request Refund di modal detail.
export async function process_refund(orderCode, reason) {
  return requestRefund(orderCode, reason)
}
