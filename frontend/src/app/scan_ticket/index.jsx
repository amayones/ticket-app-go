// View Scan Ticket — validasi QR/kode di pintu masuk (nyaman di ponsel).
// Kamera memakai BarcodeDetector bawaan browser (tanpa library); bila tidak
// tersedia, pakai input kode manual. Controller: ./controller.js (6 fungsi).
import { useEffect, useRef, useState } from 'react'
import {
  Alert,
  Badge,
  Button,
  Card,
  CardTitle,
  Skeleton,
  useToast,
  SELECT_CLASS,
  ProgressBar,
} from '../shared/all.js'
import { process_scan, read_data } from './api.js'
import { controller } from './controller.js'

export const meta = { label: 'Scan Ticket', icon: 'eye', order: 4 }

const RESULT_STYLE = {
  VALID: { tone: 'success', label: 'Valid', bar: 'bg-emerald-500 text-white dark:bg-emerald-600' },
  USED: { tone: 'warning', label: 'Sudah Dipakai', bar: 'bg-amber-500 text-white dark:bg-amber-600' },
  INVALID: { tone: 'danger', label: 'Tidak Valid', bar: 'bg-rose-600 text-white dark:bg-rose-700' },
  WRONG_EVENT: { tone: 'danger', label: 'Salah Event', bar: 'bg-rose-600 text-white dark:bg-rose-700' },
  UNPAID: { tone: 'danger', label: 'Belum Dibayar', bar: 'bg-rose-600 text-white dark:bg-rose-700' },
  REFUNDED: { tone: 'danger', label: 'Refund', bar: 'bg-rose-600 text-white dark:bg-rose-700' },
  EXPIRED: { tone: 'danger', label: 'Kedaluwarsa', bar: 'bg-rose-600 text-white dark:bg-rose-700' },
  OUTSIDE_SCHEDULE: { tone: 'warning', label: 'Di Luar Jadwal', bar: 'bg-amber-500 text-white dark:bg-amber-600' },
}

function hasCameraAPI() {
  try {
    return typeof window !== 'undefined' && 'BarcodeDetector' in window && !!navigator.mediaDevices?.getUserMedia
  } catch {
    return false
  }
}

