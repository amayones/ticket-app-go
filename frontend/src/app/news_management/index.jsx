import { useEffect, useState } from 'react'
import {
  Alert,
  Badge,
  Button,
  Card,
  CardTitle,
  StandardPage,
  StandardGrid,
  useStandardController,
  useMessageBox,
  useToast,
  SELECT_CLASS,
} from '../shared/all.js'
import { formatTime } from '../../components/format.js'
import { newsCategories, read_data, newsStatus, process_delete } from './api.js'
import { controller } from './controller.js'
import NewsModal from './FRM.jsx'

export const meta = { label: 'News Management', icon: 'file', order: 3 }

const STATUSES = [
  { value: '', label: 'Semua status' },
  { value: 'draft', label: 'Draft' },
  { value: 'published', label: 'Terbit' },
  { value: 'archived', label: 'Diarsipkan' },
]

function toneFor(s) {
  switch (s) {
    case 'PUBLISHED':
      return 'success'
    case 'DRAFT':
      return 'neutral'
    case 'ARCHIVED':
      return 'warning'
    default:
      return 'neutral'
  }
}

export default function NewsManagement({ nvdata }) {
  const toast = useToast()
  const { MessageBox, dialog } = useMessageBox()
  const [status, setStatus] = useState('')
  const [category, setCategory] = useState('')
  const [query, setQuery] = useState('')
  const [cats, setCats] = useState([])
  const [showFRM, setShowFRM] = useState(false)
  const [editing, setEditing] = useState(null)
  const params = { status, category, q: query }
  const { rows, offset, setOffset, loading, showLoading, error, setError, load } = useStandardController(read_data, params)

  useEffect(() => {
    controller.init()
    newsCategories().then(setCats).catch(() => {})
  }, [])

  function handler_create() { setEditing(null); setShowFRM(true) }
  function handler_edit(row) { setEditing(row); setShowFRM(true) }
  async function handler_status(row, action) {
    try { await newsStatus(row.code, action); toast.success('Status diperbarui.'); load() } catch (ex) { toast.error(ex.message, { title: 'Error' }) }
  }
  async function handler_delete(row) {
    try {
      const ok = await MessageBox.delete({ headline: 'Hapus berita?', details: [{ label: 'Judul', value: row.title }], danger: true, request: () => process_delete(row.code) })
      if (!ok) return
      toast.success('Berita dihapus/diarsip.')
      load()
    } catch (ex) { toast.error(ex.message, { title: 'Error' }) }
  }

  const items = [{ rows }]
  const COLUMNS = [
    { header: 'Judul', dataIndex: 'title', width: 220, render: (r) => <span className="font-semibold">{r.title}</span> },
    { header: 'Kategori', dataIndex: 'news_category_name', width: 120, render: (r) => r.news_category_name || r.news_category_code || '—' },
    { header: 'Status', dataIndex: 'status', width: 100, render: (r) => <Badge tone={toneFor(r.status)}>{r.status}</Badge> },
    { header: 'Terbit', dataIndex: 'published_at', width: 140, render: (r) => (r.published_at ? formatTime(r.published_at) : r.scheduled_at ? `Jadwal ${formatTime(r.scheduled_at)}` : '—') },
    { header: 'Dibaca', dataIndex: 'view_count', width: 80, render: (r) => Number(r.view_count || 0).toLocaleString('id-ID') },
    {
      header: 'Aksi',
      dataIndex: 'code',
      width: 260,
      render: (r) => (
        <span className="flex flex-wrap gap-1">
          <Button size="sm" variant="secondary" onClick={() => handler_edit(r)}>Ubah</Button>
          {r.status === 'DRAFT' && <Button size="sm" variant="primary" onClick={() => handler_status(r, 'publish')}>Terbitkan</Button>}
          {r.status === 'PUBLISHED' && <Button size="sm" variant="secondary" onClick={() => handler_status(r, 'unpublish')}>Tarik</Button>}
          {r.status !== 'ARCHIVED' && <Button size="sm" variant="secondary" onClick={() => handler_status(r, 'archive')}>Arsipkan</Button>}
          <Button size="sm" variant="ghost" onClick={() => handler_delete(r)}>Hapus</Button>
        </span>
      ),
    },
  ]

  return (
    <div className="flex flex-col gap-4">
      <StandardPage
        title={nvdata?.label || 'News Management'}
        description="Kelola berita. Hanya yang terbit tampil di Discovery."
        loading={loading}
        showLoading={showLoading}
        error={error}
        errorTitle="Gagal memuat berita"
        onClearError={() => setError('')}
        onRefresh={() => controller.btrefresh_click(load)}
        onCreate={() => controller.btnew_click(handler_create)}
        createLabel="Create News"
        filterBar={
          <form onSubmit={(e) => e.preventDefault()} className="mb-4 flex flex-wrap items-end gap-2">
            <label className="flex flex-col gap-1 text-xs font-medium text-zinc-500">
              Status
              <select value={status} onChange={(e) => { setStatus(e.target.value); setOffset(0) }} className={SELECT_CLASS}>
                {STATUSES.map((s) => <option key={s.label} value={s.value}>{s.label}</option>)}
              </select>
            </label>
            <label className="flex flex-col gap-1 text-xs font-medium text-zinc-500">
              Kategori
              <select value={category} onChange={(e) => { setCategory(e.target.value); setOffset(0) }} className={SELECT_CLASS}>
                <option value="">Semua kategori</option>
                {cats.map((c) => <option key={c.code} value={c.code}>{c.name}</option>)}
              </select>
            </label>
            <label className="flex min-w-[160px] flex-1 flex-col gap-1 text-xs font-medium text-zinc-500">
              Cari judul
              <input value={query} onChange={(e) => { setQuery(e.target.value); setOffset(0) }} placeholder="Ketik judul…" className={SELECT_CLASS} />
            </label>
          </form>
        }
        items={items}
        emptyTitle="Belum ada berita"
        emptyDescription="Buat berita pertama Anda."
        emptyAction={<Button size="sm" onClick={handler_create}>Create News</Button>}
        offset={offset}
        onPage={(next) => next >= 0 && setOffset(next)}
      >
        <StandardGrid columns={COLUMNS} rows={rows} minWidth={900} />
      </StandardPage>
      <NewsModal open={showFRM} editing={editing} cats={cats} onClose={() => { setShowFRM(false); setEditing(null) }} onSaved={load} />
      {dialog}
    </div>
  )
}
