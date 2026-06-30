export interface BuildInfoOptions {
  app: string
  rootDir: string
  mode?: string
  apiBaseUrl?: string
}

export interface BuildInfo {
  app: string
  version: string
  commit: string
  buildTime: string
  environment: string
  apiBaseUrl: string
  dirty: boolean
}

export function createBuildInfo(options: BuildInfoOptions): BuildInfo
