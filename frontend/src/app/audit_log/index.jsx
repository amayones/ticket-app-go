// ===== TEMPLATE SERAGAM app/* (ala <mod>.js langit_v2) =====
// Shell + controller standar; yang khas tabel ini hanya COLUMNS (./columns.js).
// Tabel DB: CPAUDITLOG via ./api.js listAudit.
import { Button, SELECT_CLASS, StandardGrid, StandardPage, useStandardController, useState } from '../shared/all.js'
import { listAudit } from './api.js'
import { ACTIONS, ENTITIES, AUDIT_COLUMNS } from './columns.jsx'

const FEATURES = { header: true, refresh: true, filter: true, tabs: false, create: false, edit: false, remove: false, pagination: true, empty: true, error: true, confirmDialog: false, extraActions: false }

const CONFIG = {
  title: 'Audit Log',
  description: 'Siapa melakukan apa, kapan, dari IP mana. Tercatat otomatis untuk login, CRUD user/role, sesi, dan notifikasi.',
  errorTitle: 'Gagal memuat audit',
  emptyTitle: 'Belum ada jejak audit',
  emptyDescription: 'Sesuaikan filter atau lakukan aksi (mis. login) agar tercatat di sini.',
}

export const meta = { label: 'Audit Log', icon: 'list', order: 5 }

export default function Audit() {
  const [action, setAction] = useState('')
  const [entity, setEntity] = useState('')
  const [actor, setActor] = useState('')
  const [actorInput, setActorInput] = useState('')

  const { rows: logs, offset, setOffset, loading, showLoading, error, setError, load } =
    useStandardController(listAudit, { action, entity, actor })

  function applyFilter(e) {
    e.preventDefault()
    setOffset(0)
    setActor(actorInput.trim())
  }

  function reset() {
    setAction('')
    setEntity('')
    setActor('')
    setActorInput('')
    setOffset(0)
  }

  return (
    <StandardPage
      title={CONFIG.title}
      description={CONFIG.description}
      features={FEATURES}
      loading={loading}
      showLoading={showLoading}
      error={error}
      errorTitle={CONFIG.errorTitle}
      onClearError={() => setError('')}
      onRefresh={load}
      filterBar={
        <form onSubmit={applyFilter} className="mb-4 flex flex-wrap items-end gap-2">
          <label className="flex flex-col gap-1 text-xs font-medium text-zinc-500">
            Aksi
            <select value={action} onChange={(e) => { setAction(e.target.value); setOffset(0) }} className={SELECT_CLASS}>
              {ACTIONS.map((a) => (
                <option key={a} value={a}>{a === '' ? 'Semua' : a}</option>
              ))}
            </select>
          </label>
          <label className="flex flex-col gap-1 text-xs font-medium text-zinc-500">
            Entitas
            <select value={entity} onChange={(e) => { setEntity(e.target.value); setOffset(0) }} className={SELECT_CLASS}>
              {ENTITIES.map((a) => (
                <option key={a} value={a}>{a === '' ? 'Semua' : a}</option>
              ))}
            </select>
          </label>
          <label className="flex flex-col gap-1 text-xs font-medium text-zinc-500">
            Kode pelaku
            <input
              value={actorInput}
              onChange={(e) => setActorInput(e.target.value)}
              placeholder="USR-…"
              className={`${SELECT_CLASS} w-32 font-mono`}
            />
          </label>
          <Button size="sm" type="submit">
            Filter
          </Button>
          <Button size="sm" variant="ghost" type="button" onClick={reset}>
            Reset
          </Button>
        </form>
      }
      items={logs}
      emptyTitle={CONFIG.emptyTitle}
      emptyDescription={CONFIG.emptyDescription}
      offset={offset}
      onPage={setOffset}
    >
      <StandardGrid columns={AUDIT_COLUMNS} rows={logs} minWidth={640} />
    </StandardPage>
  )
}
