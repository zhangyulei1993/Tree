/** 配偶关系的展示身份。保存在 relationship.relationNoteType，供树和导出共用。 */
import type { FamilyTreeResult, TreeEdge, TreeNode } from '@/types/api'

export function spouseRelationLabel(noteType?: string | null, fallbackGender?: string | null): string {
  switch (noteType) {
    case 'EX_WIFE':
      return '前妻'
    case 'EX_HUSBAND':
      return '前夫'
    case 'SECOND_WIFE':
      return '现任妻'
    case 'SECOND_HUSBAND':
      return '现任夫'
    default:
      return fallbackGender === 'MALE' ? '夫' : fallbackGender === 'FEMALE' ? '妻' : '配偶'
  }
}

function spouseTargetForEdge(edge: TreeEdge, nodeMap: Map<number, TreeNode>): TreeNode | null {
  const from = nodeMap.get(edge.fromMemberId)
  const to = nodeMap.get(edge.toMemberId)
  if (!from || !to) return null

  // The generated spouse node is the most reliable target. For imported data
  // where both endpoints are lineage nodes, the note type's gender is decisive.
  if (from.memberType === 'SPOUSE') return from
  if (to.memberType === 'SPOUSE') return to
  const expectedGender = edge.relationNoteType === 'EX_WIFE' || edge.relationNoteType === 'SECOND_WIFE'
    ? 'FEMALE'
    : edge.relationNoteType === 'EX_HUSBAND' || edge.relationNoteType === 'SECOND_HUSBAND'
      ? 'MALE'
      : ''
  if (expectedGender && from.gender === expectedGender && to.gender !== expectedGender) return from
  if (expectedGender && to.gender === expectedGender && from.gender !== expectedGender) return to
  return null
}

/** Maps each explicitly designated former/current spouse to its visual tree label. */
export function buildSpouseRelationLabelMap(tree?: Pick<FamilyTreeResult, 'nodes' | 'edges'> | null): Record<number, string> {
  const labels: Record<number, string> = {}
  const nodeMap = new Map((tree?.nodes || []).map((node) => [node.memberId, node]))
  for (const edge of tree?.edges || []) {
    if (edge.relationshipType !== 'SPOUSE' || !edge.relationNoteType?.trim()) continue
    const target = spouseTargetForEdge(edge, nodeMap)
    if (target) labels[target.memberId] = spouseRelationLabel(edge.relationNoteType, target.gender)
  }
  return labels
}

export function spouseDisplayOptions(spouseGender?: string | null) {
  if (spouseGender === 'FEMALE') {
    return [
      { label: '妻（普通配偶）', noteType: '' },
      { label: '前妻', noteType: 'EX_WIFE' },
      { label: '现任妻', noteType: 'SECOND_WIFE' }
    ]
  }
  if (spouseGender === 'MALE') {
    return [
      { label: '夫（普通配偶）', noteType: '' },
      { label: '前夫', noteType: 'EX_HUSBAND' },
      { label: '现任夫', noteType: 'SECOND_HUSBAND' }
    ]
  }
  return [{ label: '配偶（普通）', noteType: '' }]
}
