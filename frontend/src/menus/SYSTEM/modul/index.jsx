import { useCallback, useEffect, useMemo, useState } from 'react'
import { createMenu, createModule, deleteMenu, deleteModule, listMenus, listModules } from './api.js'
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

export const meta = { label: 'Modul & Menu', icon: 'list', order: 8 }

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

// Halaman Modul & Menu (master CPMATRIX + registry CPMENU).
// Urut kerja: buat modul dulu, lalu menu di dalamnya, lalu centang role
// di halaman Role. Tanpa auto-grant ke role mana pun. Menu tanpa folder
// tampil sebagai halaman 404 pemandu (bukan hilang diam-diam).
export default function Modul() {
  const toast = useToast()
  const [menus, setMenus] = useState([])
  const [modules, setModules] = useState([])
  const [loading, setLoading] = useState(true)
  const showLoading = useSmoothLoading(loading)
  const [error, setError] = useState('')
  const [showMenuCreate, setShowMenuCreate] = useState(false)
  const [mCode, setMCode] = useState('')
  const [mName, setMName] = useState('')
  const [mModule, setMModule] = useState('SYSTEM')
  const [mLabel, setMLabel] = useState('')
  const [mSort, setMSort] = useState('99')
  const [mParent, setMParent] = useState('')
  const [creatingMenu, setCreatingMenu] = useState(false)
  const [deletingMenu, setDeletingMenu] = useState(null)
  const [showModuleCreate, setShowModuleCreate] = useState(false)
  const [modCode, setModCode] = useState('')
  const [modLabel, setModLabel] = useState('')
  const [modSort, setModSort] = useState('10')
  const [creatingModule, setCreatingModule] = useState(false)
  const [deletingModule, setDeletingModule] = useState(null)

  const load = useCallback(async () => {
    setLoading(true)
    setError('')
    try {
      const [m, mods] = await Promise.all([listMenus(), listModules()])
      setMenus(Array.isArray(m) ? m : [])
      setModules(Array.isArray(mods) ? mods : [])
    } catch (err) {
      setError(err.message)
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    load()
  }, [load])

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

  async function createModuleItem() {
    if (!modCode.trim() || !modLabel.trim()) {
      toast.warning('Kode dan label modul wajib diisi.')
      return
    }
    setCreatingModule(true)
    try {
      const mod = await createModule({
        code: modCode.trim(),
        label: modLabel.trim(),
        sort_order: Number(modSort) || 99,
      })
      toast.success(`Modul ${mod.code} dibuat. Lanjut buat menu di dalamnya.`)
      setModCode('')
      setModLabel('')
      setModSort('10')
      setShowModuleCreate(false)
      load()
    } catch (err) {
      toast.error(err.message, { title: 'Gagal membuat modul' })
    } finally {
      setCreatingModule(false)
    }
  }

  async function removeModule() {
    if (!deletingModule) return
    try {
      await deleteModule(deletingModule.code)
      toast.success(`Modul ${deletingModule.code} dihapus.`)
      setDeletingModule(null)
      load()
    } catch (err) {
      toast.error(err.message, { title: 'Gagal menghapus modul' })
    }
  }

  return (
    <div className="flex flex-col gap-4">
      <Card>
        <div className="mb-4 flex flex-wrap items-center justify-between gap-3">
          <CardTitle description="Urut: buat modul dulu, lalu menu di dalamnya, lalu centang role di halaman Role. Menu tanpa folder tampil sebagai halaman 404 pemandu.">
            Modul &amp; Menu
          </CardTitle>
          <div className="flex gap-2">
            <Button variant="secondary" size="sm" onClick={() => setShowModuleCreate(true)}>
              <Icon name="plus" className="h-4 w-4" />
              Modul baru
            </Button>
            <Button size="sm" onClick={() => setShowMenuCreate(true)}>
              <Icon name="plus" className="h-4 w-4" />
              Menu baru
            </Button>
          </div>
        </div>

        {error && (
          <Alert tone="error" title="Gagal memuat" closable onClose={() => setError('')} className="mb-4">
            {error}
          </Alert>
        )}

        {showLoading ? (
          <SkeletonRows rows={3} />
        ) : (
          <>
            {modules.length > 0 && (
              <div className="mb-4 flex flex-wrap gap-2">
                {modules.map((mod) => (
                  <span
                    key={mod.code}
                    className="inline-flex items-center gap-1.5 rounded-full border border-zinc-200 px-3 py-1 text-xs font-semibold text-zinc-600 dark:border-zinc-700 dark:text-zinc-300"
                  >
                    {mod.code}
                    <button
                      type="button"
                      aria-label={`Hapus modul ${mod.code}`}
                      title={menuGroups[mod.code]?.length ? 'Gagal bila masih ada menu di dalamnya' : `Hapus modul ${mod.code}`}
                      onClick={() => setDeletingModule(mod)}
                      className="rounded-full p-0.5 text-zinc-400 transition hover:bg-rose-50 hover:text-rose-600 dark:hover:bg-rose-950"
                    >
                      <Icon name="x" className="h-3.5 w-3.5" />
                    </button>
                  </span>
                ))}
              </div>
            )}
            {menus.length === 0 ? (
              <EmptyState title="Belum ada menu" description="Buat modul dulu bila perlu, lalu menu pertama lewat tombol di atas." />
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
          </>
        )}
      </Card>

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
          <label className="block">
            <span className="mb-1.5 block text-sm font-medium text-zinc-700 dark:text-zinc-200">Modul (MCONTROL)</span>
            <select
              value={mModule}
              onChange={(e) => setMModule(e.target.value)}
              className="w-full rounded-xl border border-zinc-300 bg-white px-3.5 py-2.5 text-sm text-zinc-900 focus:border-violet-500 focus:outline-none focus:ring-2 focus:ring-violet-500/30 dark:border-zinc-700 dark:bg-zinc-900 dark:text-zinc-50"
            >
              {modules.length === 0 && <option value="SYSTEM">SYSTEM</option>}
              {modules.map((mod) => (
                <option key={mod.code} value={mod.code}>
                  {mod.code}{mod.label ? ` — ${mod.label}` : ''}
                </option>
              ))}
            </select>
            <span className="mt-1.5 block text-xs text-zinc-500 dark:text-zinc-400">
              Modul wajib sudah terdaftar di CPMATRIX (buat dulu via tombol Modul baru).
              UPPERCASE persis = nama folder menus/&lt;MODUL&gt;/.
            </span>
          </label>
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
              placeholder="mis. MENU_LAPORAN"
              hint="Kode menu induk se-modul."
            />
          </div>
          <Alert tone="info">
            Menu baru tidak otomatis diberi ke role mana pun. Centang manual di halaman Role, lalu buat
            foldernya — halaman 404 pemandu menunjukkan path persisnya.
          </Alert>
        </div>
      </Modal>

      <ConfirmDialog
        open={!!deletingMenu}
        title={`Hapus menu ${deletingMenu?.code}?`}
        message="Menu + grant-nya dihapus permanen. Gagal bila masih dipakai role atau masih punya menu anak."
        confirmLabel="Ya, hapus"
        danger
        onConfirm={removeMenu}
        onCancel={() => setDeletingMenu(null)}
      />

      <Modal
        open={showModuleCreate}
        onClose={creatingModule ? undefined : () => setShowModuleCreate(false)}
        title="Buat modul baru"
        size="sm"
        footer={
          <>
            <Button variant="secondary" onClick={() => setShowModuleCreate(false)} disabled={creatingModule}>
              Batal
            </Button>
            <Button onClick={createModuleItem} loading={creatingModule}>
              Buat modul
            </Button>
          </>
        }
      >
        <div className="flex flex-col gap-4">
          <TextField
            label="Kode modul"
            value={modCode}
            onChange={(e) => setModCode(e.target.value.toUpperCase())}
            placeholder="mis. REPORT"
            hint="UPPERCASE persis = nama folder menus/<MODUL>/, maks 40 karakter."
          />
          <TextField
            label="Label tampil"
            value={modLabel}
            onChange={(e) => setModLabel(e.target.value)}
            placeholder="mis. Report"
          />
          <TextField
            label="Urutan section"
            value={modSort}
            onChange={(e) => setModSort(e.target.value)}
            placeholder="10"
            hint="0–9999."
          />
          <Alert tone="info">
            Urut kerja: modul dulu, lalu menu di dalamnya, lalu centang role di
            halaman Role. Section modul langsung tampil di sidebar setelah ada menu
            yang diberi akses.
          </Alert>
        </div>
      </Modal>

      <ConfirmDialog
        open={!!deletingModule}
        title={`Hapus modul ${deletingModule?.code}?`}
        message="Modul dihapus permanen. Gagal bila masih ada menu di dalamnya — hapus/pindahkan dulu menunya."
        confirmLabel="Ya, hapus"
        danger
        onConfirm={removeModule}
        onCancel={() => setDeletingModule(null)}
      />
    </div>
  )
}
