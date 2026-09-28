import { Alert, Button, Card, CardTitle, EmptyState, Icon, Pagination, SkeletonRows } from '../../components'
import { PAGE_SIZE, SKELETON_ROWS } from './config.js'

// Shell standar ala <mod>.js langit_v2: toolbar (refresh/new) + items + pagination.
// Aturan isi (sama untuk semua halaman):
// - Shell mengurus: header, tabs, filter, error, skeleton, empty, pagination.
// - children = HANYA tampilan data (grid/list/kartu), tanpa cek loading/empty.
// - items = baris data; kosong -> EmptyState otomatis. Modal/Confirm di luar.
// Blok mati cukup set flag false — return tidak perlu diubah.
export default function StandardPage({
  title,
  description,
  features = {},
  loading = false,
  showLoading = false,
  error = '',
  errorTitle = 'Gagal memuat data',
  onClearError = null,
  onRefresh = null,
  onCreate = null,
  createLabel = 'Tambah baru',
  extraActions = null,
  filterBar = null,
  tabsBar = null,
  skeletonRows = SKELETON_ROWS,
  emptyTitle = 'Belum ada data',
  emptyDescription = 'Data kosong pada halaman ini.',
  emptyIcon = undefined,
  emptyAction = null,
  items = [],
  children = null,
  offset = 0,
  onPage = null,
  footer = null,
}) {
  const F = { header: true, refresh: true, filter: true, tabs: true, create: true, pagination: true, empty: true, error: true, ...features }
  const busy = showLoading || loading

  return (
    <div className="flex flex-col gap-4">
      <Card>
        {F.header && (
          <div className="mb-4 flex flex-wrap items-center justify-between gap-3">
            <CardTitle description={description}>{title}</CardTitle>
            <div className="flex gap-2">
              {F.refresh && onRefresh && (
                <Button variant="secondary" size="sm" onClick={onRefresh} loading={loading}>
                  <Icon name="refresh" className="h-4 w-4" />
                  Muat ulang
                </Button>
              )}
              {F.create && onCreate && (
                <Button size="sm" onClick={onCreate}>
                  <Icon name="plus" className="h-4 w-4" />
                  {createLabel}
                </Button>
              )}
              {extraActions}
            </div>
          </div>
        )}

        {F.tabs && tabsBar}
        {F.filter && filterBar}

        {F.error && error && (
          <Alert tone="error" title={errorTitle} closable onClose={onClearError || undefined} className="mb-4">
            {error}
          </Alert>
        )}

        {busy ? (
          <SkeletonRows rows={skeletonRows} />
        ) : items.length === 0 ? (
          F.empty ? (
            <EmptyState icon={emptyIcon} title={emptyTitle} description={emptyDescription} action={emptyAction} />
          ) : null
        ) : (
          children
        )}

        {F.pagination && !busy && items.length > 0 && onPage && (
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

        {footer}
      </Card>
    </div>
  )
}
