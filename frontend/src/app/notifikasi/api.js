// Fungsi menu Notification Template & Log.
import { apiRequest as request } from '../../api/client.js'

export async function listTemplates(activeOnly = false) {
  const data = await request(`/api/admin/notifications/templates${activeOnly ? '?active=1' : ''}`, { auth: true })
  return Array.isArray(data) ? data : []
}

export async function createTemplate(payload) {
  return request('/api/admin/notifications/templates', { method: 'POST', body: payload, auth: true })
}

export async function updateTemplate(code, payload) {
  return request(`/api/admin/notifications/templates/${code}`, { method: 'PUT', body: payload, auth: true })
}

export async function deleteTemplate(code) {
  return request(`/api/admin/notifications/templates/${code}`, { method: 'DELETE', auth: true })
}

export async function sendNotification(payload) {
  return request('/api/admin/notifications/send', { method: 'POST', body: payload, auth: true })
}

export async function listNotifLogs(limit = 20, offset = 0) {
  const data = await request(`/api/admin/notifications/logs?limit=${limit}&offset=${offset}`, { auth: true })
  return Array.isArray(data) ? data : []
}

// ===== Padanan store proxy langit_v2 (method read_data/process_*) =====
// read_data: log pengiriman (GRID utama). Selalu kembalikan array.
export async function read_data({ limit = 20, offset = 0 } = {}) {
  return listNotifLogs(limit, offset)
}

// process_create/update/delete: template (FRM handler_btsave/btdelete).
export async function process_create(dtval) {
  return createTemplate(dtval)
}

export async function process_update(code, dtval) {
  return updateTemplate(code, dtval)
}

export async function process_delete(code) {
  return deleteTemplate(code)
}
