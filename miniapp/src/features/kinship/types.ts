export const MAX_KINSHIP_DEPTH = 6

export type KinshipRelation = 'parent' | 'child' | 'sibling' | 'spouse'

export type Gender = 'male' | 'female' | 'unknown'

export type RelativeAge = 'older' | 'younger' | 'same' | 'unknown'

export interface PersonFacts {
  gender: Gender
  birthday?: string
  age?: number
}

export interface KinshipStep {
  relation: KinshipRelation
  person: PersonFacts
  relativeAge?: RelativeAge
}

export interface KinshipContext {
  self: PersonFacts
  steps: KinshipStep[]
  maxDepth: typeof MAX_KINSHIP_DEPTH
}

export interface RelationOption {
  relation: KinshipRelation
  label: string
  enabled: boolean
  disabledReason?: string
}

export interface ValidationResult {
  valid: boolean
  reason?: string
}

export interface KinshipResolution {
  status: 'resolved' | 'ambiguous' | 'unsupported'
  primaryTitle?: string
  /** @deprecated 规范称谓不再返回候选数组 */
  candidates?: string[]
  /** @deprecated 规范称谓不再返回别称 */
  aliases: string[]
  pathDescription: string
  explanation?: string
  incompleteInfo?: boolean
}

export interface KinshipRuleMatch {
  status: 'resolved' | 'ambiguous' | 'unsupported'
  primaryTitle?: string
  candidates?: string[]
  aliases?: string[]
  explanation?: string
  terminal?: boolean
}
