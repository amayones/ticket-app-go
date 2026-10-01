import { AppLogo } from '../components'
import { APP_NAME } from '../brand.js'

const categories = ['All', 'Concert', 'Festival', 'Sports', 'Workshop', 'Theater', 'Seminar', 'Exhibition']
const popular = [
  { title: 'Jakarta Soundscape Festival 2024', date: '17 August 2024', price: 'Rp 700,000', cat: 'Festival', img: 'https://images.unsplash.com/photo-1493225457124-a3eb161ffa5f?w=600&q=80', tone: 'violet' },
  { title: 'Indonesian Football League: Persija vs Persib', date: '20 September 2024', price: 'Rp 100,000', cat: 'Sports', img: 'https://images.unsplash.com/photo-1543351611-58f69d7c1781?w=600&q=80', tone: 'orange' },
  { title: 'Digital Marketing Mastery Workshop', date: '05 October 2024', price: 'Rp 1,200,000', cat: 'Workshop', img: 'https://images.unsplash.com/photo-1552664730-d307ca884978?w=600&q=80', tone: 'emerald' },
]
const jakarta = [
  { title: 'Jazz Goes To Campus', price: 'Rp 95,000', img: 'https://images.unsplash.com/photo-1511379938547-c1f69419868d?w=600&q=80' },
  { title: 'Art Jakarta 2024', price: 'Rp 150,000', img: 'https://images.unsplash.com/photo-1577720643272-265f09367456?w=600&q=80' },
  { title: 'E-Sport National', price: 'Rp 80,000', img: 'https://images.unsplash.com/photo-1511512578047-dfb367046420?w=600&q=80' },
  { title: 'Pentas Budaya Nusantara', price: 'Free', img: 'https://images.unsplash.com/photo-1533174072545-7a4b6ad7a6c3?w=600&q=80' },
]
const merch = [
  { name: 'T-Shirt Eksklusif TicketIN', price: 'Rp 149,000', img: 'https://images.unsplash.com/photo-1521572163474-6864f9cf17ab?w=400&q=80' },
  { name: 'Topi Snapback Festival', price: 'Rp 89,000', img: 'https://images.unsplash.com/photo-1588850561407-ed78c282e89b?w=400&q=80' },
  { name: 'Hoodie Jakarta Soundscape', price: 'Rp 299,000', img: 'https://images.unsplash.com/photo-1556821840-3a63f95609a7?w=400&q=80' },
  { name: 'Tumbler Stainless Steel TicketIN', price: 'Rp 125,000', img: 'https://images.unsplash.com/photo-1523369364227-24934b335841?w=400&q=80' },
]

