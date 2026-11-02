import { isRealApiMode, request } from '@/api/client'
import { families } from '@/mock/data'
import type {
  CreateFamilyInput,
  FamilyDetail,
  FamilySummary,
  LeaveFamilyResult,
  ListPublicFamilyShowcaseQuery,
  PaginatedResult,
  PublicFamily,
  PublicFamilyShowcaseItem,
  PublicPlatformStats,
  UpdateFamilyInput
} from '@/types/api'

const mockFamilyOverrides = new Map<string, Partial<FamilyDetail>>()
const mockCreatedFamiliesKey = 'tree_miniapp_mock_created_families'

type MockCreatedFamily = FamilySummary & {
  founderName?: string
  founderGender?: CreateFamilyInput['founderGender']
  founderBirthYear?: number
}

function readMockCreatedFamilies(): MockCreatedFamily[] {
  try {
    const stored = uni.getStorageSync(mockCreatedFamiliesKey)
    if (!stored) return []
    const parsed = typeof stored === 'string' ? JSON.parse(stored) : stored
    return Array.isArray(parsed) ? parsed as MockCreatedFamily[] : []
  } catch {
    uni.removeStorageSync(mockCreatedFamiliesKey)
    return []
  }
}

function writeMockCreatedFamily(family: FamilyDetail, input: CreateFamilyInput) {
  const created = readMockCreatedFamilies().filter((item) => String(item.id) !== String(family.id))
  created.unshift({
    id: family.id,
    familyName: family.familyName,
    familySurname: family.familySurname,
    nativePlace: family.nativePlace,
    regionCode: family.regionCode,
    regionText: family.regionText,
    avatarUrl: family.avatarUrl,
    status: family.status,
    dissolutionCooldownUntil: family.dissolutionCooldownUntil,
    dissolutionCooldownDays: family.dissolutionCooldownDays,
    publicDisplayStatus: family.publicDisplayStatus,
    publicDisplayEnabled: family.publicDisplayEnabled,
    role: family.role,
    founderName: input.founderName,
    founderGender: input.founderGender,
    founderBirthYear: input.founderBirthYear
  })
  uni.setStorageSync(mockCreatedFamiliesKey, JSON.stringify(created))
}

export async function listMyFamilies(): Promise<FamilySummary[]> {
  if (!isRealApiMode) {
    return readMockCreatedFamilies()
  }
  return request<FamilySummary[]>('/families')
}

export async function leaveFamily(
  familyId: number | string,
  reason?: string
): Promise<LeaveFamilyResult> {
  return request<LeaveFamilyResult, { reason?: string }>(`/families/${familyId}/leave`, {
    method: 'POST',
    data: reason ? { reason } : {}
  })
}

export async function getFamilyDetail(familyId: number | string): Promise<FamilyDetail> {
  if (!isRealApiMode) {
    const created = readMockCreatedFamilies().find((item) => String(item.id) === String(familyId))
    const source = families.find((item) => item.id === String(familyId)) || families[0]
    const detail: FamilyDetail = {
      id: created?.id || source.id,
      familyName: created?.familyName || source.name,
      familySurname: created?.familySurname || source.surname,
      nativePlace: created?.nativePlace ?? source.nativePlace,
      regionCode: created?.regionCode ?? null,
      regionText: created?.regionText ?? source.regionText,
      description: source.description,
      status: created?.status || source.status,
      dissolutionCooldownDays: 0,
      publicDisplayStatus: created?.publicDisplayStatus || 'APPROVED',
      publicDisplayEnabled: created?.publicDisplayEnabled ?? true,
      role: created?.role || 'MEMBER',
      searchable: true,
      publicContactVisible: Boolean(source.contact),
      publicContactNote: source.contact || null,
      currentFounderMemberId: 1,
      graphVersion: source.graphVersion
    }
    return { ...detail, ...mockFamilyOverrides.get(String(familyId)) }
  }
  return request<FamilyDetail>(`/families/${familyId}`)
}

