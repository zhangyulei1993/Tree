import assert from 'node:assert/strict'
import { execFileSync, spawn } from 'node:child_process'
import fs from 'node:fs'
import { dirname, join } from 'node:path'
import { fileURLToPath } from 'node:url'

import {
  adminLogin,
  apiRequest,
  assertLoopbackApiBase,
  createFreshQuotaSession,
  fetchAdminQuotaConfigs,
  parseRunId,
  resolveApiBase,
  resolveApiPort,
  sessionFilePath,
  waitForApi,
  writeSessionFile
} from './tree-quota-e2e-lib.mjs'

const here = dirname(fileURLToPath(import.meta.url))
const serveBin = join(here, 'node_modules', '.bin', 'serve')
const repoRoot = join(here, '..')
const backendDir = join(repoRoot, 'backend')
const miniappDir = join(repoRoot, 'miniapp')

export const API_BINARY = '/tmp/tree-api-e2e'
export const API_PORT = Number(process.env.TREE_E2E_API_PORT || 18080)
export const H5_PORT = Number(process.env.TREE_H5_PREVIEW_PORT || 5199)
const TEMP_ADMIN_PASSWORD = 'E2eQuota@Test1'
const TEMP_ADMIN_BCRYPT = '$2a$10$E3zh1hZWlSGFCEERlncyBuF9APuMiFdkNhVWy6IGvvrG37Z6xkeZe'
const SESSION_FILE_PREFIX = 'tree-quota-e2e-session-'

function dockerMysql(sql) {
  return execFileSync('docker', [
    'exec', 'tree-mysql', 'mysql', '-utree_user', '-ptree_pass', 'tree_platform', '-N', '-B', '-e', sql
  ], { encoding: 'utf8' }).trim()
}

function listListeningPids(port) {
  try {
    const out = execFileSync('lsof', ['-nP', `-iTCP:${port}`, '-sTCP:LISTEN', '-t'], { encoding: 'utf8' }).trim()
    return out ? out.split('\n').filter(Boolean) : []
  } catch {
    return []
  }
}

export function assertPortFree(port, label = `port ${port}`) {
  const pids = listListeningPids(port)
  assert.equal(pids.length, 0, `${label} must be free, found listeners: ${pids.join(',')}`)
}

async function waitForPortFree(port, timeoutMs = 15000) {
  const deadline = Date.now() + timeoutMs
  while (Date.now() < deadline) {
    if (listListeningPids(port).length === 0) return
    await new Promise((resolve) => setTimeout(resolve, 200))
  }
  throw new Error(`timeout waiting for port ${port} to become free`)
}

async function stopChild(child, name) {
  if (!child || child.exitCode !== null) return
  child.kill('SIGTERM')
  let exited = await new Promise((resolve) => {
    const forceTimer = setTimeout(() => resolve(false), 5000)
    child.once('exit', () => {
      clearTimeout(forceTimer)
      resolve(true)
    })
  })
  if (!exited && child.exitCode === null) {
    child.kill('SIGKILL')
    exited = await new Promise((resolve) => {
      const forceTimer = setTimeout(() => resolve(false), 3000)
      child.once('exit', () => {
        clearTimeout(forceTimer)
        resolve(true)
      })
    })
  }
  if (!exited) {
    throw new Error(`${name} did not exit during cleanup`)
  }
}

async function killPortListeners(port) {
  for (const pid of listListeningPids(port)) {
    try {
      process.kill(Number(pid), 'SIGKILL')
    } catch {}
  }
  await waitForPortFree(port)
}

function listSessionFiles() {
  return fs.readdirSync('/tmp').filter((name) => name.startsWith(SESSION_FILE_PREFIX) && name.endsWith('.json'))
}

function countActiveWechatE2EUsers() {
  const out = dockerMysql(
    "SELECT COUNT(*) FROM users WHERE status='ACTIVE' AND phone_verified=0 AND nickname LIKE 'H5测试%' AND deleted_at IS NULL"
  )
  return Number(out)
}

