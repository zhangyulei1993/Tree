import { apiClient, isRealApiMode } from '@/api/client'
import { myFamilies } from '@/mock/data'
import type { ApiResponse, FamilySummary } from '@/types/api'

export async function listMyFamilies(): Promise<FamilySummary[]> {
  if (!isRealApiMode) {
    return myFamilies.map((family) => ({
      id: family.id,
      familyName: family.name,
      familySurname: family.name.slice(0, 1),
      status: family.status,
      publicDisplayStatus: 'APPROVED',
      role: family.role
    }))
  }
  const response = await apiClient.get<ApiResponse<FamilySummary[]>>('/families')
  return response.data.data
}
