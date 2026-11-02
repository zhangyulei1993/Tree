import assert from 'node:assert/strict'
import test from 'node:test'

import { buildFamilyGraph } from './buildFamilyGraph'
import { layoutFamilyTree } from './layoutFamilyTree'
import {
  filterUnlocatedLineageMemberIds,
  findMainLineageComponent,
  findTreeRoots,
  isExternalMember,
  isLineageMember,
  isSpouseMember
} from './graph'
import { buildMemberGraph } from './graph'
import type { FamilyTreeBuildInput } from './types'
import type { TreeEdge, TreeNode } from '@/types/api'

function node(
  memberId: number,
  displayName: string,
  gender: 'MALE' | 'FEMALE',
  memberType: 'LINEAGE_MEMBER' | 'SPOUSE' | 'EXTERNAL_MEMBER' = 'LINEAGE_MEMBER',
  birthDate?: string
): TreeNode {
  return {
    memberId,
    displayName,
    gender,
    birthDate,
    memberType,
    userBindingState: 'NOT_REQUIRED',
    canExpand: false
  }
}

function pc(from: number, to: number, id: number): TreeEdge {
  return { relationshipId: id, fromMemberId: from, toMemberId: to, relationshipType: 'PARENT_CHILD' }
}

function sp(a: number, b: number, id: number): TreeEdge {
  return { relationshipId: id, fromMemberId: a, toMemberId: b, relationshipType: 'SPOUSE' }
}

function buildInput(nodes: TreeNode[], edges: TreeEdge[]): FamilyTreeBuildInput {
  return { nodes, edges, tree: [] }
}

function renderedMemberIds(input: FamilyTreeBuildInput): number[] {
  const graph = buildFamilyGraph(input)
  const ids = new Set<number>()
  const walk = (branch: { parents: TreeNode[]; childBranches: typeof branch[] }) => {
    for (const parent of branch.parents) ids.add(parent.memberId)
    for (const child of branch.childBranches) walk(child)
  }
  for (const root of graph.roots) walk(root)
  return [...ids]
}

test('grandfather and grandmother render as one couple with grandfather on the left', () => {
  const nodes = [
    node(1, '张祖父', 'MALE', 'LINEAGE_MEMBER', '1940-01-01'),
    node(2, '李祖母', 'FEMALE', 'SPOUSE', '1942-01-01'),
    node(3, '张父亲', 'MALE', 'LINEAGE_MEMBER', '1970-01-01')
  ]
  const edges = [sp(1, 2, 1), pc(1, 3, 2), pc(2, 3, 3)]
  const graph = buildFamilyGraph(buildInput(nodes, edges))
  assert.equal(graph.roots.length, 1)
  const couple = graph.roots[0]!.parents
  assert.equal(couple.length, 2)
  assert.equal(couple[0]!.memberId, 1)
  assert.equal(couple[1]!.memberId, 2)
  assert.equal(isSpouseMember(couple[1]), true)
})

test('father and mother render as one couple row', () => {
  const nodes = [
    node(1, '张祖父', 'MALE', 'LINEAGE_MEMBER', '1940-01-01'),
    node(2, '李祖母', 'FEMALE', 'SPOUSE', '1942-01-01'),
    node(3, '张父亲', 'MALE', 'LINEAGE_MEMBER', '1970-01-01'),
    node(4, '王母亲', 'FEMALE', 'SPOUSE', '1972-01-01'),
    node(5, '张我', 'MALE', 'LINEAGE_MEMBER', '2000-01-01')
  ]
  const edges = [
    sp(1, 2, 1), pc(1, 3, 2), pc(2, 3, 3),
    sp(3, 4, 4), pc(3, 5, 5), pc(4, 5, 6)
  ]
  const graph = buildFamilyGraph(buildInput(nodes, edges))
  const fatherGen = graph.roots[0]!.childBranches[0]!
  assert.equal(fatherGen.parents[0]!.memberId, 3)
  assert.equal(fatherGen.parents[1]!.memberId, 4)
})

