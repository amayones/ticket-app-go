// View News — daftar berita + halaman baca (publik, tanpa login).
// Slug artikel disimpan di hash (#/news/<slug>) supaya bisa dibagikan.
// Controller: ./controller.js (6 fungsi). Store: read_data + newsDetail.
import { useEffect, useRef, useState } from 'react'
import {
  Alert,
  Badge,
  Button,
  Card,
  CardTitle,
  SafeImage,
  Skeleton,
  StandardPage,
  formatTime,
  useDashboard,
  SELECT_CLASS,
  EventCard,
  NewsCard,
  goEvent,
  hashParam,
  replaceHash,
} from '../shared/all.js'
import { newsDetail, read_data } from './api.js'
import { controller } from './controller.js'

export const meta = { label: 'News', icon: 'list', order: 2 }

const SORTS = [
  { value: '', label: 'Terbaru' },
  { value: 'popular', label: 'Terpopuler' },
]

const PAGE_LIMIT = 12

function useDebounced(value, delay = 400) {
  const [applied, setApplied] = useState(value)
  useEffect(() => {
    const id = setTimeout(() => setApplied(value), delay)
    return () => clearTimeout(id)
  }, [value, delay])
  return applied
}

function ArticleBody({ html }) {
  if (!html) return <p className="text-sm text-zinc-500">Isi berita belum tersedia.</p>
  return (
    <div
      className="text-sm leading-relaxed text-zinc-700 dark:text-zinc-200 [&_p]:mb-3"
      dangerouslySetInnerHTML={{ __html: html }}
    />
  )
}

