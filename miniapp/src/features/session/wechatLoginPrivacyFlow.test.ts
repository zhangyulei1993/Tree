import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { dirname, join } from 'node:path'
import { fileURLToPath } from 'node:url'
import { test } from 'node:test'

import {
  clearPrivacyConsent,
  ensurePrivacyConsentForLogin,
  hasPrivacyConsent,
  setPrivacyConsent
} from '../legal/privacyConsent'

const here = dirname(fileURLToPath(import.meta.url))
const miniappSrc = join(here, '..', '..')

const storage = new Map<string, unknown>()

function installUniMock() {
  globalThis.uni = {
    getStorageSync(key: string) {
      return storage.get(key) ?? ''
    },
    setStorageSync(key: string, value: unknown) {
      storage.set(key, value)
    },
    removeStorageSync(key: string) {
      storage.delete(key)
    },
    navigateTo() {},
    showToast() {}
  } as typeof uni
}

function installWxAuthorizeMock(handler: (options: {
  success?: () => void
  fail?: (error: unknown) => void
}) => void) {
  globalThis.wx = {
    requirePrivacyAuthorize: handler
  } as typeof wx
}

async function simulateLoginClick(performLogin: () => Promise<void>) {
  await ensurePrivacyConsentForLogin()
  await performLogin()
}

test('missing local consent + wechat authorize success writes consent and calls uni.login', async () => {
  installUniMock()
  clearPrivacyConsent()
  let authorizeCalled = false
  let loginCalled = false

  installWxAuthorizeMock(({ success }) => {
    authorizeCalled = true
    success?.()
  })

  globalThis.uni = {
    ...globalThis.uni,
    login(options: UniApp.LoginOptions) {
      loginCalled = true
      options.success?.({ code: 'fresh-code' } as UniApp.LoginRes)
    }
  } as typeof uni

  await simulateLoginClick(async () => {
    await new Promise<void>((resolve, reject) => {
      uni.login({
        provider: 'weixin',
        success: () => resolve(),
        fail: reject
      })
    })
  })

  assert.equal(authorizeCalled, true)
  assert.equal(hasPrivacyConsent(), true)
  assert.equal(loginCalled, true)
})

test('existing local consent skips wechat authorize and calls uni.login directly', async () => {
  installUniMock()
  setPrivacyConsent()
  let authorizeCalled = false
  let loginCalled = false

  installWxAuthorizeMock(() => {
    authorizeCalled = true
  })

  globalThis.uni = {
    ...globalThis.uni,
    login(options: UniApp.LoginOptions) {
      loginCalled = true
      options.success?.({ code: 'fresh-code' } as UniApp.LoginRes)
    }
  } as typeof uni

  await simulateLoginClick(async () => {
    await new Promise<void>((resolve, reject) => {
      uni.login({
        provider: 'weixin',
        success: () => resolve(),
        fail: reject
      })
    })
  })

  assert.equal(authorizeCalled, false)
  assert.equal(loginCalled, true)
})

test('wechat authorize rejection does not write consent or call uni.login', async () => {
  installUniMock()
  clearPrivacyConsent()
  let loginCalled = false

  installWxAuthorizeMock(({ fail }) => {
    fail?.({ errMsg: 'requirePrivacyAuthorize:fail auth deny' })
  })

  globalThis.uni = {
    ...globalThis.uni,
    login() {
      loginCalled = true
    }
  } as typeof uni

  await assert.rejects(
    () => simulateLoginClick(async () => {
      loginCalled = true
    }),
    /拒绝微信隐私授权/
  )

  assert.equal(hasPrivacyConsent(), false)
  assert.equal(loginCalled, false)
})

test('one click completes authorize and login without requiring a second click', async () => {
  installUniMock()
  clearPrivacyConsent()
  const steps: string[] = []

  installWxAuthorizeMock(({ success }) => {
    steps.push('authorize')
    success?.()
  })

  globalThis.uni = {
    ...globalThis.uni,
    login(options: UniApp.LoginOptions) {
      steps.push('login')
      options.success?.({ code: 'one-click-code' } as UniApp.LoginRes)
    }
  } as typeof uni

  await simulateLoginClick(async () => {
    await new Promise<void>((resolve, reject) => {
      uni.login({
        provider: 'weixin',
        success: () => resolve(),
        fail: reject
      })
    })
  })

  assert.deepEqual(steps, ['authorize', 'login'])
  assert.equal(hasPrivacyConsent(), true)
})

test('wechat-login page does not depend on promptPrivacyConsentIfNeeded for login', () => {
  const source = readFileSync(join(miniappSrc, 'pages', 'auth', 'wechat-login.vue'), 'utf8')
  assert.match(source, /ensurePrivacyConsentForLogin/)
  assert.doesNotMatch(source, /promptPrivacyConsentIfNeeded/)
  assert.doesNotMatch(source, /请先同意个人信息保护提示/)
})

test('PrivacyConsentModal does not record consent when authorize fails', () => {
  const source = readFileSync(join(miniappSrc, 'components', 'legal', 'PrivacyConsentModal.vue'), 'utf8')
  const agreeBlock = source.slice(source.indexOf('async function agree'), source.indexOf('let unsubscribe'))
  const catchBody = agreeBlock.match(/catch \(error\) \{([\s\S]*?)\n  \}/)?.[1] || ''
  assert.match(catchBody, /return/)
  assert.doesNotMatch(catchBody, /setPrivacyConsent/)
})
