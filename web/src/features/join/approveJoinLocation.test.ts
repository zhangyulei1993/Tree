import assert from 'node:assert/strict'
import test from 'node:test'

import { buildApproveJoinLocation } from './approveJoinLocation'

test('father location includes LINEAGE_MEMBER', () => {
  const location = buildApproveJoinLocation({
    baseMemberId: 3,
    addType: 'ADD_FATHER',
    parentMemberType: 'LINEAGE_MEMBER'
  })
  assert.equal(location.memberType, 'LINEAGE_MEMBER')
})

test('mother location includes SPOUSE', () => {
  const location = buildApproveJoinLocation({
    baseMemberId: 3,
    addType: 'ADD_MOTHER',
    parentMemberType: 'SPOUSE'
  })
  assert.equal(location.memberType, 'SPOUSE')
})

test('non-parent location omits memberType', () => {
  const child = buildApproveJoinLocation({ baseMemberId: 3, addType: 'ADD_CHILD' })
  assert.equal(child.memberType, undefined)

  const spouse = buildApproveJoinLocation({ baseMemberId: 3, addType: 'ADD_SPOUSE' })
  assert.equal(spouse.memberType, undefined)

  const sibling = buildApproveJoinLocation({ baseMemberId: 3, addType: 'ADD_SIBLING' })
  assert.equal(sibling.memberType, undefined)
})
