import { request } from './client'

export interface InvitationRewardTier { requiredInvites: number; cumulativeRewardMembers: number }
export interface InvitationRewardSummary { enabled: boolean; active: boolean; familyId: number | string; usedMembers: number; effectiveLimit: number; validInvites: number; rewardedMembers: number; maxRewardMembers: number; nextRequiredInvites?: number; nextAddedMembers?: number; tiers: InvitationRewardTier[] }
export const getInvitationReward = (familyId: string | number) => request<InvitationRewardSummary>(`/families/${familyId}/invitation-reward`)
export const createInvitationRewardShare = (familyId: string | number) => request<{ referralCode: string; summary: InvitationRewardSummary }>(`/families/${familyId}/invitation-reward/share`, { method: 'POST' })
export const qualifyInvitationReward = (code: string) => request<InvitationRewardSummary>(`/invitation-rewards/${encodeURIComponent(code)}/qualify`, { method: 'POST' })
