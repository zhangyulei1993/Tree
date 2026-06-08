export interface UserRow {
  id: string
  maskedPhone: string
  nickname: string
  realName: string
  origin: string
  clientType: string
  phoneVerified: boolean
  status: string
  createdAt: string
  lastLoginAt: string
}

export interface FamilyRow {
  id: string
  name: string
  surname: string
  nativePlace: string
  regionText: string
  status: string
  searchable: boolean
  publicDisplayStatus: string
  founder: string
  memberCount: number
  graphVersion: number
  createdAt: string
}

export interface MemberNode {
  id: string
  name: string
  gender: string
  birthYear: number
  bindingStatus: string
  role: string
  status: string
  description: string
  children?: MemberNode[]
}

export interface AuditRow {
  id: string
  familyName: string
  applicant: string
  reason: string
  status: string
  createdAt: string
}

export interface VisitorMessageRow {
  id: string
  familyName: string
  visitorName: string
  publicContent: string
  contactSummary: string
  status: string
  createdAt: string
}

export interface FounderTransferRow {
  id: string
  familyName: string
  fromFounder: string
  toMember: string
  status: string
  createdAt: string
}

export interface DissolutionRow {
  id: string
  familyName: string
  requester: string
  reason: string
  status: string
  createdAt: string
}

export interface AdminUserRow {
  id: string
  username: string
  displayName: string
  role: string
  status: string
  lastLoginAt: string
}

export interface OperationLogRow {
  id: string
  actorType: string
  actor: string
  role: string
  module: string
  action: string
  targetType: string
  targetId: string
  result: string
  createdAt: string
  detail: Record<string, unknown>
}

export const dashboardStats = [
  { label: '用户总数', value: 1286, trend: '+12 本周' },
  { label: '家庭总数', value: 328, trend: '+6 本周' },
  { label: '公开家庭', value: 72, trend: '22%' },
  { label: '待审公开申请', value: 9, trend: '需要处理' },
  { label: '待审留言', value: 16, trend: '含 3 条高优先级' },
  { label: '待审转让', value: 2, trend: 'ROOT/SUPER' },
  { label: '待审解散', value: 1, trend: 'ROOT/SUPER' }
]

export const users: UserRow[] = [
  { id: 'user_001', maskedPhone: '138****0000', nickname: '松山', realName: '张一明', origin: 'PHONE', clientType: 'H5_WEB', phoneVerified: true, status: 'ACTIVE', createdAt: '2026-05-01 10:20', lastLoginAt: '2026-06-07 18:12' },
  { id: 'user_002', maskedPhone: '139****1111', nickname: '青禾', realName: '李清禾', origin: 'WECHAT_MINI', clientType: 'MINIAPP', phoneVerified: true, status: 'ACTIVE', createdAt: '2026-05-06 09:14', lastLoginAt: '2026-06-08 09:05' },
  { id: 'user_003', maskedPhone: '137****2222', nickname: '待绑定', realName: '王明', origin: 'WECHAT_MINI', clientType: 'MINIAPP', phoneVerified: false, status: 'PENDING', createdAt: '2026-05-18 16:44', lastLoginAt: '-' },
  { id: 'user_004', maskedPhone: '136****3333', nickname: '已禁用样例', realName: '赵安', origin: 'PHONE', clientType: 'H5_WEB', phoneVerified: true, status: 'DISABLED', createdAt: '2026-04-22 11:38', lastLoginAt: '2026-05-12 08:30' }
]

export const families: FamilyRow[] = [
  { id: 'family_001', name: '张氏家族', surname: '张', nativePlace: '山东济南', regionText: '山东省济南市', status: 'NORMAL', searchable: true, publicDisplayStatus: 'APPROVED', founder: '张一明', memberCount: 46, graphVersion: 14, createdAt: '2026-05-02 12:10' },
  { id: 'family_002', name: '李氏宗亲', surname: '李', nativePlace: '浙江杭州', regionText: '浙江省杭州市', status: 'NORMAL', searchable: true, publicDisplayStatus: 'PENDING', founder: '李清禾', memberCount: 28, graphVersion: 7, createdAt: '2026-05-10 15:20' },
  { id: 'family_003', name: '王氏家谱', surname: '王', nativePlace: '河南洛阳', regionText: '河南省洛阳市', status: 'DISSOLUTION_PENDING', searchable: false, publicDisplayStatus: 'PRIVATE', founder: '王明远', memberCount: 19, graphVersion: 5, createdAt: '2026-04-28 09:02' }
]

