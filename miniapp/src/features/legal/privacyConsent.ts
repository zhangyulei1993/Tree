export const PRIVACY_CONSENT_STORAGE_KEY = 'tree_privacy_consent_v1'

type PrivacyConsentListener = () => void
const promptListeners = new Set<PrivacyConsentListener>()

export function hasPrivacyConsent(): boolean {
  return uni.getStorageSync(PRIVACY_CONSENT_STORAGE_KEY) === 'agreed'
}

export function setPrivacyConsent(): void {
  uni.setStorageSync(PRIVACY_CONSENT_STORAGE_KEY, 'agreed')
}

export function clearPrivacyConsent(): void {
  uni.removeStorageSync(PRIVACY_CONSENT_STORAGE_KEY)
}

export function subscribePrivacyConsentPrompt(listener: PrivacyConsentListener): () => void {
  promptListeners.add(listener)
  return () => {
    promptListeners.delete(listener)
  }
}

export function promptPrivacyConsentIfNeeded(): boolean {
  if (hasPrivacyConsent()) {
    return false
  }
  for (const listener of promptListeners) {
    listener()
  }
  return true
}

export function openUserAgreement(): void {
  uni.navigateTo({ url: '/pages/legal/user-agreement' })
}

export function openPrivacyPolicy(): void {
  uni.navigateTo({ url: '/pages/legal/privacy-policy' })
}

export function resolvePrivacyAuthorizeError(error: unknown): string {
  const errMsg = typeof error === 'object' && error !== null && 'errMsg' in error
    ? String((error as { errMsg?: unknown }).errMsg || '')
    : ''
  if (/cancel|deny|拒绝|未同意|auth deny/i.test(errMsg)) {
    return '你已拒绝微信隐私授权，需同意后才能登录。'
  }
  return '微信隐私授权未完成，请重试后再登录。'
}

/** 触发微信隐私授权（仅微信小程序运行时可用） */
export function requestWechatPrivacyAuthorize(): Promise<void> {
  return new Promise((resolve, reject) => {
    const wxApi = (globalThis as { wx?: { requirePrivacyAuthorize?: (options: {
      success?: () => void
      fail?: (error: unknown) => void
    }) => void } }).wx
    if (typeof wxApi?.requirePrivacyAuthorize === 'function') {
      wxApi.requirePrivacyAuthorize({
        success: () => resolve(),
        fail: (error) => reject(error)
      })
      return
    }
    resolve()
  })
}

/** 登录前确保本地隐私同意；未同意时直接拉起微信授权，成功后写入本地 consent。 */
export async function ensurePrivacyConsentForLogin(): Promise<void> {
  if (hasPrivacyConsent()) {
    return
  }
  try {
    await requestWechatPrivacyAuthorize()
  } catch (error) {
    throw new Error(resolvePrivacyAuthorizeError(error))
  }
  setPrivacyConsent()
}
