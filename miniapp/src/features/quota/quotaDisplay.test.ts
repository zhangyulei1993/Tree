import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { dirname, join } from 'node:path'
import { fileURLToPath } from 'node:url'
import { test } from 'node:test'

import { buildCapabilitiesSummary, parseQuotaErrorMessage, quotaReachedMessage } from './quotaDisplay'

const here = dirname(fileURLToPath(import.meta.url))

function quotaApiError(message: string, code: number, data: Record<string, unknown>) {
  const error = new Error(message) as Error & { code: number; data: Record<string, unknown> }
  error.code = code
  error.data = data
  return error
}

test('capabilities api module exposes fetchCapabilities', () => {
  const source = readFileSync(join(here, '..', '..', 'api', 'capabilities.ts'), 'utf8')
  assert.match(source, /fetchCapabilities/)
  assert.match(source, /\/users\/me\/capabilities/)
  assert.match(source, /request</)
})

test('account security page refetches capabilities on each onShow for admin config changes', () => {
  const source = readFileSync(join(here, '..', '..', 'pages', 'me', 'account-security.vue'), 'utf8')
  assert.match(source, /fetchCapabilities/)
  assert.match(source, /onShow/)
  assert.match(source, /limits\.maxOwnedFamilies/)
  assert.doesNotMatch(source, /bind-phone|phone-login|验证手机号/)
})

test('quota reached message does not route to phone pages', () => {
  assert.equal(quotaReachedMessage('WECHAT_ONLY'), '当前扩展能力尚未开放')
  assert.equal(quotaReachedMessage('PHONE_BOUND'), '已达到当前账号权益上限')
})

test('parseQuotaErrorMessage uses trustTier from ApiError data', () => {
  const wechatError = quotaApiError('已达到可创建家庭数量上限', 41502, {
    trustTier: 'WECHAT_ONLY',
    quotaType: 'ownedFamilies',
    limit: 1,
    usage: 1
  })
  assert.equal(parseQuotaErrorMessage(wechatError, 'fallback'), '当前扩展能力尚未开放')

  const phoneError = quotaApiError('已达到可加入家庭数量上限', 41503, {
    trustTier: 'PHONE_BOUND',
    quotaType: 'joinedFamilies',
    limit: 5,
    usage: 5
  })
  assert.equal(parseQuotaErrorMessage(phoneError, 'fallback'), '已达到当前账号权益上限')
})

test('client apiErrorMessage applies trustTier quota messaging', () => {
  const source = readFileSync(join(here, '..', '..', 'api', 'client.ts'), 'utf8')
  assert.match(source, /parseQuotaErrorMessage/)
  assert.match(source, /data: payload\?\.data/)
})

test('buildCapabilitiesSummary formats dynamic admin limits', () => {
  const lines = buildCapabilitiesSummary({
    trustTier: 'WECHAT_ONLY',
    limits: { maxOwnedFamilies: 2, maxMembersPerOwnedFamily: 6, maxJoinedFamilies: 2 },
    usage: { ownedFamilies: 1, joinedFamilies: 0, membersPerOwnedFamily: { '1': 3 } }
  })
  assert.equal(lines[0], '可创建家庭：1/2')
  assert.equal(lines[1], '可加入家庭：0/2')
  assert.equal(lines[2], '每个家庭成员上限：6')
})
