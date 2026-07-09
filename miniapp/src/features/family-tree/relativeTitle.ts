import type { TreeEdge, TreeNode } from '@/types/api'

import type { CanonicalKinshipResult } from '../kinship/canonicalTypes'
import {
  bloodGenerationGap,
  buildRelGraph,
  childrenOf,
  findBloodDescendantPath,
  genderOf,
  isSiblingOf,
  isSpouseOf,
  parentsOf,
  siblingsOf,
  type RelGraph
} from './relativeGraph'
import { resolveCanonicalRelativeTitle } from './relativeRules'
import {
  buildTitleCacheKey,
  getCachedRelativeTitle,
  setCachedRelativeTitle
} from './relativeTitleCache'

export interface BuildKinshipTitleMapInput {
  viewerMemberId: number | null | undefined
  nodes: TreeNode[]
  edges: TreeEdge[]
  graphVersion?: number | string | null
}

export interface KinshipTitleEntry {
  canonicalTitle?: string
  rankLabel?: string
  displayLabel: string
  unsupportedCanonicalTitle: boolean
  pathDescription: string
}

function toDisplayLabel(result: CanonicalKinshipResult): string {
  if (!result.unsupportedCanonicalTitle && result.canonicalTitle) {
    return result.canonicalTitle
  }
  if (result.incompleteInfo) {
    return '关系信息不完整'
  }
  return result.pathDescription
}

function siblingFallbackLabel(graph: RelGraph, viewerMemberId: number, targetMemberId: number): string {
  if (!isSiblingOf(graph, viewerMemberId, targetMemberId)) return ''
  const gender = genderOf(graph, targetMemberId)
  if (gender === 'MALE') return '兄弟'
  if (gender === 'FEMALE') return '姐妹'
  return '兄弟姐妹'
}

function parentSiblingFallbackLabel(
  graph: RelGraph,
  viewerMemberId: number,
  targetMemberId: number
): string {
  for (const parentId of parentsOf(graph, viewerMemberId)) {
    if (!isSiblingOf(graph, parentId, targetMemberId)) continue

    const parentGender = genderOf(graph, parentId)
    const targetGender = genderOf(graph, targetMemberId)
    if (parentGender === 'MALE') {
      if (targetGender === 'MALE') return '叔伯'
      if (targetGender === 'FEMALE') return '姑母'
      return '父辈旁亲'
    }
    if (parentGender === 'FEMALE') {
      if (targetGender === 'MALE') return '舅父'
      if (targetGender === 'FEMALE') return '姨母'
      return '母辈旁亲'
    }
  }
  return ''
}

function grandparentSiblingFallbackLabel(
  graph: RelGraph,
  viewerMemberId: number,
  targetMemberId: number
): string {
  for (const parentId of parentsOf(graph, viewerMemberId)) {
    for (const grandparentId of parentsOf(graph, parentId)) {
      if (!isSiblingOf(graph, grandparentId, targetMemberId)) continue

      const parentGender = genderOf(graph, parentId)
      const grandparentGender = genderOf(graph, grandparentId)
      const targetGender = genderOf(graph, targetMemberId)

      if (parentGender === 'MALE' && grandparentGender === 'MALE') {
        if (targetGender === 'MALE') return '伯叔祖父'
        if (targetGender === 'FEMALE') return '姑祖母'
        return '祖辈旁亲'
      }
      if (parentGender === 'MALE' && grandparentGender === 'FEMALE') {
        if (targetGender === 'MALE') return '舅祖父'
        if (targetGender === 'FEMALE') return '姨祖母'
        return '祖辈旁亲'
      }
      if (parentGender === 'FEMALE' && grandparentGender === 'MALE') {
        if (targetGender === 'MALE') return '外伯叔祖父'
        if (targetGender === 'FEMALE') return '外姑祖母'
        return '外祖辈旁亲'
      }
      if (parentGender === 'FEMALE' && grandparentGender === 'FEMALE') {
        if (targetGender === 'MALE') return '外舅祖父'
        if (targetGender === 'FEMALE') return '外姨祖母'
        return '外祖辈旁亲'
      }
    }
  }
  return ''
}

