import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { dirname, join } from 'node:path'
import { fileURLToPath } from 'node:url'
import { test } from 'node:test'

import { parseQuotaErrorMessage } from '../quota/quotaDisplay'
import { optionalText, validateDateInput, validateTextFields } from '../../utils/inputValidation'

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

test('miniapp rejects unsafe input characters before submit', () => {
  assert.match(
    validateTextFields([
      { value: '张<script>', label: '成员姓名', kind: 'name', required: true }
    ]),
    /成员姓名/
  )
  assert.match(
    validateTextFields([
      { value: '请审核 {debug}', label: '申请说明', kind: 'multiLine' }
    ]),
    /申请说明/
  )
  assert.match(validateDateInput('1990-01-01<script>', '出生日期'), /出生日期/)
})

test('miniapp allows common genealogy text punctuation', () => {
  assert.equal(
    validateTextFields([
      { value: '张·明远', label: '成员姓名', kind: 'name', required: true },
      { value: '祖籍江南，迁居海上；今录入旧谱。', label: '家庭简介', kind: 'multiLine' }
    ]),
    ''
  )
  assert.equal(optionalText('  张氏家族  '), '张氏家族')
  assert.equal(optionalText('   '), undefined)
})
