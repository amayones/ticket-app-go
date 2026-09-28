// Fungsi menu Audit Log.
import { apiRequest as request } from '../../api/client.js'

export async function listAudit({ action = '', entity = '', actor = '', limit = 20, offset = 0 } = {}) {
  const q = new URLSearchParams({ action, entity, actor, limit, offset })
  const data = await request(`/api/admin/audit?${q}`, { auth: true })
  return Array.isArray(data) ? data : []
}

// ===== Padanan store proxy langit_v2 (method read_data) =====
// read_data: dipakai store/load GRID viewer. Selalu kembalikan array.
export async function read_data({ action = '', entity = '', actor = '', limit = 20, offset = 0 } = {}) {
  return listAudit({ action, entity, actor, limit, offset })
}
