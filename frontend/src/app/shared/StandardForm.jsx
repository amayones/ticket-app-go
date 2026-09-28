import { TextField, PasswordInput } from '../../components'
import { INPUT_CLASS } from './config.js'

// Form standar ala FRM<mod>.js langit_v2: kerangka sama,
// yang beda hanya FIELDS (items) per halaman.
// Field: { name, label, type: 'text'|'email'|'password'|'number'|'textarea'|'select'|'checkbox',
//   placeholder, hint, options[{value,label}], rows, required }.
export default function StandardForm({ fields = [], values = {}, onChange = () => {} }) {
  function set(name, val) {
    onChange({ ...values, [name]: val })
  }

  return (
    <div className="flex flex-col gap-4">
      {fields.map((f) => {
        if (f.type === 'password') {
          return (
            <PasswordInput
              key={f.name}
              label={f.label}
              value={values[f.name] || ''}
              onChange={(e) => set(f.name, e.target.value)}
              autoComplete="new-password"
              placeholder={f.placeholder}
              hint={f.hint}
            />
          )
        }
        if (f.type === 'textarea') {
          return (
            <label key={f.name} className="block">
              <span className="mb-1.5 block text-sm font-medium text-zinc-700 dark:text-zinc-200">{f.label}</span>
              <textarea
                value={values[f.name] || ''}
                onChange={(e) => set(f.name, e.target.value)}
                rows={f.rows || 5}
                placeholder={f.placeholder}
                className={`${INPUT_CLASS} font-mono`}
              />
              {f.hint && <span className="mt-1.5 block text-xs text-zinc-500">{f.hint}</span>}
            </label>
          )
        }
        if (f.type === 'select') {
          return (
            <label key={f.name} className="block">
              <span className="mb-1.5 block text-sm font-medium text-zinc-700 dark:text-zinc-200">{f.label}</span>
              <select value={values[f.name] || ''} onChange={(e) => set(f.name, e.target.value)} className={INPUT_CLASS}>
                {(f.options || []).map((o) => (
                  <option key={o.value} value={o.value}>{o.label || o.value}</option>
                ))}
              </select>
            </label>
          )
        }
        if (f.type === 'checkbox') {
          return (
            <label key={f.name} className="flex cursor-pointer items-center gap-2 text-sm text-zinc-700 dark:text-zinc-200">
              <input
                type="checkbox"
                checked={!!values[f.name]}
                onChange={(e) => set(f.name, e.target.checked)}
                className="h-4 w-4 accent-violet-600"
              />
              {f.label}
            </label>
          )
        }
        if (f.type === 'number') {
          return (
            <TextField
              key={f.name}
              label={f.label}
              type="number"
              value={values[f.name] ?? ''}
              onChange={(e) => set(f.name, e.target.value)}
              placeholder={f.placeholder}
              hint={f.hint}
            />
          )
        }
        return (
          <TextField
            key={f.name}
            label={f.label}
            type={f.type || 'text'}
            value={values[f.name] || ''}
            onChange={(e) => set(f.name, e.target.value)}
            autoComplete="off"
            placeholder={f.placeholder}
            hint={f.hint}
          />
        )
      })}
    </div>
  )
}
