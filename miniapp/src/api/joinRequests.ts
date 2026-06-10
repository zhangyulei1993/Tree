import { isRealApiMode, request } from '@/api/client'
import { joinRequests } from '@/mock/data'
import type {
  CancelJoinRequestInput,
  CreateJoinRequestInput,
  JoinRequest
} from '@/types/api'

const mockRequests: JoinRequest[] = joinRequests.map((item, index) => ({
  requestId: item.id,
  familyId: `family_00${index + 1}`,
  familyName: item.familyName,
  applicantUserId: 'mock_user',
  applicantMessage: item.reason,
  requestStatus: item.status,
  createdAt: new Date('2026-06-08T12:00:00+08:00').toISOString(),
  updatedAt: new Date('2026-06-08T12:00:00+08:00').toISOString()
}))

export async function createJoinRequest(
  familyId: number | string,
  input: CreateJoinRequestInput
): Promise<JoinRequest> {
  if (!isRealApiMode) {
    const now = new Date().toISOString()
    const value: JoinRequest = {
      requestId: `join_mock_${Date.now()}`,
      familyId,
      familyName: `Mock 家庭 ${familyId}`,
      applicantUserId: 'mock_user',
      applicantRealName: input.applicantRealName || null,
      applicantMessage: input.applicantMessage || null,
      requestStatus: 'PENDING',
      createdAt: now,
      updatedAt: now
    }
    mockRequests.unshift(value)
    return { ...value }
  }
  return request<JoinRequest, CreateJoinRequestInput>(
    `/families/${familyId}/join-requests`,
    {
      method: 'POST',
      data: input
    }
  )
}

export async function listMyJoinRequests(): Promise<JoinRequest[]> {
  if (!isRealApiMode) return mockRequests.map((item) => ({ ...item }))
  return request<JoinRequest[]>('/users/me/join-requests')
}

export async function cancelJoinRequest(
  familyId: number | string,
  requestId: number | string,
  input: CancelJoinRequestInput = {}
): Promise<JoinRequest> {
  if (!isRealApiMode) {
    const item = mockRequests.find((value) => String(value.requestId) === String(requestId))
    if (!item || item.requestStatus !== 'PENDING') throw new Error('加入申请状态不可操作')
    const now = new Date().toISOString()
    item.requestStatus = 'CANCELLED'
    item.cancelledAt = now
    item.updatedAt = now
    return { ...item }
  }
  return request<JoinRequest, CancelJoinRequestInput>(
    `/families/${familyId}/join-requests/${requestId}/cancel`,
    {
      method: 'POST',
      data: input
    }
  )
}
