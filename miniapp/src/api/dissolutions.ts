import { request } from '@/api/client'
import type { DissolutionRequest } from '@/types/api'

type DissolutionResponse = Omit<DissolutionRequest, 'requestId'> & {
  id: number | string
}

function normalize(value: DissolutionResponse | null): DissolutionRequest | null {
  return value ? { ...value, requestId: value.id } : null
}

export async function getCurrentDissolution(familyId: number | string) {
  const result = await request<DissolutionResponse | null>(
    `/families/${familyId}/dissolution-requests/current`
  )
  return normalize(result)
}

export async function createDissolution(familyId: number | string, requestReason?: string) {
  const result = await request<DissolutionResponse, { requestReason?: string }>(
    `/families/${familyId}/dissolution-requests`,
    { method: 'POST', data: requestReason ? { requestReason } : {} }
  )
  return normalize(result) as DissolutionRequest
}

export async function cancelDissolution(
  familyId: number | string,
  requestId: number | string,
  cancelReason?: string
) {
  const result = await request<DissolutionResponse, { cancelReason?: string }>(
    `/families/${familyId}/dissolution-requests/${requestId}/cancel`,
    { method: 'POST', data: cancelReason ? { cancelReason } : {} }
  )
  return normalize(result) as DissolutionRequest
}
