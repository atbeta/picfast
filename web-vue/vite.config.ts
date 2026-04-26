import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import tailwindcss from '@tailwindcss/vite'

export default defineConfig({
  plugins: [vue(), tailwindcss()],
  server: {
    proxy: {
      '/api': 'http://localhost:8080',
      '/i': 'http://localhost:8080',
      '/t': 'http://localhost:8080',
    },
  },
  build: {
    outDir: '../web-dist',
    emptyOutDir: true,
  },
})
