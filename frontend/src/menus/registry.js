// Registry menu modular + bersarang: folder pertama adalah modul (MCONTROL,
// UPPERCASE, mis. SYSTEM/REPORT), segmen terakhir adalah key menu, segmen
// tengah adalah parent grup visual (boleh tanpa index.jsx).
// Modul hanya untuk pengelompokan, bukan pembatasan role. Akses menu
// sepenuhnya ditentukan oleh permission MENU_<NAMA_MENU> + baris CPMENU.
//
// | Folder                          | Modul  | Key      | Parent     |
// |---------------------------------|--------|----------|------------|
// | menus/SYSTEM/users/index.jsx    | SYSTEM | users    | -          |
// | menus/REPORT/keu/laporan/idx   | REPORT | laporan  | [keu]      |
//
// Tambah menu = INSERT CPMENU/CPPERMISSION (via UI Role) + folder
// menus/<MCONTROL>/[parent/]<key>/index.jsx. Permission tanpa folder cocok
// Terdeteksi missingMenus() dan dirender sebagai halaman 404 pemandu.
const modules = import.meta.glob('./*/**/index.jsx', { eager: true })

function permissionFor(key) {
  return `MENU_${key.replaceAll('-', '_').toUpperCase()}`
}

export function keyFromCode(code) {
  return String(code || '')
    .replace(/^MENU_/, '')
    .toLowerCase()
    .replaceAll('_', '-')
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
    const m = path.match(/^\.\/([^/]+)\/(.+)\/index\.jsx$/)
    if (!m || typeof mod.default !== 'function') continue
    const [, moduleName, rest] = m
    const segs = rest.split('/')
    const key = segs[segs.length - 1]
    const parents = segs.slice(0, -1)
    const meta = mod.meta || {}
    menus.push({
      key,
      module: moduleName.toUpperCase(),
      parents,
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

// missingMenus: entri CPMENU milik user yang foldernya belum dibuat.
// myMenus = data.menus dari GET /api/users/me (atau /api/menus/mine).
export function missingMenus(myMenus = []) {
  const known = new Set(ALL_MENUS.map((m) => permissionFor(m.key)))
  return (Array.isArray(myMenus) ? myMenus : [])
    .filter((e) => e && e.code && !known.has(e.code))
    .map((e) => {
      const key = keyFromCode(e.code)
      const module = String(e.module || 'SYSTEM').toUpperCase()
      return {
        key,
        code: e.code,
        module,
        parents: [],
        label: e.label || titleCase(key),
        order: e.sort_order ?? 99,
        missing: true,
      }
    })
    .sort((a, b) => a.order - b.order || a.label.localeCompare(b.label))
}

// suggestPath: path folder yang harus dibuat untuk entri yang hilang.
export function suggestPath(entry) {
  const parts = [entry.module, ...(entry.parents || []), entry.key].filter(Boolean)
  return `frontend/src/menus/${parts.join('/')}/`
}

// groupMenus: kelompokkan item per modul untuk sidebar (urut by order terkecil).
export function groupMenus(items = []) {
  const map = new Map()
  for (const it of items) {
    const mod = (it.module || 'SYSTEM').toUpperCase()
    if (!map.has(mod)) map.set(mod, [])
    map.get(mod).push(it)
  }
  return [...map.entries()]
    .map(([module, list]) => ({ module, items: list }))
    .sort((a, b) => minOrder(a.items) - minOrder(b.items) || a.module.localeCompare(b.module))
}

function minOrder(items) {
  return items.reduce((m, it) => Math.min(m, it.order ?? 99), 99)
}
