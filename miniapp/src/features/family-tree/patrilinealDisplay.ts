import { buildMemberGraph, isLineageMember, isSpouseMember } from './graph'
import type { FamilyTreeBuildInput } from './types'

export type PatrilinealDisplayRole = 'PATRILINEAL' | 'NATAL_FEMALE' | 'AFFINAL' | 'EXTERNAL' | 'PENDING'

/**
 * 展示层的父系标记，不改动成员类型或关系数据：本支仅指父系男性后代。
 */
export function buildPatrilinealDisplayRoles(input: FamilyTreeBuildInput): Record<number, PatrilinealDisplayRole> {
  const graph = buildMemberGraph(input)
  const roles: Record<number, PatrilinealDisplayRole> = {}

  for (const node of input.nodes) {
    if (isSpouseMember(node)) roles[node.memberId] = 'AFFINAL'
    else if (!isLineageMember(node)) roles[node.memberId] = 'EXTERNAL'
    else if (node.gender === 'FEMALE') roles[node.memberId] = 'NATAL_FEMALE'
    else if (node.gender !== 'MALE') roles[node.memberId] = 'PENDING'
    else roles[node.memberId] = 'EXTERNAL'
  }

  const roots = input.nodes.filter((node) =>
    isLineageMember(node) &&
    node.gender === 'MALE' &&
    !(graph.parentIds.get(node.memberId) || []).some((parentId) => isLineageMember(graph.nodeMap.get(parentId)))
  )

  const visited = new Set<number>()
  const walkPatriline = (memberId: number) => {
    if (visited.has(memberId)) return
    visited.add(memberId)
    roles[memberId] = 'PATRILINEAL'

    for (const childId of graph.childrenIds.get(memberId) || []) {
      const child = graph.nodeMap.get(childId)
      if (!child || !isLineageMember(child)) continue
      if (child.gender === 'MALE') {
        walkPatriline(childId)
      } else if (child.gender === 'FEMALE') {
        roles[childId] = 'NATAL_FEMALE'
      } else {
        roles[childId] = 'PENDING'
      }
    }
  }

  roots.forEach((node) => walkPatriline(node.memberId))
  return roles
}
