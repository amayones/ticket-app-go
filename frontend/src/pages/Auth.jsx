import { useState } from 'react'
import { api } from '../api/client.js'
import { APP_NAME } from '../appName.js'
import { Alert, Button, Card, CardTitle, PasswordInput, TextField } from '../components'

// Form login. Mode penuh (default) = kartu ber-branding untuk halaman login.
// Mode bare = hanya field + tombol (untuk di dalam modal), tanpa Card/branding.
export function LoginForm({ onDone, bare = false }) {
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

  const form = (
    <form onSubmit={submit} className={`flex flex-col ${bare ? 'gap-3' : 'gap-4'}`}>
      {error &&
        (bare ? (
          <p role="alert" className="text-sm text-rose-600 dark:text-rose-400">
            {error}
          </p>
        ) : (
          <Alert tone="error" title="Login gagal" closable onClose={() => setError('')}>
            {error}
          </Alert>
        ))}
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
  )

  if (bare) return form

  return (
    <Card className="w-full">
      <div className="mb-5 flex flex-col items-center gap-2 text-center">
        <span className="flex h-12 w-12 items-center justify-center rounded-2xl bg-gradient-to-br from-violet-600 to-fuchsia-600 text-xl font-bold text-white">
          {APP_NAME.charAt(0).toUpperCase()}
        </span>
        <CardTitle description="Masuk untuk mengelola aplikasi. Akun dibuatkan oleh admin.">
          {APP_NAME}
        </CardTitle>
      </div>
      {form}
    </Card>
  )
}
