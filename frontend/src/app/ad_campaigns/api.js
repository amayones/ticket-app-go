import { apiRequest as request } from '../../api/client.js'

function qs(params = {}) {
  const q = new URLSearchParams()
  for (const [k, v] of Object.entries(params)) {
    if (v !== undefined && v !== null && String(v) !== '') q.append(k, String(v))
  }
  const s = q.toString()
  return s ? `?${s}` : ''
}

export async function listAds({ status, q, limit, offset } = {}, signal) {
  const items = await request(`/api/marketing/ads${qs({ status, q, limit, offset })}`, { auth: true, signal })
  return Array.isArray(items) ? items : []
}

export async function adDetail(code, signal) {
  return request(`/api/marketing/ads/${code}`, { auth: true, signal })
}

export async function adStats(code, signal) {
  return request(`/api/marketing/ads/${code}/stats`, { auth: true, signal })
}

export async function createAd(payload) {
  return request('/api/marketing/ads', { method: 'POST', body: payload, auth: true })
}

export async function updateAd(code, payload) {
  return request(`/api/marketing/ads/${code}`, { method: 'PUT', body: payload, auth: true })
}

export async function adStatus(code, action, note) {
  return request(`/api/marketing/ads/${code}/status`, { method: 'POST', body: { action, note }, auth: true })
}

export async function simulatePay(code) {
  return request(`/api/marketing/ads/${code}/simulate-pay`, { method: 'POST', auth: true })
}

export async function adPositions(signal) {
  const items = await request('/api/marketing/ad-positions', { auth: true, signal })
  return Array.isArray(items) ? items : []
}

export async function ownedEvents(signal) {
  const items = await request('/api/marketing/owned-events', { auth: true, signal })
  return Array.isArray(items) ? items : []
}

export async function uploadBanner(file) {
  const token = (() => {
    try {
      return localStorage.getItem('access_token') || ''
    } catch {
      return ''
    }
  })()
  const form = new FormData()
  form.append('banner', file)
  const res = await fetch('/api/marketing/ads/upload-banner', {
    method: 'POST',
    headers: { Authorization: `Bearer ${token}` },
    body: form,
  })
  let data = null
  try {
    data = await res.json()
  } catch {
    data = null
  }
  if (!res.ok) {
    const err = new Error((data && data.error) || `Upload gagal (${res.status})`)
    err.data = data
    err.status = res.status
    throw err
  }
  return data
}

export async function read_data(params = {}, signal) {
  return listAds(params, signal)
}

export async function process_create(dtval) {
  return createAd(dtval)
}

export async function process_update(code, dtval) {
  return updateAd(code, dtval)
}
