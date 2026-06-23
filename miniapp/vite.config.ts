import uniPlugin from '@dcloudio/vite-plugin-uni'
import { defineConfig } from 'vite'

const uni = typeof uniPlugin === 'function' ? uniPlugin : uniPlugin.default

export default defineConfig({
  plugins: [uni()],
  server: {
    // 仅 H5 dev 使用；微信小程序不会走此 proxy。
    proxy: {
      '/api': {
        target: 'http://127.0.0.1:8080',
        changeOrigin: true
      }
    }
  }
})
