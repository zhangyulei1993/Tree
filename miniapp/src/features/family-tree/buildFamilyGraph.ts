import type { TreeNode } from '@/types/api'

import {
  buildMemberGraph,
  collectChildren,
  findTreeRoots,
  sortChildIds
} from './graph'
import { compareSpouses } from './memberSort'
import type { BuiltFamilyGraph, FamilyTreeBranchNode, FamilyTreeBuildInput } from './types'

function buildCoupleBranch(
  anchorMemberId: number,
  graph: ReturnType<typeof buildMemberGraph>,
  rendered: Set<number>,
  visiting: Set<number>
): FamilyTreeBranchNode | null {
  if (visiting.has(anchorMemberId)) return null
  if (rendered.has(anchorMemberId)) return null

  const anchor = graph.nodeMap.get(anchorMemberId)
  if (!anchor) return null

  visiting.add(anchorMemberId)

  const parents: TreeNode[] = [anchor]
  rendered.add(anchorMemberId)

  const spouses = (graph.spouseIds.get(anchorMemberId) || [])
    .map((spouseId) => graph.nodeMap.get(spouseId))
    .filter((node): node is TreeNode => Boolean(node) && !rendered.has(node!.memberId))
    .sort(compareSpouses)

  for (const spouse of spouses) {
    parents.push(spouse)
    rendered.add(spouse.memberId)
    visiting.add(spouse.memberId)
  }

  // 主成员在左，配偶始终在右
  const sortedParents = [anchor, ...spouses]

  const childIdSet = new Set<number>()
  for (const parent of sortedParents) {
    for (const childId of collectChildren(parent.memberId, graph.childrenIds, graph.spouseIds)) {
      childIdSet.add(childId)
    }
  }

  const childBranches: FamilyTreeBranchNode[] = []
  for (const childId of sortChildIds([...childIdSet], graph.nodeMap)) {
    const nextVisiting = new Set(visiting)
    const branch = buildCoupleBranch(childId, graph, rendered, nextVisiting)
    if (branch) childBranches.push(branch)
  }

  visiting.delete(anchorMemberId)
  for (const spouse of spouses) visiting.delete(spouse.memberId)

  return {
    id: sortedParents
      .map((parent) => parent.memberId)
      .sort((a, b) => a - b)
      .join('-'),
    parents: sortedParents,
    childBranches
  }
}

export function buildFamilyGraph(input: FamilyTreeBuildInput): BuiltFamilyGraph {
  const memberCount = input.nodes.length
  if (memberCount === 0) {
    return {
      success: false,
      roots: [],
      orphanBranches: [],
      memberCount: 0,
      renderedCount: 0,
      message: '暂无成员'
    }
  }

  try {
    const graph = buildMemberGraph(input)
    const rendered = new Set<number>()
    const roots: FamilyTreeBranchNode[] = []

    for (const rootId of findTreeRoots(graph)) {
      if (rendered.has(rootId)) continue
      const branch = buildCoupleBranch(rootId, graph, rendered, new Set())
      if (branch) roots.push(branch)
    }

    const orphanMemberIds = [...graph.nodeMap.keys()].filter((memberId) => !rendered.has(memberId))
    const orphanBranches: FamilyTreeBranchNode[] = []

    for (const memberId of sortChildIds(orphanMemberIds, graph.nodeMap)) {
      if (rendered.has(memberId)) continue
      const branch = buildCoupleBranch(memberId, graph, rendered, new Set())
      if (branch) orphanBranches.push(branch)
    }

    return {
      success: roots.length > 0 || orphanBranches.length > 0,
      roots,
      orphanBranches,
      memberCount,
      renderedCount: rendered.size,
      message: roots.length === 0 ? '家谱结构暂时无法生成，可查看关系明细' : undefined
    }
  } catch {
    return {
      success: false,
      roots: [],
      orphanBranches: [],
      memberCount,
      renderedCount: 0,
      message: '家谱结构暂时无法生成，可查看关系明细'
    }
  }
}
