import assert from 'node:assert/strict'
import { mkdtempSync, writeFileSync } from 'node:fs'
import os from 'node:os'
import path from 'node:path'
import test from 'node:test'

import {
  FIX_ASSERT_TABLE,
  buildAuditReport,
  buildFixTransactionSql,
  buildMemberTypeFixPlan,
  classifyAuditIssues,
  collectAssertionSuccessKeys,
  createInMemoryExecutor,
  executeFixPlan,
  loadFixFile,
  memberKey,
  parseFamilyIdFilter,
  validateFixEntry,
  validateFixFileList
} from './lib/audit-fix-member-type.mjs'

function samplePlan() {
  return {
    updates: [
      {
        familyId: 30,
        memberId: 1,
        expectedType: 'LINEAGE_MEMBER',
        newType: 'SPOUSE',
        reason: '人工核对 1'
      },
      {
        familyId: 30,
        memberId: 2,
        expectedType: 'LINEAGE_MEMBER',
        newType: 'SPOUSE',
        reason: '人工核对 2'
      }
    ],
    graphVersionBumps: [30]
  }
}

test('LINEAGE_MEMBER with only spouse edges is reported but not auto-fixed', () => {
  const issues = classifyAuditIssues({
    familyId: 1,
    memberId: 2,
    displayName: '测试',
    memberType: 'LINEAGE_MEMBER',
    spouseEdges: 1,
    parentChildEdges: 0
  })
  assert.deepEqual(issues, ['SUSPICIOUS_LINEAGE_WITH_ONLY_SPOUSE_EDGES'])

  const report = buildAuditReport([{
    familyId: 1,
    memberId: 2,
    displayName: '测试',
    memberType: 'LINEAGE_MEMBER',
    spouseEdges: 1,
    parentChildEdges: 0
  }])

  assert.equal(report.totals.withIssues, 1)
  assert.equal(report.issues[0].proposedFix, undefined)
  assert.equal('autoFixable' in report.totals, false)
})

test('SPOUSE with parent-child edges is reported but not auto-fixed', () => {
  const issues = classifyAuditIssues({
    familyId: 1,
    memberId: 3,
    displayName: '配偶',
    memberType: 'SPOUSE',
    spouseEdges: 0,
    parentChildEdges: 2
  })
  assert.deepEqual(issues, ['SUSPICIOUS_SPOUSE_WITH_PARENT_CHILD_EDGES'])
})

test('buildFixTransactionSql asserts ROW_COUNT after each member UPDATE', () => {
  const sql = buildFixTransactionSql(samplePlan())
  const memberUpdates = sql.match(/UPDATE family_members/g) || []
  const memberAssertions = sql.match(/SET @member_type_fix_rc := ROW_COUNT\(\)/g) || []

  assert.equal(memberUpdates.length, 2)
  assert.equal(memberAssertions.length, 3)
  assert.match(sql, new RegExp(`INSERT INTO ${FIX_ASSERT_TABLE}`))
  assert.match(sql, /CREATE TEMPORARY TABLE member_type_fix_assert/)
  assert.match(sql, /COMMIT\s*$/)
  assert.match(sql, /IF\(@member_type_fix_rc = 1, 1001, 1\)/)
  assert.match(sql, /IF\(@member_type_fix_rc = 1, 1002, 1\)/)

  const firstMemberAssertIndex = sql.indexOf('UPDATE family_members')
  const firstRowCountIndex = sql.indexOf('SET @member_type_fix_rc := ROW_COUNT()', firstMemberAssertIndex)
  const secondMemberUpdateIndex = sql.indexOf('UPDATE family_members', firstMemberAssertIndex + 1)

  assert.ok(firstRowCountIndex > firstMemberAssertIndex)
  assert.ok(firstRowCountIndex < secondMemberUpdateIndex)
})

test('buildFixTransactionSql asserts ROW_COUNT after graph_version UPDATE', () => {
  const sql = buildFixTransactionSql(samplePlan())
  const familyUpdates = sql.match(/UPDATE families/g) || []

  assert.equal(familyUpdates.length, 1)
  assert.match(sql, /IF\(@member_type_fix_rc = 1, 1003, 1\)/)

  const familyUpdateIndex = sql.indexOf('UPDATE families')
  const familyRowCountIndex = sql.indexOf('SET @member_type_fix_rc := ROW_COUNT()', familyUpdateIndex)
  const commitIndex = sql.lastIndexOf('COMMIT')

  assert.ok(familyRowCountIndex > familyUpdateIndex)
  assert.ok(commitIndex > familyRowCountIndex)
})

test('buildFixTransactionSql uses unique assertion success keys for 1001 members and 1 family', () => {
  const updates = Array.from({ length: 1001 }, (_, index) => ({
    familyId: 99,
    memberId: index + 1,
    expectedType: 'LINEAGE_MEMBER',
    newType: 'SPOUSE',
    reason: `人工核对 ${index + 1}`
  }))
  const sql = buildFixTransactionSql({
    updates,
    graphVersionBumps: [99]
  })
  const keys = collectAssertionSuccessKeys(sql)

  assert.equal(keys.length, 1002)
  assert.equal(new Set(keys).size, 1002)
  assert.equal(keys[0], 1001)
  assert.equal(keys[1000], 2001)
  assert.equal(keys[1001], 2002)
})

