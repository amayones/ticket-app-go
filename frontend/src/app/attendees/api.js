// Fungsi menu Attendees (daftar peserta satu event, read-only).
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

export async function attendeeTypes(eventCode, signal) {
  const data = await request(`/api/events/${eventCode}/ticket-types`, { auth: true, signal })
  return Array.isArray(data && data.types) ? data.types : []
}

export async function attendeeList({ eventCode, type, attendance, payment, q, limit, offset } = {}, signal) {
  const data = await request(
    `/api/events/${eventCode}/attendees${qs({ type, attendance, payment, q, limit, offset })}`,
    { auth: true, signal },
  )
  return data || { attendees: [], summary: {} }
}

// Export CSV: fetch mentah + blob (apiRequest selalu mem-parsing JSON).
export async function exportAttendees({ eventCode, type, attendance, payment, q } = {}) {
  const res = await fetch(`/api/events/${eventCode}/attendees/export${qs({ type, attendance, payment, q })}`, {
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
// read_data: dipakai load daftar peserta + ringkasan + opsi tipe.
export async function read_data({ eventCode, type, attendance, payment, q, limit, offset } = {}, signal) {
  if (!eventCode) return { attendees: [], summary: {}, types: [] }
  const [list, types] = await Promise.all([
    attendeeList({ eventCode, type, attendance, payment, q, limit, offset }, signal),
    attendeeTypes(eventCode, signal),
  ])
  return { attendees: list.attendees || [], summary: list.summary || {}, types }
}
