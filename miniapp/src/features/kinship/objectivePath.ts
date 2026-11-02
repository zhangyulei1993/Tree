import type { Gender, KinshipContext, KinshipStep } from './types'
import { inferRelativeAgeFromFacts } from './helpers'

function canonicalStepLabel(step: KinshipStep): string {
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
      if (person.gender === 'male' && relativeAge === 'older') return '兄长'
      if (person.gender === 'male' && relativeAge === 'younger') return '弟弟'
      if (person.gender === 'female' && relativeAge === 'older') return '姐姐'
      if (person.gender === 'female' && relativeAge === 'younger') return '妹妹'
      return '兄弟姐妹'
    default:
      return '关系'
  }
}

/** 客观关系路径，如「父亲 → 兄长 → 女儿」 */
export function formatObjectivePathDescription(context: KinshipContext): string {
  if (context.steps.length === 0) return '我'
  return ['我', ...context.steps.map((step) => canonicalStepLabel(step))].join(' → ')
}

export function pathHasIncompleteInfo(context: KinshipContext): boolean {
  return context.steps.some((step, index) => {
    if (step.person.gender === 'unknown') return true
    if (step.relation === 'sibling') {
      const isTerminal = index === context.steps.length - 1
      if (!isTerminal) return false
      const reference = index === 0 ? context.self : context.steps[index - 1].person
      const inferred = inferRelativeAgeFromFacts(reference, step.person, step.relativeAge)
      if (!inferred || inferred === 'unknown') return true
    }
    return false
  })
}

export function isIncompleteGender(gender: Gender): boolean {
  return gender === 'unknown'
}
