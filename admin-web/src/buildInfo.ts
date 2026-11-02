export interface TreeBuildInfo {
  app: string
  version: string
  commit: string
  buildTime: string
  environment: string
  apiBaseUrl: string
  dirty: boolean
}

declare const __TREE_BUILD_INFO__: TreeBuildInfo

declare global {
  interface Window {
    __TREE_BUILD__?: TreeBuildInfo
  }
}

export const buildInfo = __TREE_BUILD_INFO__

if (typeof window !== 'undefined') {
  window.__TREE_BUILD__ = buildInfo
}
