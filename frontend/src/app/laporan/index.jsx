// ===== TEMPLATE SERAGAM app/*: import satu pintu + flags true/false =====
// Nonaktifkan blok cukup set false; return di bawah tidak perlu diubah.
import { Alert, Avatar, Badge, Button, Card, CardTitle, ConfirmDialog, EmptyState, Icon, Modal, Pagination, Skeleton, SkeletonRows, TextField, PasswordInput, Tooltip, formatTime, useCallback, useEffect, useMemo, useState, api, useSmoothLoading, useToast, PAGE_SIZE, SKELETON_ROWS, SELECT_CLASS, INPUT_CLASS, PageShell, usePageList, usePageListObj } from '../shared/all.js'

const FEATURES = { header: true, refresh: false, filter: false, tabs: false, create: false, edit: false, remove: false, pagination: false, empty: true, error: true, confirmDialog: false, extraActions: false }

export const meta = { label: 'Laporan', icon: 'list', order: 1 }

// Contoh menu BIASA: satu folder = satu halaman, tanpa parent.
// MCONTROL=laporan -> app/laporan/. Bandingkan dengan header PARENT
// MENU_KEUANGAN (tanpa folder) dan app/arus_kas/ (CHILD dari MENU_KEUANGAN).
// Lihat tutorial/README.md Bagian 2 untuk cara membuat menu seperti ini.
export default function Laporan() {
  return (
    <Card>
      {FEATURES.header && (
      <div className="mb-4">
        <CardTitle description="Contoh menu biasa di modul REPORT (permission MENU_LAPORAN). Belum ada data bisnis — halaman ini hanya untuk membandingkan tampilan dengan menu bersarang.">
          Laporan
        </CardTitle>
      </div>
      )}
      {FEATURES.empty && (
      <EmptyState
        title="Menu contoh"
        description="Isi halaman ini dengan tabel/grafik laporan mengikuti pola menu lain (load → SkeletonRows → Alert → EmptyState → Pagination)."
      />
      )}
    </Card>
  )
}
