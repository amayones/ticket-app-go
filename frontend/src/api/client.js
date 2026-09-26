// Core API client: base URL from VITE_API_URL, JSON envelope,
// auto refresh-token rotation on 401, no-store for auth calls.
//
// Aturan modular: file ini HANYA berisi infrastruktur (request) + fungsi
// inti (auth, sesi JWT, health). Fungsi tiap menu tinggal di
// menus/<module>/<menu>/api.js dan memakai apiRequest dari sini.
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

// Event global saat sesi benar-benar habis (refresh gagal / tidak ada).
// App.jsx mendengarkan ini untuk menampilkan popup login ulang di tempat,
// tanpa pindah halaman. Guard agar parallel 401 tidak spam event.
const SESSION_EXPIRED_EVENT = 'go-core:session-expired'
let expiredNotified = false

function notifySessionExpired() {
  if (expiredNotified) return
  expiredNotified = true
  try {
    window.dispatchEvent(new CustomEvent(SESSION_EXPIRED_EVENT))
  } catch {
    // abaikan (non-browser / mode privat)
  }
}

export function onSessionExpired(listener) {
  window.addEventListener(SESSION_EXPIRED_EVENT, listener)
  return () => window.removeEventListener(SESSION_EXPIRED_EVENT, listener)
}

function failSession() {
  store.clear()
  notifySessionExpired()
}

async function request(path, { method = 'GET', body, auth = false, retry = true } = {}) {
  const headers = { 'Content-Type': 'application/json' }
  if (auth && store.access) headers.Authorization = `Bearer ${store.access}`
  const res = await fetch(`${BASE}${path}`, {
    method,
    headers,
    body: body ? JSON.stringify(body) : undefined,
  })
  if (res.status === 401 && auth) {
    // Coba rotasi refresh-token sekali; kalau gagal / tidak ada refresh,
    // sesi dianggap habis -> bersihkan + beri tahu UI detik itu juga.
    if (retry && store.refresh) {
      const ok = await tryRefresh()
      if (ok) return request(path, { method, body, auth, retry: false })
    }
    failSession()
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

// Low-level request untuk dipakai api.js tiap menu (menus/<module>/<menu>/api.js).
export const apiRequest = request

export const api = {
  // Profil user yang sedang login, dibaca dari klaim JWT (tanpa request).
  currentUser() {
    if (!store.access) return null
    const claims = parseJwt(store.access)
    if (!claims || !claims.user_code) return null
    return { code: claims.user_code, username: claims.username || '', role: claims.role || '' }
  },
  // Ambil user + permission-nya dari backend (GET /api/users/me).
  // Dipakai saat login untuk mengetahui menu mana yang boleh ditampilkan.
  async getMe() {
    return request('/api/users/me', { auth: true })
  },
  async login(username, password) {
    const data = await request('/api/login', { method: 'POST', body: { username, password } })
    store.set(data.access_token, data.refresh_token)
    expiredNotified = false
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
  async health() {
    return request('/healthz')
  },
  isLoggedIn() {
    return !!store.access
  },
}
