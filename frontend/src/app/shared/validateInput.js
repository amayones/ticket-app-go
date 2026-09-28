// Padanan loop handler_validasi_input + validate_field[] FRM langit_v2.
// validate_field: [{ field, type: 'text'|'number'|'date', msg }]
// Kembali { valid, message, field } — field untuk fokus penanda invalid.
export function validateInput(values, validate_field = []) {
  for (const item of validate_field) {
    const value = values[item.field]
    let isInvalid = false
    if (item.type === 'number') {
      const num = parseFloat(value)
      isInvalid = value === '' || value == null || Number.isNaN(num) || num <= 0
    } else if (item.type === 'date') {
      isInvalid = value === '' || value == null || Number.isNaN(Date.parse(value))
    } else {
      isInvalid = value === '' || value == null || (typeof value === 'string' && value.trim() === '')
    }
    if (isInvalid) {
      return { valid: false, message: item.msg, field: item.field }
    }
  }
  return { valid: true, message: '', field: null }
}
