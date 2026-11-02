import type { FamilyDetail, FamilyMember, FamilyTreeResult } from '@/types/api'
import { inspectRelationshipHealth } from './relationshipHealth'

export interface FamilyHealthSummary {
  unknownGender: number
  isolated: number
  missingBirth: number
  birthOrderConflict: number
  missingSpouse: number
  requiredUnbound: number
  missingPlace: number
  total: number
}

export function calculateFamilyHealthScore(summary: FamilyHealthSummary): number {
  const deductions = [
    summary.unknownGender * 10,
    summary.isolated * 10,
    summary.missingBirth * 5,
    summary.birthOrderConflict * 10,
    summary.missingSpouse * 5,
    summary.requiredUnbound * 5,
    summary.missingPlace * 5
  ]
  return Math.max(0, 100 - deductions.reduce((total, deduction) => total + Math.min(deduction, 30), 0))
}

/** Keep the family-detail badge and the health checklist on exactly the same rules. */
export function summarizeFamilyHealth(
  family: FamilyDetail | null | undefined,
  members: FamilyMember[],
  tree: FamilyTreeResult | null | undefined
): FamilyHealthSummary {
  const unknownGender = members.filter((item) => item.gender !== 'MALE' && item.gender !== 'FEMALE').length
  const missingBirth = members.filter((item) => !item.birthDate && !item.birthYear).length
  const requiredUnbound = members.filter((item) => item.userBindingPolicy === 'REQUIRED' && !item.boundUserId).length
  const connected = new Set((tree?.edges || []).flatMap((edge) => [edge.fromMemberId, edge.toMemberId]))
  const isolated = tree
    ? members.filter((item) => members.length > 1 && !connected.has(item.memberId)).length
    : 0
  const relationshipHealth = inspectRelationshipHealth(tree)
  const birthOrderConflict = relationshipHealth.membersWithBirthOrderConflict.length
  const missingSpouse = relationshipHealth.membersWithChildrenNoSpouse.length
  const missingPlace = !family?.regionText && !family?.nativePlace ? 1 : 0
  return {
    unknownGender,
    isolated,
    missingBirth,
    birthOrderConflict,
    missingSpouse,
    requiredUnbound,
    missingPlace,
    total: unknownGender + isolated + missingBirth + birthOrderConflict + missingSpouse + requiredUnbound + missingPlace
  }
}
