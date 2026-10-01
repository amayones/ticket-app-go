import { apiRequest as request } from '../../api/client.js'

function qs(params = {}) {
  const q = new URLSearchParams()
  for (const [k, v] of Object.entries(params)) {
    if (v !== undefined && v !== null && String(v) !== '') q.append(k, String(v))
  }
  const s = q.toString()
  return s ? `?${s}` : ''
}

export async function listNews({ status, category, q, limit, offset } = {}, signal) {
  const items = await request(`/api/marketing/news${qs({ status, category, q, limit, offset })}`, { auth: true, signal })
  return Array.isArray(items) ? items : []
}

export async function newsDetail(code, signal) {
  return request(`/api/marketing/news/${code}`, { auth: true, signal })
}

export async function createNews(payload) {
  return request('/api/marketing/news', { method: 'POST', body: payload, auth: true })
}

export async function updateNews(code, payload) {
  return request(`/api/marketing/news/${code}`, { method: 'PUT', body: payload, auth: true })
}

export async function deleteNews(code) {
  return request(`/api/marketing/news/${code}`, { method: 'DELETE', auth: true })
}

export async function newsStatus(code, action) {
  return request(`/api/marketing/news/${code}/status`, { method: 'POST', body: { action }, auth: true })
}

export async function newsCategories(signal) {
  const items = await request('/api/marketing/news-categories', { auth: true, signal })
  return Array.isArray(items) ? items : []
}

export async function eventCategories(signal) {
  const items = await request('/api/marketing/event-categories', { auth: true, signal })
  return Array.isArray(items) ? items : []
}

export async function ownedEvents(signal) {
  const items = await request('/api/marketing/owned-events', { auth: true, signal })
  return Array.isArray(items) ? items : []
}

export async function read_data(params = {}, signal) {
  return listNews(params, signal)
}

export async function process_create(dtval) {
  return createNews(dtval)
}

export async function process_update(code, dtval) {
  return updateNews(code, dtval)
}

export async function process_delete(code) {
  return deleteNews(code)
}
