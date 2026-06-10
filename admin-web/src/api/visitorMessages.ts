import { apiClient, unwrapData } from '@/api/client'
import type { DeleteVisitorMessageRequest, PageResult, ReviewRequest, VisitorMessage } from '@/types/api'

export interface VisitorMessageQuery {
  status?: string
  familyId?: number
  page?: number
  pageSize?: number
}

export async function listVisitorMessages(query: VisitorMessageQuery): Promise<PageResult<VisitorMessage>> {
  const response = await apiClient.get('/admin/visitor-messages', { params: query })
  return unwrapData<PageResult<VisitorMessage>>(response)
}

export async function approveVisitorMessage(messageId: number, input: ReviewRequest): Promise<VisitorMessage> {
  const response = await apiClient.post(`/admin/visitor-messages/${messageId}/approve`, input)
  return unwrapData<VisitorMessage>(response)
}

export async function rejectVisitorMessage(messageId: number, input: ReviewRequest): Promise<VisitorMessage> {
  const response = await apiClient.post(`/admin/visitor-messages/${messageId}/reject`, input)
  return unwrapData<VisitorMessage>(response)
}

export async function deleteVisitorMessage(messageId: number, input: DeleteVisitorMessageRequest): Promise<void> {
  await apiClient.delete(`/admin/visitor-messages/${messageId}`, { data: input })
}
