import assert from 'node:assert/strict'
import test from 'node:test'

import type { TreeEdge, TreeNode } from '@/types/api'

import { bloodGenerationGap, buildRelGraph } from './relativeGraph'

function node(memberId: number, gender: 'MALE' | 'FEMALE'): TreeNode {
  return {
    memberId,
    displayName: String(memberId),
    gender,
    memberType: 'LINEAGE_MEMBER',
    userBindingState: 'NOT_REQUIRED',
    canExpand: false
  }
}

function pc(from: number, to: number, id: number): TreeEdge {
  return {
    relationshipId: id,
    fromMemberId: from,
    toMemberId: to,
    relationshipType: 'PARENT_CHILD'
  }
}

test('parent cousin child shares generation with viewer', () => {
  const nodes = [
    node(99, 'MALE'),
    node(1, 'MALE'),
    node(3, 'MALE'),
    node(4, 'MALE'),
    node(5, 'MALE'),
    node(6, 'MALE'),
    node(7, 'MALE')
  ]
  const edges = [
    pc(99, 1, 1),
    pc(99, 3, 2),
    pc(1, 4, 3),
    pc(3, 5, 4),
    pc(4, 6, 5),
    pc(5, 7, 6)
  ]
  const graph = buildRelGraph(nodes, edges)
  assert.equal(bloodGenerationGap(graph, 6, 7), 0)
})

test('cousin child is one generation younger than viewer', () => {
  const nodes = [
    node(99, 'MALE'),
    node(1, 'MALE'),
    node(3, 'MALE'),
    node(4, 'MALE'),
    node(5, 'MALE'),
    node(6, 'MALE'),
    node(7, 'MALE'),
    node(8, 'MALE')
  ]
  const edges = [
    pc(99, 1, 1),
    pc(99, 3, 2),
    pc(1, 4, 3),
    pc(3, 5, 4),
    pc(4, 6, 5),
    pc(5, 7, 6),
    pc(7, 8, 7)
  ]
  const graph = buildRelGraph(nodes, edges)
  assert.equal(bloodGenerationGap(graph, 6, 8), 1)
})
