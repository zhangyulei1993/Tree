import { apiClient, isRealApiMode } from '@/api/client'
import type {
  ApiResponse,
  CancelPublicApplicationInput,
  FamilyPublicStatus,
  PaginatedResult,
  PublicApplication,
  PublicApplicationQuery,
  SubmitPublicApplicationInput,
  TakeDownPublicFamilyInput
} from '@/types/api'

const storageKey = 'tree_web_mock_public_applications'

function readMockApplications(): PublicApplication[] {
  try {
    return JSON.parse(sessionStorage.getItem(storageKey) || '[]') as PublicApplication[]
  } catch {
    sessionStorage.removeItem(storageKey)
    return []
  }
}

function writeMockApplications(applications: PublicApplication[]) {
  sessionStorage.setItem(storageKey, JSON.stringify(applications))
}

export async function submitPublicApplication(
  familyId: number | string,
  input: SubmitPublicApplicationInput
): Promise<FamilyPublicStatus> {
  if (!isRealApiMode) {
    const now = new Date().toISOString()
    const applications = readMockApplications()
    applications.unshift({
      applicationId: Date.now(),
      familyId: Number(familyId) || 1,
      status: 'PENDING',
      reason: input.applicationReason || null,
      createdAt: now,
      updatedAt: now
    })
    writeMockApplications(applications)
    return {
      familyId: Number(familyId) || 1,
      publicDisplayStatus: 'PENDING',
      publicAppliedAt: now
    }
  }
  const response = await apiClient.post<ApiResponse<FamilyPublicStatus>>(
    `/families/${familyId}/public-applications`,
    input
  )
  return response.data.data
}

export async function listPublicApplications(
  familyId: number | string,
  query: PublicApplicationQuery = {}
): Promise<PaginatedResult<PublicApplication>> {
  if (!isRealApiMode) {
    const page = query.page || 1
    const pageSize = query.pageSize || 20
    const matching = readMockApplications().filter((application) => {
      const sameFamily = String(application.familyId) === String(Number(familyId) || 1)
      return sameFamily && (!query.status || application.status === query.status)
    })
    const offset = (page - 1) * pageSize
    return {
      items: matching.slice(offset, offset + pageSize),
      page,
      pageSize,
      total: matching.length
    }
  }
  const response = await apiClient.get<ApiResponse<PaginatedResult<PublicApplication>>>(
    `/families/${familyId}/public-applications`,
    { params: query }
  )
  return response.data.data
}

export async function cancelPublicApplication(
  familyId: number | string,
  applicationId: number | string,
  input: CancelPublicApplicationInput = {}
): Promise<PublicApplication> {
  if (!isRealApiMode) {
    const applications = readMockApplications()
    const application = applications.find(
      (item) =>
        String(item.familyId) === String(Number(familyId) || 1) &&
        String(item.applicationId) === String(applicationId)
    )
    if (!application || application.status !== 'PENDING') {
      throw new Error('公开申请不存在或状态不可取消。')
    }
    const now = new Date().toISOString()
    application.status = 'CANCELLED'
    application.cancelReason = input.cancelReason || null
    application.cancelledAt = now
    application.updatedAt = now
    writeMockApplications(applications)
    return application
  }
  const response = await apiClient.post<ApiResponse<PublicApplication>>(
    `/families/${familyId}/public-applications/${applicationId}/cancel`,
    input
  )
  return response.data.data
}

export async function closePublicFamily(
  familyId: number | string,
  input: TakeDownPublicFamilyInput = {}
): Promise<FamilyPublicStatus> {
  if (!isRealApiMode) {
    return {
      familyId: Number(familyId) || 1,
      publicDisplayStatus: 'TAKEN_DOWN',
      publicTakenDownAt: new Date().toISOString()
    }
  }
  const response = await apiClient.post<ApiResponse<FamilyPublicStatus>>(
    `/families/${familyId}/take-down-public`,
    input
  )
  return response.data.data
}