test('expectedType mismatch is rejected before writes', () => {
  const currentMembers = new Map([
    [memberKey(20, 108), {
      familyId: 20,
      memberId: 108,
      displayName: '张文昌妻',
      memberType: 'LINEAGE_MEMBER'
    }]
  ])

  assert.throws(() => {
    buildMemberTypeFixPlan([{
      familyId: 20,
      memberId: 108,
      expectedType: 'SPOUSE',
      newType: 'LINEAGE_MEMBER',
      reason: '人工核对'
    }], currentMembers)
  }, /expectedType mismatch/)
})

test('illegal newType is rejected', () => {
  assert.throws(() => {
    validateFixEntry({
      familyId: 20,
      memberId: 108,
      expectedType: 'LINEAGE_MEMBER',
      newType: 'GUEST',
      reason: 'bad'
    })
  }, /newType must be/)
})

test('empty fix file list is rejected', () => {
  assert.throws(() => validateFixFileList([]), /at least one entry/)
})

test('duplicate familyId+memberId entries are rejected', () => {
  assert.throws(() => validateFixFileList([
    {
      familyId: 30,
      memberId: 1,
      expectedType: 'LINEAGE_MEMBER',
      newType: 'SPOUSE',
      reason: '人工核对 1'
    },
    {
      familyId: 30,
      memberId: 1,
      expectedType: 'LINEAGE_MEMBER',
      newType: 'EXTERNAL_MEMBER',
      reason: '人工核对 2'
    }
  ]), /duplicate fix entry/)
})

test('expectedType equal to newType is rejected', () => {
  assert.throws(() => validateFixEntry({
    familyId: 20,
    memberId: 108,
    expectedType: 'LINEAGE_MEMBER',
    newType: 'LINEAGE_MEMBER',
    reason: '伪修复'
  }), /expectedType must differ from newType/)
})

test('invalid --family-id values are rejected', () => {
  for (const raw of ['', '0', '-1', 'abc', '1.5', null, undefined]) {
    assert.throws(() => parseFamilyIdFilter(raw), /--family-id must be a positive integer/)
  }
  assert.equal(parseFamilyIdFilter('22'), 22)
})

test('fix file familyId must match --family-id filter when both are provided', () => {
  assert.throws(() => validateFixFileList([
    {
      familyId: 21,
      memberId: 1,
      expectedType: 'LINEAGE_MEMBER',
      newType: 'SPOUSE',
      reason: '人工核对'
    }
  ], 20), /does not match --family-id=20/)
})

test('same family multiple fixes only bumps graph_version once in CLI SQL', () => {
  const sql = buildFixTransactionSql(samplePlan())
  assert.equal((sql.match(/UPDATE families/g) || []).length, 1)
})

test('any failed update rolls back all changes and graph_version bumps', async () => {
  const state = {
    members: new Map([
      [memberKey(31, 1), { familyId: 31, memberId: 1, memberType: 'LINEAGE_MEMBER' }],
      [memberKey(31, 2), { familyId: 31, memberId: 2, memberType: 'LINEAGE_MEMBER' }]
    ]),
    graphVersions: new Map([[31, 2]])
  }

  const plan = buildMemberTypeFixPlan([
    {
      familyId: 31,
      memberId: 1,
      expectedType: 'LINEAGE_MEMBER',
      newType: 'SPOUSE',
      reason: '人工核对 1'
    },
    {
      familyId: 31,
      memberId: 2,
      expectedType: 'LINEAGE_MEMBER',
      newType: 'SPOUSE',
      reason: '人工核对 2'
    }
  ], state.members)

  const executor = createInMemoryExecutor(state)
  const originalUpdate = executor.updateMember.bind(executor)
  let callCount = 0
  executor.updateMember = async (update) => {
    callCount += 1
    if (callCount === 2) return 0
    return originalUpdate(update)
  }

  await assert.rejects(() => executeFixPlan(plan, executor), /member update failed/)
  assert.equal(state.members.get(memberKey(31, 1)).memberType, 'LINEAGE_MEMBER')
  assert.equal(state.members.get(memberKey(31, 2)).memberType, 'LINEAGE_MEMBER')
  assert.equal(state.graphVersions.get(31), 2)
})

test('loadFixFile requires absolute path and valid schema', () => {
  const dir = mkdtempSync(path.join(os.tmpdir(), 'member-type-fix-'))
  const relativePath = path.join('tmp-fixes.json')
  const absolutePath = path.join(dir, 'member-type-fixes.json')

  writeFileSync(absolutePath, JSON.stringify([
    {
      familyId: 20,
      memberId: 108,
      expectedType: 'LINEAGE_MEMBER',
      newType: 'SPOUSE',
      reason: '人工核对：该节点为张文昌的配偶'
    }
  ], null, 2))

  assert.throws(() => loadFixFile(relativePath), /--fix-file must be an absolute path/)

  const fixes = loadFixFile(absolutePath)
  assert.equal(fixes.length, 1)
  assert.equal(fixes[0].newType, 'SPOUSE')
})
