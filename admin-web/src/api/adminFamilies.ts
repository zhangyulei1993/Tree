import { apiClient, unwrapData } from '@/api/client'
import type {
  FamilyPublicStatus,
  RestoreFamilyRequest,
  RestoreFamilyResult,
  TakeDownPublicFamilyRequest
} from '@/types/api'

export async function restoreFamily(familyId: number, input: RestoreFamilyRequest): Promise<RestoreFamilyResult> {
  const response = await apiClient.post(`/admin/families/${familyId}/restore`, input)
  return unwrapData<RestoreFamilyResult>(response)
}

export async function takeDownPublicFamily(familyId: number, input: TakeDownPublicFamilyRequest): Promise<FamilyPublicStatus> {
  const response = await apiClient.post(`/admin/families/${familyId}/take-down-public`, input)
  return unwrapData<FamilyPublicStatus>(response)
}
