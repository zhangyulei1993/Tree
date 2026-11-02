import assert from 'node:assert/strict'
import test from 'node:test'

function isStatusResult(value) {
  return typeof value === 'object' && value !== null && typeof value.status === 'string'
}

test('change-phone API contract expects status payload', () => {
  assert.equal(isStatusResult({ status: 'ok' }), true)
  assert.equal(isStatusResult({ accessToken: 'token' }), false)
  assert.equal(isStatusResult(null), false)
})
