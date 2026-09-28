import { useCallback, useEffect, useState } from 'react'
import { useSmoothLoading } from '../../components'
import { PAGE_SIZE } from './config.js'

// Hook seragam untuk halaman list: load(offset) -> { data, offset, ... }.
// fetcher = async (limit, offset) => rows. Untuk fetcher objek
// ({limit, offset, ...}) pakai usePageListObj di bawah.
export function usePageList(fetcher, deps = []) {
  const [data, setData] = useState([])
  const [offset, setOffset] = useState(0)
  const [loading, setLoading] = useState(true)
  const showLoading = useSmoothLoading(loading)
  const [error, setError] = useState('')

  const load = useCallback(
    async (nextOffset = offset) => {
      setLoading(true)
      setError('')
      try {
        setData(await fetcher(PAGE_SIZE, nextOffset))
      } catch (err) {
        setError(err.message)
      } finally {
        setLoading(false)
      }
    },
    // eslint-disable-next-line react-hooks/exhaustive-deps
    deps
  )

  useEffect(() => {
    load(offset)
  }, [offset, load])

  return { data, setData, offset, setOffset, loading, showLoading, error, setError, load }
}

// Varian objek: fetcher = async ({limit, offset, ...extra}) => rows.
export function usePageListObj(fetcher, extra = {}) {
  const [data, setData] = useState([])
  const [offset, setOffset] = useState(0)
  const [loading, setLoading] = useState(true)
  const showLoading = useSmoothLoading(loading)
  const [error, setError] = useState('')

  const load = useCallback(async () => {
    setLoading(true)
    setError('')
    try {
      setData(await fetcher({ ...extra, limit: PAGE_SIZE, offset }))
    } catch (err) {
      setError(err.message)
    } finally {
      setLoading(false)
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [offset, JSON.stringify(extra)])

  useEffect(() => {
    load()
  }, [load])

  return { data, setData, offset, setOffset, loading, showLoading, error, setError, load }
}
