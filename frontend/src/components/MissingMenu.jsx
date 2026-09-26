import { useState } from 'react'
import Alert from './Alert.jsx'
import Button from './Button.jsx'
import Card, { CardTitle } from './Card.jsx'
import Icon from './icons.jsx'
import { suggestPath } from '../app/registry.js'

// Halaman 404 pemandu: menu CHILD sudah ada di database (CPMENU + grant)
// tapi folder frontend app/<mcontrol>/ belum dibuat. Memberi tahu
// programmer persis di mana harus membuat file, bukan sekadar "Not found".
export default function MissingMenu({ entry }) {
  const [copied, setCopied] = useState(false)
  if (!entry) return null
  const path = suggestPath(entry)
  const copyCmd = `cp -r tutorial/templates/frontend-menu ${path}`
  const folder = entry.mcontrol || entry.key

  async function copy() {
    try {
      await navigator.clipboard.writeText(copyCmd)
      setCopied(true)
      setTimeout(() => setCopied(false), 2000)
    } catch {
      setCopied(false)
    }
  }

  return (
    <div className="flex flex-col gap-4">
      <Card>
        <CardTitle
          description={`Permission ${entry.code} sudah diberi ke role Anda, tapi halamannya belum ada.`}
        >
          <span className="font-mono">404</span> — Menu belum dibuat
        </CardTitle>

        <Alert tone="warning" title="Folder menu tidak ditemukan" className="mb-4 mt-4">
          Buat folder <span className="font-mono font-semibold">{path}</span> berisi{' '}
          <span className="font-mono">index.jsx</span> + <span className="font-mono">api.js</span>,
          lalu rebuild frontend.
        </Alert>

        <ol className="flex flex-col gap-2.5 text-sm text-zinc-700 dark:text-zinc-200">
          <li className="flex items-start gap-2">
            <Icon name="check" className="mt-0.5 h-4 w-4 shrink-0 text-emerald-500" />
            <span>
              Copy template:{' '}
              <code className="rounded bg-zinc-100 px-1.5 py-0.5 font-mono text-xs dark:bg-zinc-800">
                {copyCmd}
              </code>{' '}
              <Button variant="secondary" size="sm" onClick={copy} className="ml-1">
                {copied ? 'Disalin!' : 'Salin'}
              </Button>
            </span>
          </li>
          <li className="flex items-start gap-2">
            <Icon name="check" className="mt-0.5 h-4 w-4 shrink-0 text-emerald-500" />
            <span>
              Di <span className="font-mono">index.jsx</span> isi{' '}
              <span className="font-mono">export const meta</span> (label: {entry.label}) dan{' '}
              <span className="font-mono">export default</span>. MCONTROL{' '}
              <span className="font-mono">{folder}</span> = nama folder; judul tampil diambil dari{' '}
              <span className="font-mono">LABEL</span> di CPMENU ({entry.label}).
            </span>
          </li>
          <li className="flex items-start gap-2">
            <Icon name="check" className="mt-0.5 h-4 w-4 shrink-0 text-emerald-500" />
            <span>
              Section <span className="font-mono">{entry.module}</span> dari kolom{' '}
              <span className="font-mono">MODULE</span> di tabel{' '}
              <span className="font-mono">CPMENU</span>. Folder menu datar:{' '}
              <span className="font-mono">app/&lt;mcontrol&gt;/</span> (tanpa folder modul).
              {entry.parent && (
                <>
                  {' '}Menu ini anak dari <span className="font-mono">{entry.parent}</span>.
                </>
              )}
            </span>
          </li>
          <li className="flex items-start gap-2">
            <Icon name="check" className="mt-0.5 h-4 w-4 shrink-0 text-emerald-500" />
            <span>
              Jalankan <span className="font-mono">npm run build</span> di folder{' '}
              <span className="font-mono">frontend</span>, restart backend, lalu refresh — menu asli
              otomatis menggantikan halaman ini.
            </span>
          </li>
        </ol>
      </Card>
    </div>
  )
}
