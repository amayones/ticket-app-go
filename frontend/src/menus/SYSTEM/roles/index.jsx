import { useCallback, useEffect, useMemo, useState } from 'react'
import { createMenu, createRole, deleteMenu, deleteRole, getRole, listMenus, listPermissions, listRoles, setRolePermissions } from './api.js'

export const meta = { label: 'Role & Permission', icon: 'shield', order: 2 }
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
  Tooltip,
  useSmoothLoading,
  useToast,
} from '../../../components'

function groupPermissions(perms) {
  const groups = {}
  for (const p of perms) {
    const g = `MODULE ${p.group || 'OTHER'}`
    if (!groups[g]) groups[g] = []
    groups[g].push(p)
  }
  return groups
}

function groupMenusByModule(menus) {
  const groups = {}
  for (const m of menus) {
    const g = m.mcontrol || 'SYSTEM'
    if (!groups[g]) groups[g] = []
    groups[g].push(m)
  }
  for (const g of Object.keys(groups)) {
    groups[g].sort((a, b) => (a.sort_order ?? 99) - (b.sort_order ?? 99) || a.code.localeCompare(b.code))
  }
  return groups
}

// Daftar role di kiri
function RoleList({ roles, selected, showLoading, error, onCreate, onSelect, onErrorClose }) {
  return (
    <Card className="h-full flex flex-col">
      <div className="mb-4 flex flex-wrap items-center justify-between gap-3">
        <CardTitle description="Buat role, hapus role custom, dan pilih role untuk mengatur akses menu di kanan. Role ADMIN & USER bawaan tidak bisa dihapus.">
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
  const [menus, setMenus] = useState([])
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
  // Registry menu (CPMENU).
  const [showMenuCreate, setShowMenuCreate] = useState(false)
  const [mCode, setMCode] = useState('')
  const [mName, setMName] = useState('')
  const [mModule, setMModule] = useState('SYSTEM')
  const [mLabel, setMLabel] = useState('')
  const [mSort, setMSort] = useState('99')
  const [mParent, setMParent] = useState('')
  const [creatingMenu, setCreatingMenu] = useState(false)
  const [deletingMenu, setDeletingMenu] = useState(null)

  const load = useCallback(async () => {
    setLoading(true)
    setError('')
    try {
      const [r, p, m] = await Promise.all([listRoles(), listPermissions(), listMenus()])
      const list = Array.isArray(r) ? r : []
      setRoles(list)
      setPerms(Array.isArray(p) ? p.filter((permission) => permission.code.startsWith('MENU_')) : [])
      setMenus(Array.isArray(m) ? m : [])
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
      await deleteRole(deleting.code)
      toast.success(`Role ${deleting.code} dihapus.`)
      setDeleting(null)
      setSelected(null)
      load()
    } catch (err) {
      toast.error(err.message, { title: 'Gagal menghapus' })
    }
  }

  const menuGroups = useMemo(() => groupMenusByModule(menus), [menus])

  async function createMenuItem() {
    if (!mCode.trim() || !mName.trim() || !mModule.trim() || !mLabel.trim()) {
      toast.warning('Kode, nama, modul, dan label menu wajib diisi.')
      return
    }
    setCreatingMenu(true)
    try {
      const menu = await createMenu({
        code: mCode.trim(),
        name: mName.trim(),
        module: mModule.trim(),
        label: mLabel.trim(),
        sort_order: Number(mSort) || 99,
        parent_code: mParent.trim() || undefined,
      })
      toast.success(`Menu ${menu.code} dibuat di modul ${menu.mcontrol}. Centang role yang boleh akses, lalu buat foldernya.`)
      setMCode('')
      setMName('')
      setMLabel('')
      setMSort('99')
      setMParent('')
      setShowMenuCreate(false)
      load()
    } catch (err) {
      toast.error(err.message, { title: 'Gagal membuat menu' })
    } finally {
      setCreatingMenu(false)
    }
  }

  async function removeMenu() {
    if (!deletingMenu) return
    try {
      await deleteMenu(deletingMenu.code)
      toast.success(`Menu ${deletingMenu.code} dihapus.`)
      setDeletingMenu(null)
      load()
    } catch (err) {
      toast.error(err.message, { title: 'Gagal menghapus menu' })
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

      <Card>
        <div className="mb-4 flex flex-wrap items-center justify-between gap-3">
          <CardTitle description="Satu baris = satu menu (CPMENU). Buat menu dulu di sini (tanpa auto-grant), centang role di matriks, lalu buat foldernya — menu yang belum punya folder tampil sebagai halaman 404 pemandu.">
            Registry Menu
          </CardTitle>
          <Button size="sm" onClick={() => setShowMenuCreate(true)}>
            <Icon name="plus" className="h-4 w-4" />
            Menu baru
          </Button>
        </div>

        {showLoading ? (
          <SkeletonRows rows={3} />
        ) : menus.length === 0 ? (
          <EmptyState title="Belum ada menu" description="Buat menu pertama lewat tombol di atas." />
        ) : (
          Object.entries(menuGroups).map(([module, list]) => (
            <div key={module} className="mb-4 rounded-xl border border-zinc-100 dark:border-zinc-800">
              <div className="border-b border-zinc-100 px-4 py-2.5 dark:border-zinc-800">
                <p className="text-xs font-bold tracking-wide text-zinc-500 dark:text-zinc-400">
                  MODULE {module} <span className="font-normal">· folder menus/{module}/</span>
                </p>
              </div>
              <ul className="flex flex-col gap-1 p-2">
                {list.map((m) => (
                  <li
                    key={m.code}
                    className="flex items-center gap-3 rounded-lg px-3 py-2 text-sm hover:bg-zinc-50 dark:hover:bg-zinc-800/60"
                  >
                    <div className="min-w-0 flex-1">
                      <p className="flex flex-wrap items-center gap-2 font-medium text-zinc-800 dark:text-zinc-100">
                        {m.label} <Badge tone="neutral">{m.code}</Badge>
                        {m.parent_code && <Badge>parent: {m.parent_code}</Badge>}
                      </p>
                      <p className="text-xs text-zinc-500 dark:text-zinc-400">urutan {m.sort_order ?? 99}</p>
                    </div>
                    <Tooltip label="Hapus menu">
                      <button
                        type="button"
                        aria-label={`Hapus ${m.code}`}
                        onClick={() => setDeletingMenu(m)}
                        className="shrink-0 rounded-lg p-2 text-zinc-500 transition-colors hover:bg-white hover:text-rose-600 hover:shadow-sm dark:hover:bg-zinc-800"
                      >
                        <Icon name="trash" className="h-4 w-4" />
                      </button>
                    </Tooltip>
                  </li>
                ))}
              </ul>
            </div>
          ))
        )}
      </Card>

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

      <Modal
        open={showMenuCreate}
        onClose={creatingMenu ? undefined : () => setShowMenuCreate(false)}
        title="Buat menu baru"
        footer={
          <>
            <Button variant="secondary" onClick={() => setShowMenuCreate(false)} disabled={creatingMenu}>
              Batal
            </Button>
            <Button onClick={createMenuItem} loading={creatingMenu}>
              Buat menu
            </Button>
          </>
        }
      >
        <div className="flex flex-col gap-4">
          <TextField
            label="Kode permission"
            value={mCode}
            onChange={(e) => setMCode(e.target.value.toUpperCase())}
            placeholder="mis. MENU_LAPORAN"
            hint="Wajib prefix MENU_, huruf besar/angka/underscore, maks 40 karakter."
          />
          <TextField
            label="Nama permission"
            value={mName}
            onChange={(e) => setMName(e.target.value)}
            placeholder="mis. Akses menu Laporan"
          />
          <TextField
            label="Modul (MCONTROL)"
            value={mModule}
            onChange={(e) => setMModule(e.target.value.toUpperCase())}
            placeholder="mis. REPORT"
            hint="UPPERCASE persis = nama folder menus/<MODUL>/. Modul baru otomatis jadi grup sidebar."
          />
          <TextField
            label="Label tampil"
            value={mLabel}
            onChange={(e) => setMLabel(e.target.value)}
            placeholder="mis. Laporan"
          />
          <div className="grid grid-cols-2 gap-4">
            <TextField
              label="Urutan"
              value={mSort}
              onChange={(e) => setMSort(e.target.value)}
              placeholder="99"
              hint="0–9999."
            />
            <TextField
              label="Parent (opsional)"
              value={mParent}
              onChange={(e) => setMParent(e.target.value.toUpperCase())}
              placeholder="mis. MENU_KEUANGAN"
              hint="Kode menu induk se-modul."
            />
          </div>
          <Alert tone="info">
            Menu baru tidak otomatis diberi ke role mana pun. Centang manual di matriks, lalu buat
            foldernya — halaman 404 pemandu menunjukkan path persisnya.
          </Alert>
        </div>
      </Modal>

      <ConfirmDialog
        open={!!deletingMenu}
        title={`Hapus menu ${deletingMenu?.code}?`}
        message="Menu + permission-nya dihapus permanen. Gagal bila masih dipakai role atau masih punya menu anak."
        confirmLabel="Ya, hapus"
        danger
        onConfirm={removeMenu}
        onCancel={() => setDeletingMenu(null)}
      />
    </div>
  )
}