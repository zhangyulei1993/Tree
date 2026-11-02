import assert from 'node:assert/strict'
import test from 'node:test'

import { buildLargePatrilinealSeedPlan } from './lib/large-patrilineal-seed-plan.mjs'

test('large patrilineal plan creates a 50-member six-generation family', () => {
  const plan = buildLargePatrilinealSeedPlan(100)
  assert.equal(plan.expected.memberCount, 50)
  assert.equal(plan.expected.relationshipCount, 83)
  assert.equal(plan.expected.coupleUnits, 13)
  assert.equal(plan.expected.unlocatedCount, 1)
  assert.equal(plan.expected.renderedCount, 49)
  assert.equal(plan.families[0].baseRef, '$founder')
  assert.equal(plan.families.at(-1).children[0].name, '张念祖')
  assert.equal(plan.accountTargets.familyAdmin, '张承志')
  assert.equal(plan.accountTargets.member, '张思远')
})

test('large patrilineal plan uses unique member names and valid family sizes', () => {
  const plan = buildLargePatrilinealSeedPlan(100)
  const names = [
    plan.founderPatch.name,
    ...plan.families.flatMap((family) => [
      family.spouse.name,
      ...family.children.map((item) => item.name)
    ]),
    plan.unlocated.name
  ]
  assert.equal(new Set(names).size, names.length)
  for (const family of plan.families) {
    assert.ok(family.children.length >= 1)
  }
})
