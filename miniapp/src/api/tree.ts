import { isRealApiMode, request } from '@/api/client'
import { treeNodes } from '@/mock/data'
import type { FamilyTreeResult } from '@/types/api'

export async function getPrivateTree(familyId: number | string): Promise<FamilyTreeResult> {
  if (!isRealApiMode) {
    const idMap = new Map(treeNodes.map((node, index) => [node.memberId, index + 1]))
    const edges: FamilyTreeResult['edges'] = []
    treeNodes.forEach((node) => {
      node.childrenIds.forEach((childID) => {
        edges.push({
          relationshipId: edges.length + 1,
          fromMemberId: idMap.get(node.memberId) || 0,
          toMemberId: idMap.get(childID) || 0,
          relationshipType: 'PARENT_CHILD'
        })
      })
      node.spouseIds.forEach((spouseID) => {
        const from = idMap.get(node.memberId) || 0
        const to = idMap.get(spouseID) || 0
        if (from < to) {
          edges.push({
            relationshipId: edges.length + 1,
            fromMemberId: from,
            toMemberId: to,
            relationshipType: 'SPOUSE'
          })
        }
      })
    })
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
        isLiving: true,
        userBindingState: 'UNBOUND',
        canExpand: true
      })),
      edges,
      tree: treeNodes.map((node, index) => ({
        memberId: index + 1,
        parentIds: node.parentIds.map((id) => idMap.get(id) || 0),
        childrenIds: node.childrenIds.map((id) => idMap.get(id) || 0),
        spouseIds: node.spouseIds.map((id) => idMap.get(id) || 0)
      }))
    }
  }
  return request<FamilyTreeResult>(`/families/${familyId}/tree`)
}

export async function getPublicTree(familyId: number | string): Promise<FamilyTreeResult> {
  if (!isRealApiMode) return getPrivateTree(familyId)
  return request<FamilyTreeResult>(`/public/families/${familyId}/tree`, { public: true })
}
