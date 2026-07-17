import { apiClient, apiMode, unwrapData } from '@/api/client'
import type {
  AccountFeatureOverrideItem,
  AccountQuotaConfig,
  AccountQuotaImpactPreview,
  UpdateAccountQuotaInput
} from '@/types/api'

const mockConfigs: AccountQuotaConfig[] = [
  {
    trustTier: 'WECHAT_ONLY',
    maxOwnedFamilies: 1,
    maxMembersPerOwnedFamily: 10,
    maxJoinedFamilies: 1,
    supportsGenerationNaming: false,
    updatedAt: '2026-06-01T10:00:00+08:00'
  },
  {
    trustTier: 'PHONE_BOUND',
    maxOwnedFamilies: 1,
    maxMembersPerOwnedFamily: 20,
    maxJoinedFamilies: 5,
    supportsGenerationNaming: true,
    updatedAt: '2026-06-01T10:00:00+08:00'
  }
]

let mockGenerationNamingOverrides: AccountFeatureOverrideItem[] = []

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

export async function listAccountFeatureOverrides(featureKey: string) {
  if (apiMode === 'mock') {
    return mockGenerationNamingOverrides
  }
  const response = await apiClient.get(`/admin/account-feature-overrides/${featureKey}`)
  return unwrapData<AccountFeatureOverrideItem[]>(response)
}

function maskPhone(phone: string) {
  return phone.replace(/^(\d{3})\d{4}(\d{4})$/, '$1****$2')
}

export async function updateAccountFeatureOverrides(featureKey: string, phones: string[]) {
  if (apiMode === 'mock') {
    mockGenerationNamingOverrides = phones.map((phone) => ({
      featureKey,
      phoneMask: maskPhone(phone),
      updatedAt: new Date().toISOString()
    }))
    return mockGenerationNamingOverrides
  }
  const response = await apiClient.put(`/admin/account-feature-overrides/${featureKey}`, { phones })
  return unwrapData<AccountFeatureOverrideItem[]>(response)
}