test('spouse maternal grandparents stay off main canvas when only linked through grandmother', () => {
  const nodes = [
    node(1, '张祖父', 'MALE', 'LINEAGE_MEMBER', '1940-01-01'),
    node(2, '李祖母', 'FEMALE', 'SPOUSE', '1942-01-01'),
    node(3, '张父亲', 'MALE', 'LINEAGE_MEMBER', '1970-01-01'),
    node(10, '外曾祖父', 'MALE', 'LINEAGE_MEMBER', '1910-01-01'),
    node(11, '外曾祖母', 'FEMALE', 'SPOUSE', '1912-01-01')
  ]
  const edges = [
    sp(1, 2, 1), pc(1, 3, 2),
    sp(10, 11, 2), pc(10, 2, 3), pc(11, 2, 4)
  ]
  const input = buildInput(nodes, edges)
  const rendered = renderedMemberIds(input)
  assert.equal(rendered.includes(10), false)
  assert.equal(rendered.includes(11), false)
  const graph = buildFamilyGraph(input)
  assert.ok(graph.unlocatedMemberIds.includes(10))
  assert.equal(graph.unlocatedMemberIds.includes(11), false)
})

test('SPOUSE member is never chosen as tree root', () => {
  const nodes = [
    node(1, '张祖父', 'MALE', 'LINEAGE_MEMBER', '1940-01-01'),
    node(2, '李祖母', 'FEMALE', 'SPOUSE', '1942-01-01')
  ]
  const edges = [sp(1, 2, 1)]
  const memberGraph = buildMemberGraph(buildInput(nodes, edges))
  const roots = findTreeRoots(memberGraph)
  assert.equal(roots.length, 1)
  assert.equal(isLineageMember(memberGraph.nodeMap.get(roots[0]!)), true)
  assert.equal(isSpouseMember(memberGraph.nodeMap.get(roots[0]!)), false)
})

test('unlocated test member stays outside main canvas', () => {
  const nodes = [
    node(1, '张祖父', 'MALE', 'LINEAGE_MEMBER', '1940-01-01'),
    node(2, '李祖母', 'FEMALE', 'SPOUSE', '1942-01-01'),
    node(99, '张待安', 'MALE', 'LINEAGE_MEMBER', '1999-01-01')
  ]
  const edges = [sp(1, 2, 1)]
  const graph = buildFamilyGraph(buildInput(nodes, edges))
  assert.equal(graph.roots.length, 1)
  assert.deepEqual(graph.unlocatedMemberIds, [99])
})

test('EXTERNAL_MEMBER is not lineage', () => {
  const external = node(8, '外姓亲友', 'MALE', 'EXTERNAL_MEMBER', '1980-01-01')
  assert.equal(isLineageMember(external), false)
  assert.equal(isExternalMember(external), true)
})

test('older unlocated male does not steal main root from connected trunk', () => {
  const nodes = [
    node(1, '张祖父', 'MALE', 'LINEAGE_MEMBER', '1940-01-01'),
    node(2, '李祖母', 'FEMALE', 'SPOUSE', '1942-01-01'),
    node(3, '张父亲', 'MALE', 'LINEAGE_MEMBER', '1970-01-01'),
    node(99, '张老汉', 'MALE', 'LINEAGE_MEMBER', '1910-01-01')
  ]
  const edges = [sp(1, 2, 1), pc(1, 3, 2)]
  const memberGraph = buildMemberGraph(buildInput(nodes, edges))
  const roots = findTreeRoots(memberGraph)
  assert.deepEqual(roots, [1])
})

test('largest lineage component defines the main tree', () => {
  const nodes = [
    node(1, '张甲', 'MALE', 'LINEAGE_MEMBER', '1940-01-01'),
    node(2, '张乙', 'MALE', 'LINEAGE_MEMBER', '1970-01-01'),
    node(3, '张丙', 'MALE', 'LINEAGE_MEMBER', '2000-01-01'),
    node(4, '张丁', 'MALE', 'LINEAGE_MEMBER', '1960-01-01'),
    node(5, '张戊', 'MALE', 'LINEAGE_MEMBER', '1990-01-01'),
    node(10, '李孤立', 'MALE', 'LINEAGE_MEMBER', '1900-01-01'),
    node(11, '李独子', 'MALE', 'LINEAGE_MEMBER', '1930-01-01')
  ]
  const edges = [pc(1, 2, 1), pc(2, 3, 2), pc(2, 4, 3), pc(4, 5, 4), pc(10, 11, 5)]
  const memberGraph = buildMemberGraph(buildInput(nodes, edges))
  const mainComponent = findMainLineageComponent(memberGraph)
  assert.equal(mainComponent.length, 5)
  const roots = findTreeRoots(memberGraph)
  assert.ok(mainComponent.includes(roots[0]!))
  assert.equal(roots[0], 1)
})

