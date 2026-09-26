// Registry menu datar: satu folder di menus/ = satu menu CHILD.
// Folder = MCONTROL (snake_case CPMENU.MCONTROL, mis. users).
// Tanpa folder modul perantara. Section = CPMENU.MODULE (DB),
// header = CPMENU PARENT (MENU_KIND=PARENT, tanpa folder),
// item = CPMENU CHILD (MCONTROL wajib).
const modules = import.meta.glob('./*/index.jsx', { eager: true })

function titleCase(key) {
  return String(key || '')
    .split('_')
    .map((w) => (w ? w[0].toUpperCase() + w.slice(1) : w))
    .join(' ')
}

function loadMenus() {
  const menus = []
  for (const [path, mod] of Object.entries(modules)) {
    const m = path.match(/^\.\/([^/]+)\/index\.jsx$/)
    if (!m || typeof mod.default !== 'function') continue
    const [, mcontrol] = m
    const meta = mod.meta || {}
    menus.push({
      key: mcontrol,
      mcontrol,
      label: meta.label || titleCase(mcontrol),
      icon: meta.icon || 'list',
      order: meta.order ?? 99,
      Component: mod.default,
    })
  }
  menus.sort((a, b) => a.order - b.order || a.label.localeCompare(b.label))
  return menus
}

const ALL_MENUS = loadMenus()

function localByMcontrol() {
  const map = new Map()
  for (const m of ALL_MENUS) map.set(m.mcontrol, m)
  return map
}

// buildSidebar: entri DB (myMenus) + folder lokal -> 3 level.
// myMenus = GET /api/users/me .menus (CHILD ter-grant + ancestor PARENT).
// Return: [{ module, moduleLabel, moduleSort, roots, childrenOf }].
// Node PARENT = header (tanpa Component). Node CHILD = item (key=mcontrol).
export function buildSidebar(myMenus = []) {
  const local = localByMcontrol()
  const entries = Array.isArray(myMenus) ? myMenus : []
  const byCode = new Map()
  for (const e of entries) if (e && e.code) byCode.set(e.code, e)
  const childrenCodesOf = new Map()
  for (const e of entries) {
    if (e && e.parent_code) {
      if (!childrenCodesOf.has(e.parent_code)) childrenCodesOf.set(e.parent_code, [])
      childrenCodesOf.get(e.parent_code).push(e.code)
    }
  }
  const memo = new Map()
  function isVisible(code) {
    if (memo.has(code)) return memo.get(code)
    memo.set(code, false)
    const e = byCode.get(code)
    if (!e) return false
    if (e.kind !== 'PARENT') { memo.set(code, true); return true }
    const kids = childrenCodesOf.get(code) || []
    const v = kids.some((k) => isVisible(k))
    memo.set(code, v)
    return v
  }
  for (const e of entries) if (e && e.code) isVisible(e.code)
  const modMeta = new Map()
  for (const e of entries) {
    const mod = String(e.module || 'SYSTEM').toUpperCase()
    if (!modMeta.has(mod)) modMeta.set(mod, {
      module: mod,
      moduleLabel: e.module_label || titleCase(mod.toLowerCase()),
      moduleSort: e.module_sort ?? 99,
    })
  }
  const groups = new Map()
  function groupFor(mod) {
    if (!groups.has(mod)) {
      const meta = modMeta.get(mod) || { module: mod, moduleLabel: titleCase(mod.toLowerCase()), moduleSort: 99 }
      groups.set(mod, { ...meta, nodes: [] })
    }
    return groups.get(mod)
  }
  for (const e of entries) {
    if (!e || !e.code || !isVisible(e.code)) continue
    const mod = String(e.module || 'SYSTEM').toUpperCase()
    const g = groupFor(mod)
    if (e.kind === 'PARENT') {
      g.nodes.push({ key: e.code, code: e.code, module: mod, label: e.label || e.code, order: e.sort_order ?? 99, parent: e.parent_code || '', kind: 'PARENT', header: true })
    } else {
      const mc = e.mcontrol || ''
      const lm = local.get(mc)
      g.nodes.push({ key: mc || e.code, code: e.code, module: mod, label: e.label || (lm && lm.label) || titleCase(mc || e.code), icon: (lm && lm.icon) || 'list', order: e.sort_order ?? lm?.order ?? 99, parent: e.parent_code || '', kind: e.kind || 'CHILD', mcontrol: mc, missing: !lm, Component: lm ? lm.Component : undefined })
    }
  }
  return [...groups.values()]
    .sort((a, b) => a.moduleSort - b.moduleSort || a.module.localeCompare(b.module))
    .map((g) => {
      const codeToKey = new Map(g.nodes.map((n) => [n.code, n.key]))
      const byKey = new Set(g.nodes.map((n) => n.key))
      const roots = []
      const childrenOf = new Map()
      const sorted = [...g.nodes].sort((a, b) => ((a.kind === 'PARENT' ? -1 : 0) - (b.kind === 'PARENT' ? -1 : 0)) || a.order - b.order || a.label.localeCompare(b.label))
      for (const n of sorted) {
        const pk = n.parent ? codeToKey.get(n.parent) : undefined
        if (!n.parent || !pk || !byKey.has(pk)) roots.push(n)
        else { if (!childrenOf.has(pk)) childrenOf.set(pk, []); childrenOf.get(pk).push(n) }
      }
      return { module: g.module, moduleLabel: g.moduleLabel, moduleSort: g.moduleSort, roots, childrenOf }
    })
}

// suggestPath: lokasi folder yang harus dibuat untuk menu yang belum ada
// filenya (dipakai halaman 404 MissingMenu).
export function suggestPath(entry) {
  return `frontend/src/app/${entry.mcontrol || entry.key}/`
}
