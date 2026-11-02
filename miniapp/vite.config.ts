import uniPlugin from '@dcloudio/vite-plugin-uni'
import { defineConfig, loadEnv } from 'vite'

import { createBuildInfo } from '../scripts/build-info.mjs'

const uni = typeof uniPlugin === 'function' ? uniPlugin : uniPlugin.default

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), '')
  const apiProxyTarget = env.VITE_API_PROXY_TARGET || 'http://127.0.0.1:8080'
  const apiBaseUrl = env.VITE_API_BASE_URL || '/api'
  const buildInfo = createBuildInfo({
    app: 'tree-miniapp',
    rootDir: process.cwd(),
    mode,
    apiBaseUrl
  })

  return {
    plugins: [uni()],
    define: {
      __TREE_BUILD_INFO__: JSON.stringify(buildInfo)
    },
    server: {
      // 仅 H5 dev 使用；微信小程序不会走此 proxy。
      proxy: {
        '/api': {
          target: apiProxyTarget,
          changeOrigin: true
        }
      }
    },
    preview: {
      host: '127.0.0.1',
      port: Number(env.TREE_H5_PREVIEW_PORT || 5199),
      strictPort: true,
      proxy: {
        '/api': {
          target: apiProxyTarget,
          changeOrigin: true
        }
      }
    }
  }
})
