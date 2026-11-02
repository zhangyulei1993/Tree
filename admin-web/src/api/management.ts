import { apiClient, apiMode, unwrapData } from '@/api/client'
import {
  adminUsers as mockAdminUsers,
  dissolutionRequests as mockDissolutionRequests,
  families as mockFamilies,
  founderTransfers as mockFounderTransfers,
  memberTree as mockMemberTree,
  operationLogs as mockOperationLogs,
  publicApplications as mockPublicApplications,
  users as mockUsers
} from '@/mock/data'
import type {
  AdminRole,
  ApiResponse,
  DashboardStats,
  DissolutionRequest,
  FounderTransferRequest,
  ManagedAdmin,
  ManagedFamily,
  AdminFamilyTree,
  ManagedFamilyMember,
  ManagedUser,
  ManagedUserDetail,
  OperationLogRecord,
  PageResult,
  PublicApplication,
  RecoveryItem
} from '@/types/api'

export interface PageQuery { keyword?: string; status?: string; role?: AdminRole | ''; operatorType?: string; result?: string; familyId?: number | string; startDate?: string; endDate?: string; page?: number; pageSize?: number }

function paginate<T>(items: T[], page = 1, pageSize = 20): PageResult<T> {
  const start = (page - 1) * pageSize
  return {
    items: items.slice(start, start + pageSize),
    page,
    pageSize,
    total: items.length
  }
}

function mapManagedUsers(): ManagedUser[] {
  return mockUsers.map((item, index) => {
    const hasWechatLogin = item.clientType.includes('WECHAT') || item.origin.includes('WECHAT')
    return {
      id: index + 1,
      phone: item.maskedPhone,
      nickname: item.nickname,
      realName: item.realName,
      accountOrigin: item.origin,
      registerClient: item.clientType,
      phoneVerified: item.phoneVerified,
      hasWechatLogin,
      trustTier: item.phoneVerified ? 'PHONE_BOUND' : 'WECHAT_ONLY',
      loginMethod: hasWechatLogin ? '微信登录' : '未设置',
      status: item.status === 'PENDING' ? 'PENDING_BIND' : item.status,
      lastLoginAt: item.lastLoginAt === '-' ? null : item.lastLoginAt,
      createdAt: item.createdAt
    }
  })
}

function mapManagedFamilies(): ManagedFamily[] {
  return mockFamilies.map((item, index) => ({
    id: index + 1,
    familyName: item.name,
    familySurname: item.surname,
    nativePlace: item.nativePlace,
    regionText: item.regionText,
    description: `${item.name}的公开与私有资料管理示例。`,
    status: item.status,
    searchable: item.searchable,
    publicDisplayStatus: item.publicDisplayStatus,
    currentFounderMemberId: 1,
    graphVersion: item.graphVersion,
    memberCount: item.memberCount,
    createdAt: item.createdAt
  }))
}

function mapFamilyMembers(familyId: number | string): ManagedFamilyMember[] {
  const targetFamilyId = Number(familyId) || 1
  const members = mockMemberTree.flatMap((node) => [node, ...(node.children || [])])
  return members.map((item, index) => ({
    id: index + 1,
    familyId: targetFamilyId,
    displayName: item.name,
    gender: item.gender === '男' ? 'MALE' : item.gender === '女' ? 'FEMALE' : 'UNKNOWN',
    status: item.status,
    isLiving: true,
    userBindingPolicy: item.bindingStatus,
    boundUserId: item.bindingStatus === 'BOUND' ? index + 1 : null,
    familyRole: item.role === 'FOUNDER' ? 'FAMILY_FOUNDER' : item.role
  }))
}

function mapAdmins(): ManagedAdmin[] {
  return mockAdminUsers.map((item, index) => ({
    id: index + 1,
    username: item.username,
    displayName: item.displayName,
    role: item.role as AdminRole,
    status: item.status,
    lastLoginAt: item.lastLoginAt,
    createdAt: '2026-05-01 09:00'
  }))
}

function mapOperationLogs(): OperationLogRecord[] {
  return mockOperationLogs.map((item, index) => ({
    id: index + 1,
    operatorType: item.actorType,
    operatorRole: item.role,
    module: item.module,
    action: item.action,
    targetType: item.targetType,
    targetId: Number(item.targetId.replace(/\D/g, '')) || index + 1,
    result: item.result,
    detailJson: item.detail,
    createdAt: item.createdAt
  }))
}

