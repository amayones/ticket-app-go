// Fungsi menu Error / System Log.
import { apiRequest as request } from '../../../api/client.js'

export async function listSyslogs({ level = '', limit = 20, offset = 0 } = {}) {
  const q = new URLSearchParams({ level, limit, offset })
  const data = await request(`/api/admin/syslogs?${q}`, { auth: true })
  return Array.isArray(data) ? data : []
}

export async function pruneSyslogs(days = 30) {
  return request(`/api/admin/syslogs?days=${days}`, { method: 'DELETE', auth: true })
}
