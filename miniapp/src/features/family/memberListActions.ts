import type { FamilyMember, Invitation } from '@/types/api'

export type FamilyManageRole = 'FOUNDER' | 'FAMILY_ADMIN' | 'MEMBER' | string

export interface MemberActionContext {
  canManageFamily: boolean
  invitation?: Invitation | null
  invitationExpired?: boolean
}

function isActiveMember(member: FamilyMember) {
  return member.status === 'ACTIVE'
}

function isProtectedBoundRole(member: FamilyMember) {
  return member.boundFamilyRole === 'FOUNDER' || member.boundFamilyRole === 'FAMILY_ADMIN'
}

export function canEditMember(member: FamilyMember, canManageFamily: boolean) {
  return canManageFamily && isActiveMember(member)
}

export function canInviteMember(member: FamilyMember, canManageFamily: boolean) {
  return canManageFamily
    && isActiveMember(member)
    && !member.boundUserId
    && member.userBindingPolicy !== 'NOT_REQUIRED'
}

export function canUnbindMember(member: FamilyMember, canManageFamily: boolean) {
  return canManageFamily
    && isActiveMember(member)
    && Boolean(member.boundUserId)
    && !isProtectedBoundRole(member)
}

export function canDeleteMember(member: FamilyMember, canManageFamily: boolean) {
  return canManageFamily && isActiveMember(member)
}

export function hasPendingInvitation(ctx: MemberActionContext) {
  const item = ctx.invitation
  if (!item) return false
  if (item.status === 'PENDING' && !ctx.invitationExpired) return true
  if (ctx.invitationExpired || item.status === 'EXPIRED') return true
  return false
}

export function canShowInviteButton(member: FamilyMember, ctx: MemberActionContext) {
  return canInviteMember(member, ctx.canManageFamily) && !hasPendingInvitation(ctx)
}

export function memberActionLabels(member: FamilyMember, ctx: MemberActionContext): string[] {
  const labels: string[] = []
  if (canEditMember(member, ctx.canManageFamily)) labels.push('编辑')
  if (canUnbindMember(member, ctx.canManageFamily)) labels.push('解除绑定')
  if (canShowInviteButton(member, ctx)) labels.push('邀请本人绑定')
  if (hasPendingInvitation(ctx)) labels.push('查看邀请')
  if (canDeleteMember(member, ctx.canManageFamily)) labels.push('删除')
  return labels
}

export function memberActionSummary(member: FamilyMember, ctx: MemberActionContext): string {
  const labels = memberActionLabels(member, ctx)
  if (labels.length === 0) return ''
  return labels.join('、')
}

export function hasNodeAction(member: FamilyMember, ctx: MemberActionContext) {
  return memberActionLabels(member, ctx).length > 0
}

export function hasMoreAction(member: FamilyMember, ctx: MemberActionContext) {
  return hasNodeAction(member, ctx)
}

export function invitationHeadline(ctx: MemberActionContext) {
  if (!ctx.invitation) return ''
  if (ctx.invitationExpired || ctx.invitation.status === 'EXPIRED') return '邀请已过期'
  if (ctx.invitation.status === 'PENDING') return '邀请待接受'
  return ''
}
