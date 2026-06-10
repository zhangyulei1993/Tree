import { apiClient, unwrapData } from '@/api/client'
import type { PageResult, PublicApplication, ReviewRequest } from '@/types/api'

export interface PublicApplicationQuery {
  status?: string
  page?: number
  pageSize?: number
}

export async function listPublicApplications(query: PublicApplicationQuery): Promise<PageResult<PublicApplication>> {
  const response = await apiClient.get('/admin/family-public-applications', { params: query })
  return unwrapData<PageResult<PublicApplication>>(response)
}

export async function approvePublicApplication(applicationId: number, input: ReviewRequest): Promise<PublicApplication> {
  const response = await apiClient.post(`/admin/family-public-applications/${applicationId}/approve`, input)
  return unwrapData<PublicApplication>(response)
}

export async function rejectPublicApplication(applicationId: number, input: ReviewRequest): Promise<PublicApplication> {
  const response = await apiClient.post(`/admin/family-public-applications/${applicationId}/reject`, input)
  return unwrapData<PublicApplication>(response)
}
