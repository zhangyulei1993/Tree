import type { TreeEdge, TreeNode } from '@/types/api'

export interface RelGraph {
  nodeMap: Map<number, TreeNode>
  parents: Map<number, number[]>
  children: Map<number, number[]>
  spouses: Map<number, number[]>
}

function pushUnique(map: Map<number, number[]>, key: number, value: number) {
  if (key === value) return
  const list = map.get(key)
  if (!list) {
    map.set(key, [value])
    return
  }
  if (!list.includes(value)) list.push(value)
}

export function buildRelGraph(nodes: TreeNode[], edges: TreeEdge[]): RelGraph {
  const nodeMap = new Map<number, TreeNode>()
  for (const node of nodes) {
    nodeMap.set(node.memberId, node)
  }

  const parents = new Map<number, number[]>()
  const children = new Map<number, number[]>()
  const spouses = new Map<number, number[]>()

  for (const edge of edges) {
    if (edge.relationshipType === 'PARENT_CHILD') {
      pushUnique(children, edge.fromMemberId, edge.toMemberId)
      pushUnique(parents, edge.toMemberId, edge.fromMemberId)
      continue
    }
    if (edge.relationshipType === 'SPOUSE') {
      pushUnique(spouses, edge.fromMemberId, edge.toMemberId)
      pushUnique(spouses, edge.toMemberId, edge.fromMemberId)
    }
  }

  return { nodeMap, parents, children, spouses }
}

export function genderOf(graph: RelGraph, memberId: number): string {
  return graph.nodeMap.get(memberId)?.gender || ''
}

export function parentsOf(graph: RelGraph, memberId: number): number[] {
  return graph.parents.get(memberId) || []
}

export function childrenOf(graph: RelGraph, memberId: number): number[] {
  return graph.children.get(memberId) || []
}

export function spousesOf(graph: RelGraph, memberId: number): number[] {
  return graph.spouses.get(memberId) || []
}

export function isParentOf(graph: RelGraph, parentId: number, childId: number): boolean {
  return parentsOf(graph, childId).includes(parentId)
}

export function isSpouseOf(graph: RelGraph, aId: number, bId: number): boolean {
  return spousesOf(graph, aId).includes(bId)
}

export function sharedParentIds(graph: RelGraph, aId: number, bId: number): number[] {
  const bSet = new Set(parentsOf(graph, bId))
  return parentsOf(graph, aId).filter((id) => bSet.has(id))
}

/** 同父母兄弟姐妹（不含自己） */
export function siblingsOf(graph: RelGraph, memberId: number): number[] {
  const siblings = new Set<number>()
  for (const parentId of parentsOf(graph, memberId)) {
    for (const childId of childrenOf(graph, parentId)) {
      if (childId !== memberId) siblings.add(childId)
    }
  }
  return [...siblings]
}

export function isSiblingOf(graph: RelGraph, aId: number, bId: number): boolean {
  return sharedParentIds(graph, aId, bId).length > 0
}

/** 仅沿父母向上找血缘路径（me → … → ancestor） */
export function findBloodAncestorPath(
  graph: RelGraph,
  meId: number,
  ancestorId: number,
  maxDepth = 8
): number[] | null {
  if (meId === ancestorId) return [meId]

  const queue: number[][] = [[meId]]
  const visited = new Set<number>([meId])

  while (queue.length > 0) {
    const path = queue.shift()!
    const head = path[path.length - 1]
    if (path.length - 1 > maxDepth) continue

    for (const parentId of parentsOf(graph, head)) {
      if (visited.has(parentId)) continue
      visited.add(parentId)
      const nextPath = [...path, parentId]
      if (parentId === ancestorId) return nextPath
      queue.push(nextPath)
    }
  }

  return null
}

/** 仅沿子女向下找血缘路径（me → … → descendant） */
export function findBloodDescendantPath(
  graph: RelGraph,
  meId: number,
  descendantId: number,
  maxDepth = 8
): number[] | null {
  if (meId === descendantId) return [meId]

  const queue: number[][] = [[meId]]
  const visited = new Set<number>([meId])

  while (queue.length > 0) {
    const path = queue.shift()!
    const head = path[path.length - 1]
    if (path.length - 1 > maxDepth) continue

    for (const childId of childrenOf(graph, head)) {
      if (visited.has(childId)) continue
      visited.add(childId)
      const nextPath = [...path, childId]
      if (childId === descendantId) return nextPath
      queue.push(nextPath)
    }
  }

  return null
}

