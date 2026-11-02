import { isRealApiMode, request } from '@/api/client'
import type { FamilyTreeResult, ShowcaseFamilyItem, ShowcaseFamilyTree, TreeItem } from '@/types/api'

const mockShowcaseFamilies: ShowcaseFamilyItem[] = [
  {
    id: 40,
    displayName: '汉高祖刘邦五代家系',
    regionText: '沛县 · 西汉',
    summary: '历史示例 · 根据公开史料整理。以刘邦为起点，展示五代内的主要支系与配偶关系。',
    coverImageUrl: '',
    treeImageUrl: '',
    memberCount: 10,
    relationshipCount: 9,
    rotationPinned: true
  }
]

const mockShowcaseTrees: Record<number, ShowcaseFamilyTree> = {
  40: {
    showcaseFamilyId: 40,
    nodes: [
      { memberId: 1, displayName: '刘邦', identityText: '汉高祖', memberType: 'LINEAGE_MEMBER', gender: 'MALE', birthYear: -256 },
      { memberId: 2, displayName: '吕雉', identityText: '汉高后', memberType: 'SPOUSE', gender: 'FEMALE', birthYear: -241 },
      { memberId: 3, displayName: '刘盈', identityText: '汉惠帝', memberType: 'LINEAGE_MEMBER', gender: 'MALE', birthYear: -210 },
      { memberId: 4, displayName: '刘恒', identityText: '汉文帝', memberType: 'LINEAGE_MEMBER', gender: 'MALE', birthYear: -203 },
      { memberId: 5, displayName: '窦皇后', identityText: '汉文帝皇后', memberType: 'SPOUSE', gender: 'FEMALE', birthYear: null },
      { memberId: 6, displayName: '刘启', identityText: '汉景帝', memberType: 'LINEAGE_MEMBER', gender: 'MALE', birthYear: -188 },
      { memberId: 7, displayName: '王娡', identityText: '汉武帝生母', memberType: 'SPOUSE', gender: 'FEMALE', birthYear: null },
      { memberId: 8, displayName: '刘彻', identityText: '汉武帝', memberType: 'LINEAGE_MEMBER', gender: 'MALE', birthYear: -156 },
      { memberId: 9, displayName: '卫子夫', identityText: '汉武帝皇后', memberType: 'SPOUSE', gender: 'FEMALE', birthYear: null },
      { memberId: 10, displayName: '刘据', identityText: '戾太子', memberType: 'LINEAGE_MEMBER', gender: 'MALE', birthYear: -128 }
    ],
    edges: [
      { relationshipId: 1, fromMemberId: 1, toMemberId: 2, relationshipType: 'SPOUSE' },
      { relationshipId: 2, fromMemberId: 1, toMemberId: 3, relationshipType: 'PARENT_CHILD' },
      { relationshipId: 3, fromMemberId: 1, toMemberId: 4, relationshipType: 'PARENT_CHILD' },
      { relationshipId: 4, fromMemberId: 4, toMemberId: 5, relationshipType: 'SPOUSE' },
      { relationshipId: 5, fromMemberId: 4, toMemberId: 6, relationshipType: 'PARENT_CHILD' },
      { relationshipId: 6, fromMemberId: 6, toMemberId: 7, relationshipType: 'SPOUSE' },
      { relationshipId: 7, fromMemberId: 6, toMemberId: 8, relationshipType: 'PARENT_CHILD' },
      { relationshipId: 8, fromMemberId: 8, toMemberId: 9, relationshipType: 'SPOUSE' },
      { relationshipId: 9, fromMemberId: 8, toMemberId: 10, relationshipType: 'PARENT_CHILD' }
    ]
  }
}

export async function listPublishedShowcaseFamilies(): Promise<ShowcaseFamilyItem[]> {
  if (!isRealApiMode) return mockShowcaseFamilies
  return request<ShowcaseFamilyItem[]>('/showcase-families', { public: true })
}

export async function getPublishedShowcaseFamilyTree(showcaseFamilyId: number): Promise<FamilyTreeResult> {
  const source = !isRealApiMode
    ? mockShowcaseTrees[showcaseFamilyId] || mockShowcaseTrees[40]
    : await request<ShowcaseFamilyTree>(`/showcase-families/${showcaseFamilyId}/tree`, { public: true })
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
      identityText: member.identityText,
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
