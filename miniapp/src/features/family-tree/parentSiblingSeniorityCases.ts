import type { TreeEdge, TreeNode } from '@/types/api'

import { buildRelGraph } from './relativeGraph'
import { resolveCanonicalRelativeTitle } from './relativeRules'

function node(
  memberId: number,
  displayName: string,
  gender: 'MALE' | 'FEMALE',
  birth?: { birthDate?: string; birthYear?: number }
): TreeNode {
  return {
    memberId,
    displayName,
    gender,
    birthDate: birth?.birthDate,
    birthYear: birth?.birthYear,
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

function sp(a: number, b: number, id: number): TreeEdge {
  return {
    relationshipId: id,
    fromMemberId: a,
    toMemberId: b,
    relationshipType: 'SPOUSE'
  }
}

export function runParentSiblingSeniorityChecks(): {
  total: number
  passed: number
  failures: string[]
} {
  const failures: string[] = []

  const assertCase = (
    name: string,
    nodes: TreeNode[],
    edges: TreeEdge[],
    viewerId: number,
    targetId: number,
    expect: { title?: string; unsupported?: boolean; forbidden?: string[] }
  ) => {
    const graph = buildRelGraph(nodes, edges)
    const result = resolveCanonicalRelativeTitle(graph, viewerId, targetId)
    if (expect.unsupported) {
      if (!result.unsupportedCanonicalTitle) {
        failures.push(`${name}: expected unsupported, got ${result.canonicalTitle || result.pathDescription}`)
      }
      if (result.canonicalTitle === '伯父' || result.canonicalTitle === '叔父') {
        failures.push(`${name}: must not guess 伯父/叔父 without evidence`)
      }
      return
    }
    if (result.canonicalTitle !== expect.title) {
      failures.push(`${name}: expected ${expect.title}, got ${result.canonicalTitle || result.pathDescription}`)
    }
    for (const forbidden of expect.forbidden || []) {
      if (result.canonicalTitle === forbidden) {
        failures.push(`${name}: forbidden title ${forbidden}`)
      }
    }
  }

  const cases: Array<Parameters<typeof assertCase>> = []

  cases.push([
    'father-missing-birth-uncle-has-date',
    [
      node(10, '祖父', 'MALE', { birthDate: '1940-01-01' }),
      node(11, '祖母', 'FEMALE', { birthDate: '1942-01-01' }),
      node(1, '父亲', 'MALE'),
      node(4, '叔父', 'MALE', { birthDate: '1975-01-01' }),
      node(2, '母亲', 'FEMALE', { birthDate: '1972-01-01' }),
      node(3, '我', 'MALE', { birthDate: '2000-01-01' })
    ],
    [pc(10, 1, 1), pc(11, 1, 2), pc(10, 4, 3), pc(11, 4, 4), pc(1, 3, 5), pc(2, 3, 6)],
    3,
    4,
    { unsupported: true }
  ])

  cases.push([
    'father-has-date-uncle-missing-birth',
    [
      node(10, '祖父', 'MALE', { birthDate: '1940-01-01' }),
      node(11, '祖母', 'FEMALE', { birthDate: '1942-01-01' }),
      node(1, '父亲', 'MALE', { birthDate: '1970-01-01' }),
      node(4, '叔父', 'MALE'),
      node(2, '母亲', 'FEMALE', { birthDate: '1972-01-01' }),
      node(3, '我', 'MALE', { birthDate: '2000-01-01' })
    ],
    [pc(10, 1, 1), pc(11, 1, 2), pc(10, 4, 3), pc(11, 4, 4), pc(1, 3, 5), pc(2, 3, 6)],
    3,
    4,
    { unsupported: true }
  ])

  cases.push([
    'uncle-older-by-birth-year-only',
    [
      node(10, '祖父', 'MALE', { birthYear: 1940 }),
      node(11, '祖母', 'FEMALE', { birthYear: 1942 }),
      node(1, '父亲', 'MALE', { birthYear: 1972 }),
      node(4, '伯父', 'MALE', { birthYear: 1968 }),
      node(2, '母亲', 'FEMALE', { birthYear: 1973 }),
      node(3, '我', 'MALE', { birthYear: 2000 })
    ],
    [pc(10, 1, 1), pc(11, 1, 2), pc(10, 4, 3), pc(11, 4, 4), pc(1, 3, 5), pc(2, 3, 6)],
    3,
    4,
    { title: '伯父' }
  ])

  cases.push([
    'uncle-younger-by-birth-year-only',
    [
      node(10, '祖父', 'MALE', { birthYear: 1940 }),
      node(11, '祖母', 'FEMALE', { birthYear: 1942 }),
      node(1, '父亲', 'MALE', { birthYear: 1968 }),
      node(4, '叔父', 'MALE', { birthYear: 1972 }),
      node(2, '母亲', 'FEMALE', { birthYear: 1970 }),
      node(3, '我', 'MALE', { birthYear: 2000 })
    ],
    [pc(10, 1, 1), pc(11, 1, 2), pc(10, 4, 3), pc(11, 4, 4), pc(1, 3, 5), pc(2, 3, 6)],
    3,
    4,
    { title: '叔父' }
  ])

  cases.push([
    'uncle-older-full-dates',
    [
      node(10, '祖父', 'MALE', { birthDate: '1940-01-01' }),
      node(11, '祖母', 'FEMALE', { birthDate: '1942-01-01' }),
      node(1, '父亲', 'MALE', { birthDate: '1972-06-01' }),
      node(4, '伯父', 'MALE', { birthDate: '1968-03-01' }),
      node(2, '母亲', 'FEMALE', { birthDate: '1973-01-01' }),
      node(3, '我', 'MALE', { birthDate: '2000-01-01' })
    ],
    [pc(10, 1, 1), pc(11, 1, 2), pc(10, 4, 3), pc(11, 4, 4), pc(1, 3, 5), pc(2, 3, 6)],
    3,
    4,
    { title: '伯父' }
  ])

  cases.push([
    'uncle-same-birth-year-unsupported',
    [
      node(10, '祖父', 'MALE', { birthYear: 1940 }),
      node(11, '祖母', 'FEMALE', { birthYear: 1942 }),
      node(1, '父亲', 'MALE', { birthYear: 1970 }),
      node(4, '兄弟', 'MALE', { birthYear: 1970 }),
      node(2, '母亲', 'FEMALE', { birthYear: 1972 }),
      node(3, '我', 'MALE', { birthYear: 2000 })
    ],
    [pc(10, 1, 1), pc(11, 1, 2), pc(10, 4, 3), pc(11, 4, 4), pc(1, 3, 5), pc(2, 3, 6)],
    3,
    4,
    { unsupported: true }
  ])

  cases.push([
    'paternal-aunt-without-parent-birth',
    [
      node(10, '祖父', 'MALE', { birthDate: '1940-01-01' }),
      node(11, '祖母', 'FEMALE', { birthDate: '1942-01-01' }),
      node(1, '父亲', 'MALE'),
      node(4, '姑母', 'FEMALE', { birthDate: '1975-01-01' }),
      node(2, '母亲', 'FEMALE', { birthDate: '1972-01-01' }),
      node(3, '我', 'MALE', { birthDate: '2000-01-01' })
    ],
    [pc(10, 1, 1), pc(11, 1, 2), pc(10, 4, 3), pc(11, 4, 4), pc(1, 3, 5), pc(2, 3, 6)],
    3,
    4,
    { title: '姑母' }
  ])

  cases.push([
    'uncle-wife-insufficient-seniority',
    [
      node(10, '祖父', 'MALE', { birthDate: '1940-01-01' }),
      node(11, '祖母', 'FEMALE', { birthDate: '1942-01-01' }),
      node(1, '父亲', 'MALE'),
      node(4, '叔父', 'MALE', { birthDate: '1975-01-01' }),
      node(5, '婶母', 'FEMALE', { birthDate: '1977-01-01' }),
      node(2, '母亲', 'FEMALE', { birthDate: '1972-01-01' }),
      node(3, '我', 'MALE', { birthDate: '2000-01-01' })
    ],
    [
      pc(10, 1, 1), pc(11, 1, 2), pc(10, 4, 3), pc(11, 4, 4),
      pc(1, 3, 5), pc(2, 3, 6), sp(4, 5, 7)
    ],
    3,
    5,
    { unsupported: true, forbidden: ['伯母', '婶母'] }
  ])

  for (const item of cases) {
    assertCase(...item)
  }

  return {
    total: cases.length,
    passed: cases.length - failures.length,
    failures
  }
}