export function bloodGenerationDelta(path: number[]): number {
  return path.length - 1
}

function collectMemberChildren(graph: RelGraph, memberId: number): number[] {
  const result = new Set(childrenOf(graph, memberId))
  for (const spouseId of spousesOf(graph, memberId)) {
    for (const childId of childrenOf(graph, spouseId)) {
      result.add(childId)
    }
  }
  return [...result]
}

function stabilizeBloodGenerations(graph: RelGraph, memberIds: number[], generations: Map<number, number>) {
  let changed = true
  let guard = 0

  while (changed && guard < memberIds.length + 20) {
    changed = false
    guard += 1

    for (const memberId of memberIds) {
      for (const siblingId of siblingsOf(graph, memberId)) {
        const merged = Math.max(generations.get(memberId) ?? -1, generations.get(siblingId) ?? -1)
        if (merged < 0) continue
        if ((generations.get(memberId) ?? -1) !== merged) {
          generations.set(memberId, merged)
          changed = true
        }
        if ((generations.get(siblingId) ?? -1) !== merged) {
          generations.set(siblingId, merged)
          changed = true
        }
      }

      for (const spouseId of spousesOf(graph, memberId)) {
        const merged = Math.max(generations.get(memberId) ?? -1, generations.get(spouseId) ?? -1)
        if (merged < 0) continue
        if ((generations.get(memberId) ?? -1) !== merged) {
          generations.set(memberId, merged)
          changed = true
        }
        if ((generations.get(spouseId) ?? -1) !== merged) {
          generations.set(spouseId, merged)
          changed = true
        }
      }

      const parentGens = parentsOf(graph, memberId)
        .map((parentId) => generations.get(parentId))
        .filter((value): value is number => value !== undefined)
      if (parentGens.length > 0) {
        const childGeneration = Math.max(...parentGens) + 1
        if ((generations.get(memberId) ?? -1) !== childGeneration) {
          generations.set(memberId, childGeneration)
          changed = true
        }
      }
    }
  }

  for (const memberId of memberIds) {
    if (!generations.has(memberId)) {
      generations.set(memberId, 0)
    }
  }
}

/** 沿父母-子女（含配偶同步）为成员分配辈分序号，用于代差判断 */
export function assignBloodGenerations(graph: RelGraph): Map<number, number> {
  const memberIds = [...graph.nodeMap.keys()].map((id) => Number(id))
  const roots = memberIds.filter((id) => parentsOf(graph, id).length === 0)
  if (roots.length === 0 && memberIds.length > 0) {
    roots.push(memberIds[0])
  }

  const generations = new Map<number, number>()
  const visited = new Set<number>()
  const queue = roots.map((memberId) => ({ memberId, generation: 0 }))

  while (queue.length > 0) {
    const current = queue.shift()
    if (!current) continue
    const { memberId, generation } = current

    if (visited.has(memberId)) {
      const existing = generations.get(memberId)
      if (existing !== undefined && existing >= generation) continue
    }
    visited.add(memberId)
    generations.set(memberId, Math.max(generations.get(memberId) ?? -1, generation))

    const currentGeneration = generations.get(memberId) ?? generation

    for (const spouseId of spousesOf(graph, memberId)) {
      generations.set(spouseId, Math.max(generations.get(spouseId) ?? -1, currentGeneration))
      if (!visited.has(spouseId)) {
        queue.push({ memberId: spouseId, generation: currentGeneration })
      }
    }

    const childGeneration = currentGeneration + 1
    for (const childId of collectMemberChildren(graph, memberId)) {
      queue.push({ memberId: childId, generation: childGeneration })
    }
  }

  stabilizeBloodGenerations(graph, memberIds, generations)
  return generations
}

/** 正值表示 target 比 viewer 年轻（辈分更低） */
export function bloodGenerationGap(graph: RelGraph, viewerId: number, targetId: number): number | null {
  if (viewerId === targetId) return 0

  const generations = assignBloodGenerations(graph)
  const viewerGen = generations.get(viewerId)
  const targetGen = generations.get(targetId)
  if (viewerGen === undefined || targetGen === undefined) return null
  return targetGen - viewerGen
}

export function maternalCrossCount(graph: RelGraph, path: number[]): number {
  let count = 0
  for (let i = 1; i < path.length - 1; i += 1) {
    if (genderOf(graph, path[i]) === 'FEMALE') count += 1
  }
  return count
}
