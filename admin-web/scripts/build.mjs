import { spawnSync } from 'node:child_process'
import { mkdirSync, writeFileSync } from 'node:fs'
import { dirname, join } from 'node:path'
import { fileURLToPath } from 'node:url'

import { ADMIN_BUILD_META_FILE, computeDistDigest } from '../../scripts/admin-dist-digest.mjs'

const adminWebDir = join(dirname(fileURLToPath(import.meta.url)), '..')
const distDir = join(adminWebDir, 'dist')

const profiles = {
  real: {
    buildProfile: 'real',
    apiMode: 'real',
    apiBaseUrl: '/api',
    env: {
      VITE_API_MODE: 'real',
      VITE_API_BASE_URL: '/api'
    }
  },
  staging: {
    buildProfile: 'staging',
    apiMode: 'real',
    apiBaseUrl: '/api',
    env: {
      VITE_API_MODE: 'real',
      VITE_API_BASE_URL: '/api'
    }
  },
  mock: {
    buildProfile: 'mock',
    apiMode: 'mock',
    apiBaseUrl: '/api',
    env: {
      VITE_API_MODE: 'mock',
      VITE_API_BASE_URL: '/api'
    }
  }
}

function die(message) {
  console.error(`[admin-web build] ${message}`)
  process.exit(1)
}

function packageRunner() {
  const hasPnpm = spawnSync('pnpm', ['--version'], { stdio: 'ignore' }).status === 0
  return hasPnpm ? 'pnpm' : 'npm'
}

function runStep(command, args, env) {
  const result = spawnSync(command, args, {
    cwd: adminWebDir,
    env,
    stdio: 'inherit'
  })
  if (result.error) {
    die(`${command} ${args.join(' ')} failed: ${result.error.message}`)
  }
  if (result.status !== 0) {
    process.exit(result.status ?? 1)
  }
}

function runPackageExec(binary, extraArgs, env) {
  const runner = packageRunner()
  const args = runner === 'pnpm'
    ? ['exec', binary, ...extraArgs]
    : ['exec', '--', binary, ...extraArgs]
  runStep(runner, args, env)
}

function writeBuildMeta(profile) {
  mkdirSync(distDir, { recursive: true })
  const distDigest = computeDistDigest(distDir)
  const meta = {
    apiMode: profile.apiMode,
    apiBaseUrl: profile.apiBaseUrl,
    buildProfile: profile.buildProfile,
    builtAt: new Date().toISOString(),
    distDigest
  }
  writeFileSync(join(distDir, ADMIN_BUILD_META_FILE), `${JSON.stringify(meta, null, 2)}\n`, 'utf8')
  console.log(
    `[admin-web build] wrote ${ADMIN_BUILD_META_FILE}: apiMode=${meta.apiMode}, apiBaseUrl=${meta.apiBaseUrl}, distDigest=${distDigest.slice(0, 12)}...`
  )
}

const profileName = process.argv[2] || 'real'
const profile = profiles[profileName]
if (!profile) {
  die(`Unknown build profile "${profileName}". Use real, staging, or mock.`)
}

const buildEnv = {
  ...process.env,
  ...profile.env
}

console.log(`[admin-web build] profile=${profile.buildProfile} apiMode=${profile.apiMode} apiBaseUrl=${profile.apiBaseUrl}`)

runPackageExec('vue-tsc', ['--noEmit'], buildEnv)
runPackageExec('vite', ['build'], buildEnv)
writeBuildMeta(profile)
