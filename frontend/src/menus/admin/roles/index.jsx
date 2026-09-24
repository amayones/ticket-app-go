import { useCallback, useEffect, useMemo, useState } from 'react'
import { api } from '../api/client.js'
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
} from '../components'

function groupPermissions(perms) {
  const groups = {}
  for (const p of perms) {
    const g = p.group || 'LAINNYA'
    if (!groups[g]) groups[g] = []
    groups[g].push(p)
  }
  return groups
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
      const [r, p] = await Promise.all([api.listRoles(), api.listPermissions()])
      setRoles(Array.isArray(r) ? r : [])
      setPerms(Array.isArray(p) ? p : [])
      if (!selected && r.length > 0) setSelected(r[0].code)
    } catch (err) {
      setError(err.message)
    } finally {
      setLoading(false)
    }
  }, [selected])

  useEffect(() => {
    load()
  }, [load])

  useEffect(() => {
    if (!selected) return
    api
      .getRole(selected)
      .then((d) => setChecked(d.permissions || []))
      .catch(() => setChecked([]))
  }, [selected])

  const grouped = useMemo(() => groupPermissions(perms), [perms])

  async function save() {
    setSaving(true)
    try {
      await api.setRolePermissions(selected, checked)
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
      await api.createRole(newCode.trim(), newName.trim())
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
      await api.deleteRole(deleting.code)
      toast.success(`Role ${deleting.code} dihapus.`)
      setDeleting(null)
      setSelected(null)
      load()
    } catch (err) {
      toast.error(err.message, { title: 'Gagal menghapus' })
    }
  }

  return (
    <div className="flex flex-col gap-4">
      <Card>
        <div className="mb-4 flex flex-wrap items-center justify-between gap-3">
          <CardTitle description="Buat role, hapus role custom, dan centang permission tiap role. Role ADMIN & USER bawaan tidak bisa dihapus.">
            Role & Permission (RBAC)
          </CardTitle>
          <Button size="sm" onClick={() => setShowCreate(true)}>
            <Icon name="plus" className="h-4 w-4" />
            Role baru
          </Button>
        </div>

        {error && (
          <Alert tone="error" title="Gagal memuat" closable onClose={() => setError('')} className="mb-4">
            {error}
          </Alert>
        )}

        {showLoading ? (
          <SkeletonRows rows={3} />
        ) : roles.length === 0 ? (
          <EmptyState title="Belum ada role" description="Buat role pertama lewat tombol di atas." />
        ) : (
          <div className="flex flex-wrap gap-2">
            {roles.map((r) => (
              <button
                key={r.code}
                type="button"
                onClick={() => setSelected(r.code)}
                className={`inline-flex items-center gap-2 rounded-xl border px-3.5 py-2 text-sm font-semibold transition ${
                  selected === r.code
                    ? 'border-violet-500 bg-violet-50 text-violet-700 dark:bg-violet-950 dark:text-violet-200'
                    : 'border-zinc-200 text-zinc-600 hover:border-zinc-300 dark:border-zinc-700 dark:text-zinc-300'
                }`}
              >
                <Icon name="shield" className="h-4 w-4" />
                {r.code}
                <span className="font-normal opacity-60">{r.name}</span>
              </button>
            ))}
          </div>
        )}
      </Card>

      {selected && !showLoading && (
        <Card>
          <div className="mb-4 flex flex-wrap items-center justify-between gap-3">
            <CardTitle description="Centang permission lalu simpan. ADMIN selalu lolos semua permission tanpa perlu dicentang.">
              Matriks permission — <span className="font-mono">{selected}</span>
            </CardTitle>
            <div className="flex gap-2">
              {!['ADMIN', 'USER'].includes(selected) && (
                <Button
                  variant="danger"
                  size="sm"
                  onClick={() => setDeleting(roles.find((r) => r.code === selected))}
                >
                  <Icon name="trash" className="h-4 w-4" />
                  Hapus role
                </Button>
              )}
              <Button size="sm" onClick={save} loading={saving}>
                Simpan permission
              </Button>
            </div>
          </div>
          {Object.entries(grouped).map(([group, list]) => {
            const allOn = list.every((p) => checked.includes(p.code))
            return (
              <div key={group} className="mb-4 rounded-xl border border-zinc-100 dark:border-zinc-800">
                <div className="flex items-center justify-between gap-3 border-b border-zinc-100 px-4 py-2.5 dark:border-zinc-800">
                  <p className="text-xs font-bold tracking-wide text-zinc-500 dark:text-zinc-400">{group}</p>
                  <button
                    type="button"
                    onClick={() => toggleGroup(list, !allOn)}
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
                          onChange={() => toggle(p.code)}
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
          <p className="text-xs text-zinc-500 dark:text-zinc-400">
            Tercatat {checked.length} dari {perms.length} permission.
          </p>
        </Card>
      )}

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
        message="Role akan dihapus permanen. Gagal bila masih dipakai user — pindahkan dulu user-nya ke role lain."
        confirmLabel="Ya, hapus"
        danger
        onConfirm={remove}
        onCancel={() => setDeleting(null)}
      />
    </div>
  )
}
