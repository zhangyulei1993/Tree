import assert from 'node:assert/strict'
import { existsSync, readFileSync } from 'node:fs'
import { dirname, join } from 'node:path'
import { fileURLToPath } from 'node:url'
import { test } from 'node:test'

const here = dirname(fileURLToPath(import.meta.url))
const miniappSrc = join(here, '..', '..')

test('auth api exposes phone backup login endpoints without sms helpers', () => {
  const source = readFileSync(join(miniappSrc, 'api', 'auth.ts'), 'utf8')
  assert.match(source, /loginPhone/)
  assert.match(source, /\/auth\/login-phone/)
  assert.match(source, /bindPhoneCredential/)
  assert.match(source, /\/auth\/wechat-mini\/bind-phone-credential/)
  assert.match(source, /changePhoneLoginPassword/)
  assert.match(source, /\/auth\/wechat-mini\/change-phone-login-password/)
  assert.doesNotMatch(source, /sendCode|getPhoneNumber|phoneCode/)
})

test('session store wires phone login and credential bind with auth refresh', () => {
  const source = readFileSync(join(miniappSrc, 'stores', 'session.ts'), 'utf8')
  assert.match(source, /loginWithPhone/)
  assert.match(source, /bindPhoneCredential/)
  assert.match(source, /changePhoneLoginPassword/)
  assert.match(source, /refreshAuthContext/)
  assert.match(source, /fetchCapabilities/)
  assert.match(source, /refreshMe/)
})

test('legacy sms auth routes are not registered in backend router', () => {
  const source = readFileSync(join(miniappSrc, '..', '..', 'backend', 'internal', 'app', 'router.go'), 'utf8')
  for (const route of [
    'auth.POST("/send-code"',
    'auth.POST("/register-phone"',
    'auth.POST("/wechat-mini/phone-login"',
    'protected.POST("/wechat-mini/bind-phone"',
    'protected.POST("/change-phone"',
    'protected.POST("/cancel-account"'
  ]) {
    assert.equal(source.includes(route), false, `router still registers ${route}`)
  }
})

test('phone-login page collects phone password and legal consent only', () => {
  const source = readFileSync(join(miniappSrc, 'pages', 'auth', 'phone-login.vue'), 'utf8')
  assert.match(source, /loginWithPhone/)
  assert.match(source, /AuthLegalConsent/)
  assert.match(source, /ensurePrivacyConsentForLogin/)
  assert.match(source, /返回微信登录/)
  assert.doesNotMatch(source, /register|sendCode|getPhoneNumber|验证码/)
})

test('wechat-login exposes secondary link to phone backup login', () => {
  const source = readFileSync(join(miniappSrc, 'pages', 'auth', 'wechat-login.vue'), 'utf8')
  assert.match(source, /phone-login/)
  assert.match(source, /使用手机号密码登录/)
})

test('profile page supports optional phone backup setup and password change', () => {
  const source = readFileSync(join(miniappSrc, 'pages', 'me', 'profile.vue'), 'utf8')
  assert.match(source, /bindPhoneCredential/)
  assert.match(source, /changePhoneLoginPassword/)
  assert.match(source, /maskPhone/)
  assert.doesNotMatch(source, /sendCode|getPhoneNumber|验证码|captcha/)
})

test('pages.json registers phone-login route', () => {
  const pagesJson = readFileSync(join(miniappSrc, 'pages.json'), 'utf8')
  assert.match(pagesJson, /pages\/auth\/phone-login/)
  assert.ok(existsSync(join(miniappSrc, 'pages', 'auth', 'phone-login.vue')))
})

test('user types expose phoneLoginEnabled and PHONE_BOUND trust tier', () => {
  const source = readFileSync(join(miniappSrc, 'types', 'api.ts'), 'utf8')
  assert.match(source, /phoneLoginEnabled/)
  assert.match(source, /PHONE_BOUND/)
  assert.doesNotMatch(source, /PHONE_VERIFIED/)
})

test('maskPhone helper masks standard mobile numbers', async () => {
  const { maskPhone } = await import('../../utils/maskPhone.ts')
  assert.equal(maskPhone('13800000000'), '138****0000')
  assert.equal(maskPhone(''), '')
})
