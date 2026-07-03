import assert from 'node:assert/strict'
import { test } from 'node:test'

import {
  PRIVACY_CONSENT_STORAGE_KEY,
  clearPrivacyConsent,
  hasPrivacyConsent,
  promptPrivacyConsentIfNeeded,
  setPrivacyConsent,
  subscribePrivacyConsentPrompt
} from './privacyConsent'

const storage = new Map<string, unknown>()

;(globalThis as { uni?: Record<string, unknown> }).uni = {
  getStorageSync(key: string) {
    return storage.get(key) ?? ''
  },
  setStorageSync(key: string, value: unknown) {
    storage.set(key, value)
  },
  removeStorageSync(key: string) {
    storage.delete(key)
  },
  navigateTo() {}
}

test('privacy consent storage roundtrip', () => {
  clearPrivacyConsent()
  assert.equal(hasPrivacyConsent(), false)
  setPrivacyConsent()
  assert.equal(hasPrivacyConsent(), true)
  assert.equal(storage.get(PRIVACY_CONSENT_STORAGE_KEY), 'agreed')
  clearPrivacyConsent()
  assert.equal(hasPrivacyConsent(), false)
})

test('promptPrivacyConsentIfNeeded notifies subscribers only when consent missing', () => {
  clearPrivacyConsent()
  let promptCount = 0
  const unsubscribe = subscribePrivacyConsentPrompt(() => {
    promptCount += 1
  })

  assert.equal(promptPrivacyConsentIfNeeded(), true)
  assert.equal(promptCount, 1)

  setPrivacyConsent()
  assert.equal(promptPrivacyConsentIfNeeded(), false)
  assert.equal(promptCount, 1)

  unsubscribe()
  clearPrivacyConsent()
  assert.equal(promptPrivacyConsentIfNeeded(), true)
  assert.equal(promptCount, 1)
})
