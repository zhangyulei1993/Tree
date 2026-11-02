import assert from 'node:assert/strict'
import test from 'node:test'

import { buildSeedPlan, summarizeSeedPlan } from './lib/patrilineal-seed-plan.mjs'

test('seed plan uses founder as youngest node and avoids duplicate founder name', () => {
  const plan = buildSeedPlan({ founderMemberId: 100, founderDisplayName: '张承志' })
  const summary = summarizeSeedPlan(plan)
  assert.equal(plan.founderMemberId, 100)
  assert.equal(summary.nodesToCreate[0]?.base, '张承志')
  assert.equal(summary.nodesToCreate[0]?.addType, 'ADD_FATHER')
  assert.ok(!summary.nodesToCreate.some((node) => node.name === '张承志'))
  assert.equal(plan.expected.memberCount, 9)
  assert.equal(plan.expected.relationshipCount, 11)
  assert.equal(plan.expected.coupleUnits, 3)
})

test('seed plan rejects duplicate planned names', () => {
  assert.throws(() => {
    buildSeedPlan({ founderMemberId: 1, founderDisplayName: '张叔平' })
  }, /must not recreate founder name/)
})
