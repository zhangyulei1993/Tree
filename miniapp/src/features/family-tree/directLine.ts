import { collectChildren } from './graph'
import type { MemberGraph } from './types'

function parentsOf(memberId: number, graph: MemberGraph): number[] {
  return graph.parentIds.get(memberId) || []
}

function childrenOf(memberId: number, graph: MemberGraph): number[] {
  return collectChildren(memberId, graph.childrenIds, graph.spouseIds)
}

/** 直系血亲：祖先 + 本人 + 后代（不含配偶） */
export function collectDirectLineMemberIds(viewerId: number, graph: MemberGraph): Set<number> {
  const line = new Set<number>([viewerId])

  function walkUp(id: number, visiting: Set<number>) {
    for (const parentId of parentsOf(id, graph)) {
      if (visiting.has(parentId)) continue
      visiting.add(parentId)
      line.add(parentId)
      walkUp(parentId, visiting)
    }
  }

  function walkDown(id: number, visiting: Set<number>) {
    for (const childId of childrenOf(id, graph)) {
      if (visiting.has(childId)) continue
      visiting.add(childId)
      line.add(childId)
      walkDown(childId, visiting)
    }
  }

  walkUp(viewerId, new Set())
  walkDown(viewerId, new Set())
  return line
}

export function isParentChildLink(parentId: number, childId: number, graph: MemberGraph): boolean {
  return parentsOf(childId, graph).includes(parentId)
}
