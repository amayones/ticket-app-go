// View Home — halaman awal discovery (publik, tanpa login).
// Banner iklan (karusel + Iklan + impression/klik), promo pilihan, berita
// terbaru, event populer & terdekat, pintasan kategori, pemilih kota.
// Controller: ./controller.js (6 fungsi). Store: read_data.
import { useEffect, useRef, useState } from 'react'
import {
  Alert,
  Badge,
  Button,
  Card,
  CardTitle,
  EmptyState,
  Icon,
  SafeImage,
  Skeleton,
  StandardPage,
  useDashboard,
  useToast,
  SELECT_CLASS,
  EventCard,
  NewsCard,
  PromoCard,
  goDetail,
  goEvent,
  replaceHash,
} from '../shared/all.js'
import { read_data, recordClick, recordImpression } from './api.js'
import { controller } from './controller.js'

export const meta = { label: 'Home', icon: 'terminal', order: 1 }

function cityPref() {
  try {
    return localStorage.getItem('tikto-city') || ''
  } catch {
    return ''
  }
}

// Karusel banner: satu tampil, panah + titik, impression per banner yang
// tampil, klik mencatat + membuka tautan di tab baru.
function BannerCarousel({ banners }) {
  const toast = useToast()
  const [index, setIndex] = useState(0)
  const seenRef = useRef(new Set())

  useEffect(() => {
    const b = banners[index]
    if (!b || seenRef.current.has(b.code)) return
    seenRef.current.add(b.code)
    recordImpression(b.code)
  }, [banners, index])

  if (banners.length === 0) return null
  const active = banners[Math.min(index, banners.length - 1)]

  async function handler_click(banner) {
    await recordClick(banner.code)
    if (banner.target_url) {
      try {
        window.open(banner.target_url, '_blank', 'noopener,noreferrer')
      } catch {
        toast.info(banner.target_url, { title: banner.title })
      }
    }
  }

  function move(dir) {
    setIndex((i) => (i + dir + banners.length) % banners.length)
  }

  return (
    <div className="relative overflow-hidden rounded-2xl border border-zinc-200 dark:border-zinc-800">
      <button type="button" onClick={() => handler_click(active)} className="block w-full text-left" aria-label={active.title}>
        <SafeImage src={active.image_url} alt={active.title} className="aspect-[21/9] w-full object-cover" />
      </button>
      <span className="absolute left-2 top-2">
        <Badge tone="warning">Iklan</Badge>
      </span>
      {banners.length > 1 && (
        <>
          <button
            type="button"
            aria-label="Banner sebelumnya"
            onClick={() => move(-1)}
            className="absolute left-2 top-1/2 -translate-y-1/2 rounded-full bg-black/40 p-1.5 text-white hover:bg-black/60"
          >
            <Icon name="chevronLeft" className="h-4 w-4" />
          </button>
          <button
            type="button"
            aria-label="Banner berikutnya"
            onClick={() => move(1)}
            className="absolute right-2 top-1/2 -translate-y-1/2 rounded-full bg-black/40 p-1.5 text-white hover:bg-black/60"
          >
            <Icon name="chevronRight" className="h-4 w-4" />
          </button>
          <span className="absolute bottom-2 left-1/2 flex -translate-x-1/2 gap-1.5">
            {banners.map((b, i) => (
              <button
                key={b.code}
                type="button"
                aria-label={`Banner ${i + 1}`}
                onClick={() => setIndex(i)}
                className={`h-2 rounded-full transition-all ${i === index ? 'w-6 bg-white' : 'w-2 bg-white/50'}`}
              />
            ))}
          </span>
        </>
      )}
    </div>
  )
}

