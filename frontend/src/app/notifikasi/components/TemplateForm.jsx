// Form template ala FRM<mod>.js langit_v2: state lokal + StandardForm + FIELDS items.
import { Button, StandardForm, useState } from '../../shared/all.js'
import { TEMPLATE_FIELDS } from '../columns.jsx'

export default function TemplateForm({ initial, saving, onSubmit, onCancel }) {
  const [values, setValues] = useState({
    name: initial?.name || '',
    channel: initial?.channel || 'EMAIL',
    subject: initial?.subject || '',
    body: initial?.body || '',
    is_active: initial ? !!initial.is_active : true,
  })

  return (
    <div className="flex flex-col gap-4">
      <StandardForm fields={TEMPLATE_FIELDS} values={values} onChange={setValues} />
      <div className="flex justify-end gap-2">
        <Button variant="secondary" onClick={onCancel} disabled={saving}>
          Batal
        </Button>
        <Button onClick={() => onSubmit(values)} loading={saving}>
          Simpan template
        </Button>
      </div>
    </div>
  )
}
