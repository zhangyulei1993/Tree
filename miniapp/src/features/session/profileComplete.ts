import type { UserInfo } from '@/types/api'

export function isProfileComplete(user: UserInfo | null | undefined): boolean {
  return Boolean(user?.nickname?.trim())
}