export function listMockPublicApplications(query: { status?: string; page?: number; pageSize?: number }): PageResult<PublicApplication> {
  const items = mockPublicApplications
    .map((item, index) => ({
      applicationId: index + 1,
      familyId: index + 1,
      familyName: item.familyName,
      familyPublicDisplayStatus: item.status === 'APPROVED' ? 'APPROVED' : 'PRIVATE',
      applicantUserId: index + 1,
      status: item.status,
      reason: item.reason,
      createdAt: item.createdAt,
      updatedAt: item.createdAt
    }))
    .filter((item) => !query.status || item.status === query.status)
  return paginate(items, query.page, query.pageSize)
}

export function listMockFounderTransferRequests(query: { status?: string; page?: number; pageSize?: number }): PageResult<FounderTransferRequest> {
  const items: FounderTransferRequest[] = mockFounderTransfers
    .map((item, index) => ({
      requestId: index + 1,
      familyId: (index % mockFamilies.length) + 1,
      familyName: item.familyName,
      fromMemberId: index + 1,
      fromUserId: index + 1,
      toMemberId: index + 11,
      toUserId: index + 21,
      requestStatus: item.status,
      requestReason: `由 ${item.fromFounder} 转让给 ${item.toMember}`,
      createdAt: item.createdAt,
      updatedAt: item.createdAt
    }))
    .filter((item) => !query.status || item.requestStatus === query.status)
  return paginate(items, query.page, query.pageSize)
}

export function listMockDissolutionRequests(query: { status?: string; page?: number; pageSize?: number }): PageResult<DissolutionRequest> {
  const items: DissolutionRequest[] = mockDissolutionRequests
    .map((item, index) => ({
      requestId: index + 1,
      familyId: (index % mockFamilies.length) + 1,
      familyName: item.familyName,
      familyStatus: item.status === 'APPROVED' ? 'DISSOLVED' : 'DISSOLUTION_PENDING',
      requesterMemberId: index + 1,
      requesterUserId: index + 1,
      requestStatus: item.status,
      requestReason: item.reason,
      createdAt: item.createdAt,
      updatedAt: item.createdAt
    }))
    .filter((item) => !query.status || item.requestStatus === query.status)
  return paginate(items, query.page, query.pageSize)
}

export async function getDashboard() {
  if (apiMode === 'mock') {
    return {
      users: mockUsers.length,
      families: mockFamilies.length,
      pendingPublicApplications: mockPublicApplications.filter((item) => item.status === 'PENDING').length,
      pendingFounderTransfers: mockFounderTransfers.filter((item) => item.status === 'PENDING').length,
      pendingDissolutions: mockDissolutionRequests.filter((item) => item.status === 'PENDING').length,
      paidCapacityOrders: 0,
      activeCapacityRevenueFen: 0,
      pendingCapacityRefunds: 0,
      reviewCapacityRefunds: 0,
      failedCapacityRefunds: 0,
      pendingOperations: 0,
      newUsersLast7Days: 0,
      newFamiliesLast7Days: 0,
      acceptedInvitationsLast7Days: 0
    } satisfies DashboardStats
  }
  return unwrapData(await apiClient.get<ApiResponse<DashboardStats>>('/admin/dashboard'))
}

export async function listUsers(params: PageQuery) {
  if (apiMode === 'mock') {
    const items = mapManagedUsers().filter((item) => {
      const hitKeyword = !params.keyword || [item.id, item.phone, item.nickname, item.realName].some((value) => String(value || '').includes(params.keyword!))
      const hitStatus = !params.status || item.status === params.status
      return hitKeyword && hitStatus
    })
    return paginate(items, params.page, params.pageSize)
  }
  return unwrapData(await apiClient.get<ApiResponse<PageResult<ManagedUser>>>('/admin/users', { params }))
}

