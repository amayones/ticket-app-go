// Fungsi menu Role & Permission (RBAC).
import { apiRequest as request } from '../../../api/client.js'

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
