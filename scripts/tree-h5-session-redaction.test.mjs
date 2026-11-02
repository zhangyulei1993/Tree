import assert from 'node:assert/strict'
import { mkdtempSync, readFileSync, rmSync, statSync, writeFileSync } from 'node:fs'
import os from 'node:os'
import path from 'node:path'
import { spawnSync } from 'node:child_process'
import test from 'node:test'

import {
  assertOutputHasNoBearerToken,
  assertOutputHasNoRawJwt,
  formatPlaywrightSessionExample,
  redactToken,
  sanitizeLogValue
} from './lib/h5-session-redaction.mjs'

const repoRoot = path.resolve(import.meta.dirname, '..')
const setupScript = path.join(repoRoot, 'scripts', 'tree-h5-real-e2e-setup.mjs')

function runPrepareSession(tempDir, sessionFile) {
  const credentialsFile = path.join(tempDir, 'credentials.json')
  writeFileSync(credentialsFile, `${JSON.stringify({ password: 'dry-run-only' }, null, 2)}\n`, 'utf8')

  return spawnSync(
    process.execPath,
    [setupScript, 'prepare-h5-session', 'founder'],
    {
      cwd: repoRoot,
      env: {
        ...process.env,
        TREE_E2E_DRY_RUN_SESSION: '1',
        TREE_E2E_CREDENTIALS_FILE: credentialsFile,
        TREE_E2E_SESSION_FILE: sessionFile
      },
      encoding: 'utf8'
    }
  )
}

function assertSessionFileMode(sessionFile) {
  const sessionStat = statSync(sessionFile)
  assert.equal(sessionStat.mode & 0o777, 0o600)
}

test('redactToken never returns full JWT body', () => {
  const token = 'eyJhbGciOiJIUzI1NiJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c'
  const masked = redactToken(token)
  assert.equal(masked.includes('.'), false)
  assert.match(masked, /^eyJhbGci…\(\d+\)$/)
})

test('sanitizeLogValue redacts nested session payload fields', () => {
  const token = 'eyJhbGciOiJIUzI1NiJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c'
  const sanitized = sanitizeLogValue({
    outputFile: '/tmp/session.json',
    localStorage: {
      tree_miniapp_user_token: token,
      tree_miniapp_user: JSON.stringify({ id: 1, nickname: 'demo' })
    },
    Authorization: `Bearer ${token}`
  })
  assert.match(sanitized.localStorage.tree_miniapp_user_token, /^eyJhbGci…\(\d+\)$/)
  assert.equal(sanitized.localStorage.tree_miniapp_user, '[redacted-user-json]')
  assert.match(sanitized.Authorization, /^Bearer eyJhbGci…\(\d+\)$/)
})

test('formatPlaywrightSessionExample stays generic', () => {
  const lines = formatPlaywrightSessionExample().join('\n')
  assert.match(lines, /SESSION_INJECT_PAYLOAD/)
  assert.equal(lines.includes('eyJ'), false)
})

test('prepare-h5-session creates new session file with mode 0600', () => {
  const tempDir = mkdtempSync(path.join(os.tmpdir(), 'tree-h5-session-new-'))
  const sessionFile = path.join(tempDir, 'session-inject.json')

  const result = runPrepareSession(tempDir, sessionFile)
  const combined = `${result.stdout}\n${result.stderr}`
  assert.equal(result.status, 0, combined)
  assertOutputHasNoRawJwt(combined)
  assertOutputHasNoBearerToken(combined)
  assertSessionFileMode(sessionFile)

  const sessionPayload = JSON.parse(readFileSync(sessionFile, 'utf8'))
  assert.ok(sessionPayload.localStorage.tree_miniapp_user_token)
  assert.ok(sessionPayload.localStorage.tree_miniapp_user)

  rmSync(tempDir, { recursive: true, force: true })
})

test('prepare-h5-session tightens pre-existing 0644 session file to 0600', () => {
  const tempDir = mkdtempSync(path.join(os.tmpdir(), 'tree-h5-session-existing-'))
  const sessionFile = path.join(tempDir, 'session-inject.json')
  writeFileSync(sessionFile, '{"stale":true}\n', { encoding: 'utf8', mode: 0o644 })
  assert.equal(statSync(sessionFile).mode & 0o777, 0o644)

  const result = runPrepareSession(tempDir, sessionFile)
  const combined = `${result.stdout}\n${result.stderr}`
  assert.equal(result.status, 0, combined)
  assertOutputHasNoRawJwt(combined)
  assertOutputHasNoBearerToken(combined)
  assertSessionFileMode(sessionFile)

  rmSync(tempDir, { recursive: true, force: true })
})

test('prepare-h5-session stdout/stderr contain no raw JWT', () => {
  const tempDir = mkdtempSync(path.join(os.tmpdir(), 'tree-h5-session-test-'))
  const sessionFile = path.join(tempDir, 'session-inject.json')

  const result = runPrepareSession(tempDir, sessionFile)
  const combined = `${result.stdout}\n${result.stderr}`
  assert.equal(result.status, 0, combined)
  assertOutputHasNoRawJwt(combined)
  assertOutputHasNoBearerToken(combined)
  assert.match(combined, /session-inject\.json|outputFile/)
  assert.match(combined, /tokenPreview|…\(/)
  assert.match(combined, /SESSION_INJECT_PAYLOAD/)
  assert.equal(combined.includes('dry-run-only'), false)
  assertSessionFileMode(sessionFile)

  rmSync(tempDir, { recursive: true, force: true })
})
