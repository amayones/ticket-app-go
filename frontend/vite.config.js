import { defineConfig, loadEnv } from 'vite'
import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, '../', '')
  const appPort = env.APP_PORT || '1067'
  return {
    plugins: [react(), tailwindcss()],
    base: '/',
    // Branding aplikasi (nama + logo) untuk judul tab, sidebar, dan kartu
    // login. Satu sumber: APP_NAME / APP_LOGO di .env root, jadi ganti
    // nama aplikasi cukup edit .env lalu build ulang.
    define: {
      __APP_NAME__: JSON.stringify(env.APP_NAME || 'Go Core'),
      __APP_LOGO__: JSON.stringify(env.APP_LOGO || '/favicon.svg'),
    },
    build: {
      outDir: 'dist',
      emptyOutDir: true,
    },
    server: {
      proxy: {
        '/api': {
          target: env.VITE_API_URL || `http://localhost:${appPort}`,
          changeOrigin: true,
          secure: false,
        },
        // /healthz juga milik backend Go (dipakai badge status Dashboard).
        '/healthz': {
          target: env.VITE_API_URL || `http://localhost:${appPort}`,
          changeOrigin: true,
          secure: false,
        },
      },
    },
    preview: {
      port: 4173,
    },
  }
})
