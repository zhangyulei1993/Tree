import { isRealApiMode, request } from '@/api/client'
import { treeNodes } from '@/mock/data'
import type { FamilyMember } from '@/types/api'

export async function listFamilyMembers(familyId: number | string): Promise<FamilyMember[]> {
  if (!isRealApiMode) {
    return treeNodes.map((node, index) => ({
      memberId: index + 1,
      familyId: Number(familyId) || 1,
      name: node.name,
      gender: node.gender === '男' ? 'MALE' : node.gender === '女' ? 'FEMALE' : 'UNKNOWN',
      birthYear: Number(node.birthText) || undefined,
      isAlive: true,
      description: node.description,
      status: 'ACTIVE',
      userBindingPolicy: 'OPTIONAL',
      boundUserId: index === 0 ? 1 : undefined,
      boundFamilyRole: index === 0 ? 'FOUNDER' : undefined,
      createdAt: new Date(0).toISOString(),
      updatedAt: new Date(0).toISOString()
    }))
  }
  return request<FamilyMember[]>(`/families/${familyId}/members`)
}
