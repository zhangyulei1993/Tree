import { isRealApiMode, request } from '@/api/client'
import { treeNodes } from '@/mock/data'
import type {
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
  isAlive: true,
  description: node.description,
  memorialVisible: true,
  status: 'ACTIVE',
  userBindingPolicy: 'OPTIONAL',
  boundUserId: index === 0 ? 1 : undefined,
  boundFamilyRole: index === 0 ? 'FOUNDER' : undefined,
  createdAt: new Date(0).toISOString(),
  updatedAt: new Date(0).toISOString()
}))

export async function listFamilyMembers(familyId: number | string): Promise<FamilyMember[]> {
  if (!isRealApiMode) {
    return mockMembers.map((member) => ({ ...member, familyId: Number(familyId) || 1 }))
  }
  return request<FamilyMember[]>(`/families/${familyId}/members`)
}

export async function getFamilyMember(
  familyId: number | string,
  memberId: number | string
): Promise<FamilyMember> {
  if (!isRealApiMode) {
    const member = mockMembers
      .find((item) => item.memberId === Number(memberId))
    if (!member) throw new Error('成员不存在')
    return { ...member, familyId: Number(familyId) || 1 }
  }
  return request<FamilyMember>(`/families/${familyId}/members/${memberId}`)
}

export async function createFamilyMember(
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
      memorialVisible: input.memorialVisible ?? true,
      status: 'ACTIVE',
      userBindingPolicy: input.userBindingPolicy || 'OPTIONAL',
      createdAt: now,
      updatedAt: now
    }
    mockMembers = [...mockMembers, member]
    return { ...member }
  }
  return request<FamilyMember, CreateMemberInput>(`/families/${familyId}/members`, {
    method: 'POST',
    data: input
  })
}

export async function updateFamilyMember(
  familyId: number | string,
  memberId: number | string,
  input: UpdateMemberInput
): Promise<FamilyMember> {
  if (!isRealApiMode) {
    const current = await getFamilyMember(familyId, memberId)
    const updated: FamilyMember = {
      ...current,
      ...input,
      gender: input.gender || current.gender,
      updatedAt: new Date().toISOString()
    }
    mockMembers = mockMembers.map((member) => member.memberId === updated.memberId ? updated : member)
    return { ...updated }
  }
  return request<FamilyMember, UpdateMemberInput>(`/families/${familyId}/members/${memberId}`, {
    method: 'PUT',
    data: input
  })
}

export async function deleteFamilyMember(
  familyId: number | string,
  memberId: number | string,
  reason?: string
): Promise<{ status: string }> {
  if (!isRealApiMode) {
    mockMembers = mockMembers.filter((member) => member.memberId !== Number(memberId))
    return { status: 'ok' }
  }
  return request<{ status: string }, { reason?: string }>(
    `/families/${familyId}/members/${memberId}`,
    { method: 'DELETE', data: reason ? { reason } : {} }
  )
}

export async function unbindFamilyMemberUser(
  familyId: number | string,
  memberId: number | string,
  reason?: string
): Promise<FamilyMember> {
  if (!isRealApiMode) {
    const current = await getFamilyMember(familyId, memberId)
    const updated: FamilyMember = {
      ...current,
      boundUserId: undefined,
      boundFamilyRole: undefined,
      updatedAt: new Date().toISOString()
    }
    mockMembers = mockMembers.map((member) => member.memberId === updated.memberId ? updated : member)
    return { ...updated }
  }
  return request<FamilyMember, { reason?: string }>(
    `/families/${familyId}/members/${memberId}/unbind-user`,
    { method: 'POST', data: reason ? { reason } : {} }
  )
}
