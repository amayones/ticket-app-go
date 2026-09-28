// GRID User Account — padanan GRID<mod>.js langit_v2 (body custom kartu,
// seperti modul tree mgroup_user: bukan tabel, tapi fungsi GRID tetap sama).
// Handler nama SAMA: handler_rowbtn_edit_click, handler_btnew_click,
// handler_rowbtn_delete_click, handler_rowbtn_logout_click.
import { forwardRef, useImperativeHandle } from 'react'
import { Avatar, Badge, Icon, Tooltip, useMessageBox, useToast } from '../shared/all.js'
import { process_delete, logoutAll } from './api.js'

const GRID = forwardRef(function GRID({ rows, me, reload, openFRM, openEdit, onAccountDeleted }, ref) {
  const toast = useToast()
  const { MessageBox, dialog } = useMessageBox()

  useImperativeHandle(ref, () => ({
    // Varian A: controller.btnew_click delegasi ke sini — buka FRM input baru.
    handler_btnew_click: () => openFRM(),
  }))

  function handler_rowbtn_edit_click(user) {
    openEdit(user)
  }

  async function handler_rowbtn_delete_click(user) {
    try {
      const result = await MessageBox.delete({
        headline: 'Konfirmasi Delete Data',
        details: [
          { label: 'Username', value: `@${user.username}` },
          { label: 'Email', value: user.email },
        ],
        request: () => process_delete(user.code),
      })
      if (!result) return
      toast.success(`Akun @${user.username} dihapus.`)
      if (me?.code === user.code) {
        onAccountDeleted?.()
        return
      }
      reload()
    } catch (ex) {
      toast.error(ex.message, { title: 'Error' })
    }
  }

  async function handler_rowbtn_logout_click(user) {
    try {
      const result = await MessageBox.delete({
        headline: 'Keluarkan semua sesi?',
        details: [{ label: 'Username', value: `@${user.username}` }],
        request: () => logoutAll(user.code),
      })
      if (!result) return
      toast.success('Semua sesi berhasil dikeluarkan. Silakan login ulang bila diperlukan.')
      reload()
    } catch (ex) {
      toast.error(ex.message, { title: 'Error' })
    }
  }

  return (
    <>
      <ul className="flex flex-col gap-2.5">
        {rows.map((u) => {
          const isMe = me?.code === u.code
          return (
            <li
              key={u.code}
              className="flex items-center gap-3 rounded-xl border border-zinc-100 bg-zinc-50/60 p-3 transition hover:border-violet-200 hover:bg-violet-50/50 dark:border-zinc-800 dark:bg-zinc-900/40 dark:hover:border-violet-800"
            >
              <Avatar name={u.username} />
              <div className="min-w-0 flex-1">
                <p className="flex flex-wrap items-center gap-2 text-sm font-semibold text-zinc-900 dark:text-zinc-50">
                  <span className="truncate">@{u.username}</span>
                  {isMe && <Badge tone="brand">Anda</Badge>}
                  {u.role_code && (
                    <Badge tone={u.role_code === 'ADMIN' ? 'danger' : 'neutral'}>{u.role_code}</Badge>
                  )}
                </p>
                <p className="truncate text-xs text-zinc-500 dark:text-zinc-400">
                  <span className="font-mono">{u.code}</span> · {u.email}
                </p>
              </div>
              <div className="flex shrink-0 items-center gap-1.5">
                <Tooltip label="Edit profil">
                  <button
                    type="button"
                    aria-label={`Edit ${u.username}`}
                    onClick={() => handler_rowbtn_edit_click(u)}
                    className="rounded-lg p-2 text-zinc-500 transition-colors hover:bg-white hover:text-zinc-900 hover:shadow-sm dark:hover:bg-zinc-800 dark:hover:text-zinc-100"
                  >
                    <Icon name="pencil" className="h-4 w-4" />
                  </button>
                </Tooltip>
                <Tooltip label="Keluarkan semua sesi">
                  <button
                    type="button"
                    aria-label={`Keluarkan semua sesi ${u.username}`}
                    onClick={() => handler_rowbtn_logout_click(u)}
                    className="rounded-lg p-2 text-zinc-500 transition-colors hover:bg-white hover:text-zinc-900 hover:shadow-sm dark:hover:bg-zinc-800 dark:hover:text-zinc-100"
                  >
                    <Icon name="logout" className="h-4 w-4" />
                  </button>
                </Tooltip>
                <Tooltip label="Hapus akun">
                  <button
                    type="button"
                    aria-label={`Hapus ${u.username}`}
                    onClick={() => handler_rowbtn_delete_click(u)}
                    className="rounded-lg p-2 text-zinc-500 transition-colors hover:bg-rose-50 hover:text-rose-600 hover:shadow-sm dark:hover:bg-rose-950 dark:hover:text-rose-300"
                  >
                    <Icon name="trash" className="h-4 w-4" />
                  </button>
                </Tooltip>
              </div>
            </li>
          )
        })}
      </ul>
      {dialog}
    </>
  )
})

export default GRID
