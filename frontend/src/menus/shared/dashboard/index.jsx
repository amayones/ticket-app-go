import { useCallback, useEffect, useState } from 'react'
import { api } from '../../../api/client.js'
import { getUser, updateUser } from '../../admin/users/api.js'
import { securitySummary } from '../../admin/security/api.js'

export const meta = { label: 'Dashboard', icon: 'home', order: 0 }
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
  useSmoothLoading,
  useToast,
} from '../../../components'

// Dashboard: satu-satunya halaman untuk role USER; beranda untuk ADMIN.
// Berisi profil, ganti password, dan ringkasan + pintasan (khusus admin).
export default function Dashboard({ onNavigate }) {
  const toast = useToast()
  // Dibaca sekali saat mount agar referensi stabil (hindari fetch berulang).
  const [me] = useState(() => api.currentUser())
  const isAdmin = me?.role === 'ADMIN'
  const [profile, setProfile] = useState(null)
  const [summary, setSummary] = useState(null)
  const [loading, setLoading] = useState(true)
  const showLoading = useSmoothLoading(loading)
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
      setProfile(await getUser(me.code))
      if (isAdmin) {
        setSummary(await securitySummary().catch(() => null))
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
      await updateUser(me.code, { password: newPass })
      setNewPass('')
      setConfirmPass('')
      toast.success('Password berhasil diganti.')
    } catch (err) {
      setPassError(err.message)
    } finally {
      setSavingPass(false)
    }
  }

  // Muat pertama: fullscreen loader. Refresh berikut: konten tetap tampil
  // (tidak berkedip) lalu data diperbarui — sama halusnya dengan halaman lain.
  if (showLoading && !profile) return <PageLoader label="Memuat dashboard…" />

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

    </div>
  )
}
