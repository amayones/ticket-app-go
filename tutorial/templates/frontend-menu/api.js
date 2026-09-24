// TEMPLATE fungsi menu baru. Pola: raw request -> kembalikan data siap pakai.
// Satu file ini milik 1 menu; endpoint backend-nya dibuat mengikuti
// tutorial/backend di tutorial/README.md (atau pakai endpoint yang sudah ada).
import { apiRequest as request } from '../../../api/client.js'

// GET /api/<menu>?limit=&offset= -> selalu kembalikan array.
export async function listItems(limit = 20, offset = 0) {
  const data = await request(`/api/<menu>?limit=${limit}&offset=${offset}`, { auth: true })
  return Array.isArray(data) ? data : []
}

// GET satu item.
export async function getItem(code) {
  return request(`/api/<menu>/${code}`, { auth: true })
}

// POST tambah item.
export async function createItem(payload) {
  return request('/api/<menu>', { method: 'POST', body: payload, auth: true })
}

// PUT ubah item (kirim hanya field yang berubah).
export async function updateItem(code, patch) {
  return request(`/api/<menu>/${code}`, { method: 'PUT', body: patch, auth: true })
}

// DELETE hapus item.
export async function deleteItem(code) {
  return request(`/api/<menu>/${code}`, { method: 'DELETE', auth: true })
}
