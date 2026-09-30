// Fungsi menu Ticket Types & Pricing (tipe tiket satu event).
import { apiRequest as request } from '../../api/client.js'

export async function ticketTypes(eventCode, signal) {
  const data = await request(`/api/events/${eventCode}/ticket-types`, { auth: true, signal })
  return data || { types: [], summary: {} }
}

export async function createType(eventCode, payload) {
  return request(`/api/events/${eventCode}/ticket-types`, { method: 'POST', body: payload, auth: true })
}

export async function updateType(eventCode, typeCode, payload) {
  return request(`/api/events/${eventCode}/ticket-types/${typeCode}`, { method: 'PUT', body: payload, auth: true })
}

export async function setTypeActive(eventCode, typeCode, active) {
  return request(`/api/events/${eventCode}/ticket-types/${typeCode}/${active ? 'activate' : 'deactivate'}`, {
    method: 'POST',
    auth: true,
  })
}

export async function deleteType(eventCode, typeCode) {
  return request(`/api/events/${eventCode}/ticket-types/${typeCode}`, { method: 'DELETE', auth: true })
}

export async function reorderTypes(eventCode, order) {
  return request(`/api/events/${eventCode}/ticket-types/order`, { method: 'PUT', body: { order }, auth: true })
}

export async function publishEvent(eventCode) {
  return request(`/api/events/${eventCode}/publish`, { method: 'POST', auth: true })
}

// ===== Padanan store proxy langit_v2 (method read_data/process_*) =====
// read_data: dipakai load daftar + ringkasan tipe tiket event.
export async function read_data({ eventCode } = {}, signal) {
  if (!eventCode) return { types: [], summary: {} }
  return ticketTypes(eventCode, signal)
}

// process_create/update/delete: dipakai FRM tipe tiket.
export async function process_create(eventCode, dtval) {
  return createType(eventCode, dtval)
}

export async function process_update(eventCode, typeCode, dtval) {
  return updateType(eventCode, typeCode, dtval)
}

export async function process_delete(eventCode, typeCode) {
  return deleteType(eventCode, typeCode)
}
