import { useCallback, useEffect, useMemo, useState } from 'react'
import { createRole, deleteRole, getRole, listPermissions, listRoles, setRolePermissions } from './api.js'

import {
  Alert,
  Badge,
  Button,
  Card,
  CardTitle,
  ConfirmDialog,
  EmptyState,
  Icon,
  Modal,
  SkeletonRows,
  TextField,
  useSmoothLoading,
  useToast,
} from '../../../components'

export const meta = { label: 'Role & Permission', icon: 'shield', order: 3 }

function groupPermissions(perms) {
  const groups = {}
  for (const p of perms) {
    const g = `MODULE ${p.group || 'OTHER'}`
    if (!groups[g]) groups[g] = []
    groups[g].push(p)
  }
  return groups
}

// Daftar role di kiri
function RoleList({ roles, selected, showLoading, error, onCreate, onSelect, onErrorClose }) {
  return (
    <Card className="h-full flex flex-col">
      <div className="mb-4 flex flex-wrap items-center justify-between gap-3">
        <CardTitle description="Buat role, hapus role custom, dan pilih role untuk mengatur akses menu di kanan. Role ADMIN & USER bawaan tidak bisa dihapus. Menghapus role custom juga menghapus semua user yang memakainya beserta grant menunya.">
          Daftar Role
        </CardTitle>
        <Button size="sm" onClick={onCreate}>
          <Icon name="plus" className="h-4 w-4" />
          Role baru
        </Button>
      </div>

      {error && (
        <Alert tone="error" title="Gagal memuat" closable onClose={onErrorClose} className="mb-4">
          {error}
        </Alert>
      )}

      <div className="flex-1 overflow-y-auto">
        {showLoading ? (
          <SkeletonRows rows={5} />
        ) : roles.length === 0 ? (
          <EmptyState title="Belum ada role" description="Buat role pertama lewat tombol di atas." />
        ) : (
          <div className="flex flex-col gap-1.5">
            {roles.map((r) => (
              <button
                key={r.code}
                type="button"
                onClick={() => onSelect(r.code)}
                className={`flex items-center gap-3 w-full rounded-xl border px-3.5 py-2.5 text-sm font-semibold transition text-left ${
                  selected === r.code
                    ? 'border-violet-500 bg-violet-50 text-violet-700 dark:bg-violet-950 dark:text-violet-200'
                    : 'border-zinc-200 text-zinc-600 hover:border-zinc-300 hover:bg-zinc-50 dark:border-zinc-700 dark:text-zinc-300 dark:hover:bg-zinc-800/50'
                }`}
              >
                <Icon name="shield" className="h-5 w-5 shrink-0" />
                <div className="min-w-0 flex-1">
                  <p className="truncate font-medium">{r.code}</p>
                  <p className="truncate text-xs font-normal opacity-60">{r.name}</p>
                </div>
              </button>
            ))}
          </div>
        )}
      </div>
    </Card>
  )
}