function cousinFallbackLabel(
  graph: RelGraph,
  viewerMemberId: number,
  targetMemberId: number
): string {
  for (const parentId of parentsOf(graph, viewerMemberId)) {
    for (const parentSiblingId of siblingsOf(graph, parentId)) {
      if (!childrenOf(graph, parentSiblingId).includes(targetMemberId)) continue

      const isTang = genderOf(graph, parentId) === 'MALE' && genderOf(graph, parentSiblingId) === 'MALE'
      const prefix = isTang ? '堂' : '表'
      const targetGender = genderOf(graph, targetMemberId)
      if (targetGender === 'MALE') return `${prefix}兄弟`
      if (targetGender === 'FEMALE') return `${prefix}姐妹`
      return `${prefix}亲`
    }
  }
  return ''
}

function parentGenerationCousinLabel(prefix: '堂' | '表', gender: string): string {
  if (prefix === '堂') {
    if (gender === 'MALE') return '堂伯叔'
    if (gender === 'FEMALE') return '堂姑'
    return '堂亲长辈'
  }
  if (gender === 'MALE') return '表叔伯'
  if (gender === 'FEMALE') return '表姑姨'
  return '表亲长辈'
}

function parentGenerationCousinSpouseLabel(prefix: '堂' | '表', cousinGender: string): string {
  if (prefix === '堂') {
    if (cousinGender === 'MALE') return '堂伯叔母'
    if (cousinGender === 'FEMALE') return '堂姑父'
    return '堂亲长辈配偶'
  }
  if (cousinGender === 'MALE') return '表叔伯母'
  if (cousinGender === 'FEMALE') return '表姑姨父'
  return '表亲长辈配偶'
}

function collateralDescendantLabel(prefix: '堂' | '表', gap: number, gender: string): string {
  if (gap === 0) {
    if (gender === 'MALE') return `${prefix}兄弟`
    if (gender === 'FEMALE') return `${prefix}姐妹`
    return `${prefix}亲`
  }
  if (gap === 1) {
    if (gender === 'MALE') return `${prefix}侄`
    if (gender === 'FEMALE') return `${prefix}侄女`
  }
  if (gap === 2) {
    if (gender === 'MALE') return `${prefix}侄孙`
    if (gender === 'FEMALE') return `${prefix}侄孙女`
  }
  return ''
}

function collateralDescendantSpouseLabel(prefix: '堂' | '表', gap: number, bloodGender: string): string {
  if (gap === 0) return `${prefix}亲配偶`
  if (gap === 1) {
    if (bloodGender === 'MALE') return `${prefix}侄媳`
    if (bloodGender === 'FEMALE') return `${prefix}侄女婿`
    return `${prefix}侄辈配偶`
  }
  if (gap === 2) return `${prefix}侄孙辈配偶`
  return ''
}

function grandparentCollateralDescendantFallbackLabel(
  graph: RelGraph,
  viewerMemberId: number,
  targetMemberId: number
): string {
  for (const parentId of parentsOf(graph, viewerMemberId)) {
    for (const grandparentId of parentsOf(graph, parentId)) {
      const prefix: '堂' | '表' =
        genderOf(graph, parentId) === 'MALE' && genderOf(graph, grandparentId) === 'MALE'
          ? '堂'
          : '表'

      for (const grandparentSiblingId of siblingsOf(graph, grandparentId)) {
        for (const parentCousinId of childrenOf(graph, grandparentSiblingId)) {
          if (parentCousinId === targetMemberId) {
            return parentGenerationCousinLabel(prefix, genderOf(graph, targetMemberId))
          }
          if (isSpouseOf(graph, parentCousinId, targetMemberId)) {
            return parentGenerationCousinSpouseLabel(prefix, genderOf(graph, parentCousinId))
          }

          const bloodPath = findBloodDescendantPath(graph, parentCousinId, targetMemberId, 4)
          if (bloodPath) {
            const gap = bloodGenerationGap(graph, viewerMemberId, targetMemberId)
            if (gap == null) continue
            const label = collateralDescendantLabel(prefix, gap, genderOf(graph, targetMemberId))
            if (label) return label
          }

          for (const descendantId of collectDescendantIds(graph, parentCousinId, 4)) {
            if (!isSpouseOf(graph, descendantId, targetMemberId)) continue
            const gap = bloodGenerationGap(graph, viewerMemberId, descendantId)
            if (gap == null) continue
            const label = collateralDescendantSpouseLabel(prefix, gap, genderOf(graph, descendantId))
            if (label) return label
          }
        }
      }
    }
  }
  return ''
}

