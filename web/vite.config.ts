import react from '@vitejs/plugin-react'
import { defineConfig } from 'vitest/config'

export default defineConfig({
  plugins: [react()],
  server: {
    // `npm run dev` proxies the API to the backend running on its own port,
    // removing the /api prefix like the reverse proxy does.
    proxy: {
      '/api': {
        target: process.env.SIGTA_API_URL ?? 'http://localhost:8080',
        rewrite: (path) => path.replace(/^\/api/, ''),
      },
    },
  },
  test: {
    environment: 'jsdom',
    setupFiles: ['./src/test/setup.ts'],
  },
})
