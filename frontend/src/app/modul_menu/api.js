// Fungsi menu Modul & Menu (master CPMODULE + registry CPMENU).
// Urut kerja: buat modul dulu, lalu menu di dalamnya, lalu centang role
// di matriks halaman Role. Tanpa auto-grant ke role mana pun.
import { apiRequest as request } from '../../api/client.js'

export async function listMenus() {
  const data = await request('/api/admin/menus', { auth: true })
  return Array.isArray(data) ? data : []
}

export async function createMenu(payload) {
  return request('/api/admin/menus', { method: 'POST', body: payload, auth: true })
}

// Ubah label/urutan/modul/parent. Kode permission tidak bisa diubah
// (jadi acuan folder frontend + grant), jadi tidak ada field code di payload.
export async function updateMenu(code, payload) {
  return request(`/api/admin/menus/${code}`, { method: 'PUT', body: payload, auth: true })
}

export async function deleteMenu(code) {
  return request(`/api/admin/menus/${code}`, { method: 'DELETE', auth: true })
}

export async function listModules() {
  const data = await request('/api/admin/modules', { auth: true })
  return Array.isArray(data) ? data : []
}

export async function createModule(payload) {
  return request('/api/admin/modules', { method: 'POST', body: payload, auth: true })
}

export async function deleteModule(code) {
  return request(`/api/admin/modules/${code}`, { method: 'DELETE', auth: true })
}

// ===== Padanan store proxy langit_v2 (method read_data/process_*) =====
// read_data: dipakai store/load (modul + menu).
export async function read_data() {
  const [menus, modules] = await Promise.all([listMenus(), listModules()])
  return {
    menus: Array.isArray(menus) ? menus : [],
    modules: Array.isArray(modules) ? modules : [],
  }
}

// process_create/update/delete menu: dipakai form menu (handler_btsave/btdelete).
export async function process_create_menu(dtval) {
  return createMenu(dtval)
}

export async function process_update_menu(code, dtval) {
  return updateMenu(code, dtval)
}

export async function process_delete_menu(code) {
  return deleteMenu(code)
}

// process_create/delete modul: dipakai form modul.
export async function process_create_module(dtval) {
  return createModule(dtval)
}

export async function process_delete_module(code) {
  return deleteModule(code)
}
