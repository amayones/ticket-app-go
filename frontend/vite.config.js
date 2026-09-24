import { defineConfig, loadEnv } from 'vite'
import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, '../', '')
  const appPort = env.APP_PORT || '1067'
  return {
    plugins: [react(), tailwindcss()],
    base: '/',
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
      },
    },
    preview: {
      port: 4173,
    },
  }
})
