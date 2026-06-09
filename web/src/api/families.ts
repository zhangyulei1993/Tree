import { apiClient, isRealApiMode } from '@/api/client'
import { myFamilies, publicFamilies } from '@/mock/data'
import type {
  ApiResponse,
  CreateFamilyInput,
  FamilyDetail,
  FamilySummary,
  UpdateFamilyInput
} from '@/types/api'

const mockFamiliesStorageKey = 'tree_web_mock_families'

function restoredMockFamilies() {
  try {
    const value = sessionStorage.getItem(mockFamiliesStorageKey)
    const families = value ? JSON.parse(value) as FamilyDetail[] : []
    return new Map(families.map((family) => [String(family.id), family]))
  } catch {
    sessionStorage.removeItem(mockFamiliesStorageKey)
    return new Map<string, FamilyDetail>()
  }
}

const createdMockFamilies = restoredMockFamilies()
let mockFamilySequence = Math.max(
  100,
  ...[...createdMockFamilies.values()].map((family) => Number(family.id) || 0)
)

function persistMockFamilies() {
  sessionStorage.setItem(
    mockFamiliesStorageKey,
    JSON.stringify([...createdMockFamilies.values()])
  )
}

export async function listMyFamilies(): Promise<FamilySummary[]> {
  if (!isRealApiMode) {
    const existing = myFamilies.map((family) => ({
      id: family.id,
      familyName: family.name,
      familySurname: family.name.slice(0, 1),
      status: family.status,
      publicDisplayStatus: 'APPROVED',
      role: family.role
    }))
    const created = [...createdMockFamilies.values()].map((family) => ({
      id: family.id,
      familyName: family.familyName,
      familySurname: family.familySurname,
      nativePlace: family.nativePlace,
      regionText: family.regionText,
      status: family.status,
      publicDisplayStatus: family.publicDisplayStatus,
      role: family.role
    }))
    return [...created, ...existing]
  }
  const response = await apiClient.get<ApiResponse<FamilySummary[]>>('/families')
  return response.data.data
}

export async function createFamily(input: CreateFamilyInput): Promise<FamilyDetail> {
  if (!isRealApiMode) {
    mockFamilySequence += 1
    const family: FamilyDetail = {
      id: mockFamilySequence,
      familyName: input.familyName || `${input.surname}氏家族`,
      familySurname: input.surname,
      nativePlace: input.nativePlace,
      regionText: input.regionText,
      description: input.description,
      status: 'NORMAL',
      publicDisplayStatus: 'PRIVATE',
      role: 'FOUNDER',
      searchable: true,
      publicContactVisible: false,
      currentFounderMemberId: mockFamilySequence,
      graphVersion: 1
    }
    createdMockFamilies.set(String(family.id), family)
    persistMockFamilies()
    return family
  }
  const response = await apiClient.post<ApiResponse<FamilyDetail>>('/families', input)
  return response.data.data
}

export async function getFamilyDetail(familyId: number | string): Promise<FamilyDetail> {
  if (!isRealApiMode) {
    const created = createdMockFamilies.get(String(familyId))
    if (created) return created
    const source = publicFamilies.find((family) => family.id === String(familyId)) || publicFamilies[0]
    const mine = myFamilies.find((family) => family.id === String(familyId))
    return {
      id: source.id,
      familyName: source.name,
      familySurname: source.surname,
      nativePlace: source.nativePlace,
      regionText: source.regionText,
      description: source.description,
      status: source.publicStatus === 'APPROVED' ? 'NORMAL' : 'NORMAL',
      publicDisplayStatus: source.publicStatus,
      role: mine?.role || 'MEMBER',
      searchable: true,
      publicContactVisible: Boolean(source.publicContact),
      publicContactNote: source.publicContact || null,
      currentFounderMemberId: 1,
      graphVersion: 1
    }
  }
  const response = await apiClient.get<ApiResponse<FamilyDetail>>(`/families/${familyId}`)
  return response.data.data
}

export async function updateFamily(
  familyId: number | string,
  input: UpdateFamilyInput
): Promise<FamilyDetail> {
  if (!isRealApiMode) {
    const current = await getFamilyDetail(familyId)
    const updated = { ...current, ...input }
    createdMockFamilies.set(String(familyId), updated)
    persistMockFamilies()
    return updated
  }
  const response = await apiClient.put<ApiResponse<FamilyDetail>>(`/families/${familyId}`, input)
  return response.data.data
}
