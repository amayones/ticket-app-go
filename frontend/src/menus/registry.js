// Registry menu modular: memindai otomatis semua mainpage di
// menus/<role>/<menu>/index.jsx. Tambah menu baru = tambah folder +
// index.jsx (lihat README "cara membuat menu baru") — tanpa sentuh file ini.
//
// Aturan folder role: shared = semua role, admin = ADMIN saja,
// user = USER saja, <nama lain> = role custom bernama sama (uppercase).
// Meta opsional di index.jsx: export const meta = { label, icon, order }.
const modules = import.meta.glob('./*/*/index.jsx', { eager: true })

function titleCase(key) {
  return key
    .split('-')
    .map((w) => (w ? w[0].toUpperCase() + w.slice(1) : w))
    .join(' ')
}

function loadMenus() {
  const menus = []
  for (const [path, mod] of Object.entries(modules)) {
    const m = path.match(/^\.\/([^/]+)\/([^/]+)\/index\.jsx$/)
    if (!m || typeof mod.default !== 'function') continue
    const [, role, key] = m
    const meta = mod.meta || {}
    menus.push({
      key,
      role: role.toUpperCase(),
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

// Menu untuk satu role: folder shared + folder role-nya.
export function menusForRole(role) {
  const r = (role || '').toUpperCase()
  return ALL_MENUS.filter((m) => m.role === 'SHARED' || m.role === r)
}
