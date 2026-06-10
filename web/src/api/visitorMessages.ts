import { apiClient, isRealApiMode } from '@/api/client'
import { visitorMessages } from '@/mock/data'
import type {
  ApiResponse,
  CreateVisitorMessageInput,
  PaginatedResult,
  PaginationQuery,
  PublicVisitorMessage,
  VisitorMessageRecord
} from '@/types/api'

export async function listPublicVisitorMessages(
  familyId: number | string,
  query: PaginationQuery = {}
): Promise<PaginatedResult<PublicVisitorMessage>> {
  if (!isRealApiMode) {
    const page = query.page || 1
    const pageSize = query.pageSize || 20
    const matching = visitorMessages
      .filter((message) => message.familyId === String(familyId))
      .map((message) => ({
        messageId: message.id,
        familyId: message.familyId,
        visitorName: message.visitorName,
        messageContent: message.content,
        createdAt: message.createdAt
      }))
    const offset = (page - 1) * pageSize
    return {
      items: matching.slice(offset, offset + pageSize),
      page,
      pageSize,
      total: matching.length
    }
  }
  const response = await apiClient.get<ApiResponse<PaginatedResult<PublicVisitorMessage>>>(
    `/public/families/${familyId}/visitor-messages`,
    { params: query }
  )
  return response.data.data
}

export async function createVisitorMessage(
  familyId: number | string,
  input: CreateVisitorMessageInput
): Promise<VisitorMessageRecord> {
  if (!isRealApiMode) {
    const now = new Date().toISOString()
    return {
      messageId: `message_mock_${Date.now()}`,
      familyId,
      visitorName: input.visitorName || null,
      visitorPhone: input.visitorPhone || null,
      visitorWechat: input.visitorWechat || null,
      messageContent: input.messageContent,
      status: 'PENDING',
      createdAt: now,
      updatedAt: now
    }
  }
  const response = await apiClient.post<ApiResponse<VisitorMessageRecord>>(
    `/public/families/${familyId}/visitor-messages`,
    input
  )
  return response.data.data
}