function collectDescendantIds(graph: RelGraph, memberId: number, maxDepth: number): number[] {
  const result: number[] = []
  const queue: Array<{ id: number; depth: number }> = [{ id: memberId, depth: 0 }]
  const visited = new Set<number>([memberId])

  while (queue.length > 0) {
    const item = queue.shift()!
    if (item.depth >= maxDepth) continue
    for (const childId of childrenOf(graph, item.id)) {
      if (visited.has(childId)) continue
      visited.add(childId)
      result.push(childId)
      queue.push({ id: childId, depth: item.depth + 1 })
    }
  }
  return result
}

function kinshipFallbackLabel(
  graph: RelGraph,
  viewerMemberId: number,
  targetMemberId: number
): string {
  return (
    siblingFallbackLabel(graph, viewerMemberId, targetMemberId) ||
    parentSiblingFallbackLabel(graph, viewerMemberId, targetMemberId) ||
    grandparentSiblingFallbackLabel(graph, viewerMemberId, targetMemberId) ||
    grandparentCollateralDescendantFallbackLabel(graph, viewerMemberId, targetMemberId) ||
    cousinFallbackLabel(graph, viewerMemberId, targetMemberId)
  )
}

export function resolveKinshipTitleEntry(
  graph: RelGraph,
  viewerMemberId: number,
  targetMemberId: number
): KinshipTitleEntry {
  if (viewerMemberId === targetMemberId) {
    return {
      canonicalTitle: '我',
      displayLabel: '我',
      unsupportedCanonicalTitle: false,
      pathDescription: '我'
    }
  }
  const result = resolveCanonicalRelativeTitle(graph, viewerMemberId, targetMemberId)
  const displayLabel =
    result.unsupportedCanonicalTitle
      ? kinshipFallbackLabel(graph, viewerMemberId, targetMemberId) || toDisplayLabel(result)
      : toDisplayLabel(result)
  return {
    canonicalTitle: result.canonicalTitle,
    rankLabel: result.rankLabel,
    displayLabel,
    unsupportedCanonicalTitle: result.unsupportedCanonicalTitle,
    pathDescription: result.pathDescription
  }
}

/**
 * 基于当前视角与目标成员在关系图中的结构，动态推导规范称谓。
 */
export function getRelativeTitle(
  viewerMemberId: number | null | undefined,
  targetMemberId: number,
  graph: RelGraph
): string | undefined {
  if (viewerMemberId == null) return undefined
  const entry = resolveKinshipTitleEntry(graph, viewerMemberId, targetMemberId)
  return entry.displayLabel
}

function getRelativeTitleCached(
  viewerMemberId: number,
  targetMemberId: number,
  graph: RelGraph,
  graphVersion: number | string
): string {
  const cacheKey = buildTitleCacheKey(viewerMemberId, targetMemberId, graphVersion)
  const cached = getCachedRelativeTitle(cacheKey)
  if (cached !== undefined) return cached

  const title = getRelativeTitle(viewerMemberId, targetMemberId, graph) ?? ''
  setCachedRelativeTitle(cacheKey, title)
  return title
}

export function buildKinshipTitleMap(input: BuildKinshipTitleMapInput): Record<number, string> {
  const { viewerMemberId, nodes, edges } = input
  if (viewerMemberId == null) return {}

  const graphVersion = input.graphVersion ?? 0
  const graph = buildRelGraph(nodes, edges)
  const titles: Record<number, string> = { [viewerMemberId]: '我' }

  for (const node of nodes) {
    if (node.memberId === viewerMemberId) continue
    titles[node.memberId] = getRelativeTitleCached(
      viewerMemberId,
      node.memberId,
      graph,
      graphVersion
    )
  }

  return titles
}

/** @deprecated 使用 getRelativeTitle(viewerMemberId, targetMemberId, graph) */
export function getRelativeTitleFromMe(
  meMemberId: number,
  targetMemberId: number,
  nodes: TreeNode[],
  edges: TreeEdge[]
): string {
  const graph = buildRelGraph(nodes, edges)
  return getRelativeTitle(meMemberId, targetMemberId, graph) ?? ''
}
