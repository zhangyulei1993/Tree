import type { TreeEdge, TreeItem, TreeNode } from '@/types/api'

import type {
  BuiltFamilyTree,
  FamilyTreeBuildInput,
  FamilyTreeGeneration,
  FamilyTreeUnit,
  MemberGraph
} from './types'

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

function buildGraph(input: FamilyTreeBuildInput): MemberGraph {
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

function birthSortValue(node?: TreeNode): number {
  if (!node?.birthDate) return Number.MAX_SAFE_INTEGER
  const match = /^(\d{4})/.exec(node.birthDate)
  return match ? Number(match[1]) : Number.MAX_SAFE_INTEGER
}

function compareMembers(a: TreeNode, b: TreeNode): number {
  const byBirth = birthSortValue(a) - birthSortValue(b)
  if (byBirth !== 0) return byBirth
  return a.displayName.localeCompare(b.displayName, 'zh-CN')
}

function findRoots(
  memberIds: number[],
  parentIds: Map<number, number[]>,
  childrenIds: Map<number, number[]>,
  spouseIds: Map<number, number[]>,
  nodeMap: Map<number, TreeNode>
): number[] {
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
    roots.push(memberId)
    picked.add(memberId)
    for (const spouseId of spouseIds.get(memberId) || []) {
      if (!(parentIds.get(spouseId)?.length)) {
        picked.add(spouseId)
      }
    }
  }

  if (roots.length > 0) return roots

  const withChildren = memberIds
    .filter((id) => (childrenIds.get(id)?.length || 0) > 0)
    .sort((a, b) => birthSortValue(nodeMap.get(a)) - birthSortValue(nodeMap.get(b)))

  if (withChildren.length > 0) return [withChildren[0]]
  return memberIds.slice().sort((a, b) => birthSortValue(nodeMap.get(a)) - birthSortValue(nodeMap.get(b)))
}

function collectChildren(memberId: number, childrenIds: Map<number, number[]>, spouseIds: Map<number, number[]>): number[] {
  const result = new Set<number>(childrenIds.get(memberId) || [])
  for (const spouseId of spouseIds.get(memberId) || []) {
    for (const childId of childrenIds.get(spouseId) || []) {
      result.add(childId)
    }
  }
  return [...result]
}

function assignGenerations(
  memberIds: number[],
  graph: MemberGraph,
  roots: number[]
): Map<number, number> {
  const generations = new Map<number, number>()
  const visited = new Set<number>()
  const queue: Array<{ memberId: number; generation: number }> = roots.map((memberId) => ({
    memberId,
    generation: 0
  }))

  while (queue.length > 0) {
    const current = queue.shift()
    if (!current) continue
    const { memberId, generation } = current

    if (visited.has(memberId)) {
      const existing = generations.get(memberId)
      if (existing !== undefined && existing <= generation) continue
    }
    visited.add(memberId)
    generations.set(memberId, Math.max(generations.get(memberId) ?? -1, generation))

    const currentGeneration = generations.get(memberId) ?? generation

    for (const spouseId of graph.spouseIds.get(memberId) || []) {
      generations.set(spouseId, Math.max(generations.get(spouseId) ?? -1, currentGeneration))
      if (!visited.has(spouseId)) {
        queue.push({ memberId: spouseId, generation: currentGeneration })
      }
    }

    const childGeneration = currentGeneration + 1
    for (const childId of collectChildren(memberId, graph.childrenIds, graph.spouseIds)) {
      queue.push({ memberId: childId, generation: childGeneration })
    }
  }

  stabilizeGenerations(memberIds, graph, generations)
  return generations
}

