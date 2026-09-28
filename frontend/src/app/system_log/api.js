// Fungsi menu Error / System Log.
import { apiRequest as request } from '../../api/client.js'

export async function listSyslogs({ level = '', limit = 20, offset = 0 } = {}) {
  const q = new URLSearchParams({ level, limit, offset })
  const data = await request(`/api/admin/syslogs?${q}`, { auth: true })
  return Array.isArray(data) ? data : []
}

export async function pruneSyslogs(days = 30) {
  return request(`/api/admin/syslogs?days=${days}`, { method: 'DELETE', auth: true })
}

// ===== Padanan store proxy langit_v2 (method read_data) =====
// read_data: dipakai store/load GRID viewer. Selalu kembalikan array.
export async function read_data({ level = '', limit = 20, offset = 0 } = {}) {
  return listSyslogs({ level, limit, offset })
}
