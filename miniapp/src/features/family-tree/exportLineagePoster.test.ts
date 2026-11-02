import assert from 'node:assert/strict'
import test from 'node:test'

import { buildExportLineagePosterPages } from './exportLineagePoster'
import { buildExportLineageTree, type ExportLineageGeneration, type ExportLineageNode } from './exportLineageTree'
import { buildExportTreePosterPage, buildFamilyTreePosterPage, isExportTreeMemberDeceased } from './exportTreePoster'
import type { FamilyTreeResult } from '@/types/api'

function item(id: number): ExportLineageNode {
  return {
    id: `node-${id}`,
    generation: 1,
    role: 'PATRILINEAL',
    spouseLabels: [],
    parentLabels: [],
    children: [],
    member: { memberId: id, displayName: `张${id}`, gender: 'MALE', memberType: 'LINEAGE_MEMBER', userBindingState: 'NOT_REQUIRED', canExpand: false }
  }
}

test('poster keeps the generation layout and contains every member', () => {
  const generations: ExportLineageGeneration[] = [
    { generation: 1, members: [item(1)] },
    { generation: 2, members: Array.from({ length: 50 }, (_, index) => item(index + 2)) }
  ]
  const pages = buildExportLineagePosterPages(generations, '张氏一家亲', '宗支长卷')
  assert.ok(pages.length > 1)
  assert.ok(pages.every((page) => page.height <= 2880))
  assert.equal(pages.flatMap((page) => page.groups.flatMap((group) => group.members)).length, 51)
  assert.equal(pages[1]!.groups[0]!.continued, true)
})

test('tree poster retains only the selected export lineage and its parent-child links', () => {
  const tree: FamilyTreeResult = {
    familyId: 1,
    treeMode: 'LIST_TREE',
    graphVersion: 1,
    nodes: [
      { memberId: 1, displayName: '张父', gender: 'MALE', memberType: 'LINEAGE_MEMBER', userBindingState: 'NOT_REQUIRED', canExpand: true },
      { memberId: 2, displayName: '张女', gender: 'FEMALE', memberType: 'LINEAGE_MEMBER', userBindingState: 'NOT_REQUIRED', canExpand: true },
      { memberId: 3, displayName: '王婿', gender: 'MALE', memberType: 'SPOUSE', userBindingState: 'NOT_REQUIRED', canExpand: false },
      { memberId: 4, displayName: '外支后代', gender: 'MALE', memberType: 'LINEAGE_MEMBER', userBindingState: 'NOT_REQUIRED', canExpand: false }
    ],
    edges: [
      { relationshipId: 1, relationshipType: 'PARENT_CHILD', fromMemberId: 1, toMemberId: 2 },
      { relationshipId: 2, relationshipType: 'SPOUSE', fromMemberId: 2, toMemberId: 3 },
      { relationshipId: 3, relationshipType: 'PARENT_CHILD', fromMemberId: 2, toMemberId: 4 }
    ],
    tree: []
  }
  const page = buildExportTreePosterPage(tree, buildExportLineageTree(tree, 'SURNAME'), '张氏一家亲', '姓氏成员 · 分享亲友', 'SHARE')

  assert.ok(page.layout.success)
  assert.equal(page.displayByMemberId.has(3), false)
  assert.equal(page.displayByMemberId.has(4), true)
  assert.equal(page.layout.renderLinks.length, 2)
  assert.match(page.avatarUrlByMemberId.get(1) || '', /tree-portraits\/age-only\/base\/male-adult-/)
  assert.match(page.avatarUrlByMemberId.get(2) || '', /tree-portraits\/age-only\/base\/female-adult-/)
})

test('tree poster marks deceased members even without a recorded death date', () => {
  assert.equal(isExportTreeMemberDeceased({ isLiving: false }), true)
  assert.equal(isExportTreeMemberDeceased({ isLiving: true }), false)
  assert.equal(isExportTreeMemberDeceased({ isLiving: null }), false)
})

test('direct family-tree export keeps spouse nodes instead of applying long-scroll member filters', () => {
  const tree: FamilyTreeResult = {
    familyId: 1,
    treeMode: 'LIST_TREE',
    graphVersion: 1,
    nodes: [
      { memberId: 1, displayName: '张父', gender: 'MALE', memberType: 'LINEAGE_MEMBER', userBindingState: 'NOT_REQUIRED', canExpand: true },
      { memberId: 2, displayName: '王母', gender: 'FEMALE', memberType: 'SPOUSE', userBindingState: 'NOT_REQUIRED', canExpand: true }
    ],
    edges: [{ relationshipId: 1, relationshipType: 'SPOUSE', fromMemberId: 1, toMemberId: 2, relationNoteType: 'SECOND_WIFE' }],
    tree: []
  }
  const page = buildFamilyTreePosterPage(tree, '张氏一家亲家庭树')

  assert.equal(page.displayByMemberId.size, 2)
  assert.ok(page.avatarUrlByMemberId.has(2))
  assert.equal(page.layout.renderNodes[0]?.parents.length, 2)
  assert.equal(page.relationBadgeByMemberId[2], '现任妻')
})
