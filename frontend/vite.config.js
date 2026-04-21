import { defineConfig } from 'vite'
import { svelte } from '@sveltejs/vite-plugin-svelte'

// https://vite.dev/config/
export default defineConfig({
  plugins: [svelte()],
  preview: {
    host: '0.0.0.0',
    port: parseInt(process.env.PORT) || 3000,
    allowedHosts: 'all',
  },
  server: {
    host: '0.0.0.0',
    port: parseInt(process.env.PORT) || 3000,
  },
})
