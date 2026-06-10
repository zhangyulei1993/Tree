import uniPlugin from '@dcloudio/vite-plugin-uni'
import { defineConfig } from 'vite'

const uni = typeof uniPlugin === 'function' ? uniPlugin : uniPlugin.default

export default defineConfig({
  plugins: [uni()],
  server: {
    proxy: {
      '/api': {
        target: 'http://127.0.0.1:8080',
        changeOrigin: true
      }
    }
  }
})
