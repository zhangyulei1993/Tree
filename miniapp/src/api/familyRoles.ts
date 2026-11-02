import { request } from '@/api/client'
import type { RoleChangeResult } from '@/types/api'

export function setFamilyAdmin(
  familyId: number | string,
  memberId: number | string,
  reason?: string
) {
  return request<RoleChangeResult, { reason?: string }>(
    `/families/${familyId}/members/${memberId}/set-admin`,
    { method: 'POST', data: reason ? { reason } : {} }
  )
}

export function unsetFamilyAdmin(
  familyId: number | string,
  memberId: number | string,
  reason?: string
) {
  return request<RoleChangeResult, { reason?: string }>(
    `/families/${familyId}/members/${memberId}/unset-admin`,
    { method: 'POST', data: reason ? { reason } : {} }
  )
}
