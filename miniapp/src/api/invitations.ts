import { isRealApiMode, request } from '@/api/client'
import { invite } from '@/mock/data'
import type { Invitation, RejectInvitationInput } from '@/types/api'

const mockInvitation: Invitation = {
  invitationId: 'mock_invitation',
  familyId: 'family_001',
  familyName: invite.familyName,
  targetMemberId: 'member_001',
  targetMemberName: invite.memberName,
  inviteChannel: 'SHARE_LINK',
  inviteMessage: null,
  familyRoleAfterAccept: 'MEMBER',
  status: 'PENDING',
  expiredAt: new Date('2026-06-15T12:00:00+08:00').toISOString(),
  createdAt: new Date('2026-06-08T12:00:00+08:00').toISOString()
}

const mockInvitations: Invitation[] = [{ ...mockInvitation }]

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

export async function listMyInvitations(): Promise<Invitation[]> {
  if (!isRealApiMode) return mockInvitations.map((item) => ({ ...item }))
  return request<Invitation[]>('/users/me/invitations')
}

function updateMockInvitation(
  invitationId: number | string,
  status: string,
  timeField: 'acceptedAt' | 'rejectedAt'
) {
  const item = mockInvitations.find((value) => String(value.invitationId) === String(invitationId))
  if (!item || item.status !== 'PENDING') throw new Error('邀请状态不可操作')
  item.status = status
  item[timeField] = new Date().toISOString()
  return { ...item }
}
