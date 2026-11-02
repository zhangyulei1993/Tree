import type { FamilyTreeResult, TreeEdge, TreeNode } from '@/types/api'

import { collectLineageChildren, buildMemberGraph, birthSortValue, isLineageMember, isSpouseMember, sortChildIds } from './graph'
import { spouseRelationLabel } from './spouseRelationLabel'

export type ExportLineageMode = 'SURNAME' | 'FAMILY'
export type ExportLineageRole = 'PATRILINEAL' | 'NATAL_FEMALE' | 'NATAL_FEMALE_CHILD' | 'FAMILY_MEMBER'

export interface ExportLineageNode {
  id: string
  member: TreeNode
  role: ExportLineageRole
  generation: number
  spouseLabels: string[]
  parentLabels: string[]
  children: ExportLineageNode[]
}

export interface ExportLineageGeneration {
  generation: number
  members: ExportLineageNode[]
}

export interface ExportLineageTree {
  roots: ExportLineageNode[]
  generations: ExportLineageGeneration[]
  unlocatedMembers: TreeNode[]
  visibleMemberCount: number
  hiddenDescendantCount: number
}

function spouseTitle(member: TreeNode, spouseId: number, spouseEdges: Map<number, TreeEdge>): string {
  const spouse = spouseEdges.get(spouseId)
  return spouseRelationLabel(spouse?.relationNoteType, member.gender === 'MALE' ? 'FEMALE' : member.gender === 'FEMALE' ? 'MALE' : null)
}

function spouseYearText(spouse: TreeNode): string {
  const birthYear = spouse.birthDate?.slice(0, 4)
  const deathYear = spouse.deathDate?.slice(0, 4)
  if (!deathYear) return ''
  return `（${birthYear ? `${birthYear}-${deathYear}` : `卒：${deathYear}`}）`
}

function spouseLabelsFor(member: TreeNode, spouseIds: number[], nodeMap: Map<number, TreeNode>, spouseEdges: Map<number, TreeEdge>): string[] {
  return spouseIds
    .map((id) => nodeMap.get(id))
    .filter((spouse): spouse is TreeNode => Boolean(spouse) && isSpouseMember(spouse))
    .map((spouse) => {
      return `${spouseTitle(member, spouse.memberId, spouseEdges)}：${spouse.displayName}${spouseYearText(spouse)}`
    })
}

function spouseEdgesFor(memberId: number, edges: FamilyTreeResult['edges']): Map<number, TreeEdge> {
  const result = new Map<number, TreeEdge>()
  for (const edge of edges) {
    if (edge.relationshipType !== 'SPOUSE') continue
    if (edge.fromMemberId === memberId) result.set(edge.toMemberId, edge)
    if (edge.toMemberId === memberId) result.set(edge.fromMemberId, edge)
  }
  return result
}

function resolvedParentIds(
  memberId: number,
  graph: ReturnType<typeof buildMemberGraph>,
  mode: ExportLineageMode,
  edges: FamilyTreeResult['edges']
): number[] {
  const directParentIds = [...new Set(graph.parentIds.get(memberId) || [])]
  if (mode !== 'FAMILY') return directParentIds

  const directParents = directParentIds
    .map((id) => graph.nodeMap.get(id))
    .filter((parent): parent is TreeNode => Boolean(parent))
  const hasFather = directParents.some((parent) => parent.gender === 'MALE')
  const hasMother = directParents.some((parent) => parent.gender === 'FEMALE')
  if (hasFather && hasMother) return directParentIds

  const inferredParentIds = new Set(directParentIds)
  for (const parent of directParents) {
    const expectedGender = parent.gender === 'MALE' ? 'FEMALE' : parent.gender === 'FEMALE' ? 'MALE' : ''
    if (!expectedGender) continue
    const candidates = (graph.spouseIds.get(parent.memberId) || [])
      .filter((spouseId) => !inferredParentIds.has(spouseId))
      // Only preserve the legacy single-spouse fallback. A named former/current
      // marriage never proves biological parentage; users select that explicitly.
      .filter((spouseId) => !spouseEdgesFor(parent.memberId, edges).get(spouseId)?.relationNoteType?.trim())
      .map((spouseId) => graph.nodeMap.get(spouseId))
      .filter((spouse): spouse is TreeNode => Boolean(spouse) && spouse.gender === expectedGender)
    if (candidates.length === 1) inferredParentIds.add(candidates[0]!.memberId)
  }
  return [...inferredParentIds]
}

