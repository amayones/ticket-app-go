// Fungsi menu Audit Log.
import { apiRequest as request } from '../../api/client.js'

export async function listAudit({ action = '', entity = '', actor = '', limit = 20, offset = 0 } = {}) {
  const q = new URLSearchParams({ action, entity, actor, limit, offset })
  const data = await request(`/api/admin/audit?${q}`, { auth: true })
  return Array.isArray(data) ? data : []
}
