import type { TreeNode } from '@/types/api'

import {
  buildMemberGraph,
  collectLineageChildren,
  findTreeRoots,
  isLineageMember,
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
  if (!anchor || !isLineageMember(anchor)) return null

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

  const sortedParents = [anchor, ...spouses]

  const childIdSet = new Set<number>()
  for (const parent of sortedParents) {
    for (const childId of collectLineageChildren(graph, parent.memberId)) {
      childIdSet.add(childId)
    }
  }

  const childBranches: FamilyTreeBranchNode[] = []
  for (const childId of sortChildIds([...childIdSet], graph.nodeMap)) {
    if (!isLineageMember(graph.nodeMap.get(childId))) continue
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
      unlocatedMemberIds: [],
      memberCount: 0,
      renderedCount: 0,
      message: '暂无成员'
    }
  }

  try {
    const graph = buildMemberGraph(input)
    const rendered = new Set<number>()
    const roots: FamilyTreeBranchNode[] = []
    const rootIds = findTreeRoots(graph)

    for (const rootId of rootIds) {
      if (rendered.has(rootId)) continue
      const branch = buildCoupleBranch(rootId, graph, rendered, new Set())
      if (branch) roots.push(branch)
    }

    const unlocatedMemberIds = [...graph.nodeMap.keys()].filter(
      (memberId) => !rendered.has(memberId) && isLineageMember(graph.nodeMap.get(memberId))
    )

    return {
      success: roots.length > 0,
      roots,
      orphanBranches: [],
      unlocatedMemberIds,
      memberCount,
      renderedCount: rendered.size,
      message: roots.length === 0 ? '家谱结构暂时无法生成，可查看关系明细' : undefined
    }
  } catch {
    return {
      success: false,
      roots: [],
      orphanBranches: [],
      unlocatedMemberIds: [],
      memberCount,
      renderedCount: 0,
      message: '家谱结构暂时无法生成，可查看关系明细'
    }
  }
}
