// ===== TEMPLATE SERAGAM app/* (ala <mod>.js langit_v2) =====
// Shell + controller standar; body custom (kartu user) + FRM items (./fields.js).
// Tabel DB: CPUSER via ./api.js listUsers/createUser/deleteUser/logoutAll.
import { Alert, Avatar, Badge, Button, ConfirmDialog, Icon, Modal, StandardForm, StandardPage, Tooltip, api, useStandardController, useState, useToast } from '../shared/all.js'
import { createUser, deleteUser, listUsers, logoutAll } from './api.js'
import { listRoles } from '../role_permission/api.js'
import { userCreateFields } from './fields.js'
import EditUserModal from './components/EditUserModal.jsx'

const FEATURES = { header: true, refresh: true, filter: false, tabs: false, create: true, edit: true, remove: true, pagination: true, empty: true, error: true, confirmDialog: true, extraActions: true }

const CONFIG = {
  title: 'Daftar Pengguna',
  description: 'Kelola akun yang terdaftar. Anda hanya dapat mengubah akun milik sendiri.',
  errorTitle: 'Gagal memuat data',
  emptyTitle: 'Belum ada pengguna',
  emptyDescription: 'Data kosong pada halaman ini. Coba kembali ke halaman sebelumnya atau muat ulang.',
  createTitle: 'Tambah user baru',
}

export const meta = { label: 'User Account', icon: 'users', order: 1 }

export default function UsersList({ onAccountDeleted }) {
  const toast = useToast()
  const me = api.currentUser()
  const canManage = true
  const [editing, setEditing] = useState(null)
  const [confirm, setConfirm] = useState(null) // { type: 'delete'|'logoutAll', user }
  const [acting, setActing] = useState(false)
  const [showCreate, setShowCreate] = useState(false)
  const [roles, setRoles] = useState([])
  const [cValues, setCValues] = useState({ username: '', email: '', password: '', role: 'USER' })
  const [cError, setCError] = useState('')
  const [creating, setCreating] = useState(false)

  const { rows: users, offset, setOffset, loading, showLoading, error, setError, load } =
    useStandardController(({ limit, offset }) => listUsers(limit, offset), {})

  function openCreate() {
    setCValues({ username: '', email: '', password: '', role: 'USER' })
    setCError('')
    listRoles().then(setRoles).catch(() => setRoles([]))
    setShowCreate(true)
  }

  async function create() {
    setCError('')
    const username = (cValues.username || '').trim()
    const email = (cValues.email || '').trim().toLowerCase()
    if (!username || !email || !cValues.password) {
      setCError('Username, email, dan password wajib diisi.')
      return
    }
    setCreating(true)
    try {
      const res = await createUser(username, email, cValues.password, cValues.role)
      toast.success(`Akun @${username} dibuat (${res.code}).`, { title: 'User dibuat' })
      setShowCreate(false)
      if (offset === 0) load()
      else setOffset(0)
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
        await deleteUser(confirm.user.code)
        toast.success(`Akun @${confirm.user.username} dihapus.`)
        if (me?.code === confirm.user.code) {
          onAccountDeleted?.()
          return
        }
      } else {
        await logoutAll(confirm.user.code)
        toast.success('Semua sesi berhasil dikeluarkan. Silakan login ulang bila diperlukan.')
      }
      setConfirm(null)
      load()
    } catch (err) {
      toast.error(err.message, { title: 'Aksi gagal' })
    } finally {
      setActing(false)
    }
  }

  return (
    <div className="flex flex-col gap-4">
      <StandardPage
        title={CONFIG.title}
        description={CONFIG.description}
        features={FEATURES}
        loading={loading}
        showLoading={showLoading}
        error={error}
        errorTitle={CONFIG.errorTitle}
        onClearError={() => setError('')}
        onRefresh={() => load()}
        onCreate={canManage ? openCreate : null}
        createLabel="Tambah User"
        items={users}
        emptyTitle={CONFIG.emptyTitle}
        emptyDescription={CONFIG.emptyDescription}
        emptyAction={
          <Button variant="secondary" size="sm" onClick={() => (offset === 0 ? load() : setOffset(0))}>
            Muat ulang
          </Button>
        }
        offset={offset}
        onPage={(next) => next >= 0 && setOffset(next)}
      >
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
                {FEATURES.edit && canManage && (
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
      </StandardPage>

      {FEATURES.edit && editing && (
        <EditUserModal
          key={editing.code}
          user={editing}
          onClose={() => setEditing(null)}
          onSaved={() => load()}
        />
      )}

      {FEATURES.confirmDialog && (
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
      )}

      {FEATURES.create && (
        <Modal
          open={showCreate}
          onClose={creating ? undefined : () => setShowCreate(false)}
          title={CONFIG.createTitle}
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
            <StandardForm fields={userCreateFields(roles)} values={cValues} onChange={setCValues} />
          </div>
        </Modal>
      )}
    </div>
  )
}
