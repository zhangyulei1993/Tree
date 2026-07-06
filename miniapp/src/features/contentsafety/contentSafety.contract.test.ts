import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { dirname, join } from 'node:path'
import { fileURLToPath } from 'node:url'
import { test } from 'node:test'

import { parseQuotaErrorMessage } from '../quota/quotaDisplay'

const here = dirname(fileURLToPath(import.meta.url))

function contentSafetyApiError(message: string, code: number) {
  const error = new Error(message) as Error & { code: number }
  error.code = code
  return error
}

test('content safety rejected error surfaces backend message', () => {
  const error = contentSafetyApiError('内容可能不符合平台规范，请修改后重试', 49007)
  assert.equal(parseQuotaErrorMessage(error, 'fallback'), '内容可能不符合平台规范，请修改后重试')
})

test('backend exposes content safety error codes', () => {
  const source = readFileSync(join(here, '..', '..', '..', '..', 'backend', 'internal', 'common', 'errors', 'codes.go'), 'utf8')
  assert.match(source, /CodeContentSafetyRejected\s+Code\s*=\s*49007/)
  assert.match(source, /内容可能不符合平台规范，请修改后重试/)
})
