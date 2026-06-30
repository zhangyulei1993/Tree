import { request } from '@/api/client'
import type {
  CreateRelationshipInput,
  PlaceExistingMemberInput,
  RelationshipMutationResult,
  UpdateRelationshipInput
} from '@/types/api'

export function createRelationship(
  familyId: number | string,
  input: CreateRelationshipInput
) {
  return request<RelationshipMutationResult, CreateRelationshipInput>(
    `/families/${familyId}/relationships`,
    { method: 'POST', data: input }
  )
}

export function placeExistingMember(
  familyId: number | string,
  input: PlaceExistingMemberInput
) {
  return request<RelationshipMutationResult, PlaceExistingMemberInput>(
    `/families/${familyId}/relationships/place-existing`,
    { method: 'POST', data: input }
  )
}

export function updateRelationship(
  familyId: number | string,
  relationshipId: number | string,
  input: UpdateRelationshipInput
) {
  return request<RelationshipMutationResult, UpdateRelationshipInput>(
    `/families/${familyId}/relationships/${relationshipId}`,
    { method: 'PUT', data: input }
  )
}

export function deleteRelationship(
  familyId: number | string,
  relationshipId: number | string,
  reason?: string
) {
  return request<RelationshipMutationResult, { reason?: string }>(
    `/families/${familyId}/relationships/${relationshipId}`,
    { method: 'DELETE', data: reason ? { reason } : {} }
  )
}
