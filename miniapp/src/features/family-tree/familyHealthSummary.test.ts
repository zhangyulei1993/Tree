import assert from 'node:assert/strict'
import test from 'node:test'

import { calculateFamilyHealthScore, summarizeFamilyHealth } from './familyHealthSummary'
import type { FamilyDetail, FamilyMember, FamilyTreeResult } from '@/types/api'

const family = { regionText: '', nativePlace: '' } as FamilyDetail
const members = [
  { memberId: 1, gender: 'MALE', birthDate: '2000-01-01', userBindingPolicy: 'NOT_REQUIRED' },
  { memberId: 2, gender: 'FEMALE', birthDate: '1999-01-01', userBindingPolicy: 'REQUIRED' }
] as FamilyMember[]
const tree: FamilyTreeResult = {
  familyId: 1,
  treeMode: 'PRIVATE',
  graphVersion: 1,
  tree: [],
  nodes: [
    { memberId: 1, displayName: '父', gender: 'MALE', memberType: 'LINEAGE_MEMBER', birthDate: '2000-01-01', userBindingState: 'NOT_REQUIRED', canExpand: false },
    { memberId: 2, displayName: '子', gender: 'FEMALE', memberType: 'LINEAGE_MEMBER', birthDate: '1999-01-01', userBindingState: 'NOT_REQUIRED', canExpand: false }
  ],
  edges: [{ relationshipId: 1, relationshipType: 'PARENT_CHILD', fromMemberId: 1, toMemberId: 2 }]
}

test('family health summary counts the same actionable member items as the checklist', () => {
  const result = summarizeFamilyHealth(family, members, tree)
  assert.equal(result.birthOrderConflict, 2)
  assert.equal(result.missingSpouse, 1)
  assert.equal(result.requiredUnbound, 1)
  assert.equal(result.missingPlace, 1)
  assert.equal(result.total, 5)
  assert.equal(calculateFamilyHealthScore(result), 65)
})
