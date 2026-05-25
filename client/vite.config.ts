import { defineConfig } from 'vitest/config'
import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'

// https://vite.dev/config/
export default defineConfig({
  plugins: [react(), tailwindcss()],
  test: {
    environment: 'jsdom',
    include: ['**/*.{test,spec}.?(c|m)[jt]s?(x)', '**/*_test.?(c|m)[jt]s?(x)'],
    setupFiles: ['./src/test/setup.ts'],
  },
})
