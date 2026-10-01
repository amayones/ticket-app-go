import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { api, onSessionExpired } from './api/client.js'
import { AppLogo, Badge, Button, Icon, MissingMenu, Modal, ThemeToggle, ToastProvider, Tooltip, useToast } from './components'
import { APP_NAME } from './brand.js'
import { buildSidebar } from './app/registry.js'
import { LoginForm } from './pages/Auth.jsx'
import HomePage from './pages/HomePage.jsx'

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
  }
}

function MenuNode({ node, childrenOf, activeKey, onSelect, openParents, onToggleParent, depth = 0 }) {
  const children = childrenOf.get(node.key) || []
  if (node.header) {
    const open = openParents[node.key] !== false
    return (
      <div className="flex flex-col gap-1">
        <button
          type="button"
          onClick={() => onToggleParent(node.key)}
          aria-expanded={open}
          aria-label={`${open ? 'Tutup' : 'Buka'} submenu ${node.label}`}
          className="flex min-w-0 flex-1 items-center gap-2.5 rounded-xl border border-violet-200 bg-violet-50 py-2 pl-3 pr-1 text-violet-800 transition-colors hover:bg-violet-100 dark:border-violet-900/80 dark:bg-violet-950/50 dark:text-violet-100 dark:hover:bg-violet-900/60"
        >
          <Icon name="folder" className="h-[18px] w-[18px] shrink-0" />
          <span className="grid min-w-0 flex-1 whitespace-nowrap text-left text-[13px] font-semibold">
            <span className="min-w-0 overflow-hidden">{node.label}</span>
          </span>
          <Icon name="chevronRight" className={`h-4 w-4 shrink-0 text-zinc-400 transition-transform duration-200 ${open ? 'rotate-90' : ''}`} />
        </button>
        {open && children.length > 0 && (
          <div className="ml-4 flex flex-col gap-1 border-l-2 border-violet-200 pl-1 dark:border-violet-900/70">
            {children.map((child) => (
              <MenuNode key={child.key} node={child} childrenOf={childrenOf} activeKey={activeKey} onSelect={onSelect} openParents={openParents} onToggleParent={onToggleParent} depth={depth + 1} />
            ))}
          </div>
        )}
      </div>
    )
  }
  const expandable = children.length > 0
  const open = openParents[node.key] !== false
  return (
    <div className="flex flex-col gap-1">
      <div className="flex items-center gap-0.5">
        <button
          type="button"
          onClick={() => onSelect(node.key)}
          className={`flex min-w-0 flex-1 items-center gap-2.5 rounded-xl py-2 pr-1 transition-colors ${depth > 0 ? 'pl-5' : 'pl-3'} ${activeKey === node.key ? 'bg-zinc-900 text-white shadow-sm dark:bg-zinc-100 dark:text-zinc-900' : node.missing ? 'text-amber-700 hover:bg-amber-50 dark:text-amber-300 dark:hover:bg-amber-950/40' : 'text-zinc-600 hover:bg-zinc-100 dark:text-zinc-300 dark:hover:bg-zinc-800'}`}
        >
          <Icon name={node.missing ? 'warning' : node.icon || 'list'} className="h-4 w-4 shrink-0" />
          <span className="grid min-w-0 whitespace-nowrap text-xs font-medium">
            <span className="min-w-0 overflow-hidden">
              {node.label}
              {node.missing && <span className="ml-1.5 font-mono text-[10px] opacity-70">404</span>}
            </span>
          </span>
        </button>
        {expandable && (
          <button type="button" onClick={() => onToggleParent(node.key)} aria-expanded={open} aria-label={`${open ? 'Tutup' : 'Buka'} anak menu ${node.label}`} className="shrink-0 rounded-lg p-1 text-zinc-400 transition-colors hover:bg-zinc-100 hover:text-zinc-700 dark:hover:bg-zinc-800">
            <Icon name="chevronRight" className={`h-4 w-4 transition-transform duration-200 ${open ? 'rotate-90' : ''}`} />
          </button>
        )}
      </div>
      {expandable && open && children.length > 0 && (
        <div className="ml-4 flex flex-col gap-1 border-l border-zinc-200 pl-1 dark:border-zinc-700">
          {children.map((child) => (
            <MenuNode key={child.key} node={child} childrenOf={childrenOf} activeKey={activeKey} onSelect={onSelect} openParents={openParents} onToggleParent={onToggleParent} depth={depth + 1} />
          ))}
        </div>
      )}
    </div>
  )
}

