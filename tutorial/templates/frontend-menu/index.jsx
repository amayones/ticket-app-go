import { useCallback, useEffect, useState } from 'react'
import { listItems } from './api.js'
import {
  Alert,
  Button,
  Card,
  CardTitle,
  EmptyState,
  Icon,
  SkeletonRows,
  useSmoothLoading,
  useToast,
} from '../../../components'

// TEMPLATE mainpage menu baru. Cara pakai:
// 1. Daftarkan dulu lewat UI Modul & Menu (atau SQL ke CPMENU): isi MODULE
//    (section sidebar, mis. REPORT), CODE (mis. MENU_STOK), LABEL (judul),
//    MCONTROL snake_case (mis. stok), KIND=CHILD + PARENT_CODE opsional.
//    PARENT (header buka-tutup) TANPA mcontrol/folder.
// 2. Copy folder ini ke app/<mcontrol>/, misalnya app/stok/.
//    Judul tampil = LABEL di CPMENU; MCONTROL = nama folder.
//    Sidebar 3 level MODULE -> PARENT -> CHILD murni dari database.
// 3. Sesuaikan meta di bawah (label, icon, order).
// 4. Ganti placeholder <menu> di api.js dan isi halaman dengan UI-mu.
// 5. npm run build + restart backend; menu otomatis muncul (bila role
//    sudah di-grant via halaman Role).
export const meta = { label: 'Menu Baru', icon: 'list', order: 50 }

export default function MenuBaru() {
  const toast = useToast()
  const [items, setItems] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  const load = useCallback(async () => {
    setLoading(true)
    setError('')
    try {
      setItems(await listItems())
    } catch (err) {
      setError(err.message)
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    load()
  }, [load])

  const showLoading = useSmoothLoading(loading)

  return (
    <Card>
      <div className="mb-4 flex flex-wrap items-center justify-between gap-3">
        <CardTitle description="Ganti dengan deskripsi menumu.">
          Menu Baru
        </CardTitle>
        <Button
          size="sm"
          onClick={() => toast.info('Ganti dengan aksimu.')}
        >
          <Icon name="plus" className="h-4 w-4" />
          Aksi
        </Button>
      </div>

      {error && (
        <Alert tone="error" title="Gagal memuat" closable onClose={() => setError('')}>
          {error}
        </Alert>
      )}

      {showLoading ? (
        <SkeletonRows rows={3} />
      ) : items.length === 0 ? (
        <EmptyState
          title="Belum ada data"
          description="Ganti dengan konten menumu di sini."
        />
      ) : (
        <ul className="flex flex-col gap-2.5">
          {items.map((it) => (
            <li
              key={it.code || it.id}
              className="rounded-xl border border-zinc-100 p-3 dark:border-zinc-800"
            >
              {it.name || JSON.stringify(it)}
            </li>
          ))}
        </ul>
      )}
    </Card>
  )
}
