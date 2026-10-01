import { apiRequest as request } from '../../api/client.js'

function qs(params = {}) {
  const q = new URLSearchParams()
  for (const [k, v] of Object.entries(params)) {
    if (v !== undefined && v !== null && String(v) !== '') q.append(k, String(v))
  }
  const s = q.toString()
  return s ? `?${s}` : ''
}

export async function listPromos({ status, q, limit, offset } = {}, signal) {
  const items = await request(`/api/marketing/promos${qs({ status, q, limit, offset })}`, { auth: true, signal })
  return Array.isArray(items) ? items : []
}

export async function promoDetail(code, signal) {
  return request(`/api/marketing/promos/${code}`, { auth: true, signal })
}

export async function createPromo(payload) {
  return request('/api/marketing/promos', { method: 'POST', body: payload, auth: true })
}

export async function updatePromo(code, payload) {
  return request(`/api/marketing/promos/${code}`, { method: 'PUT', body: payload, auth: true })
}

export async function deletePromo(code) {
  return request(`/api/marketing/promos/${code}`, { method: 'DELETE', auth: true })
}

export async function togglePromo(code, active) {
  return request(`/api/marketing/promos/${code}/toggle`, { method: 'POST', body: { active }, auth: true })
}

export async function generatePromoCode() {
  const data = await request('/api/marketing/promos-generate', { auth: true })
  return data.code || data.promo_code || ''
}

export async function promoOptions(signal) {
  try {
    const cats = await request('/api/marketing/event-categories', { auth: true, signal })
    const evs = await request('/api/marketing/owned-events', { auth: true, signal })
    return { categories: Array.isArray(cats) ? cats : [], events: Array.isArray(evs) ? evs : [] }
  } catch {
    return { categories: [], events: [] }
  }
}

export async function read_data(params = {}, signal) {
  return listPromos(params, signal)
}

export async function process_create(dtval) {
  return createPromo(dtval)
}

export async function process_update(code, dtval) {
  return updatePromo(code, dtval)
}

export async function process_delete(code) {
  return deletePromo(code)
}
