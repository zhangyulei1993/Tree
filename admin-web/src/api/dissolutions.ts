import { apiClient, unwrapData } from '@/api/client'
import type { DissolutionRequest, PageResult, ReviewRequest } from '@/types/api'

export interface DissolutionQuery {
  status?: string
  familyId?: number
  page?: number
  pageSize?: number
}

export async function listDissolutionRequests(query: DissolutionQuery): Promise<PageResult<DissolutionRequest>> {
  const response = await apiClient.get('/admin/dissolution-requests', { params: query })
  return unwrapData<PageResult<DissolutionRequest>>(response)
}

export async function approveDissolutionRequest(requestId: number, input: ReviewRequest): Promise<DissolutionRequest> {
  const response = await apiClient.post(`/admin/dissolution-requests/${requestId}/approve`, input)
  return unwrapData<DissolutionRequest>(response)
}

export async function rejectDissolutionRequest(requestId: number, input: ReviewRequest): Promise<DissolutionRequest> {
  const response = await apiClient.post(`/admin/dissolution-requests/${requestId}/reject`, input)
  return unwrapData<DissolutionRequest>(response)
}
