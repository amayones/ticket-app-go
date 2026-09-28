// View Notifikasi — padanan <mod>.js langit_v2 (shell + toolbar + items).
// Controller: ./controller.js (6 fungsi). Store: read_data (log) + listTemplates.
// Tab Template = master (FRM), tab Log = viewer GRID. Tabel: CPNOTIFTEMPLATE/CPNOTIFLOG.
import { useEffect, useState } from 'react'
import { Badge, Button, Icon, Modal, StandardForm, StandardGrid, StandardPage, DownloadButton, api, useStandardController, useToast } from '../shared/all.js'
import { read_data, listTemplates, sendNotification } from './api.js'
import { controller } from './controller.js'
import GRID from './GRID.jsx'
import FRM from './FRM.jsx'

export const meta = { label: 'Notifikasi', icon: 'bell', order: 8 }

const TEMPLATE_COLUMNS = [
  { header: 'Nama', dataIndex: 'name', width: 180 },
  {
    header: 'Channel',
    dataIndex: 'channel',
    width: 90,
    render: (t) => <Badge tone="info">{t.channel}</Badge>,
  },
  {
    header: 'Status',
    dataIndex: 'is_active',
    width: 90,
    render: (t) => <Badge tone={t.is_active ? 'success' : 'neutral'}>{t.is_active ? 'Aktif' : 'Nonaktif'}</Badge>,
  },
  { header: 'Kode', dataIndex: 'code', width: 140 },
]

const SEND_FIELDS = [
  { name: 'recipient', label: 'Penerima', type: 'text', placeholder: 'email / id perangkat / kode user' },
  { name: 'nama', label: 'Variabel {{nama}}', type: 'text', placeholder: 'nama penerima' },
  { name: 'detail', label: 'Variabel {{detail}}', type: 'text', placeholder: 'detail tambahan' },
]

export default function Notifications({ nvdata }) {
  const toast = useToast()
  const [tab, setTab] = useState('templates')
  const [templates, setTemplates] = useState([])
  const [tplError, setTplError] = useState('')
  const [showFRM, setShowFRM] = useState(null) // null | 'new' | template
  const [sendOpen, setSendOpen] = useState(false)
  const [sendTpl, setSendTpl] = useState('')
  const [sendValues, setSendValues] = useState({ recipient: '', nama: '', detail: '' })
  const [sending, setSending] = useState(false)

  const { rows: logs, offset, setOffset, loading, showLoading, error, setError, load: loadLogs } =
    useStandardController(
      ({ limit, offset, tab }) => (tab === 'logs' ? read_data({ limit, offset }) : Promise.resolve([])),
      { tab }
    )

  async function loadTemplates() {
    try {
      setTemplates(await listTemplates())
    } catch (err) {
      setTplError(err.message)
    }
  }

  useEffect(() => {
    controller.init()
    loadTemplates()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  const isLogs = tab === 'logs'
  const items = isLogs ? logs : templates
  const err = isLogs ? error : tplError
  const clearErr = isLogs ? () => setError('') : () => setTplError('')
  const activeTemplates = templates.filter((t) => t.is_active)

  function refresh() {
    if (isLogs) controller.btrefresh_click(loadLogs)
    else loadTemplates()
  }

  function openSend() {
    setSendTpl(activeTemplates[0]?.code || '')
    setSendValues({ recipient: '', nama: '', detail: '' })
    setSendOpen(true)
  }

  async function send() {
    if (!sendTpl || !sendValues.recipient.trim()) {
      toast.warning('Pilih template dan isi penerima.')
      return
    }
    setSending(true)
    try {
      const me = api.currentUser()
      await sendNotification({
        template_code: sendTpl,
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
      if (isLogs) loadLogs()
    } catch (err) {
      toast.error(err.message, { title: 'Gagal mengirim' })
    } finally {
      setSending(false)
    }
  }

  return (
    <div className="flex flex-col gap-4">
      <StandardPage
        title={nvdata?.label || 'Notification Template & Log'}
        description="Kelola template pesan (dengan variabel {{…}}) dan lihat riwayat pengiriman."
        loading={isLogs ? loading : false}
        showLoading={isLogs ? showLoading : false}
        error={err}
        errorTitle="Gagal memuat"
        onClearError={clearErr}
        onRefresh={refresh}
        onCreate={isLogs ? null : () => setShowFRM('new')}
        createLabel="New Input"
        extraActions={
          <>
            <Button size="sm" onClick={openSend}>
              <Icon name="send" className="h-4 w-4" />
              Kirim / Test
            </Button>
            {isLogs && <DownloadButton rows={logs} columns={GRID.columns} filename="notifikasi" />}
          </>
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
        emptyTitle={isLogs ? 'Belum ada pengiriman' : 'Belum ada template'}
        emptyDescription={isLogs ? 'Kirim notifikasi pertama lewat tombol Kirim / Test.' : 'Buat template pertama untuk mulai mengirim notifikasi.'}
        emptyIcon={isLogs ? 'send' : 'bell'}
        emptyAction={
          !isLogs ? (
            <Button size="sm" variant="secondary" onClick={() => setShowFRM('new')}>
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
              <Button size="sm" variant="secondary" onClick={() => setShowFRM('new')}>
                <Icon name="plus" className="h-4 w-4" />
                Template baru
              </Button>
            </div>
            <StandardGrid
              columns={[
                ...TEMPLATE_COLUMNS,
                {
                  header: 'Aksi',
                  dataIndex: '__aksi',
                  width: 90,
                  render: (t) => (
                    <span className="flex gap-1">
                      <button type="button" title="Edit" onClick={() => setShowFRM(t)} className="rounded-lg p-2 text-zinc-500 hover:bg-zinc-100 hover:text-violet-600 dark:hover:bg-zinc-800">
                        <Icon name="pencil" className="h-4 w-4" />
                      </button>
                    </span>
                  ),
                },
              ]}
              rows={templates}
              minWidth={560}
            />
          </div>
        ) : (
          <GRID rows={logs} />
        )}
      </StandardPage>

      <FRM
        open={showFRM !== null}
        initial={showFRM}
        onClose={() => setShowFRM(null)}
        onSaved={() => loadTemplates()}
      />

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
            <select
              value={sendTpl}
              onChange={(e) => setSendTpl(e.target.value)}
              className="w-full rounded-xl border border-zinc-300 bg-white px-3.5 py-2.5 text-sm text-zinc-900 focus:border-violet-500 focus:outline-none focus:ring-2 focus:ring-violet-500/30 dark:border-zinc-700 dark:bg-zinc-900 dark:text-zinc-50"
            >
              <option value="">— pilih —</option>
              {activeTemplates.map((t) => (
                <option key={t.code} value={t.code}>{t.name} ({t.channel})</option>
              ))}
            </select>
          </label>
          <StandardForm fields={SEND_FIELDS} values={sendValues} onChange={setSendValues} />
        </div>
      </Modal>
    </div>
  )
}
