export type BaseRelation = 'parent' | 'child' | 'sibling' | 'spouse'
export type Gender = 'male' | 'female' | 'unknown'
export type AgeOrder = 'older' | 'younger' | 'same' | 'unknown'

export interface KinshipPerson {
  name?: string
  gender: Gender
  birthday?: string
  age?: number | string
}

export interface KinshipSelf extends KinshipPerson {}

export interface KinshipStep extends KinshipPerson {
  relation: BaseRelation
}

export interface KinshipContext {
  self: KinshipSelf
  steps?: KinshipStep[]
  path?: KinshipStep[]
}

export interface KinshipResult {
  title: string
  matched: boolean
  terminal: boolean
  reason: string
  process: string[]
  allowedNextRelations: BaseRelation[]
  displayPath: string
  normalizedPath: KinshipStep[]
}

export interface RelationOption {
  value: BaseRelation
  label: string
  disabled: boolean
  reason?: string
}
