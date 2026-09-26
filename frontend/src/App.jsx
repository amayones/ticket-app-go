import { useEffect, useMemo, useState } from 'react'
import { api, onSessionExpired } from './api/client.js'
import { Badge, Button, Icon, MissingMenu, Modal, ThemeToggle, ToastProvider, Tooltip, useToast } from './components'
import { groupMenus, menusForPermissions, missingMenus } from './menus/registry.js'
import { LoginForm } from './pages/Auth.jsx'

// Shell aplikasi: sidebar/topbar dibangun OTOMATIS dari registry menu
// (src/menus/registry.js) + entri CPMENU milik user (GET /api/users/me).
// Tambah menu = INSERT CPMENU (via UI Role) + tambah folder, tanpa sentuh file ini.
function sidebarPref() {
  try {
    return localStorage.getItem('go-core-sidebar') === 'collapsed'
  } catch {
    return false
  }
}

function loadExpanded(key) {
  try {
    return JSON.parse(localStorage.getItem(key) || '{}')
  } catch {
    return {}
  }
}

function saveExpanded(key, value) {
  try {
    localStorage.setItem(key, JSON.stringify(value))
  } catch {
    // abaikan (mode privat)
  }
}

// Tombol satu menu (leaf). Entri hilang (folder belum dibuat) tampil amber.
function MenuLeaf({ node, active, onSelect }) {
  return (
    <button
      type="button"
      onClick={() => onSelect(node.key)}
      className={`flex w-full items-center gap-2.5 rounded-xl px-3 py-2.5 transition-colors ${
        active
          ? 'bg-zinc-900 text-white shadow-sm dark:bg-zinc-100 dark:text-zinc-900'
          : node.missing
            ? 'text-amber-700 hover:bg-amber-50 dark:text-amber-300 dark:hover:bg-amber-950/40'
            : 'text-zinc-600 hover:bg-zinc-100 dark:text-zinc-300 dark:hover:bg-zinc-800'
      }`}
    >
      <Icon name={node.missing ? 'warning' : node.icon || 'list'} className="h-5 w-5 shrink-0" />
      <span className="grid whitespace-nowrap text-[13px] font-medium">
        <span className="min-w-0 overflow-hidden">
          {node.label}
          {node.missing && <span className="ml-1.5 font-mono text-[10px] opacity-70">404</span>}
        </span>
      </span>
    </button>
  )
}

// Grup parent bersarang (rekursif). Folder tanpa index.jsx = grup visual saja.
function ParentGroup({ name, pathKey, items, depth, activeKey, onSelect, openParents, onToggleParent }) {
  const open = openParents[pathKey] !== false
  const leaves = items.filter((it) => (it.parents || []).length === depth)
  const subs = new Map()
  for (const it of items) {
    if ((it.parents || []).length > depth) {
      const sub = it.parents[depth]
      if (!subs.has(sub)) subs.set(sub, [])
      subs.get(sub).push(it)
    }
  }
  return (
    <div className="flex flex-col gap-1">
      <button
        type="button"
        onClick={() => onToggleParent(pathKey)}
        aria-expanded={open}
        className="flex w-full items-center gap-2 rounded-lg px-3 py-1.5 text-xs font-bold tracking-wide text-zinc-500 transition-colors hover:bg-zinc-100 dark:text-zinc-400 dark:hover:bg-zinc-800"
      >
        <span className="flex h-4 w-4 shrink-0 items-center justify-center rounded border border-current text-[10px] leading-none">
          {open ? '−' : '+'}
        </span>
        <span className="truncate uppercase">{name}</span>
      </button>
      {open && (
        <div className="ml-3 flex flex-col gap-1 border-l border-zinc-200 pl-2 dark:border-zinc-700">
          {leaves.map((n) => (
            <MenuLeaf key={n.key} node={n} active={activeKey === n.key} onSelect={onSelect} />
          ))}
          {[...subs.entries()].map(([sub, subItems]) => (
            <ParentGroup
              key={sub}
              name={sub}
              pathKey={`${pathKey}/${sub}`}
              items={subItems}
              depth={depth + 1}
              activeKey={activeKey}
              onSelect={onSelect}
              openParents={openParents}
              onToggleParent={onToggleParent}
            />
          ))}
        </div>
      )}
    </div>
  )
}

