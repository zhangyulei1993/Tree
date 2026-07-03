import assert from 'node:assert/strict'
import test from 'node:test'

import {
  canInviteMember,
  canShowInviteButton,
  canUnbindMember,
  hasNodeAction,
  memberActionSummary
} from './memberListActions'

function member(overrides: Record<string, unknown> = {}) {
  return {
    memberId: 1,
    familyId: 22,
    name: '测试成员',
    gender: 'MALE',
    isAlive: true,
    status: 'ACTIVE',
    userBindingPolicy: 'OPTIONAL',
    createdAt: '',
    updatedAt: '',
    ...overrides
  } as import('@/types/api').FamilyMember
}

test('bound member shows edit and unbind, not invite', () => {
  const item = member({ boundUserId: 9 })
  assert.equal(canInviteMember(item, true), false)
  assert.equal(canUnbindMember(item, true), true)
  assert.equal(
    memberActionSummary(item, { canManageFamily: true }),
    '编辑、解除绑定、删除'
  )
})

test('NOT_REQUIRED member does not show invite', () => {
  const item = member({ userBindingPolicy: 'NOT_REQUIRED' })
  assert.equal(canInviteMember(item, true), false)
  assert.equal(memberActionSummary(item, { canManageFamily: true }), '编辑、删除')
})

test('optional unbound member shows invite for manager', () => {
  const item = member()
  assert.equal(canShowInviteButton(item, { canManageFamily: true }), true)
  assert.equal(
    memberActionSummary(item, { canManageFamily: true }),
    '编辑、邀请本人绑定、删除'
  )
})

test('regular member sees no management actions', () => {
  const item = member({ boundUserId: 9 })
  assert.equal(hasNodeAction(item, { canManageFamily: false }), false)
  assert.equal(memberActionSummary(item, { canManageFamily: false }), '')
})

test('pending invitation replaces invite button with view invite', () => {
  const item = member()
  const summary = memberActionSummary(item, {
    canManageFamily: true,
    invitation: {
      invitationId: 1,
      familyId: 22,
      targetMemberId: 1,
      inviteChannel: 'SHARE_LINK',
      familyRoleAfterAccept: 'MEMBER',
      status: 'PENDING',
      expiredAt: new Date(Date.now() + 86400000).toISOString(),
      createdAt: new Date().toISOString()
    }
  })
  assert.equal(summary, '编辑、查看邀请、删除')
})

test('founder bound member does not show unbind', () => {
  const item = member({ boundUserId: 9, boundFamilyRole: 'FOUNDER' })
  assert.equal(canUnbindMember(item, true), false)
  assert.equal(memberActionSummary(item, { canManageFamily: true }), '编辑、删除')
})

test('family admin bound member does not show unbind', () => {
  const item = member({ boundUserId: 9, boundFamilyRole: 'FAMILY_ADMIN' })
  assert.equal(canUnbindMember(item, true), false)
  assert.equal(memberActionSummary(item, { canManageFamily: true }), '编辑、删除')
})
