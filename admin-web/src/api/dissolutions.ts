import { apiClient, apiMode, unwrapData } from '@/api/client'
import { listMockDissolutionRequests } from '@/api/management'
import type { DissolutionRequest, PageResult, ReviewRequest } from '@/types/api'

export interface DissolutionQuery {
  status?: string
  familyId?: number
  page?: number
  pageSize?: number
}

export async function listDissolutionRequests(query: DissolutionQuery): Promise<PageResult<DissolutionRequest>> {
  if (apiMode === 'mock') {
    return listMockDissolutionRequests(query)
  }
  const response = await apiClient.get('/admin/dissolution-requests', { params: query })
  return unwrapData<PageResult<DissolutionRequest>>(response)
}

export async function approveDissolutionRequest(requestId: number, input: ReviewRequest): Promise<DissolutionRequest> {
  if (apiMode === 'mock') {
    return {
      requestId,
      familyId: 1,
      familyName: '示例家庭',
      familyStatus: 'DISSOLVED',
      requesterMemberId: 1,
      requesterUserId: 1,
      requestStatus: 'APPROVED',
      reviewComment: input.reviewComment,
      createdAt: '2026-06-29 10:00',
      updatedAt: '2026-06-29 10:00'
    }
  }
  const response = await apiClient.post(`/admin/dissolution-requests/${requestId}/approve`, input)
  return unwrapData<DissolutionRequest>(response)
}

export async function rejectDissolutionRequest(requestId: number, input: ReviewRequest): Promise<DissolutionRequest> {
  if (apiMode === 'mock') {
    return {
      requestId,
      familyId: 1,
      familyName: '示例家庭',
      familyStatus: 'NORMAL',
      requesterMemberId: 1,
      requesterUserId: 1,
      requestStatus: 'REJECTED',
      reviewComment: input.reviewComment,
      createdAt: '2026-06-29 10:00',
      updatedAt: '2026-06-29 10:00'
    }
  }
  const response = await apiClient.post(`/admin/dissolution-requests/${requestId}/reject`, input)
  return unwrapData<DissolutionRequest>(response)
}
