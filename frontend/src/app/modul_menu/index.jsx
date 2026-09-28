// View Modul & Menu — padanan form master custom langit_v2
// (layout modul+menu, bukan grid standar).
// Controller: ./controller.js (6 fungsi). Tabel: CPMODULE/CPMENU.
import { Alert, Badge, Button, Card, CardTitle, ConfirmDialog, EmptyState, Icon, Modal, SkeletonRows, TextField, Tooltip, useSmoothLoading, useToast, useCallback, useEffect, useMemo, useState } from '../shared/all.js'
import { createMenu, createModule, deleteMenu, deleteModule, listMenus, listModules, updateMenu } from './api.js'
import { controller } from './controller.js'

export const meta = { label: 'Modul & Menu', icon: 'list', order: 2 }

// Menu bertipe PARENT saja yang boleh dipilih sebagai parent menu lain.
const PARENT_KIND = 'PARENT'
const CHILD_KIND = 'CHILD'

function groupMenusByModule(menus) {
  const groups = {}
  for (const m of menus) {
    const g = m.module || 'SYSTEM'
    if (!groups[g]) groups[g] = []
    groups[g].push(m)
  }
  for (const g of Object.keys(groups)) {
    groups[g].sort((a, b) => (a.sort_order ?? 99) - (b.sort_order ?? 99) || a.code.localeCompare(b.code))
  }
  return groups
}

// parentOptions: kandidat parent untuk form (modul sama, jenis PARENT,
// bukan menu itu sendiri).
function parentOptions(menus, module, selfCode) {
  return menus.filter(
    (m) => (m.module || 'SYSTEM') === module && m.kind === PARENT_KIND && m.code !== selfCode
  )
}

