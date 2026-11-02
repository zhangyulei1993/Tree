import { fileURLToPath, URL } from 'node:url'

import vue from '@vitejs/plugin-vue'
import Components from 'unplugin-vue-components/vite'
import { ElementPlusResolver } from 'unplugin-vue-components/resolvers'
import { defineConfig, loadEnv } from 'vite'

import { createBuildInfo } from '../scripts/build-info.mjs'

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), '')
  const apiProxyTarget = env.VITE_API_PROXY_TARGET || 'http://127.0.0.1:8080'
  const apiBaseUrl = env.VITE_API_BASE_URL || '/api'
  const buildInfo = createBuildInfo({
    app: 'tree-admin-web',
    rootDir: process.cwd(),
    mode,
    apiBaseUrl
  })

  return {
    plugins: [
      vue(),
      Components({
        resolvers: [ElementPlusResolver({ importStyle: 'css' })]
      })
    ],
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
      port: 5174,
      proxy: {
        '/api': {
          target: apiProxyTarget,
          changeOrigin: true
        }
      }
    },
    build: {
      // Keep the initial shell small and give long-lived dependencies stable cache keys.
      // Route components are already lazy-loaded in src/router; these groups prevent
      // Element Plus and the Vue runtime from being pulled into the entry chunk.
      rollupOptions: {
        output: {
          manualChunks(id) {
            if (!id.includes('node_modules')) return undefined
            if (id.includes('/element-plus/')) {
              const component = id.match(/element-plus\/es\/components\/([^/]+)/)?.[1]
              return component ? `vendor-element-${component}` : 'vendor-element-core'
            }
            if (id.includes('/@element-plus/icons-vue/')) return 'vendor-element-icons'
            if (id.includes('/vue-router/') || id.includes('/pinia/') || id.includes('/vue/')) return 'vendor-vue'
            if (id.includes('/axios/')) return 'vendor-http'
            return 'vendor'
          }
        }
      }
    }
  }
})
