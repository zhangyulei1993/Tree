import { apiClient } from '@/api/client'
import type {
  ApiResponse,
  DissolutionRequest,
  FounderTransferRequest,
  RoleChangeResult
} from '@/types/api'

export async function setFamilyAdmin(familyId: number | string, memberId: number | string) {
  const response = await apiClient.post<ApiResponse<RoleChangeResult>>(
    `/families/${familyId}/members/${memberId}/set-admin`, {}
  )
  return response.data.data
}

export async function unsetFamilyAdmin(familyId: number | string, memberId: number | string) {
  const response = await apiClient.post<ApiResponse<RoleChangeResult>>(
    `/families/${familyId}/members/${memberId}/unset-admin`, {}
  )
  return response.data.data
}

export async function getCurrentFounderTransfer(familyId: number | string) {
  const response = await apiClient.get<ApiResponse<FounderTransferRequest>>(
    `/families/${familyId}/founder-transfer-requests/current`
  )
  return response.data.data
}

export async function createFounderTransfer(
  familyId: number | string,
  toMemberId: number | string,
  requestReason?: string
) {
  const response = await apiClient.post<ApiResponse<FounderTransferRequest>>(
    `/families/${familyId}/founder-transfer-requests`,
    { toMemberId: Number(toMemberId), requestReason: requestReason || undefined }
  )
  return response.data.data
}

export async function cancelFounderTransfer(
  familyId: number | string,
  requestId: number | string,
  cancelReason?: string
) {
  const response = await apiClient.post<ApiResponse<FounderTransferRequest>>(
    `/families/${familyId}/founder-transfer-requests/${requestId}/cancel`,
    { cancelReason: cancelReason || undefined }
  )
  return response.data.data
}

type DissolutionResponse = Omit<DissolutionRequest, 'requestId'> & { id: number | string }
const normalizeDissolution = (value: DissolutionResponse) => ({ ...value, requestId: value.id }) as DissolutionRequest

export async function getCurrentDissolution(familyId: number | string): Promise<DissolutionRequest | null> {
  const response = await apiClient.get<ApiResponse<DissolutionResponse | null>>(
    `/families/${familyId}/dissolution-requests/current`
  )
  return response.data.data ? normalizeDissolution(response.data.data) : null
}

export async function createDissolution(familyId: number | string, requestReason?: string) {
  const response = await apiClient.post<ApiResponse<DissolutionResponse>>(
    `/families/${familyId}/dissolution-requests`, { requestReason: requestReason || undefined }
  )
  return normalizeDissolution(response.data.data)
}

export async function cancelDissolution(
  familyId: number | string,
  requestId: number | string,
  cancelReason?: string
) {
  const response = await apiClient.post<ApiResponse<DissolutionResponse>>(
    `/families/${familyId}/dissolution-requests/${requestId}/cancel`,
    { cancelReason: cancelReason || undefined }
  )
  return normalizeDissolution(response.data.data)
}
