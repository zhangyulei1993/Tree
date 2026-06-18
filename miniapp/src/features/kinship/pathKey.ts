import { inferRelativeAgeFromFacts } from './helpers'
import type { Gender, KinshipContext, KinshipStep } from './types'

function normalizeGender(gender: Gender): Gender {
  return gender || 'unknown'
}

export function buildStepSegment(context: KinshipContext, step: KinshipStep, index: number): string {
  const gender = normalizeGender(step.person.gender)
  if (step.relation === 'sibling') {
    const reference = index === 0 ? context.self : context.steps[index - 1].person
    const relativeAge = inferRelativeAgeFromFacts(reference, step.person, step.relativeAge)
    return `sibling:${gender}:${relativeAge}`
  }
  return `${step.relation}:${gender}`
}

export function buildPathKey(context: KinshipContext): string {
  return context.steps.map((step, index) => buildStepSegment(context, step, index)).join('>')
}

export function buildRelationKey(context: KinshipContext): string {
  return context.steps.map((step) => step.relation).join('>')
}
