import { isRealApiMode, request } from '@/api/client'
import { families, myFamilies } from '@/mock/data'
import type {
  CreateFamilyInput,
  FamilyDetail,
  FamilySummary,
  ListPublicFamiliesQuery,
  LeaveFamilyResult,
  PaginatedResult,
  PublicFamily,
  PublicFamilyListItem,
  UpdateFamilyInput
} from '@/types/api'

const mockFamilyOverrides = new Map<string, Partial<FamilyDetail>>()

export async function listMyFamilies(): Promise<FamilySummary[]> {
  if (!isRealApiMode) {
    return myFamilies.map((family) => {
      const source = families.find((item) => item.id === family.id) || families[0]
      return {
        id: family.id,
        familyName: family.name,
        familySurname: source.surname,
        nativePlace: source.nativePlace,
        regionText: source.regionText,
        status: family.status,
        publicDisplayStatus: 'APPROVED',
        role: family.role
      }
    })
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
    const source = families.find((item) => item.id === String(familyId)) || families[0]
    const mine = myFamilies.find((item) => item.id === source.id)
    const detail: FamilyDetail = {
      id: source.id,
      familyName: source.name,
      familySurname: source.surname,
      nativePlace: source.nativePlace,
      regionText: source.regionText,
      description: source.description,
      status: source.status,
      publicDisplayStatus: 'APPROVED',
      role: mine?.role || 'MEMBER',
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
    return {
      id: `mock_family_${now}`,
      familyName: input.familyName || `${input.surname}氏家族`,
      familySurname: input.surname,
      nativePlace: input.nativePlace || null,
      regionText: input.regionText || null,
      description: input.description || null,
      status: 'NORMAL',
      publicDisplayStatus: 'PRIVATE',
      role: 'FOUNDER',
      searchable: true,
      publicContactVisible: false,
      currentFounderMemberId: now,
      graphVersion: 1
    }
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

function mockPublicFamilyListItem(source: (typeof families)[number]): PublicFamilyListItem {
  const hasContact = Boolean(source.contact)
  return {
    id: source.id,
    familyName: source.name,
    familySurname: source.surname,
    nativePlace: source.nativePlace,
    regionText: source.regionText,
    description: source.description,
    publicContactVisible: hasContact,
    publicContactNote: hasContact ? source.contact : null,
    publicApprovedAt: null,
    createdAt: '',
    updatedAt: ''
  }
}

export async function listPublicFamilies(
  query: ListPublicFamiliesQuery = {}
): Promise<PaginatedResult<PublicFamilyListItem>> {
  const page = query.page || 1
  const pageSize = query.pageSize || 20

  if (!isRealApiMode) {
    let items = families.map(mockPublicFamilyListItem)
    const keyword = query.keyword?.trim()
    const familySurname = query.familySurname?.trim()
    const regionText = query.regionText?.trim()

    if (keyword) {
      items = items.filter((item) =>
        [item.familyName, item.familySurname, item.nativePlace, item.regionText, item.description]
          .filter(Boolean)
          .some((value) => String(value).includes(keyword))
      )
    }
    if (familySurname) {
      items = items.filter((item) => item.familySurname.includes(familySurname))
    }
    if (regionText) {
      items = items.filter((item) => (item.regionText || '').includes(regionText))
    }

    const offset = (page - 1) * pageSize
    return {
      items: items.slice(offset, offset + pageSize),
      page,
      pageSize,
      total: items.length
    }
  }

  const params = buildPublicFamiliesQuery(query)
  return request<PaginatedResult<PublicFamilyListItem>>(`/public/families?${params}`, { public: true })
}

function buildPublicFamiliesQuery(query: ListPublicFamiliesQuery) {
  return buildQuery({
    page: String(query.page || 1),
    pageSize: String(query.pageSize || 20),
    keyword: query.keyword?.trim() || undefined,
    familySurname: query.familySurname?.trim() || undefined,
    regionText: query.regionText?.trim() || undefined
  })
}

function buildQuery(params: Record<string, string | undefined>) {
  return Object.entries(params)
    .filter(([, value]) => value !== undefined && value !== '')
    .map(([key, value]) => `${encodeURIComponent(key)}=${encodeURIComponent(value as string)}`)
    .join('&')
}
