import { useEffect, useState } from 'react'
import { updateUser, updateUserRole } from '../api.js'
import { listRoles } from '../../roles/api.js'
import { Alert, Button, Modal, TextField, useToast } from '../../../../components'

// Modal edit user: hanya mengirim field yang diubah (patch parsial).
// Dipasang dengan key={user.code} oleh induk agar form ter-reset tiap ganti user.
export default function EditUserModal({ user, onClose, onSaved }) {
  const toast = useToast()
  const canAssignRole = true
  const [username, setUsername] = useState(user?.username || '')
  const [email, setEmail] = useState(user?.email || '')
  const [password, setPassword] = useState('')
  const [role, setRole] = useState(user?.role_code || 'USER')
  const [roles, setRoles] = useState([])
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)

  useEffect(() => {
    if (canAssignRole) {
      listRoles().then(setRoles).catch(() => setRoles([]))
    }
  }, [canAssignRole])

  async function submit(e) {
    e.preventDefault()
    if (!user) return
    setError('')
    const patch = {}
    const nextUsername = username.trim()
    const nextEmail = email.trim().toLowerCase()
    if (nextUsername && nextUsername !== user.username) patch.username = nextUsername
    if (nextEmail && nextEmail !== user.email) patch.email = nextEmail
    if (password) patch.password = password
    const roleChanged = canAssignRole && role !== user.role_code
    if (Object.keys(patch).length === 0 && !roleChanged) {
      setError('Tidak ada perubahan. Ubah salah satu field terlebih dulu.')
      return
    }
    setLoading(true)
    try {
      if (Object.keys(patch).length > 0) {
        await updateUser(user.code, patch)
      }
      if (roleChanged) {
        await updateUserRole(user.code, role)
        toast.success(`Role @${user.username} diubah ke ${role}.`)
      } else {
        toast.success('Profil berhasil diperbarui.')
      }
      onSaved()
      onClose()
    } catch (err) {
      setError(err.message)
    } finally {
      setLoading(false)
    }
  }

  return (
    <Modal
      open={!!user}
      onClose={loading ? undefined : onClose}
      title={`Edit @${user?.username || ''}`}
      closeOnBackdrop={!loading}
      footer={
        <>
          <Button variant="secondary" onClick={onClose} disabled={loading}>
            Batal
          </Button>
          <Button onClick={submit} loading={loading}>
            Simpan perubahan
          </Button>
        </>
      }
    >
      <form onSubmit={submit} className="flex flex-col gap-4">
        {error && (
          <Alert tone="error" closable>
            {error}
          </Alert>
        )}
        <TextField
          label="Username"
          value={username}
          onChange={(e) => setUsername(e.target.value)}
          autoComplete="username"
          required
        />
        <TextField
          label="Email"
          type="email"
          value={email}
          onChange={(e) => setEmail(e.target.value)}
          autoComplete="email"
          required
        />
        <TextField
          label="Password baru"
          type="password"
          value={password}
          onChange={(e) => setPassword(e.target.value)}
          autoComplete="new-password"
          hint="Kosongkan bila tidak ingin mengganti password (min 8 karakter)."
        />
        {canAssignRole && (
          <label className="block">
            <span className="mb-1.5 block text-sm font-medium text-zinc-700 dark:text-zinc-200">Role</span>
            <select
              value={role}
              onChange={(e) => setRole(e.target.value)}
              className="w-full rounded-xl border border-zinc-300 bg-white px-3.5 py-2.5 text-sm text-zinc-900 focus:border-violet-500 focus:outline-none focus:ring-2 focus:ring-violet-500/30 dark:border-zinc-700 dark:bg-zinc-900 dark:text-zinc-50"
            >
              {(roles.length > 0 ? roles : [{ code: user?.role_code || 'USER', name: '' }]).map((r) => (
                <option key={r.code} value={r.code}>
                  {r.code}{r.name ? ` — ${r.name}` : ''}
                </option>
              ))}
            </select>
            <span className="mt-1.5 block text-xs text-zinc-500 dark:text-zinc-400">
              Role dapat diganti dari menu User Account.
            </span>
          </label>
        )}
      </form>
    </Modal>
  )
}
