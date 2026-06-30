import { apiClient, apiMode, unwrapData } from '@/api/client'
import { listMockVisitorMessages } from '@/api/management'
import type { DeleteVisitorMessageRequest, PageResult, ReviewRequest, VisitorMessage } from '@/types/api'

export interface VisitorMessageQuery {
  status?: string
  familyId?: number
  page?: number
  pageSize?: number
}

export async function listVisitorMessages(query: VisitorMessageQuery): Promise<PageResult<VisitorMessage>> {
  if (apiMode === 'mock') {
    return listMockVisitorMessages(query)
  }
  const response = await apiClient.get('/admin/visitor-messages', { params: query })
  return unwrapData<PageResult<VisitorMessage>>(response)
}

export async function approveVisitorMessage(messageId: number, input: ReviewRequest): Promise<VisitorMessage> {
  if (apiMode === 'mock') {
    return {
      messageId,
      familyId: 1,
      familyName: '示例家庭',
      visitorName: '访客',
      messageContent: '示例留言',
      status: 'APPROVED',
      reviewComment: input.reviewComment,
      createdAt: '2026-06-29 10:00',
      updatedAt: '2026-06-29 10:00'
    } as VisitorMessage
  }
  const response = await apiClient.post(`/admin/visitor-messages/${messageId}/approve`, input)
  return unwrapData<VisitorMessage>(response)
}

export async function rejectVisitorMessage(messageId: number, input: ReviewRequest): Promise<VisitorMessage> {
  if (apiMode === 'mock') {
    return {
      messageId,
      familyId: 1,
      familyName: '示例家庭',
      visitorName: '访客',
      messageContent: '示例留言',
      status: 'REJECTED',
      reviewComment: input.reviewComment,
      createdAt: '2026-06-29 10:00',
      updatedAt: '2026-06-29 10:00'
    } as VisitorMessage
  }
  const response = await apiClient.post(`/admin/visitor-messages/${messageId}/reject`, input)
  return unwrapData<VisitorMessage>(response)
}

export async function deleteVisitorMessage(messageId: number, input: DeleteVisitorMessageRequest): Promise<void> {
  if (apiMode === 'mock') return
  await apiClient.delete(`/admin/visitor-messages/${messageId}`, { data: input })
}
