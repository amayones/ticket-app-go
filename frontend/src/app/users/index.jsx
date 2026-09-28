// View User Account — padanan <mod>.js langit_v2 (shell + toolbar + items).
// Controller: ./controller.js (6 fungsi). Store: read_data. Tabel: CPUSER.
import { useEffect, useRef, useState } from 'react'
import { Button, StandardPage, DownloadButton, api, useStandardController } from '../shared/all.js'
import { read_data } from './api.js'
import { listRoles } from '../role_permission/api.js'
import { controller } from './controller.js'
import GRID from './GRID.jsx'
import FRM from './FRM.jsx'
import EditUserModal from './components/EditUserModal.jsx'

export const meta = { label: 'User Account', icon: 'users', order: 1 }

const GRID_COLUMNS = [
  { header: 'Kode', dataIndex: 'code' },
  { header: 'Username', dataIndex: 'username' },
  { header: 'Email', dataIndex: 'email' },
  { header: 'Role', dataIndex: 'role_code' },
]

export default function UsersList({ onAccountDeleted, nvdata }) {
  const me = api.currentUser()
  const [showFRM, setShowFRM] = useState(false)
  const [editing, setEditing] = useState(null)
  const [roles, setRoles] = useState([])
  const gridRef = useRef(null)

  const { rows: users, offset, setOffset, loading, showLoading, error, setError, load } =
    useStandardController(({ limit, offset }) => read_data({ limit, offset }), {})

  useEffect(() => {
    controller.init()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  function openFRM() {
    listRoles().then(setRoles).catch(() => setRoles([]))
    setShowFRM(true)
  }

  function afterSave() {
    if (offset === 0) load()
    else setOffset(0)
  }

  return (
    <div className="flex flex-col gap-4">
      <StandardPage
        title={nvdata?.label || 'Daftar Pengguna'}
        description="Kelola akun yang terdaftar. Anda hanya dapat mengubah akun milik sendiri."
        loading={loading}
        showLoading={showLoading}
        error={error}
        errorTitle="Gagal memuat data"
        onClearError={() => setError('')}
        onRefresh={() => controller.btrefresh_click(load)}
        onCreate={() => controller.btnew_click(() => gridRef.current?.handler_btnew_click())}
        createLabel="New Input"
        extraActions={<DownloadButton rows={users} columns={GRID_COLUMNS} filename="users" />}
        items={users}
        emptyTitle="Belum ada pengguna"
        emptyDescription="Data kosong pada halaman ini. Coba kembali ke halaman sebelumnya atau muat ulang."
        emptyAction={
          <Button variant="secondary" size="sm" onClick={() => (offset === 0 ? load() : setOffset(0))}>
            Muat ulang
          </Button>
        }
        offset={offset}
        onPage={(next) => next >= 0 && setOffset(next)}
      >
        <GRID
          ref={gridRef}
          rows={users}
          me={me}
          reload={load}
          openFRM={openFRM}
          openEdit={setEditing}
          onAccountDeleted={onAccountDeleted}
        />
      </StandardPage>

      <FRM open={showFRM} roles={roles} onClose={() => setShowFRM(false)} onSaved={afterSave} />

      {editing && (
        <EditUserModal
          key={editing.code}
          user={editing}
          onClose={() => setEditing(null)}
          onSaved={() => load()}
        />
      )}
    </div>
  )
}