export default function HomePage({ onLogin, onDashboard, onLogout, loggedIn, hasAccess, profileLoaded }) {
  return (
    <div className="min-h-screen w-full bg-white text-zinc-900 dark:bg-zinc-950 dark:text-zinc-100">
      <header className="sticky top-0 z-30 border-b border-zinc-100 bg-white/95 backdrop-blur dark:border-zinc-800 dark:bg-zinc-950/90">
        <div className="mx-auto flex max-w-[1240px] items-center gap-3 px-4 py-3 sm:px-6">
          <div className="flex items-center gap-2">
            <AppLogo className="h-7 w-7 rounded-lg" />
            <span className="text-sm font-extrabold tracking-tight">{APP_NAME}</span>
            <span className="hidden items-center gap-3 pl-6 text-xs font-medium text-zinc-500 dark:text-zinc-400 lg:flex">
              <a href="#concert" className="hover:text-zinc-900 dark:hover:text-white">Concert</a>
              <a href="#festival" className="hover:text-zinc-900 dark:hover:text-white">Festival</a>
              <a href="#sports" className="hover:text-zinc-900 dark:hover:text-white">Sports</a>
              <a href="#workshop" className="hover:text-zinc-900 dark:hover:text-white">Workshop</a>
            </span>
          </div>
          <div className="ml-auto flex items-center gap-2">
            <div className="hidden items-center gap-2 sm:flex">
              <div className="relative">
                <input placeholder="Search event..." className="w-[200px] rounded-full border border-zinc-200 bg-zinc-50 py-1.5 pl-8 pr-3 text-xs outline-none placeholder:text-zinc-400 focus:border-violet-300 focus:bg-white dark:border-zinc-700 dark:bg-zinc-900 lg:w-[260px]" />
                <svg className="pointer-events-none absolute left-2.5 top-1/2 h-3.5 w-3.5 -translate-y-1/2 text-zinc-400" fill="none" stroke="currentColor" strokeWidth="2" viewBox="0 0 24 24"><circle cx="11" cy="11" r="7" /><path d="M20 20l-3-3" /></svg>
              </div>
            </div>
            {!loggedIn ? (
              <>
                <button type="button" onClick={onLogin} className="rounded-full px-4 py-1.5 text-xs font-semibold text-zinc-700 hover:bg-zinc-100 dark:text-zinc-300 dark:hover:bg-zinc-800">Login</button>
                <button type="button" onClick={onLogin} className="rounded-full bg-[#6b5cff] px-5 py-1.5 text-xs font-semibold text-white shadow-sm hover:bg-[#5a4af0]">Sign Up</button>
              </>
            ) : profileLoaded && !hasAccess ? (
              <button type="button" onClick={onLogout} className="rounded-full border border-zinc-300 bg-white px-5 py-1.5 text-xs font-semibold text-zinc-700 hover:bg-zinc-50 dark:border-zinc-600 dark:bg-zinc-800 dark:text-zinc-200">Logout</button>
            ) : (
              <button type="button" onClick={onDashboard} className="rounded-full bg-[#6b5cff] px-5 py-1.5 text-xs font-semibold text-white shadow-sm hover:bg-[#5a4af0]">Dashboard</button>
            )}
          </div>
        </div>
      </header>

      <section className="mx-auto max-w-[1240px] px-4 pt-4 sm:px-6">
        <div className="relative overflow-hidden rounded-[24px] bg-[#0b0518]">
          <img src="https://images.unsplash.com/photo-1470225620780-dba8ba36b745?w=1200&q=80" alt="" className="absolute inset-0 h-full w-full object-cover opacity-60" />
          <div className="absolute inset-0 bg-gradient-to-r from-[#1a0533] via-[#4a1a8a]/60 to-transparent" />
          <div className="absolute inset-0 bg-gradient-to-t from-[#0b0518]/40 to-transparent" />
          <div className="relative grid gap-6 p-6 sm:p-8 lg:grid-cols-[1.1fr_0.9fr] lg:p-12">
            <div className="flex flex-col justify-center">
              <p className="mb-3 inline-flex w-fit items-center gap-1.5 rounded-full bg-white/10 px-2.5 py-1 text-[10px] font-semibold tracking-widest text-violet-200 ring-1 ring-white/10 backdrop-blur">✦ FEATURED EVENT</p>
              <h1 style={{ color: '#fff' }} className="max-w-[520px] text-[28px] font-extrabold leading-[1.05] tracking-tight text-white sm:text-[36px]">Feel the Biggest Music Vibes of the Year!</h1>
              <p className="mt-3 max-w-[480px] text-xs leading-relaxed text-violet-100/80 sm:text-sm">Get your Jakarta Soundscape Festival ticket now before it runs out. Special performances from international and local artists.</p>
              <div className="mt-6 flex flex-wrap gap-2.5">
                <button onClick={onLogin} type="button" className="rounded-full bg-[#6b5cff] px-5 py-2.5 text-xs font-semibold text-white shadow-lg shadow-violet-900/30 hover:bg-[#5a4af0]">Buy Ticket Now</button>
                <button type="button" className="rounded-full bg-white/10 px-5 py-2.5 text-xs font-semibold text-white ring-1 ring-white/20 backdrop-blur hover:bg-white/15">View Details</button>
              </div>
            </div>
          </div>
          <div className="absolute bottom-4 right-4 hidden items-center gap-2 lg:flex">
            <span className="h-1.5 w-6 rounded-full bg-white" /><span className="h-1.5 w-1.5 rounded-full bg-white/40" /><span className="h-1.5 w-1.5 rounded-full bg-white/40" />
          </div>
        </div>
      </section>

      <section className="mx-auto max-w-[1240px] px-4 pt-6 sm:px-6">
        <p className="text-xs font-bold tracking-tight text-zinc-900 dark:text-white">Browse by Category</p>
        <div className="mt-3 flex flex-wrap gap-2">
          {categories.map((c, i) => (
            <button key={c} type="button" className={`rounded-full px-3.5 py-1.5 text-xs font-semibold ring-1 ${i === 0 ? 'bg-[#6b5cff] text-white ring-[#6b5cff]' : 'bg-white text-zinc-600 ring-zinc-200 hover:bg-zinc-50 dark:bg-zinc-900 dark:text-zinc-300 dark:ring-zinc-700'}`}>{c}</button>
          ))}
        </div>
      </section>

      <section className="mx-auto max-w-[1240px] px-4 pt-8 sm:px-6" id="festival">
        <div className="flex items-end justify-between">
          <div>
            <p className="text-[10px] font-bold tracking-widest text-violet-600">✦ TRENDING NOW</p>
            <h2 className="text-base font-extrabold tracking-tight text-zinc-900 dark:text-white">Most Popular Events</h2>
          </div>
          <a href="#" className="text-xs font-medium text-zinc-400 hover:text-zinc-600">View All →</a>
        </div>
        <div className="mt-4 grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {popular.map((e) => (
            <div key={e.title} className="overflow-hidden rounded-2xl border border-zinc-100 bg-white shadow-sm dark:border-zinc-800 dark:bg-zinc-900">
              <div className="relative h-[170px] overflow-hidden">
                <img src={e.img} alt="" className="h-full w-full object-cover" />
                <span className={`absolute left-2 top-2 rounded-full px-2 py-0.5 text-[10px] font-bold text-white ${e.tone === 'violet' ? 'bg-violet-600' : e.tone === 'orange' ? 'bg-orange-500' : 'bg-emerald-600'}`}>{e.cat}</span>
              </div>
              <div className="p-3.5">
                <p className="text-[10px] text-zinc-400">{e.date}</p>
                <p className="mt-1 line-clamp-2 text-xs font-bold leading-tight text-zinc-900 dark:text-white">{e.title}</p>
                <div className="mt-3 flex items-center justify-between">
                  <p className="text-xs font-bold text-[#6b5cff]">{e.price}</p>
                  <button onClick={onLogin} type="button" className="rounded-full bg-[#6b5cff] px-3 py-1 text-[10px] font-semibold text-white">Buy Ticket</button>
                </div>
              </div>
            </div>
          ))}
        </div>
      </section>

      <section className="mx-auto max-w-[1240px] px-4 pt-8 sm:px-6">
        <div className="flex items-end justify-between">
          <div>
            <p className="text-[10px] font-bold tracking-widest text-violet-600">✦ FUN IN JAKARTA</p>
            <h2 className="text-base font-extrabold tracking-tight text-zinc-900 dark:text-white">Exciting Events in Jakarta</h2>
          </div>
          <a href="#" className="text-xs font-medium text-zinc-400 hover:text-zinc-600">View All →</a>
        </div>
        <div className="mt-4 grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
          {jakarta.map((e) => (
            <div key={e.title} className="overflow-hidden rounded-2xl border border-zinc-100 bg-white shadow-sm dark:border-zinc-800 dark:bg-zinc-900">
              <img src={e.img} alt="" className="h-[140px] w-full object-cover" />
              <div className="p-3">
                <p className="text-xs font-bold leading-tight text-zinc-900 dark:text-white">{e.title}</p>
                <div className="mt-2 flex items-center justify-between">
                  <p className="text-[11px] font-semibold text-zinc-500">{e.price}</p>
                  <button onClick={onLogin} type="button" className="rounded-full bg-[#6b5cff] px-3 py-1 text-[10px] font-semibold text-white">Buy Ticket</button>
                </div>
              </div>
            </div>
          ))}
        </div>
      </section>

      <section className="mx-auto max-w-[1240px] px-4 pt-8 sm:px-6">
        <div className="grid overflow-hidden rounded-[20px] bg-[#1d0b4a] md:grid-cols-2">
          <div className="p-6 sm:p-8">
            <p className="inline-flex rounded-full bg-white/10 px-2.5 py-1 text-[10px] font-bold tracking-widest text-violet-200">LIMITED OFFER</p>
            <h3 className="mt-3 text-xl font-extrabold leading-tight text-white">Get Up to 50% Off<br />for Selected Events!</h3>
            <p className="mt-2 text-xs leading-relaxed text-violet-200/70">Use promo code <b className="text-white">MAINEVENT</b> at checkout. Enjoy savings with various payment methods.</p>
            <button type="button" className="mt-5 rounded-full bg-[#6b5cff] px-5 py-2 text-xs font-semibold text-white hover:bg-[#5a4af0]">Claim Promo Now</button>
          </div>
          <div className="relative hidden bg-white md:block">
            <img src="https://images.unsplash.com/photo-1549465220-1a8b9238cd48?w=700&q=80" alt="" className="h-full w-full object-cover" />
            <span className="absolute left-1/2 top-1/2 -translate-x-1/2 -translate-y-1/2 rounded-xl bg-black px-4 py-2 text-lg font-black text-white shadow-xl">- 50 %</span>
          </div>
        </div>
      </section>

      <section className="mx-auto max-w-[1240px] px-4 pt-8 sm:px-6">
        <div className="flex items-end justify-between">
          <div>
            <p className="text-[10px] font-bold tracking-widest text-violet-600">✦ OFFICIAL STORE</p>
            <h2 className="text-base font-extrabold tracking-tight text-zinc-900 dark:text-white">Exclusive Merchandise</h2>
          </div>
          <a href="#" className="text-xs font-medium text-zinc-400 hover:text-zinc-600">See More →</a>
        </div>
        <div className="mt-4 grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
          {merch.map((m) => (
            <div key={m.name} className="overflow-hidden rounded-2xl border border-zinc-100 bg-white dark:border-zinc-800 dark:bg-zinc-900">
              <img src={m.img} alt="" className="h-[160px] w-full object-cover" />
              <div className="p-3 text-center">
                <p className="text-xs font-semibold text-zinc-900 dark:text-white">{m.name}</p>
                <p className="mt-1 text-xs font-bold text-[#6b5cff]">{m.price}</p>
              </div>
            </div>
          ))}
        </div>
      </section>

      <section className="mx-auto max-w-[1240px] px-4 pt-10 sm:px-6">
        <div className="rounded-[20px] bg-[#f8f7ff] px-6 py-8 dark:bg-zinc-900">
          <h2 className="text-center text-sm font-extrabold text-zinc-900 dark:text-white">How {APP_NAME} Works?</h2>
          <p className="mx-auto mt-1 max-w-[480px] text-center text-xs text-zinc-500">In just 3 easy steps, you can get your dream event ticket hassle-free.</p>
          <div className="mt-8 grid gap-6 sm:grid-cols-3">
            {[
              { n: '1. Find Event', d: 'Find your favorite concert, festival, or workshop via our smart search.', icon: '◉' },
              { n: '2. Book & Pay', d: 'Choose ticket, fill details, and pay with various payment methods.', icon: '⚡' },
              { n: '3. Receive E-Ticket', d: 'E-ticket will be sent to your email and available in your TicketIN dashboard.', icon: '🎟' },
            ].map((s) => (
              <div key={s.n} className="text-center">
                <div className="mx-auto flex h-10 w-10 items-center justify-center rounded-xl bg-violet-100 text-violet-600 dark:bg-violet-900/40">{s.icon}</div>
                <p className="mt-3 text-xs font-bold text-zinc-900 dark:text-white">{s.n}</p>
                <p className="mx-auto mt-1 max-w-[220px] text-xs leading-relaxed text-zinc-500">{s.d}</p>
              </div>
            ))}
          </div>
        </div>
        <div className="mt-6 overflow-hidden rounded-2xl border border-zinc-100 bg-white dark:border-zinc-800">
          <img src="https://images.unsplash.com/photo-1514525253161-7a46d19cd819?w=1200&q=80" alt="" className="h-[180px] w-full object-cover" />
        </div>
      </section>

      <section className="mx-auto max-w-[1240px] px-4 pt-8 sm:px-6">
        <div className="flex items-end justify-between">
          <div>
            <h2 className="text-sm font-extrabold text-zinc-900 dark:text-white">Blog & Latest News</h2>
            <p className="text-xs text-zinc-500">Update information, tips and tricks for you.</p>
          </div>
          <a href="#" className="hidden text-xs font-medium text-zinc-400 hover:text-zinc-600 sm:block">View All Articles →</a>
        </div>
        <div className="mt-4 grid gap-4 md:grid-cols-2">
          {[
            { title: '2024 Event Industry Trends: Digital Transformation', excerpt: 'How blockchain and NFT technology are changing the ticket industry...' },
            { title: 'Event Planning Tips for Beginners', excerpt: 'Want to create a successful event? Here are the must-know tips for beginners...' },
          ].map((b) => (
            <div key={b.title} className="flex gap-4 rounded-2xl border border-zinc-100 bg-white p-4 dark:border-zinc-800 dark:bg-zinc-900">
              <img src="https://images.unsplash.com/photo-1454165205744-3b78555e5572?w=300&q=80" alt="" className="h-24 w-32 shrink-0 rounded-xl object-cover" />
              <div>
                <p className="text-xs font-bold leading-tight text-zinc-900 dark:text-white">{b.title}</p>
                <p className="mt-1 line-clamp-2 text-xs leading-relaxed text-zinc-500">{b.excerpt}</p>
                <a href="#" className="mt-2 inline-block text-xs font-semibold text-[#6b5cff]">Read More →</a>
              </div>
            </div>
          ))}
        </div>
      </section>

      <section className="mx-auto max-w-[1240px] px-4 pt-8 sm:px-6">
        <div className="grid gap-6 rounded-[20px] bg-gradient-to-br from-[#1d0b4a] to-[#4b1fa7] p-6 text-white sm:p-8 lg:grid-cols-[1.2fr_0.8fr]">
          <div>
            <h3 className="text-lg font-extrabold leading-tight">Make Your Event Successful with {APP_NAME}</h3>
            <p className="mt-2 text-xs leading-relaxed text-violet-200">We provide a secure and reliable event management platform for you.</p>
            <ul className="mt-4 grid grid-cols-2 gap-2 text-xs text-violet-100">
              <li className="flex items-center gap-1.5"><span className="h-1.5 w-1.5 rounded-full bg-emerald-400" />Real-time Analytics Dashboard</li>
              <li className="flex items-center gap-1.5"><span className="h-1.5 w-1.5 rounded-full bg-emerald-400" />Check-in in Seconds</li>
              <li className="flex items-center gap-1.5"><span className="h-1.5 w-1.5 rounded-full bg-emerald-400" />Payout & Transparent Reporting</li>
              <li className="flex items-center gap-1.5"><span className="h-1.5 w-1.5 rounded-full bg-emerald-400" />Attendee Management</li>
            </ul>
            <div className="mt-5 flex gap-2">
              <button onClick={onLogin} type="button" className="rounded-full bg-white px-5 py-2 text-xs font-bold text-[#1d0b4a]">Start and Organize</button>
              <button type="button" className="rounded-full bg-white/10 px-5 py-2 text-xs font-semibold text-white ring-1 ring-white/20">Free Consultation</button>
            </div>
          </div>
          <div className="rounded-2xl bg-white/10 p-4 backdrop-blur ring-1 ring-white/10">
            <p className="flex items-center gap-2 text-xs font-bold"><span className="flex h-6 w-6 items-center justify-center rounded-full bg-pink-500 text-[11px]">♥</span>10,000+</p>
            <p className="mt-1 text-[11px] text-violet-200">Trusted Organizers</p>
            <div className="mt-4 rounded-xl bg-white p-3 text-zinc-900">
              <p className="text-[11px] text-zinc-500">Tickets Sold This Year</p>
              <p className="text-sm font-extrabold">250,430 Tickets</p>
            </div>
            <div className="mt-3 flex items-center justify-between rounded-xl bg-white/10 px-3 py-2">
              <span className="text-xs font-semibold">Satisfaction</span><span className="text-xs font-bold">4.9 / 5.0 ★★★★</span>
            </div>
          </div>
        </div>
      </section>

      <div className="mx-auto flex max-w-[1240px] flex-wrap justify-center gap-6 border-y border-zinc-100 px-4 py-4 text-[10px] font-bold tracking-widest text-zinc-400 dark:border-zinc-800 sm:px-6"> <span>EVENT INDONESIA</span><span>KEMENTERIAN PARIWISATA</span><span>MUSIC GROUP ID</span><span>SPORTS FEDERATION</span><span>WORKSHOP HUB</span></div>

      <footer className="mx-auto max-w-[1240px] px-4 py-8 text-xs text-zinc-500 dark:text-zinc-400 sm:px-6">
        <div className="grid gap-8 sm:grid-cols-4">
          <div>
            <p className="flex items-center gap-1.5 text-sm font-extrabold text-zinc-900 dark:text-white"><AppLogo className="h-6 w-6" />{APP_NAME}</p>
            <p className="mt-2 leading-relaxed">The largest ticket platform in Indonesia. Find your favorite concert, festival, sports and workshop tickets here.</p>
          </div>
          <div>
            <p className="font-bold text-zinc-900 dark:text-white">Information</p>
            <ul className="mt-2 space-y-1"><li><a href="#" className="hover:text-zinc-800">About Us</a></li><li><a href="#" className="hover:text-zinc-800">Careers</a></li><li><a href="#" className="hover:text-zinc-800">Blog</a></li><li><a href="#" className="hover:text-zinc-800">Help Center</a></li></ul>
          </div>
          <div>
            <p className="font-bold text-zinc-900 dark:text-white">Policy</p>
            <ul className="mt-2 space-y-1"><li>Terms & Conditions</li><li>Privacy Policy</li><li>Data Security</li></ul>
          </div>
          <div>
            <p className="font-bold text-zinc-900 dark:text-white">Contact Us</p>
            <ul className="mt-2 space-y-1"><li>support@tickein.com</li><li>Jl. Sudirman No. 1, Jakarta Selatan</li><li>www.tickein.com</li></ul>
          </div>
        </div>
        <div className="mt-8 flex flex-col items-center justify-between gap-2 border-t border-zinc-100 pt-4 dark:border-zinc-800 sm:flex-row">
          <p className="text-[11px]">© 2026 {APP_NAME}. All rights reserved. Unauthorized copying prohibited.</p>
          <p className="text-[11px]">Bahasa Indonesia</p>
        </div>
      </footer>
    </div>
  )
}