function ModuleGroup({ moduleLabel, module, roots, childrenOf, open, onToggle, activeKey, onSelect, openParents, onToggleParent }) {
  return (
    <div className="mt-0.5 flex flex-col gap-1">
      <button type="button" onClick={onToggle} aria-expanded={open} className="flex w-full items-center gap-2 rounded-lg border-l-2 border-violet-500/70 bg-zinc-100 py-2 pl-2.5 pr-2 text-sm font-bold uppercase tracking-wide text-zinc-600 transition-colors hover:bg-zinc-200 dark:border-violet-400/60 dark:bg-zinc-800 dark:text-zinc-200 dark:hover:bg-zinc-700">
        <span className="flex h-5 w-5 shrink-0 items-center justify-center rounded border border-current text-[11px] leading-none">{open ? '−' : '+'}</span>
        <span className="truncate">{moduleLabel || `MODULE ${module}`}</span>
      </button>
      {open && (
        <div className="flex flex-col gap-1">
          {roots.map((n) => (
            <MenuNode key={n.key} node={n} childrenOf={childrenOf} activeKey={activeKey} onSelect={onSelect} openParents={openParents} onToggleParent={onToggleParent} />
          ))}
        </div>
      )}
    </div>
  )
}

function Shell() {
  const toast = useToast()
  const [mode, setMode] = useState('homepage')
  const [view, setView] = useState('')
  const [loggedIn, setLoggedIn] = useState(api.isLoggedIn())
  const [collapsed, setCollapsed] = useState(sidebarPref)
  const [sessionExpired, setSessionExpired] = useState(false)
  const [sessionTick, setSessionTick] = useState(0)
  const [lastMe, setLastMe] = useState(() => api.currentUser())
  const [permissions, setPermissions] = useState([])
  const [myMenus, setMyMenus] = useState([])
  const [profileLoaded, setProfileLoaded] = useState(false)
  const [pendingLogin, setPendingLogin] = useState(false)
  const [openModules, setOpenModules] = useState(() => loadExpanded('go-core-modules'))
  const [openParents, setOpenParents] = useState(() => loadExpanded('go-core-parents'))
  const me = api.currentUser() || (sessionExpired ? lastMe : null)
  const isAdmin = me?.role === 'ADMIN'
  const groups = useMemo(() => buildSidebar(myMenus), [myMenus])
  const openableNodes = useMemo(() => groups.flatMap((g) => [...g.roots, ...[...g.childrenOf.values()].flat()].filter((n) => !n.header)), [groups])
  const allNodes = openableNodes
  const hasAccess = allNodes.length > 0
  const activeNode = allNodes.find((m) => m.key === view) || allNodes[0] || null
  const Active = activeNode && !activeNode.missing ? activeNode.Component : null

  useEffect(() => {
    if (view !== '') return
    let key = ''
    try {
      const m = (window.location.hash || '').match(/^#\/([A-Za-z0-9_]+)/)
      if (m) key = m[1]
    } catch {
      key = ''
    }
    if (key && allNodes.some((n) => n.key === key)) setView(key)
  }, [allNodes, view])

  function toggleOpen(setter, storeKey, key) {
    setter((prev) => {
      const next = { ...prev }
      if (next[key] === false) delete next[key]
      else next[key] = false
      saveExpanded(storeKey, next)
      return next
    })
  }

  const scrollRef = useRef(null)
  const [canUp, setCanUp] = useState(false)
  const [canDown, setCanDown] = useState(false)
  const updateScroll = useCallback(() => {
    const el = scrollRef.current
    if (!el) return
    setCanUp(el.scrollTop > 6)
    setCanDown(el.scrollTop + el.clientHeight < el.scrollHeight - 6)
  }, [])
  useEffect(() => { updateScroll() }, [groups, collapsed, openModules, openParents, updateScroll])
  useEffect(() => {
    const el = scrollRef.current
    if (!el) return
    const ro = new ResizeObserver(updateScroll)
    ro.observe(el)
    el.addEventListener('scroll', updateScroll, { passive: true })
    return () => { ro.disconnect(); el.removeEventListener('scroll', updateScroll) }
  }, [updateScroll])
  function scrollByAmount(dir) {
    const el = scrollRef.current
    if (!el) return
    el.scrollBy({ top: dir * 160, behavior: 'smooth' })
  }
  const isOpen = (map, key) => map[key] !== false

  useEffect(() => {
    if (!loggedIn) return
    return onSessionExpired(() => {
      setSessionExpired(true)
      toast.warning('Your session has expired. Please sign in again.', { title: 'Session expired' })
    })
  }, [loggedIn, toast])

  useEffect(() => {
    if (!loggedIn) { setProfileLoaded(false); return }
    if (profileLoaded) return
    api.getMe().then((data) => { setPermissions((data && data.permissions) || []); setMyMenus((data && data.menus) || []); setLastMe(data); setProfileLoaded(true) }).catch(() => { setPermissions([]); setMyMenus([]); setProfileLoaded(true) })
  }, [loggedIn, profileLoaded])

  async function logout() {
    await api.logout()
    setSessionExpired(false); setLastMe(null); setPermissions([]); setMyMenus([]); setLoggedIn(false); setView(''); setMode('homepage'); setProfileLoaded(false); setPendingLogin(false)
    toast.info('You have signed out. See you again!')
  }

  async function loadPermissions() {
    try { const data = await api.getMe(); setPermissions((data && data.permissions) || []); setMyMenus((data && data.menus) || []); setLastMe(data); setProfileLoaded(true) } catch { setPermissions([]); setMyMenus([]); setLastMe(api.currentUser()); setProfileLoaded(true) }
  }

  function handleAuth() {
    setLoggedIn(true); setView(''); setMyMenus([]); setPermissions([]); setProfileLoaded(false); setPendingLogin(true)
  }

  function handleRelogin() {
    setSessionExpired(false); setLoggedIn(true); setSessionTick((t) => t + 1); loadPermissions()
    toast.success('Session restored. Welcome back!')
  }

  useEffect(() => {
    if (!pendingLogin || !profileLoaded) return
    setPendingLogin(false)
    if (hasAccess) { setMode('app'); toast.success('Welcome back!') }
    else { setMode('homepage'); toast.info('You have no dashboard access.', { title: 'Homepage only' }) }
  }, [pendingLogin, profileLoaded, hasAccess])

  useEffect(() => {
    if (loggedIn && profileLoaded && !hasAccess && mode === 'app') setMode('homepage')
  }, [loggedIn, profileLoaded, hasAccess, mode])

  function handleAccountDeleted() {
    setSessionExpired(false); setLastMe(null); setPermissions([]); setMyMenus([]); setLoggedIn(false); setView(''); setMode('homepage'); setProfileLoaded(false); setPendingLogin(false)
    toast.warning('Your account has been deleted.', { title: 'Account deleted' })
  }

  function toggleSidebar() {
    setCollapsed((c) => { try { localStorage.setItem('go-core-sidebar', c ? 'open' : 'collapsed') } catch {} return !c })
  }

  if (mode === 'homepage') {
    return (
      <div className="anim-boot">
        <HomePage onLogin={() => setMode('login')} onDashboard={() => setMode('app')} onLogout={logout} loggedIn={loggedIn} hasAccess={hasAccess} profileLoaded={profileLoaded} />
      </div>
    )
  }

  if (!loggedIn) {
    return (
      <div className="anim-boot">
        <LoginForm onDone={handleAuth} onBack={() => setMode('homepage')} />
        <Modal open={sessionExpired} title="Session expired — sign in again" size="sm" showClose={false} closeOnBackdrop={false} bodyClassName="h-[320px] overflow-hidden px-5 py-4 text-sm text-zinc-600 dark:text-zinc-300">
          <div className="flex h-full flex-col gap-3">
            <p className="text-sm text-zinc-600 dark:text-zinc-300">Your session has expired. Please sign in again to continue on this page.</p>
            <LoginForm bare onDone={handleRelogin} />
          </div>
        </Modal>
      </div>
    )
  }

  return (
    <div className="anim-boot flex min-h-screen w-full bg-zinc-50 text-zinc-900 dark:bg-zinc-950 dark:text-zinc-100">
      <aside className={`sticky top-0 z-30 hidden h-screen shrink-0 flex-col border-r border-zinc-200 bg-white p-3 duration-200 ease-[cubic-bezier(0.4,0,0.2,1)] transition-[width] md:flex dark:border-zinc-800 dark:bg-zinc-900 ${collapsed ? 'w-[84px]' : 'w-60'}`}>
        <div className="pb-4">
          <div className={`relative flex items-center ${collapsed ? 'justify-center px-0' : 'gap-2 px-1'}`}>
            <AppLogo className="h-8 w-8" />
            <span className={`grid whitespace-nowrap text-sm font-bold tracking-tight transition-all duration-200 ease-[cubic-bezier(0.4,0,0.2,1)] ${collapsed ? 'grid-cols-[0fr] opacity-0' : 'grid-cols-[1fr] opacity-100'}`}>
              <span className="min-w-0 overflow-hidden">{APP_NAME}</span>
            </span>
            <Tooltip label={collapsed ? 'Expand sidebar' : 'Collapse sidebar'} position="right" className="absolute -right-5 top-1/2 -translate-y-1/2">
              <button type="button" onClick={toggleSidebar} aria-label={collapsed ? 'Expand sidebar' : 'Collapse sidebar'} className="shrink-0 rounded-full border border-zinc-200 bg-white p-1.5 text-zinc-500 shadow-md transition-all duration-200 hover:bg-zinc-50 hover:text-zinc-800 hover:shadow-lg active:scale-90 dark:border-zinc-700 dark:bg-zinc-800 dark:text-zinc-400 dark:hover:bg-zinc-700 dark:hover:text-zinc-100">
                <Icon name="chevronLeft" className={`h-4 w-4 transition-transform duration-200 ease-[cubic-bezier(0.4,0,0.2,1)] ${collapsed ? 'rotate-180' : ''}`} />
              </button>
            </Tooltip>
          </div>
        </div>
        <div className="relative flex min-h-0 flex-1 flex-col">
          <div className={`pointer-events-none absolute inset-x-0 top-0 z-10 h-6 bg-gradient-to-b from-white via-white/70 to-transparent transition-opacity dark:from-zinc-900 dark:via-zinc-900/70 ${canUp ? 'opacity-100' : 'opacity-0'}`} aria-hidden="true" />
          {canUp && !collapsed && <button type="button" onClick={() => scrollByAmount(-1)} aria-label="Scroll up" className="absolute left-1/2 top-1 z-20 -translate-x-1/2 rounded-full bg-white p-1 text-zinc-500 shadow-sm ring-1 ring-zinc-200 transition hover:bg-zinc-50 hover:text-zinc-700 active:scale-95 dark:bg-zinc-800 dark:text-zinc-400 dark:ring-zinc-700 dark:hover:bg-zinc-700"><Icon name="arrowUp" className="h-3.5 w-3.5" /></button>}
          <nav ref={scrollRef} onScroll={updateScroll} className="sidebar-scroll flex flex-1 flex-col gap-1 overflow-y-auto scroll-smooth">
            {collapsed ? allNodes.map((n) => {
              const btn = <button key={n.key} type="button" onClick={() => setView(n.key)} className={`flex w-full items-center justify-center gap-0 rounded-xl px-3 py-2.5 transition-colors ${activeNode?.key === n.key ? 'bg-zinc-900 text-white shadow-sm dark:bg-zinc-100 dark:text-zinc-900' : 'text-zinc-600 hover:bg-zinc-100 dark:text-zinc-300 dark:hover:bg-zinc-800'}`}><Icon name={n.missing ? 'warning' : n.icon || 'list'} className="h-5 w-5 shrink-0" /></button>
              return <Tooltip key={n.key} label={n.missing ? `${n.label} (not created)` : n.label} position="right">{btn}</Tooltip>
            }) : groups.map((g) => <ModuleGroup key={g.module} module={g.module} moduleLabel={g.moduleLabel} roots={g.roots} childrenOf={g.childrenOf} open={isOpen(openModules, g.module)} onToggle={() => toggleOpen(setOpenModules, 'go-core-modules', g.module)} activeKey={activeNode?.key} onSelect={setView} openParents={openParents} onToggleParent={(key) => toggleOpen(setOpenParents, 'go-core-parents', key)} />)}
          </nav>
          <div className={`pointer-events-none absolute inset-x-0 bottom-0 z-10 h-6 bg-gradient-to-t from-white via-white/70 to-transparent transition-opacity dark:from-zinc-900 dark:via-zinc-900/70 ${canDown ? 'opacity-100' : 'opacity-0'}`} aria-hidden="true" />
          {canDown && !collapsed && <button type="button" onClick={() => scrollByAmount(1)} aria-label="Scroll down" className="absolute bottom-1 left-1/2 z-20 -translate-x-1/2 rounded-full bg-white p-1 text-zinc-500 shadow-sm ring-1 ring-zinc-200 transition hover:bg-zinc-50 hover:text-zinc-700 active:scale-95 dark:bg-zinc-800 dark:text-zinc-400 dark:ring-zinc-700 dark:hover:bg-zinc-700"><Icon name="arrowDown" className="h-3.5 w-3.5" /></button>}
        </div>
        <div className="border-t border-zinc-200 pt-3 dark:border-zinc-800">
          <div className={`grid transition-all duration-200 ease-[cubic-bezier(0.4,0,0.2,1)] ${collapsed ? 'grid-rows-[0fr] opacity-0' : 'mb-2 grid-rows-[1fr] opacity-100'}`}>
            <div className="min-h-0 overflow-hidden">
              <div className="flex items-center gap-2 px-2">
                <span className="min-w-0 flex-1 truncate text-xs text-zinc-500"><span className="block truncate font-semibold text-zinc-700 dark:text-zinc-200">@{me?.username}</span><span className="font-mono">{me?.code}</span></span>
                <Badge tone={isAdmin ? 'danger' : 'brand'}>{me?.role}</Badge>
              </div>
            </div>
          </div>
          <div className={`flex items-center gap-1.5 ${collapsed ? 'flex-col' : ''}`}>
            {collapsed ? <Tooltip label="Sign out" position="right"><button type="button" onClick={logout} aria-label="Sign out" className="rounded-lg p-2 text-zinc-500 transition-colors hover:bg-zinc-100 hover:text-zinc-800 dark:text-zinc-400 dark:hover:bg-zinc-800 dark:hover:text-zinc-100"><Icon name="logout" className="h-5 w-5" /></button></Tooltip> : <Button variant="secondary" size="sm" fullWidth onClick={logout}><Icon name="logout" className="h-4 w-4" />Sign Out</Button>}
            <Tooltip label="Toggle theme" position="right"><ThemeToggle /></Tooltip>
          </div>
        </div>
      </aside>

      <div className="flex min-w-0 flex-1 flex-col">
        <div className="hidden items-center justify-end border-b border-zinc-200 bg-white px-4 py-2 dark:border-zinc-800 dark:bg-zinc-900 md:flex">
          <Button variant="secondary" size="sm" onClick={() => setMode('homepage')}>← Back to Homepage</Button>
        </div>
        <header className="sticky top-0 z-40 border-b border-zinc-200 bg-white/90 backdrop-blur md:hidden dark:border-zinc-800 dark:bg-zinc-950/90">
          <div className="flex items-center gap-2 px-3 py-2.5">
            <AppLogo className="h-7 w-7" />
            <button type="button" onClick={() => setMode('homepage')} className="ml-auto shrink-0 rounded-full border border-zinc-200 px-3 py-1 text-xs font-semibold text-zinc-700 dark:border-zinc-700 dark:text-zinc-300">← Homepage</button>
            <nav className="flex flex-1 items-center gap-1 overflow-x-auto">
              {allNodes.map((n) => (
                <button key={n.key} type="button" onClick={() => setView(n.key)} className={`whitespace-nowrap rounded-lg px-2.5 py-1.5 text-xs font-semibold transition ${activeNode?.key === n.key ? 'bg-violet-600 text-white' : n.missing ? 'text-amber-600 dark:text-amber-300' : 'text-zinc-600 dark:text-zinc-300'}`}>{n.label}{n.missing && ' 404'}</button>
              ))}
            </nav>
            <ThemeToggle className="shrink-0" />
            <button type="button" onClick={logout} aria-label="Sign out" className="shrink-0 rounded-lg p-1.5 text-zinc-500 hover:bg-zinc-100 dark:hover:bg-zinc-800"><Icon name="logout" className="h-5 w-5" /></button>
          </div>
        </header>

        <main className="mx-auto w-full max-w-5xl flex-1 px-3 py-4 sm:px-4 sm:py-6">
          <div key={`${activeNode?.key || 'empty'}-${sessionTick}`} className="anim-page-in">
            {activeNode?.missing ? <MissingMenu entry={activeNode} /> : Active ? <Active onNavigate={setView} onAccountDeleted={handleAccountDeleted} nvdata={activeNode} /> : <div className="rounded-2xl border border-zinc-200 bg-white p-8 text-center dark:border-zinc-800 dark:bg-zinc-900"><p className="text-sm font-semibold text-zinc-900 dark:text-zinc-50">No menu available</p><p className="mt-1 text-xs text-zinc-500">Your account has no menu access — contact admin.</p></div>}
          </div>
        </main>
        <footer className="border-t border-zinc-200 py-3 text-center text-xs text-zinc-400 dark:border-zinc-800 dark:text-zinc-500">{APP_NAME} — Go + React in one binary</footer>
      </div>

      <Modal open={sessionExpired} title="Session expired — sign in again" size="sm" showClose={false} closeOnBackdrop={false} bodyClassName="h-[320px] overflow-hidden px-5 py-4 text-sm text-zinc-600 dark:text-zinc-300">
        <div className="flex h-full flex-col gap-3">
          <p className="text-sm text-zinc-600 dark:text-zinc-300">Your session has expired. Please sign in again to continue on this page.</p>
          <LoginForm bare onDone={handleRelogin} />
        </div>
      </Modal>
    </div>
  )
}

export default function App() {
  return <ToastProvider><Shell /></ToastProvider>
}
