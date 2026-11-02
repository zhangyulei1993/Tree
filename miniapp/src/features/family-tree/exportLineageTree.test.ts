import assert from 'node:assert/strict'
import test from 'node:test'

import { buildExportLineageTree } from './exportLineageTree'
import type { FamilyTreeResult, TreeEdge, TreeNode } from '@/types/api'

function member(id: number, name: string, gender: 'MALE' | 'FEMALE', memberType = 'LINEAGE_MEMBER'): TreeNode {
  return { memberId: id, displayName: name, gender, memberType, userBindingState: 'NOT_REQUIRED', canExpand: false }
}

function pc(fromMemberId: number, toMemberId: number, relationshipId: number): TreeEdge {
  return { fromMemberId, toMemberId, relationshipId, relationshipType: 'PARENT_CHILD' }
}

function spouse(fromMemberId: number, toMemberId: number, relationshipId: number): TreeEdge {
  return { fromMemberId, toMemberId, relationshipId, relationshipType: 'SPOUSE' }
}

test('surname export excludes external spouses while family export retains their annotations', () => {
  const tree: FamilyTreeResult = {
    familyId: 1, treeMode: 'PRIVATE', graphVersion: 1, tree: [],
    nodes: [{ ...member(1, '张祖', 'MALE') }, { ...member(2, '李氏', 'FEMALE', 'SPOUSE'), deathDate: '2018-01-01' }, member(3, '张子', 'MALE')],
    edges: [spouse(1, 2, 1), pc(1, 3, 2), pc(2, 3, 3)]
  }
  const result = buildExportLineageTree(tree)
  assert.equal(result.roots.length, 1)
  assert.equal(result.roots[0]!.member.displayName, '张祖')
  assert.deepEqual(result.roots[0]!.spouseLabels, [])
  assert.equal(result.roots[0]!.children[0]!.member.displayName, '张子')
  assert.deepEqual(result.roots[0]!.children[0]!.parentLabels, ['父：张祖'])
  assert.equal(result.visibleMemberCount, 2)
  assert.deepEqual(result.generations.map((group) => group.members.map((item) => item.member.displayName)), [['张祖'], ['张子']])
  assert.deepEqual(buildExportLineageTree(tree, 'FAMILY').roots[0]!.spouseLabels, ['妻：李氏（卒：2018）'])
})

test('export annotates spouse birth and death years and distinguishes former spouses', () => {
  const tree: FamilyTreeResult = {
    familyId: 1, treeMode: 'PRIVATE', graphVersion: 1, tree: [],
    nodes: [
      { ...member(1, '张宗', 'MALE'), birthDate: '1940-01-01' },
      { ...member(2, '李氏', 'FEMALE', 'SPOUSE'), birthDate: '1942-01-01', deathDate: '2002-01-01' },
      { ...member(3, '王氏', 'FEMALE', 'SPOUSE'), birthDate: '1950-01-01' }
    ],
    edges: [
      { ...spouse(1, 2, 1), relationNoteType: 'EX_WIFE' },
      { ...spouse(1, 3, 2), relationNoteType: 'SECOND_WIFE' }
    ]
  }

  assert.deepEqual(buildExportLineageTree(tree, 'SURNAME').roots[0]!.spouseLabels, [])
  assert.deepEqual(buildExportLineageTree(tree, 'FAMILY').roots[0]!.spouseLabels, [
    '前妻：李氏（1942-2002）',
    '现任妻：王氏'
  ])
})

test('export shows a lineage female and one following generation only', () => {
  const tree: FamilyTreeResult = {
    familyId: 1, treeMode: 'PRIVATE', graphVersion: 1, tree: [],
    nodes: [member(1, '张祖', 'MALE'), member(2, '张女', 'FEMALE'), member(3, '王外孙', 'MALE'), member(4, '王曾孙', 'MALE')],
    edges: [pc(1, 2, 1), pc(2, 3, 2), pc(3, 4, 3)]
  }
  const result = buildExportLineageTree(tree)
  const daughter = result.roots[0]!.children[0]!
  assert.equal(daughter.role, 'NATAL_FEMALE')
  assert.equal(daughter.children[0]!.member.displayName, '王外孙')
  assert.equal(daughter.children[0]!.role, 'NATAL_FEMALE_CHILD')
  assert.deepEqual(daughter.children[0]!.parentLabels, ['母：张女'])
  assert.equal(daughter.children[0]!.children.length, 0)
  assert.equal(result.hiddenDescendantCount, 1)
  assert.deepEqual(result.generations.map((group) => group.members.map((item) => item.member.displayName)), [['张祖'], ['张女'], ['王外孙']])
})

test('family member export keeps descendants after a female lineage member', () => {
  const tree: FamilyTreeResult = {
    familyId: 1, treeMode: 'PRIVATE', graphVersion: 1, tree: [],
    nodes: [member(1, '张祖', 'MALE'), member(2, '张女', 'FEMALE'), member(3, '王父', 'MALE', 'SPOUSE'), member(4, '王外孙', 'MALE'), member(5, '王曾孙', 'MALE')],
    edges: [pc(1, 2, 1), spouse(2, 3, 2), pc(2, 4, 3), pc(3, 4, 4), pc(4, 5, 5)]
  }

  const result = buildExportLineageTree(tree, 'FAMILY')
  const daughter = result.roots[0]!.children[0]!
  assert.equal(daughter.role, 'FAMILY_MEMBER')
  assert.equal(daughter.children[0]!.role, 'FAMILY_MEMBER')
  assert.deepEqual(daughter.children[0]!.parentLabels, ['父：王父', '母：张女'])
  assert.equal(daughter.children[0]!.children[0]!.member.displayName, '王曾孙')
  assert.equal(result.hiddenDescendantCount, 0)
  assert.deepEqual(result.generations.map((group) => group.members.map((item) => item.member.displayName)), [['张祖'], ['张女'], ['王外孙'], ['王曾孙']])
})

test('family member export infers one missing parent from an unambiguous spouse relationship', () => {
  const tree: FamilyTreeResult = {
    familyId: 1, treeMode: 'PRIVATE', graphVersion: 1, tree: [],
    nodes: [member(1, '张父', 'MALE'), member(2, '李母', 'FEMALE', 'SPOUSE'), member(3, '张子', 'MALE')],
    edges: [spouse(1, 2, 1), pc(1, 3, 2)]
  }

  assert.deepEqual(buildExportLineageTree(tree).roots[0]!.children[0]!.parentLabels, ['父：张父'])
  assert.deepEqual(buildExportLineageTree(tree, 'FAMILY').roots[0]!.children[0]!.parentLabels, ['父：张父', '母：李母'])
})

test('family member export never infers a biological parent from a marked former or current marriage', () => {
  const tree: FamilyTreeResult = {
    familyId: 1, treeMode: 'PRIVATE', graphVersion: 1, tree: [],
    nodes: [member(1, '张父', 'MALE'), member(2, '李前妻', 'FEMALE', 'SPOUSE'), member(3, '王现妻', 'FEMALE', 'SPOUSE'), member(4, '张子', 'MALE')],
    edges: [
      { ...spouse(1, 2, 1), relationNoteType: 'EX_WIFE' },
      { ...spouse(1, 3, 2), relationNoteType: 'SECOND_WIFE' },
      pc(1, 4, 3)
    ]
  }

  assert.deepEqual(buildExportLineageTree(tree, 'FAMILY').roots[0]!.children[0]!.parentLabels, ['父：张父'])
})
