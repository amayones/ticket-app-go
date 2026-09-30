// Fungsi menu Checkout (alasan beli tiket: pilih qty, data pemesan,
// promo, order, bayar simulasi).
import { apiRequest as request } from '../../api/client.js'

export async function checkoutEvent(eventCode, signal) {
  return request(`/api/discovery/events/${encodeURIComponent(eventCode)}`, { signal })
}

export async function checkoutOrder(payload) {
  return request('/api/ticketing/checkout', { method: 'POST', body: payload, auth: true })
}

export async function payOrder(orderCode, { holder, email, phone } = {}) {
  const q = new URLSearchParams()
  if (holder) q.append('holder', holder)
  if (email) q.append('email', email)
  if (phone) q.append('phone', phone)
  const s = q.toString()
  return request(`/api/ticketing/orders/${orderCode}/pay${s ? `?${s}` : ''}`, {
    method: 'POST',
    body: { simulate: true },
    auth: true,
  })
}

export async function cancelOrder(orderCode) {
  return request(`/api/ticketing/orders/${orderCode}/cancel`, { method: 'POST', auth: true })
}
