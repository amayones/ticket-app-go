import { createContext, useContext } from 'react'

// Context toast dipisah dari provider agar file provider hanya mengekspor
// komponen (lolos aturan oxlint react/only-export-components).
export const ToastContext = createContext(null)

export default function useToast() {
  const ctx = useContext(ToastContext)
  if (!ctx) throw new Error('useToast harus dipakai di dalam <ToastProvider>')
  return ctx
}
