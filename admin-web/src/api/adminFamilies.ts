import { apiClient, apiMode, unwrapData } from '@/api/client'
import type {
  FinalizeFamilyResult,
  FamilyPublicStatus,
  RestoreFamilyRequest,
  RestoreFamilyResult,
  TakeDownPublicFamilyRequest
} from '@/types/api'

export async function restoreFamily(familyId: number, input: RestoreFamilyRequest): Promise<RestoreFamilyResult> {
  if (apiMode === 'mock') {
    return {
      familyId,
      status: 'NORMAL',
      publicDisplayStatus: 'PRIVATE',
      searchable: input.searchable ?? true,
      graphVersion: 8,
      restoredAt: '2026-06-29 10:00'
    }
  }
  const response = await apiClient.post(`/admin/families/${familyId}/restore`, input)
  return unwrapData<RestoreFamilyResult>(response)
}

export async function finalizeDissolutionFamily(familyId: number): Promise<FinalizeFamilyResult> {
  if (apiMode === 'mock') {
    return {
      familyId,
      status: 'DISSOLVED',
      publicDisplayStatus: 'PRIVATE',
      searchable: false,
      graphVersion: 8,
      dissolutionCompletedAt: '2026-06-29 10:00'
    }
  }
  const response = await apiClient.post(`/admin/families/${familyId}/finalize-dissolution`)
  return unwrapData<FinalizeFamilyResult>(response)
}

export async function takeDownPublicFamily(familyId: number, input: TakeDownPublicFamilyRequest): Promise<FamilyPublicStatus> {
  if (apiMode === 'mock') {
    return {
      familyId,
      publicDisplayStatus: 'TAKEN_DOWN',
      publicTakenDownAt: '2026-06-29 10:00'
    }
  }
  const response = await apiClient.post(`/admin/families/${familyId}/take-down-public`, input)
  return unwrapData<FamilyPublicStatus>(response)
}