function stabilizeGenerations(
  memberIds: number[],
  graph: MemberGraph,
  generations: Map<number, number>
) {
  let changed = true
  let guard = 0

  while (changed && guard < memberIds.length + 20) {
    changed = false
    guard += 1

    for (const memberId of memberIds) {
      for (const spouseId of graph.spouseIds.get(memberId) || []) {
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

      const parents = graph.parentIds.get(memberId) || []
      const parentGens = parents
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

function createUnit(memberIds: number[], nodeMap: Map<number, TreeNode>): FamilyTreeUnit {
  const members = memberIds
    .map((memberId) => nodeMap.get(memberId))
    .filter((node): node is TreeNode => Boolean(node))
    .sort(compareMembers)

  return {
    memberIds: members.map((member) => member.memberId),
    members
  }
}

function buildUnitsForGeneration(
  memberIds: number[],
  generation: Map<number, number>,
  targetGeneration: number,
  spouseIds: Map<number, number[]>,
  nodeMap: Map<number, TreeNode>
): FamilyTreeUnit[] {
  const generationMembers = memberIds
    .filter((memberId) => generation.get(memberId) === targetGeneration)
    .sort((a, b) => birthSortValue(nodeMap.get(a)) - birthSortValue(nodeMap.get(b)))

  const assigned = new Set<number>()
  const units: FamilyTreeUnit[] = []

  for (const memberId of generationMembers) {
    if (assigned.has(memberId)) continue
    const unitMemberIds = [memberId]
    assigned.add(memberId)

    for (const spouseId of spouseIds.get(memberId) || []) {
      if (generation.get(spouseId) !== targetGeneration || assigned.has(spouseId)) continue
      unitMemberIds.push(spouseId)
      assigned.add(spouseId)
    }

    units.push(createUnit(unitMemberIds, nodeMap))
  }

  return units
}

function generationLabel(index: number): string {
  return `第 ${index + 1} 代`
}

export function buildFamilyTree(input: FamilyTreeBuildInput): BuiltFamilyTree {
  const memberIds = input.nodes.map((node) => node.memberId)
  if (memberIds.length === 0) {
    return {
      success: false,
      layers: [],
      orphanUnits: [],
      memberCount: 0,
      visibleMemberCount: 0,
      message: '暂无成员'
    }
  }

  try {
    const graph = buildGraph(input)
    const roots = findRoots(memberIds, graph.parentIds, graph.childrenIds, graph.spouseIds, graph.nodeMap)
    const generations = assignGenerations(memberIds, graph, roots)
    const maxGeneration = Math.max(...memberIds.map((memberId) => generations.get(memberId) ?? 0))

    const layers: FamilyTreeGeneration[] = []
    for (let generationIndex = 0; generationIndex <= maxGeneration; generationIndex += 1) {
      const units = buildUnitsForGeneration(
        memberIds,
        generations,
        generationIndex,
        graph.spouseIds,
        graph.nodeMap
      )
      if (units.length === 0) continue
      layers.push({
        generation: generationIndex,
        label: generationLabel(generationIndex),
        units
      })
    }

    const placed = new Set<number>()
    for (const layer of layers) {
      for (const unit of layer.units) {
        for (const memberId of unit.memberIds) placed.add(memberId)
      }
    }

    const orphanMemberIds = memberIds.filter((memberId) => !placed.has(memberId))
    const orphanUnits = orphanMemberIds.length
      ? buildUnitsForGeneration(
          orphanMemberIds,
          new Map(orphanMemberIds.map((memberId) => [memberId, 0])),
          0,
          graph.spouseIds,
          graph.nodeMap
        )
      : []

    const visibleMemberCount = placed.size + orphanMemberIds.length

    return {
      success: layers.length > 0 || orphanUnits.length > 0,
      layers,
      orphanUnits,
      memberCount: memberIds.length,
      visibleMemberCount,
      message: layers.length === 0 ? '家庭树暂时无法生成，可查看关系明细' : undefined
    }
  } catch {
    return {
      success: false,
      layers: [],
      orphanUnits: [],
      memberCount: memberIds.length,
      visibleMemberCount: 0,
      message: '家庭树暂时无法生成，可查看关系明细'
    }
  }
}
