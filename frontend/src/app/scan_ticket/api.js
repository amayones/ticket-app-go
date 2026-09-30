// Fungsi menu Scan Ticket (validasi QR/kode di pintu masuk).
import { apiRequest as request } from '../../api/client.js'

function qs(params = {}) {
  const q = new URLSearchParams()
  for (const [k, v] of Object.entries(params)) {
    if (v !== undefined && v !== null && String(v) !== '') q.append(k, String(v))
  }
  const s = q.toString()
  return s ? `?${s}` : ''
}

export async function scanEvents(signal) {
  const events = await request('/api/ticketing/scan/events', { auth: true, signal })
  return Array.isArray(events) ? events : []
}

export async function scanSummary(eventCode, signal) {
  return request(`/api/ticketing/scan/summary${qs({ event: eventCode })}`, { auth: true, signal })
}

export async function scanTicket({ token, code, event }) {
  return request('/api/ticketing/scan', { method: 'POST', body: { token, code, event }, auth: true })
}

// ===== Padanan store proxy langit_v2 (method read_data/process_*) =====
// read_data: dipakai memuat event yang boleh di-scan + ringkasan masuk.
export async function read_data({ event } = {}, signal) {
  const [events, summary] = await Promise.all([
    scanEvents(signal),
    event ? scanSummary(event, signal).catch(() => null) : Promise.resolve(null),
  ])
  return { events, summary }
}

// process_scan: dipakai tiap tembakan scan (kamera/manual).
export async function process_scan({ token, code, event }) {
  return scanTicket({ token, code, event })
}
