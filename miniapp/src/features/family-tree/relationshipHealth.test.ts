import assert from 'node:assert/strict'
import test from 'node:test'

import { inspectRelationshipHealth } from './relationshipHealth'
import type { FamilyTreeResult, TreeEdge, TreeNode } from '@/types/api'

function member(memberId: number, displayName: string, birthDate?: string): TreeNode {
  return { memberId, displayName, gender: 'MALE', memberType: 'LINEAGE_MEMBER', birthDate, userBindingState: 'NOT_REQUIRED', canExpand: false }
}

function edge(relationshipType: TreeEdge['relationshipType'], fromMemberId: number, toMemberId: number, relationshipId: number): TreeEdge {
  return { relationshipType, fromMemberId, toMemberId, relationshipId }
}

test('parent with children but no spouse is prompted to supplement spouse information', () => {
  const tree: FamilyTreeResult = { familyId: 1, treeMode: 'PRIVATE', graphVersion: 1, tree: [], nodes: [member(1, '张父'), member(2, '张子')], edges: [edge('PARENT_CHILD', 1, 2, 1)] }
  const result = inspectRelationshipHealth(tree)
  assert.deepEqual(result.membersWithChildrenNoSpouse.map((item) => item.memberId), [1])
  assert.deepEqual(result.hintsByMemberId[1], ['有子女，待补配偶'])
})

test('parent with a recorded spouse does not receive a missing spouse prompt', () => {
  const tree: FamilyTreeResult = { familyId: 1, treeMode: 'PRIVATE', graphVersion: 1, tree: [], nodes: [member(1, '张父'), member(2, '李母'), member(3, '张子')], edges: [edge('PARENT_CHILD', 1, 3, 1), edge('SPOUSE', 1, 2, 2)] }
  assert.equal(inspectRelationshipHealth(tree).membersWithChildrenNoSpouse.length, 0)
})

test('parent and child with impossible birth-year order are prompted to verify their dates', () => {
  const tree: FamilyTreeResult = {
    familyId: 1,
    treeMode: 'PRIVATE',
    graphVersion: 1,
    tree: [],
    nodes: [member(1, '张父', '2000-01-01'), member(2, '张子', '1999-01-01')],
    edges: [edge('PARENT_CHILD', 1, 2, 1)]
  }
  const result = inspectRelationshipHealth(tree)
  assert.deepEqual(result.membersWithBirthOrderConflict.map((item) => item.memberId), [1, 2])
  assert.deepEqual(result.hintsByMemberId[1], ['有子女，待补配偶', '亲子出生年份待核对'])
  assert.deepEqual(result.hintsByMemberId[2], ['亲子出生年份待核对'])
})
