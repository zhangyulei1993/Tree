import assert from 'node:assert/strict'
import test from 'node:test'

import type { TreeEdge, TreeNode } from '@/types/api'

import { buildRelGraph } from './relativeGraph'
import { resolveCanonicalRelativeTitle } from './relativeRules'

function node(memberId: number, gender: 'MALE' | 'FEMALE', birthYear: number): TreeNode {
  return {
    memberId,
    displayName: String(memberId),
    gender,
    birthDate: `${birthYear}-01-01`,
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

function buildSiblingBranchFixture() {
  const nodes = [
    node(99, 'MALE', 1900),
    node(4, 'MALE', 1930),
    node(6, 'MALE', 1980),
    node(7, 'MALE', 1982),
    node(8, 'MALE', 2005),
    node(9, 'MALE', 2030),
    node(10, 'MALE', 2055),
    node(11, 'MALE', 2080)
  ]
  const edges = [
    pc(99, 4, 1),
    pc(4, 6, 2),
    pc(4, 7, 3),
    pc(7, 8, 4),
    pc(8, 9, 5),
    pc(9, 10, 6),
    pc(10, 11, 7)
  ]
  return buildRelGraph(nodes, edges)
}

function buildCousinBranchFixture() {
  const nodes = [
    node(99, 'MALE', 1900),
    node(1, 'MALE', 1928),
    node(3, 'MALE', 1926),
    node(4, 'MALE', 1960),
    node(5, 'MALE', 1958),
    node(6, 'MALE', 1990),
    node(7, 'MALE', 1992),
    node(8, 'MALE', 2015),
    node(9, 'MALE', 2040),
    node(10, 'MALE', 2065)
  ]
  const edges = [
    pc(99, 1, 1),
    pc(99, 3, 2),
    pc(1, 4, 3),
    pc(3, 5, 4),
    pc(4, 6, 5),
    pc(5, 7, 6),
    pc(7, 8, 7),
    pc(8, 9, 8),
    pc(9, 10, 9)
  ]
  return buildRelGraph(nodes, edges)
}

test('sibling branch gap 2 resolves to 侄孙', () => {
  const graph = buildSiblingBranchFixture()
  const result = resolveCanonicalRelativeTitle(graph, 6, 9)
  assert.equal(result.canonicalTitle, '侄孙')
  assert.equal(result.unsupportedCanonicalTitle, false)
})

test('sibling branch gap 3 does not resolve to 侄孙', () => {
  const graph = buildSiblingBranchFixture()
  const result = resolveCanonicalRelativeTitle(graph, 6, 10)
  assert.notEqual(result.canonicalTitle, '侄孙')
  assert.equal(result.unsupportedCanonicalTitle, true)
  assert.ok(result.pathDescription)
})

test('sibling branch gap 4 does not resolve to 侄孙', () => {
  const graph = buildSiblingBranchFixture()
  const result = resolveCanonicalRelativeTitle(graph, 6, 11)
  assert.notEqual(result.canonicalTitle, '侄孙')
  assert.equal(result.unsupportedCanonicalTitle, true)
})

test('cousin branch gap 2 resolves to 堂侄孙', () => {
  const graph = buildCousinBranchFixture()
  const result = resolveCanonicalRelativeTitle(graph, 6, 9)
  assert.equal(result.canonicalTitle, '堂侄孙')
  assert.equal(result.unsupportedCanonicalTitle, false)
})

test('cousin branch gap 3 does not resolve to 堂侄孙', () => {
  const graph = buildCousinBranchFixture()
  const result = resolveCanonicalRelativeTitle(graph, 6, 10)
  assert.notEqual(result.canonicalTitle, '堂侄孙')
  assert.equal(result.unsupportedCanonicalTitle, true)
  assert.ok(result.pathDescription)
})
