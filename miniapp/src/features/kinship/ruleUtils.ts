import { formatPathDescription } from './helpers'
import type { Gender, KinshipContext, KinshipRuleMatch, KinshipResolution } from './types'

export function genderLabel(gender: Gender, male: string, female: string, neutral: string): string {
  if (gender === 'male') return male
  if (gender === 'female') return female
  return neutral
}

export function toResolution(context: KinshipContext, match: KinshipRuleMatch): KinshipResolution {
  const aliases = match.aliases || []
  const pathDescription = formatPathDescription(context)

  if (match.status === 'resolved') {
    return {
      status: 'resolved',
      primaryTitle: match.primaryTitle,
      aliases,
      pathDescription,
      explanation: match.explanation
    }
  }

  if (match.status === 'ambiguous') {
    const candidates = match.candidates || (match.primaryTitle ? [match.primaryTitle] : [])
    return {
      status: 'ambiguous',
      primaryTitle: candidates[0],
      candidates,
      aliases,
      pathDescription,
      explanation: match.explanation || '信息不足，可能存在多种称谓，请补充性别或长幼信息。'
    }
  }

  return {
    status: 'unsupported',
    primaryTitle: match.primaryTitle,
    candidates: match.candidates,
    aliases,
    pathDescription,
    explanation: match.explanation || '暂未收录该关系的常用称谓，你仍然可以保留完整关系路径。'
  }
}
