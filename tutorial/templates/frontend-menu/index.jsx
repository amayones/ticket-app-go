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
// 1. Copy folder ini ke menus/user/<menu>/ untuk semua role yang diberi
//    permission menu, atau menus/admin/<menu>/ untuk menu khusus ADMIN.
//    Nama folder menjadi kode permission otomatis, contoh laporan -> MENU_LAPORAN.
// 2. Sesuaikan meta di bawah (label, icon, order).
// 3. Ganti placeholder <menu> di api.js dan isi halaman dengan UI-mu.
// 4. Jalankan migrate jika permission baru, lalu build dan restart backend.
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
