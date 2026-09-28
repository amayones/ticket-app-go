// Fungsi menu Role & Permission (RBAC).
import { apiRequest as request } from '../../api/client.js'

export async function listRoles() {
  return request('/api/roles', { auth: true })
}

export async function getRole(code) {
  return request(`/api/admin/roles/${code}`, { auth: true })
}

export async function createRole(code, name) {
  return request('/api/admin/roles', { method: 'POST', body: { code, name }, auth: true })
}

export async function deleteRole(code) {
  return request(`/api/admin/roles/${code}`, { method: 'DELETE', auth: true })
}

export async function listPermissions() {
  return request('/api/admin/permissions', { auth: true })
}

export async function setRolePermissions(code, permissions) {
  return request(`/api/admin/roles/${code}/permissions`, { method: 'PUT', body: { permissions }, auth: true })
}

// ===== Padanan store proxy langit_v2 (method read_data/process_*) =====
// read_data: dipakai store/load (role + permission menu).
export async function read_data() {
  const [r, p] = await Promise.all([listRoles(), listPermissions()])
  return {
    roles: Array.isArray(r) ? r : [],
    permissions: Array.isArray(p) ? p.filter((permission) => permission.code.startsWith('MENU_')) : [],
  }
}

// process_create: dipakai form role baru (handler_btsave).
export async function process_create(dtval) {
  return createRole(dtval.code, dtval.name)
}

// process_delete: dipakai hapus role (handler_btdelete).
export async function process_delete(code) {
  return deleteRole(code)
}
