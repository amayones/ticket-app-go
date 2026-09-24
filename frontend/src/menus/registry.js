// Registry menu 2 tipe: `admin` khusus ADMIN, `user` untuk SEMUA role.
// Tambah menu = tambah folder + index.jsx, tanpa sentuh file ini.
//
// | Folder | Dilihat oleh |
// |--------|--------------|
// | `menus/admin/<menu>/` | role `ADMIN` saja |
// | `menus/user/<menu>/`  | Semua role yang memiliki permission menunya |
//
// Meta opsional di index.jsx: export const meta = { label, icon, order }.
const modules = import.meta.glob(['./admin/*/index.jsx', './user/*/index.jsx'], { eager: true })

function permissionFor(key) {
  return `MENU_${key.replaceAll('-', '_').toUpperCase()}`
}

function titleCase(key) {
  return key
    .split('-')
    .map((w) => (w ? w[0].toUpperCase() + w.slice(1) : w))
    .join(' ')
}

function loadMenus() {
  const menus = []
  for (const [path, mod] of Object.entries(modules)) {
    const m = path.match(/^\.\/(admin|user)\/([^/]+)\/index\.jsx$/)
    if (!m || typeof mod.default !== 'function') continue
    const [, scope, key] = m
    const meta = mod.meta || {}
    menus.push({
      key,
      scope: scope.toUpperCase() === 'ADMIN' ? 'ADMIN' : 'ALL',
      label: meta.label || titleCase(key),
      icon: meta.icon || 'list',
      order: meta.order ?? 99,
      Component: mod.default,
    })
  }
  menus.sort((a, b) => a.order - b.order || a.label.localeCompare(b.label))
  return menus
}

const ALL_MENUS = loadMenus()

export function allMenus() {
  return ALL_MENUS
}

// Menu untuk satu role + permission:
// - ADMIN: melihat semua menu.
// - Role lain: hanya menu yang permission-nya diberikan.
export function menusForRole(role, permissions = []) {
  if ((role || '').toUpperCase() === 'ADMIN') return ALL_MENUS

  const permSet = new Set(permissions)
  return ALL_MENUS.filter((m) => {
    if (m.scope === 'ADMIN') return false
    const needed = permissionFor(m.key)
    return permSet.has(needed)
  })
}