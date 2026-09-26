import { defineConfig, loadEnv } from 'vite'
import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, '../', '')
  const appPort = env.APP_PORT || '1067'
  return {
    plugins: [react(), tailwindcss()],
    base: '/',
    // Nama aplikasi untuk branding (judul tab + kartu login). Satu sumber:
    // APP_NAME di .env root, jadi ganti nama aplikasi cukup edit satu baris.
    define: {
      __APP_NAME__: JSON.stringify(env.APP_NAME || 'Go Core'),
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
