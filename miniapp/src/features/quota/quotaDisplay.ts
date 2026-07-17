import type { UserCapabilities } from '@/types/api'

const QUOTA_EXCEEDED_CODES = new Set([41502, 41503, 41504])

type QuotaApiError = Error & {
  code?: number
  data?: {
    trustTier?: unknown
  }
}

function isQuotaApiError(error: unknown): error is QuotaApiError {
  return error instanceof Error && 'code' in error
}

function trustTierFromError(error: QuotaApiError) {
  const trustTier = error.data?.trustTier
  return typeof trustTier === 'string' ? trustTier : ''
}

export function formatQuotaLine(label: string, usage: number, limit: number) {
  return `${label}：${usage}/${limit}`
}

export function quotaReachedMessage(trustTier: string) {
  if (trustTier === 'WECHAT_ONLY') {
    return '当前扩展能力尚未开放'
  }
  return '已达到当前账号权益上限'
}

export function buildCapabilitiesSummary(capabilities: UserCapabilities) {
  const { limits, usage } = capabilities
  return [
    formatQuotaLine('可创建家庭', usage.ownedFamilies, limits.maxOwnedFamilies),
    formatQuotaLine('可加入家庭', usage.joinedFamilies, limits.maxJoinedFamilies),
    `每个家庭成员上限：${limits.maxMembersPerOwnedFamily}`,
    limits.supportsGenerationNaming
      ? '字辈体系：已开放'
      : '字辈体系：当前账号暂未开放'
  ]
}

export function parseQuotaErrorMessage(error: unknown, fallback: string) {
  if (isQuotaApiError(error)) {
    if (error.code && QUOTA_EXCEEDED_CODES.has(error.code)) {
      return quotaReachedMessage(trustTierFromError(error))
    }
    return error.message || fallback
  }
  if (!(error instanceof Error)) {
    return fallback
  }
  return error.message || fallback
}
