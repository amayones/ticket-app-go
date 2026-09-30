import { useState } from 'react'
import { placeholderImage } from './placeholder.js'

// Gambar dengan lazy loading + cadangan placeholder bila URL kosong
// atau gagal dimuat (error jaringan, expired, dsb.).
export default function SafeImage({ src, alt = '', className = '' }) {
  const [failed, setFailed] = useState(false)
  const shown = !src || failed ? placeholderImage(alt) : src
  return (
    <img
      src={shown}
      alt={alt}
      loading="lazy"
      onError={() => setFailed(true)}
      className={className}
    />
  )
}
