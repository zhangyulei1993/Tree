import { isRealApiMode, request } from '@/api/client'
import { families, myFamilies } from '@/mock/data'
import type { FamilyDetail, FamilySummary } from '@/types/api'

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

export async function getFamilyDetail(familyId: number | string): Promise<FamilyDetail> {
  if (!isRealApiMode) {
    const source = families.find((item) => item.id === String(familyId)) || families[0]
    const mine = myFamilies.find((item) => item.id === source.id)
    return {
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
  }
  return request<FamilyDetail>(`/families/${familyId}`)
}
