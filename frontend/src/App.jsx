import { useState } from 'react'
import { api } from './api/client.js'
import { Badge, Button, Icon, ThemeToggle, ToastProvider, Tooltip, useToast } from './components'
import { LoginForm } from './pages/Auth.jsx'
import Audit from './pages/Audit.jsx'
import Notifications from './pages/Notifications.jsx'
import Roles from './pages/Roles.jsx'
import Security from './pages/Security.jsx'
import Sessions from './pages/Sessions.jsx'
import Syslog from './pages/Syslog.jsx'
import UsersList from './pages/Users.jsx'

const NAV = [
  { key: 'users', label: 'User Account', icon: 'users', admin: false },
  { key: 'roles', label: 'Role & Permission', icon: 'shield', admin: true },
  { key: 'sessions', label: 'Sesi & Auth', icon: 'key', admin: false },
  { key: 'audit', label: 'Audit Log', icon: 'list', admin: true },
  { key: 'security', label: 'Security Center', icon: 'shield', admin: true },
  { key: 'syslog', label: 'System Log', icon: 'terminal', admin: true },
  { key: 'notif', label: 'Notifikasi', icon: 'bell', admin: true },
]

function sidebarPref() {
  try {
    return localStorage.getItem('go-core-sidebar') === 'collapsed'
  } catch {
    return false
  }
}

function Shell() {
  const toast = useToast()
  const [view, setView] = useState(api.isLoggedIn() ? 'users' : 'login')
  const [loggedIn, setLoggedIn] = useState(api.isLoggedIn())
  const [collapsed, setCollapsed] = useState(sidebarPref)
  const me = api.currentUser()
  const isAdmin = me?.role === 'ADMIN'
  const visibleNav = NAV.filter((n) => !n.admin || isAdmin)
  const activeNav = visibleNav.some((n) => n.key === view) ? view : 'users'

  async function logout() {
    await api.logout()
    setLoggedIn(false)
    setView('login')
    toast.info('Anda telah keluar. Sampai jumpa!')
  }

  function handleAuth() {
    setLoggedIn(true)
    setView('users')
    toast.success('Selamat datang kembali!')
  }

  function handleAccountDeleted() {
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
      {/* Sidebar (desktop) — bisa dilipat via tombol chevron.
          Hanya lebar yang ditransisikan (bukan all) agar tidak bergetar. */}
      <aside
        className={`sticky top-0 z-30 hidden h-screen shrink-0 flex-col border-r border-zinc-200 bg-white duration-200 ease-[cubic-bezier(0.4,0,0.2,1)] transition-[width,padding] md:flex dark:border-zinc-800 dark:bg-zinc-900 ${
          collapsed ? 'w-[76px] p-2' : 'w-60 p-3'
        }`}
      >
        {/* Baris logo tanpa padding-bawah: tombol tepat sejajar tengah logo.
            Teks dilipat via grid 0fr->1fr (mulus, tanpa teks tersendat). */}
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
          {visibleNav.map((n) => {
            const btn = (
              <button
                key={n.key}
                type="button"
                onClick={() => setView(n.key)}
              className={`flex w-full items-center rounded-xl px-3 py-2.5 transition-colors ${
                collapsed ? 'justify-center gap-0' : 'gap-2.5'
              } ${
                  activeNav === n.key
                    ? 'bg-zinc-900 text-white shadow-sm dark:bg-zinc-100 dark:text-zinc-900'
                    : 'text-zinc-600 hover:bg-zinc-100 dark:text-zinc-300 dark:hover:bg-zinc-800'
                }`}
              >
              <Icon name={n.icon} className="h-5 w-5 shrink-0" />
              <span
                className={`grid whitespace-nowrap text-[13px] font-medium transition-all duration-200 ease-[cubic-bezier(0.4,0,0.2,1)] ${
                  collapsed ? 'grid-cols-[0fr] opacity-0' : 'grid-cols-[1fr] opacity-100'
                }`}
              >
                <span className="min-w-0 overflow-hidden">{n.label}</span>
              </span>
            </button>
            )
            return collapsed ? (
              <Tooltip key={n.key} label={n.label} position="right">
                {btn}
              </Tooltip>
            ) : (
              btn
            )
          })}
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
              {visibleNav.map((n) => (
                <button
                  key={n.key}
                  type="button"
                  onClick={() => setView(n.key)}
                  className={`whitespace-nowrap rounded-lg px-2.5 py-1.5 text-xs font-semibold transition ${
                    activeNav === n.key
                      ? 'bg-violet-600 text-white'
                      : 'text-zinc-600 dark:text-zinc-300'
                  }`}
                >
                  {n.label}
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
          {/* key memicu animasi masuk yang halus tiap ganti menu */}
          <div key={activeNav} className="anim-page-in">
            {activeNav === 'users' && <UsersList onAccountDeleted={handleAccountDeleted} />}
            {activeNav === 'roles' && isAdmin && <Roles />}
            {activeNav === 'sessions' && <Sessions />}
            {activeNav === 'audit' && isAdmin && <Audit />}
            {activeNav === 'security' && isAdmin && <Security />}
            {activeNav === 'syslog' && isAdmin && <Syslog />}
            {activeNav === 'notif' && isAdmin && <Notifications />}
          </div>
        </main>
        <footer className="border-t border-zinc-200 py-3 text-center text-xs text-zinc-400 dark:border-zinc-800 dark:text-zinc-500">
          Go Core — Go + React dalam satu binary
        </footer>
      </div>
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
