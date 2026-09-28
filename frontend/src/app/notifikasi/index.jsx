// ===== TEMPLATE SERAGAM app/* (ala <mod>.js langit_v2) =====
// Shell + controller standar; body custom 2 tab (template + log).
// Items grid/form di ./columns.js; tabel DB: CPNOTIFTEMPLATE/CPNOTIFLOG via ./api.js.
import { Badge, Button, ConfirmDialog, Icon, Modal, StandardForm, StandardGrid, StandardPage, api, useEffect, useStandardController, useState, useToast } from '../shared/all.js'
import { createTemplate, deleteTemplate, listNotifLogs, listTemplates, sendNotification, updateTemplate } from './api.js'
import { NOTIF_LOG_COLUMNS, sendFields } from './columns.jsx'
import TemplateForm from './components/TemplateForm.jsx'

const FEATURES = { header: true, refresh: true, filter: false, tabs: true, create: true, edit: true, remove: true, pagination: true, empty: true, error: true, confirmDialog: true, extraActions: true }

const CONFIG = {
  title: 'Notification Template & Log',
  description: 'Kelola template pesan (dengan variabel {{…}}) dan lihat riwayat pengiriman.',
  errorTitle: 'Gagal memuat',
  emptyTemplatesTitle: 'Belum ada template',
  emptyTemplatesDescription: 'Buat template pertama untuk mulai mengirim notifikasi.',
  emptyLogsTitle: 'Belum ada pengiriman',
  emptyLogsDescription: 'Kirim notifikasi pertama lewat tombol Kirim / Test.',
}

export const meta = { label: 'Notifikasi', icon: 'bell', order: 8 }

export default function Notifications() {
  const toast = useToast()
  const [tab, setTab] = useState('templates')
  const [templates, setTemplates] = useState([])
  const [editing, setEditing] = useState(null) // template obj | 'new'
  const [saving, setSaving] = useState(false)
  const [deleting, setDeleting] = useState(null)
  const [sending, setSending] = useState(false)
  const [sendOpen, setSendOpen] = useState(false)
  const [sendValues, setSendValues] = useState({ template_code: '', recipient: '', nama: '', detail: '' })

  const { rows: logs, offset, setOffset, loading, showLoading, error, setError, load: loadLogs } =
    useStandardController(
      ({ limit, offset, tab }) => (tab === 'logs' ? listNotifLogs(limit, offset) : Promise.resolve([])),
      { tab }
    )

  async function loadTemplates() {
    try {
      setTemplates(await listTemplates())
    } catch (err) {
      setError(err.message)
    }
  }

  useEffect(() => {
    loadTemplates()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  async function refresh() {
    if (tab === 'logs') loadLogs()
    else {
      try {
        setTemplates(await listTemplates())
      } catch (err) {
        setError(err.message)
      }
    }
  }

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
      refresh()
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
      refresh()
    } catch (err) {
      toast.error(err.message, { title: 'Gagal menghapus' })
    }
  }

  async function send() {
    if (!sendValues.template_code || !sendValues.recipient.trim()) {
      toast.warning('Pilih template dan isi penerima.')
      return
    }
    setSending(true)
    try {
      const me = api.currentUser()
      await sendNotification({
        template_code: sendValues.template_code,
        recipient: sendValues.recipient.trim(),
        variables: {
          nama: sendValues.nama || me?.username || '',
          kode: me?.code || '',
          role: me?.role || '',
          detail: sendValues.detail,
        },
      })
      toast.success(`Notifikasi dikirim ke ${sendValues.recipient.trim()}.`)
      setSendOpen(false)
      setSendValues({ template_code: '', recipient: '', nama: '', detail: '' })
      if (tab === 'logs') loadLogs()
    } catch (err) {
      toast.error(err.message, { title: 'Gagal mengirim' })
    } finally {
      setSending(false)
    }
  }

  const isLogs = tab === 'logs'
  const items = isLogs ? logs : templates
  const activeTemplates = templates.filter((t) => t.is_active)

  return (
    <div className="flex flex-col gap-4">
      <StandardPage
        title={CONFIG.title}
        description={CONFIG.description}
        features={FEATURES}
        loading={loading}
        showLoading={tab === 'templates' ? false : showLoading}
        error={error}
        errorTitle={CONFIG.errorTitle}
        onClearError={() => setError('')}
        onRefresh={refresh}
        extraActions={
          <Button
            size="sm"
            onClick={() => {
              setSendValues((v) => ({ ...v, template_code: activeTemplates[0]?.code || '' }))
              setSendOpen(true)
            }}
          >
            <Icon name="send" className="h-4 w-4" />
            Kirim / Test
          </Button>
        }
        tabsBar={
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
        }
        items={items}
        emptyTitle={isLogs ? CONFIG.emptyLogsTitle : CONFIG.emptyTemplatesTitle}
        emptyDescription={isLogs ? CONFIG.emptyLogsDescription : CONFIG.emptyTemplatesDescription}
        emptyIcon={isLogs ? 'send' : 'bell'}
        emptyAction={
          !isLogs ? (
            <Button size="sm" variant="secondary" onClick={() => setEditing('new')}>
              <Icon name="plus" className="h-4 w-4" />
              Template baru
            </Button>
          ) : null
        }
        offset={offset}
        onPage={isLogs ? setOffset : null}
      >
        {!isLogs ? (
          <div>
            <div className="mb-3">
              <Button size="sm" variant="secondary" onClick={() => setEditing('new')}>
                <Icon name="plus" className="h-4 w-4" />
                Template baru
              </Button>
            </div>
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
          </div>
        ) : (
          <StandardGrid columns={NOTIF_LOG_COLUMNS} rows={logs} minWidth={560} />
        )}
      </StandardPage>

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
        <StandardForm fields={sendFields(activeTemplates)} values={sendValues} onChange={setSendValues} />
      </Modal>

      {FEATURES.confirmDialog && (
        <ConfirmDialog
          open={!!deleting}
          title={`Hapus template ${deleting?.name}?`}
          message="Template dihapus permanen. Riwayat pengiriman yang sudah tercatat tidak ikut terhapus."
          confirmLabel="Ya, hapus"
          danger
          onConfirm={remove}
          onCancel={() => setDeleting(null)}
        />
      )}
    </div>
  )
}
