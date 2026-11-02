import test from 'node:test'
import assert from 'node:assert/strict'

import { readQueryParam, safeDecodeURIComponent } from './queryString'

test('readQueryParam reads encoded values by supported key order', () => {
  assert.equal(
    readQueryParam('source=qr&inviteToken=abc%20123&token=fallback', ['inviteToken', 'token']),
    'abc 123'
  )
})

test('readQueryParam prefers earlier keys even when they appear later in the query', () => {
  assert.equal(readQueryParam('token=old&inviteToken=new', ['inviteToken', 'token']), 'new')
})

test('readQueryParam supports query strings with a leading question mark', () => {
  assert.equal(readQueryParam('?familyId=42', ['familyId']), '42')
})

test('readQueryParam ignores unsupported keys', () => {
  assert.equal(readQueryParam('scene=abc123', ['inviteToken', 'token']), '')
})

test('safeDecodeURIComponent keeps malformed values readable instead of throwing', () => {
  assert.equal(safeDecodeURIComponent('%E0%A4%A'), '%E0%A4%A')
})
