export const PRIVACY_CONSENT_STORAGE_KEY = 'tree_privacy_consent_v1'

export function hasPrivacyConsent(): boolean {
  return uni.getStorageSync(PRIVACY_CONSENT_STORAGE_KEY) === 'agreed'
}

export function setPrivacyConsent(): void {
  uni.setStorageSync(PRIVACY_CONSENT_STORAGE_KEY, 'agreed')
}

export function clearPrivacyConsent(): void {
  uni.removeStorageSync(PRIVACY_CONSENT_STORAGE_KEY)
}

export function openUserAgreement(): void {
  uni.navigateTo({ url: '/pages/legal/user-agreement' })
}

export function openPrivacyPolicy(): void {
  uni.navigateTo({ url: '/pages/legal/privacy-policy' })
}

/** 触发微信隐私授权（仅微信小程序） */
export function requestWechatPrivacyAuthorize(): Promise<void> {
  return new Promise((resolve, reject) => {
    // #ifdef MP-WEIXIN
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
    // #endif
    resolve()
  })
}
