import { apiClient, isRealApiMode } from '@/api/client'
import { treeEdges, treeNodes } from '@/mock/data'
import type { ApiResponse, FamilyTreeResult, RelationshipType } from '@/types/api'

export async function getPrivateTree(familyId: number | string): Promise<FamilyTreeResult> {
  if (!isRealApiMode) {
    const idByMockID = new Map(treeNodes.map((node, index) => [node.memberId, index + 1]))
    return {
      familyId: Number(familyId) || 1,
      treeMode: 'LIST_TREE',
      graphVersion: 1,
      nodes: treeNodes.map((node, index) => ({
        memberId: index + 1,
        displayName: node.name,
        gender: node.gender === '男' ? 'MALE' : node.gender === '女' ? 'FEMALE' : 'UNKNOWN',
        memberType: 'LINEAGE_MEMBER',
        birthDate: node.birthText ? `${node.birthText}-01-01` : undefined,
        deathDate: node.deathText ? `${node.deathText}-01-01` : undefined,
        isLiving: !node.deathText,
        userBindingState: 'UNBOUND',
        canExpand: true
      })),
      edges: treeEdges.map((edge, index) => ({
        relationshipId: index + 1,
        fromMemberId: idByMockID.get(edge.sourceMemberId) || 0,
        toMemberId: idByMockID.get(edge.targetMemberId) || 0,
        relationshipType: edge.relationshipType as RelationshipType
      })),
      tree: treeNodes.map((node, index) => ({
        memberId: index + 1,
        parentIds: node.parentIds.map((id) => idByMockID.get(id) || 0),
        childrenIds: node.childrenIds.map((id) => idByMockID.get(id) || 0),
        spouseIds: node.spouseIds.map((id) => idByMockID.get(id) || 0)
      }))
    }
  }
  const response = await apiClient.get<ApiResponse<FamilyTreeResult>>(`/families/${familyId}/tree`)
  return response.data.data
}