export default function Home({ nvdata, onNavigate }) {
  const toast = useToast()
  const [city, setCity] = useState(cityPref)
  const { data, loading, showLoading, error, setError, load } = useDashboard(read_data, { city })

  useEffect(() => {
    controller.init()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  function handler_city_change(value) {
    setCity(value)
    try {
      localStorage.setItem('tikto-city', value)
    } catch {
      // abaikan (mode privat)
    }
  }

  function handler_copy_code(promo) {
    const done = () => toast.success(`Kode ${promo.promo_code} disalin.`)
    try {
      if (navigator.clipboard && navigator.clipboard.writeText) {
        navigator.clipboard.writeText(promo.promo_code).then(done).catch(() => toast.info(promo.promo_code, { title: 'Kode promo' }))
      } else {
        toast.info(promo.promo_code, { title: 'Kode promo' })
      }
    } catch {
      toast.info(promo.promo_code, { title: 'Kode promo' })
    }
  }

  function goSection(menu) {
    if (onNavigate) onNavigate(menu)
  }

  function goEvents(params = {}) {
    const q = new URLSearchParams()
    for (const [k, v] of Object.entries(params)) if (v) q.append(k, v)
    const s = q.toString()
    replaceHash(`#/events${s ? `?${s}` : ''}`)
    if (onNavigate) onNavigate('events')
  }

  const home = data?.home || null
  const items = home ? [home] : []
  const cities = data?.cities || []
  const banners = home?.banners || []
  const promos = home?.promos || []
  const news = home?.news || []
  const popular = home?.popular || []
  const upcoming = home?.upcoming || []
  const categories = home?.categories || []

  return (
    <div className="flex flex-col gap-4">
      <StandardPage
        title={nvdata?.label || 'Home'}
        description="Jelajahi event, berita, dan promo terbaru."
        loading={loading}
        showLoading={showLoading}
        error={error}
        errorTitle="Gagal memuat home"
        onClearError={() => setError('')}
        onRefresh={() => controller.btrefresh_click(load)}
        filterBar={
          <div className="mb-4 flex flex-wrap items-end gap-2">
            <label className="flex flex-col gap-1 text-xs font-medium text-zinc-500">
              Kota
              <select value={city} onChange={(e) => handler_city_change(e.target.value)} className={SELECT_CLASS}>
                <option value="">Semua kota</option>
                {cities.map((c) => (
                  <option key={c.code} value={c.code}>
                    {c.name}
                  </option>
                ))}
              </select>
            </label>
          </div>
        }
        items={items}
        emptyTitle="Belum ada konten"
        emptyDescription="Belum ada event, berita, atau promo yang terbit."
        emptyAction={
          <Button variant="secondary" size="sm" onClick={() => controller.btrefresh_click(load)}>
            Muat ulang
          </Button>
        }
      >
        {banners.length > 0 && (
          <BannerCarousel key={banners.map((b) => b.code).join('|')} banners={banners} />
        )}
        {categories.length > 0 && (
          <div className="mt-4 flex gap-2 overflow-x-auto pb-1">
            {categories.map((c) => (
              <Button key={c.code} size="sm" variant="secondary" onClick={() => goEvents({ categories: c.code })}>
                {c.name}
              </Button>
            ))}
          </div>
        )}
      </StandardPage>

      <Card>
        <div className="mb-4 flex items-center justify-between gap-2">
          <CardTitle description="Penawaran yang sedang berlaku.">Promo pilihan</CardTitle>
          <Button size="sm" variant="ghost" onClick={() => goSection('promotions')}>
            Semua promo
          </Button>
        </div>
        {showLoading ? (
          <Skeleton className="h-24" />
        ) : promos.length === 0 ? (
          <p className="text-sm text-zinc-500">Belum ada promo yang berlaku.</p>
        ) : (
          <div className="grid gap-3 sm:grid-cols-2">
            {promos.map((p) => (
              <PromoCard key={p.code} promo={p} onCopy={handler_copy_code} />
            ))}
          </div>
        )}
      </Card>

      <Card>
        <div className="mb-4 flex items-center justify-between gap-2">
          <CardTitle description="Kabar terbaru seputar event.">Berita terbaru</CardTitle>
          <Button size="sm" variant="ghost" onClick={() => goSection('news')}>
            Semua berita
          </Button>
        </div>
        {showLoading ? (
          <Skeleton className="h-24" />
        ) : news.length === 0 ? (
          <p className="text-sm text-zinc-500">Belum ada berita terbit.</p>
        ) : (
          <div className="grid gap-3 sm:grid-cols-2">
            {news.map((n) => (
              <NewsCard key={n.code} news={n} onOpen={(item) => goDetail(onNavigate, 'news', item.slug)} />
            ))}
          </div>
        )}
      </Card>

      <Card>
        <div className="mb-4 flex items-center justify-between gap-2">
          <CardTitle description="Paling banyak dibeli tiketnya.">Event populer</CardTitle>
          <Button size="sm" variant="ghost" onClick={() => goEvents({ sort: 'popular' })}>
            Semua event
          </Button>
        </div>
        {showLoading ? (
          <Skeleton className="h-24" />
        ) : popular.length === 0 ? (
          <EmptyState
            icon="folder"
            title={city ? 'Belum ada event di kota ini' : 'Belum ada event'}
            description="Coba kota lain atau muat ulang."
          />
        ) : (
          <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
            {popular.map((e) => (
              <EventCard key={e.code} event={e} onOpen={(item) => goEvent(onNavigate, item.code)} />
            ))}
          </div>
        )}
      </Card>

      <Card>
        <div className="mb-4 flex items-center justify-between gap-2">
          <CardTitle description="Segera berlangsung.">Event terdekat</CardTitle>
          <Button size="sm" variant="ghost" onClick={() => goEvents({})}>
            Semua event
          </Button>
        </div>
        {showLoading ? (
          <Skeleton className="h-24" />
        ) : upcoming.length === 0 ? (
          <p className="text-sm text-zinc-500">Belum ada event mendatang.</p>
        ) : (
          <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
            {upcoming.map((e) => (
              <EventCard key={e.code} event={e} onOpen={(item) => goEvent(onNavigate, item.code)} />
            ))}
          </div>
        )}
      </Card>

      {error && (
        <Alert tone="error" title="Gagal memuat home" closable onClose={() => setError('')}>
          {error}
        </Alert>
      )}
    </div>
  )
}
