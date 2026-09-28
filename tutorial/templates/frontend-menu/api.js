// TEMPLATE proxy menu baru — padanan store proxy langit_v2.
// langit_v2: SATU url + method (read_data/process_create/update/delete/download).
// go-core: method dipetakan ke REST. GANTI <menu> dengan path endpoint
// (atau ganti seluruh fungsi bila backend-nya custom, pola tetap sama).
import { apiRequest as request } from '../../../api/client.js'

// read_data: GET list (dipakai store/load GRID). Selalu kembalikan array.
export async function read_data({ limit = 20, offset = 0 } = {}) {
  const data = await request(`/api/<menu>?limit=${limit}&offset=${offset}`, { auth: true })
  return Array.isArray(data) ? data : []
}

// process_create: POST tambah (dipakai FRM handler_btsave mode new).
export async function process_create(dtval) {
  return request('/api/<menu>', { method: 'POST', body: dtval, auth: true })
}

// process_update: PUT ubah (dipakai GRID handler_rowbtn_save + FRM edit).
export async function process_update(code, dtval) {
  return request(`/api/<menu>/${code}`, { method: 'PUT', body: dtval, auth: true })
}

// process_delete: DELETE hapus (dipakai GRID + FRM).
export async function process_delete(code) {
  return request(`/api/<menu>/${code}`, { method: 'DELETE', auth: true })
}
