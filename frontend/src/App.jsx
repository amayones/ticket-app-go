import { useState } from 'react'
import { api } from './api/client.js'
import { Badge, Button, Icon, ToastProvider, useToast } from './components'
import { LoginForm, RegisterForm } from './pages/Auth.jsx'
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

function Shell() {
  const toast = useToast()
  const [view, setView] = useState(api.isLoggedIn() ? 'users' : 'login')
  const [loggedIn, setLoggedIn] = useState(api.isLoggedIn())
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
  }

  function handleAccountDeleted() {
    setLoggedIn(false)
    setView('login')
    toast.warning('Akun Anda telah dihapus.', { title: 'Akun dihapus' })
  }

  if (!loggedIn) {
    return (
      <div className="flex min-h-screen w-full flex-col bg-zinc-50 text-zinc-900 dark:bg-zinc-950 dark:text-zinc-100">
        <header className="border-b border-zinc-200 bg-white/80 backdrop-blur dark:border-zinc-800 dark:bg-zinc-950/80">
          <div className="mx-auto flex w-full max-w-3xl items-center justify-between px-4 py-3">
            <span className="flex items-center gap-2 text-base font-bold tracking-tight">
              <span className="flex h-8 w-8 items-center justify-center rounded-lg bg-gradient-to-br from-violet-600 to-fuchsia-600 text-sm text-white">
                G
              </span>
              Go Core
            </span>
            <nav className="flex items-center gap-1.5">
              <Button variant="ghost" size="sm" onClick={() => setView('login')}>
                <span className={view === 'login' ? 'font-bold text-violet-600 dark:text-violet-300' : ''}>Login</span>
              </Button>
              <Button variant={view === 'register' ? 'primary' : 'ghost'} size="sm" onClick={() => setView('register')}>
                Register
              </Button>
            </nav>
          </div>
        </header>
        <main className="mx-auto w-full max-w-3xl flex-1 px-4 py-6 sm:py-8">
          {view === 'login' && <LoginForm onDone={handleAuth} />}
          {view === 'register' && <RegisterForm onDone={() => setView('login')} />}
        </main>
        <footer className="border-t border-zinc-200 py-4 text-center text-xs text-zinc-400 dark:border-zinc-800 dark:text-zinc-500">
          Go Core — Go + React dalam satu binary
        </footer>
      </div>
    )
  }

  return (
    <div className="flex min-h-screen w-full bg-zinc-50 text-zinc-900 dark:bg-zinc-950 dark:text-zinc-100">
      {/* Sidebar (desktop) */}
      <aside className="sticky top-0 hidden h-screen w-60 shrink-0 flex-col border-r border-zinc-200 bg-white p-4 md:flex dark:border-zinc-800 dark:bg-zinc-900">
        <span className="flex items-center gap-2 px-2 pb-4 text-base font-bold tracking-tight">
          <span className="flex h-8 w-8 items-center justify-center rounded-lg bg-gradient-to-br from-violet-600 to-fuchsia-600 text-sm text-white">
            G
          </span>
          Go Core
        </span>
        <nav className="flex flex-1 flex-col gap-1 overflow-y-auto">
          {visibleNav.map((n) => (
            <button
              key={n.key}
              type="button"
              onClick={() => setView(n.key)}
              className={`flex items-center gap-2.5 rounded-xl px-3 py-2.5 text-sm font-medium transition ${
                activeNav === n.key
                  ? 'bg-violet-600 text-white shadow-sm'
                  : 'text-zinc-600 hover:bg-zinc-100 dark:text-zinc-300 dark:hover:bg-zinc-800'
              }`}
            >
              <Icon name={n.icon} className="h-5 w-5 shrink-0" />
              {n.label}
            </button>
          ))}
        </nav>
        <div className="border-t border-zinc-200 pt-3 dark:border-zinc-800">
          <div className="mb-2 flex items-center gap-2 px-2">
            <span className="min-w-0 flex-1 truncate text-xs text-zinc-500">
              <span className="block truncate font-semibold text-zinc-700 dark:text-zinc-200">@{me?.username}</span>
              <span className="font-mono">{me?.code}</span>
            </span>
            <Badge tone={isAdmin ? 'danger' : 'brand'}>{me?.role}</Badge>
          </div>
          <Button variant="secondary" size="sm" fullWidth onClick={logout}>
            <Icon name="logout" className="h-4 w-4" />
            Logout
          </Button>
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

        <main className="mx-auto w-full max-w-4xl flex-1 px-3 py-4 sm:px-4 sm:py-6">
          {activeNav === 'users' && <UsersList onAccountDeleted={handleAccountDeleted} />}
          {activeNav === 'roles' && isAdmin && <Roles />}
          {activeNav === 'sessions' && <Sessions />}
          {activeNav === 'audit' && isAdmin && <Audit />}
          {activeNav === 'security' && isAdmin && <Security />}
          {activeNav === 'syslog' && isAdmin && <Syslog />}
          {activeNav === 'notif' && isAdmin && <Notifications />}
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
