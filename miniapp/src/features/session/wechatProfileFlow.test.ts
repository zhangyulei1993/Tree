import assert from 'node:assert/strict'
import { readFileSync, existsSync } from 'node:fs'
import { dirname, join } from 'node:path'
import { fileURLToPath } from 'node:url'
import { test } from 'node:test'

import { isProfileComplete } from '../session/profileComplete'

const here = dirname(fileURLToPath(import.meta.url))
const miniappSrc = join(here, '..', '..')

const removedLegacyPhonePages = [
  'pages/auth/bind-phone.vue',
  'pages/auth/register-phone.vue',
  'pages/account/change-phone.vue'
]

const protectedPages = [
  'pages/home/index.vue',
  'pages/family/my.vue',
  'pages/family/create.vue',
  'pages/invite/detail.vue',
  'pages/join/apply.vue'
]

test('profile completion derives from nickname only', () => {
  assert.equal(isProfileComplete({ id: 1, phoneVerified: false, status: 'ACTIVE', nickname: '  松山  ' }), true)
  assert.equal(isProfileComplete({ id: 1, phoneVerified: false, status: 'ACTIVE', nickname: '' }), false)
  assert.equal(isProfileComplete({ id: 1, phoneVerified: false, status: 'ACTIVE', nickname: null }), false)
  assert.equal(isProfileComplete({ id: 1, phoneVerified: false, status: 'ACTIVE', nickname: '松山', avatarUrl: null }), true)
})

test('session store exposes wechat profile states and phone backup actions', () => {
  const source = readFileSync(join(miniappSrc, 'stores', 'session.ts'), 'utf8')
  assert.match(source, /wechatProfileIncomplete/)
  assert.match(source, /wechatActive/)
  assert.match(source, /requireProfileComplete/)
  assert.match(source, /loginWithPhone/)
  assert.match(source, /bindPhoneCredential/)
  assert.doesNotMatch(source, /requirePhoneBound/)
})

test('legacy sms phone auth pages stay removed while backup phone-login exists', () => {
  const pagesJson = readFileSync(join(miniappSrc, 'pages.json'), 'utf8')
  for (const page of ['bind-phone', 'register-phone', 'change-phone']) {
    assert.equal(pagesJson.includes(page), false, `pages.json still references ${page}`)
  }
  for (const rel of removedLegacyPhonePages) {
    assert.equal(existsSync(join(miniappSrc, rel)), false, `page file still exists: ${rel}`)
  }
  assert.match(pagesJson, /phone-login/)
  assert.ok(existsSync(join(miniappSrc, 'pages', 'auth', 'phone-login.vue')))
})

test('protected pages use requireProfileComplete instead of phone binding gate', () => {
  for (const rel of protectedPages) {
    const source = readFileSync(join(miniappSrc, rel), 'utf8')
    assert.doesNotMatch(source, /requirePhoneBound|isPhoneBound|bind-phone|register-phone/)
    assert.match(source, /requireProfileComplete|isProfileComplete/)
  }
})

test('wechat login routes to profile onboarding and links phone backup login', () => {
  const source = readFileSync(join(miniappSrc, 'pages', 'auth', 'wechat-login.vue'), 'utf8')
  assert.match(source, /routeAfterAuth/)
  assert.match(source, /phone-login/)
})

test('profile.vue wires session store before first session usage', () => {
  const source = readFileSync(join(miniappSrc, 'pages', 'me', 'profile.vue'), 'utf8')
  assert.match(source, /import\s+\{\s*useSessionStore\s*\}\s+from\s+'@\/stores\/session'/)
  assert.match(source, /const\s+session\s*=\s*useSessionStore\(\)/)
  const sessionInit = source.indexOf('const session = useSessionStore()')
  const firstSessionUse = source.search(/session\.(user|isLoggedIn|isProfileComplete)/)
  assert.ok(sessionInit >= 0, 'profile.vue must initialize session store')
  assert.ok(firstSessionUse >= 0, 'profile.vue must use session store')
  assert.ok(sessionInit < firstSessionUse, 'session store must be created before first use')
})

test('profile onboarding saves nickname then enters core flow immediately', () => {
  const source = readFileSync(join(miniappSrc, 'pages', 'me', 'profile.vue'), 'utf8')
  assert.match(source, /onboarding/)
  assert.match(source, /finishProfile/)
  assert.match(source, /请输入昵称/)
  assert.match(source, /profileComplete\.value && !onboarding\.value/)
  assert.doesNotMatch(source, /onboarding\.value && !phoneLoginEnabled/)
})

test('cancel account uses wechat reauth API without sms phone auth', () => {
  const authApi = readFileSync(join(miniappSrc, 'api', 'auth.ts'), 'utf8')
  const cancelPage = readFileSync(join(miniappSrc, 'pages', 'account', 'cancel.vue'), 'utf8')
  assert.match(authApi, /cancel-account\/wechat-reauth/)
  assert.match(authApi, /loginPhone/)
  assert.doesNotMatch(authApi, /sendCode|bindPhone\(/)
  assert.match(cancelPage, /uni\.login/)
  assert.doesNotMatch(cancelPage, /sendCode|phoneCode/)
})

test('privacy catalog keeps sms out of scope while backup login uses password only', () => {
  const catalog = readFileSync(join(miniappSrc, 'features', 'legal', 'privacyCatalog.ts'), 'utf8')
  const legal = readFileSync(join(miniappSrc, 'features', 'legal', 'legalContent.ts'), 'utf8')
  const combined = catalog + legal
  assert.match(combined, /短信验证码/)
  assert.doesNotMatch(combined, /bind-phone|register-phone/)
  assert.match(combined, /getPhoneNumber/)
})
