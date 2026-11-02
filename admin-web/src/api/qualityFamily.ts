import { apiClient, unwrapData } from '@/api/client'
import type { ApiResponse } from '@/types/api'

export interface QualityFamilyEligibilityConfig {
  id: number
  enabled: boolean
  minMemberCount: number
  minRelationshipCount: number
  minBoundMemberCount: number
  minProfileCompleteRate: number
  minBoundMemberRate: number
  requireDescription: boolean
  requireLocation: boolean
  updatedAt: string
}

export async function getQualityFamilyEligibilityConfig() {
  return unwrapData<QualityFamilyEligibilityConfig>(await apiClient.get<ApiResponse<QualityFamilyEligibilityConfig>>('/admin/quality-family-eligibility-config'))
}

export async function updateQualityFamilyEligibilityConfig(input: Omit<QualityFamilyEligibilityConfig, 'id' | 'updatedAt'>) {
  return unwrapData<QualityFamilyEligibilityConfig>(await apiClient.put<ApiResponse<QualityFamilyEligibilityConfig>>('/admin/quality-family-eligibility-config', input))
}
