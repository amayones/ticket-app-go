// Controller standar ala C<mod>.js langit_v2 — 6 fungsi, nama SAMA PERSIS.
// Dipakai: import { controller } from './controller.js' (dibuat via
// createController, tinggal ganti mcontrol seperti kebiasaan clone di langit_v2).
// Varian btnew = A: delegasi ke GRID.handler_btnew_click().
import { formatAmount, formatDate } from '../../components/format.js'

export { formatAmount, formatDate }

function toastError(msg) {
  try {
    window.dispatchEvent(new CustomEvent('go-core:toast-error', { detail: msg }))
  } catch {
    console.error(msg)
  }
}

export function createController({ mcontrol }) {
  const var_definition = {}

  function init() {
    renderpage()
  }

  function renderpage() {
    try {
      console.log(`renderer controller ${mcontrol}`)
    } catch (ex) {
      toastError(ex.message)
    }
  }

  function btrefresh_click(reload) {
    try {
      reload()
    } catch (ex) {
      toastError(ex.message)
    }
  }

  function btnew_click(handler_btnew_click) {
    try {
      return handler_btnew_click()
    } catch (ex) {
      toastError(ex.message)
    }
  }

  return { var_definition, init, renderpage, formatAmount, formatDate, btrefresh_click, btnew_click }
}
