// Tema terang sebagai default; pilihan disimpan di localStorage.
// Dark mode Tailwind v4 memakai varian class (@custom-variant dark di index.css).
const KEY = 'go-core-theme'

export function currentTheme() {
  try {
    return localStorage.getItem(KEY) === 'dark' ? 'dark' : 'light'
  } catch {
    return 'light'
  }
}

export function applyTheme(theme) {
  const dark = theme === 'dark'
  document.documentElement.classList.toggle('dark', dark)
  try {
    localStorage.setItem(KEY, dark ? 'dark' : 'light')
  } catch {
    // abaikan (mode privat)
  }
  return dark ? 'dark' : 'light'
}

export function initTheme() {
  applyTheme(currentTheme())
}
