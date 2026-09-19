import { isRealApiMode, request } from '@/api/client'
import type { FamilyTreeResult, ShowcaseFamilyItem, ShowcaseFamilyTree, TreeItem } from '@/types/api'

export async function listPublishedShowcaseFamilies(): Promise<ShowcaseFamilyItem[]> {
  if (!isRealApiMode) return []
  return request<ShowcaseFamilyItem[]>('/showcase-families', { public: true })
}

export async function getPublishedShowcaseFamilyTree(showcaseFamilyId: number): Promise<FamilyTreeResult> {
  const source = await request<ShowcaseFamilyTree>(`/showcase-families/${showcaseFamilyId}/tree`, { public: true })
  const treeByMember = new Map<number, TreeItem>()
  for (const member of source.nodes) {
    treeByMember.set(member.memberId, {
      memberId: member.memberId,
      parentIds: [],
      childrenIds: [],
      spouseIds: []
    })
  }
  for (const edge of source.edges) {
    const from = treeByMember.get(edge.fromMemberId)
    const to = treeByMember.get(edge.toMemberId)
    if (!from || !to) continue
    if (edge.relationshipType === 'SPOUSE') {
      from.spouseIds.push(to.memberId)
      to.spouseIds.push(from.memberId)
      continue
    }
    from.childrenIds.push(to.memberId)
    to.parentIds.push(from.memberId)
  }

  return {
    familyId: source.showcaseFamilyId,
    treeMode: 'SHOWCASE_TREE',
    graphVersion: 1,
    nodes: source.nodes.map((member) => ({
      memberId: member.memberId,
      displayName: member.displayName,
      surname: member.displayName.slice(0, 1),
      gender: member.gender,
      memberType: member.memberType,
      birthDate: member.birthYear ? `${member.birthYear}-01-01` : null,
      isLiving: null,
      userBindingState: 'NOT_REQUIRED',
      canExpand: false
    })),
    edges: source.edges,
    tree: [...treeByMember.values()]
  }
}
