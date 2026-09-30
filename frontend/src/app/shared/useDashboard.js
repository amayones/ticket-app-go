import { useCallback, useEffect, useRef, useState } from 'react'
import { useSmoothLoading } from '../../components'

// Controller data dashboard ala useStandardController, ditambah:
// - refetch otomatis tiap intervalMs (default 60 detik), jeda saat tab hidden
// - request lama dibatalkan saat filter berubah (AbortController -> signal)
// - berhenti total saat halaman ditutup (cleanup unmount)
// fetcher = async (params, signal) => data.
export function useDashboard(fetcher, params = {}, { intervalMs = 60000 } = {}) {
  const [data, setData] = useState(null)
  const [loading, setLoading] = useState(true)
  const showLoading = useSmoothLoading(loading)
  const [error, setError] = useState('')
  const paramsKey = JSON.stringify(params)
  const abortRef = useRef(null)

  const load = useCallback(async () => {
    abortRef.current?.abort()
    const ctrl = new AbortController()
    abortRef.current = ctrl
    setLoading(true)
    setError('')
    try {
      const parsed = JSON.parse(paramsKey)
      const result = await fetcher(parsed, ctrl.signal)
      if (!ctrl.signal.aborted) setData(result)
    } catch (err) {
      if (err && err.name === 'AbortError') return
      if (!ctrl.signal.aborted) setError(err.message)
    } finally {
      if (!ctrl.signal.aborted) setLoading(false)
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [paramsKey])

  useEffect(() => {
    load()
  }, [load])

  useEffect(() => {
    const id = setInterval(() => {
      try {
        if (document.visibilityState !== 'hidden') load()
      } catch {
        load()
      }
    }, intervalMs)
    return () => {
      clearInterval(id)
      abortRef.current?.abort()
    }
  }, [load, intervalMs])

  return { data, loading, showLoading, error, setError, load }
}