function adminPasswordHash(username = 'admin') {
  return dockerMysql(`SELECT password_hash FROM admin_users WHERE username='${username}' AND deleted_at IS NULL LIMIT 1`)
}

function buildTreeApiBinary() {
  execFileSync('go', ['build', '-o', API_BINARY, './cmd/server/main.go'], {
    cwd: backendDir,
    stdio: 'pipe',
    env: {
      ...process.env,
      MYSQL_HOST: process.env.MYSQL_HOST || '127.0.0.1',
      REDIS_HOST: process.env.REDIS_HOST || '127.0.0.1'
    }
  })
  assert.ok(fs.existsSync(API_BINARY), `missing api binary ${API_BINARY}`)
}

function createTempAdminUsername() {
  return `e2e_quota_root_${Date.now()}`
}

function insertTempAdmin(username) {
  dockerMysql(
    `INSERT INTO admin_users (username, password_hash, display_name, role, status, remark, created_at, updated_at) VALUES ` +
    `('${username}', '${TEMP_ADMIN_BCRYPT}', 'E2E Quota Root', 'ROOT_ADMIN', 'ACTIVE', 'fresh-quota-e2e temp', NOW(), NOW())`
  )
}

function deleteTempAdmin(username) {
  dockerMysql(`DELETE FROM admin_users WHERE username='${username}'`)
}

function cleanupStaleTempAdmins() {
  dockerMysql(
    "DELETE FROM admin_users WHERE username LIKE 'e2e_quota_root_%' AND remark='fresh-quota-e2e temp'"
  )
  const remaining = Number(dockerMysql(
    "SELECT COUNT(*) FROM admin_users WHERE username LIKE 'e2e_quota_root_%' AND remark='fresh-quota-e2e temp'"
  ))
  assert.equal(remaining, 0, 'stale fresh-quota-e2e temp admins must be cleaned up before setup')
}

export async function cancelSessionUser(session) {
  assert.ok(session?.token, 'session token required for cancel')
  assert.ok(session?.wxCode, 'session wxCode required for cancel')
  await apiRequest(session.apiBase, '/auth/cancel-account/wechat-reauth', {
    method: 'POST',
    token: session.token,
    body: { code: session.wxCode, cancelReason: 'fresh-quota-e2e cleanup' }
  })
  const status = dockerMysql(`SELECT status FROM users WHERE id=${session.user.id} AND deleted_at IS NULL LIMIT 1`)
  assert.equal(status, 'CANCELLED', `user ${session.user.id} must be CANCELLED before session file removal`)
}

export async function teardownSessionArtifact(session, outFile) {
  await cancelSessionUser(session)
  if (outFile && fs.existsSync(outFile)) {
    fs.rmSync(outFile, { force: true })
  }
  assert.equal(listSessionFiles().length, 0, 'session files must be removed after cancel')
}