test('unlocated list excludes SPOUSE and EXTERNAL_MEMBER', () => {
  const nodes = [
    node(1, '张祖父', 'MALE', 'LINEAGE_MEMBER', '1940-01-01'),
    node(2, '李祖母', 'FEMALE', 'SPOUSE', '1942-01-01'),
    node(99, '张待安', 'MALE', 'LINEAGE_MEMBER', '1999-01-01'),
    node(100, '外姓亲友', 'FEMALE', 'EXTERNAL_MEMBER', '1988-01-01')
  ]
  const edges = [sp(1, 2, 1)]
  const graph = buildFamilyGraph(buildInput(nodes, edges))
  assert.deepEqual(graph.unlocatedMemberIds, [99])
})

test('lineage mother with spouse father renders as couple on main canvas', () => {
  const nodes = [
    node(1, '张祖母', 'FEMALE', 'LINEAGE_MEMBER', '1940-01-01'),
    node(2, '王祖父', 'MALE', 'SPOUSE', '1938-01-01'),
    node(3, '张母亲', 'FEMALE', 'LINEAGE_MEMBER', '1970-01-01'),
    node(4, '李父亲', 'MALE', 'SPOUSE', '1968-01-01'),
    node(5, '张我', 'MALE', 'LINEAGE_MEMBER', '2000-01-01')
  ]
  const edges = [
    sp(1, 2, 1), pc(1, 3, 2), pc(2, 3, 3),
    sp(3, 4, 4), pc(3, 5, 5), pc(4, 5, 6)
  ]
  const graph = buildFamilyGraph(buildInput(nodes, edges))
  assert.equal(graph.roots.length, 1)
  assert.equal(graph.roots[0]!.parents[0]!.memberId, 1)
  assert.equal(graph.roots[0]!.parents[1]!.memberId, 2)
  const parentGen = graph.roots[0]!.childBranches[0]!
  assert.equal(parentGen.parents[0]!.memberId, 3)
  assert.equal(parentGen.parents[1]!.memberId, 4)
})

test('missing memberType is excluded from unlocated lineage list', () => {
  const related = new Set<number>([1])
  const memberTypeById = new Map<number, string>([
    [1, 'LINEAGE_MEMBER'],
    [99, 'LINEAGE_MEMBER'],
    [100, '']
  ])
  assert.deepEqual(filterUnlocatedLineageMemberIds([99, 100], related, memberTypeById), [99])
  assert.deepEqual(filterUnlocatedLineageMemberIds([100], related, memberTypeById), [])
})

test('deep viewer keeps initial viewport anchored to the current center', () => {
  const nodes = [
    node(1, '张曾祖父', 'MALE', 'LINEAGE_MEMBER', '1920-01-01'),
    node(2, '李曾祖母', 'FEMALE', 'SPOUSE', '1922-01-01'),
    node(3, '张祖父', 'MALE', 'LINEAGE_MEMBER', '1945-01-01'),
    node(4, '张父亲', 'MALE', 'LINEAGE_MEMBER', '1970-01-01'),
    node(5, '张本人', 'MALE', 'LINEAGE_MEMBER', '2000-01-01')
  ]
  const edges = [
    sp(1, 2, 1),
    pc(1, 3, 2),
    pc(2, 3, 3),
    pc(3, 4, 4),
    pc(4, 5, 5)
  ]

  const layout = layoutFamilyTree({ ...buildInput(nodes, edges), viewerMemberId: 5 })
  const rootNode = layout.renderNodes.find((item) =>
    item.parents.some((parent) => parent.memberId === 1)
  )
  const viewerNode = layout.renderNodes.find((item) =>
    item.parents.some((parent) => parent.memberId === 5)
  )

  assert.equal(layout.scrollIntoViewId, 'tree-viewer-anchor-5')
  assert.equal(rootNode?.scrollAnchorId, undefined)
  assert.equal(viewerNode?.scrollAnchorId, 'tree-viewer-anchor-5')
})
