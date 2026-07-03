import assert from 'node:assert/strict'
import { readFileSync, existsSync } from 'node:fs'
import { dirname, join } from 'node:path'
import { fileURLToPath } from 'node:url'
import { test } from 'node:test'

import { isProfileComplete } from '../session/profileComplete'

const here = dirname(fileURLToPath(import.meta.url))
const miniappSrc = join(here, '..', '..')

const removedPhonePages = [
  'pages/auth/bind-phone.vue',
  'pages/auth/phone-login.vue',
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

test('session store exposes wechat profile states and requireProfileComplete', () => {
  const source = readFileSync(join(miniappSrc, 'stores', 'session.ts'), 'utf8')
  assert.match(source, /wechatProfileIncomplete/)
  assert.match(source, /wechatActive/)
  assert.match(source, /requireProfileComplete/)
  assert.doesNotMatch(source, /requirePhoneBound/)
  assert.doesNotMatch(source, /bindPhone/)
  assert.doesNotMatch(source, /loginPhone/)
})

test('removed phone auth pages are absent from pages.json and filesystem', () => {
  const pagesJson = readFileSync(join(miniappSrc, 'pages.json'), 'utf8')
  for (const page of ['bind-phone', 'phone-login', 'register-phone', 'change-phone']) {
    assert.equal(pagesJson.includes(page), false, `pages.json still references ${page}`)
    const full = join(miniappSrc, page.includes('account') ? 'pages/account' : 'pages/auth', `${page.split('/').pop()}.vue`)
    assert.equal(existsSync(full), false, `page file still exists: ${full}`)
  }
})

test('protected pages use requireProfileComplete instead of phone binding', () => {
  for (const rel of protectedPages) {
    const source = readFileSync(join(miniappSrc, rel), 'utf8')
    assert.doesNotMatch(source, /requirePhoneBound|isPhoneBound|bind-phone|phone-login|register-phone/)
    assert.match(source, /requireProfileComplete|isProfileComplete/)
  }
})

test('wechat login routes to profile onboarding when nickname missing', () => {
  const source = readFileSync(join(miniappSrc, 'pages', 'auth', 'wechat-login.vue'), 'utf8')
  assert.match(source, /routeAfterAuth/)
  assert.doesNotMatch(source, /bind-phone|phone-login/)
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

test('profile onboarding saves nickname then continues', () => {
  const source = readFileSync(join(miniappSrc, 'pages', 'me', 'profile.vue'), 'utf8')
  assert.match(source, /onboarding/)
  assert.match(source, /finishProfile/)
  assert.match(source, /请输入昵称/)
})

test('cancel account uses wechat reauth API', () => {
  const authApi = readFileSync(join(miniappSrc, 'api', 'auth.ts'), 'utf8')
  const cancelPage = readFileSync(join(miniappSrc, 'pages', 'account', 'cancel.vue'), 'utf8')
  assert.match(authApi, /cancel-account\/wechat-reauth/)
  assert.doesNotMatch(authApi, /sendCode|loginPhone|bindPhone/)
  assert.match(cancelPage, /uni\.login/)
  assert.doesNotMatch(cancelPage, /sendCode|phoneCode/)
})

test('privacy catalog does not declare phone sms or password collection', () => {
  const catalog = readFileSync(join(miniappSrc, 'features', 'legal', 'privacyCatalog.ts'), 'utf8')
  const legal = readFileSync(join(miniappSrc, 'features', 'legal', 'legalContent.ts'), 'utf8')
  const combined = catalog + legal
  assert.match(combined, /不收集手机号、短信验证码或登录密码|手机号、短信验证码、登录密码/)
  assert.doesNotMatch(combined, /bind-phone|register-phone|phone-login/)
})
