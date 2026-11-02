import { apiClient, isRealApiMode } from '@/api/client'
import { createMember } from '@/api/members'
import type {
  ApiResponse,
  CreateRelationshipInput,
  PlaceExistingMemberInput,
  Relationship,
  RelationshipMutationResult,
  UpdateRelationshipInput
} from '@/types/api'

let mockRelationshipSequence = 100
let mockRelationships: Relationship[] = []

export async function createRelationship(
  familyId: number | string,
  input: CreateRelationshipInput
): Promise<RelationshipMutationResult> {
  if (!isRealApiMode) {
    mockRelationshipSequence += 1
    const member = await createMember(familyId, input.newMember)
    const isSpouse = input.addType === 'ADD_SPOUSE'
    const relationship: Relationship = {
      relationshipId: mockRelationshipSequence,
      familyId: Number(familyId) || 1,
      fromMemberId: input.addType === 'ADD_FATHER' || input.addType === 'ADD_MOTHER'
        ? member.memberId
        : input.baseMemberId,
      toMemberId: input.addType === 'ADD_FATHER' || input.addType === 'ADD_MOTHER'
        ? input.baseMemberId
        : member.memberId,
      relationshipType: isSpouse ? 'SPOUSE' : 'PARENT_CHILD',
      parentLinkType: isSpouse ? null : input.relationship.parentLinkType || 'PRIMARY',
      status: 'ACTIVE',
      createdAt: new Date().toISOString(),
      updatedAt: new Date().toISOString()
    }
    mockRelationships = [...mockRelationships, relationship]
    return {
      createdMember: {
        memberId: member.memberId,
        familyId: member.familyId,
        name: member.name,
        gender: member.gender,
        status: member.status
      },
      relationships: [relationship],
      graphVersion: 2
    }
  }
  const response = await apiClient.post<ApiResponse<RelationshipMutationResult>>(
    `/families/${familyId}/relationships`,
    input
  )
  return response.data.data
}

export async function placeExistingMember(
  familyId: number | string,
  input: PlaceExistingMemberInput
): Promise<RelationshipMutationResult> {
  const response = await apiClient.post<ApiResponse<RelationshipMutationResult>>(
    `/families/${familyId}/relationships/place-existing`, input
  )
  return response.data.data
}

export async function updateRelationship(
  familyId: number | string,
  relationshipId: number | string,
  input: UpdateRelationshipInput
): Promise<RelationshipMutationResult> {
  if (!isRealApiMode) {
    const relationship = mockRelationships.find((item) => item.relationshipId === Number(relationshipId))
    if (!relationship) throw new Error('关系不存在')
    const updated = { ...relationship, ...input, updatedAt: new Date().toISOString() }
    mockRelationships = mockRelationships.map((item) => item.relationshipId === updated.relationshipId ? updated : item)
    return { relationships: [updated], graphVersion: 3 }
  }
  const response = await apiClient.put<ApiResponse<RelationshipMutationResult>>(
    `/families/${familyId}/relationships/${relationshipId}`,
    input
  )
  return response.data.data
}

export async function deleteRelationship(
  familyId: number | string,
  relationshipId: number | string,
  reason?: string
): Promise<RelationshipMutationResult> {
  if (!isRealApiMode) {
    const relationship = mockRelationships.find((item) => item.relationshipId === Number(relationshipId))
    if (!relationship) throw new Error('关系不存在')
    const deleted = {
      ...relationship,
      status: 'DELETED',
      deletedAt: new Date().toISOString(),
      updatedAt: new Date().toISOString()
    }
    mockRelationships = mockRelationships.map((item) => item.relationshipId === deleted.relationshipId ? deleted : item)
    return { relationships: [deleted], graphVersion: 3 }
  }
  const response = await apiClient.delete<ApiResponse<RelationshipMutationResult>>(
    `/families/${familyId}/relationships/${relationshipId}`,
    { data: reason ? { reason } : undefined }
  )
  return response.data.data
}
