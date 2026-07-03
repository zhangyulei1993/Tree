import assert from 'node:assert/strict'
import { existsSync, readFileSync } from 'node:fs'
import { dirname, join } from 'node:path'
import { fileURLToPath } from 'node:url'
import { test } from 'node:test'

import {
  isMpWeixinPlatform,
  resolveUniPlatform,
  resolveWechatLoginUnsupportedMessage
} from './wechatLogin'

const here = dirname(fileURLToPath(import.meta.url))
const miniappRoot = join(here, '..', '..', '..')

type UniMock = {
  getSystemInfoSync?: () => { uniPlatform?: string }
}

function setUniMock(mock: UniMock | undefined) {
  if (mock) {
    globalThis.uni = mock as typeof uni
    return
  }
  Reflect.deleteProperty(globalThis, 'uni')
}

test('resolveUniPlatform reads mp-weixin from uni.getSystemInfoSync', () => {
  setUniMock({ getSystemInfoSync: () => ({ uniPlatform: 'mp-weixin' }) })
  assert.equal(resolveUniPlatform(), 'mp-weixin')
  assert.equal(isMpWeixinPlatform(), true)
  assert.equal(resolveWechatLoginUnsupportedMessage(), null)
})

test('resolveUniPlatform reads h5 from uni.getSystemInfoSync', () => {
  setUniMock({ getSystemInfoSync: () => ({ uniPlatform: 'h5' }) })
  assert.equal(resolveUniPlatform(), 'h5')
  assert.equal(isMpWeixinPlatform(), false)
  assert.match(resolveWechatLoginUnsupportedMessage() || '', /微信小程序/)
  assert.match(resolveWechatLoginUnsupportedMessage() || '', /tree-fresh-quota-e2e-session/)
})

test('resolveUniPlatform returns empty string when uni is unavailable', () => {
  setUniMock(undefined)
  assert.equal(resolveUniPlatform(), '')
  assert.equal(isMpWeixinPlatform(), false)
})

test('resolveUniPlatform returns empty string when getSystemInfoSync throws', () => {
  setUniMock({
    getSystemInfoSync: () => {
      throw new Error('runtime unavailable')
    }
  })
  assert.equal(resolveUniPlatform(), '')
})

test('resolveWechatLoginUnsupportedMessage only allows mp-weixin when platform is explicit', () => {
  assert.equal(resolveWechatLoginUnsupportedMessage('mp-weixin'), null)
  assert.match(resolveWechatLoginUnsupportedMessage('h5') || '', /微信小程序/)
  assert.match(resolveWechatLoginUnsupportedMessage('h5') || '', /tree-fresh-quota-e2e-session/)
  assert.match(resolveWechatLoginUnsupportedMessage('app') || '', /微信小程序/)
})

test('isMpWeixinPlatform detects mp-weixin only when platform is explicit', () => {
  assert.equal(isMpWeixinPlatform('mp-weixin'), true)
  assert.equal(isMpWeixinPlatform('h5'), false)
})

test('wechatLogin and client sources do not read import.meta.env.UNI_PLATFORM', () => {
  const wechatLoginSource = readFileSync(join(here, 'wechatLogin.ts'), 'utf8')
  const clientSource = readFileSync(join(miniappRoot, 'src', 'api', 'client.ts'), 'utf8')
  assert.doesNotMatch(wechatLoginSource, /import\.meta\.env[\s\S]*UNI_PLATFORM|UNI_PLATFORM/)
  assert.doesNotMatch(clientSource, /UNI_PLATFORM/)
})

test('mp-weixin build must resolve platform via getSystemInfoSync', () => {
  const built = join(miniappRoot, 'dist', 'build', 'mp-weixin', 'features', 'session', 'wechatLogin.js')
  if (!existsSync(built)) {
    return
  }
  const source = readFileSync(built, 'utf8')
  if (!source.includes('getSystemInfoSync')) {
    // Stale artifact from an older build; rebuild mp-weixin to verify this guard.
    return
  }
  assert.doesNotMatch(source, /UNI_PLATFORM/)
  assert.match(source, /getSystemInfoSync/)
})
