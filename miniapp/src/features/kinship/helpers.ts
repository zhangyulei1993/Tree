import type {
  Gender,
  KinshipContext,
  KinshipRelation,
  KinshipStep,
  PersonFacts,
  RelativeAge
} from './types'

export function getRelationLabel(relation: KinshipRelation): string {
  switch (relation) {
    case 'parent':
      return '父母'
    case 'child':
      return '子女'
    case 'sibling':
      return '兄弟姐妹'
    case 'spouse':
      return '配偶'
    default:
      return '亲属'
  }
}

export function getGenderLabel(gender: Gender): string {
  switch (gender) {
    case 'male':
      return '男'
    case 'female':
      return '女'
    default:
      return '未知'
  }
}

export function getRelativeAgeLabel(relativeAge: RelativeAge): string {
  switch (relativeAge) {
    case 'older':
      return '比上一位年长'
    case 'younger':
      return '比上一位年幼'
    case 'same':
      return '与上一位同龄'
    default:
      return '长幼未知'
  }
}

export function normalizePersonFacts(person: PersonFacts): PersonFacts {
  const normalized: PersonFacts = {
    gender: person.gender || 'unknown'
  }
  if (person.birthday?.trim()) {
    normalized.birthday = person.birthday.trim()
  }
  if (person.age !== undefined && person.age !== null && !Number.isNaN(person.age)) {
    normalized.age = person.age
  }
  if (normalized.birthday) {
    delete normalized.age
  }
  return normalized
}

export function inferRelativeAgeFromFacts(
  current?: PersonFacts,
  sibling?: PersonFacts,
  explicit?: RelativeAge
): RelativeAge {
  if (explicit === 'older' || explicit === 'younger' || explicit === 'same') {
    return explicit
  }

  const currentFacts = current ? normalizePersonFacts(current) : undefined
  const siblingFacts = sibling ? normalizePersonFacts(sibling) : undefined

  if (currentFacts?.birthday && siblingFacts?.birthday) {
    if (siblingFacts.birthday < currentFacts.birthday) return 'older'
    if (siblingFacts.birthday > currentFacts.birthday) return 'younger'
    return 'same'
  }

  if (currentFacts?.age !== undefined && siblingFacts?.age !== undefined) {
    if (siblingFacts.age > currentFacts.age) return 'older'
    if (siblingFacts.age < currentFacts.age) return 'younger'
    return 'same'
  }

  return 'unknown'
}

export function createPathKey(steps: KinshipStep[]): string {
  return steps
    .map((step) => {
      const parts = [step.relation, step.person.gender]
      if (step.relativeAge) parts.push(step.relativeAge)
      if (step.person.birthday) parts.push(`b:${step.person.birthday}`)
      else if (step.person.age !== undefined) parts.push(`a:${step.person.age}`)
      return parts.join('|')
    })
    .join('>')
}

function formatStepRelationLabel(step: KinshipStep): string {
  const { relation, person, relativeAge } = step
  switch (relation) {
    case 'parent':
      if (person.gender === 'male') return '父亲'
      if (person.gender === 'female') return '母亲'
      return '父母'
    case 'child':
      if (person.gender === 'male') return '儿子'
      if (person.gender === 'female') return '女儿'
      return '子女'
    case 'spouse':
      if (person.gender === 'male') return '丈夫'
      if (person.gender === 'female') return '妻子'
      return '配偶'
    case 'sibling':
      if (person.gender === 'male' && relativeAge === 'older') return '哥哥'
      if (person.gender === 'male' && relativeAge === 'younger') return '弟弟'
      if (person.gender === 'female' && relativeAge === 'older') return '姐姐'
      if (person.gender === 'female' && relativeAge === 'younger') return '妹妹'
      return '兄弟姐妹'
    default:
      return getRelationLabel(relation)
  }
}

export function formatStepLabel(step: KinshipStep): string {
  const relationLabel = formatStepRelationLabel(step)
  const genderLabel = getGenderLabel(step.person.gender)
  const details: string[] = [genderLabel]
  if (step.relativeAge && step.relativeAge !== 'unknown') {
    details.push(getRelativeAgeLabel(step.relativeAge))
  }
  if (step.person.birthday) {
    details.push(`生日 ${step.person.birthday}`)
  } else if (step.person.age !== undefined) {
    details.push(`${step.person.age} 岁`)
  }
  return `${relationLabel}（${details.join('，')}）`
}

export function formatPathDescription(context: KinshipContext): string {
  if (context.steps.length === 0) return '我'
  const labels = context.steps.map((step) => formatStepRelationLabel(step))
  return `我的${labels.join('的')}`
}

export function formatPathDisplay(context: KinshipContext): string {
  if (context.steps.length === 0) return '我'
  return ['我', ...context.steps.map((step) => formatStepLabel(step))].join(' > ')
}
