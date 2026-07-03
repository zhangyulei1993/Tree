import type { KinshipRelation, ValidationResult } from './types'
import { getRelationLabel } from './helpers'

export const PARENT_CHILD_DISABLED_REASON =
  '父母的子女通常会回到本人、兄弟姐妹或旁系关系，请直接选择“兄弟姐妹”等更明确的关系。'

export const CHILD_PARENT_DISABLED_REASON =
  '子女的父母通常会回到本人、配偶或上一层关系，当前版本不通过回退路径表达。'

export const MAX_DEPTH_DISABLED_REASON = '当前版本最多支持 6 层关系路径'

export function getLastRelation(steps: { relation: KinshipRelation }[]): KinshipRelation | null {
  if (steps.length === 0) return null
  return steps[steps.length - 1].relation
}

export function isRelationAllowedByStateMachine(
  lastRelation: KinshipRelation | null,
  nextRelation: KinshipRelation
): ValidationResult {
  if (lastRelation === null) {
    return { valid: true }
  }

  if (lastRelation === 'parent' && nextRelation === 'child') {
    return { valid: false, reason: PARENT_CHILD_DISABLED_REASON }
  }

  if (lastRelation === 'child' && nextRelation === 'parent') {
    return { valid: false, reason: CHILD_PARENT_DISABLED_REASON }
  }

  const allowedMap: Record<KinshipRelation, KinshipRelation[]> = {
    parent: ['parent', 'sibling'],
    child: ['child', 'spouse'],
    sibling: ['child', 'spouse'],
    spouse: ['parent', 'sibling', 'child']
  }

  if (!allowedMap[lastRelation].includes(nextRelation)) {
    const lastLabel = getRelationLabel(lastRelation)
    const nextLabel = getRelationLabel(nextRelation)
    return {
      valid: false,
      reason: `在“${lastLabel}”之后暂不支持继续选择“${nextLabel}”。`
    }
  }

  return { valid: true }
}