export function createE2EHarness() {
  const state = {
    apiProcess: null,
    h5Process: null,
    tempAdminUsername: '',
    adminUsername: '',
    baselineActiveWechatUsers: 0,
    baselineAdminPasswordHash: '',
    apiBase: resolveApiBase()
  }

  return {
    async setup() {
      assertLoopbackApiBase(state.apiBase)
      assertPortFree(API_PORT, 'tree-api')
      assertPortFree(H5_PORT, 'h5-preview')

      state.baselineActiveWechatUsers = countActiveWechatE2EUsers()
      state.baselineAdminPasswordHash = adminPasswordHash('admin')
      cleanupStaleTempAdmins()

      buildTreeApiBinary()
      state.apiProcess = spawn(API_BINARY, [], {
        env: {
          ...process.env,
          MYSQL_HOST: process.env.MYSQL_HOST || '127.0.0.1',
          REDIS_HOST: process.env.REDIS_HOST || '127.0.0.1',
          WECHAT_MOCK_ENABLED: 'true',
          APP_PORT: String(API_PORT)
        },
        stdio: 'ignore'
      })

      await waitForApi(state.apiBase, { timeoutMs: 45000 })

      if (process.env.TREE_E2E_ADMIN_USERNAME && process.env.TREE_E2E_ADMIN_PASSWORD) {
        state.adminUsername = process.env.TREE_E2E_ADMIN_USERNAME
      } else {
        state.tempAdminUsername = createTempAdminUsername()
        insertTempAdmin(state.tempAdminUsername)
        state.adminUsername = state.tempAdminUsername
      }
    },

    async loginAdmin() {
      const password = process.env.TREE_E2E_ADMIN_PASSWORD || TEMP_ADMIN_PASSWORD
      return adminLogin(state.apiBase, { username: state.adminUsername, password })
    },

    async fetchQuotaConfigs() {
      const admin = await this.loginAdmin()
      return fetchAdminQuotaConfigs(state.apiBase, admin.accessToken)
    },

    get apiBase() {
      return state.apiBase
    },

    async forceBuildH5() {
      const distDir = join(miniappDir, 'dist', 'build', 'h5')
      fs.rmSync(distDir, { recursive: true, force: true })
      await new Promise((resolve, reject) => {
        const build = spawn('npm', ['run', 'build:h5:real'], {
          cwd: miniappDir,
          stdio: 'inherit',
          shell: false
        })
        build.on('exit', (code) => (code === 0 ? resolve() : reject(new Error(`build:h5:real failed: ${code}`))))
      })
      assert.ok(fs.existsSync(join(distDir, 'index.html')), 'h5 build output missing')
    },

    async startH5Preview() {
      assertPortFree(H5_PORT, 'h5-preview')
      const distDir = join(miniappDir, 'dist', 'build', 'h5')
      assert.ok(fs.existsSync(serveBin), `serve binary missing at ${serveBin}; run npm install in scripts/`)
      state.h5Process = spawn(serveBin, [distDir, '-l', String(H5_PORT), '--no-port-switching'], {
        cwd: here,
        stdio: 'ignore',
        shell: false
      })
      const h5Base = `http://127.0.0.1:${H5_PORT}`
      const deadline = Date.now() + 45000
      while (Date.now() < deadline) {
        try {
          const response = await fetch(`${h5Base}/`)
          if (response.ok) return h5Base
        } catch {}
        await new Promise((resolve) => setTimeout(resolve, 300))
      }
      throw new Error(`H5 preview not ready at ${h5Base}`)
    },

    async createSession(runIdPrefix) {
      const runId = parseRunId(['--run-id', `${runIdPrefix}-${Date.now()}`])
      const session = await createFreshQuotaSession({ apiBase: state.apiBase, runId })
      const outFile = writeSessionFile(session)
      return { session, outFile }
    },

    async teardown() {
      await stopChild(state.h5Process, 'h5-preview')
      state.h5Process = null
      await killPortListeners(H5_PORT)

      await stopChild(state.apiProcess, 'tree-api')
      state.apiProcess = null
      await killPortListeners(API_PORT)

      if (state.tempAdminUsername) {
        deleteTempAdmin(state.tempAdminUsername)
        state.tempAdminUsername = ''
      }

      assert.equal(countActiveWechatE2EUsers(), state.baselineActiveWechatUsers, 'ACTIVE wechat e2e users must not increase')
      assert.equal(adminPasswordHash('admin'), state.baselineAdminPasswordHash, 'admin password_hash must remain unchanged')
      assertPortFree(API_PORT, 'tree-api after teardown')
      assertPortFree(H5_PORT, 'h5-preview after teardown')
      assert.equal(listSessionFiles().length, 0, '/tmp must contain no session files after teardown')

      if (fs.existsSync(API_BINARY)) {
        fs.rmSync(API_BINARY, { force: true })
      }
      assert.ok(!fs.existsSync(API_BINARY), '/tmp/tree-api-e2e must be removed after teardown')
    }
  }
}

export function getApiOrigin() {
  return `http://127.0.0.1:${resolveApiPort()}`
}
