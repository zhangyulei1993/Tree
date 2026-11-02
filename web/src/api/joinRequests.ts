import { apiClient, isRealApiMode } from '@/api/client'
import { joinRequests as mockJoinRequests } from '@/mock/data'
import type {
  ApiResponse,
  ApproveJoinRequestInput,
  CancelJoinRequestInput,
  CreateJoinRequestInput,
  JoinRequest,
  RejectJoinRequestInput
} from '@/types/api'

const mockRequests: JoinRequest[] = mockJoinRequests.map((value, index) => ({
  requestId: value.id,
  familyId: `family_00${index + 2}`,
  familyName: value.familyName,
  applicantUserId: 'user_001',
  applicantMessage: value.reason,
  requestStatus: value.status,
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
      applicantUserId: 'user_001',
      applicantRealName: input.applicantRealName || null,
      applicantGender: input.applicantGender,
      applicantMessage: input.applicantMessage || null,
      requestStatus: 'PENDING',
      createdAt: now,
      updatedAt: now
    }
    mockRequests.unshift(value)
    return { ...value }
  }
  const response = await apiClient.post<ApiResponse<JoinRequest>>(
    `/families/${familyId}/join-requests`,
    input
  )
  return response.data.data
}

export async function listFamilyJoinRequests(familyId: number | string): Promise<JoinRequest[]> {
  const response = await apiClient.get<ApiResponse<JoinRequest[]>>(`/families/${familyId}/join-requests`)
  return response.data.data
}

export async function approveJoinRequest(
  familyId: number | string,
  requestId: number | string,
  input: ApproveJoinRequestInput
): Promise<JoinRequest> {
  const request = {
    ...input,
    memberId: input.memberId == null ? undefined : Number(input.memberId),
    location: input.location
      ? { ...input.location, baseMemberId: Number(input.location.baseMemberId) }
      : undefined
  }
  const response = await apiClient.post<ApiResponse<JoinRequest>>(
    `/families/${familyId}/join-requests/${requestId}/approve`, request
  )
  return response.data.data
}

export async function rejectJoinRequest(
  familyId: number | string,
  requestId: number | string,
  input: RejectJoinRequestInput = {}
): Promise<JoinRequest> {
  const response = await apiClient.post<ApiResponse<JoinRequest>>(
    `/families/${familyId}/join-requests/${requestId}/reject`, input
  )
  return response.data.data
}

export async function listMyJoinRequests(): Promise<JoinRequest[]> {
  if (!isRealApiMode) {
    return mockRequests.map((value) => ({ ...value }))
  }
  const response = await apiClient.get<ApiResponse<JoinRequest[]>>('/users/me/join-requests')
  return response.data.data
}

export async function cancelJoinRequest(
  familyId: number | string,
  requestId: number | string,
  input: CancelJoinRequestInput = {}
): Promise<JoinRequest> {
  if (!isRealApiMode) {
    const value = mockRequests.find((item) => String(item.requestId) === String(requestId))
    if (!value || value.requestStatus !== 'PENDING') throw new Error('加入申请状态不可操作')
    const now = new Date().toISOString()
    value.requestStatus = 'CANCELLED'
    value.cancelledAt = now
    value.updatedAt = now
    return { ...value }
  }
  const response = await apiClient.post<ApiResponse<JoinRequest>>(
    `/families/${familyId}/join-requests/${requestId}/cancel`,
    input
  )
  return response.data.data
}
