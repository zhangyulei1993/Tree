import { apiClient, isRealApiMode } from '@/api/client'
import { invitation as mockInvitation } from '@/mock/data'
import type {
  ApiResponse,
  CreatedInvitation,
  CreateInvitationInput,
  Invitation,
  InvitationActionInput
} from '@/types/api'

const initialMockInvitation: Invitation = {
  invitationId: mockInvitation.id,
  familyId: 'family_001',
  familyName: mockInvitation.familyName,
  targetMemberId: 'member_004',
  targetMemberName: mockInvitation.memberName,
  inviteChannel: 'IN_APP',
  inviteMessage: '邀请你绑定到该家庭成员节点。',
  familyRoleAfterAccept: 'MEMBER',
  status: mockInvitation.status,
  expiredAt: new Date('2026-06-15T12:00:00+08:00').toISOString(),
  createdAt: new Date('2026-06-08T12:00:00+08:00').toISOString()
}

const mockInvitations: Invitation[] = [initialMockInvitation]
const mockTokenInvitations = new Map<string, Invitation>([
  [mockInvitation.token, initialMockInvitation]
])

export async function createInvitation(
  familyId: number | string,
  memberId: number | string,
  input: CreateInvitationInput
): Promise<CreatedInvitation> {
  if (!isRealApiMode) {
    const now = new Date()
    const inviteToken = `mock_share_${Date.now()}`
    const value: Invitation = {
      invitationId: Date.now(),
      familyId,
      familyName: `Mock 家庭 ${familyId}`,
      targetMemberId: memberId,
      targetMemberName: `成员 ${memberId}`,
      inviteChannel: input.inviteChannel,
      inviteMessage: input.inviteMessage || null,
      familyRoleAfterAccept: input.familyRoleAfterAccept || 'MEMBER',
      status: 'PENDING',
      expiredAt: new Date(now.getTime() + 7 * 24 * 60 * 60 * 1000).toISOString(),
      createdAt: now.toISOString()
    }
    mockInvitations.unshift(value)
    mockTokenInvitations.set(inviteToken, value)
    return { invitation: value, inviteToken }
  }
  const response = await apiClient.post<ApiResponse<CreatedInvitation>>(
    `/families/${familyId}/members/${memberId}/invite`,
    input
  )
  return response.data.data
}

export async function getInvitationDetail(inviteToken: string): Promise<Invitation> {
  if (!isRealApiMode) {
    const value = mockTokenInvitations.get(inviteToken)
    if (!value) throw new Error('邀请不存在')
    return { ...value }
  }
  const response = await apiClient.get<ApiResponse<Invitation>>(
    `/invitations/${encodeURIComponent(inviteToken)}`
  )
  return response.data.data
}

export async function acceptInvitation(invitationId: number | string): Promise<Invitation> {
  if (!isRealApiMode) {
    return updateMockInvitation(invitationId, 'ACCEPTED', 'acceptedAt')
  }
  const response = await apiClient.post<ApiResponse<Invitation>>(
    `/invitations/${invitationId}/accept`
  )
  return response.data.data
}

export async function rejectInvitation(
  invitationId: number | string,
  input: InvitationActionInput = {}
): Promise<Invitation> {
  if (!isRealApiMode) {
    return updateMockInvitation(invitationId, 'REJECTED', 'rejectedAt')
  }
  const response = await apiClient.post<ApiResponse<Invitation>>(
    `/invitations/${invitationId}/reject`,
    input
  )
  return response.data.data
}

export async function cancelInvitation(
  invitationId: number | string,
  input: InvitationActionInput = {}
): Promise<Invitation> {
  if (!isRealApiMode) {
    return updateMockInvitation(invitationId, 'CANCELLED', 'cancelledAt')
  }
  const response = await apiClient.post<ApiResponse<Invitation>>(
    `/invitations/${invitationId}/cancel`,
    input
  )
  return response.data.data
}

export async function listMyInvitations(): Promise<Invitation[]> {
  if (!isRealApiMode) {
    return mockInvitations.filter((value) => value.inviteChannel === 'IN_APP').map((value) => ({ ...value }))
  }
  const response = await apiClient.get<ApiResponse<Invitation[]>>('/users/me/invitations')
  return response.data.data
}

export async function listFamilyInvitations(familyId: number | string): Promise<Invitation[]> {
  const response = await apiClient.get<ApiResponse<Invitation[]>>(`/families/${familyId}/invitations`)
  return response.data.data
}

export async function regenerateInvitation(invitationId: number | string): Promise<CreatedInvitation> {
  const response = await apiClient.post<ApiResponse<CreatedInvitation>>(`/invitations/${invitationId}/regenerate`)
  return response.data.data
}

function updateMockInvitation(
  invitationId: number | string,
  status: string,
  timeField: 'acceptedAt' | 'rejectedAt' | 'cancelledAt'
) {
  const value = mockInvitations.find((item) => String(item.invitationId) === String(invitationId))
  if (!value || value.status !== 'PENDING') throw new Error('邀请状态不可操作')
  value.status = status
  value[timeField] = new Date().toISOString()
  return { ...value }
}