// Matriks permission di kanan
function PermissionMatrix({ selected, checked, perms, grouped, loading, saving, onToggle, onToggleGroup, onSave, onDeleteRole, canDelete }) {
  if (!selected || loading) {
    return (
      <Card className="h-full">
        <div className="flex items-center justify-center h-full min-h-[300px]">
          <EmptyState title={selected ? 'Memuat permission…' : 'Pilih role'} description="Pilih role di sebelah kiri untuk melihat dan mengatur permission-nya." />
        </div>
      </Card>
    )
  }

  return (
    <Card className="h-full flex flex-col">
      <div className="mb-4 flex flex-wrap items-center justify-between gap-3">
        <CardTitle description="Centang menu yang boleh diakses role ini. Satu centang memberi akses ke semua fungsi menu tersebut.">
          Matriks akses menu — <span className="font-mono">{selected}</span>
        </CardTitle>
        <div className="flex gap-2">
          {canDelete && (
            <Button
              variant="danger"
              size="sm"
              onClick={onDeleteRole}
            >
              <Icon name="trash" className="h-4 w-4" />
              Hapus role
            </Button>
          )}
          <Button size="sm" onClick={onSave} loading={saving}>
            Simpan permission
          </Button>
        </div>
      </div>

      <div className="flex-1 overflow-y-auto">
        {Object.entries(grouped).map(([group, list]) => {
          const allOn = list.every((p) => checked.includes(p.code))
          return (
            <div key={group} className="mb-4 rounded-xl border border-zinc-100 dark:border-zinc-800">
              <div className="flex items-center justify-between gap-3 border-b border-zinc-100 px-4 py-2.5 dark:border-zinc-800">
                <p className="text-xs font-bold tracking-wide text-zinc-500 dark:text-zinc-400">{group}</p>
                <button
                  type="button"
                  onClick={() => onToggleGroup(list, !allOn)}
                  className="text-xs font-semibold text-violet-600 hover:underline dark:text-violet-300"
                >
                  {allOn ? 'Hapus semua' : 'Pilih semua'}
                </button>
              </div>
              <div className="grid gap-1 p-2 sm:grid-cols-2">
                {list.map((p) => {
                  const on = checked.includes(p.code)
                  return (
                    <label
                      key={p.code}
                      className={`flex cursor-pointer items-start gap-2.5 rounded-lg px-3 py-2 text-sm transition ${
                        on ? 'bg-violet-50 dark:bg-violet-950/50' : 'hover:bg-zinc-50 dark:hover:bg-zinc-800/60'
                      }`}
                    >
                      <input
                        type="checkbox"
                        checked={on}
                        onChange={() => onToggle(p.code)}
                        className="mt-1 h-4 w-4 accent-violet-600"
                      />
                      <span>
                        <span className="block font-medium text-zinc-800 dark:text-zinc-100">
                          {p.name} <Badge tone={on ? 'brand' : 'neutral'}>{p.code}</Badge>
                        </span>
                        {p.description && (
                          <span className="block text-xs text-zinc-500 dark:text-zinc-400">{p.description}</span>
                        )}
                      </span>
                    </label>
                  )
                })}
              </div>
            </div>
          )
        })}
        <p className="mt-2 text-xs text-zinc-500 dark:text-zinc-400 text-center">
          Tercatat {checked.length} dari {perms.length} permission.
        </p>
      </div>
    </Card>
  )
}

