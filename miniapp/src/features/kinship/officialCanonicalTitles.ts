import { CANONICAL_KINSHIP_TERMS } from './canonicalKinshipTerms'

/**
 * 唯一官方称谓清单：仅允许出现在 UI 中的规范亲属称谓。
 * 业务代码不得新增未登记字符串；相对/路径解析出口须通过 isOfficialCanonicalTitle 校验。
 */
export const OFFICIAL_CANONICAL_TITLE_LIST: readonly string[] = Object.freeze(
  [...new Set(Object.values(CANONICAL_KINSHIP_TERMS).map((term) => term.canonicalTitle))].sort(
    (a, b) => a.localeCompare(b, 'zh-CN')
  )
)

export const OFFICIAL_CANONICAL_TITLE_SET: ReadonlySet<string> = new Set(OFFICIAL_CANONICAL_TITLE_LIST)

/** 禁止作为最终展示结果的裸词/泛称 */
export const FORBIDDEN_BARE_CANONICAL_TITLES: ReadonlySet<string> = new Set([
  '亲属',
  '姻亲',
  '祖辈亲属',
  '父母辈亲属',
  '晚辈亲属',
  '同辈亲属',
  '祖辈旁系亲属',
  '父母',
  '子女',
  '兄弟姐妹',
  '配偶',
  '父辈堂亲',
  '母辈表亲',
  '堂叔伯'
])

/**
 * 堂/表同辈配偶口径（与兄弟姐妹之配偶「嫂子/弟媳」对称）：
 * - 堂兄之妻 → 堂嫂；堂弟之妻 → 堂弟媳
 * - 堂姐之夫 → 堂姐夫；堂妹之夫 → 堂妹夫
 * - 表亲同辈配偶 → 表嫂 / 表弟媳 / 表姐夫 / 表妹夫
 * 堂侄/表侄之配偶 → 堂侄媳 / 表侄媳（区别于堂嫂）
 *
 * 堂/表同辈长幼口径：
 * - 仅按「我」与堂/表亲本人的出生先后（或年龄）区分 兄/弟/姐/妹
 * - 不得用父母兄弟姐妹（伯父/叔父/舅父等）的长幼替代同辈长幼
 */
export const COUSIN_PEER_SENIORITY_POLICY =
  '堂表同辈称谓以本人相对年龄为准，不以父辈旁系长幼为准' as const

export const COUSIN_SPOUSE_TITLE_POLICY = {
  tang: {
    elderBrotherWife: '堂嫂',
    youngerBrotherWife: '堂弟媳',
    elderSisterHusband: '堂姐夫',
    youngerSisterHusband: '堂妹夫'
  },
  biao: {
    elderBrotherWife: '表嫂',
    youngerBrotherWife: '表弟媳',
    elderSisterHusband: '表姐夫',
    youngerSisterHusband: '表妹夫'
  }
} as const

export function isOfficialCanonicalTitle(title: string | undefined | null): boolean {
  if (!title) return false
  return OFFICIAL_CANONICAL_TITLE_SET.has(title)
}

export function isForbiddenBareCanonicalTitle(title: string | undefined | null): boolean {
  if (!title) return false
  return FORBIDDEN_BARE_CANONICAL_TITLES.has(title)
}

export function validateRegistryRelativeTitle(
  title: string | undefined | null
): { ok: true; title: string } | { ok: false; reason: 'empty' | 'forbidden' | 'unregistered' } {
  if (!title) return { ok: false, reason: 'empty' }
  if (isForbiddenBareCanonicalTitle(title)) return { ok: false, reason: 'forbidden' }
  if (!isOfficialCanonicalTitle(title)) return { ok: false, reason: 'unregistered' }
  return { ok: true, title }
}
