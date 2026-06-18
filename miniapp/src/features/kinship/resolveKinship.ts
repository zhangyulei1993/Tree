import { buildPathKey } from './pathKey'
import { matchExactRules } from './exactRules'
import { matchAncestorRules } from './ancestorRules'
import { matchCousinRules } from './cousinRules'
import { matchNephewRules } from './nephewRules'
import { matchAffinityRules } from './affinityRules'
import { matchFallbackRules } from './fallbackRules'
import { toResolution } from './ruleUtils'
import { validateContextStructure } from './terminalRules'
import type { KinshipContext, KinshipResolution } from './types'

export function resolveKinship(context: KinshipContext): KinshipResolution {
  const structure = validateContextStructure(context)
  if (!structure.valid) {
    return toResolution(context, {
      status: 'unsupported',
      explanation: structure.reason
    })
  }

  if (context.steps.length === 0) {
    return toResolution(context, {
      status: 'resolved',
      primaryTitle: '我',
      explanation: '这是关系路径的起点。'
    })
  }

  const pathKey = buildPathKey(context)

  const ruleMatchers = [
    () => matchAffinityRules(context, pathKey),
    () => matchExactRules(context, pathKey),
    () => matchAncestorRules(context, pathKey),
    () => matchCousinRules(context, pathKey),
    () => matchNephewRules(context, pathKey)
  ]

  for (const matcher of ruleMatchers) {
    const match = matcher()
    if (match) {
      return toResolution(context, match)
    }
  }

  return toResolution(context, matchFallbackRules(context))
}
