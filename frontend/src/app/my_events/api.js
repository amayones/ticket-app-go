// Fungsi menu My Events (event milik penyelenggara yang login).
import { apiRequest as request } from '../../api/client.js'

function qs(params = {}) {
  const q = new URLSearchParams()
  for (const [k, v] of Object.entries(params)) {
    if (v !== undefined && v !== null && String(v) !== '') q.append(k, String(v))
  }
  const s = q.toString()
  return s ? `?${s}` : ''
}

export async function myEvents({ status, category, city, from, to, q, sort, limit, offset } = {}, signal) {
  const items = await request(`/api/events/mine${qs({ status, category, city, from, to, q, sort, limit, offset })}`, {
    auth: true,
    signal,
  })
  return Array.isArray(items) ? items : []
}

export async function duplicateEvent(code) {
  return request(`/api/events/${code}/duplicate`, { method: 'POST', auth: true })
}

export async function publishEvent(code) {
  return request(`/api/events/${code}/publish`, { method: 'POST', auth: true })
}

export async function unpublishEvent(code) {
  return request(`/api/events/${code}/unpublish`, { method: 'POST', auth: true })
}

export async function cancelEvent(code, reason) {
  return request(`/api/events/${code}/cancel`, { method: 'POST', body: { reason }, auth: true })
}

export async function deleteEvent(code) {
  return request(`/api/events/${code}`, { method: 'DELETE', auth: true })
}

export async function createEvent(payload) {
  return request('/api/events', { method: 'POST', body: payload, auth: true })
}

export async function updateEvent(code, payload) {
  return request(`/api/events/${code}`, { method: 'PUT', body: payload, auth: true })
}

// ===== Padanan store proxy langit_v2 (method read_data/process_*) =====
// read_data: dipakai load daftar event milik sendiri.
export async function read_data(params = {}, signal) {
  return myEvents(params, signal)
}

// Event tunggal untuk FRM edit (read + upload poster).
export async function getEvent(code, signal) {
  return request(`/api/events/${code}`, { auth: true, signal })
}

export async function uploadPoster(file) {
  const token = (() => {
    try {
      return localStorage.getItem('access_token') || ''
    } catch {
      return ''
    }
  })()
  const form = new FormData()
  form.append('poster', file)
  const res = await fetch('/api/events/upload-poster', {
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

// process_publish/unpublish/cancel/delete/duplicate: dipakai tombol aksi GRID.
export async function process_publish(code) {
  return publishEvent(code)
}

export async function process_unpublish(code) {
  return unpublishEvent(code)
}

export async function process_cancel(code, reason) {
  return cancelEvent(code, reason)
}

export async function process_delete(code) {
  return deleteEvent(code)
}

export async function process_duplicate(code) {
  return duplicateEvent(code)
}

// process_create: dipakai FRM handler_btsave mode new (dtval = form values).
export async function process_create(dtval) {
  return createEvent(dtval)
}

// process_update: dipakai FRM handler_btsave mode edit.
export async function process_update(code, dtval) {
  return updateEvent(code, dtval)
}
