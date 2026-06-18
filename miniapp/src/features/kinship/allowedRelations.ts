import type { KinshipContext, KinshipRelation, RelationOption, ValidationResult } from './types'
import { MAX_KINSHIP_DEPTH } from './types'
import { getRelationLabel } from './helpers'
import {
  getLastRelation,
  isRelationAllowedByStateMachine,
  MAX_DEPTH_DISABLED_REASON
} from './stateMachine'
import { canAppendExtendedTerminalSpouse, validatePathGuard } from './pathGuardRules'
import { getTerminalReason, isTerminalContext } from './terminalRules'

const ALL_RELATIONS: KinshipRelation[] = ['parent', 'child', 'sibling', 'spouse']

export function canAppendRelation(
  context: KinshipContext,
  relation: KinshipRelation
): ValidationResult {
  const pathGuard = validatePathGuard(context, relation)
  if (!pathGuard.valid) {
    return pathGuard
  }

  if (
    context.steps.length >= MAX_KINSHIP_DEPTH &&
    !canAppendExtendedTerminalSpouse(context, relation)
  ) {
    return { valid: false, reason: MAX_DEPTH_DISABLED_REASON }
  }

  const terminalReason = getTerminalReason(context)
  if (
    terminalReason &&
    isTerminalContext(context) &&
    (context.steps.length < MAX_KINSHIP_DEPTH || !canAppendExtendedTerminalSpouse(context, relation))
  ) {
    return { valid: false, reason: terminalReason }
  }

  const lastRelation = getLastRelation(context.steps)
  return isRelationAllowedByStateMachine(lastRelation, relation)
}

export function getRelationOptions(context: KinshipContext): RelationOption[] {
  return ALL_RELATIONS.map((relation) => {
    const label = getRelationLabel(relation)
    const validation = canAppendRelation(context, relation)
    return {
      relation,
      label,
      enabled: validation.valid,
      disabledReason: validation.valid ? undefined : validation.reason
    }
  })
}

export { PARENT_CHILD_DISABLED_REASON, CHILD_PARENT_DISABLED_REASON } from './stateMachine'
export {
  PARENT_SPOUSE_DISABLED_REASON,
  CHILD_SIBLING_DISABLED_REASON,
  DIRECT_DESCENDANT_DEPTH_DISABLED_REASON,
  COLLATERAL_DESCENDANT_DEPTH_DISABLED_REASON,
  COUSIN_DESCENDANT_DEPTH_DISABLED_REASON
} from './pathGuardRules'
