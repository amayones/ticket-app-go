import { Alert, Button, Card, CardTitle, EmptyState, Icon, Pagination, SkeletonRows } from '../../components'
import { PAGE_SIZE, SKELETON_ROWS } from './config.js'

// Pembungkus return seragam: header + refresh + error + loading/empty/isi + pagination.
// Dipakai halaman list sederhana; halaman kompleks boleh tidak pakai
// (cukup import dari all.js agar struktur header tetap sama).
export default function PageShell({
  title,
  description,
  actions = null,
  showRefresh = true,
  loading = false,
  showLoading = false,
  error = '',
  errorTitle = 'Gagal memuat data',
  onClearError = null,
  onRefresh = null,
  skeletonRows = SKELETON_ROWS,
  showEmpty = true,
  emptyTitle = 'Belum ada data',
  emptyDescription = 'Data kosong pada halaman ini.',
  emptyAction = null,
  items = [],
  renderItems = null,
  showPagination = true,
  offset = 0,
  onPage = null,
  children = null,
}) {
  const busy = showLoading || loading
  return (
    <Card>
      <div className="mb-4 flex flex-wrap items-center justify-between gap-3">
        <CardTitle description={description}>{title}</CardTitle>
        <div className="flex gap-2">
          {actions}
          {showRefresh && onRefresh && (
            <Button variant="secondary" size="sm" onClick={onRefresh} loading={loading}>
              <Icon name="refresh" className="h-4 w-4" />
              Muat ulang
            </Button>
          )}
        </div>
      </div>

      {error && (
        <Alert tone="error" title={errorTitle} closable onClose={onClearError || undefined} className="mb-4">
          {error}
        </Alert>
      )}

      {busy ? (
        <SkeletonRows rows={skeletonRows} />
      ) : children ? (
        children
      ) : items.length === 0 ? (
        showEmpty ? (
          <EmptyState title={emptyTitle} description={emptyDescription} action={emptyAction} />
        ) : null
      ) : (
        renderItems?.(items)
      )}

      {showPagination && !busy && items.length > 0 && onPage && (
        <div className="mt-4 border-t border-zinc-100 pt-4 dark:border-zinc-800">
          <Pagination
            offset={offset}
            limit={PAGE_SIZE}
            count={items.length}
            hasMore={items.length === PAGE_SIZE}
            loading={loading}
            onPage={onPage}
          />
        </div>
      )}
    </Card>
  )
}
