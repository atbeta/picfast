import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'
import path from 'path'

const backendUrl = process.env.VITE_BACKEND_URL ?? 'http://localhost:8080'

// https://vite.dev/config/
export default defineConfig({
  plugins: [react(), tailwindcss()],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
    },
  },
  build: {
    rollupOptions: {
      output: {
        manualChunks(id) {
          if (!id.includes('node_modules')) {
            return undefined
          }

          if (id.includes('react') || id.includes('scheduler')) {
            return 'react-vendor'
          }
          if (id.includes('react-router')) {
            return 'router-vendor'
          }
          if (id.includes('@tanstack/react-query')) {
            return 'query-vendor'
          }
          if (id.includes('react-hook-form') || id.includes('zod') || id.includes('@hookform/resolvers')) {
            return 'form-vendor'
          }
          if (id.includes('i18next') || id.includes('react-i18next')) {
            return 'i18n-vendor'
          }
          if (id.includes('lucide-react')) {
            return 'icon-vendor'
          }
          if (id.includes('sonner')) {
            return 'toast-vendor'
          }

          return 'vendor'
        },
      },
    },
  },
  server: {
    proxy: {
      '/api': backendUrl,
      '/i': backendUrl,
      '/t': backendUrl,
    },
  },
})
