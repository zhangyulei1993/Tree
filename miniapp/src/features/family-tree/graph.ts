import type { TreeEdge, TreeItem, TreeNode } from '@/types/api'

import { birthSortValue, compareMaleChildren, sortChildMemberIds } from './memberSort'
import type { FamilyTreeBuildInput, MemberGraph } from './types'

export { birthSortValue, sortChildMemberIds } from './memberSort'

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

export function findTreeRoots(graph: MemberGraph): number[] {
  const memberIds = [...graph.nodeMap.keys()]
  const { parentIds, childrenIds, spouseIds, nodeMap } = graph

  let candidates = memberIds.filter((id) => !(parentIds.get(id)?.length))
  candidates = candidates.filter((id) => {
    const spouses = spouseIds.get(id) || []
    return !spouses.some((spouseId) => (parentIds.get(spouseId)?.length || 0) > 0)
  })

  const picked = new Set<number>()
  const roots: number[] = []
  const sorted = candidates
    .slice()
    .sort((a, b) => birthSortValue(nodeMap.get(a)) - birthSortValue(nodeMap.get(b)))

  for (const memberId of sorted) {
    if (picked.has(memberId)) continue

    const spouseList = spouseIds.get(memberId) || []
    const rootCoupleIds = [memberId, ...spouseList.filter((sid) => !(parentIds.get(sid)?.length))]
    const maleAnchor = rootCoupleIds.find((id) => nodeMap.get(id)?.gender === 'MALE')
    const anchorId = maleAnchor ?? memberId

    if (picked.has(anchorId)) continue
    roots.push(anchorId)
    picked.add(anchorId)
    for (const spouseId of spouseList) {
      if (!(parentIds.get(spouseId)?.length)) {
        picked.add(spouseId)
      }
    }
    for (const id of rootCoupleIds) picked.add(id)
  }

  if (roots.length > 0) return roots

  const withChildren = memberIds
    .filter((id) => (childrenIds.get(id)?.length || 0) > 0)
    .sort((a, b) => birthSortValue(nodeMap.get(a)) - birthSortValue(nodeMap.get(b)))

  if (withChildren.length > 0) return [withChildren[0]]
  return memberIds.slice().sort((a, b) => birthSortValue(nodeMap.get(a)) - birthSortValue(nodeMap.get(b)))
}

export function sortParents(nodes: TreeNode[]): TreeNode[] {
  return nodes.slice().sort((a, b) => {
    if (a.gender === 'MALE' && b.gender !== 'MALE') return -1
    if (b.gender === 'MALE' && a.gender !== 'MALE') return 1
    return compareMembers(a, b)
  })
}

export function sortChildIds(childIds: number[], nodeMap: Map<number, TreeNode>): number[] {
  return sortChildMemberIds(childIds, nodeMap)
}
