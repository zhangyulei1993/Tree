import type { TreeNode } from '@/types/api'

import type { RelGraph } from './relativeGraph'
import { genderOf, isSiblingOf } from './relativeGraph'

const MISSING = Number.MAX_SAFE_INTEGER

export function birthYearOf(node?: TreeNode): number {
  if (!node) return MISSING
  if (node.birthDate) {
    const match = /^(\d{4})/.exec(node.birthDate.replace(/\./g, '-').trim())
    if (match) return Number(match[1])
  }
  const extended = node as TreeNode & { birthYear?: number | null }
  if (extended.birthYear != null && !Number.isNaN(Number(extended.birthYear))) {
    return Number(extended.birthYear)
  }
  return MISSING
}

export function birthSortKey(node?: TreeNode): number {
  if (!node?.birthDate) return birthYearOf(node) === MISSING ? MISSING : birthYearOf(node) * 10000
  const normalized = node.birthDate.replace(/\./g, '-').trim()
  const match = /^(\d{4})(?:-(\d{1,2}))?(?:-(\d{1,2}))?/.exec(normalized)
  if (!match) return birthYearOf(node) === MISSING ? MISSING : birthYearOf(node) * 10000
  const year = Number(match[1])
  const month = match[2] ? Number(match[2]) : 1
  const day = match[3] ? Number(match[3]) : 1
  return year * 10000 + month * 100 + day
}

export function compareByBirth(a: TreeNode, b: TreeNode): number {
  const ka = birthSortKey(a)
  const kb = birthSortKey(b)
  if (ka !== kb && ka !== MISSING && kb !== MISSING) return ka - kb
  if (ka !== MISSING && kb === MISSING) return -1
  if (ka === MISSING && kb !== MISSING) return 1
  return a.memberId - b.memberId
}

export function groupHasCompleteBirth(graph: RelGraph, memberIds: number[]): boolean {
  if (memberIds.length <= 1) return true
  return memberIds.every((id) => {
    const node = graph.nodeMap.get(id)
    return birthSortKey(node) !== MISSING
  })
}

/** 同父母、同性别兄弟姐妹组内排序 */
export function sameParentSiblingGroup(
  graph: RelGraph,
  anchorId: number,
  gender: string
): number[] {
  const ids = siblingsOfSameGender(graph, anchorId, gender)
  return ids
    .map((id) => graph.nodeMap.get(id)!)
    .filter(Boolean)
    .sort(compareByBirth)
    .map((n) => n.memberId)
}

function siblingsOfSameGender(graph: RelGraph, memberId: number, gender: string): number[] {
  const ids = new Set<number>()
  if (genderOf(graph, memberId) === gender) ids.add(memberId)
  for (const parentId of graph.parents.get(memberId) || []) {
    for (const childId of graph.children.get(parentId) || []) {
      if (genderOf(graph, childId) === gender) ids.add(childId)
    }
  }
  return [...ids]
}

export type Seniority = 'older' | 'younger' | 'unknown'

type BirthEvidence = 'full-date' | 'year-only' | 'none'

function hasFullBirthDate(node?: TreeNode): boolean {
  if (!node?.birthDate) return false
  const normalized = node.birthDate.replace(/\./g, '-').trim()
  return /^\d{4}-\d{1,2}-\d{1,2}/.test(normalized)
}

function birthEvidence(node?: TreeNode): BirthEvidence {
  if (!node) return 'none'
  if (hasFullBirthDate(node)) return 'full-date'
  if (birthYearOf(node) !== MISSING) return 'year-only'
  return 'none'
}

/** 仅在有可靠出生证据时比较长幼，不使用 memberId 或数组顺序 */
export function compareReliableSeniority(graph: RelGraph, meId: number, targetId: number): Seniority {
  const me = graph.nodeMap.get(meId)
  const target = graph.nodeMap.get(targetId)
  if (!me || !target) return 'unknown'

  const meEvidence = birthEvidence(me)
  const targetEvidence = birthEvidence(target)
  if (meEvidence === 'none' || targetEvidence === 'none') return 'unknown'
  if (meEvidence !== targetEvidence) return 'unknown'

  if (meEvidence === 'full-date') {
    const ka = birthSortKey(me)
    const kb = birthSortKey(target)
    if (ka === MISSING || kb === MISSING || ka === kb) return 'unknown'
    if (kb < ka) return 'older'
    if (kb > ka) return 'younger'
    return 'unknown'
  }

  const aYear = birthYearOf(me)
  const bYear = birthYearOf(target)
  if (aYear === MISSING || bYear === MISSING || aYear === bYear) return 'unknown'
  if (bYear < aYear) return 'older'
  if (bYear > aYear) return 'younger'
  return 'unknown'
}

