import { request } from '@/api/client'
import type { FamilyPublicStatus, PaginatedResult, PublicApplication } from '@/types/api'

export function listPublicApplications(familyId: number | string) {
  return request<PaginatedResult<PublicApplication>>(
    `/families/${familyId}/public-applications?page=1&pageSize=50`
  )
}

export function submitPublicApplication(familyId: number | string, reason?: string) {
  return request<PublicApplication, { applicationReason?: string }>(
    `/families/${familyId}/public-applications`,
    { method: 'POST', data: reason ? { applicationReason: reason } : {} }
  )
}

export function cancelPublicApplication(
  familyId: number | string,
  applicationId: number | string,
  reason?: string
) {
  return request<PublicApplication, { cancelReason?: string }>(
    `/families/${familyId}/public-applications/${applicationId}/cancel`,
    { method: 'POST', data: reason ? { cancelReason: reason } : {} }
  )
}

export function closePublicFamily(familyId: number | string, reason?: string) {
  return request<FamilyPublicStatus, { reason?: string }>(
    `/families/${familyId}/take-down-public`,
    { method: 'POST', data: reason ? { reason } : {} }
  )
}
