import { apiClient, unwrapData } from '@/api/client'
import type { ShowcaseFamily, ShowcaseFamilyDisplaySettings, ShowcaseFamilyTree, ShowcaseFamilyTreeMember, ShowcaseFamilyTreeRelationship } from '@/types/api'

export interface ShowcaseFamilyInput {
  displayName: string
  regionText: string
  summary: string
  coverImageUrl: string
  treeImageUrl: string
  sortOrder: number
}

export interface ShowcaseFamilyMemberInput {
  displayName: string
  identityText?: string
  memberType?: 'LINEAGE_MEMBER' | 'SPOUSE'
  gender: 'MALE' | 'FEMALE' | 'UNKNOWN'
  birthYear?: number | null
  sortOrder: number
}

export interface ShowcaseFamilyRelationshipInput {
  fromMemberId: number
  toMemberId: number
  relationshipType: 'PARENT_CHILD' | 'SPOUSE'
}

export interface AddShowcaseRelativeInput {
  relationKind: 'CHILD' | 'SPOUSE'
  member: ShowcaseFamilyMemberInput
  existingSpouseMemberId?: number
  newSpouse?: ShowcaseFamilyMemberInput
}

export async function listShowcaseFamilies() {
  const response = await apiClient.get('/admin/showcase-families')
  return unwrapData<ShowcaseFamily[]>(response)
}

export async function getShowcaseFamilyDisplaySettings() {
  return unwrapData<ShowcaseFamilyDisplaySettings>(await apiClient.get('/admin/showcase-family-display-settings'))
}

export async function updateShowcaseFamilyDisplaySettings(rotationDays: number) {
  return unwrapData<ShowcaseFamilyDisplaySettings>(await apiClient.put('/admin/showcase-family-display-settings', { rotationDays }))
}

export async function createShowcaseFamily(input: ShowcaseFamilyInput) {
  const response = await apiClient.post('/admin/showcase-families', input)
  return unwrapData<ShowcaseFamily>(response)
}

export async function updateShowcaseFamily(id: number, input: ShowcaseFamilyInput) {
  const response = await apiClient.put(`/admin/showcase-families/${id}`, input)
  return unwrapData<ShowcaseFamily>(response)
}

export async function setShowcaseFamilyPublished(id: number, publish: boolean) {
  const endpoint = publish ? 'publish' : 'unpublish'
  const response = await apiClient.post(`/admin/showcase-families/${id}/${endpoint}`)
  return unwrapData<ShowcaseFamily>(response)
}

export async function deleteShowcaseFamily(id: number) {
  await apiClient.delete(`/admin/showcase-families/${id}`)
}

export async function getShowcaseFamilyTree(id: number) {
  const response = await apiClient.get(`/admin/showcase-families/${id}/tree`)
  return unwrapData<ShowcaseFamilyTree>(response)
}

export async function createShowcaseFamilyMember(showcaseFamilyId: number, input: ShowcaseFamilyMemberInput): Promise<ShowcaseFamilyTreeMember> {
  return unwrapData<ShowcaseFamilyTreeMember>(await apiClient.post(`/admin/showcase-families/${showcaseFamilyId}/members`, input))
}

export async function updateShowcaseFamilyMember(showcaseFamilyId: number, memberId: number, input: ShowcaseFamilyMemberInput): Promise<ShowcaseFamilyTreeMember> {
  return unwrapData<ShowcaseFamilyTreeMember>(await apiClient.put(`/admin/showcase-families/${showcaseFamilyId}/members/${memberId}`, input))
}

export async function deleteShowcaseFamilyMember(showcaseFamilyId: number, memberId: number) {
  await apiClient.delete(`/admin/showcase-families/${showcaseFamilyId}/members/${memberId}`)
}

export async function createShowcaseFamilyRelationship(showcaseFamilyId: number, input: ShowcaseFamilyRelationshipInput): Promise<ShowcaseFamilyTreeRelationship> {
  return unwrapData<ShowcaseFamilyTreeRelationship>(await apiClient.post(`/admin/showcase-families/${showcaseFamilyId}/relationships`, input))
}

export async function deleteShowcaseFamilyRelationship(showcaseFamilyId: number, relationshipId: number) {
  await apiClient.delete(`/admin/showcase-families/${showcaseFamilyId}/relationships/${relationshipId}`)
}

export async function addShowcaseRelative(showcaseFamilyId: number, memberId: number, input: AddShowcaseRelativeInput) {
  return unwrapData(await apiClient.post(`/admin/showcase-families/${showcaseFamilyId}/members/${memberId}/relatives`, input))
}
