export interface CanonicalKinshipResult {
  canonicalTitle?: string
  rankLabel?: string
  unsupportedCanonicalTitle: boolean
  incompleteInfo: boolean
  pathDescription: string
  explanation?: string
}

export type CanonicalTermKey = string

export interface CanonicalKinshipTerm {
  key: CanonicalTermKey
  canonicalTitle: string
  pathKeys?: string[]
  targetGender?: 'male' | 'female' | 'any'
  seniority?: 'older' | 'younger' | 'same' | 'any'
  lineageSide?: 'paternal' | 'maternal' | 'any'
}
