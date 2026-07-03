import { canonicalToKinshipResolution } from './canonicalAdapter'
import { resolveCanonicalKinship } from './resolveCanonicalKinship'
import type { KinshipContext, KinshipResolution } from './types'

export function resolveKinship(context: KinshipContext): KinshipResolution {
  return canonicalToKinshipResolution(resolveCanonicalKinship(context))
}
