import { useState } from 'react'
import { api } from '../api/client.js'
import { Alert, AppLogo, Button, PasswordInput, TextField } from '../components'
import { APP_NAME } from '../brand.js'

export function LoginForm({ onDone, bare = false, onBack }) {
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
    <form onSubmit={submit} className="flex flex-col gap-4">
      {error &&
        (bare ? (
          <p role="alert" className="rounded-xl bg-rose-50 px-3 py-2 text-sm text-rose-600 dark:bg-rose-950/50 dark:text-rose-400">
            {error}
          </p>
        ) : (
          <Alert tone="error" title="Login failed" closable onClose={() => setError('')}>
            {error}
          </Alert>
        ))}
      <TextField
        label="Username"
        value={username}
        onChange={(e) => setUsername(e.target.value)}
        autoComplete="username"
        placeholder="username"
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
      <Button type="submit" loading={loading} fullWidth className="!rounded-full !bg-[#6b5cff] !py-2.5 hover:!bg-[#5a4af0] shadow-lg shadow-violet-200 dark:shadow-none">
        {loading ? 'Checking…' : 'Sign In'}
      </Button>
    </form>
  )

  if (bare) return form

  return (
    <div className="flex min-h-screen w-full items-center justify-center bg-[#f4f1ff] px-4 py-6 dark:bg-zinc-950">
      <div className="grid w-full max-w-[980px] overflow-hidden rounded-[32px] bg-white shadow-[0_24px_64px_-24px_rgba(80,60,180,0.35)] ring-1 ring-zinc-100 md:grid-cols-[1.05fr_0.95fr] dark:bg-zinc-900 dark:ring-zinc-800">
        <div className="flex flex-col px-7 py-8 sm:px-10 sm:py-10">
          {onBack && (
            <button type="button" onClick={onBack} className="mb-4 inline-flex w-fit items-center gap-1 text-xs font-medium text-zinc-500 hover:text-zinc-700 dark:text-zinc-400 dark:hover:text-zinc-200">
              ← Back to Homepage
            </button>
          )}
          <div className="mb-8 flex items-center gap-2.5">
            <AppLogo className="h-9 w-9 rounded-xl" />
            <span className="text-sm font-bold tracking-tight text-zinc-900 dark:text-white">{APP_NAME}</span>
          </div>
          <div className="mb-6">
            <h1 className="text-[30px] font-extrabold leading-none tracking-tight text-zinc-900 dark:text-white">Sign In</h1>
            <p className="mt-2.5 text-sm leading-relaxed text-zinc-500 dark:text-zinc-400">Sign in to manage the application. Accounts are created by admin.</p>
          </div>
          {form}
          <p className="mt-6 text-center text-xs text-zinc-400 dark:text-zinc-500">
            Don&apos;t have an account? <span className="font-medium text-zinc-600 dark:text-zinc-300">Contact admin</span>
          </p>
        </div>
        <div className="relative hidden overflow-hidden bg-[#6b5cff] md:block">
          <img src="/login-card.png" alt="" className="absolute inset-0 h-full w-full object-cover" draggable={false} />
        </div>
      </div>
    </div>
  )
}
