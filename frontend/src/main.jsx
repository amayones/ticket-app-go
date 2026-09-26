import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import './index.css'
import './components/ui.css'
import { initTheme } from './components/theme.js'
import { APP_NAME } from './appName.js'
import App from './App.jsx'

initTheme()
// Judul tab mengikuti APP_NAME (.env root); index.html hanya fallback.
document.title = APP_NAME

createRoot(document.getElementById('root')).render(
  <StrictMode>
    <App />
  </StrictMode>,
)
