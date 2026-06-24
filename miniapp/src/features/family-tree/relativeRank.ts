import type { TreeNode } from '@/types/api'

import type { RelGraph } from './relativeGraph'
import { genderOf } from './relativeGraph'

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

export function compareSeniority(graph: RelGraph, meId: number, targetId: number): Seniority {
  const me = graph.nodeMap.get(meId)
  const target = graph.nodeMap.get(targetId)
  if (!me || !target) return 'unknown'
  const ka = birthSortKey(me)
  const kb = birthSortKey(target)
  if (ka === MISSING || kb === MISSING) return 'unknown'
  if (kb < ka) return 'older'
  if (kb > ka) return 'younger'
  if (targetId < meId) return 'older'
  if (targetId > meId) return 'younger'
  return 'unknown'
}

/** 组内排行前缀：大/二/三/小；信息不足返回空 */
export function rankPrefixInGroup(index: number, total: number, complete: boolean): string {
  if (!complete || total <= 1) return ''
  if (index === 0) return '大'
  if (index === total - 1 && total >= 3) return '小'
  const middle = ['二', '三', '四', '五', '六', '七', '八', '九']
  return middle[index - 1] || ''
}

/** 叔辈：在兄数之后延续 三/四/…/小 */
export function youngerUnclePrefix(
  indexInYounger: number,
  youngerTotal: number,
  olderCount: number,
  complete: boolean
): string {
  if (!complete) return ''
  if (youngerTotal <= 1) return ''
  if (indexInYounger === youngerTotal - 1) return '小'
  const globalRank = olderCount + indexInYounger + 1
  const numerals = ['', '一', '二', '三', '四', '五', '六', '七', '八', '九']
  const label = numerals[globalRank] || ''
  return label === '一' ? '大' : label
}

export function rankedSiblingTitle(
  graph: RelGraph,
  meId: number,
  targetId: number
): string {
  const gender = genderOf(graph, targetId)
  const peers = sameParentSiblingGroup(graph, meId, gender)
  const older = peers.filter((id) => compareSeniority(graph, meId, id) === 'older')
  const younger = peers.filter((id) => compareSeniority(graph, meId, id) === 'younger')
  const seniority = compareSeniority(graph, meId, targetId)

  if (gender === 'MALE') {
    if (seniority === 'older') {
      const idx = older.indexOf(targetId)
      const complete = groupHasCompleteBirth(graph, older)
      if (complete && older.length === 1) return '大哥'
      const prefix = rankPrefixInGroup(idx, older.length, complete)
      return prefix ? `${prefix}哥` : '哥哥'
    }
    if (seniority === 'younger') {
      const idx = younger.indexOf(targetId)
      const complete = groupHasCompleteBirth(graph, younger)
      if (complete && younger.length === 1) return '弟弟'
      const prefix = rankPrefixInGroup(idx, younger.length, complete)
      return prefix ? `${prefix}弟` : '弟弟'
    }
    return '兄弟'
  }
  if (gender === 'FEMALE') {
    if (seniority === 'older') {
      const idx = older.indexOf(targetId)
      const complete = groupHasCompleteBirth(graph, older)
      if (complete && older.length === 1) return '大姐'
      const prefix = rankPrefixInGroup(idx, older.length, complete)
      return prefix ? `${prefix}姐` : '姐姐'
    }
    if (seniority === 'younger') {
      const idx = younger.indexOf(targetId)
      const complete = groupHasCompleteBirth(graph, younger)
      if (complete && younger.length === 1) return '小妹'
      const prefix = rankPrefixInGroup(idx, younger.length, complete)
      return prefix ? `${prefix}妹` : '妹妹'
    }
    return '姐妹'
  }
  return '同辈亲属'
}
