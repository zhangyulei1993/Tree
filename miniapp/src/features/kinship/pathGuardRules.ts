import type { KinshipContext, KinshipRelation, ValidationResult } from './types'
import { getLastRelation } from './stateMachine'

export const PARENT_SPOUSE_DISABLED_REASON =
  '该路径通常会绕回另一位父母，建议直接选择父亲或母亲。'

export const CHILD_SIBLING_DISABLED_REASON =
  '该路径通常会绕回你的其他子女，建议直接选择子女关系。'

export const DIRECT_DESCENDANT_DEPTH_DISABLED_REASON =
  '当前版本从本人往下最多支持到重孙辈，如需继续记录更晚辈关系，请在后续版本使用家谱功能。'

export const COLLATERAL_DESCENDANT_DEPTH_DISABLED_REASON =
  '当前版本从兄弟姐妹往下最多支持到侄重孙辈，可以继续选择其配偶，但不再继续下探更晚辈。'

export const COUSIN_DESCENDANT_DEPTH_DISABLED_REASON =
  '当前版本从堂表亲往下最多支持到堂表亲侄孙辈，可以继续选择其配偶，但不再继续下探更晚辈。'

function isDirectDescendantPath(context: KinshipContext): boolean {
  return context.steps.length > 0 && context.steps.every((step) => step.relation === 'child')
}

function isSiblingDescendantPath(context: KinshipContext): boolean {
  return (
    context.steps.length > 1 &&
    context.steps[0].relation === 'sibling' &&
    context.steps.slice(1).every((step) => step.relation === 'child')
  )
}

function buildRelationKey(context: KinshipContext): string {
  return context.steps.map((step) => step.relation).join('>')
}

function isCousinDescendantPath(context: KinshipContext): boolean {
  return buildRelationKey(context) === 'parent>sibling>child>child>child'
}

export function canAppendExtendedTerminalSpouse(
  context: KinshipContext,
  nextRelation: KinshipRelation
): boolean {
  return nextRelation === 'spouse' && buildRelationKey(context) === 'parent>sibling>child>child>child'
}

export function validatePathGuard(
  context: KinshipContext,
  nextRelation: KinshipRelation
): ValidationResult {
  const lastRelation = getLastRelation(context.steps)

  if (nextRelation === 'child' && context.steps.length >= 3 && isDirectDescendantPath(context)) {
    return { valid: false, reason: DIRECT_DESCENDANT_DEPTH_DISABLED_REASON }
  }

  if (nextRelation === 'child' && context.steps.length >= 4 && isSiblingDescendantPath(context)) {
    return { valid: false, reason: COLLATERAL_DESCENDANT_DEPTH_DISABLED_REASON }
  }

  if (nextRelation === 'child' && isCousinDescendantPath(context)) {
    return { valid: false, reason: COUSIN_DESCENDANT_DEPTH_DISABLED_REASON }
  }

  if (lastRelation === 'parent' && nextRelation === 'spouse') {
    return { valid: false, reason: PARENT_SPOUSE_DISABLED_REASON }
  }

  if (lastRelation === 'child' && nextRelation === 'sibling') {
    return { valid: false, reason: CHILD_SIBLING_DISABLED_REASON }
  }

  return { valid: true }
}
