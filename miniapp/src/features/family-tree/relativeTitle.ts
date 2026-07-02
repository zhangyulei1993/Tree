import type { TreeEdge, TreeNode } from '@/types/api'

import type { CanonicalKinshipResult } from '../kinship/canonicalTypes'
import { buildRelGraph, type RelGraph } from './relativeGraph'
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
  return {
    canonicalTitle: result.canonicalTitle,
    rankLabel: result.rankLabel,
    displayLabel: toDisplayLabel(result),
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
