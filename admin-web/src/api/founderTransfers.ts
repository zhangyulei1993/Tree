import { apiClient, apiMode, unwrapData } from '@/api/client'
import { listMockFounderTransferRequests } from '@/api/management'
import type { FounderTransferRequest, PageResult, ReviewRequest } from '@/types/api'

export interface FounderTransferQuery {
  status?: string
  familyId?: number
  page?: number
  pageSize?: number
}

export async function listFounderTransferRequests(query: FounderTransferQuery): Promise<PageResult<FounderTransferRequest>> {
  if (apiMode === 'mock') {
    return listMockFounderTransferRequests(query)
  }
  const response = await apiClient.get('/admin/founder-transfer-requests', { params: query })
  return unwrapData<PageResult<FounderTransferRequest>>(response)
}

export async function approveFounderTransferRequest(requestId: number, input: ReviewRequest): Promise<FounderTransferRequest> {
  if (apiMode === 'mock') {
    return {
      requestId,
      familyId: 1,
      familyName: '示例家庭',
      fromMemberId: 1,
      fromUserId: 1,
      toMemberId: 2,
      toUserId: 2,
      requestStatus: 'APPROVED',
      reviewComment: input.reviewComment,
      createdAt: '2026-06-29 10:00',
      updatedAt: '2026-06-29 10:00'
    }
  }
  const response = await apiClient.post(`/admin/founder-transfer-requests/${requestId}/approve`, input)
  return unwrapData<FounderTransferRequest>(response)
}

export async function rejectFounderTransferRequest(requestId: number, input: ReviewRequest): Promise<FounderTransferRequest> {
  if (apiMode === 'mock') {
    return {
      requestId,
      familyId: 1,
      familyName: '示例家庭',
      fromMemberId: 1,
      fromUserId: 1,
      toMemberId: 2,
      toUserId: 2,
      requestStatus: 'REJECTED',
      reviewComment: input.reviewComment,
      createdAt: '2026-06-29 10:00',
      updatedAt: '2026-06-29 10:00'
    }
  }
  const response = await apiClient.post(`/admin/founder-transfer-requests/${requestId}/reject`, input)
  return unwrapData<FounderTransferRequest>(response)
}
