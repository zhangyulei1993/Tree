import { apiClient, isRealApiMode } from '@/api/client'
import { treeNodes } from '@/mock/data'
import type {
  ApiResponse,
  CreateMemberInput,
  FamilyMember,
  UpdateMemberInput
} from '@/types/api'

let mockMembers: FamilyMember[] = treeNodes.map((node, index) => ({
  memberId: index + 1,
  familyId: 1,
  name: node.name,
  gender: node.gender === '男' ? 'MALE' : node.gender === '女' ? 'FEMALE' : 'UNKNOWN',
  birthYear: Number(node.birthText) || undefined,
  isAlive: !node.deathText,
  description: node.description,
  status: 'ACTIVE',
  userBindingPolicy: 'OPTIONAL',
  boundFamilyRole: index === 0 ? 'FOUNDER' : undefined,
  createdAt: new Date(0).toISOString(),
  updatedAt: new Date(0).toISOString()
}))

export async function listMembers(familyId: number | string): Promise<FamilyMember[]> {
  if (!isRealApiMode) {
    return mockMembers.map((member) => ({ ...member, familyId: Number(familyId) || 1 }))
  }
  const response = await apiClient.get<ApiResponse<FamilyMember[]>>(`/families/${familyId}/members`)
  return response.data.data
}

export async function getMember(
  familyId: number | string,
  memberId: number | string
): Promise<FamilyMember> {
  if (!isRealApiMode) {
    const member = mockMembers.find((item) => item.memberId === Number(memberId))
    if (!member) throw new Error('成员不存在')
    return { ...member, familyId: Number(familyId) || 1 }
  }
  const response = await apiClient.get<ApiResponse<FamilyMember>>(
    `/families/${familyId}/members/${memberId}`
  )
  return response.data.data
}

export async function createMember(
  familyId: number | string,
  input: CreateMemberInput
): Promise<FamilyMember> {
  if (!isRealApiMode) {
    const now = new Date().toISOString()
    const member: FamilyMember = {
      memberId: Math.max(0, ...mockMembers.map((item) => item.memberId)) + 1,
      familyId: Number(familyId) || 1,
      name: input.name,
      gender: input.gender || 'UNKNOWN',
      birthDate: input.birthDate,
      birthYear: input.birthYear,
      deathDate: input.deathDate,
      deathYear: input.deathYear,
      isAlive: input.isAlive ?? true,
      avatarUrl: input.avatarUrl,
      description: input.description,
      status: 'ACTIVE',
      userBindingPolicy: input.userBindingPolicy || 'OPTIONAL',
      createdAt: now,
      updatedAt: now
    }
    mockMembers = [...mockMembers, member]
    return member
  }
  const response = await apiClient.post<ApiResponse<FamilyMember>>(
    `/families/${familyId}/members`,
    input
  )
  return response.data.data
}

export async function updateMember(
  familyId: number | string,
  memberId: number | string,
  input: UpdateMemberInput
): Promise<FamilyMember> {
  if (!isRealApiMode) {
    const member = await getMember(familyId, memberId)
    const updated = { ...member, ...input, updatedAt: new Date().toISOString() }
    mockMembers = mockMembers.map((item) => item.memberId === updated.memberId ? updated : item)
    return updated
  }
  const response = await apiClient.put<ApiResponse<FamilyMember>>(
    `/families/${familyId}/members/${memberId}`,
    input
  )
  return response.data.data
}

export async function deleteMember(
  familyId: number | string,
  memberId: number | string,
  reason?: string
): Promise<void> {
  if (!isRealApiMode) {
    mockMembers = mockMembers.filter((item) => item.memberId !== Number(memberId))
    return
  }
  await apiClient.delete<ApiResponse<{ status: string }>>(
    `/families/${familyId}/members/${memberId}`,
    { data: reason ? { reason } : undefined }
  )
}
