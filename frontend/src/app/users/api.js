// Fungsi menu User Account. Pola: raw request -> kembalikan data siap pakai.
import { apiRequest as request } from '../../api/client.js'

export async function listUsers(limit = 50, offset = 0) {
  const data = await request(`/api/users?limit=${limit}&offset=${offset}`, { auth: true })
  return Array.isArray(data) ? data : []
}

// Tambah user dari menu User Account. Akses menu diberikan oleh permission.
export async function createUser(username, email, password, roleCode) {
  const body = { username, email, password }
  if (roleCode) body.role_code = roleCode
  return request('/api/users', { method: 'POST', body, auth: true })
}

// Patch parsial: kirim hanya field yang berubah { username?, email?, password? }.
export async function updateUser(code, patch) {
  return request(`/api/users/${code}`, { method: 'PUT', body: patch, auth: true })
}

export async function deleteUser(code) {
  return request(`/api/users/${code}`, { method: 'DELETE', auth: true })
}

export async function logoutAll(code) {
  return request(`/api/users/${code}/logout-all`, { method: 'POST', auth: true })
}

export async function updateUserRole(code, roleCode) {
  return request(`/api/users/${code}/role`, { method: 'PUT', body: { role_code: roleCode }, auth: true })
}

// ===== Padanan store proxy langit_v2 (method read_data/process_*) =====
// read_data: dipakai store/load GRID. Selalu kembalikan array.
export async function read_data({ limit = 20, offset = 0 } = {}) {
  return listUsers(limit, offset)
}

// process_create: dipakai FRM handler_btsave mode new (dtval = form values).
export async function process_create(dtval) {
  return createUser(dtval.username?.trim(), dtval.email?.trim().toLowerCase(), dtval.password, dtval.role)
}

// process_update: dipakai GRID handler_rowbtn_save + FRM edit.
export async function process_update(code, dtval) {
  return updateUser(code, dtval)
}

// process_delete: dipakai GRID + FRM handler delete.
export async function process_delete(code) {
  return deleteUser(code)
}
