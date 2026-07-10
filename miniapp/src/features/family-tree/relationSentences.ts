import type { FamilyTreeResult } from '@/types/api'

function nodeName(memberId: number, nodeNameMap: Map<number, string>): string {
  return nodeNameMap.get(memberId) || '未知成员'
}

function parentLinkText(type?: string | null): string {
  switch (type) {
    case 'PRIMARY':
      return '亲生'
    case 'STEP':
      return '继亲'
    case 'ADOPTIVE':
      return '收养'
    case 'SUCCESSION':
      return '过继'
    case 'NOTE_ONLY':
      return '备注'
    case 'OTHER':
      return '其他'
    default:
      return '父母子女'
  }
}

export function buildRelationSentences(tree: Pick<FamilyTreeResult, 'edges' | 'nodes'>): string[] {
  const visibleEdges = tree.edges.filter((edge) => String(edge.relationshipType) !== 'SIBLING')
  const nodeNameMap = new Map(tree.nodes.map((node) => [node.memberId, node.displayName]))
  const nodeGenderMap = new Map(tree.nodes.map((node) => [node.memberId, node.gender]))

  return visibleEdges.map((edge) => {
    const fromName = nodeName(edge.fromMemberId, nodeNameMap)
    const toName = nodeName(edge.toMemberId, nodeNameMap)

    if (edge.relationshipType === 'SPOUSE') {
      let sentence = `${fromName} 与 ${toName} 是配偶`
      if (edge.relationNote) sentence += `，${edge.relationNote}`
      return sentence
    }

    const fromGender = nodeGenderMap.get(edge.fromMemberId)
    let parentRole = '父母'
    if (fromGender === 'MALE') parentRole = '父亲'
    else if (fromGender === 'FEMALE') parentRole = '母亲'

    let sentence = `${fromName} 是 ${toName} 的${parentRole}`
    const linkLabel = parentLinkText(edge.parentLinkType)
    if (linkLabel && linkLabel !== '父母子女') {
      sentence += `（${linkLabel}）`
    }
    if (edge.relationNote) sentence += `，${edge.relationNote}`
    return sentence
  })
}

export function countVisibleEdges(tree: Pick<FamilyTreeResult, 'edges'>): number {
  return tree.edges.filter((edge) => String(edge.relationshipType) !== 'SIBLING').length
}

export function treeViewLabel(mode: string): string {
  switch (mode) {
    case 'LIST_TREE':
      return '成员列表'
    case 'GRAPH_TREE':
      return '关系图'
    default:
      return '家庭树'
  }
}
