import { useState } from 'react'
import { api } from './api/client.js'
import { Button, Icon, ToastProvider, useToast } from './components'
import { LoginForm, RegisterForm } from './pages/Auth.jsx'
import UsersList from './pages/Users.jsx'

function Shell() {
  const toast = useToast()
  const [view, setView] = useState(api.isLoggedIn() ? 'users' : 'login')
  const [loggedIn, setLoggedIn] = useState(api.isLoggedIn())

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

  return (
    <div className="flex min-h-screen w-full flex-col bg-zinc-50 text-zinc-900 dark:bg-zinc-950 dark:text-zinc-100">
      <header className="sticky top-0 z-40 border-b border-zinc-200 bg-white/80 backdrop-blur dark:border-zinc-800 dark:bg-zinc-950/80">
        <div className="mx-auto flex w-full max-w-3xl items-center justify-between px-4 py-3">
          <span className="flex items-center gap-2 text-base font-bold tracking-tight">
            <span className="flex h-8 w-8 items-center justify-center rounded-lg bg-gradient-to-br from-violet-600 to-fuchsia-600 text-sm text-white">
              G
            </span>
            Go Core
          </span>
          <nav className="flex items-center gap-1.5">
            {!loggedIn ? (
              <>
                <Button variant="ghost" size="sm" onClick={() => setView('login')}>
                  <span className={view === 'login' ? 'font-bold text-violet-600 dark:text-violet-300' : ''}>Login</span>
                </Button>
                <Button variant={view === 'register' ? 'primary' : 'ghost'} size="sm" onClick={() => setView('register')}>
                  Register
                </Button>
              </>
            ) : (
              <>
                <Button variant="ghost" size="sm" onClick={() => setView('users')}>
                  <Icon name="users" className="h-4 w-4" />
                  Users
                </Button>
                <Button variant="secondary" size="sm" onClick={logout}>
                  <Icon name="logout" className="h-4 w-4" />
                  Logout
                </Button>
              </>
            )}
          </nav>
        </div>
      </header>
      <main className="mx-auto w-full max-w-3xl flex-1 px-4 py-6 sm:py-8">
        {view === 'login' && !loggedIn && <LoginForm onDone={handleAuth} />}
        {view === 'register' && !loggedIn && <RegisterForm onDone={() => setView('login')} />}
        {view === 'users' && loggedIn && <UsersList onAccountDeleted={handleAccountDeleted} />}
      </main>
      <footer className="border-t border-zinc-200 py-4 text-center text-xs text-zinc-400 dark:border-zinc-800 dark:text-zinc-500">
        Go Core — Go + React dalam satu binary
      </footer>
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
