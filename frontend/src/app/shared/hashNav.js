// Navigasi hash untuk halaman discovery (tautan berbagi).
// Format: #/<menu>[/param][?query]. Menu lain tidak tersentuh: mereka hanya
// menulis #/<menu> saat dipilih (App.jsx selectView).
export function hashParam(menu) {
  try {
    const h = window.location.hash || ''
    const m = h.match(new RegExp(`^#/${menu}(?:/([^?]*))?`))
    return m && m[1] ? decodeURIComponent(m[1]) : ''
  } catch {
    return ''
  }
}

export function hashQuery() {
  try {
    const h = window.location.hash || ''
    const i = h.indexOf('?')
    return new URLSearchParams(i >= 0 ? h.slice(i + 1) : '')
  } catch {
    return new URLSearchParams()
  }
}

// goDetail: pindah ke halaman detail + tulis hash berbagi, lalu navigate.
export function goDetail(navigate, menu, param) {
  try {
    window.location.hash = param ? `#/${menu}/${encodeURIComponent(param)}` : `#/${menu}`
  } catch {
    // abaikan (non-browser)
  }
  navigate(menu)
}

// goEvent: buka modal detail event dari mana saja (daftar Events, Home,
// News). Hash #/events/<code> supaya bisa dibagikan dan dibuka langsung.
export function goEvent(navigate, code) {
  try {
    window.location.hash = `#/events/${encodeURIComponent(code)}`
  } catch {
    // abaikan (non-browser)
  }
  navigate('events')
}

export function replaceHash(hash) {
  try {
    window.history.replaceState(null, '', hash)
  } catch {
    // abaikan (non-browser)
  }
}

// clearHash: kembalikan URL bersih (tanpa hash) saat kembali ke daftar.
export function clearHash() {
  try {
    window.history.replaceState(null, '', window.location.pathname || '/')
  } catch {
    // abaikan (non-browser)
  }
}
