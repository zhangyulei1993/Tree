import type { FamilyMember } from '@/types/api'

/**
 * Tree API nodes do not include boundUserId/userId.
 * Match the logged-in user through the members list boundUserId field.
 */
export function resolveCurrentMemberId(
  userId: number | string | null | undefined,
  members: FamilyMember[]
): number | null {
  if (!userId) return null
  const mine = members.find((member) => String(member.boundUserId) === String(userId))
  return mine?.memberId ?? null
}
