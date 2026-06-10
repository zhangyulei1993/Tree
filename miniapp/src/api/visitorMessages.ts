import { isRealApiMode, request } from '@/api/client'
import { messages } from '@/mock/data'
import type {
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
  const page = query.page || 1
  const pageSize = query.pageSize || 20
  if (!isRealApiMode) {
    const items = messages.map((message) => ({
      messageId: message.id,
      familyId,
      visitorName: message.visitorName,
      messageContent: message.content,
      createdAt: new Date(0).toISOString()
    }))
    const offset = (page - 1) * pageSize
    return {
      items: items.slice(offset, offset + pageSize),
      page,
      pageSize,
      total: items.length
    }
  }

  const queryString = `page=${encodeURIComponent(String(page))}&pageSize=${encodeURIComponent(String(pageSize))}`
  return request<PaginatedResult<PublicVisitorMessage>>(
    `/public/families/${familyId}/visitor-messages?${queryString}`,
    { public: true }
  )
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

  return request<VisitorMessageRecord, CreateVisitorMessageInput>(
    `/public/families/${familyId}/visitor-messages`,
    {
      method: 'POST',
      public: true,
      data: input
    }
  )
}
