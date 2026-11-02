import assert from 'node:assert/strict'
import test from 'node:test'

import { buildPatrilinealDisplayRoles } from './patrilinealDisplay'

test('only male descendants through the paternal line are labelled patrilineal', () => {
  const roles = buildPatrilinealDisplayRoles({
    nodes: [
      { memberId: 1, displayName: '张祖', gender: 'MALE', memberType: 'LINEAGE_MEMBER', userBindingState: 'UNBOUND', canExpand: true },
      { memberId: 2, displayName: '李氏', gender: 'FEMALE', memberType: 'SPOUSE', userBindingState: 'UNBOUND', canExpand: true },
      { memberId: 3, displayName: '张子', gender: 'MALE', memberType: 'LINEAGE_MEMBER', userBindingState: 'UNBOUND', canExpand: true },
      { memberId: 4, displayName: '张女', gender: 'FEMALE', memberType: 'LINEAGE_MEMBER', userBindingState: 'UNBOUND', canExpand: true },
      { memberId: 5, displayName: '王婿', gender: 'MALE', memberType: 'EXTERNAL_MEMBER', userBindingState: 'UNBOUND', canExpand: true },
      { memberId: 6, displayName: '王外孙', gender: 'MALE', memberType: 'LINEAGE_MEMBER', userBindingState: 'UNBOUND', canExpand: true }
    ],
    edges: [
      { relationshipId: 1, fromMemberId: 1, toMemberId: 3, relationshipType: 'PARENT_CHILD' },
      { relationshipId: 2, fromMemberId: 1, toMemberId: 4, relationshipType: 'PARENT_CHILD' },
      { relationshipId: 3, fromMemberId: 4, toMemberId: 6, relationshipType: 'PARENT_CHILD' },
      { relationshipId: 4, fromMemberId: 5, toMemberId: 6, relationshipType: 'PARENT_CHILD' }
    ]
  })

  assert.equal(roles[1], 'PATRILINEAL')
  assert.equal(roles[3], 'PATRILINEAL')
  assert.equal(roles[4], 'NATAL_FEMALE')
  assert.equal(roles[6], 'EXTERNAL')
})
