// Placeholder gambar tanpa hak cipta: SVG data-URI generik dari judul.
// Dipakai SafeImage saat URL kosong atau gambar gagal dimuat.
export function placeholderImage(text = 'Tikto') {
  const initials = String(text || 'Tikto')
    .split(/\s+/)
    .slice(0, 2)
    .map((w) => (w ? w[0] : ''))
    .join('')
    .toUpperCase()
  const svg =
    `<svg xmlns="http://www.w3.org/2000/svg" width="800" height="450">` +
    `<defs><linearGradient id="g" x1="0" y1="0" x2="1" y2="1">` +
    `<stop offset="0" stop-color="#7c3aed"/><stop offset="1" stop-color="#4c1d95"/>` +
    `</linearGradient></defs>` +
    `<rect width="800" height="450" fill="url(#g)"/>` +
    `<text x="400" y="245" font-family="sans-serif" font-size="96" font-weight="bold" ` +
    `fill="#ffffff" fill-opacity="0.9" text-anchor="middle">${initials}</text></svg>`
  return `data:image/svg+xml,${encodeURIComponent(svg)}`
}