export function compareSeniority(graph: RelGraph, meId: number, targetId: number): Seniority {
  return compareReliableSeniority(graph, meId, targetId)
}

/** 组内排行标签：长、次、三、幼（与基础称谓分离） */
export function rankLabelInGroup(index: number, total: number, complete: boolean): string {
  if (!complete || total <= 1) return ''
  if (index === 0) return '长'
  if (index === total - 1 && total >= 3) return '幼'
  const middle = ['次', '三', '四', '五', '六', '七', '八', '九']
  return middle[index - 1] || ''
}

/** @deprecated 使用 rankLabelInGroup；保留兼容旧调用 */
export function rankPrefixInGroup(index: number, total: number, complete: boolean): string {
  return rankLabelInGroup(index, total, complete)
}

export function youngerUncleRankLabel(
  indexInYounger: number,
  youngerTotal: number,
  olderCount: number,
  complete: boolean
): string {
  if (!complete) return ''
  if (youngerTotal <= 1) return ''
  if (indexInYounger === youngerTotal - 1) return '幼'
  const globalRank = olderCount + indexInYounger + 1
  const numerals = ['', '长', '次', '三', '四', '五', '六', '七', '八', '九']
  return numerals[globalRank] || ''
}

/** @deprecated 使用 youngerUncleRankLabel */
export function youngerUnclePrefix(
  indexInYounger: number,
  youngerTotal: number,
  olderCount: number,
  complete: boolean
): string {
  return youngerUncleRankLabel(indexInYounger, youngerTotal, olderCount, complete)
}

export function rankedSiblingTitle(
  graph: RelGraph,
  meId: number,
  targetId: number
): string {
  const result = rankedSiblingCanonical(graph, meId, targetId)
  if (!result) return ''
  return result.canonicalTitle
}

export function rankedSiblingCanonical(
  graph: RelGraph,
  meId: number,
  targetId: number
): { canonicalTitle: string; rankLabel?: string } | null {
  if (!isSiblingOf(graph, meId, targetId)) return null

  const gender = genderOf(graph, targetId)
  const peers = sameParentSiblingGroup(graph, meId, gender)
  const older = peers.filter((id) => compareSeniority(graph, meId, id) === 'older')
  const younger = peers.filter((id) => compareSeniority(graph, meId, id) === 'younger')
  const seniority = compareSeniority(graph, meId, targetId)

  if (gender === 'MALE') {
    if (seniority === 'older') {
      const idx = older.indexOf(targetId)
      const complete = groupHasCompleteBirth(graph, older)
      const rankLabel = rankLabelInGroup(idx, older.length, complete) || undefined
      return { canonicalTitle: '兄长', rankLabel }
    }
    if (seniority === 'younger') {
      const idx = younger.indexOf(targetId)
      const complete = groupHasCompleteBirth(graph, younger)
      const rankLabel = rankLabelInGroup(idx, younger.length, complete) || undefined
      return { canonicalTitle: '弟弟', rankLabel }
    }
    return null
  }
  if (gender === 'FEMALE') {
    if (seniority === 'older') {
      const idx = older.indexOf(targetId)
      const complete = groupHasCompleteBirth(graph, older)
      const rankLabel = rankLabelInGroup(idx, older.length, complete) || undefined
      return { canonicalTitle: '姐姐', rankLabel }
    }
    if (seniority === 'younger') {
      const idx = younger.indexOf(targetId)
      const complete = groupHasCompleteBirth(graph, younger)
      const rankLabel = rankLabelInGroup(idx, younger.length, complete) || undefined
      return { canonicalTitle: '妹妹', rankLabel }
    }
    return null
  }
  return null
}