export default function Roles() {
  const toast = useToast()
  const [roles, setRoles] = useState([])
  const [perms, setPerms] = useState([])
  const [selected, setSelected] = useState(null) // role code
  const [checked, setChecked] = useState([])
  const [loading, setLoading] = useState(true)
  const showLoading = useSmoothLoading(loading)
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState('')
  const [showCreate, setShowCreate] = useState(false)
  const [newCode, setNewCode] = useState('')
  const [newName, setNewName] = useState('')
  const [creating, setCreating] = useState(false)
  const [deleting, setDeleting] = useState(null)

  const load = useCallback(async () => {
    setLoading(true)
    setError('')
    try {
      const [r, p] = await Promise.all([listRoles(), listPermissions()])
      const list = Array.isArray(r) ? r : []
      setRoles(list)
      setPerms(Array.isArray(p) ? p.filter((permission) => permission.code.startsWith('MENU_')) : [])
      // Auto-select pertama tanpa bergantung pada `selected` (hindari refetch loop).
      setSelected((prev) => (prev ? prev : list.length > 0 ? list[0].code : null))
    } catch (err) {
      setError(err.message)
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    load()
  }, [load])

  useEffect(() => {
    if (!selected) return
    getRole(selected)
      .then((d) => setChecked((d.permissions || []).filter((code) => code.startsWith('MENU_'))))
      .catch(() => setChecked([]))
  }, [selected])

  const grouped = useMemo(() => groupPermissions(perms), [perms])

  async function save() {
    setSaving(true)
    try {
      await setRolePermissions(selected, checked)
      toast.success(`Permission role ${selected} diperbarui (${checked.length} item).`)
    } catch (err) {
      toast.error(err.message, { title: 'Gagal menyimpan' })
    } finally {
      setSaving(false)
    }
  }

  function toggle(code) {
    setChecked((prev) => (prev.includes(code) ? prev.filter((c) => c !== code) : [...prev, code]))
  }

  function toggleGroup(list, on) {
    setChecked((prev) => {
      const set = new Set(prev)
      for (const p of list) {
        if (on) set.add(p.code)
        else set.delete(p.code)
      }
      return [...set]
    })
  }

  async function create() {
    if (!newCode.trim() || !newName.trim()) {
      toast.warning('Kode dan nama role wajib diisi.')
      return
    }
    setCreating(true)
    try {
      await createRole(newCode.trim(), newName.trim())
      toast.success(`Role ${newCode.trim().toUpperCase()} dibuat. Atur permission-nya di matriks.`)
      setNewCode('')
      setNewName('')
      setShowCreate(false)
      setSelected(newCode.trim().toUpperCase())
      load()
    } catch (err) {
      toast.error(err.message, { title: 'Gagal membuat role' })
    } finally {
      setCreating(false)
    }
  }

  async function remove() {
    if (!deleting) return
    try {
      const res = await deleteRole(deleting.code)
      const affected = res?.deleted_users || 0
      toast.success(
        affected > 0
          ? `Role ${deleting.code} dihapus, ${affected} user ikut terhapus.`
          : `Role ${deleting.code} dihapus.`
      )
      setDeleting(null)
      setSelected(null)
      load()
    } catch (err) {
      toast.error(err.message, { title: 'Gagal menghapus' })
    }
  }


  return (
    <div className="flex flex-col gap-4">
      <div className="grid gap-4 lg:grid-cols-[280px_1fr]">
        <RoleList
          roles={roles}
          selected={selected}
          showLoading={showLoading}
          error={error}
          onCreate={() => setShowCreate(true)}
          onSelect={setSelected}
          onErrorClose={() => setError('')}
        />
        <PermissionMatrix
          selected={selected}
          checked={checked}
          perms={perms}
          grouped={grouped}
          loading={showLoading}
          saving={saving}
          onToggle={toggle}
          onToggleGroup={toggleGroup}
          onSave={save}
          onDeleteRole={() => setDeleting(roles.find((r) => r.code === selected))}
          canDelete={!!selected && !['ADMIN', 'USER'].includes(selected)}
        />
      </div>

      <Modal
        open={showCreate}
        onClose={creating ? undefined : () => setShowCreate(false)}
        title="Buat role baru"
        size="sm"
        footer={
          <>
            <Button variant="secondary" onClick={() => setShowCreate(false)} disabled={creating}>
              Batal
            </Button>
            <Button onClick={create} loading={creating}>
              Buat role
            </Button>
          </>
        }
      >
        <div className="flex flex-col gap-4">
          <TextField
            label="Kode role"
            value={newCode}
            onChange={(e) => setNewCode(e.target.value.toUpperCase())}
            placeholder="mis. EDITOR"
            hint="Huruf besar, angka, underscore, maks 20 karakter."
          />
          <TextField
            label="Nama tampil"
            value={newName}
            onChange={(e) => setNewName(e.target.value)}
            placeholder="mis. Editor Konten"
          />
        </div>
      </Modal>

      <ConfirmDialog
        open={!!deleting}
        title={`Hapus role ${deleting?.code}?`}
        message="Role, semua user yang memakainya, dan seluruh grant menu role ini akan dihapus permanen. Tindakan ini tidak bisa dibatalkan."
        confirmLabel="Ya, hapus"
        danger
        onConfirm={remove}
        onCancel={() => setDeleting(null)}
      />

    </div>
  )
}