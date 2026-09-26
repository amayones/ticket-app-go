import { useCallback, useEffect, useState } from 'react'
import { api } from '../../../api/client.js'
import {
  createTemplate,
  deleteTemplate,
  listNotifLogs,
  listTemplates,
  sendNotification,
  updateTemplate,
} from './api.js'

export const meta = { label: 'Notifikasi', icon: 'bell', order: 7 }
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
  Pagination,
  SkeletonRows,
  TextField,
  useSmoothLoading,
  formatTime,
  useToast,
} from '../../../components'

const PAGE_SIZE = 15
const CHANNELS = ['EMAIL', 'PUSH', 'INAPP']

const inputClass =
  'w-full rounded-xl border border-zinc-300 bg-white px-3.5 py-2.5 text-sm text-zinc-900 focus:border-violet-500 focus:outline-none focus:ring-2 focus:ring-violet-500/30 dark:border-zinc-700 dark:bg-zinc-900 dark:text-zinc-50'

function TemplateForm({ initial, saving, onSubmit, onCancel }) {
  const [name, setName] = useState(initial?.name || '')
  const [channel, setChannel] = useState(initial?.channel || 'EMAIL')
  const [subject, setSubject] = useState(initial?.subject || '')
  const [body, setBody] = useState(initial?.body || '')
  const [active, setActive] = useState(initial ? !!initial.is_active : true)
  return (
    <div className="flex flex-col gap-4">
      <TextField label="Nama template" value={name} onChange={(e) => setName(e.target.value)} placeholder="mis. Promo Akhir Tahun" />
      <label className="block">
        <span className="mb-1.5 block text-sm font-medium text-zinc-700 dark:text-zinc-200">Channel</span>
        <select value={channel} onChange={(e) => setChannel(e.target.value)} className={inputClass}>
          {CHANNELS.map((c) => (
            <option key={c} value={c}>{c}</option>
          ))}
        </select>
      </label>
      <TextField label="Subjek" value={subject} onChange={(e) => setSubject(e.target.value)} placeholder="dipakai untuk EMAIL" />
      <label className="block">
        <span className="mb-1.5 block text-sm font-medium text-zinc-700 dark:text-zinc-200">Isi pesan</span>
        <textarea
          value={body}
          onChange={(e) => setBody(e.target.value)}
          rows={5}
          placeholder="Halo {{nama}}, …"
          className={`${inputClass} font-mono`}
        />
        <span className="mt-1.5 block text-xs text-zinc-500">
          Variabel: <code>{'{{nama}}'}</code> <code>{'{{kode}}'}</code> <code>{'{{role}}'}</code> <code>{'{{detail}}'}</code>
        </span>
      </label>
      <label className="flex cursor-pointer items-center gap-2 text-sm text-zinc-700 dark:text-zinc-200">
        <input type="checkbox" checked={active} onChange={(e) => setActive(e.target.checked)} className="h-4 w-4 accent-violet-600" />
        Template aktif (nonaktif = tidak bisa dikirim)
      </label>
      <div className="flex justify-end gap-2">
        <Button variant="secondary" onClick={onCancel} disabled={saving}>
          Batal
        </Button>
        <Button onClick={() => onSubmit({ name, channel, subject, body, is_active: active })} loading={saving}>
          Simpan template
        </Button>
      </div>
    </div>
  )
}

