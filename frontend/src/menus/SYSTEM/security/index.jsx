import { useCallback, useEffect, useState } from 'react'
import { securitySummary } from './api.js'
import { listAudit } from '../audit/api.js'
import { Alert, Badge, Button, Card, CardTitle, Icon, Skeleton, useSmoothLoading, useToast } from '../../../components'

export const meta = { label: 'Security Center', icon: 'shield', order: 5 }

function StatCard({ icon, label, value, tone }) {
  const tones = {
    danger: 'bg-rose-100 text-rose-600 dark:bg-rose-950 dark:text-rose-300',
    neutral: 'bg-zinc-100 text-zinc-600 dark:bg-zinc-800 dark:text-zinc-300',
  }
  return (
    <div className="flex items-center gap-3 rounded-2xl border border-zinc-200 bg-white p-4 dark:border-zinc-800 dark:bg-zinc-900">
      <span className={`flex h-11 w-11 shrink-0 items-center justify-center rounded-xl ${tones[tone] || tones.neutral}`}>
        <Icon name={icon} className="h-5 w-5" />
      </span>
      <div className="min-w-0">
        <p className="truncate text-2xl font-bold text-zinc-900 dark:text-zinc-50">{value}</p>
        <p className="truncate text-xs text-zinc-500 dark:text-zinc-400">{label}</p>
      </div>
    </div>
  )
}

export default function Security() {
  const toast = useToast()
  const [summary, setSummary] = useState(null)
  const [recent, setRecent] = useState([])
  const [loading, setLoading] = useState(true)
  const showLoading = useSmoothLoading(loading)
  const [error, setError] = useState('')

  const load = useCallback(async () => {
    setLoading(true)
    setError('')
    try {
      const [s, a] = await Promise.all([
        securitySummary(),
        listAudit({ limit: 8, offset: 0 }),
      ])
      setSummary(s)
      setRecent(Array.isArray(a) ? a : [])
    } catch (err) {
      setError(err.message)
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    load()
  }, [load])

  return (
    <div className="flex flex-col gap-4">
      <Card>
        <div className="mb-4 flex flex-wrap items-center justify-between gap-3">
          <CardTitle description="Kesehatan keamanan 24 jam terakhir dalam sekali lihat.">
            Security Center
          </CardTitle>
          <Button variant="secondary" size="sm" onClick={load} loading={loading}>
            <Icon name="refresh" className="h-4 w-4" />
            Muat ulang
          </Button>
        </div>
        {error && (
          <Alert tone="error" title="Gagal memuat ringkasan" closable onClose={() => setError('')} className="mb-4">
            {error}
          </Alert>
        )}
        {showLoading || !summary ? (
          <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
            {[0, 1, 2, 3, 4, 5].map((i) => (
              <Skeleton key={i} className="h-[76px] rounded-2xl" />
            ))}
          </div>
        ) : (
          <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
            <StatCard icon="users" label="Total pengguna" value={summary.total_users} tone="neutral" />
            <StatCard icon="shield" label="Total role" value={summary.total_roles} tone="neutral" />
            <StatCard icon="key" label="Sesi aktif" value={summary.active_sessions} tone="neutral" />
            <StatCard icon="list" label="Aksi audit 24 jam" value={summary.audit_last_24h} tone="neutral" />
            <StatCard
              icon="warning"
              label="Error sistem 24 jam"
              value={summary.errors_last_24h}
              tone={summary.errors_last_24h > 0 ? 'danger' : 'neutral'}
            />
            <StatCard icon="clock" label="Perlu perhatian" value={summary.errors_last_24h > 0 ? 'Ya' : 'Tidak'} tone={summary.errors_last_24h > 0 ? 'danger' : 'neutral'} />
          </div>
        )}
      </Card>

      <Card>
        <CardTitle description="8 aktivitas terakhir dari semua user.">Aktivitas terkini</CardTitle>
        {showLoading ? (
          <Skeleton className="h-24" />
        ) : recent.length === 0 ? (
          <p className="text-sm text-zinc-500">Belum ada aktivitas tercatat.</p>
        ) : (
          <ul className="flex flex-col gap-2">
            {recent.map((l) => (
              <li key={l.code} className="flex items-center gap-3 rounded-xl bg-zinc-50 px-3 py-2 text-sm dark:bg-zinc-800/50">
                <Badge tone="neutral">{l.action}</Badge>
                <span className="min-w-0 flex-1 truncate text-xs text-zinc-600 dark:text-zinc-300">
                  {l.actor_code || 'sistem'} · {l.detail || l.entity}
                </span>
                <button
                  type="button"
                  onClick={() => toast.info(`${l.entity}${l.entity_code ? ' ' + l.entity_code : ''} — ${l.detail || 'tanpa detail'}`, { title: l.action })}
                  className="shrink-0 text-xs font-semibold text-violet-600 hover:underline dark:text-violet-300"
                >
                  Detail
                </button>
              </li>
            ))}
          </ul>
        )}
      </Card>

      <Card>
        <CardTitle description="Kebijakan yang ditegakkan backend secara otomatis.">
          Kebijakan keamanan aktif
        </CardTitle>
        <ul className="grid gap-2 text-sm sm:grid-cols-2">
          {[
            ['Access token pendek + refresh berotasi (lihat ACCESS_TOKEN_MINUTES)', true],
            ['Refresh token disimpan sebagai hash SHA-256', true],
            ['Maksimal 5 sesi per user (sesi tertua digusur)', true],
            ['Rate-limit: login 5/mnt, register 10/mnt, refresh 30/mnt', true],
            ['Password bcrypt + batas 8–72 karakter', true],
            ['JWT HS256 + issuer/audience check', true],
          ].map(([label]) => (
            <li key={label} className="flex items-start gap-2 text-zinc-600 dark:text-zinc-300">
              <span className="mt-0.5 text-emerald-500">
                <Icon name="check" className="h-4 w-4" />
              </span>
              {label}
            </li>
          ))}
        </ul>
      </Card>
    </div>
  )
}
