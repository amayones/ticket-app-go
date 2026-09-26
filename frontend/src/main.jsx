import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import './index.css'
import './components/ui.css'
import { initTheme } from './components/theme.js'
import { APP_LOGO, APP_NAME } from './brand.js'
import App from './App.jsx'

initTheme()
// Judul tab & favicon mengikuti APP_NAME / APP_LOGO (.env root).
// Nilai di index.html hanya fallback sebelum React berjalan.
document.title = APP_NAME
const icon = document.querySelector('link[rel="icon"]')
if (icon && APP_LOGO) icon.setAttribute('href', APP_LOGO)

createRoot(document.getElementById('root')).render(
  <StrictMode>
    <App />
  </StrictMode>,
)
