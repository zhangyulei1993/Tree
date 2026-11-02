import { execSync } from 'node:child_process'
import { readFileSync } from 'node:fs'
import { join } from 'node:path'

function command(commandLine, cwd) {
  try {
    return execSync(commandLine, {
      cwd,
      encoding: 'utf8',
      stdio: ['ignore', 'pipe', 'ignore']
    }).trim()
  } catch {
    return ''
  }
}

function packageVersion(rootDir) {
  try {
    const pkg = JSON.parse(readFileSync(join(rootDir, 'package.json'), 'utf8'))
    return typeof pkg.version === 'string' ? pkg.version : '0.0.0'
  } catch {
    return '0.0.0'
  }
}

function resolveCommit(rootDir) {
  return (
    process.env.TREE_GIT_COMMIT
    || process.env.VITE_GIT_COMMIT
    || command('git rev-parse --short=12 HEAD', rootDir)
    || 'unknown'
  )
}

function resolveDirty(rootDir) {
  const status = command('git status --short', rootDir)
  return Boolean(status)
}

export function createBuildInfo({ app, rootDir, mode, apiBaseUrl = '' }) {
  return {
    app,
    version: process.env.TREE_APP_VERSION || process.env.VITE_APP_VERSION || packageVersion(rootDir),
    commit: resolveCommit(rootDir),
    buildTime: process.env.TREE_BUILD_TIME || process.env.VITE_BUILD_TIME || new Date().toISOString(),
    environment: process.env.TREE_BUILD_ENV || process.env.VITE_BUILD_ENV || mode || 'development',
    apiBaseUrl,
    dirty: resolveDirty(rootDir)
  }
}
