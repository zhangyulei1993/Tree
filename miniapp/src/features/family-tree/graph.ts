import type { TreeEdge, TreeItem, TreeNode } from '@/types/api'

import { birthSortValue, compareMaleChildren, sortChildMemberIds } from './memberSort'
import type { FamilyTreeBuildInput, MemberGraph } from './types'

export { birthSortValue, sortChildMemberIds } from './memberSort'

export function isLineageMember(node?: TreeNode): boolean {
  if (!node) return false
  return String(node.memberType || '').toUpperCase() === 'LINEAGE_MEMBER'
}

export function isSpouseMember(node?: TreeNode): boolean {
  return String(node?.memberType || '').toUpperCase() === 'SPOUSE'
}

export function isExternalMember(node?: TreeNode): boolean {
  return String(node?.memberType || '').toUpperCase() === 'EXTERNAL_MEMBER'
}

function pushMap(map: Map<number, number[]>, key: number, value: number) {
  if (key === value) return
  const list = map.get(key)
  if (!list) {
    map.set(key, [value])
    return
  }
  if (!list.includes(value)) list.push(value)
}

function mergeTreeItems(graph: MemberGraph, items: TreeItem[] | undefined) {
  for (const item of items || []) {
    for (const parentId of item.parentIds || []) {
      pushMap(graph.parentIds, item.memberId, parentId)
      pushMap(graph.childrenIds, parentId, item.memberId)
    }
    for (const childId of item.childrenIds || []) {
      pushMap(graph.childrenIds, item.memberId, childId)
      pushMap(graph.parentIds, childId, item.memberId)
    }
    for (const spouseId of item.spouseIds || []) {
      pushMap(graph.spouseIds, item.memberId, spouseId)
      pushMap(graph.spouseIds, spouseId, item.memberId)
    }
  }
}

export function buildMemberGraph(input: FamilyTreeBuildInput): MemberGraph {
  const nodeMap = new Map<number, TreeNode>()
  for (const node of input.nodes) {
    nodeMap.set(node.memberId, node)
  }

  const graph: MemberGraph = {
    parentIds: new Map(),
    childrenIds: new Map(),
    spouseIds: new Map(),
    nodeMap
  }

  for (const edge of input.edges) {
    if (edge.relationshipType === 'SIBLING') continue
    if (edge.relationshipType === 'PARENT_CHILD') {
      pushMap(graph.childrenIds, edge.fromMemberId, edge.toMemberId)
      pushMap(graph.parentIds, edge.toMemberId, edge.fromMemberId)
      continue
    }
    if (edge.relationshipType === 'SPOUSE') {
      pushMap(graph.spouseIds, edge.fromMemberId, edge.toMemberId)
      pushMap(graph.spouseIds, edge.toMemberId, edge.fromMemberId)
    }
  }

  mergeTreeItems(graph, input.tree)
  return graph
}

export function compareMembers(a: TreeNode, b: TreeNode): number {
  return compareMaleChildren(a, b)
}

export function lineageParentIds(graph: MemberGraph, memberId: number): number[] {
  return (graph.parentIds.get(memberId) || []).filter((parentId) =>
    isLineageMember(graph.nodeMap.get(parentId))
  )
}

export function lineageFatherId(graph: MemberGraph, memberId: number): number | null {
  for (const parentId of lineageParentIds(graph, memberId)) {
    if (graph.nodeMap.get(parentId)?.gender === 'MALE') return parentId
  }
  return null
}

export function collectChildren(
  memberId: number,
  childrenIds: Map<number, number[]>,
  spouseIds: Map<number, number[]>
): number[] {
  const result = new Set<number>(childrenIds.get(memberId) || [])
  for (const spouseId of spouseIds.get(memberId) || []) {
    for (const childId of childrenIds.get(spouseId) || []) {
      result.add(childId)
    }
  }
  return [...result]
}

/** 仅收集 LINEAGE_MEMBER 子代，用于主干递归 */
export function collectLineageChildren(graph: MemberGraph, memberId: number): number[] {
  return collectChildren(memberId, graph.childrenIds, graph.spouseIds).filter((childId) =>
    isLineageMember(graph.nodeMap.get(childId))
  )
}

function lineageMemberIds(graph: MemberGraph): number[] {
  return [...graph.nodeMap.keys()].filter((id) => isLineageMember(graph.nodeMap.get(id)))
}

function buildLineageAdjacency(graph: MemberGraph): Map<number, Set<number>> {
  const ids = lineageMemberIds(graph)
  const idSet = new Set(ids)
  const adj = new Map<number, Set<number>>()
  for (const id of ids) adj.set(id, new Set())

  for (const id of ids) {
    for (const parentId of graph.parentIds.get(id) || []) {
      if (!idSet.has(parentId)) continue
      adj.get(id)!.add(parentId)
      adj.get(parentId)!.add(id)
    }
  }

  const parentToChildren = new Map<number, number[]>()
  for (const id of ids) {
    for (const parentId of graph.parentIds.get(id) || []) {
      const children = parentToChildren.get(parentId)
      if (children) children.push(id)
      else parentToChildren.set(parentId, [id])
    }
  }
  for (const children of parentToChildren.values()) {
    for (let i = 0; i < children.length; i += 1) {
      for (let j = i + 1; j < children.length; j += 1) {
        adj.get(children[i]!)!.add(children[j]!)
        adj.get(children[j]!)!.add(children[i]!)
      }
    }
  }

  return adj
}

