type SystemInfo = {
  uniPlatform?: string
}

type UniRuntime = {
  getSystemInfoSync?: () => SystemInfo
}

function readUniRuntime(): UniRuntime | null {
  try {
    if (typeof uni !== 'undefined' && uni && typeof uni.getSystemInfoSync === 'function') {
      return uni as UniRuntime
    }
  } catch {
    // uni may be unavailable outside uni-app runtimes.
  }
  return null
}

export function resolveUniPlatform(): string {
  const runtime = readUniRuntime()
  if (!runtime?.getSystemInfoSync) {
    return ''
  }
  try {
    const info = runtime.getSystemInfoSync()
    const platform = typeof info?.uniPlatform === 'string' ? info.uniPlatform.trim() : ''
    return platform
  } catch {
    return ''
  }
}

export function isMpWeixinPlatform(platform = resolveUniPlatform()): boolean {
  return platform === 'mp-weixin'
}

export function resolveWechatLoginUnsupportedMessage(platform = resolveUniPlatform()): string | null {
  if (isMpWeixinPlatform(platform)) {
    return null
  }
  if (platform === 'h5') {
    return '微信快捷登录仅支持微信小程序。H5 测试请使用 scripts/tree-fresh-quota-e2e-session.mjs 注入全新模拟微信登录态。'
  }
  return '微信快捷登录仅支持微信小程序，请在微信中打开小程序后登录。'
}
