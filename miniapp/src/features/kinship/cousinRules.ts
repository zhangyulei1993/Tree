import type { KinshipContext, KinshipRuleMatch } from './types'
import { inferRelativeAgeFromFacts } from './helpers'
import { buildRelationKey } from './pathKey'

const COUSIN_RULES: Record<string, KinshipRuleMatch> = {
  'parent:male>sibling:male:older>child:male': { status: 'resolved', primaryTitle: '堂兄', aliases: ['堂哥'] },
  'parent:male>sibling:male:younger>child:male': { status: 'resolved', primaryTitle: '堂弟' },
  'parent:male>sibling:male:older>child:female': { status: 'resolved', primaryTitle: '堂姐' },
  'parent:male>sibling:male:younger>child:female': { status: 'resolved', primaryTitle: '堂妹' },
  'parent:male>sibling:male:unknown>child:male': {
    status: 'ambiguous',
    candidates: ['堂兄', '堂弟', '堂兄弟'],
    explanation: '堂亲兄弟需要补充长幼信息。'
  },
  'parent:male>sibling:male:unknown>child:female': {
    status: 'ambiguous',
    candidates: ['堂姐', '堂妹', '堂姐妹'],
    explanation: '堂亲姐妹需要补充长幼信息。'
  },
  'parent:male>sibling:female:older>child:male': {
    status: 'ambiguous',
    candidates: ['表兄', '表弟', '表兄弟'],
    explanation: '姑表兄弟称谓因长幼而异，请补充信息。'
  },
  'parent:male>sibling:female:younger>child:male': {
    status: 'ambiguous',
    candidates: ['表兄', '表弟', '表兄弟'],
    explanation: '姑表兄弟称谓因长幼而异，请补充信息。'
  },
  'parent:male>sibling:female:older>child:female': {
    status: 'ambiguous',
    candidates: ['表姐', '表妹', '表姐妹'],
    explanation: '姑表姐妹称谓因长幼而异，请补充信息。'
  },
  'parent:male>sibling:female:younger>child:female': {
    status: 'ambiguous',
    candidates: ['表姐', '表妹', '表姐妹'],
    explanation: '姑表姐妹称谓因长幼而异，请补充信息。'
  },
  'parent:male>sibling:female:unknown>child:unknown': {
    status: 'ambiguous',
    candidates: ['表兄', '表弟', '表姐', '表妹', '表亲'],
    explanation: '姑表亲需要补充子女性别和长幼信息。'
  },

  'parent:female>sibling:male:older>child:male': { status: 'resolved', primaryTitle: '表兄', aliases: ['表哥'] },
  'parent:female>sibling:male:younger>child:male': { status: 'resolved', primaryTitle: '表弟' },
  'parent:female>sibling:male:older>child:female': { status: 'resolved', primaryTitle: '表姐' },
  'parent:female>sibling:male:younger>child:female': { status: 'resolved', primaryTitle: '表妹' },
  'parent:female>sibling:male:unknown>child:male': {
    status: 'ambiguous',
    candidates: ['表兄', '表弟', '表兄弟'],
    explanation: '舅表兄弟需要补充长幼信息。'
  },
  'parent:female>sibling:female:older>child:male': { status: 'resolved', primaryTitle: '表兄', aliases: ['表哥'] },
  'parent:female>sibling:female:younger>child:male': { status: 'resolved', primaryTitle: '表弟' },
  'parent:female>sibling:female:older>child:female': { status: 'resolved', primaryTitle: '表姐' },
  'parent:female>sibling:female:younger>child:female': { status: 'resolved', primaryTitle: '表妹' },
  'parent:female>sibling:female:unknown>child:female': {
    status: 'ambiguous',
    candidates: ['表姐', '表妹', '表姐妹'],
    explanation: '姨表姐妹需要补充长幼信息。'
  },
  'parent:unknown>sibling:male:older>child:male': {
    status: 'ambiguous',
    candidates: ['堂兄', '堂弟', '表兄', '表弟'],
    explanation: '需要补充父母性别以区分堂亲或表亲。'
  }
}

function getCousinSide(pathKey: string): 'paternal' | 'maternalOrCross' | 'unknown' {
  const parts = pathKey.split('>')
  const parentGender = parts[0]?.split(':')[1]
  const siblingGender = parts[1]?.split(':')[1]
  if (parentGender === 'male' && siblingGender === 'male') return 'paternal'
  if (parentGender === 'unknown' || siblingGender === 'unknown') return 'unknown'
  return 'maternalOrCross'
}

function cousinPrefixFromPathKey(pathKey: string): '堂' | '表' | null {
  const side = getCousinSide(pathKey)
  if (side === 'paternal') return '堂'
  if (side === 'maternalOrCross') return '表'
  return null
}

/**
 * 堂/表同辈称谓按「我 vs 堂表亲本人」出生先后判断，不按父母兄弟姐妹（伯父/叔父等）长幼。
 * 优先使用用户提供的长幼关系；没有长幼信息时再尝试用生日/年龄推导。
 */
