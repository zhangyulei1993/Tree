export type UserState = 'guest' | 'wechatProfileIncomplete' | 'wechatActive'

export const mockUser = {
  nickname: '松山',
  maskedPhone: '138****0000',
  status: 'ACTIVE',
  token: 'example_user_token'
}

export const families = [
  {
    id: 'family_001',
    name: '张氏家族',
    surname: '张',
    nativePlace: '山东济南',
    regionText: '山东省济南市',
    contact: '张先生 / 138****0000',
    description: '记录张氏一支的家族成员与亲缘关系，供宗亲查找和联系。',
    memberCount: 6,
    graphVersion: 14,
    status: 'NORMAL'
  },
  {
    id: 'family_002',
    name: '李氏宗亲',
    surname: '李',
    nativePlace: '浙江杭州',
    regionText: '浙江省杭州市',
    contact: '',
    description: '未设置公开联系方式的样例。',
    memberCount: 4,
    graphVersion: 8,
    status: 'NORMAL'
  }
]

export const treeNodes = [
  { memberId: 'member_001', name: '张远山', gender: '男', birthText: '1948', parentIds: [], spouseIds: ['member_002'], childrenIds: ['member_003', 'member_004'], description: '顶点成员' },
  { memberId: 'member_002', name: '林婉清', gender: '女', birthText: '1951', parentIds: [], spouseIds: ['member_001'], childrenIds: ['member_003', 'member_004'], description: '配偶节点' },
  { memberId: 'member_003', name: '张明远', gender: '男', birthText: '1976', parentIds: ['member_001', 'member_002'], spouseIds: [], childrenIds: [], description: '第二代成员' },
  { memberId: 'member_004', name: '张明禾', gender: '女', birthText: '1982', parentIds: ['member_001', 'member_002'], spouseIds: [], childrenIds: [], description: '第二代成员' }
]

export const invite = {
  token: 'mock_invite_token',
  familyName: '张氏家族',
  surname: '张',
  regionText: '山东省济南市',
  memberName: '张明禾',
  inviterName: '张明远',
  expiresAt: '2026-06-15 12:00'
}

export const myFamilies = [
  { id: 'family_001', name: '张氏家族', role: 'FOUNDER', memberName: '张远山', status: 'NORMAL' },
  { id: 'family_002', name: '李氏宗亲', role: 'MEMBER', memberName: '李清禾', status: 'NORMAL' }
]

export const joinRequests = [
  { id: 'join_001', familyName: '李氏宗亲', status: 'PENDING', reason: '希望加入同宗家庭核对族谱。' }
]
