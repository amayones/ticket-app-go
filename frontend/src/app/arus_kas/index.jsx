// ===== TEMPLATE SERAGAM app/*: import satu pintu + flags true/false =====
// Nonaktifkan blok cukup set false; return di bawah tidak perlu diubah.
import { Alert, Avatar, Badge, Button, Card, CardTitle, ConfirmDialog, EmptyState, Icon, Modal, Pagination, Skeleton, SkeletonRows, TextField, PasswordInput, Tooltip, formatTime, useCallback, useEffect, useMemo, useState, api, useSmoothLoading, useToast, PAGE_SIZE, SKELETON_ROWS, SELECT_CLASS, INPUT_CLASS, PageShell, usePageList, usePageListObj } from '../shared/all.js'

const FEATURES = { header: true, refresh: false, filter: false, tabs: false, create: false, edit: false, remove: false, pagination: false, empty: true, error: true, confirmDialog: false, extraActions: false }

export const meta = { label: 'Arus Kas', icon: 'list', order: 3 }

// Contoh menu CHILD: baris CPMENU MENU_ARUS_KAS (MCONTROL=arus_kas,
// PARENT_CODE = MENU_KEUANGAN), jadi di sidebar tampil menjorok di bawah
// header parent "Keuangan". Foldernya datar: app/arus_kas/.
export default function ArusKas() {
  return (
    <Card>
      {FEATURES.header && (
      <div className="mb-4">
        <CardTitle description="Contoh menu child di bawah menu parent (permission MENU_ARUS_KAS, parent MENU_KEUANGAN).">
          Arus Kas
        </CardTitle>
      </div>
      )}
      {FEATURES.empty && (
      <EmptyState
        title="Menu contoh (child)"
        description="Hierarki sidebar ini berasal dari CPMENU: MENU_KIND = CHILD dan PARENT_CODE = MENU_KEUANGAN."
      />
      )}
    </Card>
  )
}
