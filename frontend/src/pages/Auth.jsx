import { useState } from 'react'
import { api } from '../api/client.js'
import { Alert, Button, Card, CardTitle, PasswordInput, TextField } from '../components'

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
    <Card className="w-full">
      <div className="mb-5 flex flex-col items-center gap-2 text-center">
        <span className="flex h-12 w-12 items-center justify-center rounded-2xl bg-gradient-to-br from-violet-600 to-fuchsia-600 text-xl font-bold text-white">
          G
        </span>
        <CardTitle description="Masuk untuk mengelola aplikasi. Akun dibuatkan oleh admin.">
          Go Core
        </CardTitle>
      </div>
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
