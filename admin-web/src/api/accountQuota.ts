import { apiClient, apiMode, unwrapData } from '@/api/client'
import type { AccountQuotaConfig, AccountQuotaImpactPreview, UpdateAccountQuotaInput } from '@/types/api'

const mockConfigs: AccountQuotaConfig[] = [
  {
    trustTier: 'WECHAT_ONLY',
    maxOwnedFamilies: 1,
    maxMembersPerOwnedFamily: 10,
    maxJoinedFamilies: 1,
    updatedAt: '2026-06-01T10:00:00+08:00'
  },
  {
    trustTier: 'PHONE_BOUND',
    maxOwnedFamilies: 1,
    maxMembersPerOwnedFamily: 20,
    maxJoinedFamilies: 5,
    updatedAt: '2026-06-01T10:00:00+08:00'
  }
]

export async function listAccountQuotaConfigs() {
  if (apiMode === 'mock') {
    return mockConfigs
  }
  const response = await apiClient.get('/admin/account-quota-configs')
  return unwrapData<AccountQuotaConfig[]>(response)
}

export async function previewAccountQuotaImpact(tier: string, input: UpdateAccountQuotaInput) {
  if (apiMode === 'mock') {
    return {
      trustTier: tier,
      affectedUsers: 0,
      affectedFamilies: 0,
      ...input
    } satisfies AccountQuotaImpactPreview
  }
  const response = await apiClient.post(`/admin/account-quota-configs/${tier}/impact-preview`, input)
  return unwrapData<AccountQuotaImpactPreview>(response)
}

export async function updateAccountQuotaConfig(tier: string, input: UpdateAccountQuotaInput) {
  if (apiMode === 'mock') {
    const index = mockConfigs.findIndex((item) => item.trustTier === tier)
    if (index >= 0) {
      mockConfigs[index] = {
        ...mockConfigs[index],
        ...input,
        updatedAt: new Date().toISOString()
      }
    }
    return mockConfigs[index]
  }
  const response = await apiClient.put(`/admin/account-quota-configs/${tier}`, input)
  return unwrapData<AccountQuotaConfig>(response)
}