function parentLabelsFor(parentIds: number[], nodeMap: Map<number, TreeNode>, mode: ExportLineageMode): string[] {
  return parentIds
    .map((id) => nodeMap.get(id))
    .filter((parent): parent is TreeNode => Boolean(parent) && (mode === 'FAMILY' || isLineageMember(parent)))
    .sort((left, right) => {
      if (left.gender === 'MALE' && right.gender !== 'MALE') return -1
      if (right.gender === 'MALE' && left.gender !== 'MALE') return 1
      return left.memberId - right.memberId
    })
    .map((parent) => `${parent.gender === 'MALE' ? '父' : parent.gender === 'FEMALE' ? '母' : '亲' }：${parent.displayName}`)
}

/**
 * 姓氏成员导出采用纯宗支视图：只保留本姓成员、代际与本姓父母归属。
 * 族外配偶既不作为树节点，也不以妻／夫注记出现。
 * 本家女性保留本人及其下一代，下一代不再继续展开，避免把外支误作本支。
 * 家庭成员导出则保留所有本家成员的完整后代关系，不因成员性别截断分支。
 */
export function buildExportLineageTree(tree: FamilyTreeResult, mode: ExportLineageMode = 'SURNAME'): ExportLineageTree {
  const graph = buildMemberGraph({ nodes: tree.nodes, edges: tree.edges, tree: tree.tree })
  const lineageIds = [...graph.nodeMap.keys()].filter((id) => isLineageMember(graph.nodeMap.get(id)))
  const lineageSet = new Set(lineageIds)
  const renderedIds = new Set<number>()
  const suppressedIds = new Set<number>()

  const roots = lineageIds
    .filter((id) => !(graph.parentIds.get(id) || []).some((parentId) => lineageSet.has(parentId)))
    .sort((a, b) => {
      const left = graph.nodeMap.get(a)!
      const right = graph.nodeMap.get(b)!
      if (left.gender === 'MALE' && right.gender !== 'MALE') return -1
      if (right.gender === 'MALE' && left.gender !== 'MALE') return 1
      const byBirth = birthSortValue(left) - birthSortValue(right)
      return byBirth !== 0 ? byBirth : a - b
    })

  const buildBranch = (
    memberId: number,
    role: ExportLineageRole,
    generation: number,
    ancestry: Set<number>
  ): ExportLineageNode => {
    const member = graph.nodeMap.get(memberId)!
    renderedIds.add(memberId)
    const nextAncestry = new Set(ancestry)
    nextAncestry.add(memberId)
    const childIds = mode === 'SURNAME' && role === 'NATAL_FEMALE_CHILD'
      ? []
      : sortChildIds(collectLineageChildren(graph, memberId), graph.nodeMap)
        .filter((childId) => !nextAncestry.has(childId))

    if (mode === 'SURNAME' && role === 'NATAL_FEMALE_CHILD') {
      for (const childId of collectLineageChildren(graph, memberId)) suppressedIds.add(childId)
    }

    return {
      id: `export-lineage-${memberId}`,
      member,
      role,
      generation,
      spouseLabels: mode === 'FAMILY'
        ? spouseLabelsFor(member, graph.spouseIds.get(memberId) || [], graph.nodeMap, spouseEdgesFor(memberId, tree.edges))
        : [],
      parentLabels: parentLabelsFor(resolvedParentIds(memberId, graph, mode, tree.edges), graph.nodeMap, mode),
      children: childIds.map((childId) => {
        const child = graph.nodeMap.get(childId)!
        const childRole: ExportLineageRole = mode === 'FAMILY'
          ? 'FAMILY_MEMBER'
          : role === 'NATAL_FEMALE'
            ? 'NATAL_FEMALE_CHILD'
            : child.gender === 'FEMALE'
              ? 'NATAL_FEMALE'
              : 'PATRILINEAL'
        return buildBranch(childId, childRole, generation + 1, nextAncestry)
      })
    }
  }

  const rootNodes = roots.map((id) => {
    const member = graph.nodeMap.get(id)!
    return buildBranch(id, mode === 'FAMILY' ? 'FAMILY_MEMBER' : member.gender === 'FEMALE' ? 'NATAL_FEMALE' : 'PATRILINEAL', 1, new Set())
  })
  const unlocatedMembers = lineageIds
    .filter((id) => !renderedIds.has(id) && !suppressedIds.has(id))
    .map((id) => graph.nodeMap.get(id)!)

  const byGeneration = new Map<number, ExportLineageNode[]>()
  const collect = (branch: ExportLineageNode) => {
    const members = byGeneration.get(branch.generation) || []
    members.push(branch)
    byGeneration.set(branch.generation, members)
    branch.children.forEach(collect)
  }
  rootNodes.forEach(collect)

  return {
    roots: rootNodes,
    generations: [...byGeneration.entries()]
      .sort(([left], [right]) => left - right)
      .map(([generation, members]) => ({ generation, members })),
    unlocatedMembers,
    visibleMemberCount: renderedIds.size,
    hiddenDescendantCount: [...suppressedIds].filter((id) => !renderedIds.has(id)).length
  }
}