export async function getUser(userId: number | string) {
  if (apiMode === 'mock') {
    const user = mapManagedUsers().find((item) => String(item.id) === String(userId)) || mapManagedUsers()[0]
    return {
      user,
      families: mapManagedFamilies().slice(0, 2).map((family, index) => ({
        familyId: family.id,
        familyName: family.familyName,
        memberId: index + 1,
        memberName: mapFamilyMembers(family.id)[index]?.displayName || '成员节点',
        familyRole: index === 0 ? 'FAMILY_FOUNDER' : 'MEMBER'
      }))
    } satisfies ManagedUserDetail
  }
  return unwrapData(await apiClient.get<ApiResponse<ManagedUserDetail>>(`/admin/users/${userId}`))
}

export async function listFamilies(params: PageQuery) {
  if (apiMode === 'mock') {
    const items = mapManagedFamilies().filter((item) => {
      const hitKeyword = !params.keyword || [item.id, item.familyName, item.familySurname, item.regionText].some((value) => String(value || '').includes(params.keyword!))
      const hitStatus = !params.status || item.status === params.status
      return hitKeyword && hitStatus
    })
    return paginate(items, params.page, params.pageSize)
  }
  return unwrapData(await apiClient.get<ApiResponse<PageResult<ManagedFamily>>>('/admin/families', { params }))
}

export async function getFamily(familyId: number | string) {
  if (apiMode === 'mock') {
    return mapManagedFamilies().find((item) => String(item.id) === String(familyId)) || mapManagedFamilies()[0]
  }
  return unwrapData(await apiClient.get<ApiResponse<ManagedFamily>>(`/admin/families/${familyId}`))
}

export async function listFamilyMembers(familyId: number | string) {
  if (apiMode === 'mock') {
    return mapFamilyMembers(familyId)
  }
  return unwrapData(await apiClient.get<ApiResponse<ManagedFamilyMember[]>>(`/admin/families/${familyId}/members`))
}

export interface UpdateManagedFamilyMemberInput {
  displayName?: string
  gender?: string
  birthDate?: string | null
  isLiving?: boolean | null
  userBindingPolicy?: string
}

export type ManagedRelativeAddType = 'ADD_FATHER' | 'ADD_MOTHER' | 'ADD_CHILD' | 'ADD_SPOUSE' | 'ADD_SIBLING'

export interface CreateManagedFamilyRelativeInput {
  addType: ManagedRelativeAddType
  newMember: {
    displayName: string
    gender: string
    birthDate?: string | null
    isLiving?: boolean | null
    userBindingPolicy?: string
    memberType?: string
  }
  relationship?: {
    parentLinkType?: string
  }
}

export interface CreatedManagedFamilyRelative {
  memberId: number
  relationshipIds: number[]
  graphVersion: number
}

export async function updateFamilyMember(familyId: number | string, memberId: number | string, input: UpdateManagedFamilyMemberInput) {
  if (apiMode === 'mock') {
    return { id: Number(memberId), familyId: Number(familyId), status: 'ACTIVE', ...input } as ManagedFamilyMember
  }
  return unwrapData(await apiClient.put<ApiResponse<ManagedFamilyMember>>(`/admin/families/${familyId}/members/${memberId}`, input))
}

export async function deleteFamilyMember(familyId: number | string, memberId: number | string, reason?: string) {
  if (apiMode === 'mock') return { status: 'ok' }
  return unwrapData(await apiClient.delete<ApiResponse<{ status: string }>>(`/admin/families/${familyId}/members/${memberId}`, { data: { reason } }))
}

export async function createFamilyRelative(familyId: number | string, memberId: number | string, input: CreateManagedFamilyRelativeInput) {
  if (apiMode === 'mock') {
    return { memberId: Date.now(), relationshipIds: [Date.now() + 1], graphVersion: 2 } satisfies CreatedManagedFamilyRelative
  }
  return unwrapData(await apiClient.post<ApiResponse<CreatedManagedFamilyRelative>>(`/admin/families/${familyId}/members/${memberId}/relatives`, input))
}

export async function getFamilyTree(familyId: number | string): Promise<AdminFamilyTree> {
  if (apiMode === 'mock') {
    const nodes = mapFamilyMembers(familyId).map((member) => ({
      memberId: member.id,
      displayName: member.displayName,
      gender: member.gender,
      memberType: 'LINEAGE_MEMBER',
      isLiving: member.isLiving,
      userBindingPolicy: member.userBindingPolicy,
      userBindingState: member.boundUserId ? 'BOUND' : member.userBindingPolicy === 'NOT_REQUIRED' ? 'NOT_REQUIRED' : 'UNBOUND'
    }))
    return {
      familyId: Number(familyId),
      graphVersion: 1,
      nodes,
      edges: nodes.slice(1).map((node, index) => ({
        relationshipId: index + 1,
        fromMemberId: nodes[0].memberId,
        toMemberId: node.memberId,
        relationshipType: 'PARENT_CHILD'
      }))
    }
  }
  return unwrapData(await apiClient.get<ApiResponse<AdminFamilyTree>>(`/admin/families/${familyId}/tree`))
}

