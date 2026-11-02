export type UserState = 'guest' | 'wechatLoggedInPendingPhone' | 'phoneBoundActive'

export interface PublicFamily {
  id: string
  name: string
  surname: string
  nativePlace: string
  regionText: string
  description: string
  publicContact: string
  publicStatus: 'APPROVED' | 'PRIVATE'
  founderName: string
}

export interface TreeNode {
  memberId: string
  name: string
  gender: string
  birthText: string
  deathText?: string
  description: string
  parentIds: string[]
  spouseIds: string[]
  childrenIds: string[]
}

export interface TreeEdge {
  sourceMemberId: string
  targetMemberId: string
  relationshipType: 'PARENT_CHILD' | 'SPOUSE'
}

export const currentUser = {
  id: 'user_001',
  nickname: '松山',
  maskedPhone: '138****0000',
  status: 'ACTIVE',
  phoneVerified: true,
  token: 'example_user_token'
}

export const publicFamilies: PublicFamily[] = [
  {
    id: 'family_001',
    name: '张氏家族',
    surname: '张',
    nativePlace: '山东济南',
    regionText: '山东省济南市',
    description: '记录张氏一支的家族成员、亲缘关系与公开留言，供宗亲查找和联系。',
    publicContact: '张先生 / 138****0000',
    publicStatus: 'APPROVED',
    founderName: '张远山'
  },
  {
    id: 'family_002',
    name: '李氏宗亲',
    surname: '李',
    nativePlace: '浙江杭州',
    regionText: '浙江省杭州市',
    description: '面向同宗成员的公开展示样例，当前公开联系方式为空。',
    publicContact: '',
    publicStatus: 'APPROVED',
    founderName: '李清禾'
  },
  {
    id: 'family_003',
    name: '王氏家谱',
    surname: '王',
    nativePlace: '河南洛阳',
    regionText: '河南省洛阳市',
    description: '未公开家庭样例，仅展示申请加入入口。',
    publicContact: '',
    publicStatus: 'PRIVATE',
    founderName: '王明远'
  }
]

export const treeNodes: TreeNode[] = [
  { memberId: 'member_001', name: '张远山', gender: '男', birthText: '1948', description: '家族公开树顶点成员', parentIds: [], spouseIds: ['member_002'], childrenIds: ['member_003', 'member_004'] },
  { memberId: 'member_002', name: '林婉清', gender: '女', birthText: '1951', description: '配偶节点', parentIds: [], spouseIds: ['member_001'], childrenIds: ['member_003', 'member_004'] },
  { memberId: 'member_003', name: '张明远', gender: '男', birthText: '1976', description: '第二代成员，已公开展示', parentIds: ['member_001', 'member_002'], spouseIds: ['member_005'], childrenIds: ['member_006'] },
  { memberId: 'member_004', name: '张明禾', gender: '女', birthText: '1982', description: '第二代成员', parentIds: ['member_001', 'member_002'], spouseIds: [], childrenIds: [] },
  { memberId: 'member_005', name: '陈若兰', gender: '女', birthText: '1978', description: '配偶节点', parentIds: [], spouseIds: ['member_003'], childrenIds: ['member_006'] },
  { memberId: 'member_006', name: '张小林', gender: '男', birthText: '2008', description: '第三代成员', parentIds: ['member_003', 'member_005'], spouseIds: [], childrenIds: [] }
]

export const treeEdges: TreeEdge[] = [
  { sourceMemberId: 'member_001', targetMemberId: 'member_002', relationshipType: 'SPOUSE' },
  { sourceMemberId: 'member_001', targetMemberId: 'member_003', relationshipType: 'PARENT_CHILD' },
  { sourceMemberId: 'member_002', targetMemberId: 'member_003', relationshipType: 'PARENT_CHILD' },
  { sourceMemberId: 'member_001', targetMemberId: 'member_004', relationshipType: 'PARENT_CHILD' },
  { sourceMemberId: 'member_002', targetMemberId: 'member_004', relationshipType: 'PARENT_CHILD' },
  { sourceMemberId: 'member_003', targetMemberId: 'member_005', relationshipType: 'SPOUSE' },
  { sourceMemberId: 'member_003', targetMemberId: 'member_006', relationshipType: 'PARENT_CHILD' },
  { sourceMemberId: 'member_005', targetMemberId: 'member_006', relationshipType: 'PARENT_CHILD' }
]

export const visitorMessages = [
  { id: 'message_001', familyId: 'family_001', visitorName: '寻亲访客 A', content: '希望联系同宗分支核对族谱。', createdAt: '2026-06-08' },
  { id: 'message_002', familyId: 'family_001', visitorName: '宗亲 B', content: '公开留言样例，联系方式不展示。', createdAt: '2026-06-07' }
]

export const invitation = {
  id: 'invitation_001',
  token: 'mock_invite_token',
  familyName: '张氏家族',
  surname: '张',
  regionText: '山东省济南市',
  memberName: '张明禾',
  gender: '女',
  birthText: '1982',
  inviterName: '张明远',
  expiresAt: '2026-06-15 12:00',
  status: 'PENDING'
}

export const myFamilies = [
  { id: 'family_001', name: '张氏家族', role: 'FOUNDER', memberName: '张远山', status: 'NORMAL' },
  { id: 'family_002', name: '李氏宗亲', role: 'MEMBER', memberName: '李清禾', status: 'NORMAL' }
]

export const joinRequests = [
  { id: 'join_001', familyName: '李氏宗亲', status: 'PENDING', reason: '希望加入同宗家庭核对族谱。' }
]
