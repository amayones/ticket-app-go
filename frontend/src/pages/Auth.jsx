import { useState } from 'react'
import { api } from '../api/client.js'

export function LoginForm({ onDone }) {
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)

  async function submit(e) {
    e.preventDefault()
    setError('')
    setLoading(true)
    try {
      await api.login(username.trim(), password)
      onDone()
    } catch (err) {
      setError(err.message)
    } finally {
      setLoading(false)
    }
  }

  return (
    <form className="card" onSubmit={submit}>
      <h2>Masuk</h2>
      <label>
        Username
        <input value={username} onChange={(e) => setUsername(e.target.value)} autoComplete="username" required />
      </label>
      <label>
        Password
        <input type="password" value={password} onChange={(e) => setPassword(e.target.value)} autoComplete="current-password" required />
      </label>
      {error && <p className="error">{error}</p>}
      <button type="submit" disabled={loading}>{loading ? '...' : 'Login'}</button>
    </form>
  )
}

export function RegisterForm({ onDone }) {
  const [username, setUsername] = useState('')
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState('')
  const [ok, setOk] = useState('')
  const [loading, setLoading] = useState(false)

  async function submit(e) {
    e.preventDefault()
    setError('')
    setOk('')
    setLoading(true)
    try {
      await api.register(username.trim(), email.trim().toLowerCase(), password)
      setOk('Registrasi berhasil, silakan login.')
      onDone?.()
    } catch (err) {
      setError(err.message)
    } finally {
      setLoading(false)
    }
  }

  return (
    <form className="card" onSubmit={submit}>
      <h2>Daftar</h2>
      <label>
        Username (min 3)
        <input value={username} onChange={(e) => setUsername(e.target.value)} autoComplete="username" required />
      </label>
      <label>
        Email
        <input type="email" value={email} onChange={(e) => setEmail(e.target.value)} autoComplete="email" required />
      </label>
      <label>
        Password (min 8, max 72)
        <input type="password" value={password} onChange={(e) => setPassword(e.target.value)} autoComplete="new-password" required />
      </label>
      {error && <p className="error">{error}</p>}
      {ok && <p className="ok">{ok}</p>}
      <button type="submit" disabled={loading}>{loading ? '...' : 'Register'}</button>
    </form>
  )
}

export function UsersList() {
  const [users, setUsers] = useState([])
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)

  async function load() {
    setError('')
    setLoading(true)
    try {
      setUsers(await api.listUsers())
    } catch (err) {
      setError(err.message)
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="card">
      <h2>Users</h2>
      <button type="button" onClick={load} disabled={loading}>{loading ? '...' : 'Muat daftar'}</button>
      {error && <p className="error">{error}</p>}
      <ul className="users">
        {users.map((u) => (
          <li key={u.id}>#{u.id} {u.username} — {u.email}</li>
        ))}
      </ul>
    </div>
  )
}
