import { fileURLToPath, URL } from 'node:url'

import vue from '@vitejs/plugin-vue'
import { defineConfig, loadEnv } from 'vite'

import { createBuildInfo } from '../scripts/build-info.mjs'

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), '')
  const apiProxyTarget = env.VITE_API_PROXY_TARGET || 'http://127.0.0.1:8080'
  const apiBaseUrl = env.VITE_API_BASE_URL || '/api'
  const buildInfo = createBuildInfo({
    app: 'tree-web',
    rootDir: process.cwd(),
    mode,
    apiBaseUrl
  })

  return {
    plugins: [vue()],
    define: {
      __TREE_BUILD_INFO__: JSON.stringify(buildInfo)
    },
    resolve: {
      alias: {
        '@': fileURLToPath(new URL('./src', import.meta.url))
      }
    },
    server: {
      host: '0.0.0.0',
      port: 5173,
      proxy: {
        '/api': {
          target: apiProxyTarget,
          changeOrigin: true
        }
      }
    }
  }
})
