import { apiClient, apiMode, unwrapData } from '@/api/client'
import { listMockPublicApplications } from '@/api/management'
import type { PageResult, PublicApplication, ReviewRequest } from '@/types/api'

export interface PublicApplicationQuery {
  status?: string
  page?: number
  pageSize?: number
}

export async function listPublicApplications(query: PublicApplicationQuery): Promise<PageResult<PublicApplication>> {
  if (apiMode === 'mock') {
    return listMockPublicApplications(query)
  }
  const response = await apiClient.get('/admin/family-public-applications', { params: query })
  return unwrapData<PageResult<PublicApplication>>(response)
}

export async function approvePublicApplication(applicationId: number, input: ReviewRequest): Promise<PublicApplication> {
  if (apiMode === 'mock') {
    return {
      applicationId,
      familyId: 1,
      familyName: '示例家庭',
      familyPublicDisplayStatus: 'APPROVED',
      status: 'APPROVED',
      reviewComment: input.reviewComment,
      createdAt: '2026-06-29 10:00',
      updatedAt: '2026-06-29 10:00'
    } as PublicApplication
  }
  const response = await apiClient.post(`/admin/family-public-applications/${applicationId}/approve`, input)
  return unwrapData<PublicApplication>(response)
}

export async function rejectPublicApplication(applicationId: number, input: ReviewRequest): Promise<PublicApplication> {
  if (apiMode === 'mock') {
    return {
      applicationId,
      familyId: 1,
      familyName: '示例家庭',
      familyPublicDisplayStatus: 'PRIVATE',
      status: 'REJECTED',
      reviewComment: input.reviewComment,
      createdAt: '2026-06-29 10:00',
      updatedAt: '2026-06-29 10:00'
    } as PublicApplication
  }
  const response = await apiClient.post(`/admin/family-public-applications/${applicationId}/reject`, input)
  return unwrapData<PublicApplication>(response)
}
