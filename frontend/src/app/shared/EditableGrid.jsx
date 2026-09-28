import { Icon } from '../../components'

// Grid editable ala GRID<mod>.js Tipe A langit_v2 (plugin rowediting).
// - Baris normal: sel teks + actioncolumn edit/delete.
// - Baris editing (allowInlineEdit): sel ber-editor jadi input, aksi jadi save/cancel.
// Kolom: { header, dataIndex, width, hidden, align,
//   editor: { xtype: 'textfield'|'numberfield'|'datefield', allowBlank, maxLength },
//   render(row) }.
// Kontrol dari GRID.jsx via: editingKey, draft, setDraft, handler_rowbtn_edit_click,
// handler_rowbtn_save, handler_canceledit, handler_rowbtn_delete_click.
export default function EditableGrid({
  columns = [],
  rows = [],
  idOf = (r, i) => r.code || r.id || i,
  editingKey = null,
  draft = {},
  setDraft = () => {},
  onEdit = () => {},
  onSave = () => {},
  onCancelEdit = () => {},
  onDelete = () => {},
  minWidth = 640,
}) {
  const visible = columns.filter((c) => !c.hidden)

  function setCell(dataIndex, value) {
    setDraft({ ...draft, [dataIndex]: value })
  }

  function editorInput(c) {
    const val = draft[c.dataIndex] ?? ''
    const base = 'w-full rounded-lg border border-violet-300 bg-white px-2 py-1.5 text-xs focus:outline-none focus:ring-2 focus:ring-violet-500/30 dark:border-violet-700 dark:bg-zinc-900'
    if (c.editor.xtype === 'numberfield') {
      return (
        <input
          type="number"
          value={val}
          maxLength={c.editor.maxLength}
          onChange={(e) => setCell(c.dataIndex, e.target.value)}
          className={base}
        />
      )
    }
    if (c.editor.xtype === 'datefield') {
      return (
        <input
          type="date"
          value={String(val).slice(0, 10)}
          onChange={(e) => setCell(c.dataIndex, e.target.value)}
          className={base}
        />
      )
    }
    return (
      <input
        type="text"
        value={val}
        maxLength={c.editor.maxLength}
        onChange={(e) => setCell(c.dataIndex, e.target.value)}
        className={base}
      />
    )
  }

  return (
    <div className="overflow-x-auto rounded-xl border border-zinc-100 dark:border-zinc-800">
      <table className="w-full text-left text-sm" style={{ minWidth }}>
        <thead>
          <tr className="bg-zinc-50 text-xs uppercase tracking-wide text-zinc-500 dark:bg-zinc-800/60 dark:text-zinc-400">
            <th className="w-9 px-2 py-2.5">No</th>
            <th className="w-16 px-2 py-2.5">Aksi</th>
            {visible.map((c) => (
              <th key={c.dataIndex || c.header} className="px-3 py-2.5" style={c.width ? { width: c.width } : undefined}>
                {c.header}
              </th>
            ))}
          </tr>
        </thead>
        <tbody>
          {rows.map((row, i) => {
            const key = idOf(row, i)
            const isEditing = editingKey !== null && editingKey === key
            return (
              <tr key={key} className="border-t border-zinc-100 dark:border-zinc-800">
                <td className="px-2 py-2.5 text-xs text-zinc-400">{i + 1}</td>
                <td className="px-2 py-2.5">
                  {isEditing ? (
                    <span className="flex gap-1">
                      <button
                        type="button"
                        title="Update"
                        onClick={onSave}
                        className="rounded-lg p-1.5 text-emerald-600 hover:bg-emerald-50 dark:hover:bg-emerald-950"
                      >
                        <Icon name="check" className="h-4 w-4" />
                      </button>
                      <button
                        type="button"
                        title="Cancel"
                        onClick={onCancelEdit}
                        className="rounded-lg p-1.5 text-zinc-500 hover:bg-zinc-100 dark:hover:bg-zinc-800"
                      >
                        <Icon name="x" className="h-4 w-4" />
                      </button>
                    </span>
                  ) : (
                    <span className="flex gap-1">
                      <button
                        type="button"
                        title="Edit"
                        onClick={() => onEdit(row)}
                        className="rounded-lg p-1.5 text-zinc-500 hover:bg-zinc-100 hover:text-violet-600 dark:hover:bg-zinc-800"
                      >
                        <Icon name="pencil" className="h-4 w-4" />
                      </button>
                      <button
                        type="button"
                        title="Delete"
                        onClick={() => onDelete(row)}
                        className="rounded-lg p-1.5 text-zinc-500 hover:bg-rose-50 hover:text-rose-600 dark:hover:bg-rose-950"
                      >
                        <Icon name="trash" className="h-4 w-4" />
                      </button>
                    </span>
                  )}
                </td>
                {visible.map((c) => (
                  <td key={c.dataIndex || c.header} className="px-3 py-2.5 text-xs text-zinc-600 dark:text-zinc-300" style={c.align ? { textAlign: c.align } : undefined}>
                    {isEditing && c.editor ? editorInput(c) : c.render ? c.render(row) : (row[c.dataIndex] ?? '—')}
                  </td>
                ))}
              </tr>
            )
          })}
        </tbody>
      </table>
    </div>
  )
}
