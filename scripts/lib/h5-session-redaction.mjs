const JWT_PATTERN = /eyJ[A-Za-z0-9_-]{8,}\.[A-Za-z0-9_-]{8,}\.[A-Za-z0-9_-]{8,}/

const SENSITIVE_KEYS = new Set([
  'accessToken',
  'refreshToken',
  'token',
  'authorization',
  'Authorization',
  'tree_miniapp_user_token',
  'tree_miniapp_user',
  'password',
  'devCode',
  'session_key',
  'openid',
  'unionid'
])

export function redactToken(token) {
  if (!token || typeof token !== 'string') return ''
  return `${token.slice(0, 8)}…(${token.length})`
}

export function looksLikeJwt(value) {
  return typeof value === 'string' && JWT_PATTERN.test(value)
}

export function sanitizeLogValue(value, depth = 0) {
  if (value == null || depth > 6) return value
  if (typeof value === 'string') {
    if (looksLikeJwt(value)) return redactToken(value)
    if (value.startsWith('Bearer ')) return `Bearer ${redactToken(value.slice(7))}`
    return value
  }
  if (Array.isArray(value)) {
    return value.map((item) => sanitizeLogValue(item, depth + 1))
  }
  if (typeof value === 'object') {
    const result = {}
    for (const [key, item] of Object.entries(value)) {
      if (SENSITIVE_KEYS.has(key)) {
        if (key === 'tree_miniapp_user' && typeof item === 'string') {
          result[key] = '[redacted-user-json]'
        } else if ((key === 'Authorization' || key === 'authorization') && typeof item === 'string' && item.startsWith('Bearer ')) {
          result[key] = `Bearer ${redactToken(item.slice(7))}`
        } else {
          result[key] = redactToken(String(item))
        }
        continue
      }
      result[key] = sanitizeLogValue(item, depth + 1)
    }
    return result
  }
  return value
}

export function formatPlaywrightSessionExample() {
  return [
    'Playwright example:',
    'await page.addInitScript((payload) => {',
    "  localStorage.setItem('tree_miniapp_user_token', payload.tree_miniapp_user_token)",
    "  localStorage.setItem('tree_miniapp_user', payload.tree_miniapp_user)",
    '}, SESSION_INJECT_PAYLOAD)'
  ]
}

export function assertOutputHasNoRawJwt(output) {
  const matches = String(output).match(new RegExp(JWT_PATTERN.source, 'g')) || []
  if (matches.length > 0) {
    throw new Error(`output contains ${matches.length} unredacted JWT fragment(s)`)
  }
}

export function assertOutputHasNoBearerToken(output) {
  if (/Bearer\s+eyJ[A-Za-z0-9_-]{8,}/.test(String(output))) {
    throw new Error('output contains raw Bearer token')
  }
}
