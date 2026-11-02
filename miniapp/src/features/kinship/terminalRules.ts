import { buildPathKey } from './pathKey'
import { matchAffinityRules } from './affinityRules'
import { canAppendExtendedTerminalSpouse, validatePathGuard } from './pathGuardRules'
import {
  getLastRelation,
  isRelationAllowedByStateMachine,
  MAX_DEPTH_DISABLED_REASON
} from './stateMachine'
import type { KinshipContext, KinshipRelation, ValidationResult } from './types'
import { MAX_KINSHIP_DEPTH } from './types'

export const TERMINAL_COMPLETE_REASON = '当前关系已经形成完整称谓，暂不建议继续追加。'

export function isTerminalContext(context: KinshipContext): boolean {
  if (context.steps.length > MAX_KINSHIP_DEPTH) {
    return true
  }
  if (context.steps.length === 0) {
    return false
  }
  const pathKey = buildPathKey(context)
  return Boolean(matchAffinityRules(context, pathKey)?.terminal)
}

export function getTerminalReason(context: KinshipContext): string | undefined {
  if (isTerminalContext(context)) {
    return TERMINAL_COMPLETE_REASON
  }
  if (context.steps.length >= MAX_KINSHIP_DEPTH) {
    return MAX_DEPTH_DISABLED_REASON
  }
  return undefined
}

export function validateContextStructure(context: KinshipContext): ValidationResult {
  for (let index = 1; index < context.steps.length; index += 1) {
    const partialSteps = context.steps.slice(0, index)
    const nextRelation = context.steps[index].relation
    const partialContext: KinshipContext = {
      ...context,
      steps: partialSteps
    }
    if (
      partialSteps.length >= MAX_KINSHIP_DEPTH &&
      !canAppendExtendedTerminalSpouse(partialContext, nextRelation)
    ) {
      return { valid: false, reason: MAX_DEPTH_DISABLED_REASON }
    }
    const pathGuard = validatePathGuard(partialContext, nextRelation)
    if (!pathGuard.valid) {
      return pathGuard
    }
    const lastRelation = getLastRelation(partialSteps)
    const validation = isRelationAllowedByStateMachine(lastRelation, nextRelation)
    if (!validation.valid) {
      return validation
    }
  }
  return { valid: true }
}

export function canAppendRelationWithoutTerminal(
  context: KinshipContext,
  relation: KinshipRelation
): ValidationResult {
  if (
    context.steps.length >= MAX_KINSHIP_DEPTH &&
    !canAppendExtendedTerminalSpouse(context, relation)
  ) {
    return { valid: false, reason: MAX_DEPTH_DISABLED_REASON }
  }
  const lastRelation = getLastRelation(context.steps)
  return isRelationAllowedByStateMachine(lastRelation, relation)
}
