import { useState } from 'react'
import { api } from '../api/client.js'
import { Alert, Button, Modal, TextField, useToast } from '../components'

// Modal edit user: hanya mengirim field yang diubah (patch parsial).
// Dipasang dengan key={user.id} oleh induk agar form ter-reset tiap ganti user.
export default function EditUserModal({ user, onClose, onSaved }) {
  const toast = useToast()
  const [username, setUsername] = useState(user?.username || '')
  const [email, setEmail] = useState(user?.email || '')
  const [password, setPassword] = useState('')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)

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
    if (Object.keys(patch).length === 0) {
      setError('Tidak ada perubahan. Ubah salah satu field terlebih dulu.')
      return
    }
    setLoading(true)
    try {
      await api.updateUser(user.id, patch)
      toast.success('Profil berhasil diperbarui.')
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
      </form>
    </Modal>
  )
}
