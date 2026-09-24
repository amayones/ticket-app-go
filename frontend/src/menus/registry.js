// Registry menu modular: folder pertama adalah kategori/module, folder kedua menu.
// Module hanya untuk pengelompokan, bukan pembatasan role. Akses menu sepenuhnya
// ditentukan oleh permission MENU_<NAMA_MENU>.
//
// | Folder             | Module   |
// |--------------------|----------|
// | menus/account/...  | ACCOUNT  |
// | menus/system/...   | SYSTEM   |
//
// Tambah menu = tambah folder <module>/<menu>/index.jsx + permission menu.
const modules = import.meta.glob(['./account/*/index.jsx', './system/*/index.jsx'], { eager: true })

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
    const m = path.match(/^\.\/(account|system)\/([^/]+)\/index\.jsx$/)
    if (!m || typeof mod.default !== 'function') continue
    const [, moduleName, key] = m
    const meta = mod.meta || {}
    menus.push({
      key,
      module: moduleName.toUpperCase(),
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

export function menusForPermissions(permissions = []) {
  const permSet = new Set(permissions)
  return ALL_MENUS.filter((menu) => permSet.has(permissionFor(menu.key)))
}