export async function createFamily(input: CreateFamilyInput): Promise<FamilyDetail> {
  if (!isRealApiMode) {
    const now = Date.now()
    const created: FamilyDetail = {
      id: `mock_family_${now}`,
      familyName: input.familyName || `${input.surname}氏家族`,
      familySurname: input.surname,
      nativePlace: input.nativePlace || null,
      regionCode: input.regionCode || null,
      regionText: input.regionText || null,
      description: input.description || null,
      status: 'NORMAL',
      dissolutionCooldownDays: 0,
      publicDisplayStatus: 'PRIVATE',
      publicDisplayEnabled: false,
      role: 'FOUNDER',
      searchable: true,
      publicContactVisible: false,
      currentFounderMemberId: now,
      graphVersion: 1
    }
    mockFamilyOverrides.set(String(created.id), created)
    writeMockCreatedFamily(created, input)
    return created
  }
  return request<FamilyDetail, CreateFamilyInput>('/families', {
    method: 'POST',
    data: input
  })
}

export async function updateFamily(
  familyId: number | string,
  input: UpdateFamilyInput
): Promise<FamilyDetail> {
  if (!isRealApiMode) {
    const current = await getFamilyDetail(familyId)
    const updated = { ...current, ...input }
    mockFamilyOverrides.set(String(familyId), updated)
    return updated
  }
  return request<FamilyDetail, UpdateFamilyInput>(`/families/${familyId}`, {
    method: 'PUT',
    data: input
  })
}

export type FamilyArchiveExport = {
  scope: 'INTERNAL' | 'SHARE' | 'PUBLIC'
  generatedAt: string
  family: FamilyDetail
  members: Array<{ memberId: number; name: string; gender?: string; birthYear?: number; deathYear?: number; description?: string }>
  relationships: Array<{ fromMemberId: number; toMemberId: number; relationshipType: string }>
  quality: { incompleteMemberCount: number; missingFields: string[] }
}

export function exportFamilyArchive(familyId: number | string, scope: FamilyArchiveExport['scope']) {
  return request<FamilyArchiveExport>(`/families/${familyId}/archive-export?scope=${scope}`)
}

export async function getPublicFamilyDetail(familyId: number | string): Promise<PublicFamily> {
  if (!isRealApiMode) {
    const source = families.find((item) => item.id === String(familyId)) || families[0]
    return {
      id: source.id,
      familyName: source.name,
      familySurname: source.surname,
      nativePlace: source.nativePlace,
      regionText: source.regionText,
      description: source.description,
      publicContactNote: source.contact || null,
      publicContactVisible: Boolean(source.contact)
    }
  }
  return request<PublicFamily>(`/families/${familyId}/public`, { public: true })
}

function mockPublicFamilyShowcaseItem(source: (typeof families)[number]): PublicFamilyShowcaseItem {
  return {
    id: source.id,
    familyName: source.name,
    familySurname: source.surname,
    nativePlace: source.nativePlace,
    regionText: source.regionText,
    description: source.description,
    publicApprovedAt: null,
    createdAt: '',
    updatedAt: ''
  }
}

export async function listPublicFamilyShowcase(
  query: ListPublicFamilyShowcaseQuery = {}
): Promise<PaginatedResult<PublicFamilyShowcaseItem>> {
  const page = query.page || 1
  const pageSize = query.pageSize || 20

  if (!isRealApiMode) {
    const items = families.map(mockPublicFamilyShowcaseItem)
    const offset = (page - 1) * pageSize
    return {
      items: items.slice(offset, offset + pageSize),
      page,
      pageSize,
      total: items.length
    }
  }

  const params = buildShowcaseQuery(query)
  return request<PaginatedResult<PublicFamilyShowcaseItem>>(`/public/families?${params}`, { public: true })
}

export async function getPublicPlatformStats(): Promise<PublicPlatformStats> {
  if (!isRealApiMode) {
    return {
      familyCount: families.length,
      memberCount: 128,
      relationshipCount: 142
    }
  }
  return request<PublicPlatformStats>('/public/platform-stats', { public: true })
}

function buildShowcaseQuery(query: ListPublicFamilyShowcaseQuery) {
  return buildQuery({
    page: String(query.page || 1),
    pageSize: String(query.pageSize || 20)
  })
}

function buildQuery(params: Record<string, string | undefined>) {
  return Object.entries(params)
    .filter(([, value]) => value !== undefined && value !== '')
    .map(([key, value]) => `${encodeURIComponent(key)}=${encodeURIComponent(value as string)}`)
    .join('&')
}
