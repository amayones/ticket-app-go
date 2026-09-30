// Fungsi menu Home (discovery publik, tanpa login).
import { apiRequest as request } from '../../api/client.js'

function qs(params = {}) {
  const q = new URLSearchParams()
  for (const [k, v] of Object.entries(params)) {
    if (v !== undefined && v !== null && String(v) !== '') q.append(k, String(v))
  }
  const s = q.toString()
  return s ? `?${s}` : ''
}

export async function discoveryHome({ city } = {}, signal) {
  return request(`/api/discovery/home${qs({ city })}`, { signal })
}

export async function discoveryCities(signal) {
  const cities = await request('/api/discovery/cities', { signal })
  return Array.isArray(cities) ? cities : []
}

// Impression/klik iklan: best-effort (kegagalan tidak mengganggu halaman).
export async function recordImpression(code) {
  try {
    await request(`/api/discovery/ads/${code}/impression`, { method: 'POST' })
  } catch {
    // abaikan (anti-gaming boleh menolak; jaringan boleh gagal)
  }
}

export async function recordClick(code) {
  try {
    await request(`/api/discovery/ads/${code}/click`, { method: 'POST' })
  } catch {
    // abaikan (lihat recordImpression)
  }
}

// ===== Padanan store proxy langit_v2 (method read_data) =====
// read_data: dipakai load Home (ringkasan + daftar kota pemilih).
export async function read_data({ city } = {}, signal) {
  const [home, cities] = await Promise.all([discoveryHome({ city }, signal), discoveryCities(signal)])
  return { home: home || null, cities }
}
