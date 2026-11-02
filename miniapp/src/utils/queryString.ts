export function safeDecodeURIComponent(value: string): string {
  try {
    return decodeURIComponent(value)
  } catch {
    return value
  }
}

export function readQueryParam(query: string, keys: string[]): string {
  const normalized = query.startsWith('?') ? query.slice(1) : query
  if (!normalized) return ''

  const pairs = normalized.split('&').flatMap((segment) => {
    if (!segment) return []
    const separatorIndex = segment.indexOf('=')
    const rawKey = separatorIndex >= 0 ? segment.slice(0, separatorIndex) : segment
    const rawValue = separatorIndex >= 0 ? segment.slice(separatorIndex + 1) : ''
    return [{ key: safeDecodeURIComponent(rawKey), value: safeDecodeURIComponent(rawValue) }]
  })

  for (const expectedKey of keys) {
    const pair = pairs.find((item) => item.key === expectedKey)
    if (pair) {
      return pair.value
    }
  }

  return ''
}
