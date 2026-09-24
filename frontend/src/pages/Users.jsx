import { useCallback, useEffect, useState } from 'react'
import { api } from '../api/client.js'
import {
  Alert,
  Avatar,
  Badge,
  Button,
  Card,
  CardTitle,
  ConfirmDialog,
  EmptyState,
  Icon,
  Modal,
  Pagination,
  PasswordInput,
  SkeletonRows,
  TextField,
  Tooltip,
  useSmoothLoading,
  useToast,
} from '../components'
import EditUserModal from './EditUserModal.jsx'

const PAGE_SIZE = 10

export default function UsersList({ onAccountDeleted }) {
  const toast = useToast()
  const me = api.currentUser()
  const isAdmin = me?.role === 'ADMIN'
  const [users, setUsers] = useState([])
  const [offset, setOffset] = useState(0)
  const [loading, setLoading] = useState(true)
  const showLoading = useSmoothLoading(loading)
  const [error, setError] = useState('')
  const [editing, setEditing] = useState(null)
  const [confirm, setConfirm] = useState(null) // { type: 'delete'|'logoutAll', user }
  const [acting, setActing] = useState(false)
  const [showCreate, setShowCreate] = useState(false)
  const [roles, setRoles] = useState([])
  const [cUsername, setCUsername] = useState('')
  const [cEmail, setCEmail] = useState('')
  const [cPassword, setCPassword] = useState('')
  const [cRole, setCRole] = useState('USER')
  const [cError, setCError] = useState('')
  const [creating, setCreating] = useState(false)

  const load = useCallback(async (nextOffset) => {
    setLoading(true)
    setError('')
    try {
      setUsers(await api.listUsers(PAGE_SIZE, nextOffset))
    } catch (err) {
      setError(err.message)
    } finally {
      setLoading(false)
    }
  }, [])

  // Fetch data dari REST API saat halaman berubah (sinkronisasi ke
  // sistem eksternal — kasus sah untuk useEffect).
  useEffect(() => {
    load(offset)
  }, [offset, load])

  function changePage(next) {
    if (next < 0) return
    setOffset(next)
  }

  function openCreate() {
    setCUsername('')
    setCEmail('')
    setCPassword('')
    setCRole('USER')
    setCError('')
    api.listRoles().then(setRoles).catch(() => setRoles([]))
    setShowCreate(true)
  }

  async function create() {
    setCError('')
    if (!cUsername.trim() || !cEmail.trim() || !cPassword) {
      setCError('Username, email, dan password wajib diisi.')
      return
    }
    setCreating(true)
    try {
      const res = await api.createUser(cUsername.trim(), cEmail.trim().toLowerCase(), cPassword, cRole)
      toast.success(`Akun @${cUsername.trim()} dibuat (${res.code}).`, { title: 'User dibuat' })
      setShowCreate(false)
      setOffset(0)
      load(0)
    } catch (err) {
      setCError(err.message)
    } finally {
      setCreating(false)
    }
  }

  async function runConfirm() {
    if (!confirm) return
    setActing(true)
    try {
      if (confirm.type === 'delete') {
        await api.deleteUser(confirm.user.code)
        toast.success(`Akun @${confirm.user.username} dihapus.`)
        if (me?.code === confirm.user.code) {
          onAccountDeleted()
          return
        }
      } else {
        await api.logoutAll(confirm.user.code)
        toast.success('Semua sesi berhasil dikeluarkan. Silakan login kembali bila diperlukan.')
      }
      setConfirm(null)
      load(offset)
    } catch (err) {
      toast.error(err.message, { title: 'Aksi gagal' })
    } finally {
      setActing(false)
    }
  }

  return (
    <div className="flex flex-col gap-4">
      <Card>
        <div className="mb-4 flex flex-wrap items-center justify-between gap-3">
          <CardTitle description="Kelola akun yang terdaftar. Anda hanya dapat mengubah akun milik sendiri.">
            Daftar Pengguna
          </CardTitle>
          <div className="flex gap-2">
            {isAdmin && (
              <Button size="sm" onClick={openCreate}>
                <Icon name="plus" className="h-4 w-4" />
                Tambah User
              </Button>
            )}
            <Button variant="secondary" size="sm" onClick={() => load(offset)} loading={loading}>
              <Icon name="refresh" className="h-4 w-4" />
              Muat ulang
            </Button>
          </div>
        </div>

        {error && (
          <Alert tone="error" title="Gagal memuat data" closable onClose={() => setError('')} className="mb-4">
            {error}
          </Alert>
        )}

        {showLoading ? (
          <SkeletonRows rows={4} />
        ) : users.length === 0 ? (
          <EmptyState
            title="Belum ada pengguna"
            description="Data kosong pada halaman ini. Coba kembali ke halaman sebelumnya atau muat ulang."
            action={
              <Button variant="secondary" size="sm" onClick={() => (offset === 0 ? load(0) : setOffset(0))}>
                Muat ulang
              </Button>
            }
          />
        ) : (
          <ul className="flex flex-col gap-2.5">
            {users.map((u) => {
              const isMe = me?.code === u.code
              // Admin boleh kelola semua akun; user biasa hanya akun sendiri
              // (backend tetap menegakkan via requireSelfOrPerm).
              const canAct = isMe || isAdmin
              return (
                <li
                  key={u.code}
                  className="flex items-center gap-3 rounded-xl border border-zinc-100 bg-zinc-50/60 p-3 transition hover:border-violet-200 hover:bg-violet-50/50 dark:border-zinc-800 dark:bg-zinc-900/40 dark:hover:border-violet-800"
                >
                  <Avatar name={u.username} />
                  <div className="min-w-0 flex-1">
                    <p className="flex flex-wrap items-center gap-2 text-sm font-semibold text-zinc-900 dark:text-zinc-50">
                      <span className="truncate">@{u.username}</span>
                      {isMe && <Badge tone="brand">Anda</Badge>}
                      {u.role_code && (
                        <Badge tone={u.role_code === 'ADMIN' ? 'danger' : 'neutral'}>{u.role_code}</Badge>
                      )}
                    </p>
                    <p className="truncate text-xs text-zinc-500 dark:text-zinc-400">
                      <span className="font-mono">{u.code}</span> · {u.email}
                    </p>
                  </div>
                  {canAct && (
                    <div className="flex shrink-0 items-center gap-1.5">
                      <Tooltip label="Edit profil">
                        <button
                          type="button"
                          aria-label={`Edit ${u.username}`}
                          onClick={() => setEditing(u)}
                          className="rounded-lg p-2 text-zinc-500 transition-colors hover:bg-white hover:text-zinc-900 hover:shadow-sm dark:hover:bg-zinc-800 dark:hover:text-zinc-100"
                        >
                          <Icon name="pencil" className="h-4 w-4" />
                        </button>
                      </Tooltip>
                      <Tooltip label="Keluarkan semua sesi">
                        <button
                          type="button"
                          aria-label={`Keluarkan semua sesi ${u.username}`}
                          onClick={() => setConfirm({ type: 'logoutAll', user: u })}
                          className="rounded-lg p-2 text-zinc-500 transition-colors hover:bg-white hover:text-zinc-900 hover:shadow-sm dark:hover:bg-zinc-800 dark:hover:text-zinc-100"
                        >
                          <Icon name="logout" className="h-4 w-4" />
                        </button>
                      </Tooltip>
                      <Tooltip label="Hapus akun">
                        <button
                          type="button"
                          aria-label={`Hapus ${u.username}`}
                          onClick={() => setConfirm({ type: 'delete', user: u })}
                          className="rounded-lg p-2 text-zinc-500 transition-colors hover:bg-rose-50 hover:text-rose-600 hover:shadow-sm dark:hover:bg-rose-950 dark:hover:text-rose-300"
                        >
                          <Icon name="trash" className="h-4 w-4" />
                        </button>
                      </Tooltip>
                    </div>
                  )}
                </li>
              )
            })}
          </ul>
        )}

        {!showLoading && users.length > 0 && (
          <div className="mt-4 border-t border-zinc-100 pt-4 dark:border-zinc-800">
            <Pagination
              offset={offset}
              limit={PAGE_SIZE}
              count={users.length}
              hasMore={users.length === PAGE_SIZE}
              loading={loading}
              onPage={changePage}
            />
          </div>
        )}
      </Card>

      {editing && (
        <EditUserModal
          key={editing.code}
          user={editing}
          onClose={() => setEditing(null)}
          onSaved={() => load(offset)}
        />
      )}

      <ConfirmDialog
        open={!!confirm}
        title={confirm?.type === 'delete' ? 'Hapus akun ini?' : 'Keluarkan semua sesi?'}
        message={
          confirm?.type === 'delete'
            ? `Akun @${confirm?.user.username} beserta seluruh sesinya akan dihapus permanen dan tidak bisa dikembalikan.`
            : `Semua perangkat yang login sebagai @${confirm?.user.username} akan dikeluarkan dan harus login ulang.`
        }
        confirmLabel={confirm?.type === 'delete' ? 'Ya, hapus' : 'Ya, keluarkan'}
        danger={confirm?.type === 'delete'}
        loading={acting}
        onConfirm={runConfirm}
        onCancel={() => !acting && setConfirm(null)}
      />

      <Modal
        open={showCreate}
        onClose={creating ? undefined : () => setShowCreate(false)}
        title="Tambah user baru"
        size="sm"
        closeOnBackdrop={!creating}
        footer={
          <>
            <Button variant="secondary" onClick={() => setShowCreate(false)} disabled={creating}>
              Batal
            </Button>
            <Button onClick={create} loading={creating}>
              Buat akun
            </Button>
          </>
        }
      >
        <div className="flex flex-col gap-4">
          {cError && (
            <Alert tone="error" closable>
              {cError}
            </Alert>
          )}
          <TextField
            label="Username"
            value={cUsername}
            onChange={(e) => setCUsername(e.target.value)}
            autoComplete="off"
            placeholder="min. 3 karakter"
          />
          <TextField
            label="Email"
            type="email"
            value={cEmail}
            onChange={(e) => setCEmail(e.target.value)}
            autoComplete="off"
            placeholder="nama@email.com"
          />
          <PasswordInput
            label="Password awal"
            value={cPassword}
            onChange={(e) => setCPassword(e.target.value)}
            autoComplete="new-password"
            placeholder="min. 8 karakter"
            hint="Beritahu password ini ke user; ia bisa menggantinya via Edit profil."
          />
          <label className="block">
            <span className="mb-1.5 block text-sm font-medium text-zinc-700 dark:text-zinc-200">Role awal</span>
            <select
              value={cRole}
              onChange={(e) => setCRole(e.target.value)}
              className="w-full rounded-xl border border-zinc-300 bg-white px-3.5 py-2.5 text-sm text-zinc-900 focus:border-violet-500 focus:outline-none focus:ring-2 focus:ring-violet-500/30 dark:border-zinc-700 dark:bg-zinc-900 dark:text-zinc-50"
            >
              {(roles.length > 0 ? roles : [{ code: 'USER', name: '' }]).map((r) => (
                <option key={r.code} value={r.code}>
                  {r.code}{r.name ? ` — ${r.name}` : ''}
                </option>
              ))}
            </select>
          </label>
        </div>
      </Modal>
    </div>
  )
}
