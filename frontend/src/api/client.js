// Central API client: base URL from VITE_API_URL, JSON envelope,
// auto refresh-token rotation on 401, no-store for auth calls.
const BASE = (import.meta.env.VITE_API_URL || '').replace(/\/$/, '')

const store = {
  get access() {
    return localStorage.getItem('access_token') || ''
  },
  get refresh() {
    return localStorage.getItem('refresh_token') || ''
  },
  set(access, refresh) {
    if (access) localStorage.setItem('access_token', access)
    if (refresh) localStorage.setItem('refresh_token', refresh)
  },
  clear() {
    localStorage.removeItem('access_token')
    localStorage.removeItem('refresh_token')
  },
}

async function request(path, { method = 'GET', body, auth = false, retry = true } = {}) {
  const headers = { 'Content-Type': 'application/json' }
  if (auth && store.access) headers.Authorization = `Bearer ${store.access}`
  const res = await fetch(`${BASE}${path}`, {
    method,
    headers,
    body: body ? JSON.stringify(body) : undefined,
  })
  if (res.status === 401 && auth && retry && store.refresh) {
    const ok = await tryRefresh()
    if (ok) return request(path, { method, body, auth, retry: false })
    store.clear()
  }
  let data = null
  try {
    data = await res.json()
  } catch {
    data = null
  }
  if (!res.ok) {
    const msg = (data && data.error) || `Request failed (${res.status})`
    throw new Error(msg)
  }
  return data
}

async function tryRefresh() {
  try {
    const res = await fetch(`${BASE}/api/refresh`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ refresh_token: store.refresh }),
    })
    if (!res.ok) return false
    const data = await res.json()
    store.set(data.access_token, data.refresh_token)
    return true
  } catch {
    return false
  }
}

function parseJwt(token) {
  try {
    const payload = token.split('.')[1].replace(/-/g, '+').replace(/_/g, '/')
    return JSON.parse(atob(payload))
  } catch {
    return null
  }
}

export const api = {
  // Profil user yang sedang login, dibaca dari klaim JWT (tanpa request).
  currentUser() {
    if (!store.access) return null
    const claims = parseJwt(store.access)
    if (!claims || !claims.user_code) return null
    return { code: claims.user_code, username: claims.username || '', role: claims.role || '' }
  },
  async register(username, email, password) {
    return request('/api/users', { method: 'POST', body: { username, email, password } })
  },
  async login(username, password) {
    const data = await request('/api/login', { method: 'POST', body: { username, password } })
    store.set(data.access_token, data.refresh_token)
    return data
  },
  async logout() {
    try {
      if (store.refresh) {
        await request('/api/logout', { method: 'POST', body: { refresh_token: store.refresh } })
      }
    } finally {
      store.clear()
    }
  },
  async listUsers(limit = 50, offset = 0) {
    const data = await request(`/api/users?limit=${limit}&offset=${offset}`, { auth: true })
    return Array.isArray(data) ? data : []
  },
  async getUser(code) {
    return request(`/api/users/${code}`, { auth: true })
  },
  // Patch parsial: kirim hanya field yang berubah { username?, email?, password? }.
  async updateUser(code, patch) {
    return request(`/api/users/${code}`, { method: 'PUT', body: patch, auth: true })
  },
  async deleteUser(code) {
    return request(`/api/users/${code}`, { method: 'DELETE', auth: true })
  },
  async logoutAll(code) {
    return request(`/api/users/${code}/logout-all`, { method: 'POST', auth: true })
  },
  async listRoles() {
    return request('/api/roles', { auth: true })
  },
  // --- Role & Permission (RBAC, admin) ---
  async getRole(code) {
    return request(`/api/admin/roles/${code}`, { auth: true })
  },
  async createRole(code, name) {
    return request('/api/admin/roles', { method: 'POST', body: { code, name }, auth: true })
  },
  async deleteRole(code) {
    return request(`/api/admin/roles/${code}`, { method: 'DELETE', auth: true })
  },
  async listPermissions() {
    return request('/api/admin/permissions', { auth: true })
  },
  async setRolePermissions(code, permissions) {
    return request(`/api/admin/roles/${code}/permissions`, { method: 'PUT', body: { permissions }, auth: true })
  },
  async updateUserRole(code, roleCode) {
    return request(`/api/users/${code}/role`, { method: 'PUT', body: { role_code: roleCode }, auth: true })
  },
  // --- Authentication & Session Management ---
  async listMySessions() {
    const data = await request('/api/admin/sessions', { auth: true })
    return Array.isArray(data) ? data : []
  },
  async listAllSessions(limit = 20, offset = 0) {
    const data = await request(`/api/admin/sessions/all?limit=${limit}&offset=${offset}`, { auth: true })
    return Array.isArray(data) ? data : []
  },
  async revokeSession(id) {
    return request(`/api/admin/sessions/${id}`, { method: 'DELETE', auth: true })
  },
  // --- Audit Log ---
  async listAudit({ action = '', entity = '', actor = '', limit = 20, offset = 0 } = {}) {
    const q = new URLSearchParams({ action, entity, actor, limit, offset })
    const data = await request(`/api/admin/audit?${q}`, { auth: true })
    return Array.isArray(data) ? data : []
  },
  // --- Security Center ---
  async securitySummary() {
    return request('/api/admin/security/summary', { auth: true })
  },
  // --- Error / System Log ---
  async listSyslogs({ level = '', limit = 20, offset = 0 } = {}) {
    const q = new URLSearchParams({ level, limit, offset })
    const data = await request(`/api/admin/syslogs?${q}`, { auth: true })
    return Array.isArray(data) ? data : []
  },
  async pruneSyslogs(days = 30) {
    return request(`/api/admin/syslogs?days=${days}`, { method: 'DELETE', auth: true })
  },
  // --- Notification Template & Log ---
  async listTemplates(activeOnly = false) {
    const data = await request(`/api/admin/notifications/templates${activeOnly ? '?active=1' : ''}`, { auth: true })
    return Array.isArray(data) ? data : []
  },
  async createTemplate(payload) {
    return request('/api/admin/notifications/templates', { method: 'POST', body: payload, auth: true })
  },
  async updateTemplate(code, payload) {
    return request(`/api/admin/notifications/templates/${code}`, { method: 'PUT', body: payload, auth: true })
  },
  async deleteTemplate(code) {
    return request(`/api/admin/notifications/templates/${code}`, { method: 'DELETE', auth: true })
  },
  async sendNotification(payload) {
    return request('/api/admin/notifications/send', { method: 'POST', body: payload, auth: true })
  },
  async listNotifLogs(limit = 20, offset = 0) {
    const data = await request(`/api/admin/notifications/logs?limit=${limit}&offset=${offset}`, { auth: true })
    return Array.isArray(data) ? data : []
  },
  async health() {
    return request('/healthz')
  },
  isLoggedIn() {
    return !!store.access
  },
}
