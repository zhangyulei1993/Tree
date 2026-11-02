import assert from 'node:assert/strict'
import test from 'node:test'

import {
  auditFamilyGraph,
  buildIntegrityFixTransactionSql,
  buildSuggestedFixList,
  createIntegrityInMemoryExecutor,
  executeIntegrityFixPlan,
  findMissingSpouseEdges,
  findUnlocatedMembers,
  validateIntegrityFixFile
} from './lib/audit-family-tree-integrity.mjs'

const family20Fixture = {
  members: [
    { familyId: 20, memberId: 105, displayName: '张承志', memberType: 'LINEAGE_MEMBER', gender: 'MALE', status: 'ACTIVE' },
    { familyId: 20, memberId: 107, displayName: '张文昌', memberType: 'LINEAGE_MEMBER', gender: 'MALE', status: 'ACTIVE' },
    { familyId: 20, memberId: 108, displayName: '林晓梅', memberType: 'LINEAGE_MEMBER', gender: 'FEMALE', status: 'ACTIVE' },
    { familyId: 20, memberId: 110, displayName: '张启明', memberType: 'LINEAGE_MEMBER', gender: 'MALE', status: 'ACTIVE' },
    { familyId: 20, memberId: 111, displayName: '赵清秀', memberType: 'LINEAGE_MEMBER', gender: 'FEMALE', status: 'ACTIVE' }
  ],
  edges: [
    { familyId: 20, relationshipId: 1, fromMemberId: 107, toMemberId: 105, relationshipType: 'PARENT_CHILD', status: 'ACTIVE' },
    { familyId: 20, relationshipId: 2, fromMemberId: 108, toMemberId: 105, relationshipType: 'PARENT_CHILD', status: 'ACTIVE' },
    { familyId: 20, relationshipId: 3, fromMemberId: 110, toMemberId: 107, relationshipType: 'PARENT_CHILD', status: 'ACTIVE' },
    { familyId: 20, relationshipId: 4, fromMemberId: 111, toMemberId: 107, relationshipType: 'PARENT_CHILD', status: 'ACTIVE' }
  ]
}

test('family20 fixture reports missing spouse edges for known parent pairs', () => {
  const audit = auditFamilyGraph(20, family20Fixture.members, family20Fixture.edges)
  assert.equal(audit.missingSpouse.some((item) => item.fromDisplayName === '张文昌' && item.toDisplayName === '林晓梅'), true)
  assert.equal(audit.missingSpouse.some((item) => item.fromDisplayName === '张启明' && item.toDisplayName === '赵清秀'), true)
})

test('findUnlocatedMembers only flags members without active edges', () => {
  const issues = findUnlocatedMembers(
    [{ familyId: 1, memberId: 9, displayName: '孤立', memberType: 'LINEAGE_MEMBER', status: 'ACTIVE' }],
    []
  )
  assert.equal(issues.length, 1)
  assert.equal(issues[0].type, 'UNLOCATED_MEMBER')
})

test('findMissingSpouseEdges detects co-parent pairs without SPOUSE edge', () => {
  const issues = findMissingSpouseEdges(family20Fixture.members, family20Fixture.edges)
  assert.equal(issues.length, 2)
})

test('buildSuggestedFixList never auto-applies gender-based fixes', () => {
  const audit = auditFamilyGraph(20, family20Fixture.members, family20Fixture.edges)
  const suggested = buildSuggestedFixList(audit)
  assert.equal(suggested.memberTypeUpdates.length, 0)
  assert.equal(suggested.spouseEdges.length, 2)
})

test('integrity fix transaction rolls back when expectedType mismatches', async () => {
  const state = {
    members: new Map([
      ['20:108', { familyId: 20, memberId: 108, memberType: 'LINEAGE_MEMBER' }]
    ]),
    edges: [],
    graphVersions: new Map([[20, 42]])
  }
  const executor = createIntegrityInMemoryExecutor(state)
  await assert.rejects(() => executeIntegrityFixPlan({
    familyId: 20,
    memberTypeUpdates: [{
      familyId: 20,
      memberId: 108,
      expectedType: 'SPOUSE',
      newType: 'LINEAGE_MEMBER',
      reason: '人工确认'
    }],
    spouseEdges: []
  }, executor), /member type update failed/)
  assert.equal(state.members.get('20:108').memberType, 'LINEAGE_MEMBER')
  assert.equal(state.graphVersions.get(20), 42)
})

test('integrity fix transaction applies memberType and spouse edge with single graph_version bump', async () => {
  const state = {
    members: new Map([
      ['20:107', { familyId: 20, memberId: 107, memberType: 'LINEAGE_MEMBER' }],
      ['20:108', { familyId: 20, memberId: 108, memberType: 'LINEAGE_MEMBER' }]
    ]),
    edges: [],
    graphVersions: new Map([[20, 42]])
  }
  const executor = createIntegrityInMemoryExecutor(state)
  await executeIntegrityFixPlan({
    familyId: 20,
    memberTypeUpdates: [{
      familyId: 20,
      memberId: 108,
      expectedType: 'LINEAGE_MEMBER',
      newType: 'SPOUSE',
      reason: '人工确认'
    }],
    spouseEdges: [{
      familyId: 20,
      fromMemberId: 107,
      toMemberId: 108,
      reason: '人工确认'
    }]
  }, executor)
  assert.equal(state.members.get('20:108').memberType, 'SPOUSE')
  assert.equal(state.edges.length, 1)
  assert.equal(state.graphVersions.get(20), 43)
})

test('buildIntegrityFixTransactionSql includes member update, spouse insert and one family bump', () => {
  const sql = buildIntegrityFixTransactionSql({
    familyId: 20,
    memberTypeUpdates: [{
      familyId: 20,
      memberId: 108,
      expectedType: 'LINEAGE_MEMBER',
      newType: 'SPOUSE',
      reason: '人工确认'
    }],
    spouseEdges: [{
      familyId: 20,
      fromMemberId: 107,
      toMemberId: 108,
      reason: '人工确认'
    }]
  })
  assert.match(sql, /UPDATE family_members/)
  assert.match(sql, /INSERT INTO family_relationships/)
  assert.equal((sql.match(/UPDATE families/g) || []).length, 1)
  assert.match(sql, /COMMIT\s*$/)
})

test('validateIntegrityFixFile requires absolute-path-ready schema', () => {
  const parsed = validateIntegrityFixFile({
    families: [{
      familyId: 20,
      memberTypeUpdates: [],
      spouseEdges: [{ fromMemberId: 107, toMemberId: 108, reason: '人工确认' }]
    }]
  })
  assert.equal(parsed[0].familyId, 20)
})
