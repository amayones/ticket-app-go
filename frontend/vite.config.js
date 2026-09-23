import react from '@vitejs/plugin-react'
import { defineConfig } from 'vite'

export default defineConfig({
  plugins: [react()],
  base: '/',
  build: {
    outDir: 'dist',
  },
  server: {
    proxy: {
      '/api': 'http://localhost:1067',
    },
  },
})
