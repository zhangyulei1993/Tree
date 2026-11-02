import { apiClient, unwrapData } from '@/api/client'

export interface InvitationRewardTier { requiredInvites: number; cumulativeRewardMembers: number }
export interface InvitationRewardConfig { enabled: boolean; activationRemainingSlots: number; maxRewardMembers: number; requirePhoneBound: boolean; tiers: InvitationRewardTier[] }
export async function getInvitationRewardConfig() { return unwrapData<InvitationRewardConfig>(await apiClient.get('/admin/family-invitation-reward-config')) }
export async function updateInvitationRewardConfig(payload: InvitationRewardConfig) { return unwrapData<InvitationRewardConfig>(await apiClient.put('/admin/family-invitation-reward-config', payload)) }
