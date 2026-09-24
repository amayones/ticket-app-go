import { useCallback, useEffect, useState } from 'react'
import { api } from '../api/client.js'
import {
  Alert,
  Avatar,
  Badge,
  Button,
  Card,
  CardTitle,
  Icon,
  PageLoader,
  PasswordInput,
  useToast,
} from '../components'

// Dashboard: satu-satunya halaman untuk role USER; beranda untuk ADMIN.
// Berisi profil, status backend, sesi sendiri, ganti password, dan area
// persiapan modul inti (arah aplikasi belum ditentukan).
export default function Dashboard({ onNavigate }) {
  const toast = useToast()
  // Dibaca sekali saat mount agar referensi stabil (hindari fetch berulang).
  const [me] = useState(() => api.currentUser())
  const isAdmin = me?.role === 'ADMIN'
  const [profile, setProfile] = useState(null)
  const [sessionCount, setSessionCount] = useState(0)
  const [summary, setSummary] = useState(null)
  const [backendOk, setBackendOk] = useState(false)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [newPass, setNewPass] = useState('')
  const [confirmPass, setConfirmPass] = useState('')
  const [passError, setPassError] = useState('')
  const [savingPass, setSavingPass] = useState(false)

  const load = useCallback(async () => {
    if (!me) return
    setLoading(true)
    setError('')
    try {
      const [user, sessions, health] = await Promise.all([
        api.getUser(me.code),
        api.listMySessions(),
        api.health().catch(() => null),
      ])
      setProfile(user)
      setSessionCount(Array.isArray(sessions) ? sessions.length : 0)
      setBackendOk(!!health)
      if (isAdmin) {
        setSummary(await api.securitySummary().catch(() => null))
      }
    } catch (err) {
      setError(err.message)
    } finally {
      setLoading(false)
    }
  }, [isAdmin, me])

  useEffect(() => {
    load()
  }, [load])

  async function logoutAll() {
    try {
      await api.logoutAll(me.code)
      toast.warning('Semua sesi Anda dicabut. Muat ulang bila perlu login lagi.')
      load()
    } catch (err) {
      toast.error(err.message)
    }
  }

  async function changePassword(e) {
    e.preventDefault()
    setPassError('')
    if (newPass.length < 8) {
      setPassError('Password minimal 8 karakter.')
      return
    }
    if (newPass !== confirmPass) {
      setPassError('Konfirmasi password tidak sama.')
      return
    }
    setSavingPass(true)
    try {
      await api.updateUser(me.code, { password: newPass })
      setNewPass('')
      setConfirmPass('')
      toast.success('Password berhasil diganti.')
    } catch (err) {
      setPassError(err.message)
    } finally {
      setSavingPass(false)
    }
  }

  if (loading) return <PageLoader label="Memuat dashboard…" />

  return (
    <div className="flex flex-col gap-4">
      {error && (
        <Alert tone="error" title="Gagal memuat dashboard" closable onClose={() => setError('')}>
          {error}
        </Alert>
      )}

      <Card>
        <div className="flex flex-wrap items-center gap-4">
          <Avatar name={profile?.username || me?.username || '?'} size="lg" />
          <div className="min-w-0 flex-1">
            <p className="flex flex-wrap items-center gap-2 text-base font-bold text-zinc-900 dark:text-zinc-50">
              <span className="truncate">@{profile?.username || me?.username}</span>
              <Badge tone={isAdmin ? 'danger' : 'brand'}>{profile?.role_code || me?.role}</Badge>
              <Badge tone={backendOk ? 'success' : 'neutral'}>
                {backendOk ? 'Backend online' : 'Backend?'}
              </Badge>
            </p>
            <p className="mt-0.5 truncate text-xs text-zinc-500 dark:text-zinc-400">
              <span className="font-mono">{profile?.code || me?.code}</span>
              {profile?.email ? ` · ${profile.email}` : ''}
            </p>
          </div>
          <Button variant="secondary" size="sm" onClick={load}>
            <Icon name="refresh" className="h-4 w-4" />
            Muat ulang
          </Button>
        </div>
      </Card>

      <div className="grid gap-4 sm:grid-cols-2">
        <Card>
          <CardTitle description="Perangkat/browser yang sedang login sebagai Anda.">
            Sesi saya ({sessionCount})
          </CardTitle>
          <p className="mb-3 text-sm text-zinc-600 dark:text-zinc-300">
            Cabut semua sesi bila ada perangkat yang tidak dikenal.
          </p>
          <Button variant="secondary" size="sm" onClick={logoutAll}>
            <Icon name="logout" className="h-4 w-4" />
            Cabut semua sesi saya
          </Button>
        </Card>

        <Card>
          <CardTitle description="Ganti password akun sendiri.">
            Keamanan akun
          </CardTitle>
          <form onSubmit={changePassword} className="flex flex-col gap-3">
            {passError && (
              <Alert tone="error" closable>
                {passError}
              </Alert>
            )}
            <PasswordInput
              label="Password baru"
              value={newPass}
              onChange={(e) => setNewPass(e.target.value)}
              autoComplete="new-password"
              placeholder="min. 8 karakter"
            />
            <PasswordInput
              label="Konfirmasi password baru"
              value={confirmPass}
              onChange={(e) => setConfirmPass(e.target.value)}
              autoComplete="new-password"
              placeholder="ulangi password baru"
            />
            <div>
              <Button type="submit" size="sm" loading={savingPass}>
                Simpan password
              </Button>
            </div>
          </form>
        </Card>
      </div>

      {isAdmin && summary && (
        <Card>
          <CardTitle description="Ringkasan cepat tanpa membuka tiap menu.">
            Sekilas sistem
          </CardTitle>
          <div className="grid grid-cols-2 gap-2.5 sm:grid-cols-4">
            {[
              ['Pengguna', summary.total_users],
              ['Sesi aktif', summary.active_sessions],
              ['Role', summary.total_roles],
              ['Error 24 jam', summary.errors_last_24h],
            ].map(([label, value]) => (
              <div key={label} className="rounded-xl bg-zinc-50 px-3 py-2.5 text-center dark:bg-zinc-800/60">
                <p className="text-xl font-bold text-zinc-900 dark:text-zinc-50">{value}</p>
                <p className="text-xs text-zinc-500 dark:text-zinc-400">{label}</p>
              </div>
            ))}
          </div>
          <div className="mt-3 flex flex-wrap gap-2">
            {[
              ['users', 'User Account'],
              ['audit', 'Audit Log'],
              ['notif', 'Notifikasi'],
              ['security', 'Security Center'],
            ].map(([key, label]) => (
              <Button key={key} variant="secondary" size="sm" onClick={() => onNavigate(key)}>
                {label}
              </Button>
            ))}
          </div>
        </Card>
      )}

      <Card>
        <CardTitle description="Area ini disiapkan untuk modul inti begitu arah aplikasi ditentukan (mis. tabel M* untuk menu/halaman).">
          Modul aplikasi — segera hadir
        </CardTitle>
        <div className="grid gap-2.5 sm:grid-cols-3">
          {['Modul 1', 'Modul 2', 'Modul 3'].map((name) => (
            <div
              key={name}
              className="flex flex-col items-center gap-1.5 rounded-xl border border-dashed border-zinc-300 px-4 py-8 text-center dark:border-zinc-700"
            >
              <span className="flex h-10 w-10 items-center justify-center rounded-full bg-zinc-100 text-zinc-400 dark:bg-zinc-800">
                <Icon name="plus" className="h-5 w-5" />
              </span>
              <p className="text-sm font-semibold text-zinc-500 dark:text-zinc-400">{name}</p>
              <p className="text-xs text-zinc-400 dark:text-zinc-500">Slot modul — belum ditentukan</p>
            </div>
          ))}
        </div>
      </Card>
    </div>
  )
}
