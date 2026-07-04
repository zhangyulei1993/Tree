import { existsSync, readFileSync } from 'node:fs'
import { join } from 'node:path'
import { fileURLToPath } from 'node:url'

import { ADMIN_BUILD_META_FILE, computeDistDigest } from './admin-dist-digest.mjs'

export function verifyAdminBuildMeta(distDir) {
  if (!distDir || typeof distDir !== 'string') {
    throw new Error('Admin dist directory is required')
  }

  const indexPath = join(distDir, 'index.html')
  if (!existsSync(indexPath)) {
    throw new Error(`Admin dist index.html missing: ${indexPath}`)
  }

  const assetsDir = join(distDir, 'assets')
  if (!existsSync(assetsDir)) {
    throw new Error(`Admin dist assets directory missing: ${assetsDir}`)
  }

  const metaPath = join(distDir, ADMIN_BUILD_META_FILE)
  if (!existsSync(metaPath)) {
    throw new Error(`Admin build meta missing: ${metaPath}`)
  }

  let meta
  try {
    meta = JSON.parse(readFileSync(metaPath, 'utf8'))
  } catch {
    throw new Error(`Admin build meta is not valid JSON: ${metaPath}`)
  }

  if (meta.apiMode !== 'real') {
    throw new Error(`Admin build apiMode must be "real", got "${meta.apiMode ?? ''}"`)
  }
  if (meta.apiBaseUrl !== '/api') {
    throw new Error(`Admin build apiBaseUrl must be "/api", got "${meta.apiBaseUrl ?? ''}"`)
  }
  if (typeof meta.distDigest !== 'string' || !meta.distDigest) {
    throw new Error('Admin build distDigest is missing')
  }

  const actualDigest = computeDistDigest(distDir)
  if (actualDigest !== meta.distDigest) {
    throw new Error('Admin build distDigest mismatch; dist artifacts may be tampered or stale')
  }

  return meta
}

function isCliInvocation() {
  const entry = process.argv[1]
  if (!entry) return false
  return fileURLToPath(import.meta.url) === entry
}

if (isCliInvocation()) {
  try {
    verifyAdminBuildMeta(process.argv[2])
    process.stdout.write(`[PASS] Admin build meta verified for ${process.argv[2]}\n`)
  } catch (error) {
    const message = error instanceof Error ? error.message : String(error)
    process.stderr.write(`[FAIL] ${message}\n`)
    process.exit(1)
  }
}
