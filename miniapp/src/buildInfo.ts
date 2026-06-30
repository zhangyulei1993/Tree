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

export const buildInfo = __TREE_BUILD_INFO__
