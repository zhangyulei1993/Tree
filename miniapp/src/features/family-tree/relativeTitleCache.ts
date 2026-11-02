/** 称谓缓存：key = viewerMemberId:targetMemberId:v{graphVersion} */

const titleCache = new Map<string, string>()

export function buildTitleCacheKey(
  viewerMemberId: number,
  targetMemberId: number,
  graphVersion: number | string
): string {
  return `${viewerMemberId}:${targetMemberId}:v${graphVersion}`
}

export function getCachedRelativeTitle(key: string): string | undefined {
  return titleCache.get(key)
}

export function setCachedRelativeTitle(key: string, title: string): void {
  titleCache.set(key, title)
}

export function clearRelativeTitleCache(): void {
  titleCache.clear()
}

/** 测试 / graphVersion 变更后可选清理旧版本条目 */
export function pruneRelativeTitleCacheExcept(graphVersion: number | string): void {
  const suffix = `:v${graphVersion}`
  for (const key of titleCache.keys()) {
    if (!key.endsWith(suffix)) titleCache.delete(key)
  }
}
