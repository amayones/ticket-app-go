import { Badge, Button, Icon, formatAmount, formatTime } from '../../components'
import SafeImage from './SafeImage.jsx'

// Kartu bersama 5 menu discovery (Home, News, Events, Event Detail,
// Promotions): satu sumber tampilan agar terlihat ditulis orang yang sama.
function rupiah(value) {
  const s = formatAmount(value)
  return s === '' ? 'Rp 0' : `Rp ${s}`
}

function eventBadge(badge) {
  if (badge === 'SOLD_OUT') return <Badge tone="danger">Sold Out</Badge>
  if (badge === 'HAMPIR_HABIS') return <Badge tone="warning">Hampir Habis</Badge>
  return null
}

export function EventCard({ event, onOpen }) {
  return (
    <button
      type="button"
      onClick={() => onOpen && onOpen(event)}
      className="flex min-w-0 flex-col overflow-hidden rounded-2xl border border-zinc-200 bg-white text-left shadow-sm transition hover:border-violet-300 hover:shadow-md dark:border-zinc-800 dark:bg-zinc-900"
    >
      <div className="relative">
        <SafeImage src={event.poster_url} alt={event.title} className="aspect-video w-full object-cover" />
        <span className="absolute left-2 top-2 flex gap-1.5">
          {event.category ? <Badge tone="brand">{event.category}</Badge> : null}
          {eventBadge(event.badge)}
        </span>
      </div>
      <div className="flex min-w-0 flex-1 flex-col gap-1 p-3.5">
        <p className="truncate text-sm font-semibold text-zinc-900 dark:text-zinc-50">{event.title}</p>
        <p className="truncate text-xs text-zinc-500 dark:text-zinc-400">
          {[event.city, event.venue].filter(Boolean).join(' · ')}
        </p>
        <p className="truncate text-xs text-zinc-500 dark:text-zinc-400">{formatTime(event.start_at)}</p>
        <p className="mt-auto pt-1 text-sm font-bold text-violet-700 dark:text-violet-300">
          Mulai {rupiah(event.min_price)}
        </p>
      </div>
    </button>
  )
}

export function NewsCard({ news, onOpen }) {
  return (
    <button
      type="button"
      onClick={() => onOpen && onOpen(news)}
      className="flex min-w-0 flex-col overflow-hidden rounded-2xl border border-zinc-200 bg-white text-left shadow-sm transition hover:border-violet-300 hover:shadow-md dark:border-zinc-800 dark:bg-zinc-900"
    >
      <SafeImage src={news.image_url} alt={news.title} className="aspect-video w-full object-cover" />
      <div className="flex min-w-0 flex-1 flex-col gap-1.5 p-3.5">
        <span>
          <Badge tone="info">{news.category}</Badge>
        </span>
        <p className="line-clamp-2 text-sm font-semibold text-zinc-900 dark:text-zinc-50">{news.title}</p>
        {news.summary ? (
          <p className="line-clamp-2 text-xs text-zinc-500 dark:text-zinc-400">{news.summary}</p>
        ) : null}
        <p className="mt-auto truncate pt-1 text-[11px] text-zinc-400 dark:text-zinc-500">
          {formatTime(news.published_at)} · {Number(news.view_count || 0).toLocaleString('id-ID')} dibaca
        </p>
      </div>
    </button>
  )
}

export function PromoCard({ promo, onCopy }) {
  const upcoming = promo.state === 'UPCOMING'
  return (
    <div className="flex min-w-0 flex-col gap-2 rounded-2xl border border-dashed border-violet-300 bg-violet-50/60 p-4 dark:border-violet-800 dark:bg-violet-950/30">
      <div className="flex items-center justify-between gap-2">
        <p className="truncate font-mono text-lg font-bold text-violet-800 dark:text-violet-200">{promo.promo_code}</p>
        {upcoming ? <Badge tone="info">Segera hadir</Badge> : <Badge tone="success">Berlaku</Badge>}
      </div>
      {promo.description ? <p className="text-xs text-zinc-600 dark:text-zinc-300">{promo.description}</p> : null}
      <p className="text-[11px] text-zinc-500 dark:text-zinc-400">
        {[promo.min_purchase > 0 ? `Min. ${rupiah(promo.min_purchase)}` : '',
          promo.max_discount > 0 ? `Maks. ${rupiah(promo.max_discount)}` : '',
          promo.event_title ? promo.event_title : '',
          promo.category_name ? promo.category_name : '',
        ]
          .filter(Boolean)
          .join(' · ')}
      </p>
      <div className="mt-auto flex items-center justify-between gap-2 pt-1">
        <p className="truncate text-[11px] text-zinc-500 dark:text-zinc-400">
          {promo.remaining >= 0 ? `Sisa ${promo.remaining} · ` : ''}s/d {formatTime(promo.valid_to)}
        </p>
        {!upcoming && (
          <Button size="sm" variant="secondary" onClick={() => onCopy && onCopy(promo)}>
            <Icon name="list" className="h-4 w-4" />
            Salin
          </Button>
        )}
      </div>
    </div>
  )
}
