import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'

// https://vite.dev/config/
export default defineConfig({
  plugins: [react(), tailwindcss()],
  test: {
    include: ['**/*.{test,spec}.?(c|m)[jt]s?(x)', '**/*_test.?(c|m)[jt]s?(x)'],
  },
})
