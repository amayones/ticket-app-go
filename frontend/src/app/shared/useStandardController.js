import { useCallback, useEffect, useState } from 'react'
import { useSmoothLoading } from '../../components'
import { PAGE_SIZE } from './config.js'

// Controller standar ala C<mod>.js langit_v2: load/refresh/pagination/error.
// fetcher = async ({ limit, offset, ...params }) => rows.
// Dipakai semua halaman list agar pola hooks sama persis.
export function useStandardController(fetcher, params = {}) {
  const [rows, setRows] = useState([])
  const [offset, setOffset] = useState(0)
  const [loading, setLoading] = useState(true)
  const showLoading = useSmoothLoading(loading)
  const [error, setError] = useState('')
  const paramsKey = JSON.stringify(params)

  const load = useCallback(async () => {
    setLoading(true)
    setError('')
    try {
      const parsed = JSON.parse(paramsKey)
      const data = await fetcher({ ...parsed, limit: PAGE_SIZE, offset })
      setRows(Array.isArray(data) ? data : [])
    } catch (err) {
      setError(err.message)
    } finally {
      setLoading(false)
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [offset, paramsKey])

  useEffect(() => {
    load()
  }, [load])

  function resetPage() {
    setOffset(0)
  }

  return { rows, setRows, offset, setOffset, loading, showLoading, error, setError, load, resetPage }
}
