import type { KinshipContext, ValidationResult } from './types'
import { MAX_KINSHIP_DEPTH } from './types'
import { canAppendRelation } from './allowedRelations'

export function isTerminalContext(context: KinshipContext): boolean {
  return context.steps.length >= MAX_KINSHIP_DEPTH
}

export function getTerminalReason(context: KinshipContext): string | undefined {
  if (context.steps.length >= MAX_KINSHIP_DEPTH) {
    return '当前版本最多支持 5 层关系路径'
  }
  return undefined
}

export function validateContextStructure(context: KinshipContext): ValidationResult {
  for (let index = 1; index < context.steps.length; index += 1) {
    const partialContext: KinshipContext = {
      self: context.self,
      steps: context.steps.slice(0, index),
      maxDepth: MAX_KINSHIP_DEPTH
    }
    const nextRelation = context.steps[index].relation
    const validation = canAppendRelation(partialContext, nextRelation)
    if (!validation.valid) {
      return validation
    }
  }
  return { valid: true }
}
