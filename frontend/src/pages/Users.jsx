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
  Pagination,
  SkeletonRows,
  useToast,
} from '../components'
import EditUserModal from './EditUserModal.jsx'

const PAGE_SIZE = 10

export default function UsersList({ onAccountDeleted }) {
  const toast = useToast()
  const me = api.currentUser()
  const [users, setUsers] = useState([])
  const [offset, setOffset] = useState(0)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [editing, setEditing] = useState(null)
  const [confirm, setConfirm] = useState(null) // { type: 'delete'|'logoutAll', user }
  const [acting, setActing] = useState(false)

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
          <Button variant="secondary" size="sm" onClick={() => load(offset)} loading={loading}>
            <Icon name="refresh" className="h-4 w-4" />
            Muat ulang
          </Button>
        </div>

        {error && (
          <Alert tone="error" title="Gagal memuat data" closable onClose={() => setError('')} className="mb-4">
            {error}
          </Alert>
        )}

        {loading ? (
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
                  {isMe && (
                    <div className="flex shrink-0 items-center gap-1.5">
                      <button
                        type="button"
                        title="Edit profil"
                        aria-label={`Edit ${u.username}`}
                        onClick={() => setEditing(u)}
                        className="rounded-lg p-2 text-zinc-500 transition hover:bg-white hover:text-violet-600 hover:shadow-sm dark:hover:bg-zinc-800 dark:hover:text-violet-300"
                      >
                        <Icon name="pencil" className="h-4.5 w-4.5" />
                      </button>
                      <button
                        type="button"
                        title="Keluarkan semua sesi"
                        aria-label={`Keluarkan semua sesi ${u.username}`}
                        onClick={() => setConfirm({ type: 'logoutAll', user: u })}
                        className="rounded-lg p-2 text-zinc-500 transition hover:bg-white hover:text-amber-600 hover:shadow-sm dark:hover:bg-zinc-800 dark:hover:text-amber-300"
                      >
                        <Icon name="logout" className="h-4.5 w-4.5" />
                      </button>
                      <button
                        type="button"
                        title="Hapus akun"
                        aria-label={`Hapus ${u.username}`}
                        onClick={() => setConfirm({ type: 'delete', user: u })}
                        className="rounded-lg p-2 text-zinc-500 transition hover:bg-rose-50 hover:text-rose-600 hover:shadow-sm dark:hover:bg-rose-950 dark:hover:text-rose-300"
                      >
                        <Icon name="trash" className="h-4.5 w-4.5" />
                      </button>
                    </div>
                  )}
                </li>
              )
            })}
          </ul>
        )}

        {!loading && users.length > 0 && (
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
    </div>
  )
}