// Section satu modul di sidebar (header + tombol +/−).
function ModuleGroup({ module, items, open, onToggle, activeKey, onSelect, openParents, onToggleParent }) {
  const leaves = items.filter((it) => (it.parents || []).length === 0)
  const subs = new Map()
  for (const it of items) {
    if ((it.parents || []).length > 0) {
      const sub = it.parents[0]
      if (!subs.has(sub)) subs.set(sub, [])
      subs.get(sub).push(it)
    }
  }
  return (
    <div className="flex flex-col gap-1">
      <button
        type="button"
        onClick={onToggle}
        aria-expanded={open}
        className="flex w-full items-center gap-2 rounded-lg px-2 py-1.5 text-xs font-bold tracking-wide text-zinc-500 transition-colors hover:bg-zinc-100 dark:text-zinc-400 dark:hover:bg-zinc-800"
      >
        <span className="flex h-4 w-4 shrink-0 items-center justify-center rounded border border-current text-[10px] leading-none">
          {open ? '−' : '+'}
        </span>
        <span className="truncate">MODULE {module}</span>
      </button>
      {open && (
        <div className="flex flex-col gap-1">
          {leaves.map((n) => (
            <MenuLeaf key={n.key} node={n} active={activeKey === n.key} onSelect={onSelect} />
          ))}
          {[...subs.entries()].map(([sub, subItems]) => (
            <ParentGroup
              key={sub}
              name={sub}
              pathKey={`${module}/${sub}`}
              items={subItems}
              depth={1}
              activeKey={activeKey}
              onSelect={onSelect}
              openParents={openParents}
              onToggleParent={onToggleParent}
            />
          ))}
        </div>
      )}
    </div>
  )
}

