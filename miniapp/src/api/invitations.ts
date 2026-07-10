import { isRealApiMode, request } from '@/api/client'
import { invite } from '@/mock/data'
import type {
  CreatedInvitation,
  CreateInvitationInput,
  Invitation,
  RejectInvitationInput
} from '@/types/api'

const mockInvitation: Invitation = {
  invitationId: 'mock_invitation',
  familyId: 'family_001',
  familyName: invite.familyName,
  targetMemberId: 'member_001',
  targetMemberName: invite.memberName,
  inviteType: 'CLAIM_EXISTING_MEMBER',
  inviteChannel: 'SHARE_LINK',
  inviteMessage: null,
  familyRoleAfterAccept: 'MEMBER',
  status: 'PENDING',
  expiredAt: new Date('2026-06-15T12:00:00+08:00').toISOString(),
  createdAt: new Date('2026-06-08T12:00:00+08:00').toISOString()
}

const mockInvitations: Invitation[] = [{ ...mockInvitation }]

export async function createInvitation(
  familyId: number | string,
  memberId: number | string,
  input: CreateInvitationInput
): Promise<CreatedInvitation> {
  if (!isRealApiMode) {
    const now = new Date()
    const invitation: Invitation = {
      ...mockInvitation,
      invitationId: `mock_invitation_${Date.now()}`,
      familyId,
      targetMemberId: memberId,
      inviteChannel: input.inviteChannel,
      inviteMessage: input.inviteMessage || null,
      familyRoleAfterAccept: input.familyRoleAfterAccept,
      createdAt: now.toISOString(),
      expiredAt: new Date(now.getTime() + 7 * 24 * 60 * 60 * 1000).toISOString()
    }
    mockInvitations.unshift(invitation)
    return { invitation: { ...invitation }, inviteToken: `mock_invite_${Date.now()}` }
  }
  return request<CreatedInvitation, CreateInvitationInput>(
    `/families/${familyId}/members/${memberId}/invite`,
    {
      method: 'POST',
      data: input
    }
  )
}

export async function createFamilyInvitation(
  familyId: number | string,
  input: CreateInvitationInput
): Promise<CreatedInvitation> {
  if (!isRealApiMode) {
    const now = new Date()
    const invitation: Invitation = {
      ...mockInvitation,
      invitationId: `mock_family_invitation_${Date.now()}`,
      familyId,
      targetMemberId: `pending_member_${Date.now()}`,
      targetMemberName: input.pendingMemberLabel || '待确认成员',
      pendingMemberLabel: input.pendingMemberLabel || '待确认成员',
      inviteType: 'JOIN_FAMILY_PENDING_MEMBER',
      inviteChannel: input.inviteChannel,
      inviteMessage: input.inviteMessage || null,
      familyRoleAfterAccept: input.familyRoleAfterAccept,
      createdAt: now.toISOString(),
      expiredAt: new Date(now.getTime() + 7 * 24 * 60 * 60 * 1000).toISOString()
    }
    mockInvitations.unshift(invitation)
    return { invitation: { ...invitation }, inviteToken: `mock_family_invite_${Date.now()}` }
  }
  return request<CreatedInvitation, CreateInvitationInput>(
    `/families/${familyId}/invitations`,
    {
      method: 'POST',
      data: input
    }
  )
}

export async function getInvitationDetail(inviteToken: string): Promise<Invitation> {
  if (!isRealApiMode) return { ...mockInvitation }
  return request<Invitation>(`/invitations/${encodeURIComponent(inviteToken)}`, { public: true })
}

export async function acceptInvitation(invitationId: number | string): Promise<Invitation> {
  if (!isRealApiMode) return updateMockInvitation(invitationId, 'ACCEPTED', 'acceptedAt')
  return request<Invitation>(`/invitations/${invitationId}/accept`, { method: 'POST' })
}

export async function rejectInvitation(
  invitationId: number | string,
  input: RejectInvitationInput = {}
): Promise<Invitation> {
  if (!isRealApiMode) return updateMockInvitation(invitationId, 'REJECTED', 'rejectedAt')
  return request<Invitation, RejectInvitationInput>(
    `/invitations/${invitationId}/reject`,
    {
      method: 'POST',
      data: input
    }
  )
}

export async function cancelInvitation(
  invitationId: number | string,
  input: RejectInvitationInput = {}
): Promise<Invitation> {
  if (!isRealApiMode) return updateMockInvitation(invitationId, 'CANCELLED', 'cancelledAt')
  return request<Invitation, RejectInvitationInput>(
    `/invitations/${invitationId}/cancel`,
    {
      method: 'POST',
      data: input
    }
  )
}

export async function listMyInvitations(): Promise<Invitation[]> {
	if (!isRealApiMode) return mockInvitations.map((item) => ({ ...item }))
	return request<Invitation[]>('/users/me/invitations')
}

export async function listFamilyInvitations(
  familyId: number | string
): Promise<Invitation[]> {
  if (!isRealApiMode) {
    return mockInvitations
      .filter((item) => String(item.familyId) === String(familyId))
      .map((item) => ({ ...item }))
  }
  return request<Invitation[]>(`/families/${familyId}/invitations`)
}

export async function regenerateInvitation(
  invitationId: number | string
): Promise<CreatedInvitation> {
  if (!isRealApiMode) {
    const current = mockInvitations.find((item) => String(item.invitationId) === String(invitationId))
    if (!current || current.status !== 'PENDING') throw new Error('邀请状态不可操作')
    current.status = 'CANCELLED'
    current.cancelledAt = new Date().toISOString()
    return createInvitation(current.familyId, current.targetMemberId, {
      inviteChannel: 'SHARE_LINK',
      inviteMessage: current.inviteMessage || undefined,
      familyRoleAfterAccept: 'MEMBER'
    })
  }
  return request<CreatedInvitation>(
    `/invitations/${invitationId}/regenerate`,
    { method: 'POST' }
  )
}

function updateMockInvitation(
  invitationId: number | string,
  status: string,
  timeField: 'acceptedAt' | 'rejectedAt' | 'cancelledAt'
) {
  const item = mockInvitations.find((value) => String(value.invitationId) === String(invitationId))
  if (!item || item.status !== 'PENDING') throw new Error('邀请状态不可操作')
  item.status = status
  item[timeField] = new Date().toISOString()
  return { ...item }
}
