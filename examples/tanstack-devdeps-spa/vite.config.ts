import { defineConfig } from 'vite'
import { tanstackStart } from '@tanstack/react-start/plugin/vite'
import viteReact from '@vitejs/plugin-react'

const config = defineConfig({
  plugins: [tanstackStart(), viteReact()],
})

export default config