export default function News({ nvdata, onNavigate }) {
  const [category, setCategory] = useState('')
  const [sort, setSort] = useState('')
  const [query, setQuery] = useState('')
  const [limit, setLimit] = useState(PAGE_LIMIT)
  const [slug, setSlug] = useState(() => hashParam('news'))
  const [article, setArticle] = useState(null)
  const [articleLoading, setArticleLoading] = useState(false)
  const [articleError, setArticleError] = useState('')
  const articleAbort = useRef(null)
  const appliedQuery = useDebounced(query)

  const filters = { category, q: appliedQuery, sort, limit, offset: 0 }
  const { data, loading, showLoading, error, setError, load } = useDashboard(read_data, filters)

  useEffect(() => {
    controller.init()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  // Pindah antar berita terkait tanpa lewat sidebar (hash berubah, view sama).
  useEffect(() => {
    function onHash() {
      const s = hashParam('news')
      if (s) setSlug((prev) => (prev === s ? prev : s))
    }
    window.addEventListener('hashchange', onHash)
    return () => window.removeEventListener('hashchange', onHash)
  }, [])

  // Halaman baca: muat artikel saat slug berubah (dari kartu atau tautan).
  useEffect(() => {
    if (!slug) {
      setArticle(null)
      setArticleError('')
      return
    }
    articleAbort.current?.abort()
    const ctrl = new AbortController()
    articleAbort.current = ctrl
    setArticleLoading(true)
    setArticleError('')
    newsDetail(slug, ctrl.signal)
      .then((detail) => {
        if (!ctrl.signal.aborted) setArticle(detail)
      })
      .catch((err) => {
        if (err && err.name === 'AbortError') return
        if (!ctrl.signal.aborted) {
          setArticle(null)
          setArticleError(err.message)
        }
      })
      .finally(() => {
        if (!ctrl.signal.aborted) setArticleLoading(false)
      })
    return () => ctrl.abort()
  }, [slug])

  useEffect(() => () => articleAbort.current?.abort(), [])

  function handler_open(item) {
    setSlug(item.slug)
    replaceHash(`#/news/${encodeURIComponent(item.slug)}`)
  }

  function handler_back() {
    setSlug('')
    setArticle(null)
    replaceHash('#/news')
  }

  function handler_more() {
    setLimit((l) => l + PAGE_LIMIT)
  }

  const categories = data?.categories || []
  const items = data?.items || []
  const pageItems = [{ items }]

  // Tampilan baca artikel.
  if (slug) {
    return (
      <div className="flex flex-col gap-4">
        <Card>
          <Button variant="secondary" size="sm" onClick={handler_back}>
            Kembali ke daftar
          </Button>
          {articleLoading ? (
            <div className="mt-4">
              <Skeleton className="h-8 w-3/4" />
              <Skeleton className="mt-2 h-44" />
            </div>
          ) : articleError ? (
            <div className="mt-4">
              <Alert tone="error" title="Gagal memuat berita" closable onClose={() => setArticleError('')}>
                {articleError}
              </Alert>
              <div className="mt-3">
                <Button
                  variant="secondary"
                  size="sm"
                  onClick={() => {
                    const s = slug
                    setSlug('')
                    setTimeout(() => setSlug(s), 0)
                  }}
                >
                  Coba lagi
                </Button>
              </div>
            </div>
          ) : article ? (
            <article className="mt-4 flex min-w-0 flex-col gap-3">
              <span>
                <Badge tone="info">{article.category}</Badge>
              </span>
              <h1 className="text-xl font-bold text-zinc-900 dark:text-zinc-50">{article.title}</h1>
              <p className="text-xs text-zinc-500 dark:text-zinc-400">
                {formatTime(article.published_at)} · {Number(article.view_count || 0).toLocaleString('id-ID')} dibaca
              </p>
              <SafeImage src={article.image_url} alt={article.title} className="w-full rounded-xl object-cover" />
              <ArticleBody html={article.body} />
              {article.event_code ? (
                <div className="mt-2 rounded-2xl border border-zinc-200 p-4 dark:border-zinc-800">
                  <CardTitle description="Disebut dalam berita ini.">Event terkait</CardTitle>
                  <Button size="sm" onClick={() => goEvent(onNavigate, article.event_code)}>
                    {article.event_title || article.event_code}
                  </Button>
                </div>
              ) : null}
            </article>
          ) : null}
        </Card>

        {article && (article.related_news || []).length > 0 && (
          <Card>
            <CardTitle description="Masih seputar kategori yang sama.">Berita terkait</CardTitle>
            <div className="grid gap-3 sm:grid-cols-2">
              {article.related_news.map((n) => (
                <NewsCard key={n.code} news={n} onOpen={handler_open} />
              ))}
            </div>
          </Card>
        )}

        {article && article.related_event && (
          <Card>
            <CardTitle description="Event yang disebut dalam berita ini.">Event terkait</CardTitle>
            <EventCard
              event={article.related_event}
              onOpen={(item) => goEvent(onNavigate, item.code)}
            />
          </Card>
        )}
      </div>
    )
  }

  return (
    <div className="flex flex-col gap-4">
      <StandardPage
        title={nvdata?.label || 'News'}
        description="Kabar terbaru seputar event."
        loading={loading}
        showLoading={showLoading}
        error={error}
        errorTitle="Gagal memuat berita"
        onClearError={() => setError('')}
        onRefresh={() => controller.btrefresh_click(load)}
        filterBar={
          <form onSubmit={(e) => e.preventDefault()} className="mb-4 flex flex-wrap items-end gap-2">
            <label className="flex flex-col gap-1 text-xs font-medium text-zinc-500">
              Kategori
              <select
                value={category}
                onChange={(e) => {
                  setCategory(e.target.value)
                  setLimit(PAGE_LIMIT)
                }}
                className={SELECT_CLASS}
              >
                <option value="">Semua kategori</option>
                {categories.map((c) => (
                  <option key={c.code} value={c.code}>
                    {c.name}
                  </option>
                ))}
              </select>
            </label>
            <label className="flex flex-col gap-1 text-xs font-medium text-zinc-500">
              Cari judul
              <input
                value={query}
                onChange={(e) => {
                  setQuery(e.target.value)
                  setLimit(PAGE_LIMIT)
                }}
                placeholder="Ketik judul…"
                className={SELECT_CLASS}
              />
            </label>
            <label className="flex flex-col gap-1 text-xs font-medium text-zinc-500">
              Urutan
              <select value={sort} onChange={(e) => setSort(e.target.value)} className={SELECT_CLASS}>
                {SORTS.map((s) => (
                  <option key={s.label} value={s.value}>
                    {s.label}
                  </option>
                ))}
              </select>
            </label>
          </form>
        }
        items={pageItems}
        emptyTitle="Belum ada berita"
        emptyDescription="Coba kata kunci atau kategori lain."
        emptyAction={
          <Button variant="secondary" size="sm" onClick={() => controller.btrefresh_click(load)}>
            Muat ulang
          </Button>
        }
      >
        <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
          {items.map((n) => (
            <NewsCard key={n.code} news={n} onOpen={handler_open} />
          ))}
        </div>
        {items.length >= limit && (
          <div className="mt-4 text-center">
            <Button variant="secondary" size="sm" onClick={handler_more} loading={loading}>
              Muat lagi
            </Button>
          </div>
        )}
      </StandardPage>

      {error && (
        <Alert tone="error" title="Gagal memuat berita" closable onClose={() => setError('')}>
          {error}
        </Alert>
      )}
    </div>
  )
}
