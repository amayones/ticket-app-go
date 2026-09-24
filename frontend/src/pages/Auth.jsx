import { useState } from 'react'
import { api } from '../api/client.js'
import { Alert, Button, Card, CardTitle, PasswordInput, TextField, useToast } from '../components'

export function LoginForm({ onDone }) {
  const toast = useToast()
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
      toast.success('Selamat datang kembali!')
      onDone()
    } catch (err) {
      setError(err.message)
    } finally {
      setLoading(false)
    }
  }

  return (
    <Card className="mx-auto w-full max-w-md">
      <CardTitle description="Masuk untuk mengelola akun pengguna.">Masuk</CardTitle>
      <form onSubmit={submit} className="flex flex-col gap-4">
        {error && (
          <Alert tone="error" title="Login gagal" closable onClose={() => setError('')}>
            {error}
          </Alert>
        )}
        <TextField
          label="Username"
          value={username}
          onChange={(e) => setUsername(e.target.value)}
          autoComplete="username"
          placeholder="nama pengguna"
          required
        />
        <PasswordInput
          label="Password"
          value={password}
          onChange={(e) => setPassword(e.target.value)}
          autoComplete="current-password"
          placeholder="••••••••"
          required
        />
        <Button type="submit" loading={loading} fullWidth>
          {loading ? 'Memeriksa…' : 'Login'}
        </Button>
      </form>
    </Card>
  )
}

export function RegisterForm({ onDone }) {
  const toast = useToast()
  const [username, setUsername] = useState('')
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)

  async function submit(e) {
    e.preventDefault()
    setError('')
    setLoading(true)
    try {
      await api.register(username.trim(), email.trim().toLowerCase(), password)
      toast.success('Registrasi berhasil. Silakan login.', { title: 'Akun dibuat' })
      onDone?.()
    } catch (err) {
      setError(err.message)
    } finally {
      setLoading(false)
    }
  }

  return (
    <Card className="mx-auto w-full max-w-md">
      <CardTitle description="Buat akun baru untuk mengakses dashboard.">Daftar Akun</CardTitle>
      <form onSubmit={submit} className="flex flex-col gap-4">
        {error && (
          <Alert tone="error" title="Registrasi gagal" closable onClose={() => setError('')}>
            {error}
          </Alert>
        )}
        <TextField
          label="Username"
          value={username}
          onChange={(e) => setUsername(e.target.value)}
          autoComplete="username"
          placeholder="min. 3 karakter"
          required
        />
        <TextField
          label="Email"
          type="email"
          value={email}
          onChange={(e) => setEmail(e.target.value)}
          autoComplete="email"
          placeholder="nama@email.com"
          required
        />
        <PasswordInput
          label="Password"
          value={password}
          onChange={(e) => setPassword(e.target.value)}
          autoComplete="new-password"
          placeholder="min. 8 karakter"
          hint="Gunakan kombinasi huruf, angka, dan simbol agar lebih aman."
          required
        />
        <Button type="submit" loading={loading} fullWidth>
          {loading ? 'Mendaftar…' : 'Register'}
        </Button>
      </form>
    </Card>
  )
}
