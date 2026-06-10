import { apiClient, unwrapData } from '@/api/client'
import type { FounderTransferRequest, PageResult, ReviewRequest } from '@/types/api'

export interface FounderTransferQuery {
  status?: string
  familyId?: number
  page?: number
  pageSize?: number
}

export async function listFounderTransferRequests(query: FounderTransferQuery): Promise<PageResult<FounderTransferRequest>> {
  const response = await apiClient.get('/admin/founder-transfer-requests', { params: query })
  return unwrapData<PageResult<FounderTransferRequest>>(response)
}

export async function approveFounderTransferRequest(requestId: number, input: ReviewRequest): Promise<FounderTransferRequest> {
  const response = await apiClient.post(`/admin/founder-transfer-requests/${requestId}/approve`, input)
  return unwrapData<FounderTransferRequest>(response)
}

export async function rejectFounderTransferRequest(requestId: number, input: ReviewRequest): Promise<FounderTransferRequest> {
  const response = await apiClient.post(`/admin/founder-transfer-requests/${requestId}/reject`, input)
  return unwrapData<FounderTransferRequest>(response)
}
