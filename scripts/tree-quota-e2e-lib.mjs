import fs from 'node:fs'

const DEFAULT_API_PORT = process.env.TREE_E2E_API_PORT || '18080'

export function assertLoopbackApiBase(apiBase) {
  const url = new URL(apiBase.endsWith('/') ? apiBase.slice(0, -1) : apiBase)
  if (url.hostname !== '127.0.0.1' && url.hostname !== 'localhost') {
    throw new Error(`E2E API must use loopback host, got ${url.hostname}`)
  }
}

export function resolveApiBase() {
  const apiBase = process.env.TREE_LOCAL_API_BASE || `http://127.0.0.1:${DEFAULT_API_PORT}/api`
  assertLoopbackApiBase(apiBase)
  return apiBase
}

export function resolveApiPort() {
  const base = resolveApiBase()
  const matched = base.match(/:(\d+)\//)
  return matched ? matched[1] : DEFAULT_API_PORT
}

export function redactToken(token) {
  if (!token) return ''
  return `${token.slice(0, 8)}…(${token.length})`
}

export async function apiRequest(apiBase, pathname, { method = 'GET', token, body } = {}) {
  assertLoopbackApiBase(apiBase)
  const headers = { Accept: 'application/json' }
  if (body !== undefined) headers['Content-Type'] = 'application/json'
  if (token) headers.Authorization = `Bearer ${token}`

  const response = await fetch(`${apiBase}${pathname}`, {
    method,
    headers,
    body: body === undefined ? undefined : JSON.stringify(body)
  })
  const payload = await response.json().catch(() => ({}))
  if (!response.ok || payload.code !== 0) {
    const err = new Error(payload.message || `HTTP ${response.status}`)
    err.code = payload.code
    err.httpStatus = response.status
    err.payload = payload
    throw err
  }
  return payload.data
}

export async function waitForApi(apiBase, { timeoutMs = 30000, intervalMs = 500 } = {}) {
  const deadline = Date.now() + timeoutMs
  let lastError = null
  while (Date.now() < deadline) {
    try {
      const response = await fetch(`${apiBase}/health`)
      if (response.ok) return
      lastError = new Error(`health HTTP ${response.status}`)
    } catch (error) {
      lastError = error
    }
    await new Promise((resolve) => setTimeout(resolve, intervalMs))
  }
  throw new Error(`tree-api not ready at ${apiBase}: ${lastError?.message || 'timeout'}`)
}

export async function adminLogin(apiBase, { username, password }) {
  if (!username || !password) {
    throw new Error('admin credentials required: set TREE_E2E_ADMIN_USERNAME/PASSWORD or use harness temp admin')
  }
  const data = await apiRequest(apiBase, '/admin/auth/login', {
    method: 'POST',
    body: { username, password }
  })
  if (!data?.accessToken) {
    throw new Error('admin login missing accessToken')
  }
  return data
}

export async function fetchAdminQuotaConfigs(apiBase, adminToken) {
  const rows = await apiRequest(apiBase, '/admin/account-quota-configs', { token: adminToken })
  const map = new Map()
  for (const row of rows) {
    map.set(row.trustTier, row)
  }
  return map
}

export function buildCapabilityLines(capabilities) {
  const { limits, usage } = capabilities
  return [
    `可创建家庭：${usage.ownedFamilies}/${limits.maxOwnedFamilies}`,
    `可加入家庭：${usage.joinedFamilies}/${limits.maxJoinedFamilies}`,
    `每个家庭成员上限：${limits.maxMembersPerOwnedFamily}`,
    limits.supportsFeaturePreview
      ? '特色功能：已开放'
      : '特色功能：当前账号暂未开放'
  ]
}

export function parseRunId(argv) {
  const idx = argv.indexOf('--run-id')
  if (idx >= 0 && argv[idx + 1]) return argv[idx + 1]
  const crypto = globalThis.crypto
  const suffix = crypto?.randomUUID ? crypto.randomUUID().slice(0, 8) : `${Date.now()}`
  return `fq-h5-${Date.now()}-${suffix}`
}

export async function createFreshQuotaSession({
  apiBase = resolveApiBase(),
  runId = parseRunId([]),
  clientType = 'WECHAT_MINI_PROGRAM'
} = {}) {
  const wxCode = `fresh-quota-${runId}`
  const login = await apiRequest(apiBase, '/auth/wechat-mini/login', {
    method: 'POST',
    body: { code: wxCode, clientType }
  })
  if (!login?.accessToken || !login?.user?.id) {
    throw new Error('wechat login missing accessToken or user.id')
  }

  const nickname = `H5测试${runId.slice(-6)}`
  const profile = await apiRequest(apiBase, '/users/me/profile', {
    method: 'PATCH',
    token: login.accessToken,
    body: { nickname }
  })
  if (!profile?.id) {
    throw new Error('profile update missing user id')
  }

  const capabilities = await apiRequest(apiBase, '/users/me/capabilities', {
    token: login.accessToken
  })
  if (!capabilities?.limits || !capabilities?.usage) {
    throw new Error('capabilities payload incomplete')
  }

  return {
    runId,
    createdAt: new Date().toISOString(),
    apiBase,
    wxCode,
    token: login.accessToken,
    user: profile,
    capabilities,
    capabilityLines: buildCapabilityLines(capabilities),
    localStorage: {
      tree_miniapp_user_token: login.accessToken,
      tree_miniapp_user: JSON.stringify(profile)
    },
    playwrightInitScript: `
      localStorage.setItem('tree_miniapp_user_token', ${JSON.stringify(login.accessToken)});
      localStorage.setItem('tree_miniapp_user', ${JSON.stringify(JSON.stringify(profile))});
    `.trim()
  }
}

export function sessionFilePath(runId) {
  return `/tmp/tree-quota-e2e-session-${runId}.json`
}

export function writeSessionFile(session, outFile = sessionFilePath(session.runId)) {
  fs.writeFileSync(outFile, `${JSON.stringify(session, null, 2)}\n`, { encoding: 'utf8', mode: 0o600 })
  return outFile
}