export default function Notifications() {
  const toast = useToast()
  const [tab, setTab] = useState('templates')
  const [templates, setTemplates] = useState([])
  const [logs, setLogs] = useState([])
  const [offset, setOffset] = useState(0)
  const [loading, setLoading] = useState(true)
  const showLoading = useSmoothLoading(loading)
  const [error, setError] = useState('')
  const [editing, setEditing] = useState(null) // template obj | 'new'
  const [saving, setSaving] = useState(false)
  const [deleting, setDeleting] = useState(null)
  const [sending, setSending] = useState(false)
  const [sendOpen, setSendOpen] = useState(false)
  const [sendTpl, setSendTpl] = useState('')
  const [recipient, setRecipient] = useState('')
  const [varNama, setVarNama] = useState('')
  const [varDetail, setVarDetail] = useState('')

  const load = useCallback(async () => {
    setLoading(true)
    setError('')
    try {
      if (tab === 'logs') {
        setLogs(await listNotifLogs(PAGE_SIZE, offset))
      } else {
        setTemplates(await listTemplates())
      }
    } catch (err) {
      setError(err.message)
    } finally {
      setLoading(false)
    }
  }, [tab, offset])

  useEffect(() => {
    load()
  }, [load])

  async function saveTemplate(payload) {
    if (!payload.name.trim() || !payload.body.trim()) {
      toast.warning('Nama dan isi template wajib diisi.')
      return
    }
    setSaving(true)
    try {
      if (editing === 'new') {
        await createTemplate(payload)
        toast.success('Template dibuat.')
      } else {
        await updateTemplate(editing.code, payload)
        toast.success('Template diperbarui.')
      }
      setEditing(null)
      load()
    } catch (err) {
      toast.error(err.message, { title: 'Gagal menyimpan' })
    } finally {
      setSaving(false)
    }
  }

  async function remove() {
    if (!deleting) return
    try {
      await deleteTemplate(deleting.code)
      toast.success(`Template ${deleting.name} dihapus.`)
      setDeleting(null)
      load()
    } catch (err) {
      toast.error(err.message, { title: 'Gagal menghapus' })
    }
  }

  async function send() {
    if (!sendTpl || !recipient.trim()) {
      toast.warning('Pilih template dan isi penerima.')
      return
    }
    setSending(true)
    try {
      const me = api.currentUser()
      await sendNotification({
        template_code: sendTpl,
        recipient: recipient.trim(),
        variables: { nama: varNama || me?.username || '', kode: me?.code || '', role: me?.role || '', detail: varDetail },
      })
      toast.success(`Notifikasi dikirim ke ${recipient.trim()}.`)
      setSendOpen(false)
      setRecipient('')
      setVarNama('')
      setVarDetail('')
      if (tab === 'logs') load()
    } catch (err) {
      toast.error(err.message, { title: 'Gagal mengirim' })
    } finally {
      setSending(false)
    }
  }

  return (
    <div className="flex flex-col gap-4">
      <Card>
        <div className="mb-4 flex flex-wrap items-center justify-between gap-3">
          <CardTitle description="Kelola template pesan (dengan variabel {{…}}) dan lihat riwayat pengiriman.">
            Notification Template & Log
          </CardTitle>
          <div className="flex gap-2">
            <Button variant="secondary" size="sm" onClick={load} loading={loading}>
              <Icon name="refresh" className="h-4 w-4" />
              Muat ulang
            </Button>
            <Button
              size="sm"
              onClick={() => {
                setSendTpl(templates.find((t) => t.is_active)?.code || '')
                setSendOpen(true)
              }}
            >
              <Icon name="send" className="h-4 w-4" />
              Kirim / Test
            </Button>
          </div>
        </div>

        <div className="mb-4 flex gap-1.5 rounded-xl bg-zinc-100 p-1 dark:bg-zinc-800">
          {[
            ['templates', 'Template'],
            ['logs', 'Riwayat kirim'],
          ].map(([key, label]) => (
            <button
              key={key}
              type="button"
              onClick={() => { setTab(key); setOffset(0) }}
              className={`flex-1 rounded-lg px-3 py-1.5 text-sm font-semibold transition ${
                tab === key
                  ? 'bg-white text-zinc-900 shadow-sm dark:bg-zinc-900 dark:text-zinc-50'
                  : 'text-zinc-500 dark:text-zinc-400'
              }`}
            >
              {label}
            </button>
          ))}
        </div>

        {error && (
          <Alert tone="error" title="Gagal memuat" closable onClose={() => setError('')} className="mb-4">
            {error}
          </Alert>
        )}

        {showLoading ? (
          <SkeletonRows rows={3} />
        ) : tab === 'templates' ? (
          <>
            <div className="mb-3">
              <Button size="sm" variant="secondary" onClick={() => setEditing('new')}>
                <Icon name="plus" className="h-4 w-4" />
                Template baru
              </Button>
            </div>
            {templates.length === 0 ? (
              <EmptyState icon="bell" title="Belum ada template" description="Buat template pertama untuk mulai mengirim notifikasi." />
            ) : (
              <ul className="flex flex-col gap-2.5">
                {templates.map((t) => (
                  <li key={t.code} className="flex items-center gap-3 rounded-xl border border-zinc-100 p-3 dark:border-zinc-800">
                    <span className="flex h-10 w-10 shrink-0 items-center justify-center rounded-full bg-sky-100 text-sky-600 dark:bg-sky-950 dark:text-sky-300">
                      <Icon name="bell" className="h-5 w-5" />
                    </span>
                    <div className="min-w-0 flex-1">
                      <p className="flex flex-wrap items-center gap-2 text-sm font-semibold text-zinc-900 dark:text-zinc-50">
                        <span className="truncate">{t.name}</span>
                        <Badge tone="info">{t.channel}</Badge>
                        <Badge tone={t.is_active ? 'success' : 'neutral'}>{t.is_active ? 'Aktif' : 'Nonaktif'}</Badge>
                      </p>
                      <p className="truncate font-mono text-xs text-zinc-500">{t.code}</p>
                    </div>
                    <div className="flex shrink-0 gap-1">
                      <button type="button" title="Edit" onClick={() => setEditing(t)} className="rounded-lg p-2 text-zinc-500 hover:bg-zinc-100 hover:text-violet-600 dark:hover:bg-zinc-800">
                        <Icon name="pencil" className="h-4 w-4" />
                      </button>
                      <button type="button" title="Hapus" onClick={() => setDeleting(t)} className="rounded-lg p-2 text-zinc-500 hover:bg-rose-50 hover:text-rose-600 dark:hover:bg-rose-950">
                        <Icon name="trash" className="h-4 w-4" />
                      </button>
                    </div>
                  </li>
                ))}
              </ul>
            )}
          </>
        ) : logs.length === 0 ? (
          <EmptyState icon="send" title="Belum ada pengiriman" description="Kirim notifikasi pertama lewat tombol Kirim / Test." />
        ) : (
          <>
            <div className="overflow-x-auto rounded-xl border border-zinc-100 dark:border-zinc-800">
              <table className="w-full min-w-[560px] text-left text-sm">
                <thead>
                  <tr className="bg-zinc-50 text-xs uppercase tracking-wide text-zinc-500 dark:bg-zinc-800/60 dark:text-zinc-400">
                    <th className="px-3 py-2.5">Waktu</th>
                    <th className="px-3 py-2.5">Channel</th>
                    <th className="px-3 py-2.5">Penerima</th>
                    <th className="px-3 py-2.5">Subjek</th>
                    <th className="px-3 py-2.5">Status</th>
                  </tr>
                </thead>
                <tbody>
                  {logs.map((l) => (
                    <tr key={l.code} className="border-t border-zinc-100 dark:border-zinc-800">
                      <td className="whitespace-nowrap px-3 py-2.5 text-xs text-zinc-500">{formatTime(l.created_at)}</td>
                      <td className="px-3 py-2.5"><Badge tone="info">{l.channel}</Badge></td>
                      <td className="max-w-[180px] truncate px-3 py-2.5 text-xs">{l.recipient}</td>
                      <td className="max-w-[220px] truncate px-3 py-2.5 text-xs" title={l.body}>{l.subject || l.body}</td>
                      <td className="px-3 py-2.5"><Badge tone={l.status === 'SENT' ? 'success' : 'danger'}>{l.status}</Badge></td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
            <div className="mt-4 border-t border-zinc-100 pt-4 dark:border-zinc-800">
              <Pagination offset={offset} limit={PAGE_SIZE} count={logs.length} hasMore={logs.length === PAGE_SIZE} loading={loading} onPage={setOffset} />
            </div>
          </>
        )}
      </Card>

      <Modal open={editing !== null} onClose={saving ? undefined : () => setEditing(null)} title={editing === 'new' ? 'Template baru' : `Edit ${editing?.name || ''}`} size="lg" closeOnBackdrop={!saving}>
        {editing !== null && (
          <TemplateForm
            key={editing === 'new' ? 'new' : editing.code}
            initial={editing === 'new' ? null : editing}
            saving={saving}
            onSubmit={saveTemplate}
            onCancel={() => setEditing(null)}
          />
        )}
      </Modal>

      <Modal
        open={sendOpen}
        onClose={sending ? undefined : () => setSendOpen(false)}
        title="Kirim / test notifikasi"
        size="sm"
        footer={
          <>
            <Button variant="secondary" onClick={() => setSendOpen(false)} disabled={sending}>Batal</Button>
            <Button onClick={send} loading={sending}><Icon name="send" className="h-4 w-4" />Kirim</Button>
          </>
        }
      >
        <div className="flex flex-col gap-4">
          <label className="block">
            <span className="mb-1.5 block text-sm font-medium text-zinc-700 dark:text-zinc-200">Template</span>
            <select value={sendTpl} onChange={(e) => setSendTpl(e.target.value)} className={inputClass}>
              <option value="">— pilih —</option>
              {templates.filter((t) => t.is_active).map((t) => (
                <option key={t.code} value={t.code}>{t.name} ({t.channel})</option>
              ))}
            </select>
          </label>
          <TextField label="Penerima" value={recipient} onChange={(e) => setRecipient(e.target.value)} placeholder="email / id perangkat / kode user" />
          <TextField label="Variabel {{nama}}" value={varNama} onChange={(e) => setVarNama(e.target.value)} placeholder="nama penerima" />
          <TextField label="Variabel {{detail}}" value={varDetail} onChange={(e) => setVarDetail(e.target.value)} placeholder="detail tambahan" />
        </div>
      </Modal>

      <ConfirmDialog
        open={!!deleting}
        title={`Hapus template ${deleting?.name}?`}
        message="Template dihapus permanen. Riwayat pengiriman yang sudah tercatat tidak ikut terhapus."
        confirmLabel="Ya, hapus"
        danger
        onConfirm={remove}
        onCancel={() => setDeleting(null)}
      />
    </div>
  )
}
