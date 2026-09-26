// Fungsi menu Authentication & Session Management.
import { apiRequest as request } from '../../../api/client.js'

export async function listMySessions() {
  // Self-service (auth saja); fallback ke endpoint admin untuk backend lama.
  try {
    const data = await request('/api/sessions/mine', { auth: true })
    return Array.isArray(data) ? data : []
  } catch {
    const data = await request('/api/admin/sessions', { auth: true })
    return Array.isArray(data) ? data : []
  }
}

export async function listAllSessions(limit = 20, offset = 0) {
  const data = await request(`/api/admin/sessions/all?limit=${limit}&offset=${offset}`, { auth: true })
  return Array.isArray(data) ? data : []
}

export async function revokeSession(id) {
  // Self-service dulu (milik sendiri), fallback ke admin untuk backend lama.
  try {
    return await request(`/api/sessions/${id}`, { method: 'DELETE', auth: true })
  } catch {
    return request(`/api/admin/sessions/${id}`, { method: 'DELETE', auth: true })
  }
}
