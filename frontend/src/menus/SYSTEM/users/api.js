// Fungsi menu User Account. Pola: raw request -> kembalikan data siap pakai.
import { apiRequest as request } from '../../../api/client.js'

export async function listUsers(limit = 50, offset = 0) {
  const data = await request(`/api/users?limit=${limit}&offset=${offset}`, { auth: true })
  return Array.isArray(data) ? data : []
}

export async function getUser(code) {
  return request(`/api/users/${code}`, { auth: true })
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
