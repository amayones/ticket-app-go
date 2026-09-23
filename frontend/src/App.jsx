import { useState } from 'react'
import { api } from './api/client.js'
import { LoginForm, RegisterForm, UsersList } from './pages/Auth.jsx'
import './App.css'

export default function App() {
  const [view, setView] = useState(api.isLoggedIn() ? 'users' : 'login')
  const [loggedIn, setLoggedIn] = useState(api.isLoggedIn())

  function go(v) {
    setView(v)
  }

  async function logout() {
    await api.logout()
    setLoggedIn(false)
    setView('login')
  }

  function handleAuth() {
    setLoggedIn(true)
    setView('users')
  }

  return (
    <div className="app">
      <header className="topbar">
        <strong>Go Core</strong>
        <nav>
          {!loggedIn && (
            <>
              <button type="button" className={view === 'login' ? 'active' : ''} onClick={() => go('login')}>Login</button>
              <button type="button" className={view === 'register' ? 'active' : ''} onClick={() => go('register')}>Register</button>
            </>
          )}
          {loggedIn && (
            <>
              <button type="button" className={view === 'users' ? 'active' : ''} onClick={() => go('users')}>Users</button>
              <button type="button" onClick={logout}>Logout</button>
            </>
          )}
        </nav>
      </header>
      <main className="main">
        {view === 'login' && !loggedIn && <LoginForm onDone={handleAuth} />}
        {view === 'register' && !loggedIn && <RegisterForm onDone={() => go('login')} />}
        {view === 'users' && loggedIn && <UsersList />}
      </main>
    </div>
  )
}
