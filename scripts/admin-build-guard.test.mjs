import assert from 'node:assert/strict'
import {
  cpSync,
  existsSync,
  mkdirSync,
  mkdtempSync,
  readFileSync,
  readdirSync,
  writeFileSync
} from 'node:fs'
import { tmpdir } from 'node:os'
import { join, dirname } from 'node:path'
import { spawnSync } from 'node:child_process'
import { fileURLToPath } from 'node:url'
import { test } from 'node:test'

import { ADMIN_BUILD_META_FILE, computeDistDigest } from './admin-dist-digest.mjs'
import { verifyAdminBuildMeta } from './verify-admin-build-meta.mjs'

const scriptsDir = dirname(fileURLToPath(import.meta.url))
const repoRoot = join(scriptsDir, '..')
const adminWebDir = join(repoRoot, 'admin-web')
const adminDistDir = join(adminWebDir, 'dist')

function runAdminBuild(scriptName) {
  const hasPnpm = spawnSync('pnpm', ['--version'], { stdio: 'ignore' }).status === 0
  const runner = hasPnpm ? 'pnpm' : 'npm'
  const result = spawnSync(runner, ['run', scriptName], {
    cwd: adminWebDir,
    stdio: 'pipe',
    encoding: 'utf8'
  })
  assert.equal(
    result.status,
    0,
    `${runner} run ${scriptName} failed:\n${result.stdout}\n${result.stderr}`
  )
}

function readAdminBuildMeta() {
  return JSON.parse(readFileSync(join(adminDistDir, ADMIN_BUILD_META_FILE), 'utf8'))
}

function writeVerifiedDist(distDir, { apiMode = 'real', extraFiles = {} } = {}) {
  mkdirSync(join(distDir, 'assets'), { recursive: true })
  writeFileSync(join(distDir, 'index.html'), '<!doctype html><html><body>admin</body></html>\n')
  writeFileSync(join(distDir, 'assets', 'app.js'), 'console.log("admin")\n')
  for (const [relativePath, content] of Object.entries(extraFiles)) {
    const target = join(distDir, relativePath)
    mkdirSync(dirname(target), { recursive: true })
    writeFileSync(target, content)
  }
  const distDigest = computeDistDigest(distDir)
  writeFileSync(
    join(distDir, ADMIN_BUILD_META_FILE),
    `${JSON.stringify({ apiMode, apiBaseUrl: '/api', distDigest }, null, 2)}\n`
  )
  return distDigest
}

function simulateStagingRelease(adminSourceDir, webSourceDir) {
  const releaseRoot = mkdtempSync(join(tmpdir(), 'tree-staging-release-'))
  const releaseDir = join(releaseRoot, 'releases', 'test-release')
  const adminTarget = join(releaseDir, 'admin-web', 'dist')
  const webTarget = join(releaseDir, 'web', 'dist')
  mkdirSync(adminTarget, { recursive: true })
  mkdirSync(webTarget, { recursive: true })
  cpSync(adminSourceDir, adminTarget, { recursive: true })
  cpSync(webSourceDir, webTarget, { recursive: true })
  return { releaseDir, adminTarget, webTarget }
}

test('verify-admin-build-meta rejects missing meta file', () => {
  const tempDir = mkdtempSync(join(tmpdir(), 'tree-admin-meta-missing-'))
  mkdirSync(join(tempDir, 'assets'))
  writeFileSync(join(tempDir, 'index.html'), '<html></html>\n')
  assert.throws(
    () => verifyAdminBuildMeta(tempDir),
    /Admin build meta missing/
  )
})

test('verify-admin-build-meta rejects mock dist', () => {
  const tempDir = mkdtempSync(join(tmpdir(), 'tree-admin-meta-mock-'))
  writeVerifiedDist(tempDir, { apiMode: 'mock' })
  assert.throws(
    () => verifyAdminBuildMeta(tempDir),
    /apiMode must be "real"/
  )
})

test('verify-admin-build-meta rejects real meta without index.html', () => {
  const tempDir = mkdtempSync(join(tmpdir(), 'tree-admin-meta-no-index-'))
  mkdirSync(join(tempDir, 'assets'))
  writeFileSync(join(tempDir, 'assets', 'app.js'), 'console.log("x")\n')
  const distDigest = computeDistDigest(tempDir)
  writeFileSync(
    join(tempDir, ADMIN_BUILD_META_FILE),
    `${JSON.stringify({ apiMode: 'real', apiBaseUrl: '/api', distDigest }, null, 2)}\n`
  )
  assert.throws(
    () => verifyAdminBuildMeta(tempDir),
    /index\.html missing/
  )
})

