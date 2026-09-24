import { useEffect, useRef, useState } from 'react'

// Anti-kedip: skeleton tampil minimal `minMs` agar refresh cepat tidak
// berkedip, dan tidak berkedip mati-hidup saat loading beruntun.
// Pakai: const show = useSmoothLoading(loading); … {show ? <Skeleton/> : …}
export default function useSmoothLoading(loading, minMs = 350) {
  const [show, setShow] = useState(loading)
  const startRef = useRef(0)

  useEffect(() => {
    if (loading) {
      startRef.current = Date.now()
      setShow(true)
      return
    }
    const elapsed = Date.now() - startRef.current
    if (elapsed >= minMs) {
      setShow(false)
      return
    }
    const t = setTimeout(() => setShow(false), minMs - elapsed)
    return () => clearTimeout(t)
  }, [loading, minMs])

  return show
}