// Halaman Modul & Menu (master CPMODULE + registry CPMENU).
// Urut kerja: buat modul dulu, lalu menu di dalamnya, lalu centang role
// di halaman Role. Tanpa auto-grant ke role mana pun. Menu CHILD tanpa
// folder tampil sebagai halaman 404 pemandu (bukan hilang diam-diam).
// Sidebar 3 level: MODULE (section) -> PARENT (header buka-tutup, tanpa
// folder) -> CHILD (item, folder app/<mcontrol>/).
export default function Modul() {
  const toast = useToast()
  const [menus, setMenus] = useState([])
  const [modules, setModules] = useState([])
  const [loading, setLoading] = useState(true)
  const showLoading = useSmoothLoading(loading)
  const [error, setError] = useState('')
  const [showMenuCreate, setShowMenuCreate] = useState(false)
  const [mCode, setMCode] = useState('')
  const [mModule, setMModule] = useState('SYSTEM')
  const [mLabel, setMLabel] = useState('')
  const [mMcontrol, setMMcontrol] = useState('')
  const [mSort, setMSort] = useState('99')
  const [mKind, setMKind] = useState(CHILD_KIND)
  const [mParent, setMParent] = useState('')
  const [creatingMenu, setCreatingMenu] = useState(false)
  const [deletingMenu, setDeletingMenu] = useState(null)
  // Form edit: field yang sama persis dengan form create, minus kode permission.
  const [editingMenu, setEditingMenu] = useState(null)
  const [eModule, setEModule] = useState('')
  const [eLabel, setELabel] = useState('')
  const [eMcontrol, setEMcontrol] = useState('')
  const [eSort, setESort] = useState('99')
  const [eKind, setEKind] = useState(CHILD_KIND)
  const [eParent, setEParent] = useState('')
  const [savingEdit, setSavingEdit] = useState(false)
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
    controller.init()
    controller.btrefresh_click(load)
  }, [load])

  const menuGroups = useMemo(() => groupMenusByModule(menus), [menus])

  async function createMenuItem() {
    if (!mCode.trim() || !mModule.trim() || !mLabel.trim()) {
      toast.warning('Kode, modul, dan label menu wajib diisi.')
      return
    }
    if (mKind === CHILD_KIND && !mMcontrol.trim()) {
      toast.warning('MCONTROL wajib diisi untuk menu CHILD (nama folder).')
      return
    }
    setCreatingMenu(true)
    try {
      const menu = await createMenu({
        code: mCode.trim(),
        module: mModule.trim(),
        label: mLabel.trim(),
        mcontrol: mKind === PARENT_KIND ? '' : mMcontrol.trim().toLowerCase(),
        kind: mKind,
        sort_order: Number(mSort) || 99,
        parent_code: mKind === PARENT_KIND ? '' : mParent.trim(),
      })
      toast.success(`Menu ${menu.code} dibuat di modul ${menu.module}. Centang role yang boleh akses, lalu buat folder app/${menu.mcontrol || '(parent: tanpa folder)'}.`)
      setMCode('')
      setMLabel('')
      setMMcontrol('')
      setMSort('99')
      setMKind(CHILD_KIND)
      setMParent('')
      setShowMenuCreate(false)
      load()
    } catch (err) {
      toast.error(err.message, { title: 'Gagal membuat menu' })
    } finally {
      setCreatingMenu(false)
    }
  }

  // Buka form edit: isi field dari menu yang dipilih (sama seperti form create).
  function openMenuEdit(menu) {
    setEditingMenu(menu)
    setEModule(menu.module || 'SYSTEM')
    setELabel(menu.label || '')
    setEMcontrol(menu.mcontrol || '')
    setESort(String(menu.sort_order ?? 99))
    setEKind(menu.kind || CHILD_KIND)
    setEParent(menu.parent_code || '')
  }

  async function saveMenuEdit() {
    if (!editingMenu) return
    if (!eModule.trim() || !eLabel.trim()) {
      toast.warning('Modul dan label menu wajib diisi.')
      return
    }
    setSavingEdit(true)
    try {
      const menu = await updateMenu(editingMenu.code, {
        module: eModule.trim(),
        label: eLabel.trim(),
        mcontrol: eKind === PARENT_KIND ? '' : eMcontrol.trim().toLowerCase(),
        kind: eKind,
        sort_order: Number(eSort) || 0,
        parent_code: eKind === PARENT_KIND ? '' : eParent.trim(),
      })
      toast.success(`Menu ${menu.code} diperbarui.`)
      setEditingMenu(null)
      load()
    } catch (err) {
      toast.error(err.message, { title: 'Gagal memperbarui menu' })
    } finally {
      setSavingEdit(false)
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
          <CardTitle description="Urut: buat modul dulu, lalu menu di dalamnya, lalu centang role di halaman Role. CHILD tanpa folder tampil sebagai halaman 404 pemandu.">
            Modul &amp; Menu
          </CardTitle>
          <div className="flex gap-2">
            <Button variant="secondary" size="sm" onClick={() => controller.btnew_click(() => setShowModuleCreate(true))}>
              <Icon name="plus" className="h-4 w-4" />
              Modul baru
            </Button>
            <Button size="sm" onClick={() => controller.btnew_click(() => setShowMenuCreate(true))}>
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
                      {list[0]?.module_label || module} <span className="font-mono font-normal">· {module}</span>
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
                            <Badge tone={m.kind === PARENT_KIND ? 'brand' : 'neutral'}>
                              {m.kind || CHILD_KIND}
                            </Badge>
                            {m.mcontrol && <Badge tone="neutral">folder: {m.mcontrol}</Badge>}
                            {m.parent_code && <Badge>parent: {m.parent_code}</Badge>}
                          </p>
                          <p className="text-xs text-zinc-500 dark:text-zinc-400">
                            urutan {m.sort_order ?? 99}
                            {m.kind === PARENT_KIND ? ' • header buka-tutup (tanpa folder/halaman)' : ''}
                            {m.parent_code ? ' • anak dari parent di atas' : ''}
                          </p>
                        </div>
                        <Tooltip label="Edit menu">
                          <button
                            type="button"
                            aria-label={`Edit ${m.code}`}
                            onClick={() => openMenuEdit(m)}
                            className="shrink-0 rounded-lg p-2 text-zinc-500 transition-colors hover:bg-white hover:text-violet-600 hover:shadow-sm dark:hover:bg-zinc-800"
                          >
                            <Icon name="pencil" className="h-4 w-4" />
                          </button>
                        </Tooltip>
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
          <label className="block">
            <span className="mb-1.5 block text-sm font-medium text-zinc-700 dark:text-zinc-200">Modul (MODULE)</span>
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
              Modul wajib sudah terdaftar di CPMODULE (buat dulu via tombol Modul baru).
              Section sidebar = MODULE (CPMODULE). Folder menu terpisah: app/&lt;mcontrol&gt;/.
            </span>
          </label>
          <TextField
            label="Label tampil"
            value={mLabel}
            onChange={(e) => setMLabel(e.target.value)}
            placeholder="mis. Laporan"
          />
          {mKind === CHILD_KIND && (
            <TextField
              label="MCONTROL (nama folder)"
              value={mMcontrol}
              onChange={(e) => setMMcontrol(e.target.value.toLowerCase().replace(/[^a-z0-9_]/g, ''))}
              placeholder="mis. laporan"
              hint="snake_case (a-z, 0-9, _) = folder frontend/src/app/<mcontrol>/. PARENT tanpa mcontrol."
            />
          )}
          <TextField
            label="Urutan"
            value={mSort}
            onChange={(e) => setMSort(e.target.value)}
            placeholder="99"
            hint="0–9999. Urutan sidebar ikut nilai ini."
          />
          <label className="block">
            <span className="mb-1.5 block text-sm font-medium text-zinc-700 dark:text-zinc-200">Jenis menu</span>
            <select
              value={mKind}
              onChange={(e) => {
                setMKind(e.target.value)
                if (e.target.value === PARENT_KIND) setMParent('')
              }}
              className="w-full rounded-xl border border-zinc-300 bg-white px-3.5 py-2.5 text-sm text-zinc-900 focus:border-violet-500 focus:outline-none focus:ring-2 focus:ring-violet-500/30 dark:border-zinc-700 dark:bg-zinc-900 dark:text-zinc-50"
            >
              <option value={CHILD_KIND}>CHILD — item biasa (punya folder)</option>
              <option value={PARENT_KIND}>PARENT — header buka-tutup (tanpa folder)</option>
            </select>
            <span className="mt-1.5 block text-xs text-zinc-500 dark:text-zinc-400">
              Ditulis ke kolom MENU_KIND di CPMENU. PARENT = header buka-tutup tanpa
              folder/halaman; CHILD = item biasa dengan folder.
            </span>
          </label>
          {mKind === CHILD_KIND && (
            <label className="block">
              <span className="mb-1.5 block text-sm font-medium text-zinc-700 dark:text-zinc-200">
                Parent (menu induk, opsional)
              </span>
              <select
                value={mParent}
                onChange={(e) => setMParent(e.target.value)}
                className="w-full rounded-xl border border-zinc-300 bg-white px-3.5 py-2.5 text-sm text-zinc-900 focus:border-violet-500 focus:outline-none focus:ring-2 focus:ring-violet-500/30 dark:border-zinc-700 dark:bg-zinc-900 dark:text-zinc-50"
              >
                <option value="">— tanpa parent (menu tingkat modul) —</option>
                {parentOptions(menus, mModule).map((p) => (
                  <option key={p.code} value={p.code}>
                    {p.code} — {p.label}
                  </option>
                ))}
              </select>
              <span className="mt-1.5 block text-xs text-zinc-500 dark:text-zinc-400">
                Hanya menu ber-KIND = PARENT di modul {mModule} yang bisa dipilih.
                Nilainya ditulis ke kolom PARENT_CODE.
              </span>
            </label>
          )}
          <Alert tone="info">
            Menu baru tidak otomatis diberi ke role mana pun. Centang manual di halaman Role, lalu buat
            foldernya — halaman 404 pemandu menunjukkan path persisnya.
          </Alert>
        </div>
      </Modal>

      <Modal
        open={!!editingMenu}
        onClose={savingEdit ? undefined : () => setEditingMenu(null)}
        title={`Edit menu ${editingMenu?.code || ''}`}
        footer={
          <>
            <Button variant="secondary" onClick={() => setEditingMenu(null)} disabled={savingEdit}>
              Batal
            </Button>
            <Button onClick={saveMenuEdit} loading={savingEdit}>
              Simpan perubahan
            </Button>
          </>
        }
      >
        <div className="flex flex-col gap-4">
          <Alert tone="info">
            Kode permission <span className="font-mono">{editingMenu?.code}</span> tidak bisa
            diubah — ia acuan grant role. Ingin ganti kode? Buat menu baru, lalu hapus yang ini.
          </Alert>
          <label className="block">
            <span className="mb-1.5 block text-sm font-medium text-zinc-700 dark:text-zinc-200">Modul (MODULE)</span>
            <select
              value={eModule}
              onChange={(e) => setEModule(e.target.value)}
              className="w-full rounded-xl border border-zinc-300 bg-white px-3.5 py-2.5 text-sm text-zinc-900 focus:border-violet-500 focus:outline-none focus:ring-2 focus:ring-violet-500/30 dark:border-zinc-700 dark:bg-zinc-900 dark:text-zinc-50"
            >
              {!modules.some((mod) => mod.code === eModule) && <option value={eModule}>{eModule}</option>}
              {modules.map((mod) => (
                <option key={mod.code} value={mod.code}>
                  {mod.code}{mod.label ? ` — ${mod.label}` : ''}
                </option>
              ))}
            </select>
            <span className="mt-1.5 block text-xs text-zinc-500 dark:text-zinc-400">
              Pindah modul = pindah section sidebar (folder menu tidak ikut pindah).
            </span>
          </label>
          <TextField
            label="Label tampil"
            value={eLabel}
            onChange={(e) => setELabel(e.target.value)}
            placeholder="mis. Laporan"
            hint="Nama yang tampil di sidebar dan matriks Role."
          />
          {eKind === CHILD_KIND && (
            <TextField
              label="MCONTROL (nama folder)"
              value={eMcontrol}
              onChange={(e) => setEMcontrol(e.target.value.toLowerCase().replace(/[^a-z0-9_]/g, ''))}
              placeholder="mis. laporan"
              hint="snake_case = folder frontend/src/app/<mcontrol>/. PARENT tanpa mcontrol."
            />
          )}
          <TextField
            label="Urutan"
            value={eSort}
            onChange={(e) => setESort(e.target.value)}
            placeholder="99"
            hint="0–9999. Urutan sidebar ikut nilai ini."
          />
          <label className="block">
            <span className="mb-1.5 block text-sm font-medium text-zinc-700 dark:text-zinc-200">Jenis menu</span>
            <select
              value={eKind}
              onChange={(e) => {
                setEKind(e.target.value)
                if (e.target.value === PARENT_KIND) setEParent('')
              }}
              className="w-full rounded-xl border border-zinc-300 bg-white px-3.5 py-2.5 text-sm text-zinc-900 focus:border-violet-500 focus:outline-none focus:ring-2 focus:ring-violet-500/30 dark:border-zinc-700 dark:bg-zinc-900 dark:text-zinc-50"
            >
              <option value={CHILD_KIND}>CHILD — item biasa (punya folder)</option>
              <option value={PARENT_KIND}>PARENT — header buka-tutup (tanpa folder)</option>
            </select>
            <span className="mt-1.5 block text-xs text-zinc-500 dark:text-zinc-400">
              Menu yang sudah punya anak tidak boleh diubah jadi CHILD.
            </span>
          </label>
          {eKind === CHILD_KIND && (
            <label className="block">
              <span className="mb-1.5 block text-sm font-medium text-zinc-700 dark:text-zinc-200">
                Parent (menu induk, opsional)
              </span>
              <select
                value={eParent}
                onChange={(e) => setEParent(e.target.value)}
                className="w-full rounded-xl border border-zinc-300 bg-white px-3.5 py-2.5 text-sm text-zinc-900 focus:border-violet-500 focus:outline-none focus:ring-2 focus:ring-violet-500/30 dark:border-zinc-700 dark:bg-zinc-900 dark:text-zinc-50"
              >
                <option value="">— tanpa parent (menu tingkat modul) —</option>
                {parentOptions(menus, eModule, editingMenu?.code).map((p) => (
                  <option key={p.code} value={p.code}>
                    {p.code} — {p.label}
                  </option>
                ))}
              </select>
              <span className="mt-1.5 block text-xs text-zinc-500 dark:text-zinc-400">
                Pilih parent untuk melepas menu ini dari induknya (PARENT_CODE = kosong).
              </span>
            </label>
          )}
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
