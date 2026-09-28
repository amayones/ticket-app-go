import { Button, Icon } from '../../components'

// Padanan xtype tdk-download-excel langit_v2 (toolbar -> process_download).
// Mengunduh baris yang tampil sebagai CSV (tanpa endpoint backend khusus).
export default function DownloadButton({ rows = [], columns = [], filename = 'data', label = 'Download Data' }) {
  function download() {
    const visible = columns.filter((c) => !c.hidden && c.dataIndex)
    const head = visible.map((c) => `"${String(c.header || c.dataIndex).replace(/"/g, '""')}"`).join(',')
    const lines = rows.map((r) =>
      visible
        .map((c) => {
          const v = r[c.dataIndex]
          return `"${String(v ?? '').replace(/"/g, '""')}"`
        })
        .join(',')
    )
    const blob = new Blob([[head, ...lines].join('\r\n')], { type: 'text/csv;charset=utf-8' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `${filename}.csv`
    a.click()
    URL.revokeObjectURL(url)
  }

  return (
    <Button variant="secondary" size="sm" onClick={download}>
      <Icon name="download" className="h-4 w-4" />
      {label}
    </Button>
  )
}