export default function ScanTicket({ nvdata }) {
  const toast = useToast()
  const [events, setEvents] = useState([])
  const [event, setEvent] = useState('')
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [code, setCode] = useState('')
  const [busy, setBusy] = useState(false)
  const [result, setResult] = useState(null)
  const [summary, setSummary] = useState(null)
  const [cameraOn, setCameraOn] = useState(false)
  const [cameraError, setCameraError] = useState('')
  const videoRef = useRef(null)
  const streamRef = useRef(null)
  const scanRef = useRef(false)
  const abortRef = useRef(null)

  useEffect(() => {
    controller.init()
    abortRef.current?.abort()
    const ctrl = new AbortController()
    abortRef.current = ctrl
    setLoading(true)
    setError('')
    read_data({}, ctrl.signal)
      .then((data) => {
        if (ctrl.signal.aborted) return
        setEvents(data.events || [])
        if ((data.events || []).length === 1) setEvent(data.events[0].code)
      })
      .catch((err) => {
        if (err && err.name === 'AbortError') return
        if (!ctrl.signal.aborted) setError(err.message)
      })
      .finally(() => {
        if (!ctrl.signal.aborted) setLoading(false)
      })
    return () => ctrl.abort()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  useEffect(() => () => {
    abortRef.current?.abort()
    stopCamera()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  // Ringkasan masuk vs total per event terpilih.
  useEffect(() => {
    if (!event) {
      setSummary(null)
      return
    }
    let alive = true
    read_data({ event })
      .then((data) => {
        if (alive) setSummary(data.summary)
      })
      .catch(() => {
        if (alive) setSummary(null)
      })
    return () => {
      alive = false
    }
  }, [event, result])

  function stopCamera() {
    try {
      streamRef.current?.getTracks()?.forEach((t) => t.stop())
    } catch {
      // abaikan
    }
    streamRef.current = null
    setCameraOn(false)
  }

  async function handler_camera_toggle() {
    if (cameraOn) {
      stopCamera()
      return
    }
    setCameraError('')
    try {
      const stream = await navigator.mediaDevices.getUserMedia({ video: { facingMode: 'environment' } })
      streamRef.current = stream
      if (videoRef.current) {
        videoRef.current.srcObject = stream
        await videoRef.current.play()
      }
      setCameraOn(true)
      scanRef.current = false
      cameraLoop()
    } catch {
      setCameraError('Kamera tidak bisa dibuka. Pakai input kode manual.')
    }
  }

  async function cameraLoop() {
    try {
      const detector = new window.BarcodeDetector({ formats: ['qr_code'] })
      const tick = async () => {
        if (!streamRef.current || scanRef.current) return
        try {
          const codes = await detector.detect(videoRef.current)
          if (codes && codes.length > 0 && codes[0].rawValue && !scanRef.current) {
            scanRef.current = true
            handler_scan_click(codes[0].rawValue, '').finally(() => {
              scanRef.current = false
              setTimeout(() => {
                if (streamRef.current) tick()
              }, 1200)
            })
            return
          }
        } catch {
          // abaikan satu frame gagal
        }
        if (streamRef.current) setTimeout(tick, 400)
      }
      tick()
    } catch {
      setCameraError('Pemindai QR tak tersedia. Pakai input kode manual.')
    }
  }

  async function handler_scan_click(token, manualCode) {
    const payload = token ? { token } : { code: (manualCode || code).trim() }
    if (!payload.token && !payload.code) {
      toast.warning('Isi kode tiket dulu.', { title: 'Batal' })
      return
    }
    if (!event) {
      toast.warning('Pilih event dulu.', { title: 'Batal' })
      return
    }
    setBusy(true)
    try {
      const res = await process_scan({ ...payload, event })
      setResult(res)
      setCode('')
    } catch (ex) {
      toast.error(ex.message, { title: 'Scan gagal' })
    } finally {
      setBusy(false)
    }
  }

  const style = (result && RESULT_STYLE[result.result]) || null

  return (
    <div className="flex flex-col gap-4">
      <Card>
        <div className="mb-4 flex flex-wrap items-center justify-between gap-3">
          <CardTitle description={nvdata?.label || 'Validasi tiket di pintu masuk.'}>Scan Ticket</CardTitle>
          <Button variant="secondary" size="sm" onClick={() => window.location.reload()}>
            Muat ulang
          </Button>
        </div>
        {loading ? (
          <Skeleton className="h-24" />
        ) : error ? (
          <Alert tone="error" title="Gagal memuat" closable onClose={() => setError('')}>
            {error}
          </Alert>
        ) : (
          <div className="flex flex-col gap-3">
            <label className="flex max-w-md flex-col gap-1 text-xs font-medium text-zinc-500">
              Event
              <select value={event} onChange={(e) => { setEvent(e.target.value); setResult(null) }} className={SELECT_CLASS}>
                <option value="">— Pilih event —</option>
                {events.map((e) => (
                  <option key={e.code} value={e.code}>
                    {e.title}
                  </option>
                ))}
              </select>
            </label>
            {summary && (
              <div>
                <div className="mb-1 flex items-baseline justify-between gap-2 text-sm">
                  <span className="font-medium text-zinc-700 dark:text-zinc-200">Sudah masuk</span>
                  <span className="font-mono text-xs text-zinc-500">
                    {summary.checked_in}/{summary.total}
                  </span>
                </div>
                <ProgressBar value={summary.checked_in} max={Math.max(1, summary.total)} tone="emerald" />
              </div>
            )}
          </div>
        )}
      </Card>

      {result && style && (
        <div className={`rounded-2xl p-5 text-center shadow-sm ${style.bar}`}>
          <p className="text-2xl font-bold">{style.label}</p>
          {(result.holder_name || result.ticket_code) && (
            <p className="mt-1 truncate text-sm opacity-90">
              {[result.holder_name, result.type_name, result.ticket_code].filter(Boolean).join(' · ')}
            </p>
          )}
          {result.detail && <p className="mt-1 text-sm opacity-90">{result.detail}</p>}
          <p className="mt-2 font-mono text-xs opacity-80">
            Masuk {result.checked_in}/{result.total}
          </p>
        </div>
      )}

      <Card>
        <CardTitle description={hasCameraAPI() ? 'Arahkan kamera ke QR tiket.' : 'Perangkat ini tidak mendukung kamera, pakai input manual.'}>
          {hasCameraAPI() ? 'Kamera' : 'Input manual'}
        </CardTitle>
        {hasCameraAPI() && (
          <div className="flex flex-col gap-2">
            <Button size="sm" variant="secondary" onClick={handler_camera_toggle}>
              {cameraOn ? 'Matikan kamera' : 'Nyalakan kamera'}
            </Button>
            {cameraOn && <video ref={videoRef} playsInline muted className="w-full rounded-xl bg-black" />}
            {cameraError && <p className="text-xs text-amber-700 dark:text-amber-300">{cameraError}</p>}
          </div>
        )}
        <form
          onSubmit={(e) => {
            e.preventDefault()
            handler_scan_click('', code)
          }}
          className="mt-3 flex gap-2"
        >
          <input
            value={code}
            onChange={(e) => setCode(e.target.value)}
            placeholder="Ketik/paste kode tiket…"
            className="min-w-0 flex-1 rounded-lg border border-zinc-300 bg-white px-2.5 py-1.5 font-mono text-sm dark:border-zinc-700 dark:bg-zinc-900"
          />
          <Button size="sm" loading={busy}>
            Scan
          </Button>
        </form>
      </Card>
    </div>
  )
}
