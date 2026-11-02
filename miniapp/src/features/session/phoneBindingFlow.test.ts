import assert from 'node:assert/strict'
import { existsSync, readFileSync } from 'node:fs'
import { dirname, join } from 'node:path'
import { fileURLToPath } from 'node:url'
import { test } from 'node:test'

const here = dirname(fileURLToPath(import.meta.url))
const miniappSrc = join(here, '..', '..')

test('WeChat phone authorization supports optional quick login and logged-in account binding', () => {
  const authApi = readFileSync(join(miniappSrc, 'api', 'auth.ts'), 'utf8')
  const session = readFileSync(join(miniappSrc, 'stores', 'session.ts'), 'utf8')
  const router = readFileSync(join(miniappSrc, '..', '..', 'backend', 'internal', 'app', 'router.go'), 'utf8')

  assert.match(authApi, /bindWechatPhone/)
  assert.match(authApi, /\/auth\/wechat-mini\/bind-phone/)
  assert.match(authApi, /wechatMiniPhoneLogin/)
  assert.match(authApi, /\/auth\/wechat-mini\/phone-login/)
  assert.match(authApi, /phoneCode: string; code: string; clientType: string/)
  assert.match(session, /bindWechatPhone/)
  assert.match(session, /loginWithWechatPhone/)
  assert.match(session, /uni\.login\(\{ provider: 'weixin'/)
  assert.match(router, /protected\.POST\("\/wechat-mini\/bind-phone"/)
  assert.match(router, /auth\.POST\("\/wechat-mini\/phone-login"/)
  for (const source of [authApi, session, router]) {
    assert.doesNotMatch(source, /bind-phone-credential/)
    assert.doesNotMatch(source, /change-phone-login-password/)
  }
})

test('profile and security pages describe phone binding without password login', () => {
  const security = readFileSync(join(miniappSrc, 'pages', 'me', 'account-security.vue'), 'utf8')
  assert.match(security, /手机号绑定/)
  assert.match(security, /open-type="getPhoneNumber"/)
  assert.match(security, /bindWithWechatPhone/)
  const profile = readFileSync(join(miniappSrc, 'pages', 'me', 'profile.vue'), 'utf8')
  assert.doesNotMatch(profile, /getPhoneNumber|手机号绑定/)
  for (const source of [profile, security]) {
    assert.doesNotMatch(source, /备用登录/)
    assert.doesNotMatch(source, /登录密码/)
  }
})

test('the retired phone password login page is no longer reachable', () => {
  const pagesJson = readFileSync(join(miniappSrc, 'pages.json'), 'utf8')
  assert.doesNotMatch(pagesJson, /pages\/auth\/phone-login/)
  assert.equal(existsSync(join(miniappSrc, 'pages', 'auth', 'phone-login.vue')), false)
})
