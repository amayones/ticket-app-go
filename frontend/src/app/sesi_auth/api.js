// Fungsi menu Authentication & Session Management.
import { apiRequest as request } from '../../api/client.js'

// Self-service: auth saja (tanpa permission menu), tersedia untuk semua role.
export async function listMySessions() {
  const data = await request('/api/sessions/mine', { auth: true })
  return Array.isArray(data) ? data : []
}

// Admin view: butuh permission MENU_SESSIONS.
export async function listAllSessions(limit = 20, offset = 0) {
  const data = await request(`/api/admin/sessions/all?limit=${limit}&offset=${offset}`, { auth: true })
  return Array.isArray(data) ? data : []
}

// Revoke sesi milik sendiri; sesi milik orang lain hanya bisa oleh role
// yang punya MENU_SESSIONS (divalidasi service dari permission caller).
export async function revokeSession(id) {
  return request(`/api/sessions/${id}`, { method: 'DELETE', auth: true })
}
