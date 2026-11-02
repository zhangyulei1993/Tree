import { request } from '@/api/client'
import type { FounderTransferRequest } from '@/types/api'

export function getCurrentFounderTransfer(familyId: number | string) {
  return request<FounderTransferRequest>(`/families/${familyId}/founder-transfer-requests/current`)
}

export function createFounderTransfer(
  familyId: number | string,
  toMemberId: number | string,
  requestReason?: string
) {
  return request<FounderTransferRequest, { toMemberId: number | string; requestReason?: string }>(
    `/families/${familyId}/founder-transfer-requests`,
    { method: 'POST', data: { toMemberId, requestReason: requestReason || undefined } }
  )
}

export function cancelFounderTransfer(
  familyId: number | string,
  requestId: number | string,
  cancelReason?: string
) {
  return request<FounderTransferRequest, { cancelReason?: string }>(
    `/families/${familyId}/founder-transfer-requests/${requestId}/cancel`,
    { method: 'POST', data: cancelReason ? { cancelReason } : {} }
  )
}
