import { useState } from 'react'
import { ConfirmDialog } from '../../components'

// Padanan COMP.MessageBox.create/update/delete langit_v2.
// Pemakaian (sama persis alurnya):
//   const result = await MessageBox.create({ headline, details, request: () => process_create(dtval) })
//   if (!result) return // dibatalkan user
//   reload()
// request yang throw (gagal HTTP) ditangkap try/catch pemanggil ala TipToast.
// Render { dialog } sekali di return menu.
export function useMessageBox() {
  const [state, setState] = useState(null)

  function ask({ headline, description = '', details = [], confirmLabel = 'Ya', danger = false }) {
    return new Promise((resolve) => {
      setState({
        headline,
        description,
        details,
        confirmLabel,
        danger,
        onConfirm: () => {
          setState(null)
          resolve(true)
        },
        onCancel: () => {
          setState(null)
          resolve(false)
        },
      })
    })
  }

  async function run(kind, { headline, description = '', details = [], request }) {
    const defaults =
      kind === 'create'
        ? { headline: headline || 'Konfirmasi Simpan Data', confirmLabel: 'Ya, simpan', danger: false }
        : kind === 'update'
          ? { headline: headline || 'Konfirmasi Update Data', confirmLabel: 'Ya, update', danger: false }
          : { headline: headline || 'Konfirmasi Delete Data', confirmLabel: 'Ya, hapus', danger: true }
    const ok = await ask({ ...defaults, headline: headline || defaults.headline, description, details })
    if (!ok) return null
    const data = await request()
    return { data }
  }

  const MessageBox = {
    create: (o) => run('create', o),
    update: (o) => run('update', o),
    delete: (o) => run('delete', o),
  }

  const dialog = state ? (
    <ConfirmDialog
      open
      title={state.headline}
      message={
        <span className="flex flex-col gap-2">
          {state.description ? <span>{state.description}</span> : null}
          {state.details.map((d) => (
            <span key={d.label} className="flex justify-between gap-4 text-sm">
              <span className="text-zinc-500">{d.label}</span>
              <span className="font-semibold text-zinc-900 dark:text-zinc-50">{String(d.value ?? '—')}</span>
            </span>
          ))}
        </span>
      }
      confirmLabel={state.confirmLabel}
      danger={state.danger}
      onConfirm={state.onConfirm}
      onCancel={state.onCancel}
    />
  ) : null

  return { dialog, ask, MessageBox }
}