test('verify-admin-build-meta rejects real meta without assets directory', () => {
  const tempDir = mkdtempSync(join(tmpdir(), 'tree-admin-meta-no-assets-'))
  writeFileSync(join(tempDir, 'index.html'), '<html></html>\n')
  const distDigest = computeDistDigest(tempDir)
  writeFileSync(
    join(tempDir, ADMIN_BUILD_META_FILE),
    `${JSON.stringify({ apiMode: 'real', apiBaseUrl: '/api', distDigest }, null, 2)}\n`
  )
  assert.throws(
    () => verifyAdminBuildMeta(tempDir),
    /assets directory missing/
  )
})

test('verify-admin-build-meta accepts matching dist digest', () => {
  const tempDir = mkdtempSync(join(tmpdir(), 'tree-admin-meta-real-'))
  const distDigest = writeVerifiedDist(tempDir)
  const meta = verifyAdminBuildMeta(tempDir)
  assert.equal(meta.apiMode, 'real')
  assert.equal(meta.apiBaseUrl, '/api')
  assert.equal(meta.distDigest, distDigest)
})

test('default build produces real admin-build-meta.json with digest', { timeout: 180_000 }, () => {
  runAdminBuild('build')
  const meta = readAdminBuildMeta()
  assert.equal(meta.apiMode, 'real')
  assert.equal(meta.apiBaseUrl, '/api')
  assert.match(meta.distDigest, /^[a-f0-9]{64}$/)
  verifyAdminBuildMeta(adminDistDir)
})

test('build:staging produces real admin-build-meta.json with digest', { timeout: 180_000 }, () => {
  runAdminBuild('build:staging')
  const meta = readAdminBuildMeta()
  assert.equal(meta.apiMode, 'real')
  assert.equal(meta.apiBaseUrl, '/api')
  assert.equal(meta.buildProfile, 'staging')
  assert.match(meta.distDigest, /^[a-f0-9]{64}$/)
  verifyAdminBuildMeta(adminDistDir)
})

test('build:mock produces mock admin-build-meta.json', { timeout: 180_000 }, () => {
  runAdminBuild('build:mock')
  const meta = readAdminBuildMeta()
  assert.equal(meta.apiMode, 'mock')
  assert.equal(meta.apiBaseUrl, '/api')
  assert.throws(
    () => verifyAdminBuildMeta(adminDistDir),
    /apiMode must be "real"/
  )
})

test('verify rejects tampered dist artifact digest', { timeout: 180_000 }, () => {
  runAdminBuild('build:staging')
  verifyAdminBuildMeta(adminDistDir)

  const assetName = readdirSync(join(adminDistDir, 'assets')).find((name) => name.endsWith('.js'))
  assert.ok(assetName, 'expected built admin asset')
  writeFileSync(join(adminDistDir, 'assets', assetName), '/* tampered */\n')

  assert.throws(
    () => verifyAdminBuildMeta(adminDistDir),
    /distDigest mismatch/
  )
})

test('simulated staging release exposes admin/web dist indexes', { timeout: 180_000 }, () => {
  runAdminBuild('build:staging')

  const webSourceDir = mkdtempSync(join(tmpdir(), 'tree-web-dist-'))
  writeFileSync(join(webSourceDir, 'index.html'), '<!doctype html><html><body>web</body></html>\n')

  const { adminTarget, webTarget } = simulateStagingRelease(adminDistDir, webSourceDir)

  assert.equal(existsSync(join(adminTarget, 'index.html')), true)
  assert.equal(existsSync(join(webTarget, 'index.html')), true)
  verifyAdminBuildMeta(adminTarget)
})

test('deploy guard CLI rejects dist without meta file', () => {
  const tempDir = mkdtempSync(join(tmpdir(), 'tree-admin-deploy-missing-'))
  mkdirSync(join(tempDir, 'assets'))
  writeFileSync(join(tempDir, 'index.html'), '<html></html>\n')
  const result = spawnSync('node', [join(scriptsDir, 'verify-admin-build-meta.mjs'), tempDir], {
    encoding: 'utf8'
  })
  assert.notEqual(result.status, 0)
  assert.match(result.stderr, /Admin build meta missing/)
})

test('deploy guard CLI rejects mock dist directory', () => {
  const tempDir = mkdtempSync(join(tmpdir(), 'tree-admin-deploy-guard-'))
  writeVerifiedDist(tempDir, { apiMode: 'mock' })
  const result = spawnSync('node', [join(scriptsDir, 'verify-admin-build-meta.mjs'), tempDir], {
    encoding: 'utf8'
  })
  assert.notEqual(result.status, 0)
  assert.match(result.stderr, /apiMode must be "real"/)
})
