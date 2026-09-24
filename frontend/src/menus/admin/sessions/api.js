// Fungsi menu Authentication & Session Management.
import { apiRequest as request } from '../../../api/client.js'

export async function listMySessions() {
  const data = await request('/api/admin/sessions', { auth: true })
  return Array.isArray(data) ? data : []
}

export async function listAllSessions(limit = 20, offset = 0) {
  const data = await request(`/api/admin/sessions/all?limit=${limit}&offset=${offset}`, { auth: true })
  return Array.isArray(data) ? data : []
}

export async function revokeSession(id) {
  return request(`/api/admin/sessions/${id}`, { method: 'DELETE', auth: true })
}
