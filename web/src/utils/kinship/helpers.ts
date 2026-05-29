import type { AgeOrder, BaseRelation, Gender, KinshipContext, KinshipPerson, KinshipStep } from './types'

export const MAX_KINSHIP_LEVEL = 5

export const RELATION_LABELS: Record<BaseRelation, string> = {
  parent: '父母',
  child: '子女',
  sibling: '兄弟姐妹',
  spouse: '配偶'
}

export const BASE_RELATIONS: Array<{ label: string; value: BaseRelation }> = [
  { label: '父母', value: 'parent' },
  { label: '子女', value: 'child' },
  { label: '兄弟姐妹', value: 'sibling' },
  { label: '配偶', value: 'spouse' }
]

export function getSteps(context: KinshipContext): KinshipStep[] {
  const source = Array.isArray(context.steps) ? context.steps : context.path
  return Array.isArray(source) ? source.filter(Boolean).slice(0, MAX_KINSHIP_LEVEL) : []
}

export function pathKey(path: KinshipStep[]) {
  return path.map((item) => item.relation).join('/')
}

export function relationKey(path: KinshipStep[]) {
  return pathKey(path)
}

export function pathText(path: KinshipStep[]) {
  if (!path.length) return '本人'
  return '本人 的 ' + path.map((item) => RELATION_LABELS[item.relation]).join(' 的 ')
}

export function genderLabel(gender: Gender) {
  if (gender === 'male') return '男'
  if (gender === 'female') return '女'
  return '未知'
}

export function isKnownGender(gender: Gender): gender is 'male' | 'female' {
  return gender === 'male' || gender === 'female'
}

export function oppositeGender(gender: Gender): Gender {
  if (gender === 'male') return 'female'
  if (gender === 'female') return 'male'
  return 'unknown'
}

export function endByGender(gender: Gender, male = '父', female = '母', unknown = '辈') {
  if (gender === 'male') return male
  if (gender === 'female') return female
  return unknown
}

export function normalizeAge(value: unknown): number | null {
  if (value === null || value === undefined || value === '') return null
  const n = Number(value)
  return Number.isFinite(n) && n >= 0 ? n : null
}

export function parseBirthday(value?: string): number | null {
  if (!value) return null
  const time = new Date(value).getTime()
  return Number.isFinite(time) ? time : null
}

export function estimatedAgeFromBirthday(value?: string): number | null {
  const birthTime = parseBirthday(value)
  if (birthTime === null) return null

  const birth = new Date(birthTime)
  const today = new Date()
  let age = today.getFullYear() - birth.getFullYear()
  const monthDiff = today.getMonth() - birth.getMonth()

  if (monthDiff < 0 || (monthDiff === 0 && today.getDate() < birth.getDate())) {
    age--
  }

  return age >= 0 ? age : null
}

export function comparableAge(person: KinshipPerson): number | null {
  const directAge = normalizeAge(person.age)
  if (directAge !== null) return directAge
  return estimatedAgeFromBirthday(person.birthday)
}

export function comparePeople(target: KinshipPerson, reference: KinshipPerson): AgeOrder {
  const targetBirth = parseBirthday(target.birthday)
  const referenceBirth = parseBirthday(reference.birthday)

  if (targetBirth !== null && referenceBirth !== null) {
    if (targetBirth < referenceBirth) return 'older'
    if (targetBirth > referenceBirth) return 'younger'
    return 'same'
  }

  const targetAge = comparableAge(target)
  const referenceAge = comparableAge(reference)

  if (targetAge === null || referenceAge === null) return 'unknown'
  if (targetAge > referenceAge) return 'older'
  if (targetAge < referenceAge) return 'younger'
  return 'same'
}

export function titleByGenderAndOrder(
  gender: Gender,
  order: AgeOrder,
  maleOlder: string,
  maleYounger: string,
  maleSameOrUnknown: string,
  femaleOlder: string,
  femaleYounger: string,
  femaleSameOrUnknown: string,
  unknown: string
) {
  if (gender === 'male') {
    if (order === 'older') return maleOlder
    if (order === 'younger') return maleYounger
    return maleSameOrUnknown
  }

  if (gender === 'female') {
    if (order === 'older') return femaleOlder
    if (order === 'younger') return femaleYounger
    return femaleSameOrUnknown
  }

  return unknown
}

export function containsOnly(path: KinshipStep[], relation: BaseRelation) {
  return path.length > 0 && path.every((item) => item.relation === relation)
}

export function hasUnknownGender(path: KinshipStep[]) {
  return path.some((item) => !isKnownGender(item.gender))
}

export function cleanProcess(items: Array<string | undefined | null>) {
  return items.filter((item): item is string => Boolean(item))
}