function Shell() {
  const toast = useToast()
  const [view, setView] = useState('')
  const [loggedIn, setLoggedIn] = useState(api.isLoggedIn())
  const [collapsed, setCollapsed] = useState(sidebarPref)
  // Popup login ulang saat sesi habis: tampil di atas halaman terakhir,
  // tanpa pindah ke halaman login. sessionTick memaksa Active remount
  // agar data dimuat ulang dengan token baru setelah login sukses.
  const [sessionExpired, setSessionExpired] = useState(false)
  const [sessionTick, setSessionTick] = useState(0)
  // Cache identitas terakhir agar sidebar/halaman tidak lompat saat token
  // sudah dibersihkan (api.currentUser() -> null) tapi popup belum ditutup.
  const [lastMe, setLastMe] = useState(() => api.currentUser())
  // Permission codes milik user (diambil via GET /api/users/me saat login).
  // Digunakan untuk filter menu di sidebar.
  const [permissions, setPermissions] = useState([])
  // Entri CPMENU milik user (dari /me.menus) untuk sidebar + placeholder 404.
  const [myMenus, setMyMenus] = useState([])
  // Section modul + parent yang dibuka (persist localStorage).
  const [openModules, setOpenModules] = useState(() => loadExpanded('go-core-modules'))
  const [openParents, setOpenParents] = useState(() => loadExpanded('go-core-parents'))
  const me = api.currentUser() || (sessionExpired ? lastMe : null)
  const isAdmin = me?.role === 'ADMIN'
  const visibleMenus = menusForPermissions(permissions)
  const missing = useMemo(() => missingMenus(myMenus), [myMenus])
  const allNodes = useMemo(() => [...visibleMenus, ...missing], [visibleMenus, missing])
  const groups = useMemo(() => groupMenus(allNodes), [allNodes])
  const activeNode = allNodes.find((m) => m.key === view) || allNodes[0] || null
  const Active = activeNode && !activeNode.missing ? activeNode.Component : null

  function toggleOpen(setter, storeKey, key) {
    setter((prev) => {
      // Default terbuka; hanya kunci yang tertutup yang disimpan.
      const next = { ...prev }
      if (next[key] === false) delete next[key]
      else next[key] = false
      saveExpanded(storeKey, next)
      return next
    })
  }

  const isOpen = (map, key) => map[key] !== false

  useEffect(() => {
    if (!loggedIn) return
    return onSessionExpired(() => {
      setSessionExpired(true)
      toast.warning('Sesi Anda telah berakhir. Silakan login kembali.', { title: 'Sesi habis' })
    })
  }, [loggedIn, toast])

  useEffect(() => {
    if (!loggedIn || permissions.length > 0) return
    api.getMe()
      .then((data) => {
        setPermissions(data.permissions || [])
        setMyMenus(data.menus || [])
        setLastMe(data)
      })
      .catch(() => {
        setPermissions([])
        setMyMenus([])
      })
  }, [loggedIn, permissions.length])

  async function logout() {
    await api.logout()
    setSessionExpired(false)
    setLastMe(null)
    setPermissions([])
    setMyMenus([])
    setLoggedIn(false)
    setView('login')
    toast.info('Anda telah keluar. Sampai jumpa!')
  }

  async function loadPermissions() {
    try {
      const data = await api.getMe()
      setPermissions(data.permissions || [])
      setMyMenus(data.menus || [])
      setLastMe(data)
    } catch {
      setPermissions([])
      setMyMenus([])
      setLastMe(api.currentUser())
    }
  }

  function handleAuth() {
    setLoggedIn(true)
    setView('')
    setMyMenus([])
    loadPermissions()
    toast.success('Selamat datang kembali!')
  }

  // Login ulang dari popup sesi-habis: tetap di halaman terakhir (view
  // tidak diubah), cukup tutup popup + muat ulang konten dengan token baru.
  function handleRelogin() {
    setSessionExpired(false)
    setLoggedIn(true)
    setSessionTick((t) => t + 1)
    loadPermissions()
    toast.success('Sesi dipulihkan. Selamat melanjutkan!')
  }



  function handleAccountDeleted() {
    setSessionExpired(false)
    setLastMe(null)
    setPermissions([])
    setMyMenus([])
    setLoggedIn(false)
    setView('login')
    toast.warning('Akun Anda telah dihapus.', { title: 'Akun dihapus' })
  }

  function toggleSidebar() {
    setCollapsed((c) => {
      try {
        localStorage.setItem('go-core-sidebar', c ? 'open' : 'collapsed')
      } catch {
        // abaikan (mode privat)
      }
      return !c
    })
  }

  // Halaman login fokus: tanpa navbar, kartu di tengah layar.
  if (!loggedIn) {
    return (
      <div className="anim-boot flex min-h-screen w-full items-center justify-center bg-zinc-100 px-4 py-10 dark:bg-zinc-950">
        <div className="w-full max-w-sm">
          <LoginForm onDone={handleAuth} />
          <p className="mt-6 text-center text-xs text-zinc-400 dark:text-zinc-500">
            Go Core — hubungi admin bila belum punya akun
          </p>
        </div>
      </div>
    )
  }

  return (
    <div className="anim-boot flex min-h-screen w-full bg-zinc-50 text-zinc-900 dark:bg-zinc-950 dark:text-zinc-100">
      {/* Sidebar (desktop) — bisa dilipat via tombol chevron */}
      <aside
        className={`sticky top-0 z-30 hidden h-screen shrink-0 flex-col border-r border-zinc-200 bg-white p-3 duration-200 ease-[cubic-bezier(0.4,0,0.2,1)] transition-[width] md:flex dark:border-zinc-800 dark:bg-zinc-900 ${
          collapsed ? 'w-[84px]' : 'w-60'
        }`}
      >
        {/* Logo sejajar ikon menu; tombol lipat menonjol di kanan logo
            (tengah-tengah tinggi baris logo). */}
        <div className="pb-4">
          <div className={`relative flex items-center ${collapsed ? 'justify-center px-0' : 'gap-2 px-1'}`}>
            <span className="flex h-8 w-8 shrink-0 items-center justify-center rounded-lg bg-zinc-900 text-sm font-bold text-white dark:bg-zinc-100 dark:text-zinc-900">
              G
            </span>
            <span
              className={`grid whitespace-nowrap text-sm font-bold tracking-tight transition-all duration-200 ease-[cubic-bezier(0.4,0,0.2,1)] ${
                collapsed ? 'grid-cols-[0fr] opacity-0' : 'grid-cols-[1fr] opacity-100'
              }`}
            >
              <span className="min-w-0 overflow-hidden">Go Core</span>
            </span>
            <Tooltip
              label={collapsed ? 'Buka sidebar' : 'Tutup sidebar'}
              position="right"
              className="absolute -right-5 top-1/2 -translate-y-1/2"
            >
              <button
                type="button"
                onClick={toggleSidebar}
                aria-label={collapsed ? 'Buka sidebar' : 'Tutup sidebar'}
                className="shrink-0 rounded-full border border-zinc-200 bg-white p-1.5 text-zinc-500 shadow-md transition-all duration-200 hover:bg-zinc-50 hover:text-zinc-800 hover:shadow-lg active:scale-90 dark:border-zinc-700 dark:bg-zinc-800 dark:text-zinc-400 dark:hover:bg-zinc-700 dark:hover:text-zinc-100"
              >
                <Icon
                  name="chevronLeft"
                  className={`h-4 w-4 transition-transform duration-200 ease-[cubic-bezier(0.4,0,0.2,1)] ${collapsed ? 'rotate-180' : ''}`}
                />
              </button>
            </Tooltip>
          </div>
        </div>
        <nav className="flex flex-1 flex-col gap-1 overflow-y-auto">
          {collapsed ? (
            // Mode lipat: daftar ikon flat (grup modul tidak muat di ruang sempit).
            allNodes.map((n) => {
              const btn = (
                <button
                  key={n.key}
                  type="button"
                  onClick={() => setView(n.key)}
                  className={`flex w-full items-center justify-center gap-0 rounded-xl px-3 py-2.5 transition-colors ${
                    activeNode?.key === n.key
                      ? 'bg-zinc-900 text-white shadow-sm dark:bg-zinc-100 dark:text-zinc-900'
                      : 'text-zinc-600 hover:bg-zinc-100 dark:text-zinc-300 dark:hover:bg-zinc-800'
                  }`}
                >
                  <Icon name={n.missing ? 'warning' : n.icon || 'list'} className="h-5 w-5 shrink-0" />
                </button>
              )
              return (
                <Tooltip key={n.key} label={n.missing ? `${n.label} (belum dibuat)` : n.label} position="right">
                  {btn}
                </Tooltip>
              )
            })
          ) : (
            groups.map((g) => (
              <ModuleGroup
                key={g.module}
                module={g.module}
                items={g.items}
                open={isOpen(openModules, g.module)}
                onToggle={() => toggleOpen(setOpenModules, 'go-core-modules', g.module)}
                activeKey={activeNode?.key}
                onSelect={setView}
                openParents={openParents}
                onToggleParent={(key) => toggleOpen(setOpenParents, 'go-core-parents', key)}
              />
            ))
          )}
        </nav>
        <div className="border-t border-zinc-200 pt-3 dark:border-zinc-800">
          <div
            className={`grid transition-all duration-200 ease-[cubic-bezier(0.4,0,0.2,1)] ${
              collapsed ? 'grid-rows-[0fr] opacity-0' : 'mb-2 grid-rows-[1fr] opacity-100'
            }`}
          >
            <div className="min-h-0 overflow-hidden">
              <div className="flex items-center gap-2 px-2">
                <span className="min-w-0 flex-1 truncate text-xs text-zinc-500">
                  <span className="block truncate font-semibold text-zinc-700 dark:text-zinc-200">@{me?.username}</span>
                  <span className="font-mono">{me?.code}</span>
                </span>
                <Badge tone={isAdmin ? 'danger' : 'brand'}>{me?.role}</Badge>
              </div>
            </div>
          </div>
          <div className={`flex items-center gap-1.5 ${collapsed ? 'flex-col' : ''}`}>
            {collapsed ? (
              <Tooltip label="Logout" position="right">
                <button
                  type="button"
                  onClick={logout}
                  aria-label="Logout"
                  className="rounded-lg p-2 text-zinc-500 transition-colors hover:bg-zinc-100 hover:text-zinc-800 dark:text-zinc-400 dark:hover:bg-zinc-800 dark:hover:text-zinc-100"
                >
                  <Icon name="logout" className="h-5 w-5" />
                </button>
              </Tooltip>
            ) : (
              <Button variant="secondary" size="sm" fullWidth onClick={logout}>
                <Icon name="logout" className="h-4 w-4" />
                Logout
              </Button>
            )}
            <Tooltip label="Ganti tema" position="right">
              <ThemeToggle />
            </Tooltip>
          </div>
        </div>
      </aside>

      {/* Mobile topbar */}
      <div className="flex min-w-0 flex-1 flex-col">
        <header className="sticky top-0 z-40 border-b border-zinc-200 bg-white/90 backdrop-blur md:hidden dark:border-zinc-800 dark:bg-zinc-950/90">
          <div className="flex items-center gap-2 px-3 py-2.5">
            <span className="flex h-7 w-7 shrink-0 items-center justify-center rounded-lg bg-gradient-to-br from-violet-600 to-fuchsia-600 text-xs text-white">
              G
            </span>
            <nav className="flex flex-1 items-center gap-1 overflow-x-auto">
              {allNodes.map((n) => (
                <button
                  key={n.key}
                  type="button"
                  onClick={() => setView(n.key)}
                  className={`whitespace-nowrap rounded-lg px-2.5 py-1.5 text-xs font-semibold transition ${
                    activeNode?.key === n.key
                      ? 'bg-violet-600 text-white'
                      : n.missing
                        ? 'text-amber-600 dark:text-amber-300'
                        : 'text-zinc-600 dark:text-zinc-300'
                  }`}
                >
                  {n.label}
                  {n.missing && ' 404'}
                </button>
              ))}
            </nav>
            <ThemeToggle className="shrink-0" />
            <button
              type="button"
              onClick={logout}
              aria-label="Logout"
              className="shrink-0 rounded-lg p-1.5 text-zinc-500 hover:bg-zinc-100 dark:hover:bg-zinc-800"
            >
              <Icon name="logout" className="h-5 w-5" />
            </button>
          </div>
        </header>

        <main className="mx-auto w-full max-w-5xl flex-1 px-3 py-4 sm:px-4 sm:py-6">
          {/* key memicu animasi masuk yang halus tiap ganti menu;
              sessionTick memaksa muat ulang setelah login dari popup sesi-habis */}
          <div key={`${activeNode?.key || 'empty'}-${sessionTick}`} className="anim-page-in">
            {activeNode?.missing ? (
              <MissingMenu entry={activeNode} />
            ) : Active ? (
              <Active onNavigate={setView} onAccountDeleted={handleAccountDeleted} />
            ) : (
              <div className="rounded-2xl border border-zinc-200 bg-white p-8 text-center dark:border-zinc-800 dark:bg-zinc-900">
                <p className="text-sm font-semibold text-zinc-900 dark:text-zinc-50">Tidak ada menu tersedia</p>
                <p className="mt-1 text-xs text-zinc-500">Akun Anda belum diberi akses menu — hubungi admin.</p>
              </div>
            )}
          </div>
        </main>
        <footer className="border-t border-zinc-200 py-3 text-center text-xs text-zinc-400 dark:border-zinc-800 dark:text-zinc-500">
          Go Core — Go + React dalam satu binary
        </footer>
      </div>

      <Modal
        open={sessionExpired}
        title="Sesi habis — login lagi"
        size="sm"
        showClose={false}
        closeOnBackdrop={false}
        bodyClassName="h-[320px] overflow-hidden px-5 py-4 text-sm text-zinc-600 dark:text-zinc-300"
      >
        <div className="flex h-full flex-col gap-3">
          <p className="text-sm text-zinc-600 dark:text-zinc-300">
            Sesi Anda telah berakhir. Login kembali untuk melanjutkan di halaman ini.
          </p>
          <LoginForm bare onDone={handleRelogin} />
        </div>
      </Modal>
    </div>
  )
}

export default function App() {
  return (
    <ToastProvider>
      <Shell />
    </ToastProvider>
  )
}
