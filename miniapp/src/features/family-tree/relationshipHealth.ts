import type { FamilyTreeResult, TreeNode } from '@/types/api'

export interface RelationshipHealthResult {
  hintsByMemberId: Record<number, string[]>
  membersWithChildrenNoSpouse: TreeNode[]
  membersWithBirthOrderConflict: TreeNode[]
}

/** 提示可补全的关系信息，不将单亲等真实情况判定为错误。 */
export function inspectRelationshipHealth(tree: FamilyTreeResult | null | undefined): RelationshipHealthResult {
  const hintsByMemberId: Record<number, string[]> = {}
  if (!tree) return { hintsByMemberId, membersWithChildrenNoSpouse: [], membersWithBirthOrderConflict: [] }

  const childCount = new Map<number, number>()
  const spouseIds = new Map<number, Set<number>>()
  const membersByID = new Map(tree.nodes.map((node) => [node.memberId, node]))
  const birthOrderConflictIds = new Set<number>()
  for (const edge of tree.edges) {
    if (edge.relationshipType === 'PARENT_CHILD') {
      childCount.set(edge.fromMemberId, (childCount.get(edge.fromMemberId) || 0) + 1)
      const parent = membersByID.get(edge.fromMemberId)
      const child = membersByID.get(edge.toMemberId)
      const parentYear = birthYearOf(parent)
      const childYear = birthYearOf(child)
      if (parentYear != null && childYear != null && parentYear >= childYear) {
        birthOrderConflictIds.add(edge.fromMemberId)
        birthOrderConflictIds.add(edge.toMemberId)
      }
      continue
    }
    if (edge.relationshipType === 'SPOUSE') {
      const left = spouseIds.get(edge.fromMemberId) || new Set<number>()
      left.add(edge.toMemberId)
      spouseIds.set(edge.fromMemberId, left)
      const right = spouseIds.get(edge.toMemberId) || new Set<number>()
      right.add(edge.fromMemberId)
      spouseIds.set(edge.toMemberId, right)
    }
  }

  const membersWithChildrenNoSpouse = tree.nodes.filter((node) =>
    (childCount.get(node.memberId) || 0) > 0 && (spouseIds.get(node.memberId)?.size || 0) === 0
  )
  for (const member of membersWithChildrenNoSpouse) addHint(hintsByMemberId, member.memberId, '有子女，待补配偶')
  const membersWithBirthOrderConflict = tree.nodes.filter((node) => birthOrderConflictIds.has(node.memberId))
  for (const member of membersWithBirthOrderConflict) addHint(hintsByMemberId, member.memberId, '亲子出生年份待核对')
  return { hintsByMemberId, membersWithChildrenNoSpouse, membersWithBirthOrderConflict }
}

function birthYearOf(node: TreeNode | undefined): number | null {
  const match = /^(\d{4})/.exec(node?.birthDate || '')
  return match ? Number(match[1]) : null
}

function addHint(hintsByMemberId: Record<number, string[]>, memberId: number, hint: string) {
  const hints = hintsByMemberId[memberId] || []
  if (!hints.includes(hint)) hints.push(hint)
  hintsByMemberId[memberId] = hints
}