function componentStableKey(graph: MemberGraph, memberIds: number[]): string {
  const sorted = memberIds.slice().sort((a, b) => {
    const na = graph.nodeMap.get(a)
    const nb = graph.nodeMap.get(b)
    const birthDiff = birthSortValue(na) - birthSortValue(nb)
    if (birthDiff !== 0) return birthDiff
    return (na?.displayName || '').localeCompare(nb?.displayName || '', 'zh-CN')
  })
  return sorted.map((id) => graph.nodeMap.get(id)?.displayName || '').join('|')
}

function componentRelationScore(graph: MemberGraph, memberIds: number[]): number {
  let score = 0
  for (const id of memberIds) {
    score += (graph.childrenIds.get(id) || []).length
    score += (graph.parentIds.get(id) || []).length
    score += (graph.spouseIds.get(id) || []).length
  }
  return score
}

function compareLineageComponents(graph: MemberGraph, a: number[], b: number[]): number {
  if (b.length !== a.length) return b.length - a.length
  const scoreDiff = componentRelationScore(graph, b) - componentRelationScore(graph, a)
  if (scoreDiff !== 0) return scoreDiff
  const oldestA = Math.min(...a.map((id) => birthSortValue(graph.nodeMap.get(id))))
  const oldestB = Math.min(...b.map((id) => birthSortValue(graph.nodeMap.get(id))))
  if (oldestA !== oldestB) return oldestA - oldestB
  return componentStableKey(graph, a).localeCompare(componentStableKey(graph, b), 'zh-CN')
}

export function findMainLineageComponent(graph: MemberGraph): number[] {
  const adj = buildLineageAdjacency(graph)
  const visited = new Set<number>()
  const components: number[][] = []

  for (const startId of lineageMemberIds(graph)) {
    if (visited.has(startId)) continue
    const queue = [startId]
    const component: number[] = []
    visited.add(startId)
    while (queue.length > 0) {
      const current = queue.shift()!
      component.push(current)
      for (const nextId of adj.get(current) || []) {
        if (visited.has(nextId)) continue
        visited.add(nextId)
        queue.push(nextId)
      }
    }
    components.push(component)
  }

  if (components.length === 0) return []

  return components.slice().sort((a, b) => compareLineageComponents(graph, a, b))[0]!
}

function pickMainRootId(graph: MemberGraph, candidateIds: number[]): number | null {
  if (candidateIds.length === 0) return null
  const sorted = candidateIds.slice().sort((a, b) => {
    const na = graph.nodeMap.get(a)
    const nb = graph.nodeMap.get(b)
    if (na?.gender === 'MALE' && nb?.gender !== 'MALE') return -1
    if (nb?.gender === 'MALE' && na?.gender !== 'MALE') return 1
    const birthDiff = birthSortValue(na) - birthSortValue(nb)
    if (birthDiff !== 0) return birthDiff
    return (na?.displayName || '').localeCompare(nb?.displayName || '', 'zh-CN')
  })
  return sorted[0] ?? null
}

export function findTreeRoots(graph: MemberGraph): number[] {
  const mainComponent = findMainLineageComponent(graph)
  if (mainComponent.length === 0) return []

  const apexes = mainComponent.filter((id) => lineageParentIds(graph, id).length === 0)
  const candidates = apexes.length > 0 ? apexes : mainComponent
  const mainRoot = pickMainRootId(graph, candidates)
  return mainRoot != null ? [mainRoot] : []
}

export function sortParents(nodes: TreeNode[]): TreeNode[] {
  return nodes.slice().sort((a, b) => {
    if (a.gender === 'MALE' && b.gender !== 'MALE') return -1
    if (b.gender === 'MALE' && a.gender !== 'MALE') return 1
    return compareMembers(a, b)
  })
}

export function filterUnlocatedLineageMemberIds(
  memberIds: number[],
  relatedMemberIds: Iterable<number>,
  memberTypeById: Map<number, string>
): number[] {
  const related = new Set(relatedMemberIds)
  return memberIds.filter((memberId) => {
    if (related.has(memberId)) return false
    const memberType = memberTypeById.get(memberId)
    if (!memberType) return false
    return memberType.toUpperCase() === 'LINEAGE_MEMBER'
  })
}

export function sortChildIds(childIds: number[], nodeMap: Map<number, TreeNode>): number[] {
  return sortChildMemberIds(childIds, nodeMap)
}