export async function listAdmins(params: PageQuery) {
  if (apiMode === 'mock') {
    const items = mapAdmins().filter((item) => (!params.role || item.role === params.role) && (!params.status || item.status === params.status))
    return paginate(items, params.page, params.pageSize)
  }
  return unwrapData(await apiClient.get<ApiResponse<PageResult<ManagedAdmin>>>('/admin/admin-users', { params }))
}

export async function createAdmin(input: { username: string; password: string; displayName?: string; phone?: string; email?: string; role: AdminRole; remark?: string }) {
  return unwrapData(await apiClient.post<ApiResponse<ManagedAdmin>>('/admin/admin-users', input))
}

export async function updateAdmin(adminId: number, input: { displayName?: string; phone?: string; email?: string; role?: AdminRole; remark?: string }) {
  return unwrapData(await apiClient.put<ApiResponse<ManagedAdmin>>(`/admin/admin-users/${adminId}`, input))
}

export async function setAdminStatus(adminId: number, status: 'ACTIVE' | 'DISABLED' | 'DELETED' | 'UNLOCKED', reason?: string) {
  return unwrapData(await apiClient.post<ApiResponse<{ status: string }>>(`/admin/admin-users/${adminId}/status`, { status, reason }))
}

export async function setUserStatus(userId: number, status: 'ACTIVE' | 'DISABLED' | 'DELETED', reason?: string) {
  return unwrapData(await apiClient.post<ApiResponse<{ status: string }>>(`/admin/users/${userId}/status`, { status, reason }))
}

export async function updateFamily(familyId: number, input: { familyName?: string; description?: string; nativePlace?: string; regionText?: string; searchable?: boolean }) {
  return unwrapData(await apiClient.put<ApiResponse<ManagedFamily>>(`/admin/families/${familyId}`, input))
}

export async function setFamilyStatus(familyId: number, status: 'NORMAL' | 'DISABLED' | 'DELETED', reason?: string) {
  return unwrapData(await apiClient.post<ApiResponse<{ status: string }>>(`/admin/families/${familyId}/status`, { status, reason }))
}


export async function listOperationLogs(params: PageQuery) {
  if (apiMode === 'mock') {
    const items = mapOperationLogs().filter((item) => {
      const hitKeyword = !params.keyword || [item.module, item.action, item.targetType].some((value) => String(value || '').includes(params.keyword!))
      const hitResult = !params.result || item.result === params.result
      const hitOperatorType = !params.operatorType || item.operatorType === params.operatorType
      const hitRole = !params.role || item.operatorRole === params.role
      const createdDate = String(item.createdAt).slice(0, 10)
      const hitStart = !params.startDate || createdDate >= params.startDate
      const hitEnd = !params.endDate || createdDate <= params.endDate
      return hitKeyword && hitResult && hitOperatorType && hitRole && hitStart && hitEnd
    })
    return paginate(items, params.page, params.pageSize)
  }
  return unwrapData(await apiClient.get<ApiResponse<PageResult<OperationLogRecord>>>('/admin/operation-logs', { params }))
}

export async function listRecoveryItems(params: Pick<PageQuery, 'page' | 'pageSize'>) {
  if (apiMode === 'mock') return paginate<RecoveryItem>([], params.page, params.pageSize)
  return unwrapData(await apiClient.get<ApiResponse<PageResult<RecoveryItem>>>('/admin/recovery-items', { params }))
}

export async function restoreRecoveryItem(item: RecoveryItem) {
  if (apiMode === 'mock') return { status: 'ok' }
  const path = item.kind === 'MEMBER'
    ? `/admin/families/${item.familyId}/members/${item.id}/restore`
    : `/admin/families/${item.familyId}/relationships/${item.id}/restore`
  return unwrapData(await apiClient.post<ApiResponse<{ status: string }>>(path))
}