export function matchCousinPeerSeniority(
  context: KinshipContext,
  pathKey: string
): KinshipRuleMatch | null {
  if (buildRelationKey(context) !== 'parent>sibling>child') return null

  const childStep = context.steps[2]
  const peerSeniority = inferRelativeAgeFromFacts(context.self, childStep.person, childStep.relativeAge)
  if (peerSeniority !== 'older' && peerSeniority !== 'younger') return null

  const prefix = cousinPrefixFromPathKey(pathKey)
  if (!prefix) return null

  const childGender = childStep.person.gender
  if (childGender === 'male') {
    return {
      status: 'resolved',
      primaryTitle: peerSeniority === 'older' ? `${prefix}兄` : `${prefix}弟`
    }
  }
  if (childGender === 'female') {
    return {
      status: 'resolved',
      primaryTitle: peerSeniority === 'older' ? `${prefix}姐` : `${prefix}妹`
    }
  }
  return null
}

function genderedTitle(gender: string | undefined, male: string, female: string, neutral: string): KinshipRuleMatch {
  if (gender === 'male') return { status: 'resolved', primaryTitle: male }
  if (gender === 'female') return { status: 'resolved', primaryTitle: female }
  return {
    status: 'ambiguous',
    primaryTitle: neutral,
    candidates: [neutral, male, female],
    explanation: '需要补充性别以区分更具体的称谓。'
  }
}

function matchParentGenerationCousinRule(context: KinshipContext): KinshipRuleMatch | null {
  if (buildRelationKey(context) !== 'parent>parent>sibling>child') return null

  const parentStep = context.steps[0]
  const grandparentStep = context.steps[1]
  const targetStep = context.steps[3]
  const parentGender = parentStep.person.gender
  const grandparentGender = grandparentStep.person.gender
  const targetGender = targetStep.person.gender

  if (targetGender === 'female') {
    if (parentGender === 'male' && grandparentGender === 'male' && context.steps[2].person.gender === 'male') {
      return { status: 'resolved', primaryTitle: '堂姑' }
    }
    if (parentGender === 'female') {
      return { status: 'resolved', primaryTitle: '表姨' }
    }
    return { status: 'resolved', primaryTitle: '表姑' }
  }

  if (targetGender !== 'male') return null

  const seniority = inferRelativeAgeFromFacts(parentStep.person, targetStep.person, targetStep.relativeAge)
  if (parentGender === 'male' && grandparentGender === 'male' && context.steps[2].person.gender === 'male') {
    if (seniority === 'older') return { status: 'resolved', primaryTitle: '堂伯' }
    if (seniority === 'younger') return { status: 'resolved', primaryTitle: '堂叔' }
    return {
      status: 'ambiguous',
      candidates: ['堂伯', '堂叔'],
      explanation: '父亲的祖辈旁系男性需要补充与父亲的长幼关系。'
    }
  }

  if (parentGender === 'female') {
    if (seniority === 'older') return { status: 'resolved', primaryTitle: '表舅' }
    if (seniority === 'younger') return { status: 'resolved', primaryTitle: '表叔' }
    return {
      status: 'ambiguous',
      candidates: ['表舅', '表叔'],
      explanation: '母亲的祖辈旁系男性需要补充与母亲的长幼关系。'
    }
  }

  if (seniority === 'older') return { status: 'resolved', primaryTitle: '表伯' }
  if (seniority === 'younger') return { status: 'resolved', primaryTitle: '表叔' }
  return {
    status: 'ambiguous',
    candidates: ['表伯', '表叔'],
    explanation: '父母的祖辈旁系男性需要补充长幼关系。'
  }
}

function matchCousinDescendantRules(context: KinshipContext, pathKey: string): KinshipRuleMatch | null {
  const relationKey = buildRelationKey(context)
  if (relationKey !== 'parent>sibling>child>child' && relationKey !== 'parent>sibling>child>child>child') {
    return null
  }

  const side = getCousinSide(pathKey)
  const lastGender = pathKey.split('>').pop()?.split(':')[1]
  if (relationKey === 'parent>sibling>child>child') {
    if (side === 'paternal') {
      return genderedTitle(lastGender, '堂侄', '堂侄女', '堂侄子女')
    }
    if (side === 'maternalOrCross') {
      return genderedTitle(lastGender, '表侄', '表侄女', '表侄子女')
    }
    return {
      status: 'ambiguous',
      primaryTitle: '堂表亲侄辈',
      candidates: ['堂侄', '堂侄女', '表侄', '表侄女', '堂表亲侄辈'],
      explanation: '需要补充父母性别、旁系性别和孩子性别以区分堂亲或表亲侄辈。'
    }
  }

  if (side === 'paternal') {
    return genderedTitle(lastGender, '堂侄孙', '堂侄孙女', '堂侄孙辈')
  }
  if (side === 'maternalOrCross') {
    return genderedTitle(lastGender, '表侄孙', '表侄孙女', '表侄孙辈')
  }
  return {
    status: 'ambiguous',
    primaryTitle: '堂表亲侄孙辈',
    candidates: ['堂侄孙', '堂侄孙女', '表侄孙', '表侄孙女', '堂表亲侄孙辈'],
    explanation: '需要补充父母性别、旁系性别和后代性别以区分堂亲或表亲侄孙辈。'
  }
}

export function matchCousinRules(context: KinshipContext, pathKey: string): KinshipRuleMatch | null {
  const peerMatch = matchCousinPeerSeniority(context, pathKey)
  if (peerMatch) return peerMatch
  const parentGenerationMatch = matchParentGenerationCousinRule(context)
  if (parentGenerationMatch) return parentGenerationMatch
  return COUSIN_RULES[pathKey] || matchCousinDescendantRules(context, pathKey)
}

export { COUSIN_RULES }
