// Fungsi menu Refunds (pengajuan refund milik pembeli yang login).
import { apiRequest as request } from '../../api/client.js'

export async function buyerRefunds(signal) {
  const items = await request('/api/ticketing/refunds', { auth: true, signal })
  return Array.isArray(items) ? items : []
}

export async function requestRefund(orderCode, reason) {
  return request('/api/ticketing/refunds', { method: 'POST', body: { order_code: orderCode, reason }, auth: true })
}

// ===== Padanan store proxy langit_v2 (method read_data/process_*) =====
// read_data: dipakai load daftar pengajuan milik sendiri.
export async function read_data(params = {}, signal) {
  return buyerRefunds(signal)
}

// process_create: dipakai form pengajuan (handler_btsave).
export async function process_create(dtval) {
  return requestRefund(dtval.order_code, dtval.reason)
}
