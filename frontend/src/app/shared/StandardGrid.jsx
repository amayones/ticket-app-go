// Grid standar ala GRID<mod>.js langit_v2: kerangka tabel sama,
// yang beda hanya COLUMNS (items) per halaman.
// Kolom: { header, dataIndex, width, hidden, render(row) }.
// render = padanan renderer ExtJS (mis. formatDate / Badge).
export default function StandardGrid({ columns = [], rows = [], minWidth = 560 }) {
  const visible = columns.filter((c) => !c.hidden)
  return (
    <div className="overflow-x-auto rounded-xl border border-zinc-100 dark:border-zinc-800">
      <table className={`w-full min-w-[${minWidth}px] text-left text-sm`} style={{ minWidth }}>
        <thead>
          <tr className="bg-zinc-50 text-xs uppercase tracking-wide text-zinc-500 dark:bg-zinc-800/60 dark:text-zinc-400">
            {visible.map((c) => (
              <th key={c.dataIndex || c.header} className="px-3 py-2.5" style={c.width ? { width: c.width } : undefined}>
                {c.header}
              </th>
            ))}
          </tr>
        </thead>
        <tbody>
          {rows.map((row, i) => (
            <tr key={row.code || row.id || i} className="border-t border-zinc-100 dark:border-zinc-800">
              {visible.map((c) => (
                <td key={c.dataIndex || c.header} className="px-3 py-2.5 text-xs text-zinc-600 dark:text-zinc-300">
                  {c.render ? c.render(row) : (row[c.dataIndex] ?? '—')}
                </td>
              ))}
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}
