import { isRealApiMode, request } from '@/api/client'
import type { UserCapabilities } from '@/types/api'

const mockCapabilities: UserCapabilities = {
  trustTier: 'WECHAT_ONLY',
  limits: {
    maxOwnedFamilies: 1,
    maxMembersPerOwnedFamily: 10,
    maxJoinedFamilies: 1
  },
  usage: {
    ownedFamilies: 0,
    joinedFamilies: 0,
    membersPerOwnedFamily: {}
  }
}

export async function fetchCapabilities() {
  if (!isRealApiMode) {
    return mockCapabilities
  }
  return request<UserCapabilities>('/users/me/capabilities')
}
