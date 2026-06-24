import type { TreeEdge, TreeNode } from '@/types/api'

import { buildRelGraph, type RelGraph } from './relativeGraph'
import {
  buildTitleCacheKey,
  getCachedRelativeTitle,
  setCachedRelativeTitle
} from './relativeTitleCache'
import { resolveRelativeTitle } from './relativeRules'

export interface BuildKinshipTitleMapInput {
  viewerMemberId: number | null | undefined
  nodes: TreeNode[]
  edges: TreeEdge[]
  graphVersion?: number | string | null
}

/**
 * 基于当前视角（viewer）与目标成员在关系图中的结构，动态推导相对称谓。
 * 无 viewer 时不推导（返回 undefined）。
 */
export function getRelativeTitle(
  viewerMemberId: number | null | undefined,
  targetMemberId: number,
  graph: RelGraph
): string | undefined {
  if (viewerMemberId == null) return undefined
  if (viewerMemberId === targetMemberId) return '我'
  return resolveRelativeTitle(graph, viewerMemberId, targetMemberId)
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

  const title = getRelativeTitle(viewerMemberId, targetMemberId, graph) ?? '亲属'
  setCachedRelativeTitle(cacheKey, title)
  return title
}

/**
 * 为当前视角批量计算节点称谓（每次 viewer / graphVersion 变化时全量重算）。
 * 不写入节点数据，仅返回瞬时映射供 UI 绑定。
 */
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
  return getRelativeTitle(meMemberId, targetMemberId, graph) ?? '亲属'
}