export const memberTree: MemberNode[] = [
  {
    id: 'member_001',
    name: '张远山',
    gender: '男',
    birthYear: 1948,
    bindingStatus: 'BOUND',
    role: 'FOUNDER',
    status: 'ACTIVE',
    description: '创建者成员节点',
    children: [
      {
        id: 'member_002',
        name: '张明远',
        gender: '男',
        birthYear: 1976,
        bindingStatus: 'BOUND',
        role: 'FAMILY_ADMIN',
        status: 'ACTIVE',
        description: '家庭管理员'
      },
      {
        id: 'member_003',
        name: '张明禾',
        gender: '女',
        birthYear: 1982,
        bindingStatus: 'INVITING',
        role: 'MEMBER',
        status: 'ACTIVE',
        description: '站内邀请中'
      },
      {
        id: 'member_004',
        name: '张小林',
        gender: '男',
        birthYear: 2008,
        bindingStatus: 'NOT_REQUIRED',
        role: 'MEMBER',
        status: 'ACTIVE',
        description: '未成年人样例，无需绑定账号'
      }
    ]
  }
]

export const publicApplications: AuditRow[] = [
  { id: 'pub_app_001', familyName: '李氏宗亲', applicant: '李清禾', reason: '希望展示公开家谱供宗亲查找', status: 'PENDING', createdAt: '2026-06-07 14:20' },
  { id: 'pub_app_002', familyName: '张氏家族', applicant: '张一明', reason: '完善公开展示资料后重新提交', status: 'APPROVED', createdAt: '2026-06-05 10:11' },
  { id: 'pub_app_003', familyName: '王氏家谱', applicant: '王明远', reason: '资料不足的申请样例', status: 'REJECTED', createdAt: '2026-06-01 09:36' }
]

export const visitorMessages: VisitorMessageRow[] = [
  { id: 'msg_001', familyName: '张氏家族', visitorName: '寻亲访客 A', publicContent: '希望联系同宗分支核对族谱。', contactSummary: '联系方式已加密存储，仅后台可见', status: 'PENDING', createdAt: '2026-06-08 09:12' },
  { id: 'msg_002', familyName: '张氏家族', visitorName: '访客 B', publicContent: '公开留言展示样例。', contactSummary: '公开列表不展示联系方式', status: 'APPROVED', createdAt: '2026-06-07 19:22' },
  { id: 'msg_003', familyName: '李氏宗亲', visitorName: '访客 C', publicContent: '内容不完整的拒绝样例。', contactSummary: '联系方式已隐藏', status: 'REJECTED', createdAt: '2026-06-06 16:40' }
]

export const founderTransfers: FounderTransferRow[] = [
  { id: 'transfer_001', familyName: '张氏家族', fromFounder: '张远山', toMember: '张明远', status: 'PENDING', createdAt: '2026-06-08 08:30' },
  { id: 'transfer_002', familyName: '李氏宗亲', fromFounder: '李清源', toMember: '李清禾', status: 'REJECTED', createdAt: '2026-06-02 17:12' }
]

export const dissolutionRequests: DissolutionRow[] = [
  { id: 'dissolution_001', familyName: '王氏家谱', requester: '王明远', reason: '家庭创建错误，申请解散', status: 'PENDING', createdAt: '2026-06-08 10:05' },
  { id: 'dissolution_002', familyName: '旧测试家庭', requester: '测试管理员', reason: '测试数据清理样例', status: 'APPROVED', createdAt: '2026-05-29 15:45' }
]

export const adminUsers: AdminUserRow[] = [
  { id: 'admin_001', username: 'root_admin_demo', displayName: 'Root 管理员', role: 'ROOT_ADMIN', status: 'ACTIVE', lastLoginAt: '2026-06-08 09:00' },
  { id: 'admin_002', username: 'super_admin_demo', displayName: '超级管理员', role: 'SUPER_ADMIN', status: 'ACTIVE', lastLoginAt: '2026-06-07 20:10' },
  { id: 'admin_003', username: 'platform_admin_demo', displayName: '平台管理员', role: 'PLATFORM_ADMIN', status: 'ACTIVE', lastLoginAt: '2026-06-08 08:42' }
]

export const operationLogs: OperationLogRow[] = [
  { id: 'log_001', actorType: 'ADMIN', actor: 'super_admin_demo', role: 'SUPER_ADMIN', module: 'PUBLIC_DISPLAY', action: 'APPROVE_PUBLIC_APPLICATION', targetType: 'family_public_application', targetId: 'pub_app_002', result: 'SUCCESS', createdAt: '2026-06-08 09:16', detail: { familyId: 'family_001', result: 'APPROVED', sensitiveFields: 'omitted' } },
  { id: 'log_002', actorType: 'ADMIN', actor: 'root_admin_demo', role: 'ROOT_ADMIN', module: 'DISSOLUTION', action: 'APPROVE_DISSOLUTION_REQUEST', targetType: 'family_dissolution_request', targetId: 'dissolution_002', result: 'SUCCESS', createdAt: '2026-06-07 11:20', detail: { familyId: 'family_003', statusAfter: 'DISSOLVED' } },
  { id: 'log_003', actorType: 'USER', actor: 'user_001', role: 'FOUNDER', module: 'FAMILY_MEMBER', action: 'CREATE_MEMBER', targetType: 'family_member', targetId: 'member_004', result: 'SUCCESS', createdAt: '2026-06-06 18:33', detail: { graphVersionAfter: 14 } }
]
